<script lang="ts">
  import { Button } from "#lib/components/ui/button/index.ts";
  import * as Dialog from "#lib/components/ui/dialog/index.ts";
  import { Input } from "#lib/components/ui/input/index.ts";
  import { Label } from "#lib/components/ui/label/index.ts";
  import { Textarea } from "#lib/components/ui/textarea/index.ts";
  import { appTheme as theme } from "#lib/state/theme.svelte.ts";

  let { open = $bindable(false) }: { open?: boolean } = $props();

  let name = $state("");
  let css = $state("");
  let error = $state<string | undefined>(undefined);

  const canImport = $derived(name.trim() !== "" && css.trim() !== "");

  function reset() {
    name = "";
    css = "";
    error = undefined;
  }

  function handleImport() {
    if (!canImport) return;
    try {
      theme.add(name, css);
      open = false;
    } catch (e) {
      error = e instanceof Error ? e.message : "Couldn't import that theme.";
    }
  }

  // Start fresh each time the dialog closes.
  $effect(() => {
    if (!open) reset();
  });
</script>

<Dialog.Root bind:open>
  <Dialog.Content
    class="flex flex-col gap-4"
    style="width: min(44rem, calc(100vw - 2rem)); max-width: none; height: 60vh; max-height: 60vh;"
  >
    <Dialog.Header>
      <Dialog.Title>Import theme</Dialog.Title>
      <Dialog.Description>
        Paste the CSS from tweakcn (the :root and .dark blocks) and give the
        theme a name.
      </Dialog.Description>
    </Dialog.Header>

    <div class="flex min-h-0 flex-1 flex-col gap-4">
      <div class="flex flex-col gap-2">
        <Label for="theme-name">Name</Label>
        <Input id="theme-name" bind:value={name} placeholder="My theme" />
      </div>

      <div class="flex min-h-0 flex-1 flex-col gap-2">
        <Label for="theme-css">Theme CSS</Label>
        <Textarea
          id="theme-css"
          bind:value={css}
          placeholder={":root {\n  --background: oklch(...);\n  ...\n}\n\n.dark {\n  ...\n}"}
          class="font-mono text-xs"
          style="field-sizing: fixed; flex: 1 1 0; min-height: 10rem; resize: none; overflow-y: auto;"
          spellcheck={false}
        />
        {#if error}
          <p class="text-destructive text-sm" role="alert">{error}</p>
        {/if}
      </div>
    </div>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => (open = false)}>Cancel</Button>
      <Button disabled={!canImport} onclick={handleImport}>Import</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>