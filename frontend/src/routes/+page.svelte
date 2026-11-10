<script lang="ts">
  import type { PageProps } from "./$types";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import { ChevronDown } from "@lucide/svelte";
  import { Collapsible } from "bits-ui";
  import type {
    CharacterFolder,
    ModSaved,
    Pack,
    Subfolder,
  } from "#lib/bindings/changeme/models.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import ModCard from "#lib/components/dini/ModCard.svelte";
  import ModLibraryHeader from "#lib/components/dini/ModLibraryHeader.svelte";
  import { sumBy } from "#lib/ts-ext.ts";
  import {
    groupModsByFolders,
    type SelectedItem,
  } from "#lib/contexts/group-mods.ts";
  import LibraryConfig from "#lib/components/dini/LibraryConfig.svelte";
  import type { GameConfig } from "#lib/components/dini/types.ts";

  let { data }: PageProps = $props();

  const saved = getSavedModsContext();

  let selectedItemId = $state<string | undefined>(undefined);
  const selected = $derived.by<SelectedItem>(() => {
    if (!selectedItemId) return { type: "nothing", value: null };
    const subfolder = {
      type: "character",
      value: data.subfolders[selectedItemId],
    };
    const character = {
      type: "subfolder",
      value: data.characters[selectedItemId],
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

  const characters = $derived(Object.entries(data.characters));
  const subfolders = $derived(Object.entries(data.subfolders));

  const groupedMods = $derived(
    groupModsByFolders(saved.mods, data.rootFolder, data.characters, data.subfolders),
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

  let config = $derived<GameConfig>({
    game: {
      id: "zzmi",
      name: "Zenless Zone Zero",
      image: "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/66868c3874664.jpg",
    },
    preset: {
      id: "p1",
      name: "Preset 1",
      modIds: [],
    },
    mods: groupedMods.flatMap((group) => group.items),
    stats: {
      loaded: 4,
      total: 6,
      sizeOnDisk: "1.3 GB",
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
