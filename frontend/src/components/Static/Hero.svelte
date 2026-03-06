<script lang="ts">
import { getCopy } from "@/lib/ui/copy";
import { DEFAULT_LOCALE, type Locale, locale } from "@/lib/ui/preferences";

export let initialLocale: Locale = DEFAULT_LOCALE;

$: activeLocale = typeof window === "undefined" ? initialLocale : $locale;
$: copy = getCopy(activeLocale);
</script>

<section class="card" data-testid="home-hero">
  <p class="eyebrow-text">{copy.hero.eyebrow}</p>
  <h1 class="mt-3 text-3xl font-bold">{copy.hero.title}</h1>
  <p class="mt-3 text-body">{copy.hero.description}</p>
  <p class="mt-3 text-sm text-soft">{copy.hero.supportingText}</p>
</section>
