import type { PageLoad } from "./$types";
import { queryClient, submissionPageQuery } from "#lib/components/dini/utils.ts";

export const load: PageLoad = async ({ params }) => {

  const data = await queryClient.query(
    submissionPageQuery(params.provider, params.id),
  );

  return {
    resolvedToolId: data.resolvedToolId,
    submission: data.submission,
    files: data.submissionFiles ?? [],
  };
};
