<script lang="ts">
  import type {
    Setting,
  } from "#lib/bindings/changeme/models.ts";
  import type { SettingChange } from "#lib/contexts/settings.svelte.ts";
  import { Button } from "../ui/button";
  import SettingControl from "./SettingControl.svelte";
  import { SettingsMeta } from "./SettingsMeta";

  let {
    setting,
    value,
    onchange,
    onaction,
  }: {
    setting: Setting;
    value: unknown;
    onchange: (value: SettingChange) => void;
    onaction?: (actionId: string) => void;
  } = $props();

  const Icon = $derived(SettingsMeta.resolveIcon(setting.icon));
</script>

<div class="flex items-center gap-4 px-4 py-4">
  <span
    class="bg-primary/15 text-primary flex size-10 shrink-0 items-center justify-center rounded-lg"
  >
    <Icon class="size-5" />
  </span>

  <div class="min-w-0 flex-1">
    <label
      for="setting-{setting.key}"
      class="block  text-base leading-snug font-bold"
    >
      {setting.title}
    </label>
    {#if setting.description}
      <p class="text-muted-foreground mt-0.5 text-sm leading-snug">
        {setting.description}
      </p>
    {/if}
  </div>

  <div class="flex shrink-0 items-center gap-2">
    <SettingControl {setting} {value} {onchange} />
    {#if setting.actions}
      {#each setting.actions as action}
        <Button
          variant="outline"
          class="font-semibold"
          onclick={() => onaction?.(action.id)}
        >
          {action.label}
        </Button>
      {/each}
    {/if}
  </div>
</div>
