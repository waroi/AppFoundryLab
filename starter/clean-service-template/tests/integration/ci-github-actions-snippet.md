# GitHub Actions Snippet (Integration Smoke)

Asagidaki job, template tabanli HTTP servislerde minimal smoke kontrolunu CI pipeline'a eklemek icin kullanilabilir:

```yaml
integration-smoke:
  name: Integration Smoke (Template)
  runs-on: ubuntu-latest
  steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Start service
      run: |
        docker compose -f docker-compose.minimal.yml up --build -d

    - name: Wait for service
      run: |
        for i in $(seq 1 30); do
          if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
            exit 0
          fi
          sleep 2
        done
        exit 1

    - name: Run integration smoke
      run: |
        BASE_URL=http://127.0.0.1:8080 \
        HEALTH_PATH=/health \
        READY_PATH=/health/ready \
        ./tests/integration/smoke-http.sh

    - name: Stop service
      if: always()
      run: docker compose -f docker-compose.minimal.yml down -v
```

Not:
- Path/env degerlerini servis kontratina gore uyarlayin.
- Hazirlik endpoint'i degrade durumda `503` donebiliyorsa `EXPECT_READY_503=true` kullanin.
- Daha sert bir varyant test etmek isterseniz `docker compose -f docker-compose.minimal.yml -f docker-compose.security.yml up --build -d` kullanin.
- Docker disi bir pipeline icin servisi once `APP_CMD="..." ./scripts/run-local.sh` ile arka planda baslatip ayni smoke scriptini kullanabilirsiniz.
- Process-mode icin dogrudan `tests/integration/ci-github-actions-process-mode-snippet.md` dosyasindaki job'i temel alin.
- Opsiyonel load shedding middleware eklediyseniz ayni job'a su adimi ekleyin:

```yaml
    - name: Run optional load shedding smoke
      run: |
        BASE_URL=http://127.0.0.1:8080 \
        OVERLOAD_PATH=/internal/test/overload \
        HEALTH_PATH=/health \
        ./tests/integration/load-shed-smoke.sh
```
