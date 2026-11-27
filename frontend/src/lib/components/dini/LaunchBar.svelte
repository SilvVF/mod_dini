<script lang="ts">
  import { Play, Square } from "@lucide/svelte";
  import { Button } from "#lib/components/ui/button/index.ts";
  import Checkbox from "#lib/components/ui/checkbox/checkbox.svelte";
  import { Label } from "#lib/components/ui/label/index.ts";

  let {
    gameName,
    modCount,
    launchVanilla,
    running = false,
    onvanillachange,
    onstart,
    onstop,
  }: {
    gameName: string;
    modCount: number;
    launchVanilla: boolean;
    running?: boolean;
    onvanillachange?: (value: boolean) => void;
    onstart?: () => void;
    onstop?: () => void;
  } = $props();

  const summary = $derived(
    launchVanilla
      ? "vanilla"
      : `${modCount} ${modCount === 1 ? "mod" : "mods"}`,
  );
</script>

<div class="flex flex-col gap-3 border-t px-4 py-4">
  {#if running}
    <div
      class="bg-muted/30 flex h-14 w-full items-center justify-between gap-3 rounded-xl border px-4"
      role="status"
      aria-live="polite"
    >
      <div class="flex min-w-0 items-center gap-3">
        <span class="relative flex size-2.5 shrink-0">
          <span
            class="absolute inline-flex size-full animate-ping rounded-full bg-emerald-500 opacity-75"
          ></span>
          <span
            class="relative inline-flex size-2.5 rounded-full bg-emerald-500"
          ></span>
        </span>
        <span class="flex min-w-0 flex-col gap-0.5">
          <span class="text-sm leading-none font-semibold">Running</span>
          <span class="text-muted-foreground truncate text-xs">
            {gameName} · {summary}
          </span>
        </span>
      </div>

      <Button
        type="button"
        variant="destructive"
        size="sm"
        class="shrink-0 gap-1.5"
        onclick={onstop}
      >
        <Square class="size-3.5 fill-current" />
        Stop
      </Button>
    </div>
  {:else}
    <Button
      type="button"
      size="lg"
      class="h-14 w-full justify-between rounded-xl px-5"
      onclick={onstart}
    >
      <span class="flex flex-col items-start gap-0.5">
        <span class="text-lg leading-none font-semibold tracking-widest">
          START
        </span>
        <span class="text-primary-foreground/70 text-xs font-normal">
          {gameName} · {summary}
        </span>
      </span>
      <Play class="size-5 fill-current" />
    </Button>
  {/if}

  <div class="flex items-center gap-2 px-1">
    <Checkbox
      id="launch-vanilla"
      checked={launchVanilla}
      disabled={running}
      onCheckedChange={(v) => onvanillachange?.(v === true)}
    />
    <Label
      for="launch-vanilla"
      class="text-sm font-medium {running ? 'opacity-50' : ''}"
    >
      Launch vanilla (no mods this run)
    </Label>
  </div>
</div>
