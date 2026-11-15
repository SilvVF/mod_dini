<script>
  import { page } from "$app/state";
  import ModelViewer from "../../lib/components/dini/ModelViewer.svelte";
</script>

<ModelViewer mesh={page.data.mesh} />
