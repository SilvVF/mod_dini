// Plain data shapes only. The sidebar never fetches or mutates anything itself:
// everything comes in through props and every action goes out through callback props.

import type { ModSaved } from "#lib/bindings/changeme/models.ts";

export interface Game {
	id: string;
	name: string;
	image: string;
}

export interface Mod {
	id: string;
	name: string;
	/** Character / category the mod belongs to (e.g. "Others") */
	category: string;
	author: string;
	link: string;
	note: string;
	thumbnail: string;
	/** Preview image URLs (first one is used as the list thumbnail) */
	previews: string[];
	has3D: boolean;
	enabled: boolean;
	favorite: boolean;
	/** Name of the folder the mod lives in */
	folder: string;
}

export interface Preset {
	id: string;
	name: string;
	/** Ordered ids of the mods in this preset */
	modIds: string[];
}

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
	preset: Preset;
	/** Mods that belong to the active preset, in display order */
	mods: ModSaved[];
	stats: GameStats;
	launchVanilla: boolean;
}

/** Fields of a mod that can be edited from the sidebar */
export type ModPatch = Partial<Pick<Mod, 'author' | 'link' | 'note' | 'folder' | 'enabled' | 'favorite' | 'name'>>;

export interface SidebarProps {
	config: GameConfig;
	enabled: boolean,
	/** When null/undefined the sidebar shows the game config, otherwise the mod config */
	selectedMod?: ModSaved | null;
	/** Folder options for the mod "Folder" select */
	folders?: string[];
	onselectmod?: (id: string) => void;
	onclosemod?: () => void;
	onunloadmod?: (id: string) => void;
	onmodchange?: (id: string, patch: ModPatch) => void;
	onopenfolder?: (id: string) => void;
	onopenlink?: (id: string) => void;
	onvanillachange?: (value: boolean) => void;
	onstart?: () => void;
}
