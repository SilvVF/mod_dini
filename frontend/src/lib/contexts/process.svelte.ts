import { Events } from "@wailsio/runtime";
import { getContext, setContext, untrack } from "svelte";
import { SvelteMap } from "svelte/reactivity";
import { EventType, ModService, type ProcessEvent } from "../bindings/changeme";

const CONTEXT_KEY = Symbol("process");

export class ProcessState {
  #byKey = new SvelteMap<string, ProcessEvent>();

  all = $derived([...this.#byKey.values()]);

  running = (key: string) =>
    this.#byKey.get(key)?.type === EventType.ProcessStarted;

  set(key: string, value: ProcessEvent | undefined) {
    if (value) {
      this.#byKey.set(key, value);
    } else {
      this.#byKey.delete(key);
    }
  }

  get(key: string): ProcessEvent | undefined {
    return this.#byKey.get(key);
  }
}

export function setProcessContext(): ProcessState {
  const state = new ProcessState();

  $effect.pre(() => {
    ModService.RunningProcesses().then((data) => {
      if (!data) return;
      for (const [key, value] of Object.entries(data)) {
        state.set(key, value);
      }
    });
  });

  $effect(() => {
    const unsubscribe = Events.On("process-event", ({ data }) => {
      state.set(data.name, data);
    });

    return () => {
      unsubscribe();
    };
  });

  return setContext(CONTEXT_KEY, state);
}

export function getProcessContext(): ProcessState {
  return getContext<ProcessState>(CONTEXT_KEY);
}
