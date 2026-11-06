import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from './$types';


export const load: PageLoad = async ({ params }) => {
  const [submission, files] = await ModService.Page(Number(params.id));

  return {
    submission: submission,
    files: files ?? [],
  };
};
