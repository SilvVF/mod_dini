<script lang="ts">
  import { cn } from "cn";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "../ui/tooltip";
  import { format, formatDistanceToNow } from "date-fns";


  type Props = {
    date: Date | null | undefined;
    prefix?: string;
    inverse?: boolean;
    dateFormat?: string;
    class?: string;
  };

  let {
    date,
    prefix,
    inverse = false,
    dateFormat = "dd-MM-yyyy HH:mm",
    class: className = "",
  }: Props = $props();

  const exact = $derived(date ? format(date, dateFormat) : "");
  const prefixWithSpace = $derived(prefix ? `${prefix} ` : "");

  const distance = $derived(
    date ? `${prefixWithSpace}${formatDistanceToNow(date)} ${"ago"}` : "",
  );

  const displayText = $derived(prefix ? `${prefix} ${"at"} ${exact}` : exact);

  const tooltipText = $derived(
    prefix ? `${prefix} ${"at"} ${exact}` : `${"at"} ${exact}`,
  );
</script>

{#if date}
  <TooltipProvider>
    <Tooltip>
      <TooltipTrigger>
        <span class={cn("cursor-pointer", className)}>
          {inverse ? displayText : distance}
        </span>
      </TooltipTrigger>

      <TooltipContent>
        <span class={cn("text-xs text-muted-foreground", className)}>
          {inverse ? distance : tooltipText}
        </span>
      </TooltipContent>
    </Tooltip>
  </TooltipProvider>
{/if}