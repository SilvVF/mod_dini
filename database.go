package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

var (
	DL_BUCKET                   = []byte("Downlods")
	ErrDownloadBucketNotCreated = errors.New("download bucket has not been created yet")
)

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

func updatePackTx(tx *bolt.Tx, modID, pkgID string, update func(*Pack) error) (Pack, error) {
	m, err := updateModTx(tx, modID, func(mod *ModSaved) error {
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
