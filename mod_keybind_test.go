package main

import "testing"

const TEST_INI_PATH = "C:\\Users\\david\\SkinMods\\ZZMI\\Mods\\jane_doe_-_hazeker_redux\\JaneDoe-HazekerRedux\\JaneDoeBH.ini"
const TEST_INI_PATH2 = "C:\\Users\\david\\SkinMods\\ZZMI\\Mods\\Phoenix Reffaella\\DISABLED_Phoenix Lewd - phoenix 2\\Main\\Const.ini"
const TEST_INI_PATH3 = "C:\\Users\\david\\SkinMods\\ZZMI\\Mods\\Phoenix Reffaella\\Phoenix - Maid Bikini\\resources\\Phoenix.ini"

func TestConfigFileGeneration(t *testing.T) {
	ParseKeyind([]string{TEST_INI_PATH, TEST_INI_PATH2, TEST_INI_PATH3})

	t.Error(nil)
}
