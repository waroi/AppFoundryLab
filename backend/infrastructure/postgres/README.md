# Postgres Schema Management

Schema yonetimi migration framework ile yapilir.

- Runner: API gateway startup (`POSTGRES_AUTO_MIGRATE=true`)
- Framework: `golang-migrate`
- Migration dosyalari: `backend/services/api-gateway/internal/database/migrations/`

Not:
- `init.sql` dosyasi legacy referans olarak tutulur.
- Docker compose artik `init.sql` mount etmez.
