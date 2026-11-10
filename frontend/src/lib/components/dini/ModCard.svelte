<script lang="ts">
  import { Card } from "#lib/components/ui/card/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { Switch } from "#lib/components/ui/switch/index.ts";
  import { Settings, TrashIcon } from "@lucide/svelte/icons";

  import type {
    CharacterFolder,
    ModSaved,
    RootFolder,
    Subfolder,
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
  }

  let { item, onClick, onSettings, onDelete, onToggle }: Props = $props();

  const enabled = $derived(
    item.local.packages?.some((p) => p.enabled) ?? false,
  );

  const submission = $derived(item.local.submission);
</script>

<Card
  class="group flex min-h-360px flex-col overflow-hidden rounded-2xl border border-border/70 bg-card p-0"
  onclick={onClick}
>
  <div class="relative aspect-4/3 overflow-hidden">
    {#if submission.thumbnailUrl}
      <img
        src={submission.thumbnailUrl}
        alt={submission.name}
        class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
        loading="lazy"
      />
    {:else}
      <Skeleton
        class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
      />
    {/if}
  </div>

  <div class="flex flex-1 flex-col gap-4 px-4">
    <div class="min-w-0">
      <h3 class="line-clamp-2 text-base font-semibold leading-5 tracking-tight">
        {submission?.name ?? item.name}
      </h3>

      <p class="mt-1.5 truncate text-sm text-muted-foreground">
        By {submission?.author ?? "Unknown"}
      </p>
    </div>

    <div class="mt-auto flex items-center justify-between py-2">
      <div class="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8"
          onclick={() => onSettings?.(item)}
          aria-label="Mod settings"
        >
          <Settings class="size-4" />
        </Button>

        <span class="text-sm font-medium">{item.local.packages?.length}</span>

        <Button
          variant="ghost"
          size="icon"
          class="ml-2 h-8 w-8"
          onclick={() => onDelete?.(item)}
          aria-label="Delete mod"
        >
          <TrashIcon class="size-4" />
        </Button>
      </div>

      <Switch
        checked={enabled}
        onCheckedChange={(checked) => onToggle?.(item, checked)}
        aria-label="Enable mod"
      />
    </div>
  </div>
</Card>
