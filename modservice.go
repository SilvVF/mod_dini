package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

var ErrPackNotCold = errors.New("pack is not cold")

type ModService struct {
	app            *application.App
	keybindsWindow *application.WebviewWindow
	providers      ProviderRegistry
	gbApi          GbApi
	ks             *KeybindStore
	fs             *FolderStore
	db             *bolt.DB
	sweeper        *Sweeper
	settings       *SettingsStore
	modLocks       sync.Map
}

func (ms *ModService) lockMod(modID string) (unlock func()) {
	v, _ := ms.modLocks.LoadOrStore(modID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (ms *ModService) OpenKeybindsWindow() {

	if ms.keybindsWindow != nil {
		ms.keybindsWindow.Close()
	}

	ms.keybindsWindow = ms.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:          "keybinds-window",
		Width:         420,
		Height:        600,
		DisableResize: true,
		AlwaysOnTop:   true,
		URL:           "/keybinds",
		Windows:       application.WindowsWindow{},
	})
}

func (ms *ModService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ms.providers = ProviderRegistry{
		"gamebanana": NewGameBanana(&GbApi{}),
		"huihui":     &HuiHuiProvider{},
	}
	ms.ks = &KeybindStore{
		db: ms.db,
		ms: ms,
	}

	err := ms.settings.Sync(ms.DefaultSettings())
	if err != nil {
		panic(err)
	}

	err = ms.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DL_BUCKET)
		return err
	})
	if err != nil {
		panic(err)
	}

	ms.fs.CreateIfNotExists(
		"zzmi",
		"Zenless Zone Zero",
	)

	_, err = ms.fs.Update("zzmi", func(folder *RootFolder) error {
		folder.Characters = ZenlessCharacters
		return nil
	})

	if err != nil {
		panic(err)
	}

	changes := ms.settings.Changes()
	go func() {
		for c := range changes {
			log.Printf("changed setting: %v", c)
			ms.app.Event.Emit("setting-updated", c)
		}
	}()

	ms.sweeper = NewSweeper(ms)
	go ms.sweeper.Run()

	// res, err := ms.Initial()
	// if err != nil {
	// 	return err
	// }
	// for _, m := range res {
	// 	ms.ks.SaveDefault(m.ID)
	// }

	return nil
}

func (ms *ModService) FolderConfig(toolId string) (RootFolder, error) {
	f, err := ms.fs.Get(toolId)

	if err != nil {
		log.Println("folder get ", err)
	}
	return f, err
}

func (ms *ModService) GetKeyBinds() (map[string]map[string]ModKeyBindPreset, error) {
	return ms.ks.GetAll()
}

func (ms *ModService) Settings() ([]Setting, error) {
	return ms.settings.All()
}

func (ms *ModService) OnSettingAction(key string, action string) error {
	return ms.settings.DoAction(key, action)
}

func (ms *ModService) UpdateSetting(key string, value json.RawMessage) error {
	_, err := ms.settings.SetValue(key, value)
	return err
}

func (ms *ModService) Initial() (map[string]ModSaved, error) {

	res := map[string]ModSaved{}

	err := ms.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(DL_BUCKET)
		if b == nil {
			return ErrDownloadBucketNotCreated
		}
		return b.ForEach(func(k, v []byte) error {
			var mod ModSaved
			err := json.Unmarshal(v, &mod)
			if err != nil {
				return err
			}

			res[string(k)] = mod

			return nil
		})
	})

	return res, err
}

func (ms *ModService) rescanXXMI() error {
	// todo(get path for xxmi)
	path := ""
	if path != "" {
		_, err := SetSettingValue(ms.settings, "xxmi.folder", path)
		return err
	}

	return nil
}

func (ms *ModService) setModFolder(toolId string, reset bool) error {
	if reset {
		path, err := GetSetting[string](ms.settings, "xxmi.folder")
		if err != nil || path == "" {
			return err
		}
		_, err = SetSettingValue(ms.settings, toolId, filepath.Join(path, toolId+"/mods"))
		return err
	} else {
		path, err := ms.app.Dialog.
			OpenFile().
			SetTitle("Select Folder for" + toolId).
			CanChooseFiles(false).
			CanChooseDirectories(true).
			PromptForSingleSelection()

		if err != nil || path == "" {
			return err
		}

		_, err = SetSettingValue(ms.settings, toolId+".folder", path)
		return err
	}
}

func (ms *ModService) getMod(id string) (ModSaved, error) {
	var mod ModSaved
	err := ms.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(DL_BUCKET)
		data := bucket.Get([]byte(id))

		return json.Unmarshal(data, &mod)
	})
	return mod, err
}

