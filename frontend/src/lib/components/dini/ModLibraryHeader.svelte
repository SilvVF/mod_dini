<script lang="ts">
  import { Button } from "#lib/components/ui/button/index.ts";
  import type { SelectedItem } from "#lib/contexts/group-mods.ts";
  import { ArrowDownToLine } from "@lucide/svelte";
  import Skeleton from "../ui/skeleton/skeleton.svelte";

  interface Props {
    selected: SelectedItem;
    game?: string;
    image?: string;
    charactersCount: number;
    subfoldersCount: number;
    mods: number;
    loaded: number;
    onUnloadAll?: () => void;
  }

  let {
    selected,
    game = "ZENLESS ZONE ZERO",
    image,
    charactersCount,
    subfoldersCount,
    mods,
    loaded,
    onUnloadAll,
  }: Props = $props();

  let config = $derived.by(() => {
    switch (selected.type) {
      case "character":
        return {
          eyebrowSuffix: "CHARACTER",
          title: selected.value.name,
          icon: selected.value.icon,
          subtitle: `${mods} installed for ${selected.value.name}`,
        };
      case "subfolder":
        return {
          eyebrowSuffix: "CATEGORY",
          title: selected.value.name,
          icon: undefined,
          subtitle: `${mods} installed for ${selected.value.name}`,
        };
      default:
        return {
          eyebrowSuffix: "LIBRARY",
          title: "Mod Library",
          icon: undefined,
          subtitle: `${charactersCount} characters · ${subfoldersCount} subfolders`,
        };
    }
  });
</script>

<header
  class="relative isolate flex w-full max-h-60 items-end overflow-hidden rounded-xl border border-border/40 bg-card"
>
  {#if image}
    <img
      loading="eager"
      fetchpriority="high"
      src={image}
      alt={game}
      class="absolute inset-0 -z-20 h-full w-full object-cover object-center"
    />
  {/if}

  <div
    class="absolute inset-0 -z-10 bg-linear-to-r from-black/85 via-black/55 to-black/30"
  ></div>

  <div class="flex w-full items-end justify-between gap-6 p-6">
    <div class="min-w-0">
      {#if config.icon}
        <img
          src={config.icon}
          alt="character icon"
          class="mb-3 size-16 rounded-full border-2 border-white/30 object-cover"
        />
      {:else}
        <div class="mb-3 size-16"></div>
      {/if}

      <p class="mb-3 text-xs font-semibold tracking-[0.3em] text-indigo-300">
        {game} · {config.eyebrowSuffix}
      </p>

      <h1 class="truncate text-4xl font-bold tracking-tight text-white">
        {config.title}
      </h1>

      <p class="mt-4 text-sm font-medium text-white/75">
        {config.subtitle}
      </p>
    </div>

    <div class="flex shrink-0 items-start gap-6 self-start">
      <div class="text-center">
        <p class="text-3xl font-bold text-white">{mods}</p>
        <p class="text-xs font-semibold tracking-widest text-white/70">MODS</p>
      </div>

      <div class="text-center">
        <p class="text-3xl font-bold text-indigo-300">{loaded}</p>
        <p class="text-xs font-semibold tracking-widest text-white/70">
          LOADED
        </p>
      </div>
    </div>
  </div>

  {#if onUnloadAll}
    <Button
      variant="outline"
      class="absolute bottom-5 right-5 rounded-full border-white/40 bg-black/30 text-white hover:bg-black/50"
      onclick={onUnloadAll}
    >
      <ArrowDownToLine class="size-4" />
      Unload all
    </Button>
  {/if}
</header>
