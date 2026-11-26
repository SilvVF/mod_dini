import { Events } from "@wailsio/runtime";
import { getContext, onMount, setContext, untrack } from "svelte";
import {
  ModService,
  StorageState,
  type DataProgress,
  type ModSaved,
  type Pack,
  type Preset,
  type PresetItem,
  type Submission,
  type SubmissionFile,
} from "../bindings/changeme";
import { compact } from "#lib/ts-ext.ts";
import { appState } from "#lib/state/app.svelte.ts";
import type { ModGroup } from "./group-mods";

export interface SavedMods {
  mods: Record<string, ModSaved>;
  presets: Record<number, Preset>;
  dlProgress: Record<string, DataProgress>;
  add(sub: Submission, files: SubmissionFile[], packs: Pack[]): void;
  setAsEnabledPkgs(id: string, ids: string[]): void;
  enablePkg(modId: string, id: string): void;
  toggleAllPkgs(id: string): void;
  remove(modID: string): void;
  createPreset(name: string, toolId: string): void;
  removeFromPreset(presetId: number, modId: string): void;
  renamePreset(presetId: number, name: string): void;
  enablePreset(presetId: number): void;
}

const SAVED_MODS_KEY = Symbol("savedMods");

function createPresetRecord(data: Preset[]) {
  const presets: { [key: string]: Preset } = {};
  for (const item of data) {
    presets[item.id] = item;
  }
  return presets;
}

export function setSavedModsContext(
  initial: () => Record<string, ModSaved>,
  initialPresets: () => Preset[],
) {
  let context = $state<SavedMods>({
    mods: initial(),
    presets: createPresetRecord(initialPresets()),
    dlProgress: {},
    add(sub, files, packs) {
      ModService.Save(appState.activeTool, sub, files, packs).catch((e) => {
        console.log(e);
      });
    },
    remove(modID) {
      ModService.Remove(modID, true);
    },
    toggleAllPkgs(id) {
      const mod = this.mods[id];
      if (!mod) return;
      const anyEnabled = mod.packages?.some((p) => p.enabled);
      if (anyEnabled) {
        this.setAsEnabledPkgs(mod.id, []);
      } else {
        const packs = mod.packages?.map((p) => p.id) ?? [];
        this.setAsEnabledPkgs(mod.id, packs);
      }
    },
    setAsEnabledPkgs(id, ids) {
      ModService.EnablePacks(id, ids).catch((e) => {
        console.log(e);
      });
    },
    enablePkg(modId, id) {
      const mod = this.mods[modId];
      if (!mod || !mod.packages) return;

      const enabled = new Set(
        mod.packages.filter((p) => p.enabled).map((p) => p.id),
      );
      if (enabled.has(id)) {
        enabled.delete(id);
      } else {
        enabled.add(id);
      }
      this.setAsEnabledPkgs(mod.id, Array.from(enabled));
    },
    removeFromPreset(presetId: number, modId: string) {
      const preset = this.presets[presetId];
      if (!preset) return;

      const updated = preset.mods?.filter((item) => item.modId !== modId) ?? [];
      ModService.UpdatePreset(presetId, preset.name, updated);
    },
    renamePreset(presetId: number, name: string) {
      const preset = this.presets[presetId];
      if (!preset) return;
      ModService.UpdatePreset(preset.id, name, preset.mods);
    },
    createPreset(name: string, toolId: string) {
      const enabled = Object.values(this.mods)
        .filter((m) => m.toolId === toolId)
        .filter((m) =>
          m.packages?.some((p) => p.storage === StorageState.StorageActive),
        );
      const presetItems: PresetItem[] = [];
      for (const mod of enabled) {
        presetItems.push({
          modId: mod.id,
          packIds:
            mod.packages
              ?.filter((p) => p.storage === StorageState.StorageActive)
              ?.map((p) => p.id) ?? [],
        });
      }

      ModService.CreatePreset(toolId, name, presetItems);
    },
    enablePreset(presetId: number) {
      ModService.EnablePreset(presetId).catch(console.log);
    },
  });

  $effect.pre(() => {
    const incoming = initial();
    untrack(() => {
      context.mods = incoming;
    });
  });

  $effect.pre(() => {
    const incoming = initialPresets();

    untrack(() => {
      context.presets = createPresetRecord(incoming);
    });
  });

  setContext(SAVED_MODS_KEY, context);

  onMount(() => {
    const unsubscribe = Events.On("group-update", ({ data }) => {
      if (!data) return;
      context.presets[data.id] = data;
    });

    return () => unsubscribe();
  });

  onMount(() => {
    const unsubscribe = Events.On("group-removed", ({ data }) => {
      if (!data) return;
      delete context.presets[data];
    });

    return () => unsubscribe();
  });

  onMount(() => {
    const unsubscribe = Events.On("mod-removed", ({ data }) => {
      if (!data) return;
      delete context.mods[data];
    });

    return () => unsubscribe();
  });

  onMount(() => {
    const unsubscribe = Events.On("mods-loaded", ({ data }) => {
      if (!data) return;
      console.log("loaded", data);
      context.mods = compact(data);
    });

    return () => unsubscribe();
  });

  onMount(() => {
    const unsubscribe = Events.On("mod-update", ({ data }) => {
      console.log("mod-update", data);
      if (!data) return;
      context.mods[data.id] = data;
    });

    return () => unsubscribe();
  });

  onMount(() => {
    const unsubscribe = Events.On("mod-download-progress", ({ data }) => {
      if (!data) return;
      context.dlProgress[data.id] = data;
    });

    return () => unsubscribe();
  });

  return context;
}

export function getSavedModsContext() {
  return getContext<SavedMods>(SAVED_MODS_KEY);
}
