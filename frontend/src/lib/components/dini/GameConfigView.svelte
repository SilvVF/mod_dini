<script lang="ts">
  import { Diamond, ChevronRight, X } from "@lucide/svelte";
  import type { GameConfig } from "./types.js";
  import { getSavedModsContext } from "../../contexts/saved-mods.svelte.js";
  import * as DropdownMenu from "../ui/dropdown-menu/index.js";
  import * as Dialog from "../ui/dialog/index.js";
  import { Input } from "../ui/input/index.js";
  import { Button } from "../ui/button/index.js";

  let {
    config,
    onselectmod,
    onselectpreset,
  }: {
    config: GameConfig;
    onselectmod?: (id: string) => void;
    /** Called when a preset is picked from the dropdown */
    onselectpreset?: (id: number | undefined) => void;
  } = $props();

  const saved = getSavedModsContext();

  const stats = $derived([
    {
      value: String(config.stats.loaded),
      suffix: `/${config.stats.total}`,
      label: "loaded",
    },
    { value: config.stats.sizeOnDisk, label: "on disk" },
    { value: String(config.stats.overlaps), label: "overlaps" },
    { value: String(config.stats.updates), label: "updates" },
  ]);

  const activeId = $derived(config.viewedPreset?.id);

  let createOpen = $state(false);
  let newName = $state("");

  function submitCreate(e: SubmitEvent) {
    e.preventDefault();
    const name = newName.trim();
    if (!name) return;
    saved.createPreset(name, config.game.id);
    newName = "";
    createOpen = false;
  }

  let renameDialogOpen = $state(false);

  function enablePreset() {
    if (config.viewedPreset) {
      saved.enablePreset(config.viewedPreset.id);
    }
  }

  function commitRename(id: number, current: string, next: string) {
    const name = next.trim();
    if (name && name !== current) saved.renamePreset(id, name);
  }
</script>

