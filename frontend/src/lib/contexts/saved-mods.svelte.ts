import { Events } from "@wailsio/runtime";
import { getContext, onMount, setContext, untrack } from "svelte";
import {
  ModService,
  type DataProgress,
  type ModFileTree,
  type ModSaved,
  type Submission,
  type SubmissionFile,
} from "../bindings/changeme";
import { compact } from "#lib/ts-ext.ts";

export interface SavedMods {
  mods: Record<string, ModSaved>;
  dlProgress: Record<string, DataProgress>;
  add(sub: Submission, files: SubmissionFile[], tree: ModFileTree): void;
  remove(id: number): void;
}

const SAVED_MODS_KEY = Symbol("savedMods");

export function setSavedModsContext(
  initial: () => Record<string, ModSaved>,
) {

  let context = $state<SavedMods>({
    mods: initial(),
    dlProgress: {},
    add(sub, files, tree) {
      ModService.Save(sub, files, tree).catch((e) => {
        console.log(e);
      });
    },
    remove(id) {},
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
