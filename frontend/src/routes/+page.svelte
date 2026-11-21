<script lang="ts">
  import { page } from "$app/state";
  import {
    createInfiniteQuery,
    infiniteQueryOptions,
  } from "@tanstack/svelte-query";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { untrack } from "svelte";

  import { ModService } from "#lib/bindings/changeme/index.ts";
  import SubmissionPreview from "#lib/components/dini/SubmissionPreview.svelte";

  let searchQuery = $state(page.url.searchParams.get("search") ?? "");
  let debouncedSearch = $state(searchQuery);

  let filters = $state({
    categories: [] as string[],
    heroes: [] as string[],
    sort: "default",
    hideNsfw: false,
    favoritesOnly: false,
  });

  const PAGE_SIZE = 50;
  const ESTIMATED_ROW_HEIGHT = 340;
  const LOAD_MORE_THRESHOLD = 4;

  let parentElement: HTMLDivElement | undefined = $state(undefined);

  let columns = $state(4);

  let resizeObserver: ResizeObserver | undefined;

  let resizeElement = $state<HTMLElement | null>(null);

  $effect(() => {
    if (!resizeElement) return;

    resizeObserver?.disconnect();

    resizeObserver = new ResizeObserver(([entry]) => {
      const width = entry.contentRect.width;

      columns =
        width >= 1536
          ? 6
          : width >= 1280
            ? 5
            : width >= 1024
              ? 4
              : width >= 768
                ? 3
                : width >= 640
                  ? 2
                  : 1;
    });

    resizeObserver.observe(resizeElement);

    return () => resizeObserver?.disconnect();
  });

  let debounceTimeout: ReturnType<typeof setTimeout>;

  $effect(() => {
    const query = searchQuery;

    clearTimeout(debounceTimeout);

    debounceTimeout = setTimeout(() => {
      debouncedSearch = query;
    }, 300);

    return () => clearTimeout(debounceTimeout);
  });

  const queryFilters = $derived({
    search: debouncedSearch.trim(),
    categories: filters.categories,
    heroes: filters.heroes,
    sort: filters.sort,
    hideNsfw: filters.hideNsfw,
    favoritesOnly: filters.favoritesOnly,
  });

  const query = createInfiniteQuery(() =>
    infiniteQueryOptions({
      queryKey: ["source", page.params.provider, page.params.cat],
      initialPageParam: Number(page.url.searchParams.get("page") ?? 1),
      queryFn: ({ pageParam }) =>
        ModService.List({
          provider: page.params.provider!,
          cat: page.params.cat!,
          page: pageParam,
        }),
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.next : undefined,
    }),
  );

  const submissions = $derived(
    query.data?.pages.flatMap((p) => p.submissions) ?? [],
  );

  const rows = $derived.by(() => {
    const result: (typeof submissions)[] = [];

    for (let i = 0; i < submissions.length; i += columns) {
      result.push(submissions.slice(i, i + columns));
    }

    return result;
  });

  const virtualizer = createVirtualizer({
    get count() {
      return rows.length;
    },
    getScrollElement: () => parentElement!,
    estimateSize: () => ESTIMATED_ROW_HEIGHT,
    overscan: 3,
  });

  const virtualRows = $derived($virtualizer.getVirtualItems());

  const lastVirtualRowIndex = $derived(virtualRows.at(-1)?.index ?? -1);

  $effect(() => {
    if (
      !query.hasNextPage ||
      query.isFetchingNextPage ||
      query.isFetchNextPageError ||
      lastVirtualRowIndex < rows.length - LOAD_MORE_THRESHOLD
    ) {
      return;
    }

    query.fetchNextPage();
  });

  $effect(() => {
    untrack(() => {
      if (parentElement) {
        parentElement.scrollTo({
          top: 0,
          behavior: "instant",
        });
      }
    });
  });

  const updateSearch = (value: string) => {
    searchQuery = value;
  };

  const updateFilters = (next: Partial<typeof filters>) => {
    filters = { ...filters, ...next };
  };
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
  <div class="flex items-center gap-3">
    <input
      class="h-10 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm"
      type="search"
      placeholder="Search submissions..."
      bind:value={searchQuery}
    />

    <select
      class="h-10 rounded-md border bg-background px-3 text-sm"
      value={filters.sort}
      onchange={(event) => updateFilters({ sort: event.currentTarget.value })}
    >
      <option value="default">Default</option>
      <option value="lastUpdated">Last updated</option>
      <option value="downloadCount">Downloads</option>
      <option value="rating">Rating</option>
      <option value="releaseDate">Release date</option>
    </select>
  </div>

  <div class="flex min-h-0 flex-1 flex-col" aria-busy={query.isFetching}>
    {#if query.isPending}
      <div
        class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
      >
        {#each Array(8) as _, i (`skeleton-${i}`)}
          <div class="h-80 animate-pulse rounded-xl bg-muted"></div>
        {/each}
      </div>
    {:else if query.isError}
      <div
        class="flex flex-1 items-center justify-center text-sm text-destructive"
      >
        {query.error.message}
      </div>
    {:else if submissions.length === 0}
      <div
        class="flex flex-1 flex-col items-center justify-center gap-2 text-muted-foreground"
      >
        <p class="text-lg font-semibold">No submissions found</p>
        <p class="text-sm">
          {searchQuery ||
          Object.values(filters).some((v) =>
            Array.isArray(v) ? v.length > 0 : v !== false && v !== "default",
          )
            ? "Try changing your search or filters."
            : "There are no submissions available."}
        </p>
      </div>
    {:else}
      <div
        bind:this={parentElement}
        bind:this={resizeElement}
        class="min-h-0 flex-1 overflow-auto"
      >
        <div
          class="relative w-full"
          style="height: {$virtualizer.getTotalSize()}px;"
        >
          {#each virtualRows as virtualRow (virtualRow.key)}
            <div
              class="absolute left-0 top-0 w-full"
              style="
                height: {virtualRow.size}px;
                transform: translateY({virtualRow.start}px);
                contain: strict;
              "
            >
              <div
                class="grid gap-4 px-1 pr-2"
                style="grid-template-columns: repeat({columns}, minmax(0, 1fr));"
              >
                {#each rows[virtualRow.index] ?? [] as submission (`${submission?.provider}:${submission?.submissionId}`)}
                  {#if submission}
                    <SubmissionPreview {submission} />
                  {/if}
                {/each}
              </div>
            </div>
          {/each}
        </div>

        {#if query.isFetchingNextPage}
          <div class="flex justify-center py-6">
            <span class="text-sm text-muted-foreground">
              Loading more submissions...
            </span>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
