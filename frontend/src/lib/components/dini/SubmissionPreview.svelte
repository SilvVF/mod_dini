<script lang="ts">
  import {
    InstallState,
    StorageState,
    type Submission,
  } from "../../bindings/changeme";

  import { epochSecondsToDate } from "./utils";

  import { Badge } from "#lib/components/ui/badge/index.js";
  import { Button } from "#lib/components/ui/button/index.ts";

  import {
    Heart,
    CalendarDays,
    ViewIcon,
    DownloadIcon,
    CheckIcon,
    LoaderCircleIcon,
    TriangleAlertIcon,
    ArchiveRestoreIcon,
  } from "@lucide/svelte";
  import { cn } from "#lib/utils.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import type { ClassValue } from "svelte/elements";

  let {
    submission,
    onDownload,
    class: className
  }: {
    submission: Submission;
    /** Called when the user clicks Download (or Retry after a failed install). */
    onDownload?: (submission: Submission) => void;
    class?: ClassValue | undefined | null; 
  } = $props();

  const saved = getSavedModsContext();

  const local = $derived(saved.mods[`${submission.provider}:${submission.submissionId}`]);

  const packs = $derived(local?.packages ?? []);

  type DownloadStatus =
    | "none"
    | "queued"
    | "downloading"
    | "extracting"
    | "restoring"
    | "error"
    | "downloaded";

  const status: DownloadStatus = $derived.by(() => {
    // Not saved locally at all -> can be downloaded.
    if (!local) return "none";

    if (packs.some((p) => p.state === InstallState.InstallErr)) return "error";
    if (packs.some((p) => p.state === InstallState.InstallDownloading))
      return "downloading";
    if (packs.some((p) => p.state === InstallState.InstallExtracting))
      return "extracting";
    if (packs.some((p) => p.state === InstallState.InstallQueued))
      return "queued";
    if (
      packs.some(
        (p) =>
          p.storage === StorageState.StorageRestoring ||
          p.storage === StorageState.StorageCompressing,
      )
    )
      return "restoring";

    return "downloaded";
  });

  const busy = $derived(
    status === "queued" ||
      status === "downloading" ||
      status === "extracting" ||
      status === "restoring",
  );

  const statusLabel = $derived(
    {
      none: "Download",
      queued: "Queued",
      downloading: "Downloading…",
      extracting: "Extracting…",
      restoring: "Restoring…",
      error: "Retry",
      downloaded: "Downloaded",
    }[status],
  );

  function formatCount(value: number): string {
    return new Intl.NumberFormat("en", {
      notation: "compact",
      maximumFractionDigits: 1,
    }).format(value);
  }

  function formatDate(value: number): string {
    return new Intl.DateTimeFormat("en", {
      month: "short",
      day: "numeric",
      year: "numeric",
    }).format(epochSecondsToDate(value));
  }
</script>

<article
  class={cn(
    "group flex h-full min-w-0 flex-col overflow-hidden rounded-lg border bg-card transition-colors hover:border-primary/50",
    status === "error" && "border-destructive/60",
    className
  )}
>
  <!-- Thumbnail -->
  <a
    data-sveltekit-preload-data="false"
    href={`/mod/${submission.provider}/${submission.slug}`}
    class={cn(
      local ? "opacity-75" : "",
      "relative block aspect-video overflow-hidden bg-muted",
    )}
  >
    {#if submission.thumbnailUrl}
      <img
        src={submission.thumbnailUrl}
        alt={submission.name}
        loading="lazy"
        class="size-full object-cover transition-transform duration-300 group-hover:scale-105"
      />
    {:else}
      <div
        class="flex size-full items-center justify-center text-sm text-muted-foreground"
      >
        No image available
      </div>
    {/if}

    <div
      class="absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100"
    ></div>

    {#if submission.isNsfw || submission.isObsolete || submission.isTombstoned}
      <div class="absolute left-2 top-2 flex flex-wrap gap-1.5">
        {#if submission.isNsfw}
          <Badge
            variant="destructive"
            class="border-0 bg-red-500/90 text-white"
          >
            NSFW
          </Badge>
        {/if}

        {#if submission.isObsolete}
          <Badge
            class="border-0 bg-yellow-500/90 text-black hover:bg-yellow-500/90"
          >
            Obsolete
          </Badge>
        {/if}

        {#if submission.isTombstoned}
          <Badge variant="secondary" class="border-0 bg-black/70 text-white">
            Removed
          </Badge>
        {/if}
      </div>
    {/if}
  </a>

  <!-- Content -->
  <div class="flex min-w-0 flex-1 flex-col gap-2 p-3">
    <a href={`/mod/${submission.provider}/${submission.slug}`} class="min-w-0">
      <h3
        class="truncate text-base font-semibold leading-tight tracking-tight transition-colors group-hover:text-primary"
        title={submission.name}
      >
        {submission.name}
      </h3>
    </a>

    <p class="truncate text-sm text-muted-foreground">
      By {submission.author}
    </p>

    <div
      class="mt-auto flex items-center gap-4 pt-2 text-xs text-muted-foreground"
    >
      <span class="inline-flex items-center gap-1.5" title="Views">
        <ViewIcon class="size-3.5" />
        {formatCount(submission.viewCount)}
      </span>

      <span class="inline-flex items-center gap-1.5" title="Likes">
        <Heart class="size-3.5" />
        {formatCount(submission.likes)}
      </span>

      {#if submission.category}
        <span class="ml-auto truncate text-xs" title={submission.category}>
          {submission.category}
        </span>
      {/if}
    </div>

    <div
      class="flex items-center justify-between gap-2 border-t pt-2 text-xs text-muted-foreground"
    >
      <span class="inline-flex min-w-0 items-center gap-1.5">
        <CalendarDays class="size-3.5 shrink-0" />
        <span class="truncate">{formatDate(submission.remoteUpdatedAt)}</span>
      </span>

      <Button
        variant={status === "error"
          ? "destructive"
          : status === "downloaded"
            ? "secondary"
            : "outline"}
        size="sm"
        class="h-7 shrink-0 gap-1.5 px-2.5 text-xs"
        disabled={busy || status === "downloaded"}
        onclick={() => onDownload?.(submission)}
        title={statusLabel}
        aria-label={statusLabel}
      >
        {#if status === "none"}
          <DownloadIcon class="size-3.5" />
        {:else if status === "restoring"}
          <ArchiveRestoreIcon class="size-3.5" />
        {:else if busy}
          <LoaderCircleIcon class="size-3.5 animate-spin" />
        {:else if status === "error"}
          <TriangleAlertIcon class="size-3.5" />
        {:else}
          <CheckIcon class="size-3.5" />
        {/if}
        {statusLabel}
      </Button>
    </div>
  </div>
</article>
