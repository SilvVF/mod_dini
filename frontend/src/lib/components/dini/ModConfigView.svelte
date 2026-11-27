<script lang="ts">
  import {
    Box,
    Pencil,
    Star,
    X,
    ArrowUp,
    StickyNote,
    Eye,
    Play,
  } from "@lucide/svelte";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import { Button } from "../ui/button/index.js";
  import { Label } from "../ui/label/index.js";
  import { Switch } from "../ui/switch/index.js";
  import Input from "../ui/input/input.svelte";
  import {
    InstallState,
    StorageState,
    type ModKeyBind,
    type ModKeyBindPreset,
    type ModKeybindUpdate,
    type ModSaved,
    type Pack,
  } from "#lib/bindings/changeme/models.js";
  import { Checkbox } from "../ui/checkbox/index.js";
  import { formatSize } from "./utils.js";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.js";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { ModService } from "#lib/bindings/changeme/index.js";
  import { Events } from "@wailsio/runtime";

  let {
    mod,
    /** Height of the preview image as a percentage (0-100) of the panel. Bind it to keep the size across mods. */
    imageSize = $bindable(35),
    onclose,
    onunload,
    onopenfolder,
    onopenlink,
    onTogglePack,
  }: {
    mod: ModSaved;
    onTogglePack: (pack: Pack) => void;
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

  const isActive = (p: Pack) => p.storage === StorageState.StorageActive;
  const enabled = $derived(mod.packages?.some(isActive));
  const savedCtx = getSavedModsContext();

  function handleToggleMod() {
    savedCtx.toggleAllPkgs(mod.id);
  }

  const queryKey = $derived(["current_keybinds", mod.id]);

  const currentQuery = createQuery(() => ({
    queryKey,
    queryFn: () => ModService.GetCurrentBinds(mod.id),
  }));

  const binds = $derived<ModKeyBind[]>(currentQuery.data?.values ?? []);

  const queryClient = useQueryClient();

  const editKeyMutation = createMutation(() => ({
    mutationFn: (update: ModKeybindUpdate) => ModService.UpdateKey(update),
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  }));

  $effect(() => {
    const unsubscribe = Events.On("ini-file-updated", ({data}) => {
      if (!data) return
      queryClient.invalidateQueries({ queryKey: ["current_keybinds", data] })
    })

    return () => unsubscribe();
  })

  let capturing = $state<{ packId: string; section: string } | undefined>();
  let heldKeys = $state<string[]>([]);

  const MOD_ORDER = ["ctrl", "alt", "shift"];
  const sortCombo = (keys: string[]) =>
    [...keys].sort((a, b) => {
      const ai = MOD_ORDER.indexOf(a);
      const bi = MOD_ORDER.indexOf(b);
      return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi);
    });

  const SPECIAL_KEYS: Record<string, string> = {
    ArrowLeft: "VK_LEFT",
    ArrowRight: "VK_RIGHT",
    ArrowUp: "VK_UP",
    ArrowDown: "VK_DOWN",
    " ": "VK_SPACE",
    Enter: "VK_RETURN",
    Tab: "VK_TAB",
    Backspace: "VK_BACK",
    Delete: "VK_DELETE",
    Insert: "VK_INSERT",
    Home: "VK_HOME",
    End: "VK_END",
    PageUp: "VK_PRIOR",
    PageDown: "VK_NEXT",
  };

  function toIniKey(e: KeyboardEvent): string | undefined {
    switch (e.key) {
      case "Control":
        return "ctrl";
      case "Alt":
        return "alt";
      case "Shift":
        return "shift";
      case "Meta":
        return undefined;
    }
    if (SPECIAL_KEYS[e.key]) return SPECIAL_KEYS[e.key];
    if (/^F\d{1,2}$/.test(e.key)) return e.key;
    if (e.code.startsWith("Key")) return e.code.slice(3).toLowerCase();
    if (e.code.startsWith("Digit")) return e.code.slice(5);
    if (e.code.startsWith("Numpad")) return "VK_NUMPAD" + e.code.slice(6);
    return e.key.length === 1 ? e.key.toLowerCase() : undefined;
  }

  function formatKey(raw: string | undefined) {
    if (!raw) return "—";
    return raw
      .split(/\s+/)
      .map((k) => k.replace(/^VK_/i, ""))
      .map((k) =>
        MOD_ORDER.includes(k.toLowerCase())
          ? k[0].toUpperCase() + k.slice(1).toLowerCase()
          : k.toUpperCase(),
      )
      .join(" + ");
  }

  const sectionLabel = (section: string) => section.replace(/^Key/i, "");

  function commitKey(
    target: { packId: string; section: string },
    combo: string[],
  ) {
    capturing = undefined;
    heldKeys = [];
    editKeyMutation.mutate({
      modId: mod.id,
      packId: target.packId,
      section: target.section,
      key: "key", // ini key name
      value: combo.join(" "),
    });
  }

  function cancelCapture() {
    capturing = undefined;
    heldKeys = [];
  }

  // wait for a key (or a held combo, committed on the first release)
  $effect(() => {
    const target = capturing;
    if (!target) return;
    let combo: string[] = [];

    const onKeyDown = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (e.repeat) return;
      if (e.key === "Escape") return cancelCapture();
      const name = toIniKey(e);
      if (name && !combo.includes(name)) {
        combo = sortCombo([...combo, name]);
        heldKeys = combo;
      }
    };
    const onKeyUp = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (combo.length) commitKey(target, combo);
    };
    const onPointerDown = (e: PointerEvent) => {
      if (!(e.target as HTMLElement).closest("[data-key-capture]")) {
        cancelCapture();
      }
    };

    window.addEventListener("keydown", onKeyDown, true);
    window.addEventListener("keyup", onKeyUp, true);
    window.addEventListener("pointerdown", onPointerDown, true);
    return () => {
      window.removeEventListener("keydown", onKeyDown, true);
      window.removeEventListener("keyup", onKeyUp, true);
      window.removeEventListener("pointerdown", onPointerDown, true);
    };
  });

  // no backend method exists for "press it in the game" yet
  function handlePressInGame(bind: ModKeyBind) {
    // e.g. ModService.PressKey(mod.id, bind.key)
  }
