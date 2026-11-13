<script lang="ts">
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import { appState } from "#lib/state/app.svelte.ts";
  import { compact } from "#lib/ts-ext.ts";
  import type { PageProps } from "./$types";
  import Keybinds from "./Keybinds.svelte";

  let { data }: PageProps = $props();

  let saved = getSavedModsContext();

  const characters = $derived(
    compact(data.rootFolder[appState.activeTool].characters.characters ?? {}),
  );
  const names = $derived(
    Object.fromEntries(
      Object.entries(characters).flatMap(([_, folder]) =>
        Object.entries(folder.sources ?? {}).map(([provider, categoryId]) => [
          `${provider}:${categoryId}`,
          folder.name,
        ]),
      ),
    ),
  );

  const presetsForEnabledMods = $derived.by(() => {
    const enabledMods = Object.values(saved.mods).filter(
      (mod) =>  mod.toolId === appState.activeTool && mod.packages?.some((p) => p.enabled),
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
