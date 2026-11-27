import { ModService } from "#lib/bindings/changeme/index.ts";
import { queryClient } from "#lib/components/dini/utils.ts";
import type { PageLoad } from "../../../model/[modId]/[packId]/$types";

export const load: PageLoad = async ({ params }) => {
  const packPreview = await queryClient.query({
    queryKey: ["preview3d", params.modId, params.packId],
    queryFn: async () => {
      const pack = await ModService.PreviewPack(params.modId, params.packId);
      return pack;
    },
  });

  return {
    preview: packPreview,
  };
};
