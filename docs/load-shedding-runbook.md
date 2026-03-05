# Load Shedding Runbook

## Purpose

Explain how gateway overload is handled and what to inspect first.

## Main controls

- `MAX_INFLIGHT_REQUESTS`
- `LOAD_SHED_EXEMPT_PREFIXES`
- metric: `api_gateway_load_shed_total`

## Basic response

- HTTP `503`
- `Retry-After: 1`

## Review rule

Any release that changes overload behavior should trigger `load-shed-policy-review`.
