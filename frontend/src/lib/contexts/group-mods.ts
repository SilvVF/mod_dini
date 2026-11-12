import {
  InstallState,
  type CharacterFolder,
  type ModSaved,
  type Pack,
  type RootFolder,
  type Subfolder,
  type SubmissionFile,
} from "../bindings/changeme";

export type ModGroup = {
  name: string;
  folder: CharacterFolder | Subfolder | RootFolder;
  count: number;
  items: ModSaved[];
};

export type PackageStatus =
  | "available"
  | "queued"
  | "downloading"
  | "extracting"
  | "installed"
  | "update"
  | "error";

export function statusOf(
  file: SubmissionFile,
  local: Pack | undefined,
): PackageStatus {
  if (!local) return "available";
  switch (local.state) {
    case InstallState.InstallErr:
      return "error";
    case InstallState.InstallQueued:
      return "queued";
    case InstallState.InstallDownloading:
      return "downloading";
    case InstallState.InstallExtracting:
      return "extracting";
    case InstallState.InstallReady:
      return local.manifest.hash && local.manifest.hash !== file.md5Checksum
        ? "update"
        : "installed";
    default:
      return "available";
  }
}

export type PackageView = {
  file: SubmissionFile;
  pack: Pack;
  status: PackageStatus;
  enabled: boolean;
};

export type SelectedItem =
  | { type: "nothing"; value: null }
  | { type: "subfolder"; value: Subfolder }
  | { type: "character"; value: CharacterFolder };

export function groupModsByFolders(
  mods: ModSaved[],
  root: RootFolder,
  characters: Record<string, CharacterFolder>,
  subfolders: Record<string, Subfolder>,
): ModGroup[] {
  const modsByCategory = new Map<string, ModSaved[]>();

  for (const mod of mods) {
    const key = `${mod.submission.provider}:${mod.submission.category}`;
    const items = modsByCategory.get(key);

    if (items) {
      items.push(mod);
    } else {
      modsByCategory.set(key, [mod]);
    }
  }

  const groupBySources = (
    name: string,
    folder: CharacterFolder | Subfolder,
  ): ModGroup[] =>
    Object.entries(folder.sources ?? {}).map(([source, id]) => {
      const items = modsByCategory.get(`${source}:${id}`) ?? [];

      return {
        name,
        folder,
        count: items.length,
        items,
      };
    });

  const groups = [
    ...Object.entries(characters).flatMap(([name, folder]) =>
      groupBySources(name, folder),
    ),
    ...Object.entries(subfolders).flatMap(([name, folder]) =>
      groupBySources(name, folder),
    ),
  ];

  const groupedMods = new Set(groups.flatMap((group) => group.items));

  const uncategorizedMods = Object.values(mods).filter(
    (mod) => mod.toolId === root.id && !groupedMods.has(mod),
  );

  if (uncategorizedMods.length > 0) {
    groups.push({
      name: root.id,
      folder: root,
      count: uncategorizedMods.length,
      items: uncategorizedMods,
    });
  }

  return groups;
}
