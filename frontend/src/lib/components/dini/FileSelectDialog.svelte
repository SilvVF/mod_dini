<script lang="ts">
  import { Badge } from "../ui/badge";
  import { Checkbox } from "../ui/checkbox";
  import * as Dialog from "../ui/dialog";
  import File from "@lucide/svelte/icons/file";
  import Archive from "@lucide/svelte/icons/archive";
  import FolderOpen from "@lucide/svelte/icons/folder-open";
  import Download from "@lucide/svelte/icons/download";
  import { formatSize } from "./utils";
  import { Button } from "../ui/button";
  import DialogTrigger from "../ui/dialog/dialog-trigger.svelte";
  import DialogClose from "../ui/dialog/dialog-close.svelte";
  import type { ModFile, ModFileTree } from "#lib/bindings/changeme/models.ts";

  interface Props {
    fileTree: ModFileTree | null;
    onConfirm: (selectedFileTree: ModFileTree) => void;
    onCancel: () => void;
    modName?: string;
  }

  let { fileTree, onConfirm, onCancel, modName = "Mod" }: Props = $props();

  let localFileTree = $state<ModFileTree | null>(null);

  $effect(() => {
    if (fileTree) {
      localFileTree = fileTree;
    } else {
      localFileTree = null;
    }
  });

  const selectedFiles = $derived(
    localFileTree?.files?.filter((file) => file.is_selected) ?? [],
  );

  const selectedCount = $derived(selectedFiles.length);

  const totalSelectedSize = $derived(
    selectedFiles.reduce((acc, file) => acc + file.size, 0),
  );

  const allSelected = $derived(
    localFileTree !== null &&
      localFileTree.files !== null &&
      localFileTree.files.length > 0 &&
      selectedCount === localFileTree.files.length,
  );

  const someSelected = $derived(
    localFileTree !== null &&
      localFileTree.files !== null &&
      selectedCount > 0 &&
      selectedCount < localFileTree.files.length,
  );

  const filesByArchive = $derived.by(() => {
    const groups = new Map<string, ModFile[]>();

    for (const file of localFileTree?.files ?? []) {
      const files = groups.get(file.archive_name) ?? [];
      files.push(file);
      groups.set(file.archive_name, files);
    }

    return Array.from(groups.entries());
  });

  const hasMultipleArchives = $derived(filesByArchive.length > 1);

  function handleFileToggle(fileIndex: number, checked: boolean) {
    if (!localFileTree) return;

    localFileTree = {
      ...localFileTree,
      files:
        localFileTree?.files?.map((file, index) =>
          index === fileIndex ? { ...file, is_selected: checked } : file,
        ) ?? [],
    };
  }

  function handleSelectAll(checked: boolean) {
    if (!localFileTree) return;

    localFileTree = {
      ...localFileTree,
      files:
        localFileTree?.files?.map((file) => ({
          ...file,
          is_selected: checked,
        })) ?? [],
    };
  }

  function handleArchiveToggle(archiveName: string, checked: boolean) {
    if (!localFileTree) return;

    localFileTree = {
      ...localFileTree,
      files:
        localFileTree?.files?.map((file) =>
          file.archive_name === archiveName
            ? { ...file, is_selected: checked }
            : file,
        ) ?? [],
    };
  }

  function getArchiveSelectedCount(files: ModFile[]): number {
    return files.filter((file) => file.is_selected).length;
  }

  function isArchiveFullySelected(files: ModFile[]): boolean {
    return files.length > 0 && files.every((file) => file.is_selected);
  }

  function isArchivePartiallySelected(files: ModFile[]): boolean {
    const selected = getArchiveSelectedCount(files);
    return selected > 0 && selected < files.length;
  }

  function confirmSelection() {
    if (!localFileTree || selectedCount === 0) return;
    onConfirm(localFileTree);
  }
</script>

