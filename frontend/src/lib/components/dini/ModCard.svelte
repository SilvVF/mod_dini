<script lang="ts">
  import { Card } from "#lib/components/ui/card/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { Switch } from "#lib/components/ui/switch/index.ts";
  import {
    Box,
    Settings,
    TrashIcon,
    LoaderCircleIcon,
    TriangleAlertIcon,
    RotateCcwIcon,
    ArchiveRestoreIcon,
    ArrowUpRightIcon,
    SnowflakeIcon,
  } from "@lucide/svelte/icons";

  import {
    InstallState,
    StorageState,
    type CharacterFolder,
    type ModSaved,
    type RootFolder,
    type Subfolder,
  } from "#lib/bindings/changeme/index.ts";
  import Skeleton from "../ui/skeleton/skeleton.svelte";

  export interface GroupedCharacter {
    name: string;
    folder: CharacterFolder | Subfolder | RootFolder;
    local: ModSaved;
  }

  interface Props {
    item: GroupedCharacter;
    onClick: () => void;
    onSettings?: (item: GroupedCharacter) => void;
    onDelete?: (item: GroupedCharacter) => void;
    onToggle?: (item: GroupedCharacter, enabled: boolean) => void;
    onRetry?: (item: GroupedCharacter) => void;
    conflict?: boolean;
  }

  let {
    item,
    onClick,
    onSettings,
    onDelete,
    onToggle,
    onRetry,
    conflict,
  }: Props = $props();

  const INSTALLING = [
    InstallState.InstallQueued,
    InstallState.InstallDownloading,
    InstallState.InstallExtracting,
  ];
  const RESTORING = [
    StorageState.StorageRestoring,
    StorageState.StorageCompressing,
  ];
  const ERROR = [InstallState.InstallErr];

  type Status = "error" | "installing" | "restoring" | "ready";

  const submission = $derived(item.local.submission);
  const packs = $derived(item.local.packages ?? []);

  const enabled = $derived(packs.some((p) => p.enabled));

  // Nothing on disk to enable (no extracted files and no archive)
  const noStorage = $derived(
    packs.length > 0 &&
      packs.every((p) => p.storage === StorageState.StorageNone),
  );

  // At least one pack is compressed into an archive
  const cold = $derived(
    packs.some((p) => p.storage === StorageState.StorageCold),
  );

  const status: Status = $derived.by(() => {
    if (packs.some((p) => ERROR.includes(p.state))) {
      return "error";
    } else if (packs.some((p) => INSTALLING.includes(p.state))) {
      return "installing";
    } else if (packs.some((p) => RESTORING.includes(p.storage))) {
      return "restoring";
    } else {
      return "ready";
    }
  });

  const busy = $derived(status === "installing" || status === "restoring");

  // Fully disabled: finished installing, but no pack is enabled
  const dimmed = $derived(status === "ready" && (!enabled || noStorage));

  const installLabel = $derived.by(() => {
    const states = packs.map((p) => p.state);
    if (states.includes(InstallState.InstallDownloading)) return "Downloading…";
    if (states.includes(InstallState.InstallExtracting)) return "Extracting…";
    return "Queued";
  });

  const statusLabel = $derived(
    status === "restoring" ? "Restoring from archive…" : installLabel,
  );

  const detailsHref = $derived(
    `/mod/${submission?.provider}/${submission?.submissionId}`,
  );

  // Keeps card-level onClick from firing when pressing inner controls.
  const stop = (fn?: () => void) => (e: MouseEvent) => {
    e.stopPropagation();
    fn?.();
  };
</script>

<Card
  class="group flex cursor-pointer flex-col overflow-hidden rounded-2xl border bg-card p-0 {status ===
  'error'
    ? 'border-destructive/60'
    : 'border-border/70'}"
  onclick={onClick}
