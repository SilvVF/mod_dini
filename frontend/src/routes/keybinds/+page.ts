import { ModService } from '../../lib/bindings/changeme';
import { type ModKeyBindPreset } from '../../lib/bindings/changeme/models';
import type { PageLoad } from './$types';


export const load: PageLoad = async () => {
  const data = await ModService.GetKeyBinds()

  const binds = (data ?? {}) as Record<string, Record<string, ModKeyBindPreset>>;

  return {
    keybinds: binds
  };
};
