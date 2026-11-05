<script lang="ts">
  import { Category } from "#lib/bindings/changeme/models.ts";
  import SettingsView from "#lib/components/settings/SettingsView.svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";

  const categoryValues = new Set<string>(Object.values(Category));

  function categoryFromString(value: string | undefined): Category {
    return categoryValues.has(value ?? "") ? value as Category : Category.$zero;
  }

  const category = $derived(categoryFromString(page.params.cat))
  function handleNavigate(category: Category) {
    goto(`/settings/${category}`)
  } 
</script>

<SettingsView category={category} onnavigate={handleNavigate} />
