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
