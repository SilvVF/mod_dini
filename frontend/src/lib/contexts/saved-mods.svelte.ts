import { Events } from "@wailsio/runtime";
import { getContext, onMount, setContext, untrack } from "svelte";
import {
  ModService,
  type DataProgress,
  type ModSaved,
  type Pack,
  type Submission,
  type SubmissionFile,
} from "../bindings/changeme";
import { compact } from "#lib/ts-ext.ts";

export interface SavedMods {
  mods: Record<string, ModSaved>;
  dlProgress: Record<string, DataProgress>;
  add(sub: Submission, files: SubmissionFile[], packs: Pack[]): void;
  setAsEnabledPkgs(id: string, ids: string[]): void;
  enablePkg(modId: string, id: string): void;
  toggleAllPkgs(id: string): void;
  remove(id: number): void;
}

const SAVED_MODS_KEY = Symbol("savedMods");

export function setSavedModsContext(initial: () => Record<string, ModSaved>) {
  let context = $state<SavedMods>({
    mods: initial(),
    dlProgress: {},
    add(sub, files, packs) {
      ModService.Save(sub, files, packs).catch((e) => {
        console.log(e);
      });
    },
    remove(id) {},
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
  });

  $effect.pre(() => {
    const incoming = initial();
    untrack(() => {
      context.mods = incoming;
    });
  });

  setContext(SAVED_MODS_KEY, context);

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
