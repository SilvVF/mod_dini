<script lang="ts">
  import "./layout.css";
  import favicon from "#lib/assets/favicon.svg";
  import type { LayoutProps } from "./$types";
  import { setSavedMods } from "#lib/contexts/saved-mods.svelte.ts";
  import { setSettings, getSettings } from "#lib/contexts/settings.svelte.ts";
  import SettingsPage from "#lib/components/SettingsPage.svelte";
  import { Dialog, DialogContent } from "#lib/components/ui/dialog/index.ts";
  import Button from "#lib/components/ui/button/button.svelte";
  import { Globe, Settings, X } from "@lucide/svelte";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
  import { SvelteQueryDevtools } from "@tanstack/svelte-query-devtools";

  let { children }: LayoutProps = $props();

  const queryClient = new QueryClient();

  setSavedMods();
  setSettings();

  const settings = getSettings();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<QueryClientProvider client={queryClient}>
  <div class="flex flex-col overflow-clip">
    <header
      class="flex h-[10vh] shrink-0 items-center justify-start border-b px-4"
    >
      <Button variant="ghost" size="icon" onclick={settings.open}>
        <Settings class="size-4" />
      </Button>

      <Button variant="ghost" size="icon" href="/source/gamebanana/30305">
        <Globe class="size-4" />
      </Button>
    </header>

    <main class="min-h-0 min-w-0 max-h-[90vh] flex-1 overflow-y-auto">
      {@render children()}
    </main>
  </div>
  <Dialog
    open={settings.isOpen}
    onOpenChange={(open) => {
      if (open) settings.open();
      else settings.close();
    }}
  >
    <DialogContent
      class="max-h-[90vh] overflow-y-auto border-border bg-background p-0 sm:max-w-5xl"
    >
      <SettingsPage />

      <Button
        variant="ghost"
        size="icon"
        class="absolute right-4 top-4"
        onclick={settings.close}
      >
        <X class="size-4" />
      </Button>
    </DialogContent>
  </Dialog>
  <SvelteQueryDevtools />
</QueryClientProvider>
