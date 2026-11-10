<script lang="ts">
  import "./layout.css";
  import favicon from "#lib/assets/favicon.svg";
  import type { LayoutProps } from "./$types";
  import { setSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import Button from "#lib/components/ui/button/button.svelte";
  import { Globe, Settings } from "@lucide/svelte";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
  import { SvelteQueryDevtools } from "@tanstack/svelte-query-devtools";
  import { setSettingsContext } from "#lib/contexts/settings.svelte.ts";
  import * as Dialog from "#lib/components/ui/dialog/index.ts";
  import { page } from "$app/state";
  import { Category } from "#lib/bindings/changeme/models.ts";
  import { goto } from "$app/navigation";
  import SettingsView from "#lib/components/settings/SettingsView.svelte";

  let { children, data }: LayoutProps = $props();

  const queryClient = new QueryClient();

  setSavedModsContext(() => data.mods);
  setSettingsContext(() => data.settings);
  const category = $derived(page.state.settings?.category);

  function openSettings(
    e: MouseEvent & { currentTarget: HTMLAnchorElement },
    category: Category,
  ) {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0)
      return;
    e.preventDefault();
    goto(`/settings/${category}`, {
      shallow: true,
      state: { settings: { category } },
    });
  }
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<QueryClientProvider client={queryClient}>
  <div class="flex flex-col overflow-clip">
    <header
      class="flex h-[10vh] shrink-0 items-center justify-start border-b px-4"
    >
      <a
        href="/settings/general"
        onclick={(e) => openSettings(e, Category.CategoryMods)}
      >
        <Button variant="ghost" size="icon" href="/settings">
          <Settings class="size-4" />
        </Button>
      </a>

      <Button variant="ghost" size="icon" href="/source/gamebanana/30305">
        <Globe class="size-4" />
      </Button>
    </header>

    <main class="min-h-0 min-w-0 max-h-[90vh] flex-1 overflow-hidden">
      {@render children()}
    </main>
  </div>
  <Dialog.Root
    open={!!page.state.settings}
    onOpenChange={(open) => {
      if (!open) history.back();
    }}
  >
    <Dialog.Content
      showCloseButton={false}
      class="h-[min(85vh,44rem)] w-[min(94vw,60rem)] max-w-none gap-0 overflow-hidden p-0 sm:max-w-none"
    >
      {#if category}
        <SettingsView
          {category}
          onnavigate={(c) => {
            goto(`/settings/${c}`, {
              shallow: true,
              replace: true,
              state: { settings: { category: c } },
            });
          }}
          onclose={() => history.back()}
        />
      {/if}
    </Dialog.Content>
  </Dialog.Root>

  <SvelteQueryDevtools />
</QueryClientProvider>
