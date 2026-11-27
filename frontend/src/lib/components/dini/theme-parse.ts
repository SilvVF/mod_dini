// Parses the CSS that tweakcn exports (":root { ... }" and ".dark { ... }")
// into plain variable maps. Only known theme variables are kept, and values
// are checked, because the result is injected into a <style> element.
 
const ALLOWED = new Set([
  "background", "foreground", "card", "card-foreground",
  "popover", "popover-foreground", "primary", "primary-foreground",
  "secondary", "secondary-foreground", "muted", "muted-foreground",
  "accent", "accent-foreground", "destructive", "destructive-foreground",
  "border", "input", "ring",
  "chart-1", "chart-2", "chart-3", "chart-4", "chart-5",
  "sidebar", "sidebar-foreground", "sidebar-primary",
  "sidebar-primary-foreground", "sidebar-accent",
  "sidebar-accent-foreground", "sidebar-border", "sidebar-ring",
  "radius",
]);
 
// oklch(...), hsl(...), hex, rem/px values and calc() only.
const SAFE_VALUE = /^[\w\s.,%#()/+*-]+$/;
const UNSAFE = /url\s*\(|expression|@import/i;
 
export type ThemeVars = Record<string, string>;
export interface ParsedTheme {
  light: ThemeVars;
  dark: ThemeVars;
}
 
export function parseTweakcnCss(input: string): ParsedTheme {
  const css = input.replace(/\/\*[\s\S]*?\*\//g, "");
  const light: ThemeVars = {};
  const dark: ThemeVars = {};
 
  for (const block of css.matchAll(/(:root|\.dark)\s*\{([^}]*)\}/g)) {
    const target = block[1] === ":root" ? light : dark;
    for (const decl of block[2].matchAll(/--([\w-]+)\s*:\s*([^;]+);?/g)) {
      const name = decl[1];
      const value = decl[2].trim();
      if (!ALLOWED.has(name)) continue;
      if (!SAFE_VALUE.test(value) || UNSAFE.test(value)) continue;
      target[name] = value;
    }
  }
 
  if (!light.background || !light.primary) {
    throw new Error(
      "Couldn't find a :root block with --background and --primary. Paste the full CSS from tweakcn.",
    );
  }
 
  // A theme without a .dark block uses the same colors in dark mode.
  return { light, dark: Object.keys(dark).length ? dark : { ...light } };
}
 