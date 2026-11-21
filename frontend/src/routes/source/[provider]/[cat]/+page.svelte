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

  type Submission = NonNullable<
    Awaited<ReturnType<typeof ModService.List>>["submissions"]
  >[number];

  type Indexed = { submission: Submission; haystack: string };

  // Row height includes the bottom gap (pb-4 on each row).
  const ROW_HEIGHT = 380;
  const GAP_X = 16;
  const MIN_CARD_WIDTH = 280;
  const OVERSCAN = 3;
  const PREFETCH_ROWS = 4;

  const provider = $derived(page.params.provider);
  const category = $derived(page.params.cat);

  let searchInput = $state("");
  let search = $state("");
  let heroFilter = $state("all");
  let sortBy = $state("updated");
  let showNSFW = $state(true);
  let showObsolete = $state(true);

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
      queryKey: ["source", provider, category],
      initialPageParam: Number(page.url.searchParams.get("page") ?? 1),
      queryFn: ({ pageParam }) =>
        ModService.List({
          provider: provider!,
          cat: category!,
          page: pageParam,
        }),
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.next : undefined,
      staleTime: 60_000,
      refetchOnWindowFocus: false,
    }),
  );

  // Dedupe + precompute a lowercase search haystack once per data change,
  // instead of lowercasing four fields per item on every keystroke.
  const indexed = $derived.by<Indexed[]>(() => {
    const unique = new Map<Submission["submissionId"], Indexed>();

    for (const result of query.data?.pages ?? []) {
      for (const submission of result.submissions ?? []) {
        if (!submission || unique.has(submission.submissionId)) continue;

        unique.set(submission.submissionId, {
          submission,
          haystack: [
            submission.name,
            submission.description,
            submission.author,
            submission.hero,
          ]
            .filter(Boolean)
            .join("\n")
            .toLowerCase(),
        });
      }
    }

    return [...unique.values()];
  });

  const heroes = $derived(
    [
      ...new Set(
        indexed
          .map(({ submission }) => submission.hero?.trim())
          .filter((hero): hero is string => Boolean(hero)),
      ),
    ].sort((a, b) => a.localeCompare(b)),
  );

  const filteredSubmissions = $derived.by<Submission[]>(() => {
    const needle = search.trim().toLowerCase();

    // Single pass for all filters.
    const result: Submission[] = [];
    for (const { submission, haystack } of indexed) {
      if (needle && !haystack.includes(needle)) continue;
      if (!showNSFW && submission.isNsfw) continue;
      if (!showObsolete && submission.isObsolete) continue;
      if (heroFilter !== "all" && submission.hero?.trim() !== heroFilter)
        continue;
      result.push(submission);
    }

    switch (sortBy) {
      case "name":
        result.sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""));
        break;
      case "views":
        result.sort((a, b) => (b.viewCount ?? 0) - (a.viewCount ?? 0));
        break;
      case "likes":
        result.sort((a, b) => (b.likes ?? 0) - (a.likes ?? 0));
        break;
      case "updated":
        result.sort(
          (a, b) => (b.remoteUpdatedAt ?? 0) - (a.remoteUpdatedAt ?? 0),
        );
        break;
    }

    return result;
  });

  const rowCount = $derived(Math.ceil(filteredSubmissions.length / columns));

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
    const timeout = setTimeout(() => (search = value), 150);
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
    void heroFilter;
    void sortBy;
    void showNSFW;
    void showObsolete;

    untrack(() => container?.scrollTo({ top: 0 }));
  });

  // Sorting is not a filter; these are.
  const filtersActive = $derived(
    search.trim() !== "" || heroFilter !== "all" || !showNSFW || !showObsolete,
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

        <Select.Root type="single" bind:value={heroFilter}>
          <Select.Trigger class="w-full md:w-[200px]">
            <Select.Value placeholder="All characters" />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="all">All characters</Select.Item>
            {#each heroes as hero (hero)}
              <Select.Item value={hero}>{hero}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>

        <Select.Root type="single" bind:value={sortBy}>
          <Select.Trigger class="w-full md:w-[170px]">
            <Select.Value placeholder="Sort by" />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="updated">Recently updated</Select.Item>
            <Select.Item value="name">Name</Select.Item>
            <Select.Item value="views">Most viewed</Select.Item>
            <Select.Item value="likes">Most liked</Select.Item>
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
