
<script lang="ts">
  import { Camera, Check, ChevronDown, FolderOpen, Loader2, Sparkles, X } from '@lucide/svelte';
    import type { RetextureInfo } from './types';

    interface Props {
        displayName: string;
        modPath?: string;
        isGbPreview: boolean;
        embedded?: boolean;
        retextureInfo: RetextureInfo;
        showBaseModelDropdown: boolean;
        setShowBaseModelDropdown: (show: boolean) => void;
        loadMod: (folderName: string) => void;
        handleBrowseBaseModel: () => void;
        onInstallGbMod?: () => Promise<void> | void;
        installingGb: boolean;
        setInstallingGb: (val: boolean) => void;
        gbInstallSuccess: boolean;
        setGbInstallSuccess: (val: boolean) => void;
        onClose: () => void;
    }

    let {
        displayName,
        modPath,
        isGbPreview,
        embedded = false,
        retextureInfo,
        showBaseModelDropdown,
        setShowBaseModelDropdown,
        loadMod,
        handleBrowseBaseModel,
        onInstallGbMod,
        installingGb,
        setInstallingGb,
        gbInstallSuccess,
        setGbInstallSuccess,
        onClose
    }: Props = $props();

    function t(key: string, fallback?: string): string {
        return fallback ?? key;
    }

    function getConfidenceLabel(candidate: RetextureInfo['candidates'][number]): string {
        switch (candidate.confidence) {
            case 'exact':
                return `${candidate.matchedBuffersCount}/${candidate.totalRequiredBuffers}`;
            case 'partial':
                return `${candidate.matchedBuffersCount}/${candidate.totalRequiredBuffers}`;
            default:
                return `${candidate.matchedBuffersCount} matches`;
        }
    }

    async function installGbMod() {
        if (installingGb || gbInstallSuccess || !onInstallGbMod) return;

        try {
            setInstallingGb(true);
            await onInstallGbMod();
            setGbInstallSuccess(true);
        } catch (error) {
            console.error('Failed to install from preview:', error);
        } finally {
            setInstallingGb(false);
        }
    }

    function useAsPreview() {
     
    }
</script>

