// Plain data shapes only. The sidebar never fetches or mutates anything itself:
// everything comes in through props and every action goes out through callback props.

import type { CharacterFolder, ModSaved, Pack, Preset } from "#lib/bindings/changeme/models.ts";

export interface Game {
  id: string;
  name: string;
  image: string;
}

export type Conflict = {
  name: string;
  folderName: string;
  mods: ModSaved[]; // mods with at least one enabled pack
  packCount: number; // total enabled packs across those mods
};

export interface GameStats {
  loaded: number;
  total: number;
  /** Pre-formatted, e.g. "1.3 GB" */
  sizeOnDisk: string;
  overlaps: number;
  updates: number;
}

/** Everything the sidebar needs to render the per-game configuration */
export interface GameConfig {
  game: Game;
  presets: Preset[];
  viewedPreset: Preset | undefined;
  /** Mods that belong to the active preset, in display order */
  mods: ModSaved[];
  stats: GameStats;
  launchVanilla: boolean;
}

export interface SidebarProps {
  config: GameConfig;
  /** When null/undefined the sidebar shows the game config, otherwise the mod config */
  selectedMod?: ModSaved | null;
  /** Folder options for the mod "Folder" select */
  folders?: string[];
  onselectmod?: (id: string) => void;
  onclosemod?: () => void;
  onunloadmod?: (id: string) => void;
  onopenfolder?: (mod: ModSaved) => void;
  onopenlink?: (mod: ModSaved) => void;
  onvanillachange?: (value: boolean) => void;
  onTogglePack: (mod: ModSaved, value: Pack) => void;
  onstart?: () => void;
  onSelectPreset: (id: number | undefined) => void;
}
