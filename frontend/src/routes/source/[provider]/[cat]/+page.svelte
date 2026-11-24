<script lang="ts">
  import { page } from "$app/state";
  import {
    createInfiniteQuery,
    infiniteQueryOptions,
  } from "@tanstack/svelte-query";
  import { createVirtualizer } from "@tanstack/svelte-virtual";

  import { ModService } from "#lib/bindings/changeme/index.ts";
  import SubmissionPreview from "#lib/components/dini/SubmissionPreview.svelte";
  import { Button } from "#lib/components/ui/button/index.ts";
  import * as Select from "#lib/components/ui/select/index.ts";
  import { Input } from "#lib/components/ui/input/index.ts";
  import Switch from "#lib/components/ui/switch/switch.svelte";
  import { cn } from "#lib/utils.ts";
  import { Search, RefreshCw, SlidersHorizontal } from "@lucide/svelte";
  import { untrack } from "svelte";

  const ROW_HEIGHT = 380;
  const GAP_X = 16;
  const MIN_CARD_WIDTH = 280;
  const OVERSCAN = 3;
  const PREFETCH_ROWS = 4;

  const provider = $derived(page.params.provider);
  const category = $derived(page.params.cat);

  let searchInput = $state("");
  let search = $state("");
  let subcategory = $state("all");
  let sortBy = $state("MostLiked");
  let showNSFW = $state(true);
  let showObsolete = $state(true);

  const subcategories = $derived(page.data.subcategories);

  let container = $state<HTMLDivElement>();
  let containerWidth = $state(0);

  const columns = $derived(
    Math.max(
      1,
      Math.floor((containerWidth + GAP_X) / (MIN_CARD_WIDTH + GAP_X)),
    ),
  );

  const query = createInfiniteQuery(() =>
    infiniteQueryOptions({
      queryKey: ["source", provider, category, sortBy, search, subcategory],
      initialPageParam: Number(page.url.searchParams.get("page") ?? 1),
      queryFn: ({ pageParam }) =>
        ModService.List({
          provider: provider!,
          cat: subcategory === "all" ? category! : subcategory,
          query: search,
          sort: sortBy,
          page: pageParam,
        }),
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.next : undefined,
      staleTime: 60_000,
      refetchOnWindowFocus: false,
    }),
  );

  const filteredSubmissions = $derived(
    query.data?.pages?.map((p) => p.submissions ?? [])?.flat() ?? [],
  );

  const rowCount = $derived(Math.ceil(filteredSubmissions.length / columns));

  // TODO
  const heros: string[] = [];

  // Virtualize *rows*; each row renders `columns` cards. Row height is fixed,
  // so no per-item measuring is needed.
  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => container ?? null,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });

  const virtualRows = $derived($virtualizer.getVirtualItems());
  const totalSize = $derived($virtualizer.getTotalSize());

  function rowItems(rowIndex: number) {
    const start = rowIndex * columns;
    return filteredSubmissions.slice(start, start + columns);
  }

  function refresh() {
    query.refetch();
  }

  function fetchNextPage() {
    if (!query.hasNextPage || query.isFetching) return;
    query.fetchNextPage();
  }

  // Debounce search input.
  $effect(() => {
    const value = searchInput;
    const timeout = setTimeout(() => (search = value), 300);
    return () => clearTimeout(timeout);
  });

  // Keep the virtualizer in sync with the scroll element + row count, and
  // re-measure when the column count changes.
  $effect(() => {
    const count = rowCount;
    const el = container;
    void columns;

    untrack(() => {
      $virtualizer.setOptions({
        ...$virtualizer.options,
        count,
        getScrollElement: () => el ?? null,
      });
      $virtualizer.measure();
    });
  });

  // Jump back to the top when the result set is reshaped by the controls.
  $effect(() => {
    void search;
    void subcategory;
    void sortBy;
    void showNSFW;
    void showObsolete;

    untrack(() => container?.scrollTo({ top: 0 }));
  });

  // Sorting is not a filter; these are.
  const filtersActive = $derived(
    search.trim() !== "" || subcategory !== "all" || !showNSFW || !showObsolete,
  );

  // Infinite scroll: prefetch when the last rendered row nears the end,
  // but only when no filter is active. With a filter on, pages load manually
  // via the "Load next page" button.
  $effect(() => {
    if (filtersActive) return;

    const last = virtualRows.at(-1);
    if (!last) return;

    if (last.index >= rowCount - PREFETCH_ROWS) {
      untrack(fetchNextPage);
    }
  });
</script>

<svelte:head>
  <title>Browse submissions</title>