<Dialog.Root>
  <DialogTrigger>
    <Button>DL</Button>
  </DialogTrigger>
  {#if localFileTree}
    <Dialog.Content
      class="flex h-[80vh]max-h-[80vh] max-w-3xl flex-col gap-0 overflow-hidden bg-background p-0"
      onclick={(event) => event.stopPropagation()}
    >
      <Dialog.Header class="shrink-0 border-b px-6 py-4">
        <Dialog.Title class="flex items-center gap-2">
          <FolderOpen class="h-5 w-5" />
          Select files to install {modName}
        </Dialog.Title>

        <Dialog.Description>
          Choose the files you want to install.
          {selectedCount} of {localFileTree.total_files} selected
          {#if selectedCount > 0}
            ({formatSize(totalSelectedSize)})
          {/if}
        </Dialog.Description>
      </Dialog.Header>

      <div
        class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden px-6 py-4 no-scrollbar"
      >
        <div class="flex shrink-0 items-center gap-2 border-b pb-2">
          <Checkbox
            checked={allSelected}
            class={someSelected ? "data-[state=checked]:bg-yellow-500" : ""}
            onCheckedChange={(checked) => handleSelectAll(checked === true)}
          />

          <span class="text-sm font-medium">
            Select all ({localFileTree.total_files})
          </span>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto pr-2">
          <div class="space-y-4">
            {#if hasMultipleArchives}
              {#each filesByArchive as [archiveName, files] (archiveName)}
                <div class="space-y-2">
                  <div class="flex items-center gap-2 border-b pb-2">
                    <Checkbox
                      checked={isArchiveFullySelected(files)}
                      class={isArchivePartiallySelected(files)
                        ? "data-[state=checked]:bg-yellow-500"
                        : ""}
                      onCheckedChange={(checked) =>
                        handleArchiveToggle(archiveName, checked === true)}
                    />

                    <Archive class="h-4 w-4 text-muted-foreground" />

                    <span class="text-sm font-medium">
                      {archiveName}
                    </span>

                    <Badge variant="outline" class="text-xs">
                      {getArchiveSelectedCount(files)}/{files.length}
                    </Badge>
                  </div>

                  <div class="ml-6 space-y-1">
                    {#each files as file}
                      <div
                        class="flex w-full items-center justify-between rounded-md px-3 py-2 text-left hover:bg-muted/50"
                        data-install-file={JSON.stringify([
                          file.archive_name,
                          file.path,
                        ])}
                      >
                        <div class="flex min-w-0 items-center gap-3">
                          <File
                            class="h-4 w-4 shrink-0 text-muted-foreground"
                          />

                          <div class="min-w-0 space-y-1">
                            <div class="break-all font-mono text-sm">
                              {file.name}
                            </div>

                            {#if file.path !== file.name}
                              <div
                                class="break-all text-xs text-muted-foreground"
                              >
                                {file.path}
                              </div>
                            {/if}
                          </div>
                        </div>

                        <span
                          class="ml-3 shrink-0 text-xs text-muted-foreground"
                        >
                          {formatSize(file.size)}
                        </span>
                      </div>
                    {/each}
                  </div>
                </div>
              {/each}
            {:else}
              <div class="space-y-2">
                {#each localFileTree.files as file, index (file.path)}
                  <div
                    class="flex w-full items-center justify-between rounded-md px-3 py-2 text-left hover:bg-muted/50"
                    data-install-file={JSON.stringify([
                      file.archive_name,
                      file.path,
                    ])}
                  >
                    <div class="flex min-w-0 items-center gap-3">
                      <File class="h-4 w-4 shrink-0 text-muted-foreground" />

                      <div class="min-w-0 space-y-1">
                        <div class="break-all font-mono text-sm">
                          {file.name}
                        </div>

                        {#if file.path !== file.name}
                          <div class="break-all text-xs text-muted-foreground">
                            {file.path}
                          </div>
                        {/if}
                      </div>
                    </div>

                    <span class="ml-3 shrink-0 text-xs text-muted-foreground">
                      {formatSize(file.size)}
                    </span>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      </div>

      <Dialog.Footer
        class="mt-auto flex shrink-0 flex-col gap-3 border-t bg-background px-6 py-4 sm:flex-row sm:items-center sm:justify-between"
      >
        <div class="text-sm text-muted-foreground">
          {#if selectedCount === 0}
            No files selected
          {:else}
            {selectedCount} files selected ({formatSize(totalSelectedSize)})
          {/if}
        </div>

        <div class="flex items-center justify-end gap-2">
          <DialogClose>
            <Button variant="outline" onclick={onCancel}>Cancel</Button>
          </DialogClose>

          <DialogClose>
            <Button disabled={selectedCount === 0} onclick={confirmSelection}>
              <Download class="mr-2 h-4 w-4" />
              Install selected
            </Button>
          </DialogClose>
        </div>
      </Dialog.Footer>
    </Dialog.Content>
  {:else}
    <Dialog.Content class="max-h-[80vh] max-w-md">
      <Dialog.Header>
        <Dialog.Title>Loading</Dialog.Title>
        <Dialog.Description>Analyzing mod files...</Dialog.Description>
      </Dialog.Header>
    </Dialog.Content>
  {/if}
</Dialog.Root>
