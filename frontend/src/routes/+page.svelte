<script lang="ts">
  import type { PageProps } from "./$types";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import { ChevronDown } from "@lucide/svelte";
  import { Collapsible } from "bits-ui";
  import type {
    CharacterFolder,
    ModSaved,
    Subfolder,
  } from "#lib/bindings/changeme/models.ts";
  import { getSavedMods } from "#lib/contexts/saved-mods.svelte.ts";
  import ModCard from "#lib/components/dini/ModCard.svelte";
  import { testModRecord } from "#lib/test/mod-card-test-data.ts";
  import ModLibraryHeader from "#lib/components/dini/ModLibraryHeader.svelte";
  import { sumBy } from "#lib/ts-ext.ts";
  import {
    groupModsByFolders,
    type SelectedItem,
  } from "#lib/contexts/group-mods.ts";
  import LibraryConfig from "#lib/components/dini/LibraryConfig.svelte";
  import type { GameConfig } from "#lib/components/dini/types.ts";

  let { data }: PageProps = $props();

  const subFolders = $derived(
    Object.entries(data.subfolders).filter(
      (entry): entry is [string, Subfolder] => entry[1] != null,
    ),
  );
  const characters = $derived<[string, CharacterFolder][]>(
    Object.entries(data.characters).filter(
      (entry): entry is [string, CharacterFolder] => entry[1] != null,
    ),
  );

  const saved = getSavedMods();

  let selected = $state<SelectedItem>({ type: "nothing", value: null });
  let selectedMod = $state<ModSaved | undefined>(undefined);

  $effect(() => {
    selected;
    selectedMod = undefined;
  });

  const groupedMods = $derived(
    groupModsByFolders(testModRecord(), characters, subFolders),
  );

  const selectedGroups = $derived(
    groupedMods.filter((m) => (!selected ? true : m.folder === selected.value)),
  );

  const selectedCount = $derived(
    sumBy(selectedGroups, (group) => group.items.length),
  );

  let config = $derived<GameConfig>({
    game: {
      id: "zzz",
      name: "Zenless Zone Zero",
      image:
        "https://images.gamebanana.com/img/Webpage/Game/Profile/Background/66868c3874664.jpg",
    },
    preset: {
      id: "p1",
      name: "Preset 1",
      modIds: selectedGroups.flatMap((group) =>
        group.items.map((item) => item.id),
      ),
    },
    mods: selectedGroups.flatMap((group) => group.items),
    stats: {
      loaded: 4,
      total: 6,
      sizeOnDisk: "1.3 GB",
      overlaps: 0,
      updates: 0,
    },
    launchVanilla: false,
  });

  function handleCharacterSelected(folder: CharacterFolder) {
    if (selected.value?.name === folder.name) {
      selected = { type: "nothing", value: null };
    } else {
      selected = { type: "character", value: folder };
    }
  }

  function handleCloseMod() {
    selectedMod = undefined;
  }

  function handleSelectMod(id: string) {
    const mods = groupedMods.flatMap((group) => group.items);
    selectedMod = mods.find((mod) => mod.id === id);
  }

  function hanldeModCardClick(mod: ModSaved) {
    selectedMod = mod;
  }

  function handleVanillaChange(v: boolean) {}

  function hanldeUnloadMod() {}
</script>

<div class="overflow-clip">
  <Resizable.PaneGroup
    direction="horizontal"
    class="h-full rounded-sm border w-full overflow-clip max-h-[90vh]"
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
              <Sidebar.GroupContent />
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
          subfoldersCount={subFolders.length}
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
          {#each groupedMods as group}
            {#each group.items as mod (`${group.name}:${mod.id}`)}
              <ModCard
                item={{
                  name: group.name,
                  folder: group.folder,
                  local: mod,
                }}
                onClick={() => hanldeModCardClick(mod)}
                enabled={false}
                onSettings={() => {}}
                onDelete={() => {}}
                onToggle={() => {}}
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
        enabled={false}
        folders={["Character", "Others"]}
        onselectmod={handleSelectMod}
        onclosemod={handleCloseMod}
        onunloadmod={hanldeUnloadMod}
        onvanillachange={handleVanillaChange}
      />
    </Resizable.Pane>
  </Resizable.PaneGroup>
</div>
