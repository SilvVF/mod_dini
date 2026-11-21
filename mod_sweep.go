package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archives"
	"go.etcd.io/bbolt"
)

const (
	DISABLE_MAX_TIME = 24 * time.Hour
	SWEEP_INTERVAL   = 10 * time.Minute
)

var (
	ErrPackNotWarm        = errors.New("pack is not warm")
	ErrPackNotReady       = errors.New("pack is not ready for compression")
	ErrPackNotCompressing = errors.New("pack is not compressing")
)

type sweepTask struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

type Sweeper struct {
	ms    *ModService
	ctx   context.Context
	mutex sync.Mutex
	tasks map[string]*sweepTask
}

func NewSweeper(ms *ModService, ctx context.Context) *Sweeper {
	return &Sweeper{
		ms:    ms,
		ctx:   ctx,
		tasks: make(map[string]*sweepTask),
	}
}

func (s *Sweeper) Run() {
	ticker := time.NewTicker(SWEEP_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-ticker.C:
			s.doSweep()
		}
	}
}

func (s *Sweeper) CancelAndWait(modID string) {
	s.mutex.Lock()
	t := s.tasks[modID]
	s.mutex.Unlock()
	if t == nil {
		return
	}
	t.cancel()
	<-t.done
}

func (s *Sweeper) startTask(modID string) (*sweepTask, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.ctx.Err() != nil {
		return nil, false
	}

	if _, exists := s.tasks[modID]; exists {
		return nil, false
	}

	ctx, cancel := context.WithCancel(s.ctx)

	task := &sweepTask{
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	s.tasks[modID] = task

	return task, true
}

func (s *Sweeper) finishTask(modID string, task *sweepTask) {
	s.mutex.Lock()
	if s.tasks[modID] == task {
		delete(s.tasks, modID)
	}
	s.mutex.Unlock()

	close(task.done)
	task.cancel()
}

func (s *Sweeper) doSweep() {
	var modIDs []string

	err := s.ms.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(DL_BUCKET)
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, _ []byte) error {
			modIDs = append(modIDs, string(k))
			return nil
		})
	})

	if err != nil {
		log.Printf("sweeper: failed to enumerate mods: %v", err)
		return
	}

	for _, modID := range modIDs {
		if s.ctx.Err() != nil {
			return
		}

		task, started := s.startTask(modID)
		if !started {
			continue
		}

		unlock := s.ms.lockMod(modID)

		if err := s.sweepMod(modID, task.ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("sweeper: mod %s: %v", modID, err)
			}
		}

		unlock()
		s.finishTask(modID, task)
	}
}

func (s *Sweeper) sweepMod(
	modID string,
	ctx context.Context,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var mod ModSaved

	err := s.ms.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(DL_BUCKET)
		data := bucket.Get([]byte(modID))
		return json.Unmarshal(data, &mod)
	})

	if err != nil {
		return err
	}

	now := time.Now()

	for _, pack := range mod.Packs {
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, err := s.ms.updatePack(modID, pack.ID, func(p *Pack) error {
			if p.Storage != StorageWarm {
				return ErrPackNotWarm
			}

			disabledAt := time.Unix(p.DisabledAt, 0)
			if now.Before(disabledAt.Add(DISABLE_MAX_TIME)) {
				return ErrPackNotReady
			}

			p.Storage = StorageCompressing
			return ctx.Err()
		}); err != nil {
			continue
		}

		err = s.compressPack(mod, pack, ctx)
		if err != nil {
			log.Printf(
				"sweeper: failed to compress mod %s pack %s: %v",
				mod.ID,
				pack.ID,
				err,
			)

			s.restoreWarm(mod.ID, pack.ID)
			continue
		}
	}

	return nil
}

func (s *Sweeper) restoreWarm(
	modID string,
	packID string,
) {
	_, err := s.ms.updatePack(modID, packID, func(p *Pack) error {
		if p.Storage != StorageCompressing {
			return nil
		}
		p.Storage = StorageWarm
		return nil
	})

	if err != nil {
		log.Printf(
			"sweeper: failed to restore pack %s to warm: %v",
			packID,
			err,
		)
	}
}

func (s *Sweeper) compressPack(
	m ModSaved,
	p Pack,
	ctx context.Context,
) error {
	format := archives.CompressedArchive{
		Archival: archives.Zip{
			Compression: zip.Deflate,
		},
	}

	toolDir, err := GetSetting[string](
		s.ms.settings,
		m.ToolId+".folder",
	)
	if err != nil {
		return err
	}

	modDir := filepath.Join(toolDir, m.Submission.Provider, m.Submission.SubmissionID)
	packDir := filepath.Join(modDir, "DISABLED_"+p.ID)
	archivePath := filepath.Join(modDir, p.ID+".part")

	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(archivePath)

	files, err := archives.FilesFromDisk(
		ctx,
		nil,
		map[string]string{
			packDir + string(filepath.Separator): "",
		},
	)
	if err != nil {
		return err
	}

	if err := format.Archive(ctx, out, files); err != nil {
		return err
	}

	out.Close()

	_, err = s.ms.updatePack(m.ID, p.ID, func(p *Pack) error {

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if p.Storage != StorageCompressing {
			return ErrPackNotCompressing
		}

		finalPath := strings.ReplaceAll(archivePath, ".part", ".zip")
		if err = os.Rename(archivePath, finalPath); err != nil {
			return err
		}

		info, err := os.Stat(finalPath)
		if err != nil {
			return err
		}

		p.Storage = StorageCold
		p.ArchiveFormat = ".zip"
		p.ArchiveSize = info.Size()

		if err := os.RemoveAll(packDir); err != nil {
			log.Printf(
				"sweeper: failed to remove original pack directory %s: %v",
				packDir,
				err,
			)
			return err
		}

		return nil
	})

	return err
}
