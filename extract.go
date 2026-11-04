package main

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

type ExtractOptions struct {
	ctx        context.Context
	archive    *os.File
	dst        string
	overwrite  bool
	onProgress func(progress int64, total int64)
}

func ArchiveExtract(options ExtractOptions) error {
	ctx := options.ctx
	f := options.archive
	dst := options.dst
	overwrite := options.overwrite
	onProgress := options.onProgress

	if f == nil {
		return errors.New("archive file is nil")
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	archivePath := f.Name()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	format, stream, err := archives.Identify(ctx, archivePath, f)
	if err != nil {
		return err
	}

	total, err := uncompressedSize(ctx, archivePath)
	if err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	ex, ok := format.(archives.Extractor)
	if !ok {
		return errors.New("nothing to extract")
	}

	progress := int64(0)

	if onProgress != nil {
		onProgress(progress, total)
	}

	return ex.Extract(ctx, stream, func(ctx context.Context, entry archives.FileInfo) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		name := filepath.Clean(filepath.FromSlash(entry.NameInArchive))

		if name == "." || filepath.IsAbs(name) || name == ".." ||
			strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return errors.New("invalid archive entry path: " + entry.NameInArchive)
		}

		dstPath := filepath.Join(dst, name)

		// Verify the resulting path remains inside the destination.
		rel, err := filepath.Rel(dst, dstPath)
		if err != nil {
			return err
		}

		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return errors.New("archive entry escapes destination: " + entry.NameInArchive)
		}

		if entry.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if !overwrite && FileExists(dstPath) {
			if onProgress != nil {
				progress += entry.Size()
				onProgress(progress, total)
			}
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		src, err := entry.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			return err
		}

		_, copyErr := io.Copy(out, src)
		srcErr := src.Close()
		outErr := out.Close()

		if copyErr != nil {
			return copyErr
		}
		if srcErr != nil {
			return srcErr
		}
		if outErr != nil {
			return outErr
		}

		if onProgress != nil {
			progress += entry.Size()
			onProgress(progress, total)
		}

		return nil
	})
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func uncompressedSize(ctx context.Context, archivePath string) (int64, error) {
	total := int64(0)

	fsys, err := archives.FileSystem(ctx, archivePath, nil)
	if err != nil {
		return -1, err
	}

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		total += info.Size()
		return nil
	})

	return total, err
}
