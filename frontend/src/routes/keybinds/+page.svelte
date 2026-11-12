<script lang="ts">
    import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
    import type { PageProps } from "../$types";
    import Keybinds from "./Keybinds.svelte";

    let { data }: PageProps = $props();

    let saved = getSavedModsContext();

    const toolId = "zzmi";
    const names = $derived(
        Object.fromEntries(
            Object.entries(data.characters).flatMap(([_, folder]) =>
                Object.entries(folder.sources ?? {}).map(
                    ([provider, categoryId]) => [
                        `${provider}:${categoryId}`,
                        folder.name,
                    ],
                ),
            ),
        ),
    );

    const groupedPresets = $derived.by(() =>
        Object.entries(data.keybinds).flatMap(([modId, presets]) => {
            const mod = saved.mods[modId];
            const categoryKey = `${mod.submission.provider}:${mod.submission.category}`;
            return [
                {
                    mod,
                    name: names[categoryKey] ?? "",
                    presets: Object.values(presets),
                },
            ];
        }),
    );
</script>

<div>
    <Keybinds data={groupedPresets} toolId="zzmi" />
</div>
