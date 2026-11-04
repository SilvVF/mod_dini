<script lang="ts">
  import type { PageProps } from "./$types";
  import * as Resizable from "#lib/components/ui/resizable/index.js";
  import * as Sidebar from "#lib/components/ui/sidebar/index.ts";
  import { ChevronDown } from "@lucide/svelte";
  import { Collapsible } from "bits-ui";
  import type { CharacterFolder } from "#lib/bindings/changeme/models.ts";
  import { getSavedMods } from "#lib/contexts/saved-mods.svelte.ts";
  import { groupModsByCharacter } from "#lib/components/dini/utils.ts";

  let { data }: PageProps = $props();

  const characters = $derived<[string, CharacterFolder][]>(
    Object.entries(data.characters.characters ?? {}).filter(
      (entry): entry is [string, CharacterFolder] => entry[1] != null,
    ),
  );

  const saved = getSavedMods();

  let selectedCharacter = $state("");

  const groupedMods = $derived(groupModsByCharacter(saved.mods, characters));

  const selectedMods = $derived(
    groupedMods.filter((m) =>
      selectedCharacter === "" ? true : m.name === selectedCharacter,
    ),
  );

  $inspect(selectedMods);
  $inspect(groupedMods);

  function handleCharacterSelected(name: string) {
    if (selectedCharacter === name) {
      selectedCharacter = "";
    } else {
      selectedCharacter = name;
    }
  }
</script>

<div class="overflow-clip">
  <Resizable.PaneGroup
    direction="horizontal"
    class="h-full rounded-sm border w-full overflow-clip max-h-[90vh]"
  >
    <Resizable.Pane defaultSize={20} maxSize={40} minSize={10}>
      <div class="overflow-y-scroll h-[90vh]">
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
                          isActive={selectedCharacter === name}
                          onclick={() => handleCharacterSelected(name)}
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
      <div class="flex h-full items-center justify-center p-2">
        <div>Selected {selectedCharacter}</div>
        {#each selectedMods as mod}
          <div>{mod.items}</div>
        {/each}
      </div>
    </Resizable.Pane>
    <Resizable.Handle withHandle />
    <Resizable.Pane defaultSize={20} maxSize={40} minSize={10}>
      <div class="flex h-full items-center justify-center p-2">
        <span class="font-semibold">Sidebar</span>
      </div>
    </Resizable.Pane>
  </Resizable.PaneGroup>
</div>
