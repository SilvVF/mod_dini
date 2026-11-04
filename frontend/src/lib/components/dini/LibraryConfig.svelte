<script lang="ts">
  import type { SidebarProps } from "./types.js";
  import GameConfigView from "./GameConfigView.svelte";
  import ModConfigView from "./ModConfigView.svelte";
  import LaunchBar from "./LaunchBar.svelte";

  let {
    config,
    selectedMod = null,
    folders = [],
    onselectmod,
    onclosemod,
    onunloadmod,
    onmodchange,
    onopenfolder,
    onopenlink,
    onvanillachange,
    onstart,
  }: SidebarProps = $props();

  const activeCount = $derived(config.launchVanilla ? 0 : config.stats.loaded);
</script>

<aside
  class="bg-card text-card-foreground flex h-full w-full flex-col overflow-hidden"
>
  <div class="relative flex min-h-0 flex-1 flex-col overflow-hidden">
    <!-- Game config is always rendered -->
    <div class="flex min-h-0 flex-1 flex-col" inert={!!selectedMod}>
      <GameConfigView {config} {onselectmod} {onunloadmod} />
    </div>

    <!--
			Mod config panel: always mounted, parked off-screen to the right.
			The transition class only exists while a mod is selected, so:
			  game -> mod : translate-x-full -> translate-x-0 animates (transition appears with the change)
			  mod  -> mod : transform doesn't change, content just swaps, no animation
			  mod  -> game: transition class is removed in the same change, so it snaps closed
		-->
    <div
      inert={!selectedMod}
      aria-hidden={!selectedMod}
      class="bg-card absolute inset-0 z-10 flex flex-col {selectedMod
        ? 'translate-x-0 transition-transform duration-300 ease-out motion-reduce:transition-none'
        : 'translate-x-full'}"
    >
      {#if selectedMod}
        {#key selectedMod.id}
          <ModConfigView
            mod={{
              id: selectedMod.id,
              category: selectedMod.submission.category,
              author: selectedMod.submission.author,
              name: selectedMod.submission.name,
              link: selectedMod.submission.profileUrl,
              note: selectedMod.submission.description,
              thumbnail: selectedMod.submission.thumbnailUrl ?? "",
              previews: selectedMod.submission.images.split(","),
              has3D: false,
              enabled: false,
              favorite: false,
              folder: selectedMod.tree.files?.join() ?? "",
            }}
            {folders}
            onclose={onclosemod}
            onunload={() => onunloadmod?.(selectedMod.id)}
            onchange={(patch) => onmodchange?.(selectedMod.id, patch)}
            onopenfolder={() => onopenfolder?.(selectedMod.id)}
            onopenlink={() => onopenlink?.(selectedMod.id)}
          />
        {/key}
      {/if}
    </div>
  </div>

  <LaunchBar
    gameName={config.game.name}
    modCount={activeCount}
    launchVanilla={config.launchVanilla}
    {onvanillachange}
    {onstart}
  />
</aside>
