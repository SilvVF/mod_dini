
<script lang="ts">
import { Compass, User, RotateCcw } from '@lucide/svelte';
import type { CameraPreset } from './types';

interface Props {
  meshCount: number;
  loading: boolean;
  error: string | null;
  activeCameraPreset: CameraPreset | null;
  handleSelectCameraPreset: (preset: CameraPreset) => void;
}

let {
  meshCount,
  loading,
  error,
  activeCameraPreset,
  handleSelectCameraPreset,
}: Props = $props();

const presets: CameraPreset[] = [
  'front',
  'back',
  'left',
  'right',
  'face',
  'reset',
];
</script>

{#if meshCount > 0 && !loading && !error}
  <div class="absolute bottom-3 left-3 z-10 flex items-center rounded-xl border border-white/15 bg-[#161b22]/90 p-1 backdrop-blur-md shadow-2xl gap-1">
    <span class="px-2 text-[10px] font-semibold uppercase tracking-wider text-white/40 flex items-center gap-1">
      <Compass class="h-3 w-3 text-purple-400" />
      <span class="hidden sm:inline">Angle</span>
    </span>

    {#each presets as preset (preset)}
      <button
        onclick={() => handleSelectCameraPreset(preset)}
        class={`rounded-md px-2 py-0.5 text-[11px] font-medium transition-all cursor-pointer ${
          activeCameraPreset === preset && preset !== 'reset'
            ? 'bg-purple-600 text-white font-semibold shadow-sm'
            : 'text-white/60 hover:bg-white/10 hover:text-white'
        }`}
        title={preset === 'face' ? 'Face' : preset === 'reset' ? 'Reset' : preset}
      >
        {#if preset === 'face'}
          <span class="flex items-center gap-1">
            <User class="h-2.5 w-2.5" />
            Face
          </span>
        {:else if preset === 'reset'}
          <span class="flex items-center gap-1">
            <RotateCcw class="h-2.5 w-2.5" />
            Reset
          </span>
        {:else}
          {preset}
        {/if}
      </button>
    {/each}
  </div>
{/if}
