<script lang="ts" module>
  import type {
    CharacterFolder,
    ModSaved,
    RootFolder,
    Subfolder,
  } from "#lib/bindings/changeme/index.ts";

  export interface GroupedCharacter {
    name: string;
    folder: CharacterFolder | Subfolder | RootFolder;
    local: ModSaved;
  }
</script>

<script lang="ts">
  import {
    ArchiveRestore,
    ArrowUpRight,
    Box,
    Ellipsis,
    LoaderCircle,
    RotateCcw,
    Snowflake,
    Trash2,
    TriangleAlert,
  } from "@lucide/svelte";
  import { goto } from "$app/navigation";
  import {
    InstallState,
    StorageState,
    type Pack,
  } from "#lib/bindings/changeme/index.ts";
  import { openPreview } from "#lib/state/app.svelte.ts";
  import { cn } from "#lib/utils.ts";
  import { Badge } from "#lib/components/ui/badge/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { Card } from "#lib/components/ui/card/index.ts";
  import * as DropdownMenu from "#lib/components/ui/dropdown-menu/index.ts";
  import Skeleton from "#lib/components/ui/skeleton/skeleton.svelte";
  import { Switch } from "#lib/components/ui/switch/index.ts";

  interface Props {
    item: GroupedCharacter;
    onClick: () => void;
    onDelete?: (item: GroupedCharacter) => void;
    onToggle?: (item: GroupedCharacter, enabled: boolean) => void;
    onRetry?: (item: GroupedCharacter) => void;
    conflict?: boolean;
  }

  let { item, onClick, onDelete, onToggle, onRetry, conflict }: Props =
    $props();

  const INSTALLING = [
    InstallState.InstallQueued,
    InstallState.InstallDownloading,
    InstallState.InstallExtracting,
  ];
  const RESTORING = [
    StorageState.StorageRestoring,
    StorageState.StorageCompressing,
  ];

  const isActive = (p: Pack) => p.storage === StorageState.StorageActive;
  const isBusy = (p: Pack) => RESTORING.includes(p.storage);
  const canPreview = (p: Pack) =>
    p.state === InstallState.InstallReady &&
    p.storage !== StorageState.StorageNone &&
    !isBusy(p);

  const submission = $derived(item.local.submission);
  const packs = $derived(item.local.packages ?? []);

  const packLabel = (p: Pack) => p.name || `Pack ${packs.indexOf(p) + 1}`;

  const activeCount = $derived(packs.filter(isActive).length);
  const enabled = $derived(activeCount > 0);
  const noStorage = $derived(
    packs.length > 0 &&
      packs.every((p) => p.storage === StorageState.StorageNone),
  );
  const cold = $derived(
    packs.some((p) => p.storage === StorageState.StorageCold),
  );
  const previewable = $derived(packs.filter(canPreview));

  const status = $derived.by<"error" | "installing" | "restoring" | "ready">(
    () => {
      if (packs.some((p) => p.state === InstallState.InstallErr))
        return "error";
      if (packs.some((p) => INSTALLING.includes(p.state))) return "installing";
      if (packs.some(isBusy)) return "restoring";
      return "ready";
    },
  );
  const ready = $derived(status === "ready");
  const busy = $derived(status === "installing" || status === "restoring");
  const dimmed = $derived(ready && !enabled);

  const busyLabel = $derived.by(() => {
    const hasState = (s: InstallState) => packs.some((p) => p.state === s);
    if (status === "restoring") {
      return packs.some((p) => p.storage === StorageState.StorageCompressing)
        ? "Compressing…"
        : "Restoring from archive…";
    }
    if (hasState(InstallState.InstallDownloading)) return "Downloading…";
    if (hasState(InstallState.InstallExtracting)) return "Extracting…";
    return "Queued";
  });

  const conflictHint = $derived(
    activeCount > 1
      ? "More than one pack of this mod is on"
      : "Another mod for this character is also on",
  );

  const detailsHref = $derived(
    `/mod/${submission?.provider}/${submission?.submissionId}`,
  );

  let menuOpen = $state(false);

  function previewPack(e: Event, pack: Pack) {
    openPreview(e, item.local.id, pack.id);
    menuOpen = false;
  }

  const stop = (e: Event) => e.stopPropagation();

  const menuItem = "gap-2.5 py-2";
  const menuIcon = "text-muted-foreground size-4";
</script>

<Card
  class={cn(
    "group hover:border-primary/40 cursor-pointer gap-0 overflow-hidden rounded-xl p-0 transition-colors",
    status === "error" && "border-destructive/60 hover:border-destructive",
    "h-full w-full",
  )}
  onclick={onClick}
