<script lang="ts">
  import { Category, type Setting } from "#lib/bindings/changeme/index.ts";
  import { getSettingsContext } from "#lib/contexts/settings.svelte.ts";
  import { Inbox, Search, SearchX, X } from "@lucide/svelte";
  import type { HTMLAttributes } from "svelte/elements";
  import { cn } from "#lib/utils.ts";
  import { Button } from "../ui/button";
  import { SettingsMeta } from "./SettingsMeta";
  import { Input } from "../ui/input";
  import SettingRow from "./SettingRow.svelte";
  import { groupBy } from "#lib/ts-ext.ts";

  let {
    category = Category.CategoryGeneral,
    onnavigate,
    onclose,
    class: className,
    ...restProps
  }: Omit<HTMLAttributes<HTMLDivElement>, "onchange" | "onclose"> & {
    category?: Category;
    onnavigate?: (category: Category) => void;
    onclose?: () => void;
  } = $props();

  const settings = getSettingsContext();

  let query = $state("");

  /** While searching, look through every category; otherwise show the selected one. */
  const visible = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (q) {
      return settings.all.filter((s) => {
        return [
          s.title,
          s.description,
          s.section,
          SettingsMeta.categoryLabel(s.category),
          ...(s.keywords ?? []),
        ]
          .filter(Boolean)
          .some((t) => t!.toLowerCase().includes(q));
      });
    } else {
      return settings.inCategory(category);
    }
  });

  const groups = $derived(
    groupBy(
      visible,
      (s) => s.section || SettingsMeta.categoryLabel(s.category),
    ),
  );

  const entries = $derived(Object.entries(groups));
</script>

<div class={cn("flex size-full min-h-0 flex-col", className)} {...restProps}>
  <header class="flex shrink-0 items-center justify-between px-4 pt-4 pb-3">
    <h1 class="font-serif text-xl font-bold">Settings</h1>
    {#if onclose}
      <Button
        variant="outline"
        size="icon"
        aria-label="Close settings"
        onclick={onclose}
      >
        <X class="size-4" />
      </Button>
    {/if}
  </header>
  <div
    class="grid min-h-0 flex-1 grid-cols-[11rem_minmax(0,1fr)] grid-rows-[minmax(0,1fr)] gap-4 px-4 pb-4"
  >
    <nav aria-label="Settings categories" class="min-h-0 space-y-1 overflow-y-auto">
      {#each SettingsMeta.categories as c (c.id)}
        {@const active = !query && c.id === category}
        <button
          type="button"
          aria-current={active ? "page" : undefined}
          onclick={() => {
            query = "";
            onnavigate?.(c.id);
          }}
          class="focus-visible:ring-ring w-full rounded-lg px-3 py-2.5 text-left outline-none focus-visible:ring-2 {active
            ? 'bg-muted font-bold'
            : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'}"
        >
          {c.label}
        </button>
      {/each}
    </nav>

    <div class="flex min-h-0 min-w-0 flex-col gap-4">
      <div class="relative shrink-0">
        <Search
          class="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2"
        />
        <Input
          type="search"
          placeholder="Search settings..."
          aria-label="Search settings"
          bind:value={query}
          class="h-11 pl-9"
        />
      </div>
      <div class="flex min-h-0 flex-1 flex-col gap-6 overflow-y-auto pr-1">
        {#each entries as [title, items] (title)}
          <section aria-labelledby="section-{title}" class="shrink-0">
            <h2
              id="section-{title}"
              class="text-muted-foreground mb-2 px-1 font-serif text-xs font-bold tracking-wider uppercase"
            >
              {title}
            </h2>
            <div class="divide-border bg-card divide-y rounded-xl border">
              {#each items as setting (setting.key)}
                <SettingRow
                  {setting}
                  value={settings.current(setting.key)}
                  onchange={(v) => settings.update(setting, v)}
                  onaction={(id) => {}}
                />
              {/each}
            </div>
          </section>
        {:else}
          <div
            class="text-muted-foreground flex flex-1 flex-col items-center justify-center gap-3 rounded-xl border border-dashed p-8 text-center"
          >
            {#if query}
              <SearchX class="size-8 opacity-60" />
              <p class="text-sm">No settings match “{query.trim()}”.</p>
            {:else}
              <Inbox class="size-8 opacity-60" />
              <p class="text-sm">No settings in this section yet.</p>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>