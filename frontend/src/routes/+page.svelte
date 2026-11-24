<script lang="ts">
  import type { PageProps } from "./$types";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import { Input } from "#lib/components/ui/input/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { ChevronDown, Layers, Search, X } from "@lucide/svelte";
  import { Collapsible } from "bits-ui";
  import {
    StorageState,
    type CharacterFolder,
    type ModSaved,
    type Pack,
    type Subfolder,
  } from "#lib/bindings/changeme/models.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import ModCard from "#lib/components/dini/ModCard.svelte";
  import ModLibraryHeader from "#lib/components/dini/ModLibraryHeader.svelte";
  import { compact, sumBy } from "#lib/ts-ext.ts";
  import {
    groupModsByFolders,
    type SelectedItem,
  } from "#lib/contexts/group-mods.ts";
  import LibraryConfig from "#lib/components/dini/LibraryConfig.svelte";
  import type { GameConfig } from "#lib/components/dini/types.ts";
  import { formatSize } from "#lib/components/dini/utils.ts";
  import { appState, toolBanner, toolName } from "#lib/state/app.svelte.ts";

  let { data }: PageProps = $props();

  const saved = getSavedModsContext();

  type ModFilter = "all" | "enabled" | "disabled";
  type SidebarTab = "all" | "with-mods";

  const isEnabled = (m: ModSaved) => m.packages?.some((p) => p.enabled) ?? false;

  const gridClass = `grid w-full grid-cols-1 items-stretch gap-4
    sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5`;

  let selectedItemId = $state<string | undefined>(undefined);

  // Sidebar filters
  let sidebarTab = $state<SidebarTab>("all");
  let sidebarSearch = $state("");

  // Mod grid filters
  let modFilter = $state<ModFilter>("all");
  let modSearch = $state("");
  let modSearchEl = $state<HTMLInputElement | null>(null);
  let groupMods = $state(true);

  const toolId = $derived(appState.activeTool);
  const rootFolder = $derived(data.rootFolder[toolId]);
  const subfoldersRecord = $derived(compact(rootFolder?.subfolders ?? {}));
  const charactersRecord = $derived(
    compact(rootFolder.characters.characters ?? {}),
  );

  const selected = $derived.by<SelectedItem>(() => {
    if (!selectedItemId) return { type: "nothing", value: null };
    const subfolder = {
      type: "character",
      value: subfoldersRecord[selectedItemId],
    };
    const character = {
      type: "subfolder",
      value: charactersRecord[selectedItemId],
    };

    return (
      subfolder.value
        ? subfolder
        : character.value
          ? character
          : { type: "nothing", value: null }
    ) as SelectedItem;
  });

  let selectedModId = $state<string | undefined>(undefined);

  $effect(() => {
    selectedItemId;
    selectedModId = undefined;
  });

  const characters = $derived(Object.entries(charactersRecord));
  const subfolders = $derived(Object.entries(subfoldersRecord));

  const mods = $derived(
    Object.values(saved.mods).filter((m) => m.toolId === toolId),
  );
  const groupedMods = $derived(
    groupModsByFolders(mods, rootFolder, charactersRecord, subfoldersRecord),
  );

  // enabled / total mods per folder (keyed by folder name)
  const folderStats = $derived.by(() => {
    const map = new Map<string, { enabled: number; total: number }>();
    for (const group of groupedMods) {
      const s = map.get(group.folder.name) ?? { enabled: 0, total: 0 };
      for (const mod of group.items) {
        s.total++;
        if (isEnabled(mod)) s.enabled++;
      }
      map.set(group.folder.name, s);
    }
    return map;
  });

  const statsFor = (folderName: string) =>
    folderStats.get(folderName) ?? { enabled: 0, total: 0 };

  const allStats = $derived({
    enabled: mods.filter(isEnabled).length,
    total: mods.length,
  });

  // Sidebar filtering
  const sidebarQuery = $derived(sidebarSearch.trim().toLowerCase());

  const filteredSubfolders = $derived(
    subfolders.filter(([name]) => name.toLowerCase().includes(sidebarQuery)),
  );

  const filteredCharacters = $derived(
    characters.filter(([name, folder]) => {
      if (!name.toLowerCase().includes(sidebarQuery)) return false;
      if (sidebarTab === "with-mods" && statsFor(folder.name).total === 0)
        return false;
      return true;
    }),
  );

  // Groups for the current sidebar selection
  const selectedGroups = $derived(
    groupedMods.filter((m) => {
      return selected.type === "nothing"
        ? true
        : m.folder.name === selected.value.name;
    }),
  );

  // Apply the mod search first so the tab counts reflect it
  const modQuery = $derived(modSearch.trim().toLowerCase());

  const searchedGroups = $derived(
    selectedGroups
      .map((group) => ({
        ...group,
        items: group.items.filter((mod) => {
          if (!modQuery) return true;
          return (
            mod.submission?.name?.toLowerCase().includes(modQuery) ||
            mod.submission?.author?.toLowerCase().includes(modQuery) ||
            group.name.toLowerCase().includes(modQuery)
          );
        }),
      }))
      .filter((group) => group.items.length > 0),
  );

  const filterCounts = $derived.by(() => {
    let enabledCount = 0;
    let total = 0;
    for (const group of searchedGroups) {
      for (const mod of group.items) {
        total++;
        if (isEnabled(mod)) enabledCount++;
      }
    }
    return {
      all: total,
      enabled: enabledCount,
      disabled: total - enabledCount,
    };
  });

  const visibleGroups = $derived(
    searchedGroups
      .map((group) => ({
        ...group,
        items: group.items.filter((mod) =>
          modFilter === "all"
            ? true
            : modFilter === "enabled"
              ? isEnabled(mod)
              : !isEnabled(mod),
        ),
      }))
      .filter((group) => group.items.length > 0),
  );

  // Flat list used when grouping is off
  const visibleFlat = $derived(
    visibleGroups.flatMap((group) =>
      group.items.map((mod) => ({ group, mod })),
    ),
  );

  const visibleCount = $derived(
    sumBy(visibleGroups, (group) => group.items.length),
  );

  const loadedInSelection = $derived(
    sumBy(selectedGroups, (group) => group.items.filter(isEnabled).length),
  );

  // enabled/total for a group header, ignoring the tab/search filters
  function groupStats(groupName: string) {
    const full = selectedGroups.find((g) => g.name === groupName);
    const items = full?.items ?? [];
    return { enabled: items.filter(isEnabled).length, total: items.length };
  }

  const selectedMod = $derived(
    selectedModId ? saved.mods[selectedModId] : undefined,
  );

  const enabled = $derived(
    mods.filter((m) =>
      m.packages?.some((p) => p.enabled && m.toolId === toolId),
    ).length,
  );
  const totalSize = $derived.by(() => {
    let total = 0;

    for (const mod of mods) {
      for (const pack of mod.packages ?? []) {
        if (pack.storage === StorageState.StorageCold) {
          total += pack.archiveSize;
        } else {
          total += pack.manifest.size;
        }
      }
    }

    return total;
  });

  let config = $derived<GameConfig>({
    game: {
      id: appState.activeTool,
      name: toolName(appState.activeTool),
      image: toolBanner(appState.activeTool),
    },
    preset: {
      id: "p1",
      name: "Preset 1",
      modIds: [],
    },
    mods: groupedMods.flatMap((group) => group.items),
    stats: {
      loaded: enabled,
      total: mods.length,
      sizeOnDisk: formatSize(totalSize),
      overlaps: 0,
      updates: 0,
    },
    launchVanilla: false,
  });

  const modFilterTabs = $derived<{ id: ModFilter; label: string }[]>([
    { id: "all", label: "All" },
    { id: "enabled", label: "Enabled" },
    { id: "disabled", label: "Disabled" },
  ]);

  function handleSubfolderSelected(folder: Subfolder) {
    selectedItemId = folder.id;
  }

  function handleCharacterSelected(folder: CharacterFolder) {
    selectedItemId = folder.name;
  }

  function handleCloseMod() {
    selectedModId = undefined;
  }

  function handleSelectMod(id: string) {
    selectedModId = id;
  }

  function hanldeModCardClick(mod: ModSaved) {
    selectedModId = mod.id;
  }

  function handlePkgSelected(mod: ModSaved, pkg: Pack) {
    saved.enablePkg(mod.id, pkg.id);
  }

  function handleModToggle(mod: ModSaved) {
    saved.toggleAllPkgs(mod.id);
  }

  function handleRemoveMod(mod: ModSaved) {
    saved.remove(mod.id);
  }

  function handleVanillaChange(v: boolean) {}

  function handleWindowKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "f") {
      e.preventDefault();
      modSearchEl?.focus();
      modSearchEl?.select();
    }
  }