</svelte:head>

<div class="flex h-[calc(90dvh-4rem)] flex-col gap-4 overflow-hidden m-4">
  <div class="shrink-0 space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="min-w-0">
        <h1 class="text-xl font-semibold tracking-tight">Browse submissions</h1>
        <p class="text-sm text-muted-foreground">
          {filteredSubmissions.length.toLocaleString()} submissions
        </p>
      </div>

      <Button
        variant="outline"
        size="sm"
        onclick={refresh}
        disabled={query.isFetching}
      >
        <RefreshCw
          class={cn("mr-2 size-4", query.isFetching && "animate-spin")}
        />
        Refresh
      </Button>
    </div>

    <div class="flex flex-col gap-3 rounded-lg border bg-card p-3">
      <div class="flex flex-col gap-3 md:flex-row">
        <div class="relative min-w-0 flex-1">
          <Search
            class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
          />
          <Input
            bind:value={searchInput}
            placeholder="Search submissions..."
            class="pl-9"
          />
        </div>

        <Select.Root type="single" bind:value={subcategory}>
          <Select.Trigger class="w-full md:w-50">
            <Select.Value placeholder="All characters" />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="all">All characters</Select.Item>
            {#each subcategories as subcat}
              <Select.Item value={subcat.category}>
                <img
                  src={subcat.iconUrl}
                  alt=""
                  class="size-5 rounded-full object-cover"
                />
                {subcat.name}
              </Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>

        <Select.Root type="single" bind:value={sortBy}>
          <Select.Trigger class="w-full md:w-50">
            <Select.Value placeholder="Sort by" />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="NewAndUpdated">Recently updated</Select.Item>
            <Select.Item value="Alphabetically">Name</Select.Item>
            <Select.Item value="MostViewed">Most viewed</Select.Item>
            <Select.Item value="MostLiked">Most liked</Select.Item>
          </Select.Content>
        </Select.Root>
      </div>

      <div class="flex flex-wrap items-center gap-x-5 gap-y-3">
        <div class="flex items-center gap-2">
          <Switch bind:checked={showNSFW} />
          <span class="text-sm text-muted-foreground">Show NSFW</span>
        </div>

        <div class="flex items-center gap-2">
          <Switch bind:checked={showObsolete} />
          <span class="text-sm text-muted-foreground">Show obsolete</span>
        </div>

        <div
          class="ml-auto flex items-center gap-2 text-sm text-muted-foreground"
        >
          <SlidersHorizontal class="size-4" />
          <span>{columns} {columns === 1 ? "column" : "columns"}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- The only scroll container on the page. -->
  <div
    bind:this={container}
    bind:clientWidth={containerWidth}
    class="min-h-0 flex-1 overflow-y-auto overflow-x-hidden"
  >
    {#if query.isPending}
      <div class="grid h-full place-items-center text-sm text-muted-foreground">
        Loading submissions...
      </div>
    {:else if query.isError}
      <div class="grid h-full place-content-center gap-3 text-center">
        <p class="text-sm text-destructive">Failed to load submissions.</p>
        <Button variant="outline" size="sm" onclick={refresh}>Try again</Button>
      </div>
    {:else if filteredSubmissions.length === 0}
      <div class="grid h-full place-items-center text-sm text-muted-foreground">
        {#if query.hasNextPage}
          No matches in the loaded pages yet.
        {:else}
          No submissions found.
        {/if}
      </div>
    {:else}
      <div class="relative w-full" style:height="{totalSize}px">
        {#each virtualRows as row (row.key)}
          <div
            class="absolute left-0 top-0 grid w-full gap-x-4 pb-4"
            style:height="{ROW_HEIGHT}px"
            style:transform="translateY({row.start}px)"
            style:grid-template-columns="repeat({columns}, minmax(0, 1fr))"
          >
            {#each rowItems(row.index) as submission (submission.submissionId)}
              <SubmissionPreview class="max-h-340px" {submission} />
            {/each}
          </div>
        {/each}
      </div>
    {/if}

    <div class="mt-auto">
      <!-- Status bar lives outside the scroller so it is always visible. -->
      {#if !query.isPending && !query.isError}
        <div
          class="flex shrink-0 justify-center pb-1 text-sm text-muted-foreground"
        >
          {#if query.isFetchingNextPage}
            Loading more submissions...
          {:else if query.hasNextPage}
            <Button
              variant="outline"
              size="sm"
              onclick={fetchNextPage}
              disabled={query.isFetching}
            >
              Load next page
            </Button>
          {:else}
            You've reached the end.
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>
