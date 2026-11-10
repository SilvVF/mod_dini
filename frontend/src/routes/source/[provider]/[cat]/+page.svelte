<script lang="ts">
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import SubmissionPreview from "#lib/components/dini/SubmissionPreview.svelte";
  import { page } from "$app/state";
  import {
    createInfiniteQuery,
    infiniteQueryOptions,
  } from "@tanstack/svelte-query";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { untrack } from "svelte";

  let search = $state("");
  let container = $state<HTMLDivElement>();
  let sentinel = $state<HTMLDivElement>();

  const pageSearchParam = page.url.searchParams.get("page");

  const query = createInfiniteQuery(() =>
    infiniteQueryOptions({
      queryKey: ["source", page.params.provider, page.params.cat] as const,
      queryFn: ({ pageParam }) =>
        ModService.List({
          provider: page.params.provider!,
          cat: page.params.cat!,
          page: pageParam,
        }),
      initialPageParam: pageSearchParam ? Number(pageSearchParam) : 1,
      getNextPageParam: (lastPage) =>
        lastPage.hasNext ? lastPage.next : undefined,
    }),
  );

  const submissions = $derived(
    query.data?.pages.flatMap((page) => page.submissions) ?? [],
  );

  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => container!,
    estimateSize: () => 160,
    overscan: 5,
    getItemKey: (index) => submissions[index]?.submissionId ?? "",
  });

  $effect(() => {
    submissions.length;

    untrack(() => {
      $virtualizer.setOptions({
        ...$virtualizer.options,
        count: submissions.length,
      });
    });
  });

  const virtualItems = $derived($virtualizer.getVirtualItems());

  $effect(() => {
    if (!sentinel || !container) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (
          entry.isIntersecting &&
          query.hasNextPage &&
          !query.isFetchingNextPage
        ) {
          query.fetchNextPage();
        }
      },
      {
        root: container,
        rootMargin: "0px 0px 300px 0px",
      },
    );

    observer.observe(sentinel);

    return () => observer.disconnect();
  });
</script>

<div class="space-y-4">
  <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
    <h2 class="text-xl font-semibold text-foreground">Submissions</h2>

    <input
      bind:value={search}
      type="search"
      placeholder="Search submissions..."
      class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring sm:w-64"
    />
  </div>

  {#if query.isPending}
    <p class="text-foreground">Loading...</p>
  {:else if query.error}
    <p class="text-destructive">Error: {query.error.message}</p>
  {:else}
    <div bind:this={container} class="h-[75vh] overflow-y-auto">
      <div
        class="relative w-full"
        style:height={`${$virtualizer.getTotalSize()}px`}
      >
        {#each virtualItems as item (item.key)}
          <div
            class="absolute left-0 top-0 w-full"
            style:transform={`translateY(${item.start}px)`}
          >
            <SubmissionPreview submission={submissions[item.index]!} />
          </div>
        {/each}
      </div>

      <div bind:this={sentinel} class="h-1"></div>

      {#if query.isFetchingNextPage}
        <p class="py-4 text-center text-sm text-muted-foreground">
          Loading more submissions...
        </p>
      {:else if !query.hasNextPage}
        <p class="py-4 text-center text-sm text-muted-foreground">
          No more submissions.
        </p>
      {/if}
    </div>
  {/if}
</div>
