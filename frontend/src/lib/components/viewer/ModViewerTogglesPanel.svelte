<script lang="ts">
  import type { ViewerToggle } from "#lib/bindings/changeme/modpreview/models.ts";
  import { Sliders, X } from "@lucide/svelte";

  interface Props {
    toggles: ViewerToggle[];
    toggleValues: Record<string, number>;
    handleSetToggleValue: (variable: string, val: number) => void;
    handleCycleToggle: (tog: ViewerToggle) => void;
    variant?: "sidebar" | "floating";
    onCloseFloating?: () => void;
  }

  let {
    toggles,
    toggleValues,
    handleSetToggleValue,
    handleCycleToggle,
    variant = "sidebar",
    onCloseFloating,
  }: Props = $props();

  function getCurrentValue(toggle: ViewerToggle): number {
    return toggleValues[toggle.variable.toLowerCase()] ?? toggle.currentValue;
  }
</script>

{#if variant === "floating"}
  <div
    class="flex max-h-80 w-80 flex-col overflow-hidden rounded-2xl border border-purple-500/30 bg-[#161b22]/95 shadow-2xl backdrop-blur-xl animate-in fade-in slide-in-from-bottom-2 duration-150 sm:w-96"
  >
    <div
      class="flex items-center justify-between border-b border-white/10 bg-white/[0.03] px-3 py-2"
    >
      <div
        class="flex items-center gap-1.5 text-xs font-semibold text-purple-300"
      >
        <Sliders class="h-3.5 w-3.5" />
        <span>Outfit Variations & Toggles ({toggles.length})</span>
      </div>

      {#if onCloseFloating}
        <button
          type="button"
          onclick={onCloseFloating}
          class="cursor-pointer rounded p-1 text-white/50 transition-colors hover:bg-white/10 hover:text-white"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      {/if}
    </div>

    <div class="max-h-64 flex-1 space-y-2 overflow-y-auto p-2.5">
      {#each toggles as tog (tog.id)}
        {@const curVal = getCurrentValue(tog)}

        <div class="rounded-lg border border-white/10 bg-white/5 p-2">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <span class="truncate text-xs font-medium text-white">
              {tog.name}
            </span>

            {#if tog.key}
              <span
                class="rounded border border-purple-500/30 bg-purple-500/20 px-1.5 py-0.5 text-[10px] font-mono font-bold text-purple-300"
              >
                [{tog.key.toUpperCase()}]
              </span>
            {/if}
          </div>

          <div class="flex flex-wrap items-center gap-1">
            {#each tog.values as value, idx (`${tog.id}-${value}`)}
              {@const isActive = curVal === value}
              {@const label = tog.labels?.[idx] ?? `Style ${idx + 1}`}

              <button
                type="button"
                onclick={() => handleSetToggleValue(tog.variable, value)}
                class={`cursor-pointer rounded px-2 py-0.5 text-[11px] font-medium transition-all ${
                  isActive
                    ? "bg-purple-600 font-semibold text-white shadow-sm"
                    : "bg-white/10 text-white/70 hover:bg-white/20 hover:text-white"
                }`}
              >
                {label}
              </button>
            {/each}

            <button
              type="button"
              onclick={() => handleCycleToggle(tog)}
              class="cursor-pointer rounded bg-white/10 px-2 py-0.5 text-[11px] font-medium text-purple-300 transition-colors hover:bg-purple-500/25 hover:text-white"
              title="Cycle"
            >
              ↻
            </button>
          </div>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <div class="flex flex-1 flex-col overflow-hidden">
    <div class="border-b border-white/10 bg-white/[0.02] px-3 py-2">
      <span class="text-xs font-semibold text-white/80">
        Outfit Variations & Toggles
      </span>

      <p class="mt-0.5 text-[10px] text-white/40">
        Press hotkey or click buttons to cycle
      </p>
    </div>

    <div class="flex-1 space-y-2 overflow-y-auto p-2">
      {#each toggles as tog (tog.id)}
        {@const curVal = getCurrentValue(tog)}

        <div class="rounded-lg border border-white/10 bg-white/5 p-2.5">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <span class="truncate text-xs font-medium text-white">
              {tog.name}
            </span>

            {#if tog.key}
              <span
                class="rounded border border-purple-500/30 bg-purple-500/20 px-1.5 py-0.5 text-[10px] font-mono font-bold text-purple-300"
              >
                [{tog.key.toUpperCase()}]
              </span>
            {/if}
          </div>

          <div class="flex flex-wrap items-center gap-1">
            {#each tog.values as value, idx (`${tog.id}-${value}`)}
              {@const isActive = curVal === value}
              {@const label = tog.labels?.[idx] ?? `Style ${idx + 1}`}

              <button
                type="button"
                onclick={() => handleSetToggleValue(tog.variable, value)}
                class={`cursor-pointer rounded px-2.5 py-1 text-xs font-medium transition-all ${
                  isActive
                    ? "bg-purple-600 font-semibold text-white shadow-sm"
                    : "bg-white/10 text-white/70 hover:bg-white/20 hover:text-white"
                }`}
              >
                {label}
              </button>
            {/each}

            <button
              type="button"
              onclick={() => handleCycleToggle(tog)}
              class="cursor-pointer rounded bg-white/10 px-2 py-1 text-xs font-medium text-purple-300 transition-colors hover:bg-purple-500/25 hover:text-white"
              title="Cycle"
            >
              ↻
            </button>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}
