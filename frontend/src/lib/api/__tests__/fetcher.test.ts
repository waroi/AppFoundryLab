import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

function acceptAnyPayload(payload: unknown): payload is Record<string, never> {
  return payload !== undefined;
}

describe("fetchTyped", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses the backend error code when the API returns an error envelope", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: { code: "invalid_credentials" } }),
      }),
    );

    const { fetchTyped } = await import("@/lib/api/fetcher");

    await expect(fetchTyped("/api/v1/auth/token", acceptAnyPayload)).rejects.toEqual(
      expect.objectContaining({
        status: 401,
        message: "invalid_credentials",
      }),
    );
  });

  it("falls back to a status-based code when the response body is not JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        text: async () => "service unavailable",
      }),
    );

    const { fetchTyped } = await import("@/lib/api/fetcher");

    await expect(fetchTyped("/health", acceptAnyPayload)).rejects.toEqual(
      expect.objectContaining({
        status: 503,
        message: "request_failed_503",
      }),
    );
  });
});
