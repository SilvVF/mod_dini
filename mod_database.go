package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

var (
	DL_BUCKET                   = []byte("Downlods")
	GROUP_BUCKET                = []byte("MOD_GROUP")
	ErrDownloadBucketNotCreated = errors.New("download bucket has not been created yet")
)

type ModGroup struct {
	ID   uint64
	Name string
	Mods []ModGroupItem
}

type ModGroupItem struct {
	Id    string
	Packs []string
}

func (ms *ModService) GetGroups() []ModGroup {
	ret := []ModGroup{}

	ms.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(GROUP_BUCKET)
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			var group ModGroup
			err := json.Unmarshal(v, &group)
			if err != nil {
				return nil
			}

			ret = append(ret, group)
			return nil
		})
	})

	return ret
}

func (ms *ModService) updateGroup(id uint64, update func(group *ModGroup) error) (ModGroup, error) {
	var group ModGroup
	var err error

	err = ms.db.Update(func(tx *bolt.Tx) error {

		bkt := tx.Bucket(GROUP_BUCKET)
		if bkt == nil {
			return errors.New("group does not exist")
		}

		key := []byte(strconv.FormatUint(id, 10))
		value := bkt.Get(key)
		if value == nil {
			return errors.New("group does not exist")
		}

		err = json.Unmarshal(value, &group)
		if err != nil {
			return err
		}

		err = update(&group)
		if err != nil {
			return err
		}

		data, err := json.Marshal(group)
		if err != nil {
			return err
		}

		return bkt.Put(key, data)
	})

	if err != nil {
		return group, err
	}

	ms.app.Event.Emit("group-update", group)

	return group, nil
}

func (ms *ModService) createGroup(name string) (ModGroup, error) {
	group := ModGroup{
		Name: name,
		Mods: make([]ModGroupItem, 0),
	}

	err := ms.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(GROUP_BUCKET)
		if err != nil {
			return err
		}

		nextID, err := bkt.NextSequence()
		if err != nil {
			return err
		}
		group.ID = nextID

		data, err := json.Marshal(group)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(strconv.FormatUint(nextID, 10)), data)
	})

	if err != nil {
		return group, err
	}

	ms.app.Event.Emit("group-update", group)
	return group, nil
}

// broadcasts update event when no err
func (ms *ModService) updateMod(modId string, update func(mod *ModSaved) error) (ModSaved, error) {

	var mod ModSaved
	var err error

	err = ms.db.Update(func(tx *bolt.Tx) error {
		mod, err = updateModTx(tx, modId, update)
		return err
	})

	if err != nil {
		return mod, err
	}

	ms.app.Event.Emit("mod-update", mod)

	return mod, nil
}

func getModTx(tx *bolt.Tx, modId string) (ModSaved, error) {
	var mod ModSaved

	bucket := tx.Bucket(DL_BUCKET)
	data := bucket.Get([]byte(modId))

	if err := json.Unmarshal(data, &mod); err != nil {
		return mod, err
	}

	return mod, nil
}

func updateModTx(tx *bolt.Tx, modId string, update func(mod *ModSaved) error) (ModSaved, error) {
	var mod ModSaved

	bucket := tx.Bucket(DL_BUCKET)
	data := bucket.Get([]byte(modId))

	if err := json.Unmarshal(data, &mod); err != nil {
		return mod, err
	}

	if err := update(&mod); err != nil {
		return mod, err
	}

	b, err := json.Marshal(mod)
	if err != nil {
		return mod, err
	}

	err = bucket.Put([]byte(modId), b)

	return mod, err
}

// func updatePackTx(tx *bolt.Tx, modID, pkgID string, update func(*Pack) error) (Pack, error) {
// 	m, err := updateModTx(tx, modID, func(mod *ModSaved) error {
// 		pack, idx, ok := mod.GetPackage(pkgID)
// 		if !ok {
// 			return fmt.Errorf("pack %s not found", pkgID)
// 		}

// 		if err := update(&pack); err != nil {
// 			return err
// 		}
// 		mod.Packs[idx] = pack

// 		return nil
// 	})

// 	if err != nil {
// 		log.Printf("[mod] update package failed mod=%s pkg=%s: %v", modID, pkgID, err)
// 		return Pack{}, err
// 	}

// 	p, _, _ := m.GetPackage(pkgID)

// 	return p, nil
// }

// broadcasts update when no err
func (ms *ModService) updatePack(modID, pkgID string, update func(*Pack) error) (Pack, error) {

	m, err := ms.updateMod(modID, func(mod *ModSaved) error {
		pack, idx, ok := mod.GetPackage(pkgID)
		if !ok {
			return fmt.Errorf("pack %s not found", pkgID)
		}

		if err := update(&pack); err != nil {
			return err
		}
		mod.Packs[idx] = pack

		return nil
	})

	if err != nil {
		log.Printf("[mod] update package failed mod=%s pkg=%s: %v", modID, pkgID, err)
		return Pack{}, err
	}

	p, _, _ := m.GetPackage(pkgID)

	return p, nil
}
