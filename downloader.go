package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"go.etcd.io/bbolt"
)

func (ms ModSaved) UpdateState(state ModState, m *ModService) error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(DL_BUCKET)

		var mod ModSaved
		mbytes := b.Get([]byte(ms.ID))
		json.Unmarshal(mbytes, &mod)

		mod.State = state

		mbytes, err := json.Marshal(mod)
		if err != nil {
			return err
		}

		err = b.Put([]byte(mod.ID), mbytes)

		if err != nil {
			return err
		}

		m.app.Event.Emit("mod-update", mod)

		return nil
	})
}

func (m *ModService) DownloadMod(mod ModSaved, catID string) (err error) {

	defer func() {
		if err != nil {
			mod.UpdateState(ModStateErr, m)
		}
	}()

	downloadList := []SubmissionFile{}
	for _, f := range mod.Files {

		selected := slices.ContainsFunc(mod.Tree.Files, func(mf ModFile) bool {
			return mf.ArchiveName == f.FileName
		})

		if selected {
			downloadList = append(downloadList, f)
		}
	}

	sf := downloadList[0]

	mod.UpdateState(ModStateDownloading, m)

	f, err := httpDownload(sf, func(total, progress int64) {
		m.app.Event.Emit("mod-download-progress", DataProgress{
			ID:       mod.ID,
			State:    1,
			Total:    total,
			Progress: progress,
		})
	})
	if err != nil {
		return
	}

	defer func() {
		name := f.Name()
		f.Close()
		os.Remove(name)
	}()

	fname := filepath.Base(f.Name())
	dirName := strings.TrimSuffix(fname, filepath.Ext(fname))

	mod.UpdateState(ModStateExtracting, m)

	err = ArchiveExtract(ExtractOptions{
		ctx:       context.Background(),
		archive:   f,
		dst:       filepath.Join(MOD_TEST_FOLDER, dirName),
		overwrite: true,
		onProgress: func(progress int64, total int64) {
			m.app.Event.Emit("mod-download-progress", DataProgress{
				ID:       mod.ID,
				State:    int(ModStateExtracting),
				Total:    total,
				Progress: progress,
			})
		},
	})

	mod.UpdateState(ModStateActive, m)

	return
}

func httpDownload(
	sf SubmissionFile,
	onProgress func(total, progress int64),
) (file *os.File, err error) {

	fmt.Println("starting download", sf.DownloadURL)

	resp, err := http.Head(sf.DownloadURL)
	if err != nil {
		return
	}

	totalSize, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		return
	}

	res, err := http.Get(sf.DownloadURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	byteCounter := io.TeeReader(res.Body, &countingWriter{
		onProgress: func(progress int64) {
			onProgress(int64(totalSize), progress)
		},
	})

	f, err := os.Create(filepath.Join(MOD_TEST_FOLDER, sf.FileName))
	if err != nil {
		return
	}

	_, err = io.Copy(f, byteCounter)
	if err != nil {
		return
	}

	return f, err
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

var unarrExtensions = map[string]struct{}{
	".zip":     {},
	".rar":     {},
	".7z":      {},
	".tar":     {},
	".gz":      {},
	".bz2":     {},
	".xz":      {},
	".tar.gz":  {},
	".tar.bz2": {},
	".tar.xz":  {},
	".iso":     {},
	".cab":     {},
	".lzma":    {},
	".cpio":    {},
	".z":       {},
	".lz":      {},
}
