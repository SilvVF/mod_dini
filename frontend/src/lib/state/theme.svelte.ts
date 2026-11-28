import { Events } from "@wailsio/runtime";
import {
  parseTweakcnCss,
  type ThemeVars,
} from "#lib/components/dini/theme-parse.ts";

export const BUILT_IN_THEMES = [
  { id: "default", label: "Default" },
  { id: "slate", label: "Slate" },
] as const;

export type ThemeId = string;

export interface CustomTheme {
  id: string;
  label: string;
  light: ThemeVars;
  dark: ThemeVars;
}

interface ThemeSync {
  current: ThemeId;
  custom: CustomTheme[];
}

const THEME_KEY = "theme";
const CUSTOM_KEY = "customThemes";
const STYLE_ID = "custom-themes";
const THEME_EVENT = "theme:changed";

function readCustom(): CustomTheme[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(CUSTOM_KEY) ?? "[]");
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

const toRule = (selector: string, vars: ThemeVars) =>
  `${selector} {\n${Object.entries(vars)
    .map(([k, v]) => `  --${k}: ${v};`)
    .join("\n")}\n}`;

class ThemeState {
  custom = $state<CustomTheme[]>(readCustom());
  current = $state<ThemeId>(localStorage.getItem(THEME_KEY) ?? "default");

  all = $derived([
    ...BUILT_IN_THEMES,
    ...this.custom.map(({ id, label }) => ({ id, label })),
  ]);

  constructor() {
    this.renderStyles();
    this.apply();

    Events.On(THEME_EVENT, (event) => {
      this.receive(event.data as ThemeSync);
    });
  }

  set(id: ThemeId) {
    this.current = id;
    localStorage.setItem(THEME_KEY, id);
    this.apply(this.current);
    this.broadcast();
  }

  // Parses pasted tweakcn CSS, saves it under `name` and switches to it.
  // Throws an Error with a readable message if the CSS can't be used.
  add(name: string, css: string): CustomTheme {
    const label = name.trim();
    if (!label) throw new Error("Give the theme a name.");

    const { light, dark } = parseTweakcnCss(css);
    const slug = label
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "");
    const theme: CustomTheme = {
      id: `custom-${slug || "theme"}-${Date.now().toString(36)}`,
      label,
      light,
      dark,
    };

    this.custom = [...this.custom, theme];
    this.persistCustom();
    this.renderStyles();
    this.set(theme.id); // set() also broadcasts the updated custom list
    return theme;
  }

  remove(id: ThemeId) {
    this.custom = this.custom.filter((t) => t.id !== id);
    this.persistCustom();
    this.renderStyles();
    if (this.current === id) this.set("default");
    else this.broadcast();
  }

  // The default theme is the plain :root / .dark variables, so it has no attribute.
  // Pass an id to show a theme without saving it; omit it to show the saved one.
  apply(id: ThemeId = this.current) {
    console.trace("apply()", JSON.stringify(id));
    const root = document.documentElement;
    if (!id || id === "default") root.removeAttribute("data-theme");
    else root.setAttribute("data-theme", id);
  }

  preview(id: ThemeId) {
    this.apply(id);
  }

  clearPreview() {
    this.apply();
  }

  private broadcast() {
    const payload: ThemeSync = {
      current: this.current,
      // $state.snapshot returns a plain object that can be serialized over IPC
      custom: $state.snapshot(this.custom) as CustomTheme[],
    };
    void Events.Emit(THEME_EVENT, payload);
  }

  // Applies state pushed from another window. The sender has already written
  // localStorage (shared per origin), so only update in-memory state and the DOM.
  private receive({ current, custom }: ThemeSync) {
    this.custom = custom;
    this.renderStyles();
    const known =
      BUILT_IN_THEMES.some((t) => t.id === current) ||
      custom.some((t) => t.id === current);
    this.current = known ? current : "default";
    this.apply();
  }

  private persistCustom() {
    localStorage.setItem(CUSTOM_KEY, JSON.stringify(this.custom));
  }

  // Writes one rule pair per custom theme, mirroring the slate blocks in app.css.
  private renderStyles() {
    let el = document.getElementById(STYLE_ID) as HTMLStyleElement | null;
    if (!el) {
      el = document.createElement("style");
      el.id = STYLE_ID;
      document.head.appendChild(el);
    }
    el.textContent = this.custom
      .map((t) =>
        [
          toRule(`[data-theme="${t.id}"]`, t.light),
          toRule(`[data-theme="${t.id}"].dark`, t.dark),
        ].join("\n"),
      )
      .join("\n");
  }
}

export const appTheme = new ThemeState();
