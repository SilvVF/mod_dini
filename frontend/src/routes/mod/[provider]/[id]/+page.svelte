<script lang="ts">
  import { formatDistanceToNow } from "date-fns";
  import type { PageProps } from "./$types";
  import { page } from "$app/state";

  import {
    InstallState,
    ModService,
    StorageState,
    type ModSaved,
    type Pack,
  } from "#lib/bindings/changeme/index.ts";
  import { statusOf, type PackageView } from "#lib/contexts/group-mods.ts";
  import { getSavedModsContext } from "#lib/contexts/saved-mods.svelte.ts";
  import { toolName, tools } from "#lib/state/app.svelte.ts";
  import { compact } from "#lib/ts-ext.ts";

  import FileSelectDialog from "#lib/components/dini/FileSelectDialog.svelte";
  import Gallery from "#lib/components/dini/Gallery.svelte";
  import { formatSize } from "#lib/components/dini/utils.ts";
  import { Avatar, AvatarImage } from "#lib/components/ui/avatar/index.ts";
  import AvatarFallback from "#lib/components/ui/avatar/avatar-fallback.svelte";
  import { Badge } from "#lib/components/ui/badge/index.ts";
  import { Button } from "#lib/components/ui/button/index.ts";
  import Progress from "#lib/components/ui/progress/progress.svelte";
  import * as Select from "#lib/components/ui/select/index.ts";
  import Spinner from "#lib/components/ui/spinner/spinner.svelte";
  import { Switch } from "#lib/components/ui/switch/index.ts";
  import { Tabs } from "#lib/components/ui/tabs/index.ts";
  import TabsContent from "#lib/components/ui/tabs/tabs-content.svelte";
  import TabsList from "#lib/components/ui/tabs/tabs-list.svelte";
  import TabsTrigger from "#lib/components/ui/tabs/tabs-trigger.svelte";

  import ArrowLeft from "@lucide/svelte/icons/arrow-left";
  import Box from "@lucide/svelte/icons/box";
  import CalendarClock from "@lucide/svelte/icons/calendar-clock";
  import CalendarPlus from "@lucide/svelte/icons/calendar-plus";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";
  import Download from "@lucide/svelte/icons/download";
  import Files from "@lucide/svelte/icons/files";
  import Hash from "@lucide/svelte/icons/hash";
  import Heart from "@lucide/svelte/icons/heart";
  import History from "@lucide/svelte/icons/history";
  import Info from "@lucide/svelte/icons/info";
  import MessageCircle from "@lucide/svelte/icons/message-circle";
  import Package from "@lucide/svelte/icons/package";
  import RotateCw from "@lucide/svelte/icons/rotate-cw";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import Wrench from "@lucide/svelte/icons/wrench";
  import SubmissionFilePreview from "#lib/components/dini/SubmissionFilePreview.svelte";
  import ModViewerModel from "#lib/components/viewer/ModViewer.svelte";
  import ModViewer from "#lib/components/viewer/ModViewer.svelte";
  import { createQuery } from "@tanstack/svelte-query";

  let { data }: PageProps = $props();
  const saved = getSavedModsContext();

  const submission = $derived(data.submission);
  const files = $derived(data.files);

  const localMod: ModSaved | undefined = $derived(
    saved.mods[`${submission.provider}:${submission.submissionId}`],
  );
  let toolId = $derived(localMod?.toolId ?? data.resolvedToolId ?? undefined);
  const needsTool = $derived(!localMod && !toolId);

  const characters = $derived(
    compact(
      toolId ? (data.rootFolder[toolId]?.characters.characters ?? {}) : {},
    ),
  );

  const packages = $derived.by<PackageView[]>(() => {
    const localPacks = localMod?.packages;
    return files.map((file) => {
      const installed = localPacks?.find((pack) => pack.fileId === file.fileId);

      return {
        file,
        pack: installed ?? {
          id: file.fileId.toString(),
          modId: `${submission.provider}:${submission.slug}`,
          fileId: file.fileId,
          name: file.fileName,
          state: InstallState.$zero,
          enabled: false,
          storage: StorageState.$zero,
          archiveFormat: "",
          archiveSize: -1,
          disabledAt: -1,
          extractedSize: -1,
          manifest: { relPath: "", size: file.size, hash: file.md5Checksum },
        },
        status: statusOf(file, installed),
        enabled: installed !== undefined, // "selected" for the dialog; real state is pack.enabled
      };
    });
  });

  const installedPacks = $derived(
    localMod?.packages?.filter((p) => p.storage !== StorageState.StorageNone) ??
      [],
  );
  const isInstalled = $derived(installedPacks.length > 0);
  const enabledCount = $derived(installedPacks.filter((p) => p.enabled).length);

  const heroFromApi = $derived(submission.hero);

  let selectedHero = $derived(heroFromApi);

  const canSelectHero = $derived(heroFromApi == null && localMod == null);

  const hero = $derived(selectedHero ?? heroFromApi ?? undefined);

  let selectedPackPreviewId: string | undefined = $derived(undefined);

  const previewQuery = createQuery(() => ({
    queryKey: ["preview-model", localMod?.id, selectedPackPreviewId],
    queryFn: async () => {
      if (!localMod || !selectedPackPreviewId) return;
      return await ModService.PreviewPack(localMod.id, selectedPackPreviewId);
    },
    enabled: localMod !== undefined && selectedPackPreviewId !== undefined,
  }));

  const IN_PROGRESS = [
    InstallState.InstallQueued,
    InstallState.InstallDownloading,
    InstallState.InstallExtracting,
  ];
  const installing = $derived(
    localMod?.packages?.some((p) => IN_PROGRESS.includes(p.state)) ?? false,
  );

  const dlProgress = $derived(
    localMod ? saved.dlProgress[localMod.id] : undefined,
  );

  let dialogOpen = $state(false);
  let tab = $state("overview");

  const tabs = [
    { value: "overview", label: "Overview", icon: Info },
    { value: "3d", label: "3D", icon: Box },
    { value: "changelog", label: "Changelog", icon: History },
    { value: "comments", label: "Comments", icon: MessageCircle },
    { value: "files", label: "Files", icon: Files },
  ];
  const comingSoon = ["changelog", "comments"];

  const isBusy = (p: Pack) =>
    p.storage === StorageState.StorageCompressing ||
    p.storage === StorageState.StorageRestoring;

  const progressPct = (p: Pack) =>
    dlProgress && dlProgress.packageId === p.id && dlProgress.total > 0
      ? (dlProgress.progress / dlProgress.total) * 100
      : 0;

  // unix seconds or milliseconds
  const toDate = (ts: number) => new Date(ts < 1e12 ? ts * 1000 : ts);
  const ago = (ts: number) =>
    ts ? formatDistanceToNow(toDate(ts), { addSuffix: true }) : "-";

  const initials = (s: string) => s.slice(0, 2).toUpperCase();

  const downloads = $derived(files.reduce((n, f) => n + f.downloadCount, 0));

  const authorUrl = $derived(
    submission.provider === "gamebanana" && submission.authorRemoteId
      ? `https://gamebanana.com/members/${submission.authorRemoteId}`
      : "",
  );

  let heroQuery = $state("");

  const filteredCharacters = $derived(
    Object.entries(characters).filter(([name]) =>
      name.toLowerCase().includes(heroQuery.trim().toLowerCase()),
    ),
  );

  function handleConfirmDownload(packs: PackageView[]) {
    if (!toolId) return;

    if (canSelectHero && selectedHero !== null) {
      submission.hero = selectedHero;
    }

    saved.add(
      submission,
      files,
      packs.map((p) => p.pack),
    );
  }

  function reinstall() {
    if (!localMod) return;
    saved.add(submission, files, installedPacks);
  }

  function setPackEnabled(pack: Pack, on: boolean) {
    if (!localMod) return;
    const ids = installedPacks
      .filter((p) => p.enabled && p.id !== pack.id)
      .map((p) => p.id);
    if (on) ids.push(pack.id);
    saved.setAsEnabledPkgs(localMod.id, ids);
  }

  function toggleAll() {
    if (localMod) saved.toggleAllPkgs(localMod.id);
  }

  function deleteMod() {
    if (localMod) saved.remove(localMod.id);
  }

  const card = "flex items-center gap-4 rounded-xl border bg-muted/30 p-4";
  const tile =
    "flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground";
  const label = "text-xs uppercase tracking-wider text-muted-foreground";
