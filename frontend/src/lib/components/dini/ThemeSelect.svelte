<script lang="ts">
  import { Palette, Plus, Trash2 } from "@lucide/svelte";
  import { Button } from "#lib/components/ui/button/index.ts";
  import * as DropdownMenu from "#lib/components/ui/dropdown-menu/index.ts";
  import { appTheme, type ThemeId } from "#lib/state/theme.svelte.ts";
  import ThemeImportDialog from "./ThemeImportDialog.svelte";
 
  let importOpen = $state(false);
 
  const currentLabel = $derived(
    appTheme.all.find((t) => t.id === appTheme.current)?.label,
  );
 
  // Closing the menu without picking restores the saved theme.
  function handleOpenChange(open: boolean) {
    if (!open) appTheme.clearPreview();
  }
</script>
 
<DropdownMenu.Root onOpenChange={handleOpenChange}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button {...props} variant="outline" size="sm" class="h-8 gap-2">
        <Palette class="size-4" />
        {currentLabel}
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
 
  <DropdownMenu.Content align="end" class="w-52">
    <DropdownMenu.Label>Theme</DropdownMenu.Label>
    <DropdownMenu.Separator />
 
    <DropdownMenu.RadioGroup
      value={appTheme.current}
      onValueChange={(v) => appTheme.set(v as ThemeId)}
    >
      {#each appTheme.all as t (t.id)}
        <DropdownMenu.RadioItem
          value={t.id}
          onpointerenter={() => appTheme.preview(t.id)}
          onfocus={() => appTheme.preview(t.id)}
        >
          <span class="truncate">{t.label}</span>
        </DropdownMenu.RadioItem>
      {/each}
    </DropdownMenu.RadioGroup>
 
    <DropdownMenu.Separator />
 
    <DropdownMenu.Item
      class="gap-2"
      onpointerenter={() => appTheme.clearPreview()}
      onSelect={() => (importOpen = true)}
    >
      <Plus class="size-4" />
      Import theme…
    </DropdownMenu.Item>
 
    {#if appTheme.custom.length > 0}
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger class="gap-2">
          <Trash2 class="size-4" />
          Remove theme
        </DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent class="w-48">
          {#each appTheme.custom as t (t.id)}
            <DropdownMenu.Item onSelect={() => appTheme.remove(t.id)}>
              <span class="truncate">{t.label}</span>
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
 
<ThemeImportDialog bind:open={importOpen} />
 