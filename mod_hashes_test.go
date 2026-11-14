package main

import (
	"log"
	"testing"
)

func TestGetModHashes(t *testing.T) {

	hashes, err := getHashes("leotorrez", "ZZ-Model-Importer-Assets", "PlayerCharacterData")

	log.Println(hashes)

	t.Error(err)
}
