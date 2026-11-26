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

  let { class: className }: { class?: string } = $props();

  const pathOf = (u: URL) => u.pathname + u.search + u.hash;

  // ---- history tracking (the History API can't tell us if "forward" exists) ----
  let stack = $state<string[]>([]);
  let index = $state(-1);

  afterNavigate((nav) => {
    const navToUrl = nav.to?.url
    const url = navToUrl ? pathOf(navToUrl) : page.url.pathname + page.url.search + page.url.hash;

    if (nav.type === "popstate" && nav.delta !== undefined && stack.length) {
      index = Math.min(Math.max(index + nav.delta, 0), stack.length - 1);
    } else if (nav.type === "enter" || index < 0) {
      stack = [url];
      index = 0;
    } else if (stack[index] !== url) {
      stack = [...stack.slice(0, index + 1), url];
      index = stack.length - 1;
    }
  });

  const canBack = $derived(index > 0);
  const canForward = $derived(index >= 0 && index < stack.length - 1);

  // ---- address bar ----
  let inputEl = $state<HTMLInputElement | null>(null);
  let editing = $state(false);
  let draft = $state("");

  const current = $derived(page.url.pathname + page.url.search + page.url.hash);

  function submit() {
    let target = draft.trim();
    if (!target) return;
    if (!target.startsWith("/")) target = "/" + target;
    inputEl?.blur();
    goto(target);
  }

  function openNewWindow() {
    ModService.OpenWindow(current);
  }

  function handleWindowKeydown(e: KeyboardEvent) {
    const mod = e.ctrlKey || e.metaKey;
    if (mod && e.key.toLowerCase() === "l") {
      e.preventDefault();
      inputEl?.focus();
    } else if (mod && e.key.toLowerCase() === "n") {
      e.preventDefault();
      openNewWindow();
    }
  }

  // Ctrl/Cmd+click or middle-click on an internal link opens it in a new window
  function handleWindowClick(e: MouseEvent) {
    const wantsNewWindow =
      e.button === 1 || (e.button === 0 && (e.ctrlKey || e.metaKey));
    if (!wantsNewWindow) return;
    const a = (e.target as HTMLElement).closest("a");
    const href = a?.getAttribute("href");
    if (!a || !href?.startsWith("/")) return;
    e.preventDefault();
    e.stopPropagation();
    ModService.OpenWindow(href);
  }
</script>

<svelte:window
  onkeydown={handleWindowKeydown}
  onclickcapture={handleWindowClick}
  onauxclickcapture={handleWindowClick}
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
    value={editing ? draft : current}
    class="bg-muted/60 text-muted-foreground focus:text-foreground h-8 min-w-0 flex-1 rounded-md border px-3 font-mono text-sm outline-none focus-visible:ring-1 focus-visible:ring-ring"
    spellcheck="false"
    oninput={(e) => (draft = e.currentTarget.value)}
    onfocus={(e) => {
      draft = current;
      editing = true;
      e.currentTarget.select();
    }}
    onblur={() => (editing = false)}
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
