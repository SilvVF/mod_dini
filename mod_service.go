package main

import (
	"changeme/mod3d"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

var ErrPackNotCold = errors.New("pack is not cold")
var ErrModBusy = errors.New("this mod is being installed or changed")

type ModService struct {
	app            *application.App
	keybindsWindow *application.WebviewWindow
	providers      ProviderRegistry
	ks             *KeybindStore
	fs             *FolderStore
	db             *bolt.DB
	sweeper        *Sweeper
	settings       *SettingsStore
	modLocks       sync.Map
	pw             *ProcessWatcher
}

func NewModService(app *application.App, db *bolt.DB) *ModService {
	settingsStore, err := NewSettingsStore(db)
	if err != nil {
		panic(err)
	}

	ms := &ModService{
		settings:       settingsStore,
		keybindsWindow: nil,
		app:            app,
		db:             db,
		fs:             &FolderStore{db: db},
		providers: ProviderRegistry{
			"gamebanana": NewGameBanana(&GbApi{}),
			"huihui":     &HuiHuiProvider{},
		},
		pw: NewProcessWatcher(time.Second),
	}

	ms.ks = &KeybindStore{
		db: ms.db,
		ms: ms,
	}

	return ms
}

func (ms *ModService) lockMod(modID string) (unlock func()) {
	v, _ := ms.modLocks.LoadOrStore(modID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (ms *ModService) tryLockMod(modID string) (unlock func(), ok bool) {
	v, _ := ms.modLocks.LoadOrStore(modID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	if !mu.TryLock() {
		return nil, false
	}
	return mu.Unlock, true
}

func (ms *ModService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ms.sweeper = NewSweeper(ms, ctx)

	if err := ms.settings.Sync(ms.DefaultSettings()); err != nil {
		panic(err)
	}

	if err := ms.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DL_BUCKET)
		return err
	}); err != nil {
		panic(err)
	}

	if err := ms.seedTools(); err != nil {
		panic(err)
	}

	// todo(should be fast enough to call blocking possibly move to goroutine + event when finished if startup is slow)
	if err := ms.reconcile(); err != nil {
		log.Printf("reconcile: %v", err)
	}

	changes := ms.settings.Changes()
	go func() {
		for c := range changes {
			log.Printf("changed setting: %v", c)
			ms.app.Event.Emit("setting-updated", c)
		}
	}()

	go ms.sweeper.Run()
	processEvents := ms.pw.Watch(ctx, "GenshinImpact.exe", "StarRail.exe", "ZenlessZoneZero.exe", "Client-Win64-Shipping.exe")

	go func() {
		for {
			event, ok := <-processEvents
			if !ok {
				break
			}
			ms.app.Event.Emit("process-event", event)
		}
	}()

	return nil
}

func (ms *ModService) RunningProcesses() map[string]ProcessEvent {
	return ms.pw.Current("GenshinImpact.exe", "StarRail.exe", "ZenlessZoneZero.exe", "Client-Win64-Shipping.exe")
}

type PackPreview struct {
	Payload *mod3d.ViewerPayload `json:"payload"`
	Mod     ModSaved             `json:"mod"`
	Pack    Pack                 `json:"pack"`
}

func (ms *ModService) PreviewPack(modID, packID string) (PackPreview, error) {

	ms.sweeper.CancelAndWait(modID)

	unlock := ms.lockMod(modID)
	defer unlock()

	mod, err := ms.getMod(modID)
	if err != nil {
		return PackPreview{}, err
	}

	pack, _, ok := mod.GetPackage(packID)
	if !ok || !slices.Contains([]StorageState{StorageWarm, StorageActive}, pack.Storage) {
		return PackPreview{}, err
	}

	paths, err := ms.packPaths(mod, pack)
	if err != nil {
		return PackPreview{}, err
	}

	baseModelMods := ms.getBaseModelMods(mod)

	payload, err := mod3d.ParseMod(
		expectedPath(paths, pack.Storage),
		nil,
		baseModelMods,
		mod3d.QualityHigh,
	)

	return PackPreview{
		Payload: payload,
		Mod:     mod,
		Pack:    pack,
	}, err
}

func (ms *ModService) StartGame(tool ToolID) error {
	exeSubPath := "Resources\\Bin\\XXMI Launcher.exe"
	dir, err := GetSetting[string](ms.settings, "xxmi.folder")
	if err != nil {
		return err
	}

	exeFilePath := filepath.Join(dir, exeSubPath)
	cmd := exec.Command(
		exeFilePath,
		"--nogui",
		"--xxmi",
		tool,
	)

	return cmd.Run()
}

func (ms *ModService) getBaseModelMods(mod ModSaved) []mod3d.BaseModelMod {
	baseModelMods := []mod3d.BaseModelMod{}
	mods, err := ms.Initial()
	if err != nil {
		return baseModelMods
	}

	for _, m := range mods {

		if m.ID == mod.ID {
			continue
		}

		sameCharacter := (m.CharacterId == mod.CharacterId && m.CharacterId != "")
		sameCategory := (m.Submission.Category == mod.Submission.Category && m.Submission.Category != "")
		if sameCategory || sameCharacter {
			for _, p := range m.Packs {
				paths, err := ms.packPaths(m, p)
				if err != nil {
					continue
				}
				baseModelMods = append(baseModelMods, mod3d.BaseModelMod{
					Provider: m.Submission.Provider,
					ModID:    mod.ID,
					FileID:   p.FileID,
					Name:     p.Name,
					Path:     expectedPath(paths, p.Storage),
				})
			}
		}
	}

	log.Println("base model mods: ", baseModelMods)

	return baseModelMods
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

func (ms *ModService) GetKeyBinds() (map[string]map[string]ModKeyBindPreset, error) {
	return ms.ks.GetAll()
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

func (ms *ModService) toolRoot(toolID ToolID) (string, error) {
	root, err := GetSetting[string](ms.settings, toolID+".folder")
	if err != nil {
		return "", err
	}
	if root == "" {
		return "", fmt.Errorf("the %s folder is not set", toolID)
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return "", fmt.Errorf("the %s folder %q is not available", toolID, root)
	}
	return root, nil
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
	part     string
}

func (ms *ModService) packPaths(mod ModSaved, p Pack) (packPaths, error) {
	root, err := ms.toolRoot(mod.ToolId)
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
		part:     filepath.Join(dir, base+".part"),
	}, nil
}

func expectedPath(paths packPaths, state StorageState) string {
	switch state {
	case StorageActive:
		return paths.active
	case StorageCold:
		return paths.archive
	case StorageCompressing:
		return paths.temp
	case StorageNone:
		return ""
	case StorageRestoring:
		return ""
	case StorageWarm:
		return paths.disabled
	default:
		panic(fmt.Sprintf("unexpected main.StorageState: %#v", state))
	}
}

func (ms *ModService) Remove(modId string, deleteKeybinds bool) error {

	ms.sweeper.CancelAndWait(modId)

	unlock := ms.lockMod(modId)
	defer unlock()

	mod, err := ms.getMod(modId)
	if err != nil {
		return err
	}

	toolDir, err := ms.toolRoot(mod.ToolId)
	if err != nil {
		return err
	}

	err = ms.db.Update(func(tx *bolt.Tx) error {
		modDir := filepath.Join(toolDir, mod.Submission.Provider, mod.Submission.SubmissionID)
		if err := os.RemoveAll(modDir); err != nil {
			return err
		}

		err = tx.Bucket(DL_BUCKET).Delete([]byte(mod.ID))
		if err != nil {
			return err
		}

		if deleteKeybinds {
			keybinds := tx.Bucket(KEYBIND_BUCKET)
			if keybinds != nil {
				keybinds.DeleteBucket([]byte(mod.ID))
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	ms.app.Event.Emit("mod-removed", modId)

	return nil
}

func (ms *ModService) Save(toolId ToolID, sub Submission, files []SubmissionFile, selected []Pack) error {
	if _, ok := toolByID(toolId); !ok {
		return fmt.Errorf("unknown tool %q", toolId)
	}

	outputDir, err := ms.toolRoot(toolId)
	if err != nil {
		return err
	}

	modID := sub.Provider + ":" + sub.SubmissionID
	log.Printf("save mod=%s tool=%s packs=%d", modID, toolId, len(selected))

	ms.sweeper.CancelAndWait(modID)

	// need to do this because files on disk are edited and rely on previous state for same mod.
	unlock, ok := ms.tryLockMod(modID)
	if !ok {
		return ErrModBusy
	}
	defer unlock()

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

	tool, local, hero, ok := ms.resolveModMeta(sub.Provider, sub.Category)
	if ok {
		mod.ToolId = tool
		mod.CharacterId = local
		mod.Submission.Hero = &hero
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

	return ms.downloadMod(ms.app.Context(), mod)
}

func (ms *ModService) resolveModMeta(providerID, category string) (toolID, local, hero string, ok bool) {
	if category == "" {
		return "", "", "", false
	}

	cfg, err := ms.FolderConfig()
	if err != nil {
		log.Printf("resolve character: %v", err)
		return "", "", "", false
	}

	for _, root := range cfg {
		for charID, folder := range root.Characters.Characters {
			if folder.Sources[providerID] == category {
				return root.ID, charID, folder.Name, true
			}
		}
	}
	return "", "", "", false
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

	if toolId, _, character, ok := ms.resolveModMeta(providerId, p.Submission.Category); ok {
		p.ResolvedToolId = &toolId
		p.Submission.Hero = &character
	}

	log.Println("resolved", p.Submission.Hero, p.ResolvedToolId)

	return p, nil
}

func (ms *ModService) Subcategories(provider string, category string) (ProviderSubcategoryList, error) {
	prov, err := ms.providers.Get(provider)
	if err != nil {
		return ProviderSubcategoryList{}, err
	}

	return prov.Subcategories(context.Background(), category)
}

func (ms *ModService) List(p ListParams) (SubmissionList, error) {
	log.Println("list params", p)

	prov, err := ms.providers.Get(p.Provider)
	if err != nil {
		return SubmissionList{}, err
	}

	return prov.List(context.Background(), p)
}
