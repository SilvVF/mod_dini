import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  const { resolvedToolId, submission, submissionFiles } = await ModService.Page(
    params.provider,
    params.id,
  );

  return {
    resolvedToolId: resolvedToolId,
    submission: submission,
    files: submissionFiles ?? [],
  };
};
