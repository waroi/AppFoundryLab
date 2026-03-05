# Hizli Baslangic

## 1. Ilk yerel calistirma

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Varsayilan local portlar doluysa `.env.docker.local` icinde `FRONTEND_HOST_PORT`, `API_GATEWAY_HOST_PORT` ve `LOGGER_HOST_PORT` degerlerini guncelleyin veya bu degiskenleri export edip stack'i yeniden baslatin. Ornek: `FRONTEND_HOST_PORT=14321 API_GATEWAY_HOST_PORT=18080 LOGGER_HOST_PORT=18090 ./scripts/dev-up.sh standard security`. Local Docker yayinlari varsayilan olarak `DOCKER_HOST_BIND_ADDRESS=127.0.0.1` ile host-local tutulur.

## 2. Stack'i ac

- Frontend: `http://127.0.0.1:<FRONTEND_HOST_PORT>/` (varsayilan: `http://127.0.0.1:4321/`)
- Frontend test sayfasi: `http://127.0.0.1:<FRONTEND_HOST_PORT>/test` (varsayilan: `http://127.0.0.1:4321/test`)
- Turkce ana sayfa: `http://127.0.0.1:<FRONTEND_HOST_PORT>/tr`
- Turkce test sayfasi: `http://127.0.0.1:<FRONTEND_HOST_PORT>/tr/test`
- API gateway: `http://127.0.0.1:<API_GATEWAY_HOST_PORT>` (varsayilan: `http://127.0.0.1:8080`)
- Logger metrics: `http://127.0.0.1:<LOGGER_HOST_PORT>/metrics` (varsayilan: `http://127.0.0.1:8090/metrics`)

## 3. Locale ve theme dogrulamasi

- `/` ve `/test` sayfalarindaki sag ust toolbar'i kullan
- `EN` ve `TR` arasinda gecis yap
- `Light` ve `Dark` arasinda gecis yap
- dil gecisinin `/` ile `/tr` veya `/test` ile `/tr/test` arasinda navigation yaptigini dogrula
- Sayfayi yenileyip mevcut URL'nin locale'i korudugunu ve secilen theme'in kalici oldugunu kontrol et
- Document icindeki `html[lang]` ve `html[data-theme]` degerlerinin secime gore degistigini kontrol et

## 4. Admin girisinden sonra ne denenmeli

- runtime config'i incele
- runtime metrics'i incele
- runtime report indir
- incident report indir
- son incident event'leri incele

## 5. Sonraki dokumanlar

- [gelistirme-rehberi.md](/mnt/d/w/AppFoundryLab/docs/tr/gelistirme-rehberi.md)
- [operasyonlar.md](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- [incident-yonetimi.md](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- [deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)

## 6. Tek sunucu deployment paketini localde denemek istersen

```bash
cp .env.single-host.example .env.single-host
./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/archive-runtime-report.sh http://127.0.0.1:<API_GATEWAY_HOST_PORT> admin guclu_sifre
```
