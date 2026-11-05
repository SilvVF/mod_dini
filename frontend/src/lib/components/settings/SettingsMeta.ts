import type { Component } from "svelte";
import {
  Bell,
  Globe,
  LayoutGrid,
  Moon,
  Palette,
  Settings2,
  Sun,
  X,
  Zap,
} from "@lucide/svelte";
import { Category, Kind } from "#lib/bindings/changeme/models.ts";

// categories: { id: Category; label: string }[]
export const SettingsMeta = {
  categories: [
    { id: Category.CategoryGeneral, label: "General" },
    { id: Category.CategoryAppearance, label: "Appearance" },
    { id: Category.CategoryMods, label: "Mods" },
    { id: Category.CategoryInGame, label: "In game" },
    { id: Category.CategoryGameBanana, label: "GameBanana" },
    { id: Category.CategoryFolders, label: "Folders" },
    { id: Category.CategoryTools, label: "Tools" },
    { id: Category.CategoryAdvanced, label: "Advanced" },
    { id: Category.CategoryAbout, label: "About" },
  ],
  categoryLabel: (id: Category) => {
    return SettingsMeta.categories.find((c) => c.id === id)?.label ?? id;
  },
  /** Resolves Setting.icon / Option.icon names sent by the backend. Add entries as you add settings. */
  resolveIcon: (name?: string): Component => (name && icons[name]) || Settings2,
};

const icons: Record<string, Component> = {
  globe: Globe,
  x: X,
  "layout-grid": LayoutGrid,
  bell: Bell,
  palette: Palette,
  zap: Zap,
  "theme-dark": Moon,
  "theme-light": Sun,
};
