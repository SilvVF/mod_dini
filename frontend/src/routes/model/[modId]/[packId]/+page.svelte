<script>
  import ModViewerModel from "#lib/components/viewer/ModViewerModel.svelte";
  import { page } from "$app/state";
</script>


<ModViewerModel 
mod={page.data.preview.mod} 
pack={page.data.preview.pack}
payload={page.data.preview.payload}
onClose={() => window.navigation.back()}/>