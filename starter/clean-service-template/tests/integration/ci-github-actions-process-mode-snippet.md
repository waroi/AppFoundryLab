# GitHub Actions Snippet (Process Mode Smoke)

Asagidaki job, template tabanli HTTP servislerde Docker kullanmadan process-mode smoke kontrolunu CI pipeline'a eklemek icin kullanilabilir:

```yaml
integration-smoke-process-mode:
  name: Integration Smoke (Process Mode)
  runs-on: ubuntu-latest
  steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Setup runtime
      run: |
        # Ornek: Go servisi icin
        # go version
        # veya gerekli runtime/install adimlarini ekle
        true

    - name: Run process-mode smoke
      run: |
        APP_CMD="go run ./cmd/service" \
        BASE_URL=http://127.0.0.1:8080 \
        HEALTH_PATH=/health \
        READY_PATH=/health/ready \
        ./tests/integration/process-mode-smoke.sh
```

Not:
- `APP_CMD` degerini kendi servisinin baslatma komutuna gore uyarlayin.
- Runtime/toolchain kurulumunu kendi stack'inize gore ekleyin.
- Opsiyonel load shedding smoke icin `RUN_LOAD_SHED_SMOKE=true` kullanin.
