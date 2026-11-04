package main

import (
	"encoding/json"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type RootFolder struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Config     FolderConfig         `json:"config"`
	Subfolders map[string]Subfolder `json:"subfolders"`
	Characters CharacterCategory    `json:"characters"`
	Order      []string             `json:"order"`
}

type Subfolder struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Custom  bool              `json:"custom"`
	Config  FolderConfig      `json:"config"`
	Sources map[string]string `json:"sources"`
}

type CharacterCategory struct {
	Config     FolderConfig               `json:"config"`
	Characters map[string]CharacterFolder `json:"characters"`
}

type CharacterFolder struct {
	Name    string            `json:"name"`
	Sources map[string]string `json:"sources"`
	Icon    string            `json:"icon"`
}

type FolderConfig struct {
	Visible   bool       `json:"visible"`
	Pinned    bool       `json:"pinned"`
	Collapsed bool       `json:"collapsed"`
	Sort      SortConfig `json:"sort"`
}

type SortConfig struct {
	By        SortBy `json:"by"`
	Ascending bool   `json:"ascending"`
}

type SortBy string

const (
	SortByName SortBy = "name"
)

var (
	ErrFolderNotFound = errors.New("root folder not found")
	FOLDER_BUCKET     = []byte("root_folders")
)

type FolderStore struct {
	db *bolt.DB
}

func (s *FolderStore) Get(id string) (RootFolder, error) {
	var folder RootFolder

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FOLDER_BUCKET)
		if bucket == nil {
			return ErrFolderNotFound
		}

		data := bucket.Get([]byte(id))
		if data == nil {
			return ErrFolderNotFound
		}

		return json.Unmarshal(data, &folder)
	})

	return folder, err
}

func (s *FolderStore) Save(folder RootFolder) error {
	if folder.ID == "" {
		return errors.New("root folder ID cannot be empty")
	}

	data, err := json.Marshal(folder)
	if err != nil {
		return fmt.Errorf("marshal root folder: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FOLDER_BUCKET)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(folder.ID), data)
	})
}

func (s *FolderStore) Delete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FOLDER_BUCKET)

		if bucket.Get([]byte(id)) == nil {
			return ErrFolderNotFound
		}

		return bucket.Delete([]byte(id))
	})
}

func (s *FolderStore) CreateIfNotExists(
	id string,
	name string,
) (RootFolder, error) {
	var folder RootFolder

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(FOLDER_BUCKET)
		if err != nil {
			return err
		}

		data := bucket.Get([]byte(id))
		if data == nil {
			folder = DefaultRootFolder(id, name)
			b, err := json.Marshal(folder)

			if err != nil {
				return err
			}
			return bucket.Put([]byte(id), b)
		} else {
			err = json.Unmarshal(data, &folder)
			return err
		}
	})

	return folder, err
}

func (s *FolderStore) Update(
	id string,
	update func(*RootFolder) error,
) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(FOLDER_BUCKET)

		data := bucket.Get([]byte(id))
		if data == nil {
			return ErrFolderNotFound
		}

		var folder RootFolder

		if err := json.Unmarshal(data, &folder); err != nil {
			return fmt.Errorf("unmarshal root folder: %w", err)
		}

		if err := update(&folder); err != nil {
			return err
		}

		updated, err := json.Marshal(folder)
		if err != nil {
			return fmt.Errorf("marshal root folder: %w", err)
		}

		return bucket.Put([]byte(id), updated)
	})
}

func (s *FolderStore) UpdateSubfolder(
	rootID string,
	subfolderID string,
	update func(*Subfolder) error,
) error {
	return s.Update(rootID, func(root *RootFolder) error {
		subfolder, exists := root.Subfolders[subfolderID]
		if !exists {
			return fmt.Errorf(
				"subfolder %q does not exist",
				subfolderID,
			)
		}

		if err := update(&subfolder); err != nil {
			return err
		}

		root.Subfolders[subfolderID] = subfolder

		return nil
	})
}

func DefaultRootFolder(id, name string) RootFolder {
	return RootFolder{
		ID:   id,
		Name: name,
		Config: FolderConfig{
			Visible: true,
			Sort: SortConfig{
				By:        SortByName,
				Ascending: true,
			},
		},
		Characters: CharacterCategory{
			Config:     FolderConfig{},
			Characters: map[string]CharacterFolder{},
		},
		Subfolders: map[string]Subfolder{
			"weapons": {
				ID:      "weapons",
				Name:    "Weapons",
				Custom:  false,
				Sources: map[string]string{"local": "weapons"},
				Config: FolderConfig{
					Visible: true,
					Sort: SortConfig{
						By:        SortByName,
						Ascending: true,
					},
				},
			},
			"others": {
				ID:      "others",
				Name:    "Others",
				Custom:  false,
				Sources: map[string]string{"local": "others", "gamebanna": "29874"},
				Config: FolderConfig{
					Visible: true,
					Sort: SortConfig{
						By:        SortByName,
						Ascending: true,
					},
				},
			},
		},
		Order: []string{
			"weapons",
			"others",
		},
	}
}
