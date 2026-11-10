import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from './$types';


export const load: PageLoad = async ({ params }) => {
  const { submission, submissionFiles } = await ModService.Page(params.provider, params.id);

  console.log(submission, submissionFiles)

  return {
    submission: submission,
    files: submissionFiles ?? [],
  };
};
