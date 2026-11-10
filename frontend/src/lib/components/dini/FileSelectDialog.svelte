<script lang="ts">
    import { untrack } from "svelte";
    import { SvelteSet } from "svelte/reactivity";
    import * as Dialog from "../ui/dialog";
    import { Button } from "../ui/button";
    import { Checkbox } from "../ui/checkbox";
    import Download from "@lucide/svelte/icons/download";
    import HardDrive from "@lucide/svelte/icons/hard-drive";
    import { formatSize } from "./utils";
    import type { PackageView } from "#lib/contexts/group-mods.ts";

    interface Props {
        packageView: PackageView[];
        onConfirm: (packs: PackageView[]) => void;
        onCancel: () => void;
        modName?: string;
    }

    let { packageView, onConfirm, onCancel, modName = "Mod" }: Props = $props();

    let open = $state(false);
    const selectedIds = new SvelteSet<string>();

    // Rows mid-install can't be toggled
    const isLocked = (p: PackageView) =>
        p.status === "downloading" || p.status === "extracting";

    // Reset the selection from the stored state each time the dialog opens
    $effect.pre(() => {
        if (!open) return;
        untrack(() => {
            selectedIds.clear();
            for (const p of packageView) {
                if (p.enabled) selectedIds.add(p.pack.id);
            }
        });
    });

    const toggleable = $derived(packageView.filter((p) => !isLocked(p)));
    const selected = $derived(
        packageView.filter((p) => selectedIds.has(p.pack.id)),
    );
    const totalSize = $derived(
        selected.reduce((sum, p) => sum + p.file.size, 0),
    );

    const allSelected = $derived(
        toggleable.length > 0 &&
            toggleable.every((p) => selectedIds.has(p.pack.id)),
    );
    const someSelected = $derived(selected.length > 0 && !allSelected);

    function toggle(id: string, checked: boolean) {
        if (checked) selectedIds.add(id);
        else selectedIds.delete(id);
    }

    function toggleAll(checked: boolean) {
        for (const p of toggleable) toggle(p.pack.id, checked);
    }

    function confirm() {
        onConfirm(selected);
        open = false;
    }

    function cancel() {
        onCancel();
        open = false;
    }

    const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });
    const units: [Intl.RelativeTimeFormatUnit, number][] = [
        ["year", 31536000],
        ["month", 2592000],
        ["day", 86400],
        ["hour", 3600],
        ["minute", 60],
    ];

    // addedAt may be unix seconds or milliseconds
    function timeAgo(ts: number) {
        const ms = ts < 1e12 ? ts * 1000 : ts;
        const diff = (ms - Date.now()) / 1000;
        for (const [unit, secs] of units) {
            if (Math.abs(diff) >= secs)
                return rtf.format(Math.round(diff / secs), unit);
        }
        return "just now";
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Trigger>
        {#snippet child({ props })}
            <Button {...props}>DL</Button>
        {/snippet}
    </Dialog.Trigger>

    <Dialog.Content class="max-w-xl gap-6 p-8">
        {#if packageView.length > 0}
            <Dialog.Header class="gap-0 text-left">
                <div
                    class="mb-6 flex size-12 items-center justify-center rounded-full bg-muted"
                >
                    <Download class="size-5" />
                </div>

                <Dialog.Title class="text-xl font-semibold">
                    Select Files to Download
                </Dialog.Title>

                <Dialog.Description class="mt-1">
                    {modName} contains {packageView.length}
                    {packageView.length === 1 ? "file" : "files"}. Select which
                    ones to install.
                </Dialog.Description>
            </Dialog.Header>

            <div class="overflow-hidden rounded-lg border bg-muted/30">
                <label
                    class="flex cursor-pointer items-center gap-3 border-b bg-muted/40 px-4 py-3"
                >
                    <Checkbox
                        checked={allSelected}
                        indeterminate={someSelected}
                        disabled={toggleable.length === 0}
                        onCheckedChange={(checked) =>
                            toggleAll(checked === true)}
                    />
                    <span
                        class="flex-1 text-xs font-medium uppercase tracking-wider text-muted-foreground"
                    >
                        Select all
                    </span>
                    <span class="text-xs font-medium text-muted-foreground">
                        {selected.length}/{packageView.length}
                    </span>
                </label>

                <div class="max-h-80 divide-y overflow-y-auto">
                    {#each packageView as p (p.pack.id)}
                        {@const locked = isLocked(p)}
                        {@const checked = selectedIds.has(p.pack.id)}
                        <label
                            class="flex items-start gap-3 px-4 py-3 transition-colors {checked
                                ? 'bg-muted/60'
                                : 'hover:bg-muted/30'} {locked
                                ? 'cursor-not-allowed opacity-60'
                                : 'cursor-pointer'}"
                        >
                            <Checkbox
                                class="mt-1"
                                {checked}
                                disabled={locked}
                                onCheckedChange={(c) =>
                                    toggle(p.pack.id, c === true)}
                            />

                            <div class="min-w-0 flex-1 space-y-0.5">
                                <div class="break-all font-semibold">
                                    {p.file.fileName}
                                </div>

                                {#if p.file.description}
                                    <div class="text-sm text-muted-foreground">
                                        {p.file.description}
                                    </div>
                                {/if}

                                <div
                                    class="flex gap-3 text-sm text-muted-foreground"
                                >
                                    <span class="font-medium"
                                        >{formatSize(p.file.size)}</span
                                    >
                                    <span
                                        >Updated {timeAgo(p.file.addedAt)}</span
                                    >
                                </div>
                            </div>
                        </label>
                    {/each}
                </div>
            </div>

            <Dialog.Footer
                class="flex-row items-center justify-between sm:justify-between"
            >
                <div
                    class="flex items-center gap-2 text-sm text-muted-foreground"
                >
                    <HardDrive class="size-4" />
                    Total: {formatSize(totalSize)}
                </div>

                <div class="flex items-center gap-2">
                    <Button variant="outline" onclick={cancel}>Cancel</Button>
                    <Button onclick={confirm}>
                        <Download class="size-4" />
                        Download Selected
                    </Button>
                </div>
            </Dialog.Footer>
        {:else}
            <Dialog.Header>
                <Dialog.Title>Loading</Dialog.Title>
                <Dialog.Description>Analyzing mod files...</Dialog.Description>
            </Dialog.Header>
        {/if}
    </Dialog.Content>
</Dialog.Root>
