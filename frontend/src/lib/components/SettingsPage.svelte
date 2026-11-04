<script lang="ts">
  import { Button } from "#lib/components/ui/button/index.ts";
  import { Input } from "#lib/components/ui/input/index.ts";
  import { getSettings } from "#lib/contexts/settings.svelte.ts";
  import { Folder, Search, X } from "@lucide/svelte";

  const settings = getSettings();
  const folders = $derived(Object.entries(settings.folders));

  function changeFolder(id: string) {
    settings.setFolder(id);
  }

  function useXXMI(id: string) {}
</script>

<div class="min-w-0 space-y-6">
  <div class="relative">
    <Search
      class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
    />

    <Input
      bind:value={settings.query}
      placeholder="Search settings..."
      aria-label="Search settings"
      class="pl-9 pr-9"
    />

    {#if settings.query}
      <Button
        variant="ghost"
        size="icon"
        class="absolute right-1 top-1/2 size-7 -translate-y-1/2"
        aria-label="Clear search"
        onclick={() => (settings.query = "")}
      >
        <X class="size-4" />
      </Button>
    {/if}
  </div>

  <section class="space-y-3">
    <h2 class="text-sm font-semibold tracking-wide text-muted-foreground">
      Folders
    </h2>

    <div class="divide-y rounded-lg border">
      {#each folders as [id, folder] (id)}
        <div class="flex flex-wrap items-center gap-4 p-4 sm:flex-nowrap">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Folder class="size-4" />
          </div>

          <div class="min-w-0 flex-1 space-y-1 text-foreground">
            <p class="font-medium">{folder.title}</p>

            <p class="truncate text-sm text-muted-foreground" title={folder.path}>
              {folder.path}
            </p>
          </div>

          <div class="flex shrink-0 items-center gap-2">
            <Button variant="outline" size="sm" onclick={() => changeFolder(id)}>
              Change
            </Button>

            <Button
              variant="ghost"
              size="sm"
              class="text-muted-foreground"
              onclick={() => useXXMI(id)}
            >
              Use XXMI's
            </Button>
          </div>
        </div>
      {/each}
    </div>
  </section>
</div>