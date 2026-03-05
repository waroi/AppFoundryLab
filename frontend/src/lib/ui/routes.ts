import { DEFAULT_LOCALE, type Locale, normalizeLocale } from "@/lib/ui/preferences";
import type { PageTitleKey } from "@/lib/ui/copy";

export const NON_DEFAULT_LOCALES: Locale[] = ["tr"];

const ROUTE_SEGMENTS: Record<PageTitleKey, "" | "test"> = {
  home: "",
  test: "test",
};

export function getLocalizedPath(page: PageTitleKey, locale: Locale = DEFAULT_LOCALE): string {
  const normalizedLocale = normalizeLocale(locale);
  const prefix = normalizedLocale === DEFAULT_LOCALE ? "" : `/${normalizedLocale}`;
  const segment = ROUTE_SEGMENTS[page];
  return segment ? `${prefix}/${segment}` : prefix || "/";
}
