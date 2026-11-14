package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func fileSize(path string) int64 {
	if info, err := os.Stat(path); err == nil {
		return info.Size()
	}
	return 0
}

type observed struct {
	active, disabled, archive, temp, part bool
}

func observe(p packPaths) observed {
	return observed{
		active:   exists(p.active),
		disabled: exists(p.disabled),
		archive:  exists(p.archive),
		temp:     exists(p.temp),
		part:     exists(p.part),
	}
}

func formOnDisk(was StorageState, d observed) StorageState {
	// With one copy on disk there is nothing to decide. The recorded state only breaks ties
	// when an archive sits beside a leftover folder:
	//   Cold / Restoring -> the archive was committed, the folder is a half-finished delete
	//   Compressing      -> the archive isn't trusted until Cold is recorded, so the folder wins
	archiveWins := d.archive && !d.active && (was == StorageCold || was == StorageRestoring)
	if archiveWins {
		return StorageCold
	}

	switch {
	case d.active:
		return StorageActive
	case d.disabled:
		return StorageWarm
	case d.archive:
		return StorageCold
	}
	return StorageNone
}

// corrected record + files to delete out.
func planPack(p Pack, paths packPaths, d observed, archiveSize, now int64) (fixed Pack, remove []string) {
	switch p.State {
	// app closed while downloading move to the error state
	case InstallQueued, InstallDownloading, InstallExtracting:
		p.State = InstallErr
		p.Storage = StorageNone
		p.Enabled = false
		return p, nil
	case InstallErr:
		return p, nil
	}

	was := p.Storage
	form := formOnDisk(was, d)

	// Leftovers from transitions.
	if d.part {
		remove = append(remove, paths.part)
	}
	if d.temp {
		remove = append(remove, paths.temp)
	}

	// folder exists archive not deleted for some reason
	compressInterrupted := was == StorageCompressing && form == StorageWarm
	// something failed while renaming
	restoreFinished := was == StorageRestoring && form == StorageActive
	archiveIsStale := d.archive && (compressInterrupted || restoreFinished)
	if archiveIsStale {
		remove = append(remove, paths.archive)
	}

	switch form {
	case StorageNone:
		// Delete nothing that could be the only copy.
		p.State = InstallErr
		p.Storage = StorageNone
		// todo(drop enabled flag)
		p.Enabled = false
		return p, remove
	case StorageActive:
		p.DisabledAt = 0
		p.ArchiveFormat = ""
		p.ArchiveSize = 0
	case StorageWarm:
		// restarts the warm window
		if p.DisabledAt <= 0 {
			p.DisabledAt = now
		}
		p.ArchiveFormat = ""
		p.ArchiveSize = 0
	case StorageCold:
		p.ArchiveFormat = filepath.Ext(paths.archive)
		p.ArchiveSize = archiveSize
	}

	p.Storage, p.Enabled = form, form == StorageActive
	return p, remove
}

// sets the state of the bolt mod entries to match what is on the disk and cleans up leftovers
func (ms *ModService) reconcile() error {
	mods, err := ms.Initial()
	if err != nil {
		return err
	}
	now := time.Now().Unix()

	for _, mod := range mods {
		root, err := ms.toolRoot(string(mod.ToolId))
		if err != nil {
			// Unset folder touch nothing.
			log.Printf("reconcile: skipping %s: %v", mod.ID, err)
			continue
		}

		for _, p := range mod.Packs {
			var paths packPaths
			var d observed
			var size int64

			if p.State == InstallReady { // other states have no RelPath yet
				if paths, err = ms.packPaths(mod, p); err != nil {
					log.Printf("reconcile: skipping %s/%s: %v", mod.ID, p.ID, err)
					continue
				}
				d = observe(paths)
				if d.archive {
					size = fileSize(paths.archive)
				}
			}

			fixed, remove := planPack(p, paths, d, size, now)

			if fixed != p {
				if _, err := ms.updatePack(mod.ID, p.ID, func(cur *Pack) error {
					*cur = fixed // fine at startup: nothing else is writing yet
					return nil
				}); err != nil {
					log.Printf("reconcile: save %s/%s: %v", mod.ID, p.ID, err)
					continue // never delete anything unless the corrected state was saved
				}
				log.Printf("reconcile: %s/%s removing %v", mod.ID, p.ID, remove)
			}

			for _, path := range remove {
				if err := removeInside(root, path); err != nil {
					log.Printf("reconcile: remove %s: %v", path, err)
					continue
				}
				log.Printf("reconcile: removed leftover %s", path)
			}
		}
	}
	return nil
}

func removeInside(root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing to delete %q outside %q", path, root)
	}
	return os.RemoveAll(path)
}
