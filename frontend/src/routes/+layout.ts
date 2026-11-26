import { ModService } from "#lib/bindings/changeme/index.js";
import { appState } from "#lib/state/app.svelte.ts";
import { compact } from "#lib/ts-ext.ts";
import type { LayoutLoad } from "./$types";

export const prerender = true;
export const ssr = false;

export const load: LayoutLoad = async ({ params }) => {
  const rootFolder = await ModService.FolderConfig();
  const settings = await ModService.Settings();

  const mods = await ModService.Initial();

  const presets = await ModService.GetPresets();

  return {
    rootFolder: compact(rootFolder ?? {}),
    settings: settings ?? [],
    mods: compact(mods ?? {}),
    presets: presets ?? [],
  };
};