>
  <div class="relative aspect-4/3 overflow-hidden">
    {#if conflict && enabled && status === "ready"}
      <span
        class="absolute z-10 left-2 top-2 rounded-md bg-amber-300 px-2 py-0.5 text-xs font-semibold text-amber-950 shadow-sm"
        title={packs.filter((p) => p.enabled).length > 1
          ? "More than one pack of this mod is on"
          : "Another mod for this character is also on"}
      >
        Conflict
      </span>
    {/if}
    {#if submission?.thumbnailUrl}
      <img
        src={submission.thumbnailUrl}
        alt={submission.name}
        class="h-full w-full object-cover transition duration-300 group-hover:scale-[1.02] {busy
          ? 'opacity-60'
          : ''} {status === 'error' ? 'opacity-40 grayscale' : ''} {dimmed
          ? 'brightness-50'
          : ''}"
        loading="lazy"
      />
    {:else}
      <Skeleton class="h-full w-full" />
    {/if}

    {#if busy}
      <div
        class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-background/50 backdrop-blur-[1px]"
      >
        {#if status === "restoring"}
          <ArchiveRestoreIcon class="size-6 text-primary" />
        {:else}
          <LoaderCircleIcon class="size-6 animate-spin text-primary" />
        {/if}
        <span class="text-sm font-medium">{statusLabel}</span>
      </div>
    {:else if status === "error"}
      <div
        class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-destructive/10"
      >
        <TriangleAlertIcon class="size-7 text-destructive" />
        <span class="text-sm font-medium text-destructive">
          Install failed
        </span>
      </div>
    {:else if cold}
      <div
        class="absolute right-2 top-2 rounded-full bg-background/80 p-1.5 text-sky-500 shadow-sm backdrop-blur-sm"
        title="Archived (cold storage)"
      >
        <SnowflakeIcon class="size-4" aria-label="Archived" />
      </div>
    {/if}
  </div>

  <div class="flex flex-1 flex-col gap-2 px-4 pb-3">
    <div class="min-w-0">
      <h3 class="line-clamp-2 text-base font-semibold leading-5 tracking-tight">
        {submission?.name ?? item.name}
      </h3>

      <p class="mt-1.5 truncate text-sm text-muted-foreground">
        By {submission?.author ?? "Unknown"}
      </p>
    </div>

    {#if busy}
      <div class="mt-auto flex flex-col gap-2 py-2">
        <div class="h-1 w-full overflow-hidden rounded-full bg-muted">
          <div
            class="h-full w-full animate-pulse rounded-full bg-primary/70"
          ></div>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-muted-foreground">{statusLabel}</span>
          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            href={detailsHref}
            aria-label="View mod details"
            onclick={stop()}
          >
            <ArrowUpRightIcon class="size-4" />
          </Button>
        </div>
      </div>
    {:else if status === "error"}
      <div class="mt-auto flex flex-col gap-2 py-2">
        <p class="text-xs text-destructive">
          Something went wrong while installing this mod.
        </p>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              class="h-8 gap-1.5"
              onclick={stop(() => onRetry?.(item))}
            >
              <RotateCcwIcon class="size-3.5" />
              Retry
            </Button>
            <Button
              variant="ghost"
              size="icon"
              class="size-8"
              onclick={stop(() => onDelete?.(item))}
              aria-label="Delete mod"
            >
              <TrashIcon class="size-4" />
            </Button>
          </div>
          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            href={detailsHref}
            aria-label="View mod details"
            onclick={stop()}
          >
            <ArrowUpRightIcon class="size-4" />
          </Button>
        </div>
      </div>
    {:else}
      <div class="mt-auto flex flex-col gap-1 py-2">
        <div class="flex items-center">
          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            onclick={stop(() => onSettings?.(item))}
            aria-label="Mod settings"
          >
            <Settings class="size-4" />
          </Button>

          <span class="px-0.5 text-sm font-medium">{packs.length}</span>

          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            onclick={stop(() => onDelete?.(item))}
            aria-label="Delete mod"
          >
            <TrashIcon class="size-4" />
          </Button>

          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            href={`/model/${item.local.id}/${packs.at(0)?.id}`}
            aria-label="Preview model"
            onclick={stop()}
          >
            <Box class="size-4" />
          </Button>

          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            href={detailsHref}
            aria-label="View mod details"
            onclick={stop()}
          >
            <ArrowUpRightIcon class="size-4" />
          </Button>
        </div>

        <div
          class="flex justify-end"
          title={noStorage ? "Mod files are not available" : undefined}
          onclick={(e) => e.stopPropagation()}
          role="presentation"
        >
          <Switch
            checked={enabled && !noStorage}
            disabled={noStorage}
            onCheckedChange={(checked) => {
              if (noStorage) return;
              onToggle?.(item, checked);
            }}
            aria-label="Enable mod"
          />
        </div>
      </div>
    {/if}
  </div>
</Card>