</script>

<svelte:window onkeydown={handleWindowKeydown} />

{#snippet countBadge(stats: { enabled: number; total: number })}
  {#if stats.total > 0}
    <span
      class="ms-auto shrink-0 text-xs tabular-nums {stats.enabled > 0
        ? 'text-primary'
        : 'text-muted-foreground'}"
    >
      {stats.enabled}/{stats.total}
    </span>
  {/if}
{/snippet}

{#snippet modCard(group: (typeof visibleGroups)[number], mod: ModSaved)}
  <ModCard
    item={{
      name: group.name,
      folder: group.folder,
      local: mod,
    }}
    onClick={() => hanldeModCardClick(mod)}
    onSettings={() => {}}
    onDelete={() => handleRemoveMod(mod)}
    onToggle={() => handleModToggle(mod)}
  />
{/snippet}

<div class="overflow-clip">
  <Resizable.PaneGroup
    direction="horizontal"
    class="h-full rounded-sm border w-full overflow-clip min-h-[90vh] max-h-[90vh]"
  >
    <Resizable.Pane defaultSize={20} maxSize={40} minSize={10}>
      <div class="flex h-full min-h-0 flex-col">
        <!-- Sidebar filters -->
        <div class="flex flex-col gap-2 border-b p-2">
          <div class="flex items-center gap-4 px-1">
            <button
              type="button"
              class="border-b-2 pb-1 text-sm font-medium transition-colors {sidebarTab ===
              'with-mods'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'}"
              onclick={() => (sidebarTab = "with-mods")}
            >
              With mods
            </button>
            <button
              type="button"
              class="border-b-2 pb-1 text-sm font-medium transition-colors {sidebarTab ===
              'all'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'}"
              onclick={() => (sidebarTab = "all")}
            >
              All
            </button>
          </div>

          <div class="relative">
            <Search
              class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              bind:value={sidebarSearch}
              placeholder="Find character or folder…"
              class="h-8 pl-8 pr-8"
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

          <Collapsible.Root open class="group/collapsible">
            <Sidebar.Group>
              <Sidebar.GroupLabel>
                {#snippet child({ props })}
                  <Collapsible.Trigger {...props}>
                    Other · {filteredSubfolders.length}
                    <ChevronDown
                      class="ms-auto transition-transform group-data-[state=open]/collapsible:rotate-180"
                    />
                  </Collapsible.Trigger>
                {/snippet}
              </Sidebar.GroupLabel>
              <Collapsible.Content>
                <Sidebar.GroupContent>
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
                </Sidebar.GroupContent>
              </Collapsible.Content>
            </Sidebar.Group>
          </Collapsible.Root>

          <Collapsible.Root open class="group/collapsible">
            <Sidebar.Group>
              <Sidebar.GroupLabel>
                {#snippet child({ props })}
                  <Collapsible.Trigger {...props}>
                    Characters · {filteredCharacters.length}
                    <ChevronDown
                      class="ms-auto transition-transform group-data-[state=open]/collapsible:rotate-180"
                    />
                  </Collapsible.Trigger>
                {/snippet}
              </Sidebar.GroupLabel>
              <Collapsible.Content>
                <Sidebar.Group>
                  <Sidebar.GroupContent>
                    <Sidebar.Menu>
                      {#each filteredCharacters as [name, folder] (name)}
                        <Sidebar.MenuItem>
                          <Sidebar.MenuButton
                            isActive={folder.name === selected.value?.name}
                            onclick={() => handleCharacterSelected(folder)}
                          >
                            {#if folder.icon}
                              <img
                                class="size-8 object-contain rounded-sm"
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
                    </Sidebar.Menu>
                  </Sidebar.GroupContent>
                </Sidebar.Group>
              </Collapsible.Content>
            </Sidebar.Group>
          </Collapsible.Root>
        </div>
      </div>
    </Resizable.Pane>
    <Resizable.Handle withHandle />
    <Resizable.Pane defaultSize={60} minSize={40}>
      <div class="flex h-full min-h-0 w-full overflow-clip flex-col">
        <ModLibraryHeader
          {selected}
          image={config.game.image}
          charactersCount={characters.length}
          subfoldersCount={subfolders.length}
          mods={visibleCount}
          game={config.game.name.toUpperCase()}
          loaded={loadedInSelection}
        />

        <!-- Mod filters -->
        <div
          class="flex flex-wrap items-center justify-between gap-2 border-b px-3 py-2"
        >
          <div class="flex items-center gap-4">
            {#each modFilterTabs as tab (tab.id)}
              <button
                type="button"
                class="flex items-center gap-1.5 border-b-2 pb-1 text-sm font-medium transition-colors {modFilter ===
                tab.id
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'}"
                onclick={() => (modFilter = tab.id)}
              >
                {tab.label}
                <span class="text-xs tabular-nums text-muted-foreground">
                  {filterCounts[tab.id]}
                </span>
              </button>
            {/each}
          </div>

          <div class="flex w-full items-center gap-2 sm:w-auto">
            <div class="relative w-full sm:w-64">
              <Search
                class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                bind:ref={modSearchEl}
                bind:value={modSearch}
                placeholder="Search mods… (Ctrl+F)"
                class="h-8 pl-8 pr-8"
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

        <div class="min-h-0 flex-1 overflow-y-auto p-2">
          {#if visibleCount === 0}
            <p class="py-12 text-center text-sm text-muted-foreground">
              No mods match the current filters.
            </p>
          {:else if groupMods}
            <div class="flex flex-col gap-6">
              {#each visibleGroups as group}
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
            <div class="{gridClass} auto-rows-min">
              {#each visibleFlat as { group, mod } (`${group.name}:${mod.id}`)}
                {@render modCard(group, mod)}
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </Resizable.Pane>
    <Resizable.Handle withHandle />
    <Resizable.Pane defaultSize={20} maxSize={40} minSize={10}>
      <LibraryConfig
        {config}
        {selectedMod}
        folders={["Character", "Others"]}
        onTogglePack={handlePkgSelected}
        onselectmod={handleSelectMod}
        onclosemod={handleCloseMod}
        onvanillachange={handleVanillaChange}
      />
    </Resizable.Pane>
  </Resizable.PaneGroup>
</div>