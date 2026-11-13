<script lang="ts">
  import "./layout.css";
  import favicon from "#lib/assets/favicon.svg";
  import type { LayoutProps } from "./$types";
  import { setSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import Button from "#lib/components/ui/button/button.svelte";
  import { Globe, Globe2, Keyboard, Settings } from "@lucide/svelte";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
  import { SvelteQueryDevtools } from "@tanstack/svelte-query-devtools";
  import { setSettingsContext } from "#lib/contexts/settings.svelte.ts";
  import * as Dialog from "#lib/components/ui/dialog/index.ts";
  import { page } from "$app/state";
  import { Category } from "#lib/bindings/changeme/models.ts";
  import { goto } from "$app/navigation";
  import SettingsView from "#lib/components/settings/SettingsView.svelte";
  import { cn } from "#lib/utils.ts";
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import {
    appState,
    getRouteForProvider,
    setActiveTool,
    setToolFromEvent,
    tools,
  } from "#lib/state/app.svelte.ts";
  import { onMount } from "svelte";
  import { Events } from "@wailsio/runtime";

  let { children, data }: LayoutProps = $props();

  const queryClient = new QueryClient();

  setSavedModsContext(() => data.mods);
  setSettingsContext(() => data.settings);

  onMount(() => {
    const unsubscribe = Events.On("active-tool-change", ({ data }) => {
      setToolFromEvent(data);
    });

    return () => unsubscribe();
  });

  const activeToolConfig = $derived(
    tools.find((t) => t.id === appState.activeTool)!,
  );
  const category = $derived(page.state.settings?.category);
  const topBarHidden = $derived(page.route.id !== "/")

  function openKeybindsWindow() {
    ModService.OpenKeybindsWindow();
  }

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
      class={cn(
        "flex h-[10vh] shrink-0 items-center justify-start border-b px-4",
        topBarHidden  ? "hidden" : "",
      )}
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
      <Button variant="ghost" size="icon" href="/source/huihui/zzz">
        <Globe2 class="size-4" />
      </Button>
      <Button variant="ghost" size="icon" onclick={openKeybindsWindow}>
        <Keyboard class="size-4" />
      </Button>
      <div class="flex items-center gap-3">
        {#each tools as tool}
          <Button
            variant="ghost"
            size="icon"
            class={cn(
              "size-11 shrink-0 rounded-xl border-2 p-0 transition-all",
              appState.activeTool === tool.id
                ? "border-primary"
                : "border-transparent",
            )}
            onclick={() => setActiveTool(tool.id)}
            aria-label={tool.id}
          >
            <img
              src={tool.src}
              alt={tool.id}
              class={cn(
                "size-9 rounded-lg object-cover transition-opacity",
                appState.activeTool === tool.id ? "opacity-100" : "opacity-50",
              )}
            />
          </Button>
        {/each}
        {#each activeToolConfig.providers as provider}
          <Button href={getRouteForProvider(provider)}>
            {provider}
          </Button>
        {/each}
      </div>
    </header>

    <main class="min-h-0 overflow-clip">
      {@render children()}
    </main>
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
  </div></QueryClientProvider
>
