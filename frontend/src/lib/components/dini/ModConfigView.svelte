<script lang="ts">
  import {
    Box,
    Pencil,
    Star,
    X,
    ArrowUp,
    StickyNote,
    Eye,
  } from "@lucide/svelte";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import { Button } from "../ui/button/index.js";
  import { Label } from "../ui/label/index.js";
  import { Switch } from "../ui/switch/index.js";
  import Input from "../ui/input/input.svelte";
  import { Textarea } from "../ui/textarea/index.js";
  import {
    InstallState,
    type ModKeyBindPreset,
    type ModSaved,
    type Pack,
  } from "#lib/bindings/changeme/models.js";
  import { Checkbox } from "../ui/checkbox/index.js";
  import { formatSize } from "./utils.js";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.js";
  import Keybinds from "./Keybinds.svelte";

  let {
    mod,
    /** Height of the preview image as a percentage (0-100) of the panel. Bind it to keep the size across mods. */
    imageSize = $bindable(35),
    onclose,
    onunload,
    onopenfolder,
    onopenlink,
    onTogglePack,
    keybindData,
  }: {
    mod: ModSaved;
    onTogglePack: (pack: Pack) => void;
    keybindData?: ModKeyBindPreset[];
    folders?: string[];
    imageSize?: number;
    onclose?: () => void;
    onunload?: () => void;
    onopenfolder?: () => void;
    onopenlink?: () => void;
  } = $props();

  let noteOpen = $state(false);
  let hideDescription = $derived(
    localStorage.getItem("desc-hide-" + mod.id) === "1",
  );

  $effect(() => {
    localStorage.setItem("desc-hide-" + mod.id, hideDescription ? "1" : "0");
  });
  const hasNote = $derived(mod.submission.description.trim().length > 0);
  const anyInstalled = $derived(
    mod.packages?.some((p) => p.state === InstallState.InstallReady),
  );
  const enabled = $derived(mod.packages?.some((p) => p.enabled));
  const savedCtx = getSavedModsContext();

  function handleToggleMod() {
    savedCtx.toggleAllPkgs(mod.id);
  }
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
        {#if mod.submission.thumbnailUrl}
          <img
            src={mod.submission.thumbnailUrl}
            alt=""
            class="size-full object-cover"
          />
        {/if}

        {#if anyInstalled}
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
              {mod.submission.category}
            </p>
            <div class="mt-1 flex items-start gap-3">
              <h2
                class="min-w-0 flex-1 text-2xl leading-tight font-bold wrap-break-word"
              >
                {mod.submission.name}
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
                  aria-label={false
                    ? "Remove from favorites"
                    : "Add to favorites"}
                  aria-pressed={true}
                >
                  <Star class="size-4 {false ? 'fill-current' : ''}" />
                </Button>
              </div>
            </div>
            <p class="text-muted-foreground mt-2 text-sm">
              by {mod.submission.author || "Unknown"}
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
              {enabled ? "Enabled" : "Disabled"}
            </Label>
            <Switch
              id="mod-enabled"
              checked={enabled}
              onCheckedChange={handleToggleMod}
            />
          </div>
          {#if (mod.packages?.length ?? 0) > 1}
            {#each mod.packages as pack (pack.id)}
              <label
                class="flex items-start gap-3 px-4 py-3 transition-colors {pack.enabled
                  ? 'bg-muted/60'
                  : 'hover:bg-muted/30'} {pack.state !==
                InstallState.InstallReady
                  ? 'cursor-not-allowed opacity-60'
                  : 'cursor-pointer'}"
              >
                <Checkbox
                  class="mt-1"
                  checked={pack.enabled}
                  onCheckedChange={() => onTogglePack(pack)}
                />

                <div class="min-w-0 flex-1 space-y-0.5">
                  <div class="break-all font-semibold">
                    {pack.manifest.relPath}
                  </div>
                  <div class="flex gap-3 text-sm text-muted-foreground">
                    <span class="font-medium"
                      >{formatSize(pack.manifest.size)}</span
                    >
                  </div>
                </div>
              </label>
            {/each}
          {/if}
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
                value={mod.submission.author}
                placeholder="Unknown"
              />

              <Label for="mod-link" class="text-muted-foreground">Link</Label>
              <div class="flex gap-2">
                <Input
                  id="mod-link"
                  type="url"
                  value={mod.submission.slug}
                  placeholder="https://"
                />
                <Button
                  variant="secondary"
                  size="icon"
                  class="shrink-0"
                  aria-label="Open link"
                  disabled={!mod.submission.slug}
                  onclick={onopenlink}
                >
                  <ArrowUp class="size-4" />
                </Button>
              </div>
            </div>

            <Button
              onclick={() => {
                hideDescription = !hideDescription;
              }}
            >
              <Eye />
            </Button>
            {#if !hideDescription && (noteOpen || hasNote)}
              {@html mod.submission.description}
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
        </div>
        {#if keybindData}
          <Keybinds
            data={[
              {
                mod,
                name: "",
                presets: keybindData,
              },
            ]}
            toolId={mod.toolId}
            title=""
          />
        {/if}
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
