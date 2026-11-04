import type { CharacterFolder, ModSaved } from "#lib/bindings/changeme/models.ts";

export const epochSecondsToDate = (timestamp: number): Date => {
  return new Date(timestamp * 1000);
};

const BYTE_UNITS = ["B", "KB", "MB", "GB"] as const;

export const formatByteSize = (bytes: number): string => {
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < BYTE_UNITS.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(2)} ${BYTE_UNITS[unitIndex]}`;
};

export const formatSize = (bytes: number) =>
  bytes >= 1024 * 1024
    ? `${(bytes / 1024 / 1024).toFixed(1)} MB`
    : `${Math.max(1, Math.round(bytes / 1024))} KB`;

export function groupModsByCharacter(
  mods: Record<string, ModSaved>,
  characters: [string, CharacterFolder][],
) {
  const modsByCategory = new Map<string, (typeof mods)[string][]>();

    for (const mod of Object.values(mods)) {
      const key = `${mod.submission.provider}:${mod.submission.category}`;

      const items = modsByCategory.get(key);

      if (items) {
        items.push(mod);
      } else {
        modsByCategory.set(key, [mod]);
      }
    }

    return characters.flatMap(([name, folder]) =>
      Object.entries(folder.sources ?? {}).map(([source, id]) => {
        const items = modsByCategory.get(`${source}:${id}`) ?? [];

        return {
          name,
          folder,
          count: items.length,
          items,
        };
      }),
    );
}
