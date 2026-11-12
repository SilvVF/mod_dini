package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"go.etcd.io/bbolt"
	"gopkg.in/ini.v1"
)

type ModKeyBind struct {
	Section       string   `json:"section"`
	Key           string   `json:"key"`
	ActionType    string   `json:"actionType"`
	PersistVar    string   `json:"persistVar"`
	PersistValue  string   `json:"persistValue"`
	PersistValues []string `json:"persistValues"`
}

type ModKeyBindPreset struct {
	ID        string       `json:"id"`
	ModId     string       `json:"modId"`
	ToolId    string       `json:"toolId"`
	Name      string       `json:"name"`
	Values    []ModKeyBind `json:"values"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

var KEYBIND_BUCKET = []byte("keybinds_bucket")
var ErrKeybindBucketNotCreated = errors.New("keybinds bucket was not created yet")

var ORIGINAL_KEY = []byte("original")

const TEST_MOD_PATH = "C:\\Users\\david\\SkinMods\\ZZMI\\Mods\\jane_doe_-_hazeker_redux"

type KeybindStore struct {
	db *bbolt.DB
	ms *ModService
}

func (ks *KeybindStore) GetAll() (map[string]map[string]ModKeyBindPreset, error) {
	binds := map[string]map[string]ModKeyBindPreset{}
	err := ks.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket(KEYBIND_BUCKET)
		return root.ForEachBucket(func(modId []byte) error {
			modBucket := root.Bucket(modId)

			binds[string(modId)] = map[string]ModKeyBindPreset{}

			return modBucket.ForEach(func(presetId, v []byte) error {
				var preset ModKeyBindPreset
				if err := json.Unmarshal(v, &preset); err != nil {
					return err
				}
				binds[string(modId)][string(presetId)] = preset
				return nil
			})
		})
	})

	return binds, err
}

func (ms *ModService) ToolDir(tool string) (string, error) {
	switch tool {
	case "zzmi":
		return GetSetting[string](ms.settings, tool+".folder")
	default:
		return "", fmt.Errorf("invalid tool %s", tool)
	}
}

func (ks *KeybindStore) GetCurrent(mod ModSaved) (ModKeyBindPreset, error) {
	return ModKeyBindPreset{}, nil
}

func (ks *KeybindStore) SaveDefault(modId string) error {
	return ks.db.Update(func(tx *bbolt.Tx) error {

		mod, err := getModTx(tx, modId)
		if err != nil {
			return err
		}

		root, err := tx.CreateBucketIfNotExists(KEYBIND_BUCKET)
		if err != nil {
			return ErrKeybindBucketNotCreated
		}

		bucket, err := root.CreateBucketIfNotExists([]byte(mod.ID))
		if err != nil {
			return err
		}

		modBinds := []ModKeyBind{}
		for _, pack := range mod.Packs {
			paths, err := ks.ms.packPaths(mod, pack)
			if err != nil {
				return err
			}

			currPath := ""
			if pack.Storage == StorageActive {
				currPath = paths.active
			} else if pack.Storage == StorageWarm {
				currPath = paths.disabled
			}

			if currPath == "" {
				return errors.New("unable to set deafault keybinds from archive")
			}

			iniFiles, err := FindIniFiles(currPath)
			binds, err := ParseKeyind(iniFiles)
			if err != nil {
				return err
			}
			modBinds = append(modBinds, binds...)
		}

		preset := ModKeyBindPreset{
			ID:        string(ORIGINAL_KEY),
			ModId:     mod.ID,
			ToolId:    mod.ToolId,
			Name:      string(ORIGINAL_KEY),
			Values:    modBinds,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		b, err := json.Marshal(preset)

		return bucket.Put(ORIGINAL_KEY, b)
	})
}

func FindIniFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(
		path string,
		entry fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".ini") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func ParseKeyind(
	files []string,
) ([]ModKeyBind, error) {

	args := make([]interface{}, len(files))
	for i, value := range files {
		args[i] = value
	}

	keybinds := []ModKeyBind{}

	cfg, err := ini.LoadSources(
		ini.LoadOptions{SkipUnrecognizableLines: true},
		[]byte{},
		args...,
	)
	if err != nil {
		return keybinds, err
	}

	constants := cfg.Section("Constants")
	options := map[string]struct{}{}
	for _, k := range constants.Keys() {
		if strings.HasPrefix(k.Name(), "global persist") {
			option := strings.Trim(strings.TrimPrefix(k.Name(), "global persist"), " ")
			options[option] = struct{}{}
		}
	}

	for _, s := range cfg.Sections() {

		if !strings.HasPrefix(s.Name(), "Key") {
			continue
		}

		persistVar := ""
		persistValuesStr := ""

		for _, c := range s.Keys() {
			if _, ok := options[c.Name()]; ok {
				persistVar = c.Name()
				persistValuesStr = c.Value()
			}
		}

		persistValues := strings.Split(persistValuesStr, ",")
		for i, v := range persistValues {
			persistValues[i] = strings.TrimSpace(v)
		}

		keybinds = append(keybinds, ModKeyBind{
			Section:       s.Name(),
			Key:           s.Key("key").Value(),
			ActionType:    s.Key("type").Value(),
			PersistVar:    persistVar,
			PersistValues: persistValues,
		})
	}

	return keybinds, nil
}