>
  <div class="bg-muted relative shrink-0 overflow-hidden h-7/12">
    {#if submission?.thumbnailUrl}
      <img
        src={submission.thumbnailUrl}
        alt={submission.name}
        loading="lazy"
        class={cn(
          "size-full object-cover transition duration-300 group-hover:scale-[1.02]",
          busy && "opacity-60",
          status === "error" && "opacity-40 grayscale",
          dimmed && "grayscale",
        )}
      />
    {:else}
      <Skeleton class="size-full rounded-none" />
    {/if}

    {#if busy || status === "error"}
      <div
        class={cn(
          "absolute inset-0 flex flex-col items-center justify-center gap-2",
          busy ? "bg-background/50 backdrop-blur-[1px]" : "bg-destructive/10",
        )}
      >
        {#if status === "error"}
          <TriangleAlert class="text-destructive size-7" />
          <span class="text-destructive text-sm font-medium">
            Install failed
          </span>
        {:else}
          {#if status === "restoring"}
            <ArchiveRestore class="text-primary size-6" />
          {:else}
            <LoaderCircle class="text-primary size-6 animate-spin" />
          {/if}
          <span class="text-sm font-medium">{busyLabel}</span>
        {/if}
      </div>
    {/if}

    {#if busy}
      <div
        class="bg-primary/70 absolute inset-x-0 bottom-0 h-1 animate-pulse"
      ></div>
    {/if}

    {#if conflict && enabled && ready}
      <Badge
        class="absolute top-2 left-2 z-10 border-transparent bg-amber-300 text-amber-950"
        title={conflictHint}
      >
        Conflict
      </Badge>
    {/if}

    {#if cold && ready}
      <Badge
        variant="secondary"
        class="absolute top-2 right-2 gap-1 text-sky-500"
        title="Archived (cold storage)"
      >
        <Snowflake class="size-3" />
        Archived
      </Badge>
    {/if}

    {#if packs.length > 1}
      <Badge variant="secondary" class="absolute bottom-3 left-2">
        {packs.length} packs
      </Badge>
    {/if}
  </div>

  <div class="flex flex-1 flex-col gap-4 p-4 pb-3">
    <div class="min-w-0">
      <h3 class="line-clamp-2 text-lg leading-6 font-semibold tracking-tight">
        {submission?.name ?? item.name}
      </h3>
      <p class="text-muted-foreground mt-1 truncate text-sm line-clamp-1">
        By {submission?.author ?? "Unknown"}
      </p>
    </div>

    <div class="mt-auto flex items-center justify-between gap-2">
      <div class="flex items-center" role="presentation" onclick={stop}>
        {#if !busy}
          <Button
            variant="ghost"
            size="icon"
            class="text-muted-foreground hover:text-destructive size-8"
            onclick={() => onDelete?.(item)}
            aria-label="Delete mod"
            title="Delete mod"
          >
            <Trash2 class="size-4" />
          </Button>
        {/if}
      </div>

      <div class="flex items-center gap-2" role="presentation" onclick={stop}>
        {#if ready}
          <span title={noStorage ? "Mod files are not available" : undefined}>
            <Switch
              checked={enabled}
              disabled={noStorage}
              onCheckedChange={(checked) => onToggle?.(item, checked)}
              aria-label="Enable mod"
            />
          </span>
        {/if}

        <DropdownMenu.Root bind:open={menuOpen}>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button
                {...props}
                variant="ghost"
                size="icon"
                class="size-8"
                aria-label="More actions"
                title="More actions"
              >
                <Ellipsis class="size-4" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>

          <DropdownMenu.Content align="end" class="w-56 rounded-xl p-1.5">
            {#if status === "error"}
              <DropdownMenu.Item
                class={menuItem}
                onSelect={() => onRetry?.(item)}
              >
                <RotateCcw class={menuIcon} />
                Retry install
              </DropdownMenu.Item>
              <DropdownMenu.Separator />
            {/if}

            {#if ready && previewable.length === 1}
              <DropdownMenu.Item
                class={menuItem}
                onSelect={(e) => previewPack(e, previewable[0])}
              >
                <Box class={menuIcon} />
                Preview in 3D
              </DropdownMenu.Item>
              <DropdownMenu.Separator />
            {:else if ready && previewable.length > 1}
              <DropdownMenu.Sub>
                <DropdownMenu.SubTrigger class={menuItem}>
                  <Box class={menuIcon} />
                  Preview in 3D
                </DropdownMenu.SubTrigger>
                <DropdownMenu.SubContent class="rounded-xl p-1.5">
                  {#each previewable as pack (pack.id)}
                    <DropdownMenu.Item
                      class={menuItem}
                      onSelect={(e) => previewPack(e, pack)}
                    >
                      {packLabel(pack)}
                    </DropdownMenu.Item>
                  {/each}
                </DropdownMenu.SubContent>
              </DropdownMenu.Sub>
              <DropdownMenu.Separator />
            {/if}

            <DropdownMenu.Item
              class={menuItem}
              onSelect={() => goto(detailsHref)}
            >
              <ArrowUpRight class={menuIcon} />
              View details
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  </div>
</Card>