</script>

<div class="flex min-h-0 flex-1 flex-col">
  <Resizable.PaneGroup direction="vertical" class="min-h-0 flex-1">
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
                class="flex items-start gap-3 px-4 py-3 transition-colors {isActive(
                  pack,
                )
                  ? 'bg-muted/60'
                  : 'hover:bg-muted/30'} {pack.state !==
                InstallState.InstallReady
                  ? 'cursor-not-allowed opacity-60'
                  : 'cursor-pointer'}"
              >
                <Checkbox
                  class="mt-1"
                  checked={isActive(pack)}
                  onCheckedChange={() => onTogglePack(pack)}
                />

                <div class="min-w-0 flex-1 space-y-0.5">
                  <div class="break-all font-semibold">
                    {pack.manifest.relPath}
                  </div>
                  <div class="text-muted-foreground flex gap-3 text-sm">
                    <span class="font-medium"
                      >{formatSize(pack.manifest.size)}</span
                    >
                  </div>
                </div>
              </label>
            {/each}
          {/if}
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

        <section class="px-5 pb-6">
          <div class="flex flex-wrap items-baseline gap-x-2">
            <h3
              class="text-muted-foreground text-xs font-medium tracking-widest uppercase"
            >
              In-game keys
            </h3>
            <span class="text-muted-foreground/70 text-xs">
              click a key to change it, the arrow to press it in the game
            </span>
          </div>

          {#if currentQuery.data}
            <p class="text-primary mt-1 font-semibold">
              {currentQuery.data.name || "Key sheet"}
            </p>

            <ul class="divide-border mt-2 divide-y divide-dashed">
              {#each binds as bind (bind.packId + ":" + bind.section)}
                {@const isCapturing =
                  capturing?.packId === bind.packId &&
                  capturing?.section === bind.section}
                <li class="flex items-center justify-between gap-3 py-3">
                  <div class="min-w-0 truncate">
                    <span class="font-semibold"
                      >{sectionLabel(bind.section)}</span
                    >
                    <span class="text-muted-foreground">{bind.actionType}</span>
                  </div>

                  <div class="flex shrink-0 items-center gap-2">
                    <button
                      type="button"
                      data-key-capture
                      class="min-w-14 rounded-lg border px-3 py-1.5 font-mono text-sm font-bold transition-colors
                        {isCapturing
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'bg-muted/60 hover:bg-muted'}"
                      onclick={() => {
                        if (isCapturing) return cancelCapture();
                        heldKeys = [];
                        capturing = {
                          packId: bind.packId,
                          section: bind.section,
                        };
                      }}
                    >
                      {#if isCapturing}
                        {heldKeys.length
                          ? formatKey(heldKeys.join(" "))
                          : "Press a key…"}
                      {:else}
                        {formatKey(
                          editKeyMutation.isPending &&
                            editKeyMutation.variables.packId === bind.packId &&
                            editKeyMutation.variables.section === bind.section
                            ? editKeyMutation.variables.value
                            : bind.key,
                        )}
                      {/if}
                    </button>

                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-muted-foreground size-8"
                      aria-label="Press {sectionLabel(
                        bind.section,
                      )} in the game"
                      onclick={() => handlePressInGame(bind)}
                    >
                      <Play class="size-4 fill-current" />
                    </Button>
                  </div>
                </li>
              {:else}
                <li class="text-muted-foreground py-3 text-sm">
                  No keybinds in this mod.
                </li>
              {/each}
            </ul>
          {/if}
        </section>
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
