import { SvelteMap } from "svelte/reactivity";
import {
  ModService,
  type Action,
  type Category,
  type Setting,
  type SettingChange,
} from "../bindings/changeme";
import { Events } from "@wailsio/runtime";
import { getContext, setContext, untrack } from "svelte";

const CONTEXT_KEY = Symbol("settings");

export class SettingsState {
  #byKey = new SvelteMap<string, Setting>();

  /** Every setting in display order. */
  all = $derived([...this.#byKey.values()].sort((a, b) => a.order - b.order));

  constructor(initial: Setting[]) {
    for (const s of initial) this.#byKey.set(s.key, s);
  }

  get(key: string): Setting | undefined {
    return this.#byKey.get(key);
  }

  /** The effective value: the user's value, or the default when unset. */
  current<T>(key: string): T | undefined {
    const s = this.#byKey.get(key);
    if (!s) return undefined;
    return (s.value !== undefined ? s.value : s.default) as T;
  }

  inCategory(category: Category): Setting[] {
    return this.all.filter((s) => s.category === category);
  }

  /** Replace a setting with the updated one sent by the backend. */
  apply(setting: Setting) {
    this.#byKey.set(setting.key, setting);
  }

  replace(settings: Setting[]) {
    this.#byKey.clear();
    for (const s of settings) this.#byKey.set(s.key, s);
  }

  async action(setting: Setting, actionId: string) {
    await ModService.OnSettingAction(setting.key, actionId);
  }

  async update(setting: Setting, change: SettingChange) {
    await ModService.UpdateSetting(setting.key, change);
  }
}

/**
 * Call once from a component's script (e.g. the settings page or app root)
 * with the full list of settings. Subscribes to the Wails change event for
 * as long as that component is mounted.
 */
export function setSettingsContext(initial: () => Setting[]): SettingsState {
  const state = new SettingsState(initial());

  $effect.pre(() => {
    const incoming = initial();
    untrack(() => state.replace(incoming));
  });

  $effect(() => {
    const unsubscribe = Events.On("setting-updated", ({ data }) =>
      state.apply(data),
    );

    return () => {
      unsubscribe();
    };
  });

  return setContext(CONTEXT_KEY, state);
}

export function getSettingsContext(): SettingsState {
  return getContext<SettingsState>(CONTEXT_KEY);
}
