<script lang="ts">
  import { Play } from "@lucide/svelte";
  import Checkbox from "../ui/checkbox/checkbox.svelte";
  import { Label } from "../ui/label";

  let {
    gameName,
    modCount,
    launchVanilla,
    onvanillachange,
    onstart,
  }: {
    gameName: string;
    modCount: number;
    launchVanilla: boolean;
    onvanillachange?: (value: boolean) => void;
    onstart?: () => void;
  } = $props();
</script>

<div class="space-y-3 px-4 pt-3 pb-4">
  <button
    type="button"
    onclick={onstart}
    class="focus-visible:ring-ring ring-offset-card flex h-16 w-full items-center justify-between bg-linear-to-r from-violet-500 to-indigo-500 px-7 text-white outline-none transition hover:brightness-110 focus-visible:ring-2 focus-visible:ring-offset-2 active:brightness-95"
    style="clip-path: polygon(5% 0, 100% 0, 95% 100%, 0 100%);"
  >
    <span class="flex flex-1 flex-col items-center">
      <span class=" text-2xl leading-none font-bold tracking-[0.2em]"
        >START</span
      >
      <span class="mt-1 text-xs opacity-80">
        {gameName} · {launchVanilla
          ? "vanilla"
          : `${modCount} ${modCount === 1 ? "mod" : "mods"}`}
      </span>
    </span>
    <Play class="size-5 fill-current" />
  </button>

  <div class="flex items-center gap-2 px-1">
    <Checkbox
      id="launch-vanilla"
      checked={launchVanilla}
      onCheckedChange={(v) => onvanillachange?.(v === true)}
      class="rounded-full"
    />
    <Label for="launch-vanilla" class="font-semibold"
      >Launch vanilla (no mods this run)</Label
    >
  </div>
</div>
