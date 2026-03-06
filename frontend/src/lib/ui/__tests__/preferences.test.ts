import { describe, expect, it } from "vitest";
import { DEFAULT_LOCALE, DEFAULT_THEME, normalizeLocale, normalizeTheme } from "../preferences";

describe("preferences UI library", () => {
  describe("normalizeLocale", () => {
    it("should return 'tr' when value is 'tr'", () => {
      expect(normalizeLocale("tr")).toBe("tr");
    });

    it("should return DEFAULT_LOCALE when value is 'en'", () => {
      expect(normalizeLocale("en")).toBe(DEFAULT_LOCALE);
    });

    it("should return DEFAULT_LOCALE for unknown string values", () => {
      expect(normalizeLocale("fr")).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale("")).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale("invalid")).toBe(DEFAULT_LOCALE);
    });

    it("should return DEFAULT_LOCALE for non-string values", () => {
      expect(normalizeLocale(null)).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale(undefined)).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale(123)).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale({})).toBe(DEFAULT_LOCALE);
      expect(normalizeLocale([])).toBe(DEFAULT_LOCALE);
    });
  });

  describe("normalizeTheme", () => {
    it("should return 'dark' when value is 'dark'", () => {
      expect(normalizeTheme("dark")).toBe("dark");
    });

    it("should return DEFAULT_THEME when value is 'light'", () => {
      expect(normalizeTheme("light")).toBe(DEFAULT_THEME);
    });

    it("should return DEFAULT_THEME for unknown string values", () => {
      expect(normalizeTheme("blue")).toBe(DEFAULT_THEME);
      expect(normalizeTheme("")).toBe(DEFAULT_THEME);
      expect(normalizeTheme("invalid")).toBe(DEFAULT_THEME);
    });

    it("should return DEFAULT_THEME for non-string values", () => {
      expect(normalizeTheme(null)).toBe(DEFAULT_THEME);
      expect(normalizeTheme(undefined)).toBe(DEFAULT_THEME);
      expect(normalizeTheme(123)).toBe(DEFAULT_THEME);
      expect(normalizeTheme({})).toBe(DEFAULT_THEME);
      expect(normalizeTheme([])).toBe(DEFAULT_THEME);
    });
  });
});
