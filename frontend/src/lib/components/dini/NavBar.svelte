<script lang="ts">
  import { afterNavigate, goto } from "$app/navigation";
  import { page } from "$app/state";
  import {
    ArrowLeft,
    ArrowRight,
    Home,
    SquareArrowOutUpRight,
  } from "@lucide/svelte";
  import { Button } from "#lib/components/ui/button/index.ts";
  import { ModService } from "#lib/bindings/changeme/index.ts";
  import { cn } from "#lib/utils.ts";
  import { providerRoute } from "./utils";

  let { class: className }: { class?: string } = $props();

  type PathLike = { pathname: string; search: string; hash: string };
  const pathOf = (u: PathLike) => u.pathname + u.search + u.hash;

  const current = $derived(pathOf(page.url));

  let stack = $state<string[]>([]);
  let index = $state(-1);

  afterNavigate(({ type, to }) => {
    const url = pathOf(to?.url ?? page.url);

    if (type === "popstate" && stack.length) {
      index = Math.min(Math.max(index, 0), stack.length - 1);
    } else if (type === "enter" || index < 0) {
      stack = [url];
      index = 0;
    } else if (stack[index] !== url) {
      stack = [...stack.slice(0, index + 1), url];
      index = stack.length - 1;
    }
  });

  const canBack = $derived(index > 0);
  const canForward = $derived(index < stack.length - 1);

  let draft = $derived(current);
  let inputEl = $state<HTMLInputElement | null>(null);

  const openNewWindow = () => ModService.OpenWindow(current);

  function handleKeydown(e: KeyboardEvent) {
    if (!(e.ctrlKey || e.metaKey)) return;
    const key = e.key.toLowerCase();
    if (key === "l") {
      e.preventDefault();
      inputEl?.focus();
      inputEl?.select();
    } else if (key === "n") {
      e.preventDefault();
      openNewWindow();
    }
  }

  // Ctrl/Cmd+click or middle-click on an internal link opens a new window
  function handleClick(e: MouseEvent) {
    const wantsNewWindow =
      e.button === 1 || (e.button === 0 && (e.ctrlKey || e.metaKey));

    if (!wantsNewWindow) return;

    const target = e.target;
    if (!(target instanceof HTMLElement)) return;

    const anchor = target.closest("a");
    if (!anchor) return;

    const href = anchor.getAttribute("href");
    if (!href) return;

    // Only handle internal application routes
    if (!href.startsWith("/") || href.startsWith("//")) return;

    e.preventDefault();
    e.stopPropagation();

    ModService.OpenWindow(href);
  }

  function submit() {
    const target = draft.trim();
    if (!target) return;
    inputEl?.blur();
    goto(
      providerRoute(target) ?? (target.startsWith("/") ? target : "/" + target),
    );
  }
</script>

<svelte:window
  onkeydown={handleKeydown}
  onclickcapture={handleClick}
  onauxclickcapture={handleClick}
/>

<div class={cn("flex min-w-0 items-center gap-1", className)}>
  <Button
    variant="ghost"
    size="icon"
    class="size-8 shrink-0"
    aria-label="Back"
    title="Back"
    disabled={!canBack}
    onclick={() => history.back()}
  >
    <ArrowLeft class="size-4" />
  </Button>
  <Button
    variant="ghost"
    size="icon"
    class="size-8 shrink-0"
    aria-label="Forward"
    title="Forward"
    disabled={!canForward}
    onclick={() => history.forward()}
  >
    <ArrowRight class="size-4" />
  </Button>
  <Button
    variant="ghost"
    size="icon"
    class="size-8 shrink-0"
    aria-label="Home"
    title="Home"
    href="/"
    disabled={page.url.pathname === "/"}
  >
    <Home class="size-4" />
  </Button>

  <input
    bind:this={inputEl}
    bind:value={draft}
    class="bg-muted/60 text-muted-foreground focus:text-foreground h-8 min-w-0 flex-1 rounded-md border px-3 font-mono text-sm outline-none focus-visible:ring-1 focus-visible:ring-ring"
    spellcheck="false"
    onfocus={(e) => e.currentTarget.select()}
    onblur={() => (draft = current)}
    onkeydown={(e) => {
      if (e.key === "Enter") submit();
      else if (e.key === "Escape") inputEl?.blur();
    }}
  />

  <Button
    variant="ghost"
    size="icon"
    class="size-8 shrink-0"
    aria-label="Open in new window"
    title="Open in new window (Ctrl+N)"
    onclick={openNewWindow}
  >
    <SquareArrowOutUpRight class="size-4" />
  </Button>
</div>
