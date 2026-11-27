<script lang="ts">
  import {
    Box,
    Check,
    ChevronDown,
    ExternalLink,
    Eye,
    EyeOff,
    FolderOpen,
    Pencil,
    X,
  } from "@lucide/svelte";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { Events } from "@wailsio/runtime";
  import {
    InstallState,
    StorageState,
    type ModKeyBind,
    type ModKeybindUpdate,
    type ModSaved,
    type Pack,
  } from "#lib/bindings/changeme/models.ts";
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import { cn } from "#lib/utils.ts";
  import { Badge } from "#lib/components/ui/badge/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { Checkbox } from "#lib/components/ui/checkbox/index.ts";
  import * as DropdownMenu from "#lib/components/ui/dropdown-menu/index.ts";
  import Input from "#lib/components/ui/input/input.svelte";
  import { Label } from "#lib/components/ui/label/index.ts";
  import * as Resizable from "#lib/components/ui/resizable/index.ts";
  import * as Select from "#lib/components/ui/select/index.ts";
  import { Switch } from "#lib/components/ui/switch/index.ts";
  import { formatSize } from "./utils.ts";
  import { openPreview } from "#lib/state/app.svelte.ts";

  let {
    mod,
    folders = [],
    imageSize = $bindable(35),
    onclose,
    onunload,
    onopenfolder,
    onopenlink,
    onrename,
    onrecategorize,
    onTogglePack,
  }: {
    mod: ModSaved;
    /** Folders of the selected tool this mod can be moved to. */
    folders?: string[];
    /** Height of the preview image as a percentage (0-100) of the panel. Bind it to keep the size across mods. */
    imageSize?: number;
    onclose?: () => void;
    onunload?: () => void;
    onopenfolder?: () => void;
    onopenlink?: () => void;
    onrename?: (name: string) => void;
    onrecategorize?: (folder: string) => void;
    onTogglePack: (pack: Pack) => void;
  } = $props();

  const savedCtx = getSavedModsContext();
  const queryClient = useQueryClient();

  const packs = $derived(mod.packages ?? []);

  const isActive = (p: Pack) => p.storage === StorageState.StorageActive;
  const isBusy = (p: Pack) =>
    p.storage === StorageState.StorageCompressing ||
    p.storage === StorageState.StorageRestoring;
  const canPreview = (p: Pack) =>
    p.state === InstallState.InstallReady &&
    p.storage !== StorageState.StorageNone &&
    !isBusy(p);
  const packLabel = (p: Pack) => p.name || p.manifest.relPath;

  const enabled = $derived(packs.some(isActive));
  const activeCount = $derived(packs.filter(isActive).length);
  const previewable = $derived(packs.filter(canPreview));

  let previewMenuOpen = $state(false);

  let renaming = $state(false);
  let nameDraft = $state("");
  let nameInput = $state<HTMLInputElement | null>(null);

  function startRename() {
    nameDraft = mod.submission.name;
    renaming = true;
  }

  function commitRename() {
    if (!renaming) return;
    renaming = false;
    const name = nameDraft.trim();
    if (name && name !== mod.submission.name) onrename?.(name);
  }

  $effect(() => {
    if (!renaming) return;
    nameInput?.focus();
    nameInput?.select();
  });

  let folderQuery = $state("");
  const filteredFolders = $derived(
    folders.filter((f) =>
      f.toLowerCase().includes(folderQuery.trim().toLowerCase()),
    ),
  );

  $effect(() => {
    void mod.id;
    renaming = false;
    folderQuery = "";
  });

  const hasDescription = $derived(
    mod.submission.description.trim().length > 0,
  );
  const descriptionKey = $derived(`desc-hide-${mod.id}`);
  let hideDescription = $derived(
    localStorage.getItem(descriptionKey) === "1",
  );

  function toggleDescription() {
    hideDescription = !hideDescription;
    localStorage.setItem(descriptionKey, hideDescription ? "1" : "0");
  }

  const queryKey = $derived(["current_keybinds", mod.id]);

  const currentQuery = createQuery(() => ({
    queryKey,
    queryFn: () => ModService.GetCurrentBinds(mod.id),
  }));

  const binds = $derived<ModKeyBind[]>(currentQuery.data?.values ?? []);

  const editKeyMutation = createMutation(() => ({
    mutationFn: (update: ModKeybindUpdate) => ModService.UpdateKey(update),
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  }));

  $effect(() => {
    const unsubscribe = Events.On("ini-file-updated", ({ data }) => {
      if (!data) return;
      queryClient.invalidateQueries({ queryKey: ["current_keybinds", data] });
    });
    return () => unsubscribe();
  });

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

  const isCapturing = (bind: ModKeyBind) =>
    capturing?.packId === bind.packId && capturing?.section === bind.section;

  function keyLabel(bind: ModKeyBind) {
    if (isCapturing(bind)) {
      return heldKeys.length ? formatKey(heldKeys.join(" ")) : "Press a key…";
    }
    const pending = editKeyMutation.isPending
      ? editKeyMutation.variables
      : undefined;
    const isPending =
      pending?.packId === bind.packId && pending?.section === bind.section;
    return formatKey(isPending ? pending?.value : bind.key);
  }

  function toggleCapture(bind: ModKeyBind) {
    if (isCapturing(bind)) return cancelCapture();
    heldKeys = [];
    capturing = { packId: bind.packId, section: bind.section };
  }

  function cancelCapture() {
    capturing = undefined;
    heldKeys = [];
  }

  function commitKey(
    target: { packId: string; section: string },
    combo: string[],
  ) {
    cancelCapture();
    editKeyMutation.mutate({
      modId: mod.id,
      packId: target.packId,
      section: target.section,
      key: "key", // ini key name
      value: combo.join(" "),
    });
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

  const sectionTitle =
    "text-muted-foreground text-xs font-medium tracking-widest uppercase";
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
        {:else}
          <Box
            class="text-muted-foreground/40 absolute inset-0 m-auto size-12"
          />
        {/if}

        <div
          class="from-background/80 pointer-events-none absolute inset-x-0 bottom-0 h-24 bg-linear-to-t to-transparent"
        ></div>

        <Button
          variant="secondary"
          size="icon"
          class="absolute top-3 right-3 size-9 rounded-full backdrop-blur"
          aria-label="Close"
          onclick={onclose}
        >
          <X class="size-4" />
        </Button>

        <div class="absolute bottom-3 left-3">
          {#if previewable.length === 1}
            <Button
              variant="secondary"
              size="sm"
              class="gap-2 rounded-full backdrop-blur"
              onclick={(e) => openPreview(e, mod.id, previewable[0].id)}
            >
              <Box class="size-4" />
              3D preview
            </Button>
          {:else if previewable.length > 1}
            <DropdownMenu.Root bind:open={previewMenuOpen}>
              <DropdownMenu.Trigger>
                {#snippet child({ props })}
                  <Button
                    {...props}
                    variant="secondary"
                    size="sm"
                    class="gap-2 rounded-full backdrop-blur"
                  >
                    <Box class="size-4" />
                    3D preview
                    <ChevronDown class="size-3.5" />
                  </Button>
                {/snippet}
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="start">
                <DropdownMenu.Label>Preview pack</DropdownMenu.Label>
                {#each previewable as pack (pack.id)}
                  <DropdownMenu.Item
                    onSelect={(e) => {
                      openPreview(e, mod.id, pack.id);
                      previewMenuOpen = false;
                    }}
                  >
                    {packLabel(pack)}
                  </DropdownMenu.Item>
                {/each}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          {/if}
        </div>
      </div>
    </Resizable.Pane>

    <Resizable.Handle withHandle aria-label="Resize preview image" />

    <Resizable.Pane minSize={25}>
      <div class="h-full overflow-y-auto">
        <div class="space-y-6 px-5 py-5">
          <header class="space-y-2">
            {#if mod.submission.category}
              <Badge variant="secondary">{mod.submission.category}</Badge>
            {/if}

            {#if renaming}
              <div class="flex items-center gap-2">
                <Input
                  bind:ref={nameInput}
                  bind:value={nameDraft}
                  aria-label="Mod name"
                  class="h-10 text-lg font-semibold"
                  onblur={commitRename}
                  onkeydown={(e) => {
                    if (e.key === "Enter") commitRename();
                    else if (e.key === "Escape") renaming = false;
                  }}
                />
                <Button
                  variant="secondary"
                  size="icon"
                  class="size-9 shrink-0"
                  aria-label="Save name"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={commitRename}
                >
                  <Check class="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-9 shrink-0"
                  aria-label="Cancel rename"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={() => (renaming = false)}
                >
                  <X class="size-4" />
                </Button>
              </div>
            {:else}
              <div class="flex items-start gap-2">
                <h2
                  class="min-w-0 flex-1 text-2xl leading-tight font-bold wrap-break-word"
                >
                  {mod.submission.name}
                </h2>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-8 shrink-0"
                  aria-label="Rename"
                  onclick={startRename}
                >
                  <Pencil class="size-4" />
                </Button>
              </div>
            {/if}

            <div class="text-muted-foreground flex items-center gap-1 text-sm">
              <span>by {mod.submission.author || "Unknown"}</span>
              {#if onopenlink}
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-6"
                  aria-label="Open post"
                  onclick={onopenlink}
                >
                  <ExternalLink class="size-3.5" />
                </Button>
              {/if}
            </div>
          </header>

          <div
            class="bg-muted/40 flex items-center justify-between rounded-xl border px-4 py-3"
          >
            <div>
              <Label
                for="mod-enabled"
                class={cn("text-base font-semibold", enabled && "text-primary")}
              >
                {enabled ? "Enabled" : "Disabled"}
              </Label>
              {#if packs.length > 1}
                <p class="text-muted-foreground text-xs">
                  {activeCount} of {packs.length} packs active
                </p>
              {/if}
            </div>
            <Switch
              id="mod-enabled"
              checked={enabled}
              onCheckedChange={() => savedCtx.toggleAllPkgs(mod.id)}
            />
          </div>

          {#if packs.length > 1}
            <section class="space-y-2">
              <h3 class={sectionTitle}>Packs</h3>
              <div class="divide-y overflow-hidden rounded-xl border">
                {#each packs as pack (pack.id)}
                  {@const ready = pack.state === InstallState.InstallReady}
                  <label
                    class={cn(
                      "flex items-center gap-3 px-4 py-3 transition-colors",
                      isActive(pack) && "bg-muted/50",
                      ready
                        ? "hover:bg-muted/30 cursor-pointer"
                        : "cursor-not-allowed opacity-60",
                    )}
                  >
                    <Checkbox
                      checked={isActive(pack)}
                      disabled={!ready}
                      onCheckedChange={() => onTogglePack(pack)}
                    />
                    <div class="min-w-0 flex-1">
                      <div class="text-sm font-medium break-all">
                        {packLabel(pack)}
                      </div>
                      <div class="text-muted-foreground text-xs">
                        {formatSize(pack.manifest.size)}
                      </div>
                    </div>
                    {#if pack.storage === StorageState.StorageCold}
                      <Badge variant="outline">Compressed</Badge>
                    {/if}
                  </label>
                {/each}
              </div>
            </section>
          {/if}

          {#if folders.length > 0}
            <section class="space-y-2">
              <h3 class={sectionTitle}>Category</h3>
              <Select.Root
                type="single"
                value={mod.characterId ?? ""}
                onValueChange={(folder) => {
                  if (folder && folder !== mod.characterId) {
                    onrecategorize?.(folder);
                  }
                }}
                onOpenChange={(open) => {
                  if (!open) folderQuery = "";
                }}
              >
                <Select.Trigger class="w-full">
                  {mod.characterId || "Uncategorized"}
                </Select.Trigger>
                <Select.Content>
                  <div class="bg-popover sticky top-0 z-10 p-1">
                    <Input
                      bind:value={folderQuery}
                      placeholder="Search folders..."
                      class="h-8"
                      onkeydown={(e) => e.stopPropagation()}
                    />
                  </div>
                  {#each filteredFolders as folder (folder)}
                    <Select.Item value={folder} label={folder} />
                  {:else}
                    <div
                      class="text-muted-foreground px-2 py-4 text-center text-sm"
                    >
                      No folders found
                    </div>
                  {/each}
                </Select.Content>
              </Select.Root>
            </section>
          {/if}

          {#if hasDescription}
            <section class="space-y-2">
              <div class="flex items-center justify-between">
                <h3 class={sectionTitle}>Description</h3>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-7"
                  aria-label={hideDescription
                    ? "Show description"
                    : "Hide description"}
                  onclick={toggleDescription}
                >
                  {#if hideDescription}
                    <EyeOff class="size-4" />
                  {:else}
                    <Eye class="size-4" />
                  {/if}
                </Button>
              </div>
              {#if !hideDescription}
                <div
                  class="text-muted-foreground [&_a]:text-primary text-sm leading-relaxed [&_a]:underline [&_img]:max-w-full [&_img]:rounded-md"
                >
                  {@html mod.submission.description}
                </div>
              {/if}
            </section>
          {/if}

          <section class="space-y-2">
            <div class="flex items-baseline justify-between gap-2">
              <h3 class={sectionTitle}>In-game keys</h3>
              <span class="text-muted-foreground/70 text-xs">
                click a key to change it
              </span>
            </div>

            {#if currentQuery.data}
              <p class="text-primary text-sm font-semibold">
                {currentQuery.data.name || "Key sheet"}
              </p>

              <ul class="divide-y overflow-hidden rounded-xl border">
                {#each binds as bind (bind.packId + ":" + bind.section)}
                  {@const active = isCapturing(bind)}
                  <li class="flex items-center justify-between gap-3 px-4 py-2.5">
                    <div class="min-w-0 truncate text-sm">
                      <span class="font-semibold">
                        {sectionLabel(bind.section)}
                      </span>
                      <span class="text-muted-foreground">
                        {bind.actionType}
                      </span>
                    </div>

                    <button
                      type="button"
                      data-key-capture
                      class={cn(
                        "min-w-14 shrink-0 rounded-md border px-3 py-1 font-mono text-xs font-bold transition-colors",
                        active
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-muted/60 hover:bg-muted",
                      )}
                      onclick={() => toggleCapture(bind)}
                    >
                      {keyLabel(bind)}
                    </button>
                  </li>
                {:else}
                  <li class="text-muted-foreground px-4 py-3 text-sm">
                    No keybinds in this mod.
                  </li>
                {/each}
              </ul>
            {/if}
          </section>
        </div>
      </div>
    </Resizable.Pane>
  </Resizable.PaneGroup>

  <footer class="bg-card flex gap-3 border-t px-5 py-3">
    <Button
      variant="secondary"
      class="gap-2 rounded-full px-5 font-semibold"
      onclick={onopenfolder}
    >
      <FolderOpen class="size-4" />
      Open folder
    </Button>
    <Button
      variant="outline"
      class="flex-1 rounded-full font-semibold"
      onclick={onunload}
    >
      Unload
    </Button>
  </footer>
</div>