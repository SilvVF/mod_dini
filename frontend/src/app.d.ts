// See https://svelte.dev/docs/kit/types#app.d.ts

import { Category } from "./lib/bindings/changeme";

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
    interface PageState {
      settings?: { category: Category };
    }
  }
}

export {};
