# Clean Service Template

Purpose:
- Provide a minimal starter structure for future projects without heavy boilerplate.
- Keep layers explicit so architecture decisions stay clear from day one.

Structure:
- `src/domain/`: business rules and entities.
- `src/application/`: use-cases and orchestration logic.
- `src/infrastructure/`: adapters for DB, messaging, external services.
- `src/interfaces/`: API, CLI, or UI entry points.
- `src/interfaces/http/`: optional HTTP-specific helpers (ornek load shedding middleware dahil).
- `tests/`: focused tests by feature or risk area.
- `docs/adr/`: architecture decision records.

Usage:
1. Copy this folder into a new project.
2. Rename or split modules by bounded context.
3. Start with one vertical slice before broad scaffolding.
4. Keep SOLID/KISS/YAGNI decisions in ADR notes.
5. Baslangic smoke testi icin `tests/integration/smoke-http.sh` scriptini hedef endpointlere gore calistir.
6. CI entegrasyonu icin `tests/integration/ci-github-actions-snippet.md` ornegini kullan.
7. Docker ile hizli baseline calistirmasi icin `docker-compose.minimal.yml` dosyasini projene al ve servis `Dockerfile` yolunu uyarla.
8. Burst riskin varsa `src/interfaces/http/load_shedding.go.example` ve `tests/integration/load-shed-smoke.sh` dosyalarini opt-in olarak uyarlayip kullan.
9. Daha sert bir local/prod-benzeri baseline icin `docker-compose.security.yml` override'ini `docker-compose.minimal.yml` ile birlikte kullan.
10. Docker kullanmayacaksan `scripts/run-local.sh` ile process-mode lokal calistirma baslangici al.
11. Process-mode smoke/CI akisi icin `tests/integration/process-mode-smoke.sh` ve `tests/integration/ci-github-actions-process-mode-snippet.md` dosyalarini kullan.

## Variant Selection
- Kanonik karar tablosu: `docs/variant-selection-guide.md`

| Ihtiyac | Baslangic Secimi | Not |
|---|---|---|
| En kucuk Docker baseline | `docker-compose.minimal.yml` | En hizli bring-up |
| Daha sert Docker baseline | `docker-compose.minimal.yml` + `docker-compose.security.yml` | Read-only fs ve privilege dusurme |
| Docker'siz gelistirme dongusu | `scripts/run-local.sh` | Process-mode lokal baslangic |
| Process-mode CI smoke | `tests/integration/process-mode-smoke.sh` | `ci-github-actions-process-mode-snippet.md` ile birlikte |
| Overload/load shedding ihtiyaci | `src/interfaces/http/load_shedding.go.example` | Sadece gercek ihtiyac varsa opt-in |

Kural:
- Tek bir base runtime yolu secerek basla: compose-mode veya process-mode.
- Ek varyantlari sadece olculen ihtiyac varsa ekle.
