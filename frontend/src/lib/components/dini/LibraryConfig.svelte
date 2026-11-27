<script lang="ts">
  import type { SidebarProps } from "./types.js";
  import GameConfigView from "./GameConfigView.svelte";
  import ModConfigView from "./ModConfigView.svelte";
  import LaunchBar from "./LaunchBar.svelte";
  import { getProcessContext } from "#lib/contexts/process.svelte.js";
  import { tools } from "#lib/state/app.svelte.js";

  let {
    config,
    selectedMod = null,
    onselectmod,
    onclosemod,
    onopenfolder,
    onopenlink,
    onvanillachange,
    onTogglePack,
    onSelectPreset,
    onstart,
  }: SidebarProps = $props();

  const activeCount = $derived(config.launchVanilla ? 0 : config.stats.loaded);
  const processState = getProcessContext();
  
  const process = $derived(tools.find(t => t.id === config.game.id)?.process)
  const gameRunning = $derived(process ? processState.running(process) : undefined)
</script>

<aside
  class="bg-card text-card-foreground flex h-full w-full flex-col overflow-hidden"
>
  <div class="relative flex min-h-0 flex-1 flex-col overflow-hidden">
    <!-- Game config is always rendered -->
    <div class="flex min-h-0 flex-1 flex-col" inert={!!selectedMod}>
      <GameConfigView {config} {onselectmod} onselectpreset={onSelectPreset} />
    </div>
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
            mod={selectedMod}
            onTogglePack={(pack) => onTogglePack(selectedMod, pack)}
            onclose={onclosemod}
            onopenfolder={() => onopenfolder?.(selectedMod)}
            onopenlink={() => onopenlink?.(selectedMod)}
          />
        {/key}
      {/if}
    </div>
  </div>

  <LaunchBar
    gameName={config.game.name}
    modCount={activeCount}
    running={gameRunning}
    launchVanilla={config.launchVanilla}
    {onvanillachange}
    {onstart}
  />
</aside>
