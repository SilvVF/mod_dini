package main

import (
	"testing"
)

func TestGetModHashes(t *testing.T) {

	_, err := getHashes("leotorrez", "ZZ-Model-Importer-Assets", "PlayerCharacterData")

	t.Error(err)
}
