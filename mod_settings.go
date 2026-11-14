package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

func (ms *ModService) Settings() ([]Setting, error) {
	return ms.settings.All()
}

func (ms *ModService) OnSettingAction(key string, action string) error {
	return ms.settings.DoAction(key, action)
}

func (ms *ModService) UpdateSetting(key string, value json.RawMessage) error {
	_, err := ms.settings.SetValue(key, value)
	return err
}

func (ms *ModService) SetActiveTool(toolId ToolID) error {
	raw, err := maybeJSON(toolId)
	if err != nil {
		return err
	}

	_, err = ms.settings.SetValue("tool.active", raw)
	return err
}

func (ms *ModService) setModFolder(toolId ToolID, reset bool) error {
	settingsKey := fmt.Sprintf("%s.folder", toolId)
	if reset {
		path, err := GetSetting[string](ms.settings, "xxmi.folder")
		if err != nil || path == "" {
			return err
		}
		_, err = SetSettingValue(ms.settings, settingsKey, filepath.Join(path, "/mods"))
		return err
	} else {
		path, err := ms.app.Dialog.
			OpenFile().
			SetTitle("Select Folder for " + toolId).
			CanChooseFiles(false).
			CanChooseDirectories(true).
			PromptForSingleSelection()

		if err != nil || path == "" {
			return err
		}

		_, err = SetSettingValue(ms.settings, settingsKey, path)
		return err
	}
}

func (ms *ModService) DefaultSettings() []Setting {
	return []Setting{
		NewText(
			"zzmi.folder",
			CategoryFolders,
			"ZenlessZoneZero Mods Folder",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.zzmi", "Change", func() error {
				return ms.setModFolder("zzmi", false)
			}).
			WithActionHandler(ms.settings, "zzmi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("zzmi", true)
			}).
			MakeUneditable(),
		NewText(
			"gimi.folder",
			CategoryFolders,
			"Genshin Impact Mods Folder",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.gimi", "Change", func() error {
				return ms.setModFolder("gimi", false)
			}).
			WithActionHandler(ms.settings, "gimi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("gimi", true)
			}).
			MakeUneditable(),
		NewText(
			"wwmi.folder",
			CategoryFolders,
			"Wuthering Waves Mods Folder",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.wwmi", "Change", func() error {
				return ms.setModFolder("wwmi", false)
			}).
			WithActionHandler(ms.settings, "wwmi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("wwmi", true)
			}).
			MakeUneditable(),
		NewText(
			"srmi.folder",
			CategoryFolders,
			"Honkai Star rail Mods Folder",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.srmi", "Change", func() error {
				return ms.setModFolder("srmi", false)
			}).
			WithActionHandler(ms.settings, "srmi.xxmi", "Use XXMI's", func() error {
				return ms.setModFolder("srmi", true)
			}).
			MakeUneditable(),
		NewText(
			"xxmi.folder",
			CategoryFolders,
			"XXMI Launcher",
			"",
			"",
		).
			WithIcon("folder").
			WithActionHandler(ms.settings, "change.xxmi", "Change", func() error {
				return ms.setModFolder("xxmi", false)
			}).
			WithActionHandler(ms.settings, "xxmi.xxmi", "Rescan", func() error {
				return ms.rescanXXMI()
			}).
			MakeUneditable(),
		// ---- General ----
		NewDropdown(
			"general.language",
			CategoryGeneral,
			"Language",
			"The language of the app's menus and messages. It changes right away.",
			[]Option{
				{Value: "en", Label: "English"},
				{Value: "de", Label: "Deutsch"},
				{Value: "ja", Label: "日本語"},
			},
			"en",
		).
			WithSection("General").
			WithIcon("globe").
			WithKeywords("locale", "translation"),
		NewDropdown(
			"tool.active",
			CategoryGeneral,
			"Tool",
			"Set the active tool",
			[]Option{
				{Value: "zzmi", Label: "ZZMI"},
				{Value: "wwmi", Label: "WWMI"},
				{Value: "gimi", Label: "GIMI"},
				{Value: "srmi", Label: "SRMI"},
			},
			"zzmi",
		).
			WithSection("General").
			WithIcon("globe").
			WithKeywords("active", "tool"),
		NewToggle(
			"general.hide_unconfigured_games",
			CategoryGeneral,
			"Hide games that aren't set up",
			"Leaves games without an XXMI folder out of the game list. They come back when you turn this off.",
			false,
		).
			WithSection("General").
			WithIcon("layout-grid"),
		NewDropdown(
			"appearance.theme",
			CategoryAppearance,
			"Theme",
			"Fine-tune colors, cards and the background in Customize.",
			[]Option{
				{Value: "dark", Label: "Dark", Icon: "theme-dark"},
				{Value: "light", Label: "Light", Icon: "theme-light"},
			},
			"dark",
		).
			WithSection("Appearance").
			WithIcon("palette"),
	}
}
