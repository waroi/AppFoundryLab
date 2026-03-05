# Operator Observability Runbook

This runbook covers the operator-only Prometheus access path for the single-host deployment model.

## 1. Access modes

- `basic-auth`: use [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml) and a password hash
- `mtls`: use [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml) and client certificates

The base Prometheus port should remain host-local on `127.0.0.1:9090`. Operator access exists to avoid publishing the raw port.

## 2. Generate mTLS material

```bash
./scripts/generate-operator-mtls-certs.sh
```

Default output:

- `deploy/observability/operator-certs/ca.crt`
- `deploy/observability/operator-certs/server.crt`
- `deploy/observability/operator-certs/server.key`
- `deploy/observability/operator-certs/client.crt`
- `deploy/observability/operator-certs/client.key`
- `deploy/observability/operator-certs/client-ca.crt`
- `deploy/observability/operator-certs/manifest.env`

## 3. Required environment variables

Set the following in `.env.single-host` when using mTLS:

- `ENABLE_OPERATOR_PROMETHEUS_ACCESS=true`
- `PROMETHEUS_OPERATOR_ACCESS_MODE=mtls`
- `PROMETHEUS_OPERATOR_TLS_CERT_FILE`
- `PROMETHEUS_OPERATOR_TLS_KEY_FILE`
- `PROMETHEUS_OPERATOR_CLIENT_CA_FILE`

## 4. Readiness check

```bash
./scripts/check-operator-mtls-readiness.sh ./deploy/observability/operator-certs ./.env.single-host
```

The check verifies:

- expected files exist
- certificates are not close to expiry
- the env file points at the mTLS access mode and cert paths

## 5. Rotation guidance

- rotate the CA and issued client/server certificates together on a fixed cadence
- store the generated material in your secret manager or operator-only secure storage, not in git
- run the readiness check before rollout and after replacement
- review operator connectivity after rotation with a concrete scrape or browser test through the proxy

## 6. Related files

- [scripts/generate-operator-mtls-certs.sh](/mnt/d/w/AppFoundryLab/scripts/generate-operator-mtls-certs.sh)
- [scripts/check-operator-mtls-readiness.sh](/mnt/d/w/AppFoundryLab/scripts/check-operator-mtls-readiness.sh)
- [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml)
- [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml)
- [deploy/caddy/Caddyfile.prometheus-operator.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.example)
- [deploy/caddy/Caddyfile.prometheus-operator.mtls.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.mtls.example)
