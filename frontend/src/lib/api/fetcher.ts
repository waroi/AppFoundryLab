function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

function readRuntimeApiBaseUrl(): string {
  if (typeof window === "undefined") {
    return "";
  }

  const configured = window.__APP_CONFIG__?.apiBaseUrl?.trim();
  if (configured) {
    return trimTrailingSlash(configured);
  }

  return "";
}

const runtimeDefaultBase =
  typeof window !== "undefined"
    ? `http://${window.location.hostname}:8080`
    : "http://localhost:8080";

export const API_BASE_URL = trimTrailingSlash(
  readRuntimeApiBaseUrl() || import.meta.env.PUBLIC_API_BASE_URL || runtimeDefaultBase,
);

const RETRY_MAX_ATTEMPTS = Number(import.meta.env.PUBLIC_API_RETRY_MAX_ATTEMPTS ?? 2);
const RETRY_BASE_DELAY_MS = Number(import.meta.env.PUBLIC_API_RETRY_BASE_DELAY_MS ?? 200);
const CIRCUIT_FAILURE_THRESHOLD = Number(import.meta.env.PUBLIC_API_CIRCUIT_FAILURE_THRESHOLD ?? 5);
const CIRCUIT_COOLDOWN_MS = Number(import.meta.env.PUBLIC_API_CIRCUIT_COOLDOWN_MS ?? 15000);

const RETRYABLE_STATUS = new Set([408, 425, 429, 500, 502, 503, 504]);

type CircuitState = {
  failureCount: number;
  openedUntil: number;
};

type ErrorEnvelope = {
  error?: {
    code?: unknown;
    message?: unknown;
  };
};

const circuitStates = new Map<string, CircuitState>();

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function canRetry(method: string, attempt: number, status: number): boolean {
  if (attempt >= RETRY_MAX_ATTEMPTS) {
    return false;
  }
  if (!(method === "GET" || method === "HEAD" || method === "OPTIONS")) {
    return false;
  }
  return RETRYABLE_STATUS.has(status);
}

function circuitKeyFor(path: string, method: string): string {
  const normalizedPath = path.split("?")[0];
  return `${method}:${normalizedPath}`;
}

function getCircuitState(key: string): CircuitState {
  const existing = circuitStates.get(key);
  if (existing) {
    return existing;
  }

  const created: CircuitState = { failureCount: 0, openedUntil: 0 };
  circuitStates.set(key, created);
  return created;
}

function checkCircuit(key: string): void {
  const state = getCircuitState(key);
  if (Date.now() < state.openedUntil) {
    throw new ApiError(503, "api_circuit_open");
  }
}

function onRequestSuccess(key: string): void {
  const state = getCircuitState(key);
  state.failureCount = 0;
  state.openedUntil = 0;
}

function onRequestFailure(key: string): void {
  const state = getCircuitState(key);
  state.failureCount += 1;
  if (state.failureCount >= CIRCUIT_FAILURE_THRESHOLD) {
    state.openedUntil = Date.now() + CIRCUIT_COOLDOWN_MS;
  }
}

async function errorCodeFromResponse(response: Response): Promise<string> {
  const fallback = `request_failed_${response.status}`;
  const text = await response.text();
  if (!text) {
    return fallback;
  }

  try {
    const payload = JSON.parse(text) as ErrorEnvelope;
    if (typeof payload.error?.code === "string" && payload.error.code.length > 0) {
      return payload.error.code;
    }
    if (typeof payload.error?.message === "string" && payload.error.message.length > 0) {
      return payload.error.message;
    }
  } catch {
    return fallback;
  }

  return fallback;
}

export async function fetchTyped<T>(
  path: string,
  validate: (payload: unknown) => payload is T,
  init?: RequestInit,
  token?: string,
): Promise<T> {
  const method = (init?.method ?? "GET").toUpperCase();
  const circuitKey = circuitKeyFor(path, method);
  checkCircuit(circuitKey);

  for (let attempt = 0; ; attempt += 1) {
    try {
      const response = await fetch(`${API_BASE_URL}${path}`, {
        ...init,
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          ...(init?.headers ?? {}),
        },
      });

      if (!response.ok) {
        if (canRetry(method, attempt, response.status)) {
          const backoff = RETRY_BASE_DELAY_MS * (attempt + 1);
          await sleep(backoff);
          continue;
        }
        throw new ApiError(response.status, await errorCodeFromResponse(response));
      }

      const payload: unknown = await response.json();
      if (!validate(payload)) {
        throw new ApiError(500, "invalid response contract");
      }

      onRequestSuccess(circuitKey);
      return payload;
    } catch (error) {
      const status = error instanceof ApiError ? error.status : 0;
      const retryableNetworkError = status === 0 && canRetry(method, attempt, 503);
      if (retryableNetworkError) {
        const backoff = RETRY_BASE_DELAY_MS * (attempt + 1);
        await sleep(backoff);
        continue;
      }

      onRequestFailure(circuitKey);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(503, "network_error");
    }
  }
}
