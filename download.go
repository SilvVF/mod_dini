package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func (s InstallState) String() string {
	switch s {
	case InstallErr:
		return "error"
	case InstallQueued:
		return "queued"
	case InstallDownloading:
		return "downloading"
	case InstallExtracting:
		return "extracting"
	case InstallReady:
		return "ready"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

func (ms *ModService) DownloadMod(mod ModSaved) error {
	log.Printf("[mod] download start mod=%s packages=%d", mod.ID, len(mod.Packs))
	start := time.Now()

	outputDir, err := GetSetting[string](ms.settings, "zzmi.folder")
	if err != nil {
		log.Printf("[mod] read output dir setting failed: %v", err)
		return err
	}

	var errs []error
	downloaded, skipped := 0, 0

	for _, pkg := range mod.Packs {
		// already downloaded; Err and Queued packages get (re)tried
		if pkg.State == InstallReady {
			log.Printf("[mod] skip mod=%s pkg=%s (already ready)", mod.ID, pkg.ID)
			skipped++
			continue
		}

		if err := ms.downloadPackage(context.Background(), mod, pkg, outputDir); err != nil {
			log.Printf("[mod] package failed mod=%s pkg=%s: %v", mod.ID, pkg.ID, err)

			ms.updatePack(mod.ID, pkg.ID, func(p *Pack) error {
				p.State = InstallErr
				return nil
			})

			errs = append(errs, fmt.Errorf("package %s: %w", pkg.ID, err))
			continue
		}
		downloaded++
	}

	log.Printf(
		"[mod] download done mod=%s ok=%d skipped=%d failed=%d took=%s",
		mod.ID, downloaded, skipped, len(errs), time.Since(start).Round(time.Millisecond),
	)

	return errors.Join(errs...)
}

func (m ModSaved) GetPackage(id string) (Pack, int, bool) {
	i := slices.IndexFunc(m.Packs, func(p Pack) bool { return p.ID == id })
	if i < 0 {
		return Pack{}, -1, false
	}
	return m.Packs[i], i, true
}

func (m ModSaved) SubmissionFileForPackage(pkg Pack) (SubmissionFile, bool) {
	i := slices.IndexFunc(m.SubmissionFiles, func(f SubmissionFile) bool {
		return f.FileID == pkg.FileID
	})

	if i < 0 {
		return SubmissionFile{}, false
	}

	return m.SubmissionFiles[i], true
}

func (ms *ModService) EmitDownloadProgress(mod ModSaved, pkg Pack, total, progress int64) bool {
	return ms.app.Event.Emit("mod-download-progress", DataProgress{
		ID:        mod.ID,
		PackageID: pkg.ID,
		State:     int(pkg.State),
		Total:     total,
		Progress:  progress,
	})
}

func (ms *ModService) downloadPackage(
	ctx context.Context,
	mod ModSaved,
	pkg Pack,
	outputDir string,
) error {
	dlLog := newProgressLogger(fmt.Sprintf("download mod=%s pkg=%s", mod.ID, pkg.ID))
	subFile, ok := mod.SubmissionFileForPackage(pkg)

	if !ok {
		return fmt.Errorf("no submission file with id %s", pkg.FileID)
	}

	log.Printf(
		"[mod] package start mod=%s pkg=%s file=%q size=%d url=%s",
		mod.ID, pkg.ID, subFile.FileName, subFile.Size, subFile.DownloadURL,
	)
	start := time.Now()

	pkg, err := ms.updatePack(mod.ID, pkg.ID, func(p *Pack) error {
		p.State = InstallDownloading
		return nil
	})
	if err != nil {
		return err
	}

	f, err := httpDownload(ctx, subFile, func(total, progress int64) {
		dlLog(total, progress)
		ms.EmitDownloadProgress(mod, pkg, total, progress)
	})

	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	defer os.Remove(f.Name())
	defer f.Close()

	pkg, err = ms.updatePack(mod.ID, pkg.ID, func(p *Pack) error {
		p.State = InstallExtracting
		return nil
	})
	if err != nil {
		return err
	}

	dst := filepath.Join(outputDir, mod.Submission.Provider, mod.Submission.SubmissionID, pkg.ID)
	relPath, err := filepath.Rel(outputDir, dst)
	if err != nil {
		return fmt.Errorf("rel path: %w", err)
	}

	log.Printf("[mod] extract start mod=%s pkg=%s dst=%s", mod.ID, pkg.ID, dst)

	err = ArchiveExtract(ExtractOptions{
		ctx:       ctx,
		archive:   f,
		dst:       dst,
		overwrite: true,
		onProgress: func(progress, total int64) {
			dlLog(total, progress)
			ms.EmitDownloadProgress(mod, pkg, total, progress)
		},
	})

	if err != nil {
		log.Printf("[mod] extract failed mod=%s pkg=%s, removing %s: %v", mod.ID, pkg.ID, dst, err)
		os.RemoveAll(dst)
		return fmt.Errorf("extract: %w", err)
	}

	size, err := DirSize(dst)
	if err != nil {
		return fmt.Errorf("stat extracted dir: %w", err)
	}

	pkg, err = ms.updatePack(mod.ID, pkg.ID, func(p *Pack) error {
		p.State = InstallReady
		p.Storage = StorageActive
		p.Enabled = true
		p.FileID = subFile.FileID
		p.Manifest = ManifestEntry{
			RelPath: relPath,
			Size:    size,
			Hash:    subFile.MD5Checksum,
		}
		return nil
	})

	log.Printf(
		"[mod] package ready mod=%s pkg=%s took=%s",
		mod.ID, pkg.ID, time.Since(start).Round(time.Millisecond),
	)
	return err
}

func httpDownload(
	ctx context.Context,
	sf SubmissionFile,
	onProgress func(total, progress int64),
) (*os.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sf.DownloadURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Printf(
		"[mod] http response file=%q status=%s content-length=%d",
		sf.FileName, res.Status, res.ContentLength,
	)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", res.Status)
	}

	total := res.ContentLength
	if total <= 0 {
		log.Printf("[mod] content-length unknown for %q, using stored size %d", sf.FileName, sf.Size)
		total = sf.Size
	}

	f, err := os.CreateTemp(MOD_TEST_FOLDER, "*-"+sf.FileName)
	if err != nil {
		return nil, err
	}
	log.Printf("[mod] temp file created %s", f.Name())

	cw := &countingWriter{onProgress: func(progress int64) { onProgress(total, progress) }}
	n, err := io.Copy(f, io.TeeReader(res.Body, cw))
	if err != nil {
		log.Printf("[mod] copy failed file=%q after %d bytes: %v", sf.FileName, n, err)
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	log.Printf("[mod] downloaded file=%q bytes=%d", sf.FileName, n)

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}

	return f, nil
}

// newProgressLogger returns a callback that logs at most once per 25% step.
func newProgressLogger(label string) func(total, progress int64) {
	lastStep := -1
	return func(total, progress int64) {
		if total <= 0 {
			return
		}
		step := int(progress * 4 / total) // 0..4
		if step > lastStep {
			lastStep = step
			log.Printf("[mod] %s %d%% (%d/%d)", label, step*25, progress, total)
		}
	}
}

type countingWriter struct {
	count      int64
	onProgress func(int64)
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n := len(p)
	cw.count += int64(n)
	cw.onProgress(cw.count)
	return n, nil
}
