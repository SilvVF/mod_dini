<script lang="ts">
  import type { PageProps } from "./$types";
  import { tick } from "svelte";
  import { SvelteSet } from "svelte/reactivity";
  import { Collapsible } from "bits-ui";
  import { Browser } from "@wailsio/runtime";
  import {
    AlertTriangle,
    Check,
    ChevronDown,
    Layers,
    Plus,
    Power,
    PowerOff,
    Search,
    Trash2,
    X,
  } from "@lucide/svelte";
 
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import * as Dialog from "#lib/components/ui/dialog/index.ts";
  import { Input } from "#lib/components/ui/input/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import Switch from "#lib/components/ui/switch/switch.svelte";
 
  import ModCard from "#lib/components/dini/ModCard.svelte";
  import ModLibraryHeader from "#lib/components/dini/ModLibraryHeader.svelte";
  import LibraryConfig from "#lib/components/dini/LibraryConfig.svelte";
  import type { Conflict, GameConfig } from "#lib/components/dini/types.ts";
  import { formatSize, modLink } from "#lib/components/dini/utils.ts";
 
  import {
    StorageState,
    type CharacterFolder,
    type ModSaved,
    type Pack,
    type Subfolder,
  } from "#lib/bindings/changeme/models.ts";
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import {
    groupModsByFolders,
    type SelectedItem,
  } from "#lib/contexts/group-mods.ts";
  import { appState, toolBanner, toolName } from "#lib/state/app.svelte.ts";
  import { compact, sumBy } from "#lib/ts-ext.ts";
 
  type ModFilter = "all" | "enabled" | "disabled";
  type SidebarTab = "all" | "with-mods";
  type Stats = { enabled: number; total: number };
 
  let { data }: PageProps = $props();
  const saved = getSavedModsContext();
 
  // Columns follow the pane width (it is resizable), not the window width.
  const gridClass = `grid w-full auto-rows-[22rem] items-stretch gap-4
    grid-cols-[repeat(auto-fill,minmax(18rem,1fr))]`;
 
  const modFilterTabs: { id: ModFilter; label: string }[] = [
    { id: "all", label: "All" },
    { id: "enabled", label: "Enabled" },
    { id: "disabled", label: "Disabled" },
  ];
 
  let groupMods = $state(localStorage.getItem("groupMods") === "1");
  let hideConflicts = $state(localStorage.getItem("hideConflicts") === "1");
 
  $effect(() => {
    localStorage.setItem("groupMods", groupMods ? "1" : "0");
  });
  $effect(() => {
    localStorage.setItem("hideConflicts", hideConflicts ? "1" : "0");
  });
 
  const toolId = $derived(appState.activeTool);
  const rootFolder = $derived(data.rootFolder[toolId]);
  const subfoldersRecord = $derived(compact(rootFolder?.subfolders ?? {}));
  const charactersRecord = $derived(
    compact(rootFolder?.characters?.characters ?? {}),
  );
  const subfolders = $derived(Object.entries(subfoldersRecord));
  const characters = $derived(Object.entries(charactersRecord));
 
  const mods = $derived(
    Object.values(saved.mods).filter((m) => m.toolId === toolId),
  );
  const groupedMods = $derived(
    groupModsByFolders(mods, rootFolder, charactersRecord, subfoldersRecord),
  );
 
  const isActivePack = (p: Pack) => p.storage === StorageState.StorageActive;
  const activePackCount = (m: ModSaved) =>
    (m.packages ?? []).filter(isActivePack).length;
  const isEnabled = (m: ModSaved) => activePackCount(m) > 0;
 
  function countEnabled(items: ModSaved[]): Stats {
    return { enabled: items.filter(isEnabled).length, total: items.length };
  }
 
  let selectedItemId = $state<string | undefined>(undefined);
  let selectedModId = $state<string | undefined>(undefined);
 
  const selected = $derived.by<SelectedItem>(() => {
    if (!selectedItemId) return { type: "nothing", value: null };
 
    const subfolder = subfoldersRecord[selectedItemId];
    if (subfolder) return { type: "subfolder", value: subfolder };
 
    const character = charactersRecord[selectedItemId];
    if (character) return { type: "character", value: character };
 
    return { type: "nothing", value: null };
  });
 
  const selectedMod = $derived(
    selectedModId ? saved.mods[selectedModId] : undefined,
  );
 
  // Changing the sidebar selection closes any open mod.
  $effect(() => {
    selectedItemId;
    selectedModId = undefined;
  });
 
  // Several groups can share a folder, so their counts are summed.
  const folderStats = $derived.by(() => {
    const map = new Map<string, Stats>();
    for (const group of groupedMods) {
      const { enabled, total } = countEnabled(group.items);
      const prev = map.get(group.folder.name) ?? { enabled: 0, total: 0 };
      map.set(group.folder.name, {
        enabled: prev.enabled + enabled,
        total: prev.total + total,
      });
    }
    return map;
  });
  const statsFor = (folderName: string): Stats =>
    folderStats.get(folderName) ?? { enabled: 0, total: 0 };
  const allStats = $derived(countEnabled(mods));
 
  let sidebarTab = $state<SidebarTab>("all");
  let sidebarSearch = $state("");
  const sidebarQuery = $derived(sidebarSearch.trim().toLowerCase());
  const matchesSidebar = (name: string) =>
    name.toLowerCase().includes(sidebarQuery);
 
  const filteredSubfolders = $derived(
    subfolders.filter(([name]) => matchesSidebar(name)),
  );
  const filteredCharacters = $derived(
    characters.filter(
      ([name, folder]) =>
        matchesSidebar(name) &&
        (sidebarTab !== "with-mods" || statsFor(folder.name).total > 0),
    ),
  );
 
  let modFilter = $state<ModFilter>("all");
  let modSearch = $state("");
  let modSearchEl = $state<HTMLInputElement | null>(null);
  const modQuery = $derived(modSearch.trim().toLowerCase());
 
  const selectedGroups = $derived(
    selected.type === "nothing"
      ? groupedMods
      : groupedMods.filter((g) => g.folder.name === selected.value.name),
  );
 
  type Group = (typeof groupedMods)[number];
 
  function filterGroups(
    groups: Group[],
    predicate: (mod: ModSaved, group: Group) => boolean,
  ) {
    return groups
      .map((group) => ({
        ...group,
        items: group.items.filter((mod) => predicate(mod, group)),
      }))
      .filter((group) => group.items.length > 0);
  }
 
  function matchesSearch(mod: ModSaved, group: Group) {
    if (!modQuery) return true;
    return [mod.submission?.name, mod.submission?.author, group.name].some(
      (field) => field?.toLowerCase().includes(modQuery),
    );
  }
 
  function matchesFilter(mod: ModSaved) {
    if (modFilter === "all") return true;
    return isEnabled(mod) === (modFilter === "enabled");
  }
 
  // Search is applied first so the tab counts reflect it.
  const searchedGroups = $derived(filterGroups(selectedGroups, matchesSearch));
  const visibleGroups = $derived(
    filterGroups(searchedGroups, (mod) => matchesFilter(mod)),
  );
  const visibleFlat = $derived(
    visibleGroups.flatMap((group) =>
      group.items.map((mod) => ({ group, mod })),
    ),
  );
  const visibleCount = $derived(visibleFlat.length);
 
  const filterCounts = $derived.by<Record<ModFilter, number>>(() => {
    const searched = searchedGroups.flatMap((g) => g.items);
    const enabledCount = searched.filter(isEnabled).length;
    return {
      all: searched.length,
      enabled: enabledCount,
      disabled: searched.length - enabledCount,
    };
  });
 
  const loadedInSelection = $derived(
    sumBy(selectedGroups, (g) => g.items.filter(isEnabled).length),
  );
 
  // enabled/total for a group header, ignoring tab/search filters
  const groupStats = (groupName: string): Stats =>
    countEnabled(selectedGroups.find((g) => g.name === groupName)?.items ?? []);
 
  let conflictsOpen = $state(false);
  const showConflictUi = $derived(!hideConflicts);
 
  const conflicts = $derived.by<Conflict[]>(() =>
    groupedMods.flatMap((group) => {
      if (!(group.folder.name in charactersRecord)) return [];
      const on = group.items.filter(isEnabled);
      const packCount = sumBy(on, activePackCount);
      if (packCount <= 1) return [];
      return [
        {
          name: group.name,
          folderName: group.folder.name,
          mods: on,
          packCount,
        },
      ];
    }),
  );
 
  // every mod with an enabled pack on a conflicting character
  const conflictModIds = $derived(
    new Set(conflicts.flatMap((c) => c.mods.map((m) => m.id))),
  );
 
  async function openConflictMod(c: Conflict, mod: ModSaved) {
    conflictsOpen = false;
    selectedItemId = c.folderName;
    await tick();
    selectedModId = mod.id;
  }
 
  let selectionMode = $state(false);
  const selectedModIds = new SvelteSet<string>();
 
  const selectedMods = $derived(
    [...selectedModIds]
      .map((id) => saved.mods[id])
      .filter((mod): mod is ModSaved => !!mod && mod.toolId === toolId),
  );
  const selectedModCount = $derived(selectedMods.length);
  const allVisibleSelected = $derived(
    visibleFlat.length > 0 &&
      visibleFlat.every(({ mod }) => selectedModIds.has(mod.id)),
  );
 
  // Switching tools resets selection.
  $effect(() => {
    toolId;
    clearModSelection();
  });
 
  function toggleModSelection(id: string) {
    if (selectedModIds.has(id)) selectedModIds.delete(id);
    else selectedModIds.add(id);
  }
 
  function toggleSelectionMode() {
    if (selectionMode) clearModSelection();
    else selectionMode = true;
  }
 
  function clearModSelection() {
    selectedModIds.clear();
    selectionMode = false;
  }
 
  function handleSelectAllVisible() {
    // Capture first: the derived flips as soon as the first id is removed.
    const deselect = allVisibleSelected;
    for (const { mod } of visibleFlat) {
      if (deselect) selectedModIds.delete(mod.id);
      else selectedModIds.add(mod.id);
    }
  }
 
  function handleDeleteSelected() {
    const ids = selectedMods.map((m) => m.id);
    clearModSelection();
    ids.forEach((id) => saved.remove(id));
  }
 
  function handleEnableSelected() {
    for (const mod of selectedMods) {
      const packIds = (mod.packages ?? []).map((pack) => pack.id);
      if (packIds.length > 0) saved.setAsEnabledPkgs(mod.id, packIds);
    }
    clearModSelection();
  }
 
  function handleDisableSelected() {
    for (const mod of selectedMods) {
      if (isEnabled(mod)) saved.setAsEnabledPkgs(mod.id, []);
    }
    clearModSelection();
  }
 
  let presetDialogOpen = $state(false);
  let presetName = $state("");
  let viewingPreset = $state<number | undefined>(undefined);
 
  // modId -> chosen packId, for mods that are off and have several packs
  let packChoices = $state<Record<string, string>>({});
 
  const packLabel = (pack: Pack) => pack.name ?? pack.id;
  const needsPackChoice = (mod: ModSaved) =>
    !isEnabled(mod) && (mod.packages?.length ?? 0) > 1;
 
  const modsNeedingChoice = $derived(selectedMods.filter(needsPackChoice));
 
  const canCreatePreset = $derived(
    presetName.trim() !== "" &&
      selectedModCount > 0 &&
      modsNeedingChoice.every((mod) => packChoices[mod.id]),
  );
 
  function presetPackIds(mod: ModSaved): string[] {
    const packs = mod.packages ?? [];
    if (isEnabled(mod)) return packs.filter(isActivePack).map((p) => p.id);
    if (packs.length === 1) return [packs[0].id];
    return packChoices[mod.id] ? [packChoices[mod.id]] : [];
  }
 
  function closePresetDialog() {
    presetDialogOpen = false;
    presetName = "";
    packChoices = {};
  }
 
  function handleCreatePreset() {
    if (!canCreatePreset) return;
    saved.createPresetFromItems(
      presetName.trim(),
      toolId,
      selectedMods.map((mod) => ({
        modId: mod.id,
        packIds: presetPackIds(mod),
      })),
    );
    closePresetDialog();
    clearModSelection();
  }
 
  const totalSize = $derived(
    sumBy(
      mods.flatMap((m) => m.packages ?? []),
      (pack) =>
        pack.storage === StorageState.StorageCold
          ? pack.archiveSize
          : pack.manifest.size,
    ),
  );
 
  const config = $derived.by<GameConfig>(() => {
    const presets = Object.values(saved.presets).filter(
      (preset) => preset.toolId === toolId,
    );
    const viewedPreset = presets.find((p) => p.id === viewingPreset);
    const presetModIds = new Set(viewedPreset?.mods?.map((m) => m.modId));
 
    return {
      game: {
        id: toolId,
        name: toolName(toolId),
        image: toolBanner(toolId),
      },
      presets,
      viewedPreset,
      mods: viewedPreset
        ? mods.filter((m) => presetModIds.has(m.id))
        : mods.filter(isEnabled),
      stats: {
        loaded: allStats.enabled,
        total: mods.length,
        sizeOnDisk: formatSize(totalSize),
        overlaps: 0,
        updates: 0,
      },
      launchVanilla: false,
    };
  });
 
  const handleSubfolderSelected = (folder: Subfolder) =>
    (selectedItemId = folder.id);
  const handleCharacterSelected = (folder: CharacterFolder) =>
    (selectedItemId = folder.name);
 
  function handleModCardClick(mod: ModSaved) {
    if (selectionMode) toggleModSelection(mod.id);
    else selectedModId = mod.id;
  }
 
  function handleRetryInstall(mod: ModSaved) {
    if (!mod.submissionFiles || !mod.packages) return;
    saved.add(mod.submission, mod.submissionFiles, mod.packages);
  }
 
  function handleOpenLink(mod: ModSaved) {
    const link = modLink(mod);
    if (link) Browser.OpenURL(link);
  }
 
  function handleWindowKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "f") {
      e.preventDefault();
      modSearchEl?.focus();
      modSearchEl?.select();
    }
  }
