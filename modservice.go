package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	bolt "go.etcd.io/bbolt"
)

type ModService struct {
	app   *application.App
	gbApi GbApi
	fs    *FolderStore
	db    *bolt.DB
}

var DL_BUCKET = []byte("Downlods")

func (m *ModService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {

	err := m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DL_BUCKET)
		return err
	})

	if err != nil {
		panic(err)
	}
	m.fs = &FolderStore{db: m.db}

	m.fs.CreateIfNotExists(
		"zzz",
		"Zenless Zone Zero",
	)

	err = m.fs.Update("zzz", func(folder *RootFolder) error {
		folder.Characters = ZenlessCharacters

		return nil
	})

	if err != nil {
		panic(err)
	}

	return nil
}

func (m *ModService) FolderConfig() (RootFolder, error) {
	f, err := m.fs.Get("zzz")

	if err != nil {
		log.Println("folder get ", err)
	}
	return f, err
}

func (m *ModService) Settings() (map[string]ModFolder, error) {
	res := map[string]ModFolder{}

	return res, nil
}

func (m *ModService) Initial() (map[string]ModSaved, error) {

	res := map[string]ModSaved{}

	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(DL_BUCKET)
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

	if err != nil {
		return nil, err
	}

	log.Println(res)

	return res, nil
}

func (m *ModService) SetModFolder(id string) error {

	path, err := m.app.Dialog.OpenFile().
		SetTitle("Select Folder").
		CanChooseFiles(false).
		CanChooseDirectories(true).
		PromptForSingleSelection()

	if err != nil || path == "" {
		return err
	}
	//todo
	return nil
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
