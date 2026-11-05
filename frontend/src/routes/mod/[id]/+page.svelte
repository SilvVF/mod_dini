<script lang="ts">
  import {
    ModService,
    type ModFileTree,
    type Submission,
    type SubmissionFile,
  } from "#lib/bindings/changeme/index.ts";
  import FileSelectDialog from "#lib/components/dini/FileSelectDialog.svelte";
  import SubmissionFilePreview from "#lib/components/dini/SubmissionFilePreview.svelte";
  import SubmissionPreview from "#lib/components/dini/SubmissionPreview.svelte";
  import Progress from "#lib/components/ui/progress/progress.svelte";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import { page } from "$app/state";

  let submission = $state<Submission>();
  let files = $state<SubmissionFile[]>([]);

  const id = $derived(page.params.id);

  const saved = getSavedModsContext();
  const localMod = $derived(
    Object.values(saved.mods).find(
      (v) => v.submission.slug === submission?.slug,
    ),
  );

  const dlProgress = $derived(
    localMod ? saved.dlProgress[localMod.id] : undefined,
  );

  $effect(() => {
    ModService.Page(Number(id)).then(([data, sfiles]) => {
      submission = data;
      files = sfiles ?? [];
    });
  });

  const tree = $derived.by<ModFileTree>(() => {
    const f =
      files.map((f) => ({
        name: f.fileName,
        path: "",
        size: f.size,
        is_selected: false,
        archive_name: f.fileName,
      })) ?? [];

    return {
      files: localMod?.tree?.files ?? f,
      total_files: files.length,
      has_multiple_files: files.length > 1,
    };
  });

  function handleConfirmDownload(selectedFileTree: ModFileTree) {
    if (!submission) return;
    saved.add(submission, files, selectedFileTree);
  }
</script>

{#if submission}
  <div class="flex flex-col space-y-4">
    {#if localMod}
      <p>Local mod: {localMod.id}</p>
    {/if}
    <Progress
      value={dlProgress?.progress}
      max={dlProgress?.total}
      class="w-[60%]"
    />
    <SubmissionPreview {submission} />
    <SubmissionFilePreview {files} isDownloadable={true} />

    <FileSelectDialog
      fileTree={tree}
      onConfirm={handleConfirmDownload}
      onCancel={() => {}}
    />
  </div>
{/if}
