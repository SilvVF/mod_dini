<script lang="ts">
  import { ChevronDown } from "@lucide/svelte";
  import { Switch } from "../ui/switch";
  import { Input } from "../ui/input";
  import {
    Kind,
    type Setting,
  } from "../../bindings/changeme";
  import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from "../ui/select";
  import { ToggleGroup, ToggleGroupItem } from "../ui/toggle-group";
  import Popover from "../ui/popover/popover.svelte";
  import PopoverTrigger from "../ui/popover/popover-trigger.svelte";
  import { PopoverContent } from "../ui/popover";
  import { Checkbox } from "../ui/checkbox";
  import { Label } from "../ui/label";
  import { SettingsMeta } from "./SettingsMeta";
  import type { SettingChange } from "#lib/contexts/settings.svelte.ts";

  let {
    setting,
    value,
    onchange,
  }: {
    setting: Setting;
    value: unknown;
    onchange: (value: SettingChange) => void;
  } = $props();

  const options = $derived(setting.options ?? []);
  const labelOf = (v: unknown) => {
    return options.find((o) => o.value === v)?.label ?? String(v ?? "");
  };

  const selected = $derived(Array.isArray(value) ? (value as string[]) : []);
  const titleId = $derived(`setting-${setting.key}`);

  function onMultiSelect(optionValue: string, on: boolean) {
    const newValue = new Set(selected);
    if (on) {
      newValue.add(optionValue);
    } else {
      newValue.delete(optionValue);
    }
    onchange({
      kind: Kind.KindMultiSelect,
      value: Array.from(newValue),
    });
  }

  function onToggle(on: boolean) {
    onchange({ kind: Kind.KindToggle, value: on });
  }

  function onDropdown(option: string) {
    onchange({ kind: Kind.KindDropdown, value: option });
  }

  function onSegmented(option: string) {
    onchange({ kind: Kind.KindSegmented, value: option });
  }

  function onTextEdit(text: string) {
    onchange({ kind: Kind.KindText, value: text });
  }

  function onEditNumber(e: Event & { currentTarget: HTMLInputElement }) {
    let n = e.currentTarget.valueAsNumber;
    if (Number.isNaN(n)) {
      e.currentTarget.value = "";
      return;
    }

    if (setting.min) n = Math.max(setting.min, n);
    if (setting.max) n = Math.min(setting.max, n);
    e.currentTarget.value = String(n);

    onchange({ kind: Kind.KindNumber, value: n });
  }

  const summary = $derived(
    selected.length === 0
      ? "None"
      : selected.length <= 2
        ? selected.map(labelOf).join(", ")
        : `${selected.length} selected`,
  );
</script>

{#if setting.kind === Kind.KindToggle}
  <Switch
    disabled={!setting.editable}
    id={titleId}
    aria-label={setting.title}
    checked={value === true}
    onCheckedChange={onToggle}
  />
{:else if setting.kind === Kind.KindDropdown}
  {@const TriggerIcon = options.find((o) => o.value === value)?.icon}
  <Select
    type="single"
    value={String(value ?? "")}
    onValueChange={onDropdown}
    disabled={!setting.editable}
  >
    <SelectTrigger id={titleId} aria-label={setting.title} class="w-44">
      <span class="flex items-center gap-2">
        {#if TriggerIcon}
          {@const Icon = SettingsMeta.resolveIcon(TriggerIcon)}
          <Icon class="size-4" />
        {/if}
        {labelOf(value)}
      </span>
    </SelectTrigger>
    <SelectContent>
      {#each options as o (o.value)}
        {@const Icon = o.icon ? SettingsMeta.resolveIcon(o.icon) : null}
        <SelectItem value={o.value} label={o.label}>
          {#if Icon}<Icon class="mr-2 size-4" />{/if}
          {o.label}
        </SelectItem>
      {/each}
    </SelectContent>
  </Select>
{:else if setting.kind === Kind.KindSegmented}
  <ToggleGroup
    disabled={!setting.editable}
    type="single"
    aria-label={setting.title}
    value={String(value ?? "")}
    onValueChange={onSegmented}
    class="bg-muted/60 rounded-lg p-1"
  >
    {#each options as o (o.value)}
      <ToggleGroupItem
        value={o.value}
        class="text-muted-foreground data-[state=on]:bg-background data-[state=on]:text-foreground rounded-md px-4 font-semibold data-[state=on]:shadow-sm"
      >
        {o.label}
      </ToggleGroupItem>
    {/each}
  </ToggleGroup>
{:else if setting.kind === Kind.KindMultiSelect}
  <Popover>
    <PopoverTrigger
      id={titleId}
      aria-label={setting.title}
      class="border-input bg-background hover:bg-accent focus-visible:ring-ring flex h-9 w-44 items-center justify-between gap-2 rounded-md border px-3 text-sm outline-none focus-visible:ring-2"
    >
      <span class="truncate">{summary}</span>
      <ChevronDown class="size-4 shrink-0 opacity-60" />
    </PopoverTrigger>
    <PopoverContent class="w-56 space-y-1 p-2" align="end">
      {#each options as o (o.value)}
        <div
          class="hover:bg-accent flex items-center gap-2 rounded-md px-2 py-1.5"
        >
          <Checkbox
            disabled={!setting.editable}
            id="{titleId}-{o.value}"
            checked={selected.includes(o.value)}
            onCheckedChange={(c) => onMultiSelect(o.value, c)}
          />
          <Label
            for="{titleId}-{o.value}"
            class="flex-1 cursor-pointer font-normal">{o.label}</Label
          >
        </div>
      {/each}
    </PopoverContent>
  </Popover>
{:else if setting.kind === Kind.KindText}
  <Input
    disabled={!setting.editable}
    id={titleId}
    aria-label={setting.title}
    class="w-56"
    value={String(value ?? "")}
    onchange={(e) => onTextEdit(e.currentTarget.value)}
  />
{:else if setting.kind === Kind.KindNumber}
  <Input
    disabled={!setting.editable}
    id={titleId}
    aria-label={setting.title}
    type="number"
    class="w-28"
    value={Number(value ?? 0)}
    min={setting.min}
    max={setting.max}
    step={setting.step}
    onchange={onEditNumber}
  />
{/if}
