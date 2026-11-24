import { ModService } from "#lib/bindings/changeme/index.ts";
import type { PageLoad } from "../../../source/[provider]/[cat]/$types";

export const load: PageLoad = async ({params}) => {

    const subcats = await ModService.Subcategories(params.provider, params.cat)

    return {
        subcategories: subcats
    }
};