<div class="absolute top-0 left-0 right-0 z-30 flex h-10 items-center justify-between border-b border-white/10 bg-[#0d1117]/90 px-3 backdrop-blur-md">
    <div class="flex min-w-0 items-center gap-2">
        <span class="shrink-0 text-xs font-semibold uppercase tracking-wider text-purple-400">
            {t('mod_viewer_title', 'Mod Viewer')}
        </span>

        <span class="text-white/40">•</span>

        <span
            class="truncate text-xs font-medium text-white/80"
            title={displayName}
        >
            {displayName}
        </span>
    </div>

    <div class="flex shrink-0 items-center gap-2">
        {#if retextureInfo.isRetexture}
            <div class="relative">
                <button
                    type="button"
                    onclick={() => setShowBaseModelDropdown(!showBaseModelDropdown)}
                    class={`flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs font-medium transition-colors ${
                        retextureInfo.baseModelSource
                            ? 'border-amber-500/40 bg-amber-500/10 text-amber-300 hover:bg-amber-500/20'
                            : 'border-white/10 bg-white/5 text-white/60 hover:bg-white/10'
                    }`}
                    title={t('mod_viewer_base_model_label', 'Base Model')}
                >
                    <Sparkles class="h-3.5 w-3.5 text-amber-400" />

                    <span class="max-w-140px truncate">
                        {retextureInfo.baseModelSource
                            ? retextureInfo.baseModelSource.replace(/^DISABLED_?/, '')
                            : t('mod_viewer_select_base_model', 'Select Base Model')}
                    </span>

                    <ChevronDown class="h-3 w-3 text-white/40" />
                </button>

                {#if showBaseModelDropdown}
                    <div class="absolute right-0 top-full z-50 mt-1 w-72 rounded-xl border border-white/15 bg-[#161b22] p-1.5 shadow-2xl backdrop-blur-xl">
                        <div class="px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-white/40">
                            {t('mod_viewer_candidates_available', 'Compatible models found:')}
                        </div>

                        <button
                            type="button"
                            onclick={() => {
                                setShowBaseModelDropdown(false);
                                loadMod('');
                            }}
                            class={`flex w-full items-start gap-2 rounded-lg px-2.5 py-2 text-left text-xs transition-colors ${
                                !retextureInfo.baseModelSource
                                    ? 'bg-purple-600/20 font-medium text-purple-300'
                                    : 'text-white/70 hover:bg-white/5'
                            }`}
                        >
                            <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-white/30"></span>

                            <div class="min-w-0 flex-1">
                                <div class="flex items-center justify-between">
                                    <span>{t('mod_viewer_base_model_none', 'No Base Model')}</span>

                                    {#if !retextureInfo.baseModelSource}
                                        <Check class="h-3.5 w-3.5 text-purple-400" />
                                    {/if}
                                </div>

                                <div class="text-[10px] text-white/40">
                                    Texture details & clues only
                                </div>
                            </div>
                        </button>

                        {#each retextureInfo.candidates as candidate (candidate.folderName)}
                            {@const isSelected = retextureInfo.baseModelSource === candidate.folderName}

                            <button
                                type="button"
                                onclick={() => {
                                    setShowBaseModelDropdown(false);
                                    loadMod(candidate.folderName);
                                }}
                                class={`mt-1 flex w-full items-start gap-2 rounded-lg px-2.5 py-2 text-left text-xs transition-colors ${
                                    isSelected
                                        ? 'bg-amber-500/20 font-medium text-amber-200'
                                        : 'text-white/80 hover:bg-white/5'
                                }`}
                            >
                                <span
                                    class={`mt-1 h-2 w-2 shrink-0 rounded-full ${
                                        candidate.confidence === 'exact'
                                            ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.6)]'
                                            : candidate.confidence === 'partial'
                                                ? 'bg-amber-400'
                                                : 'bg-white/30'
                                    }`}
                                ></span>

                                <div class="min-w-0 flex-1">
                                    <div class="flex items-center justify-between gap-1">
                                        <span class="truncate">
                                            {candidate.displayName}
                                        </span>

                                        <span
                                            class={`rounded border px-1.5 py-0.5 text-[9px] font-mono font-medium ${
                                                candidate.confidence === 'exact'
                                                    ? 'border-emerald-500/30 bg-emerald-500/20 text-emerald-300'
                                                    : candidate.confidence === 'partial'
                                                        ? 'border-amber-500/30 bg-amber-500/20 text-amber-300'
                                                        : 'border-white/10 bg-white/10 text-white/50'
                                            }`}
                                        >
                                            {getConfidenceLabel(candidate)}
                                        </span>
                                    </div>

                                    <div class="truncate text-[10px] text-white/40">
                                        {candidate.matchDetails}
                                    </div>
                                </div>
                            </button>
                        {/each}

                        <div class="mt-1 border-t border-white/10 pt-1">
                            <button
                                type="button"
                                onclick={() => {
                                    setShowBaseModelDropdown(false);
                                    handleBrowseBaseModel();
                                }}
                                class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-left text-xs text-purple-300 transition-colors hover:bg-purple-500/15"
                            >
                                <FolderOpen class="h-3.5 w-3.5 shrink-0 text-purple-400" />

                                <span class="truncate">
                                    {t('mod_viewer_browse_base_model', 'Browse custom folder...')}
                                </span>
                            </button>
                        </div>
                    </div>
                {/if}
            </div>
        {/if}

        {#if onInstallGbMod}
            <button
                type="button"
                onclick={installGbMod}
                disabled={installingGb || gbInstallSuccess}
                class={`flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-bold shadow-lg transition-all ${
                    gbInstallSuccess
                        ? 'border border-emerald-500/40 bg-emerald-500/20 text-emerald-300'
                        : installingGb
                            ? 'cursor-wait bg-emerald-600/50 text-white'
                            : 'bg-emerald-600 text-white shadow-emerald-950/40 hover:scale-105 hover:bg-emerald-500 active:scale-95'
                }`}
            >
                {#if gbInstallSuccess}
                    <Check class="h-3.5 w-3.5" />
                    {t('mod_viewer_gb_install_success', 'Installed!')}
                {:else if installingGb}
                    <Loader2 class="h-3.5 w-3.5 animate-spin" />
                    {t('mod_viewer_gb_installing', 'Installing...')}
                {:else}
                    <Sparkles class="h-3.5 w-3.5" />
                    {t('mod_viewer_gb_install', 'Install Mod to Library')}
                {/if}
            </button>
        {/if}

        {#if !embedded}
            <button
                type="button"
                onclick={onClose}
                class="cursor-pointer rounded-md p-1.5 text-white/60 transition-colors hover:bg-white/10 hover:text-white"
                title="Close"
            >
                <X class="h-4 w-4" />
            </button>
        {/if}
    </div>
</div>
