<script lang="ts">
  import type {
    ModKeyBindPreset,
    ModSaved,
  } from "#lib/bindings/changeme/models.ts";

  type ModKeyBind = NonNullable<ModKeyBindPreset["values"]>[number];

  type Props = {
    data: {
      mod: ModSaved;
      name: string;
      presets: ModKeyBindPreset[];
    }[];
    toolId: string;
    title?: string;
  };

  let { data, toolId, title = "Keybinds" }: Props = $props();
</script>

<div
  class="flex max-h-screen min-h-0 flex-col overflow-hidden rounded-2xl border border-border bg-card text-card-foreground shadow-xl"
>
  <!-- Header -->
  <div
    class="flex shrink-0 items-center gap-3 border-b border-border px-4 py-3"
  >
    <h2 class="truncate text-sm font-bold">{title}</h2>

    <span
      class="rounded-full bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground"
    >
      {data.length}
    </span>

    <div class="flex-1"></div>

    <span class="text-xs text-muted-foreground">
      {toolId}
    </span>
  </div>

  <!-- Preset list -->
  <div class="min-h-0 flex-1 space-y-3 overflow-y-scroll p-3">
    {#each data as item}
      <div class="rounded-xl border border-border bg-muted/40 p-3">
        <div class="relative">
          <span
            class="float-right ml-3 max-w-[50%] pt-0.5 text-right text-[10px] font-semibold uppercase tracking-wide text-muted-foreground"
          >
            {item.name}
          </span>

          <h3 class="text-sm font-bold leading-snug">
            {item.mod.submission.name}
          </h3>
        </div>
        <!-- Bind rows -->
        {#if item.presets.length > 0}
          <div class="space-y-1.5">
            {#each item.presets as preset ((preset.modId, preset.id))}
              {#each preset.values as bind, i ((bind.section, i))}
                <div class="flex min-h-7 items-center justify-between gap-3">
                  <span class="min-w-0 truncate text-sm text-foreground/90">
                    {bind.section.startsWith("Key")
                      ? bind.section.slice(3, bind.section.length)
                      : bind.section}
                  </span>

                  <div class="flex shrink-0 flex-wrap justify-end gap-1">
                    {#each bind.key.split(" ") as key, j (`${j}-${key}`)}
                      {#if key}
                        <kbd
                          class="rounded-md border border-border bg-background/70 px-2 py-0.5 text-xs font-semibold leading-4 shadow-sm"
                        >
                          {key}
                        </kbd>
                      {/if}
                    {/each}
                  </div>
                </div>
              {/each}
            {/each}
          </div>
        {:else}
          <p class="py-2 text-center text-xs text-muted-foreground">
            No keybinds available
          </p>
        {/if}
      </div>
    {:else}
      <div
        class="flex min-h-32 items-center justify-center text-sm text-muted-foreground"
      >
        No keybind presets found for this tool.
      </div>
    {/each}
  </div>
</div>