<div class="flex min-h-0 flex-1 flex-col">
  <header class="px-5 pt-5 pb-3">
    <DropdownMenu.Root>
      <div class="flex flex-row justify-between">
        <DropdownMenu.Trigger
          class="focus-visible:ring-ring -mx-1 flex items-center gap-2 rounded-md px-1 outline-none focus-visible:ring-2"
        >
          <Diamond class="text-primary size-4 fill-current" />
          <h2 class="text-xl font-bold">
            {config.viewedPreset?.name ?? "Default"}
          </h2>
          <ChevronRight class="text-muted-foreground size-3" />
        </DropdownMenu.Trigger>
        {#if config.viewedPreset}
          <Button onclick={enablePreset}>Enable</Button>
        {/if}
      </div>

      <DropdownMenu.Content align="start" class="w-72">
        <DropdownMenu.Item
          onSelect={() => onselectpreset?.(undefined)}
          class="flex items-center justify-between gap-4 {!activeId
            ? 'text-primary focus:text-primary'
            : ''}"
        >
          <span class="truncate">Default</span>
          <span class="text-muted-foreground shrink-0 text-xs">
            {!activeId ? "Active now" : `${config.mods?.length ?? 0} mods`}
          </span>
        </DropdownMenu.Item>
        {#each config.presets as preset (preset.id)}
          {@const active = preset.id === activeId}
          <DropdownMenu.Item
            onSelect={() =>
              active
                ? onselectpreset?.(undefined)
                : onselectpreset?.(preset.id)}
            class="flex items-center justify-between gap-4 {active
              ? 'text-primary focus:text-primary'
              : ''}"
          >
            <span class="truncate">{preset.name}</span>
            <span class="text-muted-foreground shrink-0 text-xs">
              {active ? "Active now" : `${preset.mods?.length ?? 0} mods`}
            </span>
          </DropdownMenu.Item>
        {/each}

        {#if config.presets.length > 0}
          <DropdownMenu.Separator />
        {/if}

        <DropdownMenu.Item onSelect={() => (createOpen = true)}>
          Save as new preset
        </DropdownMenu.Item>
        <DropdownMenu.Item onSelect={() => (renameDialogOpen = true)}>
          Manage presets...
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </header>

  <ol class="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-3 pb-4">
    {#each config.mods as mod, i (mod.id)}
      <li class="group relative">
        <button
          type="button"
          onclick={() => onselectmod?.(mod.id)}
          class="hover:bg-accent focus-visible:ring-ring flex w-full items-center
      gap-3 rounded-lg px-2 py-2 pr-9 text-left outline-none focus-visible:ring-2"
        >
          <span
            class="text-muted-foreground w-5 shrink-0 text-right text-xs tabular-nums"
            >{i + 1}</span
          >
          <span class="bg-muted size-10 shrink-0 overflow-hidden rounded-md">
            {#if (mod.submission.images?.length ?? 0) > 0}
              <img
                src={mod.submission.images![0]}
                alt=""
                class="size-full object-cover"
              />
            {/if}
          </span>
          <span class="min-w-0 flex-1">
            <span class="block truncate text-sm font-semibold"
              >{mod.submission.name}</span
            >
            <span class="text-muted-foreground block truncate text-xs"
              >{mod.submission.category}</span
            >
          </span>
        </button>
        {#if config.viewedPreset}
          <button
            type="button"
            aria-label="Remove {mod.submission.name} from preset"
            onclick={() =>
              saved.removeFromPreset(config.viewedPreset!.id, mod.id)}
            class="text-muted-foreground hover:text-foreground focus-visible:ring-ring absolute top-1/2 right-3 -translate-y-1/2 rounded p-1 opacity-0 outline-none group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:ring-2"
          >
            <X class="size-3.5" />
          </button>
        {/if}
      </li>
    {:else}
      <li class="text-muted-foreground px-3 py-8 text-center text-sm">
        No mods in this preset yet.
      </li>
    {/each}
  </ol>

  <dl class="grid grid-cols-2 gap-x-6 gap-y-4 border-t px-5 py-4">
    {#each stats as s}
      <div>
        <dd class="text-xl leading-none font-bold">
          {s.value}{#if s.suffix}<span
              class="text-muted-foreground text-sm font-normal">{s.suffix}</span
            >{/if}
        </dd>
        <dt class="text-muted-foreground mt-1 text-xs">{s.label}</dt>
      </div>
    {/each}
  </dl>
</div>

<Dialog.Root bind:open={createOpen}>
  <Dialog.Content class="sm:max-w-sm">
    <form onsubmit={submitCreate} class="grid gap-4">
      <Dialog.Header>
        <Dialog.Title>Save as new preset</Dialog.Title>
        <Dialog.Description>Give your preset a name.</Dialog.Description>
      </Dialog.Header>
      <Input bind:value={newName} placeholder="Preset name" autofocus />
      <Dialog.Footer>
        <Button type="submit" disabled={!newName.trim()}>Save</Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={renameDialogOpen}>
  <Dialog.Content class="sm:max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Manage presets</Dialog.Title>
      <Dialog.Description>
        Edit a name and press Enter or click away to save.
      </Dialog.Description>
    </Dialog.Header>
    <ul class="grid gap-2">
      {#each config.presets as preset (preset.id)}
        <li class="flex items-center gap-2">
          <Input
            value={preset.name}
            aria-label="Preset name"
            onkeydown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                commitRename(preset.id, preset.name, e.currentTarget.value);
                renameDialogOpen = false;
              } else if (e.key === "Escape") {
                e.currentTarget.value = preset.name;
              }
            }}
            onblur={(e) => (e.currentTarget.value = preset.name)}
          />
          <span class="text-muted-foreground shrink-0 text-xs">
            {preset.mods?.length ?? 0} mods
          </span>
        </li>
      {/each}
    </ul>
  </Dialog.Content>
</Dialog.Root>
