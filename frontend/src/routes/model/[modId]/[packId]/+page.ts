import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from "../../../model/[modId]/[packId]/$types";

export const load: PageLoad = async ({ params }) => {
  const packPreview = await ModService.PreviewPack(
    params.modId,
    params.packId,
  )

  
  return {
    preview: packPreview,
  };
};
