package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"go.etcd.io/bbolt"
	"gopkg.in/ini.v1"
)

var KEYBIND_BUCKET = []byte("keybinds_bucket")
var ORIGINAL_KEY = []byte("original")

var ErrKeybindBucketNotCreated = errors.New("keybinds bucket was not created yet")

type KeybindStore struct {
	db *bbolt.DB
	ms *ModService
}

type ModKeyBind struct {
	PackId        string   `json:"packId"`
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

type ModKeybindUpdate struct {
	ModId   string `json:"modId"`
	PackId  string `json:"packId"`
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

func (ks *KeybindStore) GetForMod(modId string) (map[string]ModKeyBindPreset, error) {
	binds := map[string]ModKeyBindPreset{}
	err := ks.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket(KEYBIND_BUCKET)
		modBucket := root.Bucket([]byte(modId))
		if modBucket == nil {
			return errors.New("no keybinds for mod")
		}

		return modBucket.ForEach(func(presetId, v []byte) error {
			var preset ModKeyBindPreset
			if err := json.Unmarshal(v, &preset); err != nil {
				return err
			}
			binds[string(presetId)] = preset
			return nil
		})
	})

	return binds, err
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

func (ks *KeybindStore) UpdateKey(u ModKeybindUpdate) error {

	ks.ms.sweeper.CancelAndWait(u.ModId)

	unlock := ks.ms.lockMod(u.ModId)
	defer unlock()

	mod, err := ks.ms.getMod(u.ModId)
	if err != nil {
		return err
	}

	pack, _, ok := mod.GetPackage(u.PackId)
	if !ok {
		return errors.New("pack not found")
	}

	if pack.Storage != StorageActive && pack.Storage != StorageWarm {
		return errors.New("pack not warm or active")
	}

	paths, err := ks.ms.packPaths(mod, pack)
	if err != nil {
		return err
	}
	path := expectedPath(paths, pack.Storage)
	iniFiles, err := FindIniFiles(path)
	if err != nil {
		return err
	}

	errs := []error{}
	for _, f := range iniFiles {
		errs = append(errs, setIniKey(f, u.Section, u.Key, u.Value))
	}

	anyCompleted := slices.Contains(errs, nil)
	if !anyCompleted {
		return errors.Join(errs...)
	}

	return nil
}

func setIniKey(path, section, key, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	inSection := false
	keyRe := regexp.MustCompile(`(?i)^(\s*` + regexp.QuoteMeta(key) + `\s*=\s*).*$`)

	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			inSection = strings.EqualFold(t[1:len(t)-1], section)
			continue
		}
		if inSection && keyRe.MatchString(line) {
			lines[i] = keyRe.ReplaceAllString(line, "${1}"+value)
			break
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func (ks *KeybindStore) GetCurrent(modId string) (ModKeyBindPreset, error) {

	ks.ms.sweeper.CancelAndWait(modId)

	unlock := ks.ms.lockMod(modId)
	defer unlock()

	mod, err := ks.ms.getMod(modId)
	if err != nil {
		return ModKeyBindPreset{}, err
	}

	return ks.getCurrentKeybinds(mod)
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

		preset, err := ks.getCurrentKeybinds(mod)
		if err != nil {
			return err
		}
		preset.Name = string(ORIGINAL_KEY)
		preset.ID = string(ORIGINAL_KEY)

		b, err := json.Marshal(preset)
		if err != nil {
			return err
		}

		return bucket.Put(ORIGINAL_KEY, b)
	})
}

func (ks *KeybindStore) getCurrentKeybinds(mod ModSaved) (ModKeyBindPreset, error) {
	modBinds := []ModKeyBind{}
	for _, pack := range mod.Packs {
		paths, err := ks.ms.packPaths(mod, pack)
		if err != nil {
			return ModKeyBindPreset{}, err
		}

		if pack.Storage != StorageActive && pack.Storage != StorageWarm {
			return ModKeyBindPreset{}, errors.New("pack not warm or active")
		}

		currPath := expectedPath(paths, pack.Storage)

		iniFiles, err := FindIniFiles(currPath)
		binds, err := ParseKeybind(iniFiles)

		for i := range binds {
			binds[i].PackId = pack.ID
		}

		if err != nil {
			return ModKeyBindPreset{}, err
		}
		modBinds = append(modBinds, binds...)
	}

	preset := ModKeyBindPreset{
		ModId:     mod.ID,
		ToolId:    mod.ToolId,
		Values:    modBinds,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return preset, nil
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

		if strings.HasPrefix(strings.ToLower(entry.Name()), "disabled") {
			return nil
		}

		if strings.EqualFold(filepath.Ext(entry.Name()), ".ini") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func ParseKeybind(
	files []string,
) ([]ModKeyBind, error) {

	args := make([]any, len(files))
	for i, value := range files {
		args[i] = value
	}

	keybinds := []ModKeyBind{}

	cfg, err := ini.LoadSources(
		ini.LoadOptions{
			SkipUnrecognizableLines:    true,
			AllowDuplicateShadowValues: true,
			AllowNonUniqueSections:     true,
		},
		[]byte{},
		args...,
	)
	if err != nil {
		return keybinds, err
	}

	constants := cfg.Section("Constants")
	options := map[string]struct{}{}
	for _, k := range constants.Keys() {
		if after, ok := strings.CutPrefix(k.Name(), "global persist"); ok {
			option := strings.Trim(after, " ")
			options[option] = struct{}{}
		}
	}

	for _, s := range cfg.Sections() {

		if !strings.HasPrefix(s.Name(), "Key") {
			continue
		}

		persistVar := ""
		persistValuesStr := ""

		for _, row := range s.Keys() {
			if _, ok := options[row.Name()]; ok {
				persistVar = row.Name()
				persistValuesStr = row.Value()
				break
			}
		}

		persistValues := strings.Split(persistValuesStr, ",")
		for i, v := range persistValues {
			persistValues[i] = strings.TrimSpace(v)
		}

		key := ""
		if s.HasKey("key") {
			key = s.Key("key").Value()
		} else {
			key = s.Key("Key").Value()
		}

		keybinds = append(keybinds, ModKeyBind{
			Section:       s.Name(),
			Key:           key,
			ActionType:    s.Key("type").Value(),
			PersistVar:    persistVar,
			PersistValues: persistValues,
		})
	}

	return keybinds, nil
}