</script>

<!-- ---- snippets ---------------------------------------------------------- -->

{#snippet stat(Icon: typeof Download, name: string, value: string)}
  <div class={card}>
    <div class={tile}><Icon class="size-4" /></div>
    <div class="min-w-0">
      <div class={label}>{name}</div>
      <div class="truncate font-mono font-semibold" title={value}>{value}</div>
    </div>
  </div>
{/snippet}

{#snippet heroCard()}
  <div class={card}>
    <Avatar class="size-10">
      {#if hero && characters[hero]}
        <AvatarImage src={characters[hero].icon} alt={hero} />
      {/if}
      <AvatarFallback>{initials(hero ?? "?")}</AvatarFallback>
    </Avatar>

    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2">
        <span class={label}>Hero</span>
        {#if heroFromApi}
          <Badge variant="outline" class="h-5 px-1.5 text-[10px]">API</Badge>
        {:else if hero}
          <Badge variant="outline" class="h-5 px-1.5 text-[10px]">Manual</Badge>
        {/if}
      </div>

      {#if canSelectHero}
        <Select.Root
          type="single"
          value={selectedHero ?? ""}
          onValueChange={(value) => (selectedHero = value || null)}
          onOpenChange={(open) => {
            if (!open) heroQuery = "";
          }}
        >
          <Select.Trigger
            class="h-7 w-full border-0 bg-transparent p-0 text-left font-semibold shadow-none"
          >
            {hero ?? "Select a character"}
          </Select.Trigger>

          <Select.Content>
            <div class="sticky top-0 z-10 bg-popover p-1">
              <input
                bind:value={heroQuery}
                type="text"
                placeholder="Search characters..."
                class="h-8 w-full rounded-md border bg-background px-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring"
                onkeydown={(e) => e.stopPropagation()}
              />
            </div>

            {#each filteredCharacters as [name, character] (name)}
              <Select.Item value={name} label={name}>
                <div class="flex items-center gap-2">
                  {#if character.icon}
                    <img
                      src={character.icon}
                      alt=""
                      class="size-5 rounded-full object-cover"
                    />
                  {/if}
                  <span>{name}</span>
                </div>
              </Select.Item>
            {:else}
              <div class="px-2 py-4 text-center text-sm text-muted-foreground">
                No characters found
              </div>
            {/each}
          </Select.Content>
        </Select.Root>
      {:else}
        <div class="truncate font-semibold">
          {hero ?? "Unknown"}
        </div>
      {/if}
    </div>
  </div>
{/snippet}

{#snippet authorCard()}
  <div class={card}>
    <Avatar class="size-10">
      <AvatarFallback>{initials(submission.author)}</AvatarFallback>
    </Avatar>
    <div class="min-w-0">
      <div class={label}>Author</div>
      <div class="truncate font-semibold">{submission.author}</div>
    </div>
    {#if authorUrl}
      <a
        href={authorUrl}
        target="_blank"
        rel="noreferrer"
        class="ml-auto inline-flex shrink-0 items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        View profile
        <ChevronRight class="size-4" />
      </a>
    {/if}
  </div>
{/snippet}

{#snippet toolCard(span: string)}
  <div class="{card} {span} {needsTool ? 'border-destructive/60' : ''}">
    <div class={tile}><Wrench class="size-4" /></div>
    <div class="min-w-0 flex-1">
      <div class={label}>Install to</div>
      {#if localMod || data.resolvedToolId}
        <div class="truncate font-mono font-semibold">
          {localMod
            ? toolName(localMod.toolId)
            : toolName(data.resolvedToolId ?? "")}
        </div>
      {:else}
        <Select.Root
          type="single"
          value={toolId ?? ""}
          onValueChange={(v) => (toolId = v)}
        >
          <Select.Trigger
            class="h-7 w-full border-0 bg-transparent p-0 font-mono font-semibold shadow-none"
          >
            {toolId ? toolName(toolId) : "Select a tool"}
          </Select.Trigger>
          <Select.Content>
            {#each tools as t (t.id)}
              <Select.Item value={t.id} label={toolName(t.id)} />
            {/each}
          </Select.Content>
        </Select.Root>
      {/if}
    </div>
  </div>
{/snippet}

{#snippet fileRow(pkg: PackageView)}
  {@const working =
    pkg.status === "queued" ||
    pkg.status === "downloading" ||
    pkg.status === "extracting"}
  <div class="flex items-center gap-4 px-6 py-4">
    <div class="min-w-0 flex-1 space-y-0.5">
      <div class="break-all font-semibold">{pkg.file.fileName}</div>
      {#if pkg.file.description}
        <div class="text-sm text-muted-foreground">{pkg.file.description}</div>
      {/if}
      <div class="flex gap-3 text-sm text-muted-foreground">
        <span class="font-medium">{formatSize(pkg.file.size)}</span>
        <span>Updated {ago(pkg.file.addedAt)}</span>
        <span>{pkg.file.downloadCount.toLocaleString()} downloads</span>
      </div>
    </div>

    {#if pkg.pack.storage === StorageState.StorageCold}
      <Badge variant="outline">Compressed</Badge>
    {/if}

    {#if working}
      <div class="w-40 space-y-1">
        <Progress value={progressPct(pkg.pack)} />
        <div class="text-xs capitalize text-muted-foreground">
          {pkg.status}…
        </div>
      </div>
    {:else if pkg.status === "available"}
      <Button
        size="sm"
        variant="outline"
        class="gap-2"
        disabled={!toolId}
        onclick={() => (dialogOpen = true)}
      >
        <Download class="size-4" />
        Install
      </Button>
    {:else if pkg.status === "error"}
      <Badge variant="destructive">Failed</Badge>
      <Button
        size="sm"
        variant="outline"
        disabled={!toolId}
        onclick={() => (dialogOpen = true)}
      >
        Retry
      </Button>
    {:else}
      {#if pkg.status === "update"}
        <Badge variant="outline">Update available</Badge>
      {/if}
      <Switch
        checked={pkg.pack.enabled}
        disabled={isBusy(pkg.pack)}
        onCheckedChange={(on) => setPackEnabled(pkg.pack, on)}
      />
    {/if}
  </div>
{/snippet}

<!-- ---- page -------------------------------------------------------------- -->

<FileSelectDialog
  bind:open={dialogOpen}
  packageView={packages}
  onConfirm={handleConfirmDownload}
  onCancel={() => (dialogOpen = false)}
/>

<div
  class="mx-auto flex h-full min-h-0 w-full flex-col gap-6 overflow-y-auto p-6"
>
  <Button
    variant="ghost"
    size="sm"
    class="w-fit gap-2"
    onclick={() => history.back()}
  >
    <ArrowLeft class="size-4" />
    Back to Mods
  </Button>

  <section class="rounded-2xl border bg-card">
    <!-- banner -->
    <div class="relative h-80 overflow-hidden rounded-t-2xl bg-muted">
      {#if submission.thumbnailUrl}
        <img
          src={submission.thumbnailUrl}
          alt=""
          class="absolute inset-0 size-full object-cover"
        />
      {/if}
      <div
        class="absolute inset-0 bg-linear-to-t from-background/90 via-background/30 to-transparent"
      ></div>

      <Badge
        variant="outline"
        class="absolute left-6 top-6 gap-1 rounded-full bg-background/60 px-3 py-1.5 backdrop-blur"
      >
        <Hash class="size-3" />
        {submission.submissionId}
      </Badge>

      <div
        class="absolute inset-x-6 bottom-6 flex items-end justify-between gap-4"
      >
        <div class="min-w-0">
          <h1 class="text-4xl font-bold tracking-tight">{submission.name}</h1>
          <p class="mt-1 text-lg text-muted-foreground">
            {submission.category}
          </p>
        </div>

        <Button
          variant="outline"
          class="shrink-0 gap-2 rounded-full bg-background/60 backdrop-blur"
          onclick={() => (tab = "3d")}
        >
          <Box class="size-4" />
          View in 3D
        </Button>
      </div>
    </div>

    <div class="flex flex-col gap-4 p-8">
      <!-- hero / author / downloads / likes -->
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {@render heroCard()}
        {@render authorCard()}
        {@render stat(Download, "Downloads", downloads.toLocaleString())}
        {@render stat(Heart, "Likes", submission.likes.toLocaleString())}
      </div>

      <hr />

      <!-- dates / install target / installed files -->
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {@render stat(CalendarPlus, "Published", ago(submission.remoteAddedAt))}
        {@render stat(
          CalendarClock,
          "Last modified",
          ago(submission.remoteUpdatedAt),
        )}
        {@render toolCard(isInstalled ? "" : "lg:col-span-2")}
        {#if isInstalled}
          {@render stat(
            Package,
            "Installed files:",
            `${enabledCount} of ${files.length} files enabled`,
          )}
        {/if}
      </div>
    </div>

    <!-- footer actions -->
    <div
      class="flex flex-wrap items-center justify-between gap-3 border-t px-8 py-5"
    >
      <a
        href={submission.profileUrl}
        target="_blank"
        rel="noreferrer"
        class="font-semibold text-primary hover:underline"
      >
        View Post on {submission.provider}
      </a>

      <div class="flex flex-wrap items-center gap-3">
        {#if isInstalled}
          <Button variant="outline" class="gap-2" onclick={reinstall}>
            <RotateCw class="size-4" />
            Reinstall
          </Button>

          <label
            class="flex cursor-pointer items-center gap-3 rounded-md border bg-background px-3 py-2 text-sm font-medium"
          >
            <Switch checked={enabledCount > 0} onCheckedChange={toggleAll} />
            Downloaded
          </label>

          <Button variant="destructive" class="gap-2" onclick={deleteMod}>
            <Trash2 class="size-4" />
            Delete Mod
          </Button>
        {:else if installing}
          <Button disabled size="sm">
            <Spinner data-icon="inline-start" />
            Installing...
          </Button>
        {:else}
          <Button disabled={!toolId} onclick={() => (dialogOpen = true)}>
            Download Mod
          </Button>
        {/if}
      </div>
    </div>
  </section>

  <Tabs bind:value={tab}>
    <TabsList>
      {#each tabs as t (t.value)}
        {@const Icon = t.icon}
        <TabsTrigger value={t.value} class="gap-2">
          <Icon class="size-4" />
          {t.label}
          {#if t.value === "files"}
            <span class="text-muted-foreground">{files.length}</span>
          {/if}
        </TabsTrigger>
      {/each}
    </TabsList>

    <TabsContent value="overview" class="mt-4 flex flex-col gap-4">
      <section class="rounded-2xl border bg-card p-8">
        <h2 class="mb-4 text-xl font-semibold">Description</h2>
        {@html submission.description}
      </section>

      {#if submission.images && submission.images.length > 0}
        <Gallery
          images={submission.images}
          nsfwSettings={{ disableBlur: false, blurStrength: 16 }}
        />
      {/if}
    </TabsContent>

    <TabsContent value="files" class="mt-4">
      <section class="divide-y overflow-hidden rounded-2xl border bg-card">
        {#each packages as pkg (pkg.pack.id)}
          {@render fileRow(pkg)}
        {/each}
      </section>
    </TabsContent>

    <TabsContent value="3d" class="mt-4">
      <section class="rounded-2xl border bg-card p-8 text-muted-foreground">
        Install to view in 3d.
      </section>
      <Select.Root
        type="single"
        value={localMod?.packages?.at(0)?.id ?? ""}
        onValueChange={(v) => (selectedPackPreviewId = v)}
      >
        <Select.Trigger
          class="h-7 w-full border-0 bg-transparent p-0 font-mono font-semibold shadow-none"
        >
          {selectedPackPreviewId ? selectedPackPreviewId : "Select a pack"}
        </Select.Trigger>
        <Select.Content>
          {#each localMod.packages as p (p.id)}
            <Select.Item value={p.id} label={p.name} />
          {/each}
        </Select.Content>
      </Select.Root>
      {#if selectedPackPreviewId && previewQuery.data && previewQuery.data.payload}
        <div class="h-800px">
        <ModViewer
          mod={previewQuery.data.mod}
          pack={previewQuery.data.pack}
          payload={previewQuery.data.payload}
          embedded={true}
          onClose={() => {}}
        />
        </div>
      {/if}
    </TabsContent>

    {#each comingSoon as v (v)}
      <TabsContent value={v} class="mt-4">
        <section class="rounded-2xl border bg-card p-8 text-muted-foreground">
          Not available yet.
        </section>
      </TabsContent>
    {/each}
  </Tabs>
</div>
