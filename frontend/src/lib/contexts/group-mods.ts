import {
  InstallState,
  type CharacterFolder,
  type ModSaved,
  type Package,
  type Subfolder,
  type SubmissionFile,
} from "../bindings/changeme";

export type ModGroup = {
  name: string;
  folder: CharacterFolder | Subfolder;
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
  local: Package | undefined,
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
  pack: Package;
  status: PackageStatus;
  enabled: boolean;
};

export type SelectedItem =
  | { type: "nothing"; value: null }
  | { type: "subfolder"; value: Subfolder }
  | { type: "character"; value: CharacterFolder };

export function groupModsByFolders(
  mods: Record<string, ModSaved>,
  characters: Record<string, CharacterFolder>,
  subfolders: Record<string, Subfolder>,
): ModGroup[] {
  const modsByCategory = new Map<string, ModSaved[]>();

  for (const mod of Object.values(mods)) {
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

  return [
    ...Object.entries(characters).flatMap(([name, folder]) =>
      groupBySources(name, folder),
    ),
    ...Object.entries(subfolders).flatMap(([name, folder]) =>
      groupBySources(name, folder),
    ),
  ];
}
