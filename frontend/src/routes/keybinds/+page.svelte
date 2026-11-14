<script lang="ts">
  import Keybinds from "#lib/components/dini/Keybinds.svelte";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import { appState } from "#lib/state/app.svelte.ts";
  import { compact } from "#lib/ts-ext.ts";
  import type { PageProps } from "./$types";
  let { data }: PageProps = $props();

  let saved = getSavedModsContext();

  const characters = $derived(
    compact(data.rootFolder[appState.activeTool].characters.characters ?? {}),
  );
  const names = $derived.by(() => {
    const nameMap: { [key: string]: string } = {};
    for (const folder of Object.values(characters)) {
      for (const [provider, categoryId] of Object.entries(folder.sources ?? {})) {
        nameMap[`${provider}:${categoryId}`] = folder.name;
      }
    }
    return nameMap;
  });

  const presetsForEnabledMods = $derived.by(() => {
    const enabledMods = Object.values(saved.mods).filter(
      (mod) =>
        mod.toolId === appState.activeTool &&
        mod.packages?.some((p) => p.enabled),
    );

    return enabledMods.map((mod) => ({
      mod,
      name: names[`${mod.submission.provider}:${mod.submission.category}`],
      presets: Object.values(data.keybinds[mod.id]),
    }));
  });
</script>

<div>
  <Keybinds data={presetsForEnabledMods} toolId={appState.activeTool} />
</div>
