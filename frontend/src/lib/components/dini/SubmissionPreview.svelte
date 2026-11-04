<script lang="ts">
  import type { Submission } from "../../bindings/changeme";
  import DateDisplay from "./DateDisplay.svelte";
  import { epochSecondsToDate } from "./utils";

  let { submission }: { submission: Submission } = $props();

  function formatCount(value: number): string {
    return new Intl.NumberFormat("en", {
      notation: "compact",
      maximumFractionDigits: 1,
    }).format(value);
  }

  function getTags(tags: string): string[] {
    try {
      return JSON.parse(tags || "[]");
    } catch {
      return [];
    }
  }
</script>

<a
  href={`/mod/${submission.submissionId}`}
  class="flex gap-3 rounded-lg border bg-card p-3 transition-colors hover:bg-accent"
>
  <div class="h-20 w-28 shrink-0 overflow-hidden rounded-md bg-muted">
    {#if submission.thumbnailUrl}
      <img
        src={submission.thumbnailUrl}
        alt=""
        loading="lazy"
        class="h-full w-full object-cover"
      />
    {:else}
      <div
        class="flex h-full items-center justify-center text-xs text-muted-foreground"
      >
        No image
      </div>
    {/if}
  </div>

  <div class="flex min-w-0 flex-1 flex-col justify-between gap-2">
    <div class="space-y-1">
      <div class="flex flex-wrap items-center gap-2">
        <h3 class="truncate font-semibold">{submission.name}</h3>

        {#if submission.isNsfw}
          <span
            class="rounded bg-red-500/10 px-1.5 py-0.5 text-xs text-red-500"
          >
            NSFW
          </span>
        {/if}

        {#if submission.isObsolete}
          <span
            class="rounded bg-yellow-500/10 px-1.5 py-0.5 text-xs text-yellow-600"
          >
            Obsolete
          </span>
        {/if}

        {#if submission.isTombstoned}
          <span
            class="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground"
          >
            Removed
          </span>
        {/if}
      </div>

      <p class="line-clamp-2 text-sm text-muted-foreground">
        {submission.hero || "No description available."}
      </p>

      <div class="line-clamp-2 text-sm text-muted-foreground">
        {@html submission.description || "No description available."}
      </div>
    </div>

    <div
      class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground"
    >
      <span>{submission.author}</span>
      <span>{submission.category}</span>
      <DateDisplay
        date={epochSecondsToDate(submission.remoteUpdatedAt)}
        prefix="added"
        class="text-xs text-muted-foreground"
      />
    </div>

    <div class="flex flex-wrap gap-1">
      {#each getTags(submission.tags).slice(0, 5) as tag}
        <span
          class="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground"
        >
          {tag}
        </span>
      {/each}
    </div>
  </div>

  <div class="hidden shrink-0 flex-col justify-center gap-2 text-right sm:flex">
    <div>
      <p class="text-sm font-semibold">{formatCount(submission.viewCount)}</p>
      <p class="text-xs text-muted-foreground">Views</p>
    </div>

    <div>
      <p class="text-sm font-semibold">{formatCount(submission.likes)}</p>
      <p class="text-xs text-muted-foreground">Likes</p>
    </div>
  </div>
</a>
