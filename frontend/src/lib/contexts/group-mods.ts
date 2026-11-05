import type {
  CharacterFolder,
  ModSaved,
  Subfolder,
} from "../bindings/changeme";

export type ModGroup = {
  name: string;
  folder: CharacterFolder | Subfolder;
  count: number;
  items: ModSaved[];
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
    ...Object.entries(characters).flatMap(([name, folder]) => groupBySources(name, folder)),
    ...Object.entries(subfolders).flatMap(([name, folder]) => groupBySources(name, folder)),
  ];
}
