<script lang="ts">
  import { Grid3x3, RotateCcw, RotateCw, Zap } from "@lucide/svelte";
  import type { RenderMode, LightingPreset } from "./types";

  interface Props {
    meshCount: number;
    loading: boolean;
    error: string | null;
    renderMode: RenderMode;
    handleSetRenderMode: (mode: RenderMode) => void;
    lightingPreset: LightingPreset;
    handleSetLighting: (preset: LightingPreset) => void;
    turntableActive: boolean;
    toggleTurntable: () => void;
    wireframe: boolean;
    toggleWireframe: () => void;
    rotateModelY: (rad: number) => void;
    fitCamera: () => void;
    potatoMode: boolean;
    togglePotatoMode: () => void;
  }

  let {
    meshCount,
    loading,
    error,
    renderMode,
    handleSetRenderMode,
    lightingPreset,
    handleSetLighting,
    turntableActive,
    toggleTurntable,
    wireframe,
    toggleWireframe,
    rotateModelY,
    fitCamera,
    potatoMode,
    togglePotatoMode,
  }: Props = $props();

  const renderModes: RenderMode[] = ["pbr", "diffuse", "clay", "normal"];
  const lightingPresets: LightingPreset[] = ["studio", "anime", "dramatic"];

  function getRenderModeLabel(mode: RenderMode): string {
    switch (mode) {
      case "normal":
        return "Normals";
      case "diffuse":
        return "Anime";
      default:
        return mode.toUpperCase();
    }
  }
</script>

{#if meshCount > 0 && !loading && !error}
  <div
    class="absolute top-3 right-3 z-20 flex max-w-[calc(100%-24px)] flex-wrap items-center justify-end gap-1 rounded-xl border border-white/15 bg-[#161b22]/95 p-1 shadow-2xl backdrop-blur-md"
  >
    <div
      class="flex items-center rounded-lg border border-white/10 bg-white/5 p-0.5"
    >
      {#each renderModes as mode}
        <button
          type="button"
          onclick={() => handleSetRenderMode(mode)}
          class={`flex cursor-pointer items-center gap-1 rounded px-2 py-0.5 text-xs font-medium transition-all ${
            renderMode === mode
              ? "bg-purple-600 font-semibold text-white shadow-sm"
              : "text-white/60 hover:bg-white/10 hover:text-white"
          }`}
          title={`mod_viewer_mode_${mode}`}
        >
          {getRenderModeLabel(mode)}
        </button>
      {/each}
    </div>

    <div
      class="flex items-center rounded-lg border border-white/10 bg-white/5 p-0.5"
    >
      {#each lightingPresets as preset}
        <button
          type="button"
          onclick={() => handleSetLighting(preset)}
          class={`cursor-pointer rounded px-1.5 py-0.5 text-xs font-medium capitalize transition-all ${
            lightingPreset === preset
              ? "bg-white/15 font-semibold text-white"
              : "text-white/50 hover:bg-white/10 hover:text-white/80"
          }`}
          title={`mod_viewer_lighting_${preset}`}
        >
          {preset}
        </button>
      {/each}
    </div>

    <div class="mx-0.5 hidden h-4 w-px bg-white/15 sm:block"></div>

    <button
      type="button"
      onclick={toggleTurntable}
      class={`flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium transition-all ${
        turntableActive
          ? "bg-purple-600 text-white shadow-sm"
          : "text-white/60 hover:bg-white/10 hover:text-white"
      }`}
      title={"mod_viewer_turntable Turntable (Auto-Spin) [T]"}
    >
      <RotateCw
        class={`h-3.5 w-3.5 ${turntableActive ? "animate-spin" : ""}`}
      />
      <span class="hidden sm:inline">
        {"mod_viewer_turntable Turntable"}
      </span>
    </button>

    <button
      type="button"
      onclick={toggleWireframe}
      class={`cursor-pointer rounded-lg p-1.5 transition-colors ${
        wireframe
          ? "bg-purple-500/20 text-purple-400"
          : "text-white/60 hover:bg-white/10 hover:text-white"
      }`}
      title={"mod_viewer_wireframe"}
    >
      <Grid3x3 class="h-4 w-4" />
    </button>

    <button
      type="button"
      onclick={() => rotateModelY(Math.PI / 2)}
      class="cursor-pointer rounded-lg p-1.5 text-white/60 transition-colors hover:bg-white/10 hover:text-white"
      title={"mod_viewer_rotate"}
    >
      <RotateCw class="h-4 w-4" />
    </button>

    <button
      type="button"
      onclick={fitCamera}
      class="cursor-pointer rounded-lg p-1.5 text-white/60 transition-colors hover:bg-white/10 hover:text-white"
      title={"mod_viewer_reset_camera"}
    >
      <RotateCcw class="h-4 w-4" />
    </button>

    <div class="mx-0.5 h-4 w-px bg-white/15"></div>

    <button
      type="button"
      onclick={togglePotatoMode}
      class={`flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium transition-all ${
        potatoMode
          ? "border border-amber-500/40 bg-amber-500/25 text-amber-300 shadow-sm"
          : "text-white/60 hover:bg-white/10 hover:text-white"
      }`}
      title={potatoMode
        ? "Potato Mode ON (Fast Preview, 512px diffuse only)"
        : "Switch to Potato Mode (Fast Preview)"}
    >
      <Zap
        class={`h-3.5 w-3.5 ${
          potatoMode ? "fill-amber-400/30 text-amber-400" : ""
        }`}
      />
      <span class="hidden sm:inline">
        {potatoMode ? "Potato" : "Full Quality"}
      </span>
    </button>
  </div>
{/if}
