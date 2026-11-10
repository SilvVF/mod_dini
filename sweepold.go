package main

//
// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"sync"
// 	"time"
//
// 	"github.com/mholt/archives"
// 	"go.etcd.io/bbolt"
// )
//
// type Sweeper struct {
// 	ms    *ModService
// 	ctx   context.Context
// 	mutex sync.Mutex
// 	tasks map[string]context.CancelFunc
// }
//
// const (
// 	DISABLE_MAX_TIME = 10 * time.Minute
// )
//
// func NewSweeper(ms *ModService) *Sweeper {
// 	return &Sweeper{
// 		ms:    ms,
// 		ctx:   ms.app.Context(),
// 		tasks: map[string]context.CancelFunc{},
// 	}
// }
//
// func (s *Sweeper) Cancel(modId string) {
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()
//
// 	cancel := s.tasks[modId]
// 	if cancel != nil {
// 		cancel()
// 	}
// }
//
// func (s *Sweeper) Run() {
// 	ticker := time.NewTicker(time.Second * 30)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-s.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			s.doSweep()
// 		}
// 	}
// }
//
// func (s *Sweeper) doSweep() {
// 	toUpdate := map[string]struct{}{}
// 	s.ms.db.View(func(tx *bbolt.Tx) error {
// 		bucket := tx.Bucket(DL_BUCKET)
// 		return bucket.ForEach(func(k, v []byte) error {
// 			toUpdate[string(k)] = struct{}{}
// 			return nil
// 		})
// 	})
//
// 	for mid, _ := range toUpdate {
// 		ctx, cancel := context.WithCancel(s.ctx)
//
// 		s.mutex.Lock()
// 		s.tasks[mid] = cancel
// 		s.mutex.Unlock()
//
// 		var mod ModSaved
// 		s.ms.db.View(func(tx *bbolt.Tx) error {
// 			data := tx.Bucket(DL_BUCKET).Get([]byte(mid))
// 			return json.Unmarshal(data, &mod)
// 		})
//
// 		for _, p := range mod.Packs {
//
// 			if ctx.Err() != nil {
// 				break
// 			}
//
// 			if p.Storage != StorageWarm {
// 				continue
// 			}
//
// 			disabledAtUnix := time.Unix(p.DisabledAt, 0)
// 			if time.Now().After(disabledAtUnix.Add(DISABLE_MAX_TIME)) {
// 				if _, err := s.ms.updatePack(mod.ID, p.ID, func(p *Pack) error {
// 					p.Storage = StorageCompressing
// 					return ctx.Err()
// 				}); err != nil {
// 					continue
// 				}
//
// 				err := s.compressPack(mod, p, ctx)
// 				log.Println(err)
// 			}
// 		}
//
// 		s.mutex.Lock()
// 		s.tasks[mid] = nil
// 		s.mutex.Unlock()
// 		cancel()
// 	}
// }
//
// func (s *Sweeper) compressPack(m ModSaved, p Pack, ctx context.Context) error {
// 	format := archives.CompressedArchive{
// 		Archival: archives.Zip{},
// 	}
// 	toolDir, err := GetSetting[string](s.ms.settings, m.ToolId+".folder")
// 	if err != nil {
// 		return err
// 	}
//
// 	modDir := filepath.Join(toolDir, m.Submission.Provider, m.Submission.Slug)
// 	packDir := filepath.Join(modDir, "DISABLED_"+p.ID)
// 	compressFilePath := filepath.Join(modDir, p.ID+".part")
//
// 	out, err := os.Create(compressFilePath)
// 	if err != nil {
// 		return err
// 	}
//
// 	files, err := archives.FilesFromDisk(ctx, nil, map[string]string{
// 		packDir: "",
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	err = format.Archive(ctx, out, files)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = os.Rename(compressFilePath, strings.ReplaceAll(compressFilePath, ".part", ".zip"))
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = s.ms.updatePack(m.ID, p.ID, func(p *Pack) error {
// 		p.Storage = StorageCold
// 		p.ArchiveFormat = ".zip"
// 		// todo(get archive size)
// 		p.ArchiveSize = 0
//
// 		return ctx.Err()
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	return os.RemoveAll(packDir)
// }