func (ms *ModService) EnablePacks(modID string, ids []string) error {
	ms.sweeper.CancelAndWait(modID)

	unlock := ms.lockMod(modID)
	defer unlock()

	mod, err := ms.getMod(modID)
	if err != nil {
		return err
	}

	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}

	var errs []error
	for _, p := range mod.Packs {
		var err error
		switch {
		case want[p.ID] && p.Storage == StorageActive:
			// nothing to do
		case want[p.ID]:
			err = ms.activate(mod, p) // Warm: rename. Cold: Restoring, extract to temp, rename, Active, delete archive
		case p.Storage == StorageActive:
			err = ms.deactivate(mod, p) // rename, then state=Warm with DisabledAt=now
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("pack %s: %w", p.ID, err))
		}
	}
	return errors.Join(errs...)
}

func (ms *ModService) deactivate(mod ModSaved, p Pack) error {
	if p.Storage != StorageActive {
		return fmt.Errorf("cannot deactivate pack %s in storage state %d", p.ID, p.Storage)
	}

	paths, err := ms.packPaths(mod, p)
	if err != nil {
		return err
	}

	if _, err := os.Stat(paths.disabled); err == nil {
		return fmt.Errorf("%s already exists", paths.disabled)
	}

	if err := os.Rename(paths.active, paths.disabled); err != nil {
		return fmt.Errorf("disable %s (is the game or tool using it?): %w", p.ID, err)
	}

	if _, err := ms.updatePack(mod.ID, p.ID, func(cur *Pack) error {
		cur.Enabled = false
		cur.Storage = StorageWarm
		cur.DisabledAt = time.Now().Unix()
		return nil
	}); err != nil {
		if rerr := os.Rename(paths.disabled, paths.active); rerr != nil {
			log.Printf("deactivate: rollback rename failed for %s: %v", p.ID, rerr)
		}
		return err
	}

	return nil
}

func (ms *ModService) activate(mod ModSaved, p Pack) error {
	paths, err := ms.packPaths(mod, p)
	if err != nil {
		return err
	}

	switch p.Storage {
	case StorageWarm:
		if err := os.Rename(paths.disabled, paths.active); err != nil {
			return err
		}
	case StorageCold:
		if err := ms.restore(mod, p, paths); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot activate pack %s in storage state %d", p.ID, p.Storage)
	}

	if _, err := ms.updatePack(mod.ID, p.ID, func(cur *Pack) error {
		cur.Enabled = true
		cur.Storage = StorageActive
		cur.DisabledAt = 0
		cur.ArchiveFormat = ""
		cur.ArchiveSize = 0
		return nil
	}); err != nil {
		return err
	}

	if p.Storage == StorageCold {
		if err := os.Remove(paths.archive); err != nil {
			log.Printf("activate: could not remove archive %s: %v", paths.archive, err)
		}
	}
	return nil
}

// restore extracts a cold pack into a temp folder and renames it into place.
// On any failure the archive is untouched and the pack goes back to Cold.
func (ms *ModService) restore(mod ModSaved, p Pack, paths packPaths) (err error) {
	if _, err := ms.updatePack(mod.ID, p.ID, func(cur *Pack) error {
		if cur.Storage != StorageCold {
			return ErrPackNotCold
		}
		cur.Storage = StorageRestoring
		return nil
	}); err != nil {
		return err
	}

	defer func() {
		os.RemoveAll(paths.temp)
		if err != nil {
			ms.setStorage(mod.ID, p.ID, StorageCold)
		}
	}()

	if err = os.RemoveAll(paths.temp); err != nil { // leftover from a crash
		return err
	}
	if err = ms.extractArchive(mod.ID, p.ID, paths.archive, paths.temp); err != nil {
		return err
	}

	return os.Rename(paths.temp, paths.active)
}

func (ms *ModService) extractArchive(modID, packID, archivePath, dst string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return ArchiveExtract(ExtractOptions{
		ctx:       context.Background(),
		archive:   f,
		dst:       dst,
		overwrite: true,
		onProgress: func(progress, total int64) {
			ms.app.Event.Emit("mod-download-progress", DataProgress{
				ID:        modID,
				PackageID: packID,
				Kind:      "storage",
				State:     int(StorageRestoring),
				Total:     total,
				Progress:  progress,
			})
		},
	})
}

func (ms *ModService) setStorage(modID, packID string, state StorageState) {
	if _, err := ms.updatePack(modID, packID, func(cur *Pack) error {
		cur.Storage = state
		return nil
	}); err != nil {
		log.Printf("set storage %d for pack %s: %v", state, packID, err)
	}
}

type packPaths struct {
	active   string // <root>/<dir>/<name>                the tool loads this
	disabled string // <root>/<dir>/DISABLED_<name>        warm form
	temp     string // <root>/<dir>/DISABLED_<name>.tmp    extraction target, ignored by the tool
	archive  string // <root>/<dir>/<name><ArchiveFormat>   cold form
}

