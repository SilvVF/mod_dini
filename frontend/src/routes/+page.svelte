<script lang="ts">
  import type { PageProps } from "./$types";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import { ChevronDown } from "@lucide/svelte";
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

  let selectedItemId = $state<string | undefined>(undefined);
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

  const selectedGroups = $derived(
    groupedMods.filter((m) => {
      return selected.type === "nothing"
        ? true
        : m.folder.name === selected.value.name;
    }),
  );

  const selectedCount = $derived(
    sumBy(selectedGroups, (group) => group.items.length),
  );

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

  function handleVanillaChange(v: boolean) {}
</script>

<div class="overflow-clip">
  <Resizable.PaneGroup
    direction="horizontal"
    class="h-full rounded-sm border w-full overflow-clip min-h-[90vh] max-h-[90vh]"
  >
    <Resizable.Pane defaultSize={20} maxSize={40} minSize={10}>
      <div class="overflow-y-scroll h-full">
        <Collapsible.Root open class="group/collapsible">
          <Sidebar.Group>
            <Sidebar.GroupLabel>
              {#snippet child({ props })}
                <Collapsible.Trigger {...props}>
                  Other
                  <ChevronDown
                    class="ms-auto transition-transform group-data-[state=open]/collapsible:rotate-180"
                  />
                </Collapsible.Trigger>
              {/snippet}
            </Sidebar.GroupLabel>
            <Collapsible.Content>
              <Sidebar.GroupContent>
                {#each subfolders as [name, folder] (name)}
                  <Sidebar.MenuItem>
                    <Sidebar.MenuButton
                      isActive={folder.name === selected.value?.name}
                      onclick={() => handleSubfolderSelected(folder)}
                    >
                      <span>{name}</span>
                    </Sidebar.MenuButton>
                  </Sidebar.MenuItem>
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
                  Characters
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
                    {#each characters as [name, folder] (name)}
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
                          <span>{name}</span>
                        </Sidebar.MenuButton>
                      </Sidebar.MenuItem>
                    {/each}
                  </Sidebar.Menu>
                </Sidebar.GroupContent>
              </Sidebar.Group>
            </Collapsible.Content>
          </Sidebar.Group>
        </Collapsible.Root>
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
          mods={selectedCount}
          game={config.game.name.toUpperCase()}
          loaded={0}
        />
        <div
          class="grid min-h-0 flex-1 w-full p-2 grid-cols-1
    auto-rows-min items-stretch gap-4
    overflow-y-auto
    sm:grid-cols-2
    md:grid-cols-3
    lg:grid-cols-3
    xl:grid-cols-4
    2xl:grid-cols-5"
        >
          {#each selectedGroups as group}
            {#each group.items as mod (`${group.name}:${mod.id}`)}
              <ModCard
                item={{
                  name: group.name,
                  folder: group.folder,
                  local: mod,
                }}
                onClick={() => hanldeModCardClick(mod)}
                onSettings={() => {}}
                onDelete={() => {}}
                onToggle={() => handleModToggle(mod)}
              />
            {/each}
          {/each}
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
