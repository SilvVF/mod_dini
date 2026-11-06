package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

type ModService struct {
	app       *application.App
	providers ProviderRegistry
	gbApi     GbApi
	fs        *FolderStore
	db        *bolt.DB
	settings  *SettingsStore
}

func (ms *ModService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
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
		"zzz",
		"Zenless Zone Zero",
	)

	_, err = ms.fs.Update("zzz", func(folder *RootFolder) error {
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

	return nil
}

func (ms *ModService) FolderConfig() (RootFolder, error) {
	f, err := ms.fs.Get("zzz")

	if err != nil {
		log.Println("folder get ", err)
	}
	return f, err
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

func (ms *ModService) setModFolder(key string, reset bool) error {
	if reset {
		path, err := GetSetting[string](ms.settings, "xxmi.folder")
		if err != nil || path == "" {
			return err
		}
		_, err = SetValueOf(ms.settings, key, filepath.Join(path, "WWMI/mods"))
		return err
	} else {
		path, err := ms.app.Dialog.
			OpenFile().
			SetTitle("Select Folder").
			CanChooseFiles(false).
			CanChooseDirectories(true).
			PromptForSingleSelection()

		if err != nil || path == "" {
			return err
		}

		_, err = SetValueOf(ms.settings, key, path)
		return err
	}
}

func (ms *ModService) EnablePacks(modId string, ids []string) error {
	log.Println("settings ", modId, " packs: ", ids)
	_, err := ms.updateMod(modId, func(mod *ModSaved) error {
		for i, p := range mod.Packs {
			mod.Packs[i].Enabled = slices.Contains(ids, p.ID)
		}
		return nil
	})

	return err
}

func (ms *ModService) Save(sub Submission, files []SubmissionFile, selected []Pack) error {
	log.Println("saving ModSaved", selected)

	modID := sub.Provider + ":" + sub.Slug
	outputDir, err := GetSetting[string](ms.settings, "zzmi.folder")

	if err != nil || outputDir == "" {
		log.Println("output dir not canceling save ", modID)
		return err
	}

	modDir := filepath.Join(outputDir, sub.Provider, sub.Slug)
	if err := os.RemoveAll(modDir); err != nil {
		return err
	}

	mod := ModSaved{
		ID:              modID,
		Submission:      sub,
		SubmissionFiles: files,
		Packs:           []Pack{},
	}

	for i, s := range selected {
		mod.Packs = append(mod.Packs, Pack{
			ID:       s.ID,
			ModID:    modID,
			FileID:   s.FileID,
			Name:     s.Name,
			State:    InstallQueued,
			Priority: i,
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

	return prov.Page(context.Background(), pageId)
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
				return ms.setModFolder("zzmi.folder", false)
			}).
			WithActionHandler(ms.settings, "zzmi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("zzmi.folder", true)
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
				return ms.setModFolder("xxmi.folder", false)
			}).
			WithAction("zzmi.xxmi", "Use XXMI's").
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

		NewSegmented(
			"general.on_close",
			CategoryGeneral,
			"When the window is closed",
			"Ask each time, keep the app running in the tray, or quit it completely.",
			[]Option{
				{Value: "ask", Label: "Ask"},
				{Value: "tray", Label: "Tray"},
				{Value: "quit", Label: "Quit"},
			},
			"ask",
		).
			WithSection("General").
			WithIcon("x"),

		NewToggle(
			"general.hide_unconfigured_games",
			CategoryGeneral,
			"Hide games that aren't set up",
			"Leaves games without an XXMI folder out of the game list. They come back when you turn this off.",
			false,
		).
			WithSection("General").
			WithIcon("layout-grid"),

		NewMultiSelect(
			"general.notifications",
			CategoryGeneral,
			"Notifications",
			"Choose which events show a notification.",
			[]Option{
				{Value: "updates", Label: "Mod updates"},
				{Value: "downloads", Label: "Downloads"},
				{Value: "errors", Label: "Errors"},
			},
			[]string{"updates", "errors"},
		).
			WithSection("General").
			WithIcon("bell"),

		// ---- Appearance ----
		NewSegmented(
			"appearance.interface",
			CategoryAppearance,
			"Interface",
			"The new look, or the Classic layout the app had before.",
			[]Option{
				{Value: "modern", Label: "Modern"},
				{Value: "classic", Label: "Classic"},
			},
			"modern",
		).
			WithSection("Appearance").
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

		NewToggle(
			"appearance.performance_mode",
			CategoryAppearance,
			"Performance mode",
			"No background picture, blur, transparency or heavy shadows. Lightest on slower PCs.",
			false,
		).
			WithSection("Appearance").
			WithIcon("zap"),
	}
}
