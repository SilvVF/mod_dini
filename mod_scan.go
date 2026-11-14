package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/ini.v1"
)

type hashComponent struct {
	Name          string       `json:"component_name"`
	DrawVB        string       `json:"draw_vb"`
	PositionVB    string       `json:"position_vb"`
	BlendVB       string       `json:"blend_vb"`
	TexcoordVB    string       `json:"texcoord_vb"`
	IB            string       `json:"ib"`
	TextureHashes [][][]string `json:"texture_hashes"` // per object: [slot, ext, hash]
}

const (
	weightIB       = 3.0
	weightPosition = 3.0
	weightBlend    = 2.0
	weightTexcoord = 2.0
	weightDraw     = 1.0
	weightTexture  = 0.5

	strongWeight = 2.0 // IB and vertex-buffer hashes identify a specific mesh
	minHashScore = 4.0
)

var hex8 = regexp.MustCompile(`^[0-9a-f]{8}$`)

// weights returns each distinct hash of the component with how identifying it is.
func (c hashComponent) weights() map[string]float64 {
	out := map[string]float64{}
	add := func(h string, w float64) {
		h = strings.ToLower(h)
		if hex8.MatchString(h) && w > out[h] {
			out[h] = w
		}
	}

	add(c.IB, weightIB)
	add(c.PositionVB, weightPosition)
	add(c.BlendVB, weightBlend)
	add(c.TexcoordVB, weightTexcoord)
	add(c.DrawVB, weightDraw)
	for _, object := range c.TextureHashes {
		for _, tex := range object {
			if len(tex) == 3 {
				add(tex[2], weightTexture)
			}
		}
	}
	return out
}

// key: HASH -> map{ owner: weight }, owner is "<tool>:<local character id>"
var HASH_BUCKET = []byte("hash_index")

// indexCharacter only ever adds. Mods made before a game patch still use the old hashes,
// so refreshing the JSON must not delete what an earlier version contributed.
func (ms *ModService) indexCharacter(owner string, comps []hashComponent) error {
	allHashesWeighted := map[string]float64{}
	for _, c := range comps {
		for hash, weight := range c.weights() {
			allHashesWeighted[hash] = max(allHashesWeighted[hash], weight)
		}
	}

	return ms.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(HASH_BUCKET)
		if err != nil {
			return err
		}
		for hash, weight := range allHashesWeighted {
			owners := map[string]float64{}
			if raw := bkt.Get([]byte(hash)); raw != nil {
				_ = json.Unmarshal(raw, &owners)
			}
			owners[owner] = max(owners[owner], weight)

			data, err := json.Marshal(owners)
			if err != nil {
				return err
			}
			if err := bkt.Put([]byte(hash), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ms *ModService) matchCharacter(toolID, sourceName string) (id string, ok bool) {
	cfg, err := ms.FolderConfig()
	if err != nil {
		return "", false
	}
	// matching is scoped to the tool: the same name can exist in two games
	for charID, characterFolder := range cfg[toolID].Characters.Characters {
		// should be the same check
		if characterFolder.Name == sourceName || charID == sourceName {
			return charID, true
		}
	}
	return "", false
}

// imports the character hashes from game resources and adds them to the
func (ms *ModService) importCharacterHashes(toolID, sourceName string, raw []byte) error {
	localCharId, ok := ms.matchCharacter(toolID, sourceName)
	if !ok {
		return fmt.Errorf("no %s character matches %q (add an alias)", toolID, sourceName)
	}

	var comps []hashComponent
	if err := json.Unmarshal(raw, &comps); err != nil {
		return fmt.Errorf("parse hashes for %q: %w", sourceName, err)
	}
	return ms.indexCharacter(toolID+":"+localCharId, comps)
}

// ---- detection ----

// returns tool and local character id whose known hashes best explain
// the mod's hashes (from iniHashes).
func (ms *ModService) detectCharacter(hashes map[string]struct{}) (toolID, charID string, ok bool) {
	type tally struct {
		score  float64
		strong bool
	}
	tallies := map[string]*tally{}

	_ = ms.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(HASH_BUCKET)
		if bkt == nil {
			return nil
		}
		for hash := range hashes {
			raw := bkt.Get([]byte(hash))
			if raw == nil {
				continue
			}
			var owners map[string]float64
			if json.Unmarshal(raw, &owners) != nil {
				continue
			}
			for owner, w := range owners {
				t := tallies[owner]
				if t == nil {
					t = &tally{}
					tallies[owner] = t
				}
				t.score += w / float64(len(owners)) // a hash shared by several characters counts for less
				t.strong = t.strong || w >= strongWeight
			}
		}
		return nil
	})

	var best string
	var bestScore, second float64
	for owner, t := range tallies {
		switch {
		case t.score > bestScore:
			best, second, bestScore = owner, bestScore, t.score
		case t.score > second:
			second = t.score
		}
	}
	if best == "" || !tallies[best].strong || bestScore < minHashScore || bestScore < 2*second {
		return "", "", false
	}

	toolID, charID, _ = strings.Cut(best, ":")
	return toolID, charID, true
}

func iniHashes(dir string) (map[string]struct{}, error) {
	files, err := FindIniFiles(dir)
	if err != nil {
		return nil, err
	}

	args := make([]interface{}, len(files))
	for i, value := range files {
		args[i] = value
	}

	cfg, err := ini.LoadSources(
		ini.LoadOptions{SkipUnrecognizableLines: true},
		[]byte{},
		args...,
	)

	hashes := make(map[string]struct{})

	for _, section := range cfg.Sections() {
		hashValue, err := section.GetKey("hash")
		if err != nil {
			continue
		}
		hashes[hashValue.String()] = struct{}{}
	}

	return hashes, nil
}

func (ms *ModService) detectAndStoreCharacter(mod ModSaved, dir string) {
	hashes, err := iniHashes(dir)
	if err != nil || len(hashes) == 0 {
		return
	}

	toolID, charID, ok := ms.detectCharacter(hashes)
	if !ok {
		return
	}
	if toolID != string(mod.ToolId) {
		// Don't assign a character from a different game. This is also your wrong-tool warning.
		log.Printf("mod %s is installed under %s but its hashes match %s/%s", mod.ID, mod.ToolId, toolID, charID)
		return
	}

	if _, err := ms.updateMod(mod.ID, func(m *ModSaved) error {
		if m.CharacterId == "" { // re-check: another pack may have set it already
			m.CharacterId = charID
			m.CharacterSource = "hash"
		}
		return nil
	}); err != nil {
		log.Printf("store detected character for %s: %v", mod.ID, err)
	}
}
