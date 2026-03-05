<script lang="ts">
import { getCopy, getPageTitle, type PageTitleKey } from "@/lib/ui/copy";
import { getLocalizedPath } from "@/lib/ui/routes";
import { DEFAULT_LOCALE, initializePreferenceSync, locale, setLocale, setTheme, theme, type Locale } from "@/lib/ui/preferences";
import { onMount } from "svelte";

export let pageTitleKey: PageTitleKey;
export let initialLocale: Locale = DEFAULT_LOCALE;

let localePending: Locale | null = null;

onMount(() => {
  initializePreferenceSync();
  setLocale(initialLocale);
});

$: activeLocale = typeof window === "undefined" ? initialLocale : $locale;
$: copy = getCopy(activeLocale);
$: if (typeof document !== "undefined") {
  document.title = getPageTitle(activeLocale, pageTitleKey);
}

function switchLocale(nextLocale: Locale): void {
  if (typeof window === "undefined" || nextLocale === activeLocale || localePending) {
    return;
  }

  localePending = nextLocale;
  setLocale(nextLocale);

  const url = new URL(window.location.href);
  url.pathname = getLocalizedPath(pageTitleKey, nextLocale);
  window.location.assign(url.toString());
}
</script>

<aside class="preference-toolbar" data-testid="preference-toolbar">
  <div class="preference-group" aria-label={copy.toolbar.language}>
    <span class="preference-label">{copy.toolbar.language}</span>
    <div class="segmented-control" role="group">
      <button
        class:segment-active={activeLocale === "en"}
        class="segment-button"
        type="button"
        on:click={() => switchLocale("en")}
        aria-pressed={activeLocale === "en"}
        disabled={localePending !== null}
        data-testid="locale-en"
      >
        {copy.toolbar.english}
      </button>
      <button
        class:segment-active={activeLocale === "tr"}
        class="segment-button"
        type="button"
        on:click={() => switchLocale("tr")}
        aria-pressed={activeLocale === "tr"}
        disabled={localePending !== null}
        data-testid="locale-tr"
      >
        {copy.toolbar.turkish}
      </button>
    </div>
  </div>

  <div class="preference-group" aria-label={copy.toolbar.theme}>
    <span class="preference-label">{copy.toolbar.theme}</span>
    <div class="segmented-control" role="group">
      <button
        class:segment-active={$theme === "light"}
        class="segment-button"
        type="button"
        on:click={() => setTheme("light")}
        aria-pressed={$theme === "light"}
        data-testid="theme-light"
      >
        {copy.toolbar.light}
      </button>
      <button
        class:segment-active={$theme === "dark"}
        class="segment-button"
        type="button"
        on:click={() => setTheme("dark")}
        aria-pressed={$theme === "dark"}
        data-testid="theme-dark"
      >
        {copy.toolbar.dark}
      </button>
    </div>
  </div>
</aside>
