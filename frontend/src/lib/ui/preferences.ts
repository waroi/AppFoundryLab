import { writable } from "svelte/store";

export type Locale = "en" | "tr";
export type Theme = "light" | "dark";

export const DEFAULT_LOCALE: Locale = "en";
export const DEFAULT_THEME: Theme = "light";
export const THEME_STORAGE_KEY = "appfoundrylab.theme";

export function normalizeLocale(value: unknown): Locale {
  return value === "tr" ? "tr" : DEFAULT_LOCALE;
}

export function normalizeTheme(value: unknown): Theme {
  return value === "dark" ? "dark" : DEFAULT_THEME;
}

function readInitialLocale(): Locale {
  if (typeof document === "undefined") {
    return DEFAULT_LOCALE;
  }

  return normalizeLocale(document.documentElement.lang);
}

function readInitialTheme(): Theme {
  if (typeof document === "undefined") {
    return DEFAULT_THEME;
  }

  return normalizeTheme(document.documentElement.dataset.theme);
}

export const locale = writable<Locale>(readInitialLocale());
export const theme = writable<Theme>(readInitialTheme());
let syncInitialized = false;

export function setLocale(value: Locale): void {
  locale.set(normalizeLocale(value));
}

export function setTheme(value: Theme): void {
  theme.set(normalizeTheme(value));
}

export function initializePreferenceSync(): void {
  if (typeof window === "undefined" || syncInitialized) {
    return;
  }
  syncInitialized = true;

  locale.subscribe((value) => {
    const normalized = normalizeLocale(value);
    document.documentElement.lang = normalized;
  });

  theme.subscribe((value) => {
    const normalized = normalizeTheme(value);
    document.documentElement.dataset.theme = normalized;
    document.documentElement.style.colorScheme = normalized;
    window.localStorage.setItem(THEME_STORAGE_KEY, normalized);
  });

  window.addEventListener("storage", (event) => {
    if (event.key === THEME_STORAGE_KEY && event.newValue) {
      theme.set(normalizeTheme(event.newValue));
    }
  });
}
