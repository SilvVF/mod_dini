import { Events } from "@wailsio/runtime";
import { getContext, onMount, setContext } from "svelte";
import {
  ModService,
  type DataProgress,
  type ModFileTree,
  type ModSaved,
  type Submission,
  type SubmissionFile,
} from "../bindings/changeme";

export interface SavedMods {
  mods: Record<string, ModSaved>;
  dlProgress: Record<string, DataProgress>;
  add(sub: Submission, files: SubmissionFile[], tree: ModFileTree): void;
  remove(id: number): void;
}

const modsFromData = (
  data: { [x: `${number}`]: ModSaved | undefined } | undefined | null,
) =>
  data
    ? Object.fromEntries(
        Object.entries(data).filter(
          (entry): entry is [string, ModSaved] => entry[1] != null,
        ),
      )
    : {};

const SAVED_MODS_KEY = Symbol("savedMods");

export function setSavedMods() {
  let context = $state<SavedMods>({
    mods: {},
    dlProgress: {},
    add(sub, files, tree) {
      ModService.Save(sub, files, tree).catch((e) => {
        console.log(e);
      });
    },
    remove(id) {},
  });

  setContext(SAVED_MODS_KEY, context);

  onMount(() => {
    ModService.Initial().then((data) => {
      if (!data) return;
      console.log("initial", data);
      context.mods = modsFromData(data);
    });
    const unsubscribe = Events.On("mods-loaded", ({ data }) => {
      if (!data) return;
      console.log("loaded", data);
      context.mods = modsFromData(data);
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

export function getSavedMods() {
  return getContext<SavedMods>(SAVED_MODS_KEY);
}
