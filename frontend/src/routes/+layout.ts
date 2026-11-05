import { ModService } from "#lib/bindings/changeme/index.js";
import { compact } from "#lib/ts-ext.ts";
import type { LayoutLoad } from "./$types";

export const prerender = true;
export const ssr = false;

export const load: LayoutLoad = async ({ params }) => {
  const rootFolder = await ModService.FolderConfig();
  const settings = await ModService.Settings();

  const mods = await ModService.Initial();

  return {
    rootFolder,
    characters: compact(rootFolder.characters.characters ?? {}),
    subfolders: compact(rootFolder.subfolders ?? {}),
    settings: settings ?? [],
    mods: compact(mods ?? {}),
  };
};
