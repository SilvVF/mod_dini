package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

type ModService struct {
	app      *application.App
	gbApi    GbApi
	fs       *FolderStore
	db       *bolt.DB
	settings *SettingsStore
}

var (
	DL_BUCKET                   = []byte("Downlods")
	ErrDownloadBucketNotCreated = errors.New("download bucket has not been created yet")
)

func (m *ModService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	err := m.settings.Sync(m.DefaultSettings())
	if err != nil {
		panic(err)
	}

	err = m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DL_BUCKET)
		return err
	})
	if err != nil {
		panic(err)
	}

	m.fs.CreateIfNotExists(
		"zzz",
		"Zenless Zone Zero",
	)

	_, err = m.fs.Update("zzz", func(folder *RootFolder) error {
		folder.Characters = ZenlessCharacters
		return nil
	})

	if err != nil {
		panic(err)
	}

	changes := m.settings.Changes()
	go func() {
		for c := range changes {
			log.Printf("changed setting: %v", c)
			m.app.Event.Emit("setting-updated", c)
		}
	}()

	return nil
}

func (m *ModService) FolderConfig() (RootFolder, error) {
	f, err := m.fs.Get("zzz")

	if err != nil {
		log.Println("folder get ", err)
	}
	return f, err
}

func (m *ModService) Settings() ([]Setting, error) {
	return m.settings.All()
}

func (m *ModService) OnSettingAction(key string, action string) error {
	return m.settings.DoAction(key, action)
}

func (m *ModService) UpdateSetting(key string, change SettingChange) error {
	_, err := m.settings.SetValue(key, change)
	return err
}

func (m *ModService) Initial() (map[string]ModSaved, error) {

	res := map[string]ModSaved{}

	err := m.db.View(func(tx *bolt.Tx) error {
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

func (m *ModService) SetModFolder(key string) error {

	path, err := m.app.Dialog.
		OpenFile().
		SetTitle("Select Folder").
		CanChooseFiles(false).
		CanChooseDirectories(true).
		PromptForSingleSelection()

	if err != nil || path == "" {
		return err
	}

	_, err = SetValueOf(m.settings, key, KindText, path)
	return err
}

func (m *ModService) Save(sub Submission, files []SubmissionFile, tree ModFileTree) error {

	mod := ModSaved{
		Submission: sub,
		Files:      files,
		Tree:       tree,
	}

	err := m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(DL_BUCKET)

		mod.ID = sub.Provider + ":" + sub.Slug
		data, err := json.Marshal(mod)

		if err != nil {
			return err
		}

		err = b.Put([]byte(mod.ID), data)
		if err != nil {
			return err
		}

		m.app.Event.Emit("mod-update", mod)

		return nil
	})
	if err != nil {
		return err
	}

	err = m.DownloadMod(mod, "zzmi")
	log.Println(err)

	return err
}

func (m *ModService) Page(id int) (Submission, []SubmissionFile, error) {
	res, err := m.gbApi.ModPage(id)
	if err != nil {
		return Submission{}, nil, err
	}

	s, files := MapModPageResponse(res)

	return s, files, nil
}

type ListParams struct {
	Provider string `json:"provider"`
	Cat      string `json:"cat"`
	Page     int    `json:"page"`
}

type ListPage struct {
	Submissions []Submission `json:"submissions"`
	HasNext     bool         `json:"hasNext"`
	Next        int          `json:"next"`
	Prev        int          `json:"prev"`
}

func (m *ModService) List(p ListParams) (ListPage, error) {
	q := NewCategoryQuery()
	q.Page = max(p.Page, 1)
	q.ID = p.Cat
	q.Sort = MostLiked
	q.PerPage = 50

	log.Println("list results ", p)
	lp := ListPage{}

	if p.Provider == "gamebanana" {
		res, err := m.gbApi.CategoryContent(q)
		if err != nil {
			return ListPage{}, err
		}

		submissions := make([]Submission, len(res.ARecords))
		for i, record := range res.ARecords {
			submissions[i] = MapCategoryRecord(record)
		}

		lp.Submissions = submissions
		lp.Next = min(q.Page+1, res.AMetadata.NRecordCount/res.AMetadata.NPerpage)
		lp.Prev = max(q.Page-1, 1)
		lp.HasNext = lp.Next != q.Page

		return lp, nil
	} else {
		return lp, nil
	}
}

func (m *ModService) DefaultSettings() []Setting {
	return []Setting{
		NewText("zzmi.folder", CategoryFolders, "ZenlessZoneZero Mods Folder", "", "").
			WithIcon("folder").
			WithActionHandler(m.settings, "change.zzmi", "Change", func() error {
				return m.SetModFolder("zzmi.folder")
			}).
			WithAction("zzmi.xxmi", "Use XXMI's").
			MakeUneditable(),
		NewDropdown(
			"general.language",
			CategoryGeneral,
			"Language",
			"The language of the app's menus and messages. It changes right away.",
			[]Option{{Value: "en", Label: "English"}, {Value: "de", Label: "Deutsch"},
				{Value: "ja", Label: "日本語"}},
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
			[]Option{{Value: "ask", Label: "Ask"}, {Value: "tray", Label: "Tray"},
				{Value: "quit", Label: "Quit"}},
			"ask",
		).WithSection("General").WithIcon("x"),

		NewToggle(
			"general.hide_unconfigured_games",
			CategoryGeneral,
			"Hide games that aren't set up",
			"Leaves games without an XXMI folder out of the game list. They come back when you turn this off.",
			false,
		).WithSection("General").WithIcon("layout-grid"),

		NewMultiSelect("general.notifications", CategoryGeneral,
			"Notifications", "Choose which events show a notification.",
			[]Option{{Value: "updates", Label: "Mod updates"}, {Value: "downloads", Label: "Downloads"}, {Value: "errors", Label: "Errors"}},
			[]string{"updates", "errors"}).WithSection("General").WithIcon("bell"),

		NewSegmented("appearance.interface", CategoryAppearance,
			"Interface", "The new look, or the Classic layout the app had before.",
			[]Option{{Value: "modern", Label: "Modern"}, {Value: "classic", Label: "Classic"}},
			"modern").WithSection("Appearance").WithIcon("layout-grid"),

		NewDropdown("appearance.theme", CategoryAppearance,
			"Theme", "Fine-tune colors, cards and the background in Customize.",
			[]Option{{Value: "dark", Label: "Dark", Icon: "theme-dark"}, {Value: "light", Label: "Light", Icon: "theme-light"}},
			"dark").WithSection("Appearance").WithIcon("palette"),

		NewToggle("appearance.performance_mode", CategoryAppearance,
			"Performance mode", "No background picture, blur, transparency or heavy shadows. Lightest on slower PCs.",
			false).WithSection("Appearance").WithIcon("zap"),
	}
}
