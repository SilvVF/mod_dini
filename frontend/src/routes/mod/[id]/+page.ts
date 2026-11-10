import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from './$types';


export const load: PageLoad = async ({ params }) => {
  const { submission, submissionFiles } = await ModService.Page("gamebanana", params.id);

  return {
    submission: submission,
    files: submissionFiles ?? [],
  };
};
