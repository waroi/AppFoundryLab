<script lang="ts">
import { getCopy } from "@/lib/ui/copy";
import { DEFAULT_LOCALE, locale, type Locale } from "@/lib/ui/preferences";

export let initialLocale: Locale = DEFAULT_LOCALE;

$: activeLocale = typeof window === "undefined" ? initialLocale : $locale;
$: copy = getCopy(activeLocale);
</script>

<section class="card" data-testid="test-page-intro">
  <h1 class="text-2xl font-semibold">{copy.testPage.title}</h1>
  <p class="mt-2 text-body">{copy.testPage.description}</p>
</section>
