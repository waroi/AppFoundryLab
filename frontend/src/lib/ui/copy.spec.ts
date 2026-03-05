import { describe, it, expect } from "bun:test";
import { translateError } from "./copy";

describe("translateError", () => {
  it("returns the English mapped error message for a known code", () => {
    expect(translateError("en", "network_error")).toBe("Network error");
    expect(translateError("en", "login_required")).toBe("Sign in before running this action.");
  });

  it("returns the Turkish mapped error message for a known code", () => {
    expect(translateError("tr", "network_error")).toBe("Ag hatasi");
    expect(translateError("tr", "login_required")).toBe("Bu islemi calistirmadan once giris yapin.");
  });

  it("returns the original error code when the code is unknown", () => {
    const unknownCode = "some_random_error_code";
    expect(translateError("en", unknownCode)).toBe(unknownCode);
    expect(translateError("tr", unknownCode)).toBe(unknownCode);
  });

  it("falls back to the default locale (en) for an unknown locale", () => {
    // @ts-expect-error - intentionally passing an invalid locale
    expect(translateError("fr", "network_error")).toBe("Network error");
  });
});