</script>
 
<svelte:window onkeydown={handleWindowKeydown} />
 
{#snippet countBadge(stats: Stats)}
  {#if stats.total > 0}
    <span
      class="ms-auto shrink-0 text-xs tabular-nums
        {stats.enabled > 0 ? 'text-primary' : 'text-muted-foreground'}"
    >
      {stats.enabled}/{stats.total}
    </span>
  {/if}
{/snippet}
 
{#snippet tabButton(
  label: string,
  active: boolean,
  onclick: () => void,
  count?: number,
)}
  <button
    type="button"
    class="flex items-center gap-1.5 border-b-2 pb-1 text-sm font-medium transition-colors
      {active
      ? 'border-primary text-foreground'
      : 'border-transparent text-muted-foreground hover:text-foreground'}"
    {onclick}
  >
    {label}
    {#if count !== undefined}
      <span class="text-xs tabular-nums text-muted-foreground">{count}</span>
    {/if}
  </button>
{/snippet}
 
{#snippet sidebarSection(
  title: string,
  count: number,
  content: import("svelte").Snippet,
)}
  <Collapsible.Root open class="group/collapsible">
    <Sidebar.Group>
      <Sidebar.GroupLabel>
        {#snippet child({ props })}
          <Collapsible.Trigger {...props}>
            {title} · {count}
            <ChevronDown
              class="ms-auto transition-transform group-data-[state=open]/collapsible:rotate-180"
            />
          </Collapsible.Trigger>
        {/snippet}
      </Sidebar.GroupLabel>
      <Collapsible.Content>
        <Sidebar.GroupContent>
          <Sidebar.Menu>
            {@render content()}
          </Sidebar.Menu>
        </Sidebar.GroupContent>
      </Collapsible.Content>
    </Sidebar.Group>
  </Collapsible.Root>
{/snippet}
 
{#snippet modCard(group: (typeof visibleGroups)[number], mod: ModSaved)}
  {@const isSelected = selectedModIds.has(mod.id)}
  <div class="relative isolate h-full min-w-0 rounded-lg">
    <div class="h-full min-w-0 *:h-full *:w-full *:min-w-0">
      <ModCard
        item={{ name: group.name, folder: group.folder, local: mod }}
        conflict={showConflictUi && conflictModIds.has(mod.id)}
        onClick={() => handleModCardClick(mod)}
        onRetry={() => handleRetryInstall(mod)}
        onDelete={() => saved.remove(mod.id)}
        onToggle={() => saved.toggleAllPkgs(mod.id)}
      />
    </div>
 
    {#if selectionMode}
      {#if isSelected}
        <div
          aria-hidden="true"
          class="pointer-events-none absolute inset-0 z-20 rounded-lg border-2 border-primary shadow-sm"
        ></div>
      {/if}
      <Button
        type="button"
        variant={isSelected ? "default" : "outline"}
        size="icon"
        class="absolute right-2 top-2 z-30 size-8 shadow-md"
        aria-label={isSelected ? "Deselect mod" : "Select mod"}
        aria-pressed={isSelected}
        onclick={(e) => {
          e.stopPropagation();
          toggleModSelection(mod.id);
        }}
      >
        {#if isSelected}
          <Check class="size-4" />
        {/if}
      </Button>
    {/if}
  </div>
{/snippet}
 
<div class="overflow-clip">
  <Resizable.PaneGroup
    direction="horizontal"
    class="h-full max-h-[90vh] min-h-[90vh] w-full overflow-clip rounded-sm border"
  >
    <!-- Left: sidebar -->
    <Resizable.Pane defaultSize={20} minSize={10} maxSize={40}>
      <div class="flex h-full min-h-0 flex-col">
        <div class="flex flex-col gap-2 border-b p-2">
          <div class="flex items-center gap-4 px-1">
            {@render tabButton(
              "With mods",
              sidebarTab === "with-mods",
              () => (sidebarTab = "with-mods"),
            )}
            {@render tabButton(
              "All",
              sidebarTab === "all",
              () => (sidebarTab = "all"),
            )}
          </div>
 
          <div class="relative">
            <Search
              class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              bind:value={sidebarSearch}
              placeholder="Find character or folder…"
              class="h-8 px-8"
            />
            {#if sidebarSearch}
              <button
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                aria-label="Clear search"
                onclick={() => (sidebarSearch = "")}
              >
                <X class="size-4" />
              </button>
            {/if}
          </div>
        </div>
 
        <div class="min-h-0 flex-1 overflow-y-auto">
          <Sidebar.Group>
            <Sidebar.GroupContent>
              <Sidebar.Menu>
                <Sidebar.MenuItem>
                  <Sidebar.MenuButton
                    isActive={selected.type === "nothing"}
                    onclick={() => (selectedItemId = undefined)}
                  >
                    <span class="font-medium">All mods</span>
                    {@render countBadge(allStats)}
                  </Sidebar.MenuButton>
                </Sidebar.MenuItem>
              </Sidebar.Menu>
            </Sidebar.GroupContent>
          </Sidebar.Group>
 
          {#snippet otherItems()}
            {#each filteredSubfolders as [name, folder] (name)}
              <Sidebar.MenuItem>
                <Sidebar.MenuButton
                  isActive={folder.name === selected.value?.name}
                  onclick={() => handleSubfolderSelected(folder)}
                >
                  <span class="truncate">{name}</span>
                  {@render countBadge(statsFor(folder.name))}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>
            {:else}
              <p class="px-2 py-1 text-xs text-muted-foreground">
                No folders found
              </p>
            {/each}
          {/snippet}
 
          {#snippet characterItems()}
            {#each filteredCharacters as [name, folder] (name)}
              <Sidebar.MenuItem>
                <Sidebar.MenuButton
                  isActive={folder.name === selected.value?.name}
                  onclick={() => handleCharacterSelected(folder)}
                >
                  {#if folder.icon}
                    <img
                      class="size-8 rounded-sm object-contain"
                      src={folder.icon}
                      alt={name}
                    />
                  {/if}
                  <span class="truncate">{name}</span>
                  {@render countBadge(statsFor(folder.name))}
                </Sidebar.MenuButton>
              </Sidebar.MenuItem>
            {:else}
              <p class="px-2 py-1 text-xs text-muted-foreground">
                No characters found
              </p>
            {/each}
          {/snippet}
 
          {@render sidebarSection(
            "Other",
            filteredSubfolders.length,
            otherItems,
          )}
          {@render sidebarSection(
            "Characters",
            filteredCharacters.length,
            characterItems,
          )}
        </div>
      </div>
    </Resizable.Pane>
 
    <Resizable.Handle withHandle />
 
    <!-- Center: mod library -->
    <Resizable.Pane defaultSize={60} minSize={40}>
      <div class="flex h-full min-h-0 w-full flex-col overflow-clip">
        <ModLibraryHeader
          {selected}
          image={config.game.image}
          charactersCount={characters.length}
          subfoldersCount={subfolders.length}
          mods={visibleCount}
          game={config.game.name.toUpperCase()}
          loaded={loadedInSelection}
        />
 
        <!-- Toolbar -->
        <div class="flex flex-col gap-3 border-b p-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-4">
              {#each modFilterTabs as tab (tab.id)}
                {@render tabButton(
                  tab.label,
                  modFilter === tab.id,
                  () => (modFilter = tab.id),
                  filterCounts[tab.id],
                )}
              {/each}
            </div>
 
            <div class="flex flex-wrap items-center gap-2">
              <div class="relative w-full sm:w-64">
                <Search
                  class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                  bind:ref={modSearchEl}
                  bind:value={modSearch}
                  placeholder="Search mods… (Ctrl+F)"
                  class="h-8 px-8"
                />
                {#if modSearch}
                  <button
                    type="button"
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    aria-label="Clear search"
                    onclick={() => (modSearch = "")}
                  >
                    <X class="size-4" />
                  </button>
                {/if}
              </div>
 
              <Button
                variant="ghost"
                size="icon"
                class="relative size-8 shrink-0"
                aria-label="Mod conflicts"
                title="Mod conflicts"
                onclick={() => (conflictsOpen = true)}
              >
                <AlertTriangle
                  class="size-4 {conflicts.length > 0 ? 'text-amber-500' : ''}"
                />
                {#if showConflictUi && conflicts.length > 0}
                  <span
                    class="absolute -right-1 -top-1 flex size-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-semibold text-white"
                  >
                    {conflicts.length}
                  </span>
                {/if}
              </Button>
 
              <Button
                variant={selectionMode ? "secondary" : "ghost"}
                size="sm"
                class="h-8 shrink-0 gap-1.5"
                aria-pressed={selectionMode}
                onclick={toggleSelectionMode}
              >
                <Check class="size-4" />
                Select
              </Button>
 
              <Button
                variant={groupMods ? "secondary" : "ghost"}
                size="sm"
                class="h-8 shrink-0 gap-1.5"
                aria-pressed={groupMods}
                title={groupMods ? "Ungroup mods" : "Group mods by character"}
                onclick={() => (groupMods = !groupMods)}
              >
                <Layers class="size-4" />
                Group
              </Button>
            </div>
          </div>
 
          {#if selectionMode}
            <div
              class="flex flex-wrap items-center justify-between gap-2 rounded-md border bg-muted/30 px-3 py-2"
            >
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-sm font-medium tabular-nums">
                  {selectedModCount} selected
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8"
                  onclick={handleSelectAllVisible}
                >
                  {allVisibleSelected ? "Deselect visible" : "Select visible"}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 gap-1"
                  onclick={clearModSelection}
                >
                  <X class="size-4" />
                  Clear
                </Button>
              </div>
 
              <div class="flex flex-wrap items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  class="h-8 gap-1"
                  disabled={selectedModCount === 0}
                  onclick={handleEnableSelected}
                >
                  <Power class="size-4" />
                  Enable all packs
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  class="h-8 gap-1"
                  disabled={selectedModCount === 0}
                  onclick={handleDisableSelected}
                >
                  <PowerOff class="size-4" />
                  Disable all packs
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  class="h-8 gap-1"
                  disabled={selectedModCount === 0}
                  onclick={() => (presetDialogOpen = true)}
                >
                  <Plus class="size-4" />
                  Add to preset
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  class="h-8 gap-1"
                  disabled={selectedModCount === 0}
                  onclick={handleDeleteSelected}
                >
                  <Trash2 class="size-4" />
                  Delete
                </Button>
              </div>
            </div>
          {/if}
        </div>
 
        <!-- Mod grid -->
        <div class="min-h-0 flex-1 overflow-y-auto p-3">
          {#if visibleCount === 0}
            <p class="py-12 text-center text-sm text-muted-foreground">
              No mods match the current filters.
            </p>
          {:else if groupMods}
            <div class="flex flex-col gap-4">
              {#each visibleGroups as group (group.name)}
                {@const stats = groupStats(group.name)}
                <section class="flex flex-col gap-2">
                  <h2
                    class="flex items-baseline gap-2 px-1 text-xs font-semibold uppercase tracking-widest text-muted-foreground"
                  >
                    <span class="truncate">{group.name}</span>
                    <span
                      class="tabular-nums {stats.enabled > 0
                        ? 'text-primary'
                        : ''}"
                    >
                      {stats.enabled}/{stats.total}
                    </span>
                  </h2>
                  <div class={gridClass}>
                    {#each group.items as mod (`${group.name}:${mod.id}`)}
                      {@render modCard(group, mod)}
                    {/each}
                  </div>
                </section>
              {/each}
            </div>
          {:else}
            <div class={gridClass}>
              {#each visibleFlat as { group, mod } (`${group.name}:${mod.id}`)}
                {@render modCard(group, mod)}
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </Resizable.Pane>
 
    <Resizable.Handle withHandle />
 
    <!-- Right: config -->
    <Resizable.Pane defaultSize={20} minSize={10} maxSize={40}>
      <LibraryConfig
        {config}
        {selectedMod}
        folders={["Character", "Others"]}
        onTogglePack={(mod: ModSaved, pkg: Pack) =>
          saved.enablePkg(mod.id, pkg.id)}
        onselectmod={(id: string) => (selectedModId = id)}
        onclosemod={() => (selectedModId = undefined)}
        onvanillachange={() => {}}
        onopenfolder={(mod: ModSaved) => ModService.OpenModFolder(mod.id)}
        onopenlink={handleOpenLink}
        onstart={() => ModService.StartGame(toolId)}
        onSelectPreset={(id) => (viewingPreset = id)}
      />
    </Resizable.Pane>
  </Resizable.PaneGroup>
</div>
 
<Dialog.Root bind:open={conflictsOpen}>
  <Dialog.Content class="sm:max-w-xl">
    <Dialog.Header>
      <Dialog.Title>
        {conflicts.length} character{conflicts.length === 1 ? "" : "s"} with several
        mods or packs on
      </Dialog.Title>
      <Dialog.Description>
        These characters have more than one mod or pack on. They usually clash
        in game, so keep one on for each. Click a name to open it.
      </Dialog.Description>
    </Dialog.Header>
 
    {#if conflicts.length > 0}
      <div class="flex max-h-80 flex-col gap-4 overflow-y-auto">
        {#each conflicts as c (c.folderName)}
          <div class="flex flex-col gap-2">
            <p class="text-sm text-muted-foreground">
              {c.name} · {c.packCount} packs on
              {#if c.mods.length > 1}across {c.mods.length} mods{/if}
            </p>
            <div class="flex flex-wrap gap-2">
              {#each c.mods as mod (mod.id)}
                {@const n = activePackCount(mod)}
                <Button
                  variant="secondary"
                  size="sm"
                  class="rounded-full font-semibold"
                  onclick={() => openConflictMod(c, mod)}
                >
                  {mod.submission?.name ?? mod.id}
                  {#if n > 1}
                    <span class="font-normal text-muted-foreground"
                      >· {n} packs</span
                    >
                  {/if}
                </Button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
 
    <div class="flex items-center justify-between gap-4 border-t pt-3">
      <div class="space-y-0.5">
        <p class="text-sm font-semibold">Don't highlight conflicts</p>
        <p class="text-xs text-muted-foreground">
          Hides the yellow tag on mods and the red number on the warning button.
          This list still works.
        </p>
      </div>
      <Switch
        bind:checked={hideConflicts}
        aria-label="Don't highlight conflicts"
      />
    </div>
  </Dialog.Content>
</Dialog.Root>
 
<Dialog.Root bind:open={presetDialogOpen}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>Create preset</Dialog.Title>
      <Dialog.Description>
        Create a new preset containing {selectedModCount} selected mods. Enabled
        mods keep their enabled packs, and mods with a single pack use it.
      </Dialog.Description>
    </Dialog.Header>
 
    <Input
      bind:value={presetName}
      placeholder="Preset name"
      onkeydown={(e) => e.key === "Enter" && handleCreatePreset()}
    />
 
    {#if modsNeedingChoice.length > 0}
      <div class="flex max-h-64 flex-col gap-3 overflow-y-auto">
        <p class="text-sm text-muted-foreground">
          These mods are off and have several packs. Pick the one to include.
        </p>
        {#each modsNeedingChoice as mod (mod.id)}
          <div class="flex flex-col gap-1.5">
            <p class="truncate text-sm font-medium">
              {mod.submission?.name ?? mod.id}
            </p>
            <div class="flex flex-wrap gap-2">
              {#each mod.packages ?? [] as pack (pack.id)}
                <Button
                  variant={packChoices[mod.id] === pack.id
                    ? "default"
                    : "outline"}
                  size="sm"
                  aria-pressed={packChoices[mod.id] === pack.id}
                  onclick={() => (packChoices[mod.id] = pack.id)}
                >
                  {packLabel(pack)}
                </Button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
 
    <Dialog.Footer>
      <Button variant="outline" onclick={closePresetDialog}>Cancel</Button>
      <Button disabled={!canCreatePreset} onclick={handleCreatePreset}>
        Create preset
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
 