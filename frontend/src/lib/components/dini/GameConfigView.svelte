<script lang="ts">
  import { Diamond, ChevronRight, X } from "@lucide/svelte";
  import type { GameConfig } from "./types.js";

  let {
    config,
    onselectmod,
    onunloadmod,
  }: {
    config: GameConfig;
    onselectmod?: (id: string) => void;
    onunloadmod?: (id: string) => void;
  } = $props();

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
</script>

<div class="flex min-h-0 flex-1 flex-col">
  <header class="flex items-center gap-2 px-5 pt-5 pb-3">
    <Diamond class="text-primary size-4 fill-current" />
    <h2 class="font-serif text-xl font-bold">{config.preset.name}</h2>
    <ChevronRight class="text-muted-foreground size-3" />
  </header>

  <ol class="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-3 pb-4">
    {#each config.mods as mod, i (mod.id)}
      <li class="group relative">
        <button
          type="button"
          onclick={() => onselectmod?.(mod.id)}
          class="hover:bg-accent focus-visible:ring-ring flex w-full items-center
		  gap-3 rounded-lg px-2 py-2 pr-9 text-left outline-none focus-visible:ring-2
		  {false ? '' : 'opacity-50'}"
        >
          <span
            class="text-muted-foreground w-4 text-right text-xs tabular-nums"
            >{i + 1}</span
          >
          <span class="bg-muted size-10 shrink-0 overflow-hidden rounded-md">
            {#if mod.submission.images}
              <img
                src={mod.submission.images}
                alt=""
                class="size-full object-cover"
              />
            {/if}
          </span>
          <span class="min-w-0">
            <span class="block truncate text-sm font-semibold"
              >{mod.submission.name}</span
            >
            <span class="text-muted-foreground block truncate text-xs"
              >{mod.submission.category}</span
            >
          </span>
        </button>
        <button
          type="button"
          aria-label="Unload {mod.submission.name}"
          onclick={() => onunloadmod?.(mod.id)}
          class="text-muted-foreground hover:text-foreground focus-visible:ring-ring absolute top-1/2 right-3 -translate-y-1/2 rounded p-1 opacity-0 outline-none group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:ring-2"
        >
          <X class="size-3.5" />
        </button>
      </li>
    {:else}
      <li class="text-muted-foreground px-3 py-8 text-center text-sm">
        No mods in this preset yet.
      </li>
    {/each}
  </ol>

  <dl class="border-t px-5 py-4 grid grid-cols-2 gap-x-6 gap-y-4">
    {#each stats as s}
      <div>
        <dd class="font-serif text-xl leading-none font-bold">
          {s.value}{#if s.suffix}<span
              class="text-muted-foreground text-sm font-normal">{s.suffix}</span
            >{/if}
        </dd>
        <dt class="text-muted-foreground mt-1 text-xs">{s.label}</dt>
      </div>
    {/each}
  </dl>
</div>
