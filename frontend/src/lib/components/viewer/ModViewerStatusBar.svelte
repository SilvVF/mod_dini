
<script lang="ts">
  import { Loader2, Sparkles, X } from '@lucide/svelte';
    import type { RetextureInfo } from './types';

    interface Props {
        retextureInfo: RetextureInfo;
        onOpenBaseModelDropdown: () => void;
        loading: boolean;
        error: string | null;
        onClose: () => void;
    }

    let {
        retextureInfo,
        onOpenBaseModelDropdown,
        loading,
        error,
        onClose
    }: Props = $props();

    const baseModelName = $derived(
        retextureInfo.baseModelSource?.replace(/^DISABLED_?/, '') ?? ''
    );
</script>

{#if retextureInfo.isRetexture && retextureInfo.baseModelSource}
    <div class="absolute left-3 top-13 z-10 flex items-center gap-2 rounded-lg border border-amber-500/30 bg-amber-950/70 px-3 py-1.5 shadow-lg backdrop-blur-md">
        <Sparkles class="h-4 w-4 shrink-0 text-amber-400" />

        <div class="text-xs text-amber-200">
            <span class="font-semibold">Retexture based on:</span>
            {' '}{baseModelName}
        </div>

        <button
            type="button"
            onclick={onOpenBaseModelDropdown}
            class="ml-1 cursor-pointer text-[11px] font-medium text-amber-400 underline transition-colors hover:text-amber-300"
        >
            Change Base Model
        </button>
    </div>
{/if}

{#if loading}
    <div class="absolute inset-0 z-20 flex flex-col items-center justify-center gap-3 bg-[#0d1117]/90 p-6 text-center backdrop-blur-sm">
        <Loader2 class="h-8 w-8 animate-spin text-purple-400" />

        <div class="space-y-1">
            <span class="text-sm font-semibold text-white/90">
                Loading Mod...
            </span>

            <p class="max-w-xs text-xs text-white/50">
                Please wait while the mod preview is being prepared.
            </p>
        </div>

        <button
            type="button"
            onclick={onClose}
            class="mt-3 flex items-center gap-1.5 rounded-lg border border-white/20 bg-white/5 px-4 py-1.5 text-xs font-semibold text-white/80 shadow-lg transition-colors hover:bg-white/10 hover:text-white"
        >
            <X class="h-3.5 w-3.5" />
            Cancel
        </button>
    </div>
{:else if error}
    <div class="absolute inset-0 z-20 flex flex-col items-center justify-center gap-3 bg-[#0d1117]/90">
        <div class="max-w-md rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-center">
            <p class="text-sm font-medium text-red-400">
                Failed to Load Mod
            </p>

            <p class="mt-1 text-xs text-red-400/70">
                {error}
            </p>
        </div>
    </div>
{/if}
