<script lang="ts">
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import type { PageProps } from "./$types";
  import Keybinds from "./Keybinds.svelte";

  let { data }: PageProps = $props();

  let saved = getSavedModsContext();

  const toolId = "zzmi";
  const names = $derived(
    Object.fromEntries(
      Object.entries(data.characters).flatMap(([_, folder]) =>
        Object.entries(folder.sources ?? {}).map(([provider, categoryId]) => [
          `${provider}:${categoryId}`,
          folder.name,
        ]),
      ),
    ),
  );

  const presetsForEnabledMods = $derived.by(() => {
    const enabledMods = Object.values(saved.mods).filter(
      (mod) => mod.packages?.some((p) => p.enabled) && mod.toolId === toolId,
    );

    return enabledMods.map((mod) => ({
      mod,
      name: names[`${mod.submission.provider}:${mod.submission.category}`],
      presets: Object.values(data.keybinds[mod.id]),
    }));
  });
</script>

<div>
  <Keybinds data={presetsForEnabledMods} toolId="zzmi" />
</div>
