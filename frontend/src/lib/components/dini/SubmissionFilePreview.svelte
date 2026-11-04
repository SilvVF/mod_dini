<script lang="ts">
  import type { SubmissionFile } from "#lib/bindings/changeme/models.ts";
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "../ui/card";
  import DateDisplay from "./DateDisplay.svelte";
  import { epochSecondsToDate, formatByteSize } from "./utils";

  type Props = {
    files: SubmissionFile[];
    isDownloadable?: boolean;
  };

  let { files, isDownloadable = false }: Props = $props();
</script>

{#if isDownloadable && files && files.length > 0}
  <Card class="shadow-none contain-[layout_style_paint]">
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        {"Files"}
      </CardTitle>

      <CardDescription>
        {"Files that can be installed. Different versions of the mod are available."}
        {#if files.length > 1}
          {"Multiple versions available"}
        {/if}
      </CardDescription>
    </CardHeader>

    <CardContent>
      <ul class="list-disc space-y-2">
        {#each files as file (file.fileId)}
          <li class="flex flex-col gap-0.5 text-sm">
            <div class="flex items-center justify-between">
              <div class="flex min-w-0 flex-1 items-center gap-2">
                <span class="truncate" title={file.fileName}>
                  {file.fileName}

                  <DateDisplay
                    date={epochSecondsToDate(file.addedAt)}
                    prefix="added"
                    class="text-xs text-muted-foreground"
                  />
                </span>
              </div>

              <span class="ml-2 text-xs text-muted-foreground">
                {formatByteSize(file.size)}
                {#if file.md5Checksum}
                  (MD5: {file.md5Checksum})
                {/if}
              </span>
            </div>

            {#if file.description}
              <p class="line-clamp-2 text-xs text-muted-foreground">
                {file.description}
              </p>
            {/if}
          </li>
        {/each}
      </ul>
    </CardContent>
  </Card>
{/if}
