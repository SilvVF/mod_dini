import { ModService } from "#lib/bindings/changeme/index.ts";
import type { LayoutLoad } from "./$types";

export const prerender = true;
export const ssr = false;

export const load: LayoutLoad = async ({ params }) => {
  const rootFolder = await ModService.FolderConfig();

  return {
    rootFolder,
    characters: rootFolder.characters,
  };
};