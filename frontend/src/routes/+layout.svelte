<script lang="ts">
  import "./layout.css";
  import favicon from "#lib/assets/favicon.svg";
  import type { LayoutProps } from "./$types";
  import { setSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import Button from "#lib/components/ui/button/button.svelte";
  import {
    Keyboard,
    LoaderCircle,
    MoonIcon,
    Settings,
    SunIcon,
    TriangleAlert,
  } from "@lucide/svelte";
  import { createQuery, QueryClientProvider } from "@tanstack/svelte-query";
  import { SvelteQueryDevtools } from "@tanstack/svelte-query-devtools";
  import {
    getSettingsContext,
    setSettingsContext,
  } from "#lib/contexts/settings.svelte.ts";
  import * as Dialog from "#lib/components/ui/dialog/index.ts";
  import { navigating, page } from "$app/state";
  import { Category } from "#lib/bindings/changeme/models.ts";
  import { goto } from "$app/navigation";
  import SettingsView from "#lib/components/settings/SettingsView.svelte";
  import { cn } from "#lib/utils.ts";
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import {
    appState,
    getRouteForProvider,
    setActiveTool,
    tools,
  } from "#lib/state/app.svelte.ts";
  import {
    getProcessContext,
    setProcessContext,
  } from "#lib/contexts/process.svelte.ts";
  import { queryClient } from "#lib/components/dini/utils.ts";
  import NavBar from "#lib/components/dini/NavBar.svelte";
  import ModViewer from "#lib/components/viewer/ModViewer.svelte";
  import { toggleMode, ModeWatcher } from "mode-watcher";
  import ThemeSelect from "#lib/components/dini/ThemeSelect.svelte";
  import { appTheme } from "#lib/state/theme.svelte.ts";
  import { onMount } from "svelte";

  let { children, data }: LayoutProps = $props();

  const initial = appTheme.current === "default" ? "" : appTheme.current;

  onMount(() => {
    appTheme.apply();
  });

  setSavedModsContext(
    () => data.mods,
    () => data.presets,
  );
  setSettingsContext(() => data.settings);
  setProcessContext();

  const settings = getSettingsContext();
  const processState = getProcessContext();

  $effect(() => {
    appState.activeTool = settings.get("tool.active")?.value;
  });

  const activeToolConfig = $derived(
    tools.find((t) => t.id === appState.activeTool)!,
  );
  const category = $derived(page.state.settings?.category);

  let loading = $state(false);
  let progress = $state(0);

  const preview = createQuery(
    () => ({
      queryKey: [
        "preview3d",
        page.state.preview?.modId,
        page.state.preview?.packId,
      ],
      queryFn: () =>
        ModService.PreviewPack(
          page.state.preview!.modId,
          page.state.preview!.packId,
        ),
      gcTime: 60_000,
      retry: false,
      enabled: page.state.preview !== undefined,
    }),
    () => queryClient,
  );

  const previewError = $derived(
    preview.error instanceof Error
      ? preview.error.message
      : String(preview.error),
  );

  $effect(() => {
    if (navigating.to) {
      loading = true;
      progress = 10;
      const timer = setInterval(() => (progress += (90 - progress) * 0.1), 100);
      return () => clearInterval(timer);
    }
    if (!loading) return;
    progress = 100;
    const timer = setTimeout(() => {
      loading = false;
      progress = 0;
    }, 200);
    return () => clearTimeout(timer);
  });

  function openSettings(e: MouseEvent, category: Category) {
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

<ModeWatcher defaultTheme={initial} />
<QueryClientProvider client={queryClient}>
  <div class="flex flex-col overflow-clip">
    <header
      class={cn(
        "relative flex h-[10vh] shrink-0 items-center justify-start border-b px-4",
        page.route.id === "/keybinds" && "hidden",
      )}
    >
      <NavBar class="mr-4 w-96 shrink" />
      <ThemeSelect />
      <Button onclick={toggleMode} variant="outline" size="icon">
        <SunIcon
          class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all! dark:scale-0 dark:-rotate-90"
        />
        <MoonIcon
          class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all! dark:scale-100 dark:rotate-0"
        />
        <span class="sr-only">Toggle theme</span>
      </Button>
      <Button
        variant="ghost"
        size="icon"
        href={`/settings/${Category.CategoryMods}`}
        onclick={(e) => openSettings(e, Category.CategoryMods)}
      >
        <Settings class="size-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        onclick={() => ModService.OpenKeybindsWindow()}
      >
        <Keyboard class="size-4" />
      </Button>
      <div class="flex items-center gap-3">
        {#each tools as tool}
          {@const active = appState.activeTool === tool.id}
          <Button
            variant="ghost"
            size="icon"
            class={cn(
              "size-11 shrink-0 rounded-xl border-2 p-0 transition-all",
              active ? "border-primary" : "border-transparent",
            )}
            onclick={() => setActiveTool(tool.id)}
            aria-label={tool.id}
          >
            <span class="relative flex size-9 items-center justify-center">
              <img
                src={tool.src}
                alt={tool.id}
                class={cn(
                  "size-9 rounded-lg object-cover transition-opacity",
                  active ? "opacity-100" : "opacity-50",
                )}
              />
              {#if processState.running(tool.process)}
                <LoaderCircle
                  class="absolute size-5 animate-spin text-primary"
                  strokeWidth={2.5}
                />
              {/if}
            </span>
          </Button>
        {/each}
        {#each activeToolConfig.providers as provider}
          <Button href={getRouteForProvider(provider)}>{provider}</Button>
        {/each}
      </div>

      {#if loading}
        <div class="pointer-events-none absolute inset-x-0 bottom-0 h-0.5">
          <div
            class="bg-primary h-full transition-[width] duration-200 ease-out"
            style:width="{progress}%"
          ></div>
        </div>
      {/if}
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
            onnavigate={(c) =>
              goto(`/settings/${c}`, {
                shallow: true,
                replace: true,
                state: { settings: { category: c } },
              })}
            onclose={() => history.back()}
          />
        {/if}
      </Dialog.Content>
    </Dialog.Root>

    <Dialog.Root
      open={!!page.state.preview}
      onOpenChange={(open) => {
        if (!open) history.back();
      }}
    >
      <Dialog.Content
        showCloseButton={false}
        escapeKeydownBehavior={preview.data ? "ignore" : "close"}
        class="h-[min(90vh,52rem)] w-[min(96vw,80rem)] max-w-none gap-0 overflow-hidden p-0 sm:max-w-none"
      >
        {#if preview.data}
          <ModViewer
            mod={preview.data.mod}
            pack={preview.data.pack}
            payload={preview.data.payload}
            embedded
            onClose={() => history.back()}
          />
        {:else if preview.isError}
          <div
            class="m-auto flex max-w-md flex-col items-center gap-4 p-6 text-center"
          >
            <TriangleAlert class="text-destructive size-8" />
            <div class="space-y-1">
              <p class="font-medium">Couldn't load the preview</p>
              <p class="text-muted-foreground text-sm wrap-break-word">
                {previewError}
              </p>
            </div>
            <div class="flex gap-2">
              <Button variant="outline" onclick={() => preview.refetch()}>
                Retry
              </Button>
              <Button variant="ghost" onclick={() => history.back()}>
                Close
              </Button>
            </div>
          </div>
        {:else}
          <LoaderCircle class="m-auto size-6 animate-spin" />
        {/if}
      </Dialog.Content>
    </Dialog.Root>

    <SvelteQueryDevtools />
  </div>
</QueryClientProvider>
