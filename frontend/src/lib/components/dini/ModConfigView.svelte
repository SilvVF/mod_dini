<script lang="ts">
  import { Box, Pencil, Star, X, ArrowUp, StickyNote } from "@lucide/svelte";
  import type { Mod, ModPatch } from "./types.js";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import { Button } from "../ui/button/index.js";
  import { Label } from "../ui/label/index.js";
  import { Switch } from "../ui/switch/index.js";
  import Input from "../ui/input/input.svelte";
  import { Textarea } from "../ui/textarea/index.js";

  let {
    mod,
    folders = [],
    /** Height of the preview image as a percentage (0-100) of the panel. Bind it to keep the size across mods. */
    imageSize = $bindable(35),
    onclose,
    onunload,
    onchange,
    onopenfolder,
    onopenlink,
  }: {
    mod: Mod;
    folders?: string[];
    imageSize?: number;
    onclose?: () => void;
    onunload?: () => void;
    onchange?: (patch: ModPatch) => void;
    onopenfolder?: () => void;
    onopenlink?: () => void;
  } = $props();

  let noteOpen = $state(false);
  const hasNote = $derived(mod.note.trim().length > 0);
  const folderOptions = $derived(
    folders.includes(mod.folder) ? folders : [mod.folder, ...folders],
  );
</script>

<div class="flex min-h-0 flex-1 flex-col">
  <Resizable.PaneGroup direction="vertical" class="min-h-0 flex-1">
    <!-- Preview image: drag the handle below it to resize -->
    <Resizable.Pane
      defaultSize={imageSize}
      minSize={15}
      maxSize={75}
      onResize={(size) => (imageSize = size)}
    >
      <div class="bg-muted relative size-full overflow-hidden">
        {#if mod.thumbnail}
          <img src={mod.thumbnail} alt="" class="size-full object-cover" />
        {/if}

        {#if mod.has3D}
          <span
            class="bg-background/80 absolute bottom-2 left-3 flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-semibold backdrop-blur"
          >
            <Box class="size-3.5" /> 3D
          </span>
        {/if}

        <div class="absolute top-3 right-3 flex gap-2">
          <Button
            variant="secondary"
            size="icon"
            class="size-9 rounded-full"
            aria-label="Edit previews"
          >
            <Pencil class="size-4" />
          </Button>
          <Button
            variant="secondary"
            size="icon"
            class="size-9 rounded-full"
            aria-label="Close"
            onclick={onclose}
          >
            <X class="size-4" />
          </Button>
        </div>
      </div>
    </Resizable.Pane>

    <Resizable.Handle withHandle aria-label="Resize preview image" />

    <Resizable.Pane minSize={25}>
      <div class="h-full overflow-y-auto">
        <div class="space-y-5 px-5 pt-4 pb-6">
          <!-- Title -->
          <div>
            <p class="text-primary text-xs font-medium tracking-wide">
              {mod.category}
            </p>
            <div class="mt-1 flex items-start gap-3">
              <h2
                class="min-w-0 flex-1 font-serif text-2xl leading-tight font-bold break-words"
              >
                {mod.name}
              </h2>
              <div class="flex shrink-0 gap-2">
                <Button
                  variant="secondary"
                  size="icon"
                  class="size-9"
                  aria-label="Rename"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  variant="secondary"
                  size="icon"
                  class="size-9"
                  aria-label={mod.favorite
                    ? "Remove from favorites"
                    : "Add to favorites"}
                  aria-pressed={mod.favorite}
                  onclick={() => onchange?.({ favorite: !mod.favorite })}
                >
                  <Star class="size-4 {mod.favorite ? 'fill-current' : ''}" />
                </Button>
              </div>
            </div>
            <p class="text-muted-foreground mt-2 text-sm">
              by {mod.author || "Unknown"}
            </p>
          </div>

          <!-- Enabled -->
          <div
            class="bg-muted/60 flex items-center justify-between rounded-xl px-4 py-3.5"
          >
            <Label
              for="mod-enabled"
              class="text-primary text-base font-semibold"
            >
              {mod.enabled ? "Enabled" : "Disabled"}
            </Label>
            <Switch
              id="mod-enabled"
              checked={mod.enabled}
              onCheckedChange={(v) => onchange?.({ enabled: v })}
            />
          </div>

          <!-- Info -->
          <section class="space-y-3">
            <h3 class="text-muted-foreground text-xs">
              Info <span class="ml-1 opacity-70">saved automatically</span>
            </h3>

            <div class="grid grid-cols-[4rem_1fr] items-center gap-x-3 gap-y-3">
              <Label for="mod-author" class="text-muted-foreground"
                >Author</Label
              >
              <Input
                id="mod-author"
                value={mod.author}
                placeholder="Unknown"
                oninput={(e) => onchange?.({ author: e.currentTarget.value })}
              />

              <Label for="mod-link" class="text-muted-foreground">Link</Label>
              <div class="flex gap-2">
                <Input
                  id="mod-link"
                  type="url"
                  value={mod.link}
                  placeholder="https://"
                  oninput={(e) => onchange?.({ link: e.currentTarget.value })}
                />
                <Button
                  variant="secondary"
                  size="icon"
                  class="shrink-0"
                  aria-label="Open link"
                  disabled={!mod.link}
                  onclick={onopenlink}
                >
                  <ArrowUp class="size-4" />
                </Button>
              </div>
            </div>

            {#if noteOpen || hasNote}
              <Textarea
                value={mod.note}
                rows={3}
                placeholder="Write a note about this mod"
                aria-label="Note"
                oninput={(e) => onchange?.({ note: e.currentTarget.value })}
              />
            {:else}
              <button
                type="button"
                onclick={() => (noteOpen = true)}
                class="text-muted-foreground hover:text-foreground flex items-center gap-2 text-sm"
              >
                <StickyNote class="size-4" /> Add a note
              </button>
            {/if}
          </section>

          <!-- Folder -->
          <section class="space-y-2">
            <Label for="mod-folder" class="text-muted-foreground text-xs"
              >Folder</Label
            >
            <select
              id="mod-folder"
              value={mod.folder}
              onchange={(e) => onchange?.({ folder: e.currentTarget.value })}
              class="border-input bg-background focus-visible:ring-ring h-10 w-full rounded-md border px-3 text-sm outline-none focus-visible:ring-2"
            >
              {#each folderOptions as f (f)}
                <option value={f}>{f}</option>
              {/each}
            </select>
          </section>
        </div>
      </div>
    </Resizable.Pane>
  </Resizable.PaneGroup>

  <!-- Sticky actions -->
  <footer class="bg-card flex gap-3 border-t px-5 py-3">
    <Button
      variant="secondary"
      class="rounded-full px-5 font-semibold"
      onclick={onopenfolder}>Open folder</Button
    >
    <Button
      variant="outline"
      class="flex-1 rounded-full font-semibold"
      onclick={onunload}>Unload</Button
    >
  </footer>
</div>
