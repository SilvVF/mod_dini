<script lang="ts">
  import { Eye, EyeOff } from "@lucide/svelte/icons";
  import NSFWBadge from "./NSFWBadge.svelte";
  import { cn } from "#lib/utils.ts";
  import { getSettingsContext } from "#lib/contexts/settings.svelte.ts";

  type Props = {
    isNSFW: boolean;
    blurStrength?: number;
    showControls?: boolean;
    disableBlur?: boolean;
    class?: string;
    onToggleVisibility?: (visible: boolean) => void;
    children: import("svelte").Snippet;
  };

  let {
    isNSFW,
    blurStrength = 16,
    showControls = true,
    disableBlur = false,
    class: className = "",
    onToggleVisibility,
    children,
  }: Props = $props();

  let isVisible = $state(false);
  const settings = getSettingsContext();

  function handleToggle(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    isVisible = !isVisible;
    onToggleVisibility?.(isVisible);
  }
</script>

{#if !isNSFW || disableBlur || settings.current<boolean>("show.nsfw.images")}
  <div class={className}>
    {@render children()}
  </div>
{:else}
  <div class={cn("relative", className)}>
    <div
      class={cn(
        "h-full w-full",
        !isVisible && "will-change-[filter] contain-[paint]",
      )}
      style={!isVisible ? `filter: blur(${blurStrength}px)` : undefined}
    >
      {@render children()}
    </div>

    {#if !isVisible}
      <div
        class="absolute inset-0 flex items-center justify-center bg-black/40"
      >
        <div class="flex flex-col items-center gap-3">
          <NSFWBadge class="px-3 py-1 text-xs" />

          {#if showControls}
            <button
              type="button"
              onclick={handleToggle}
              class="group flex items-center gap-1.5 rounded-full border border-white/15 bg-white/5 px-4 py-1.5 text-xs font-medium text-white/80 backdrop-blur-xl transition-all duration-300 hover:border-white/30 hover:bg-white/15 hover:text-white hover:shadow-[0_0_16px_rgba(255,255,255,0.08)]"
            >
              <Eye
                class="h-3.5 w-3.5 transition-transform duration-300 group-hover:scale-110"
              />
              Show
            </button>
          {/if}
        </div>
      </div>
    {/if}

    {#if isVisible && showControls}
      <div class="absolute left-2 top-2">
        <button
          type="button"
          onclick={handleToggle}
          class="group flex items-center gap-1.5 rounded-full border border-white/15 bg-black/50 px-3 py-1.5 text-xs font-medium text-white/70 backdrop-blur-xl transition-all duration-300 hover:border-white/25 hover:bg-black/70 hover:text-white"
        >
          <EyeOff
            class="h-3.5 w-3.5 transition-transform duration-300 group-hover:scale-110"
          />
          Hide
        </button>
      </div>
    {/if}
  </div>
{/if}
