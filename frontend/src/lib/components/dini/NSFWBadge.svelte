<script lang="ts">
  import { cn } from "#lib/utils.ts";
  import Badge from "../ui/badge/badge.svelte";




  let { class: className = "" }: { class?: string } = $props();
</script>

<Badge
  variant="outline"
  class={cn(
    "border-red-500/30 bg-red-950/80 font-bold uppercase tracking-widest text-red-300",
    className
  )}
>
  NSFW
</Badge>