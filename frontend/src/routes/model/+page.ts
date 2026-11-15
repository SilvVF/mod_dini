import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from "./$types";

export const load: PageLoad = async () => {
  const mesh = await ModService.GetParsedData().catch((e) => console.log(e));

  return {
    mesh,
  };
};
