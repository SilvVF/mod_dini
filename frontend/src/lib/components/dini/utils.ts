import { ModService } from "#lib/bindings/changeme/index.ts";
import {
  InstallState,
  StorageState,
  type CharacterFolder,
  type ModSaved,
  type Submission,
  type SubmissionFile,
} from "#lib/bindings/changeme/models.ts";
import { statusOf, type PackageView } from "#lib/contexts/group-mods.ts";
import { QueryClient, queryOptions } from "@tanstack/svelte-query";

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

export const queryClient = new QueryClient();

export const submissionPageQuery = (
  provider: string | undefined,
  id: string | number | undefined,
) =>
  queryOptions({
    queryKey: ["submission-page", provider, String(id)],
    queryFn: () => ModService.Page(provider ?? "", String(id)),
    staleTime: 60_000,
    enabled: id !== undefined && provider !== undefined,
  });

export function buildPackageViews(
  submission: Submission,
  files: SubmissionFile[],
  localMod?: ModSaved,
): PackageView[] {
  const localPacks = localMod?.packages;
  return files.map((file) => {
    const installed = localPacks?.find((pack) => pack.fileId === file.fileId);

    return {
      file,
      pack: installed ?? {
        id: file.fileId.toString(),
        modId: `${submission.provider}:${submission.slug}`,
        fileId: file.fileId,
        name: file.fileName,
        state: InstallState.$zero,
        enabled: false,
        storage: StorageState.$zero,
        archiveFormat: "",
        archiveSize: -1,
        disabledAt: -1,
        extractedSize: -1,
        manifest: { relPath: "", size: file.size, hash: file.md5Checksum },
      },
      status: statusOf(file, installed),
      enabled: installed !== undefined, // "selected" for the dialog
    };
  });
}

export type Matcher = {
  host: RegExp;
  route: (u: URL) => string | null;
  link: (id: string) => string;
};

export const providers: Record<string, Matcher> = {
  gamebanana: {
    host: /(^|\.)gamebanana\.com$/,
    route(u) {
      const mod = u.pathname.match(/^\/mods\/(\d+)\/?$/);
      if (mod) return `/mod/gamebanana/${mod[1]}`;
      const cat = u.pathname.match(/^\/mods\/cats\/(\d+)\/?$/);
      if (cat) return `/source/gamebanana/${cat[1]}${u.search}`;
      return null;
    },
    link: (id) => `https://gamebanana.com/mods/${id}`,
  },
  huihui: {
    host: /(^|\.)huihui648\.org$/,
    route(u) {
      const cat = u.pathname.match(/^\/category\/([^/]+)\/?$/);
      if (cat) return `/source/huihui/${cat[1]}${u.search}`;
      const post = u.pathname.match(/^\/([^/]+)\/?$/);
      if (post && post[1] !== "category") return `/mod/huihui/${post[1]}`;
      return null;
    },
    link: (id) => `https://huihui648.org/${encodeURIComponent(id)}/`,
  },
};

export function providerRoute(raw: string): string | null {
  let u: URL;
  try {
    u = new URL(raw.trim());
  } catch {
    return null;
  }
  if (u.protocol !== "http:" && u.protocol !== "https:") return null;

  for (const m of Object.values(providers)) {
    if (m.host.test(u.hostname)) return m.route(u);
  }
  return null;
}

export function modLink(mod: ModSaved): string | null {
  const { provider, submissionId, profileUrl } = mod.submission;
  const link = providers[provider]?.link(submissionId);
  return link ?? (profileUrl || null);
}
