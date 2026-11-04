import { Events } from "@wailsio/runtime";
import { getContext, onMount, setContext } from "svelte";
import { ModService } from "../bindings/changeme";
import { page } from "$app/state";
import { goto } from "$app/navigation";

export type FolderItem = {
  title: string;
  path: string;
};

export interface Settings {
  activeSection: string;
  folders: Record<string, FolderItem>;
  query: string;
  setFolder(id: string): void;
  isOpen: boolean;
  open: () => void;
  close: () => void;
}

export function setSettings() {
  let context = $state<Settings>({
    isOpen: false,
    activeSection: "",
    query: "",
    folders: {
      xxmi: {
        title: "XXMI Launcher",
        path: "",
      },
      gimi: {
        title: "Genshin Impact Mods folder",
        path: "",
      },
      srmi: {
        title: "Honkai: Star Rail Mods folder",
        path: "",
      },
      zzmi: {
        title: "Zenless Zone Zero Mods folder",
        path: "",
      },
      wwmi: {
        title: "Wuthering Waves Mods folder",
        path: "",
      },
    },
    setFolder(id) {
      ModService.SetModFolder(id);
    },
    close() {
      if (page.state.modal) {
        goto(page.state.background ?? "/", {
          state: {
            modal: undefined,
            background: undefined,
          },
        });
      }
    },
    open() {
      goto("/settings", {
        state: {
          modal: page.url.pathname !== "/settings",
          background: page.url.pathname,
        },
        shallow: true,
      });
    },
  });

  setContext(SETTINGS_KEY, context);

  $effect(() => {
    context.isOpen = page.state.modal === true;
  });

  onMount(() => {
    ModService.Settings().then((data) => {
      if (!data) return;
      for (const [id, folder] of Object.entries(data)) {
        if (context.folders[id]) {
          context.folders[id].path = folder?.path ?? "";
        }
      }
    });

    const unsubscribe = Events.On("folder-updated", ({ data }) => {
      if (!data) return;
      context.folders[data.id].path = data.path;
    });

    return () => unsubscribe();
  });

  return context;
}

const SETTINGS_KEY = Symbol("settings");

export function getSettings() {
  return getContext<Settings>(SETTINGS_KEY);
}