func (ms *ModService) packPaths(mod ModSaved, p Pack) (packPaths, error) {
	root, err := GetSetting[string](ms.settings, mod.ToolId+".folder")
	if err != nil {
		return packPaths{}, err
	}

	dir := filepath.Join(root, filepath.Dir(p.Manifest.RelPath))
	base := filepath.Base(p.Manifest.RelPath)

	return packPaths{
		active:   filepath.Join(dir, base),
		disabled: filepath.Join(dir, "DISABLED_"+base),
		temp:     filepath.Join(dir, "DISABLED_"+base+".tmp"),
		archive:  filepath.Join(dir, base+p.ArchiveFormat),
	}, nil
}

func (ms *ModService) Save(sub Submission, files []SubmissionFile, selected []Pack) error {
	log.Println("saving ModSaved", selected)

	//todo(pass tool id dynamically)
	toolId := "zzmi"
	modID := sub.Provider + ":" + sub.SubmissionID
	outputDir, err := GetSetting[string](ms.settings, toolId+".folder")

	if err != nil || outputDir == "" {
		log.Println("output dir not canceling save ", modID)
		return err
	}

	modDir := filepath.Join(outputDir, sub.Provider, sub.SubmissionID)
	if err := os.RemoveAll(modDir); err != nil {
		return err
	}

	mod := ModSaved{
		ID:              modID,
		ToolId:          toolId,
		Submission:      sub,
		SubmissionFiles: files,
		Packs:           []Pack{},
	}

	for _, s := range selected {
		mod.Packs = append(mod.Packs, Pack{
			ID:      s.ID,
			ModID:   modID,
			FileID:  s.FileID,
			Name:    s.Name,
			State:   InstallQueued,
			Storage: StorageNone,
		})
	}

	data, err := json.Marshal(mod)
	if err != nil {
		return err
	}

	err = ms.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(DL_BUCKET).Put([]byte(modID), data)
	})
	if err != nil {
		return err
	}

	ms.app.Event.Emit("mod-update", mod)

	go ms.DownloadMod(mod)

	return nil
}

func (ms *ModService) Page(providerId, pageId string) (SubmissionPage, error) {
	prov, err := ms.providers.Get(providerId)
	if err != nil {
		return SubmissionPage{}, err
	}

	p, err := prov.Page(context.Background(), pageId)

	if err != nil {
		return p, err
	}

	// resolve the local name for the category of this page
	f, err := ms.fs.Get("zzmi")
	for _, character := range f.Characters.Characters {
		categoryId := character.Sources[providerId]
		if categoryId == p.Submission.Category {
			p.Submission.Hero = &character.Name
			break
		}
	}

	return p, nil
}

func (ms *ModService) List(p ListParams) (SubmissionList, error) {
	log.Println("list params", p)

	prov, err := ms.providers.Get(p.Provider)
	if err != nil {
		return SubmissionList{}, err
	}

	return prov.List(context.Background(), p)
}

func (ms *ModService) DefaultSettings() []Setting {
	return []Setting{
		NewText(
			"zzmi.folder",
			CategoryFolders,
			"ZenlessZoneZero Mods Folder",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.zzmi", "Change", func() error {
				return ms.setModFolder("zzmi", false)
			}).
			WithActionHandler(ms.settings, "zzmi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("zzmi", true)
			}).
			MakeUneditable(),
		NewText(
			"xxmi.folder",
			CategoryFolders,
			"XXMI Launcher",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.xxmi", "Change", func() error {
				return ms.setModFolder("xxmi", false)
			}).
			WithActionHandler(ms.settings, "xxmi.xxmi", "Rescan", func() error {
				return ms.rescanXXMI()
			}).
			MakeUneditable(),
		// ---- General ----
		NewDropdown(
			"general.language",
			CategoryGeneral,
			"Language",
			"The language of the app's menus and messages. It changes right away.",
			[]Option{
				{Value: "en", Label: "English"},
				{Value: "de", Label: "Deutsch"},
				{Value: "ja", Label: "日本語"},
			},
			"en",
		).
			WithSection("General").
			WithIcon("globe").
			WithKeywords("locale", "translation"),
		NewToggle(
			"general.hide_unconfigured_games",
			CategoryGeneral,
			"Hide games that aren't set up",
			"Leaves games without an XXMI folder out of the game list. They come back when you turn this off.",
			false,
		).
			WithSection("General").
			WithIcon("layout-grid"),
		NewDropdown(
			"appearance.theme",
			CategoryAppearance,
			"Theme",
			"Fine-tune colors, cards and the background in Customize.",
			[]Option{
				{Value: "dark", Label: "Dark", Icon: "theme-dark"},
				{Value: "light", Label: "Light", Icon: "theme-light"},
			},
			"dark",
		).
			WithSection("Appearance").
			WithIcon("palette"),
	}
}
