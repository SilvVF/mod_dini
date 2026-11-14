package main

import (
	"fmt"
	"slices"
)

type ToolID = string

const (
	ZZMI ToolID = "zzmi"
	SRMI ToolID = "srmi"
	GIMI ToolID = "gimi"
	WWMI ToolID = "wwmi"
)

type toolDef struct {
	ID            string
	Name          string // game name, used for labels and folder config
	Label         string // short name shown in dropdowns
	setCharacters func(*RootFolder)
}

// The single source of truth for supported tools. Order here is the order in the UI.
var supportedTools = []toolDef{
	{ID: ZZMI, Name: "Zenless Zone Zero", Label: "ZZMI", setCharacters: func(f *RootFolder) { f.Characters = ZenlessCharacters }},
	{ID: GIMI, Name: "Genshin Impact", Label: "GIMI", setCharacters: func(f *RootFolder) { f.Characters = GenshinCharacters }},
	{ID: WWMI, Name: "Wuthering Waves", Label: "WWMI", setCharacters: func(f *RootFolder) { f.Characters = WutheringWavesCharacters }},
	{ID: SRMI, Name: "Honkai Star Rail", Label: "SRMI", setCharacters: func(f *RootFolder) { f.Characters = StarRailCharacters }},
}

func toolByID(id ToolID) (toolDef, bool) {
	i := slices.IndexFunc(supportedTools, func(t toolDef) bool { return t.ID == id })
	if i < 0 {
		return toolDef{}, false
	}
	return supportedTools[i], true
}

func toolOptions() []Option {
	opts := make([]Option, len(supportedTools))
	for i, t := range supportedTools {
		opts[i] = Option{Value: t.ID, Label: t.Label}
	}
	return opts
}

// seedTools makes sure every tool has a folder record with its character list.
func (ms *ModService) seedTools() error {
	for _, t := range supportedTools {
		ms.fs.CreateIfNotExists(t.ID, t.Name)

		if _, err := ms.fs.Update(t.ID, func(f *RootFolder) error {
			t.setCharacters(f)
			return nil
		}); err != nil {
			return fmt.Errorf("seed tool %s: %w", t.ID, err)
		}
	}
	return nil
}

func (ms *ModService) FolderConfig() (map[ToolID]RootFolder, error) {
	out := make(map[string]RootFolder, len(supportedTools))
	for _, t := range supportedTools {
		f, err := ms.fs.Get(t.ID)
		if err != nil {
			return nil, fmt.Errorf("folder config %s: %w", t.ID, err)
		}
		out[t.ID] = f
	}
	return out, nil
}
