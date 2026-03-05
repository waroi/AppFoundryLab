# Gelistirme Rehberi

## 1. Proje haritasi

- `frontend/`: Astro + Svelte uygulamasi
- `backend/services/api-gateway/`: HTTP API, JWT auth, RBAC, health, metrics, incident monitor
- `backend/services/logger/`: request log ingest ve kalici incident journal
- `backend/core/calculator/`: Rust gRPC worker
- `scripts/`: yerel otomasyon ve kalite kapilari
- `.github/workflows/`: CI/CD

## 2. Tavsiye edilen yerel dongu

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

## 3. Frontend degisiklik haritasi

Frontend sunum degisikliklerini su alanlarda gelistir:

- document shell ve pre-paint preference bootstrap: `frontend/src/layouts/BaseLayout.astro`
- localized route mapping: `frontend/src/lib/ui/routes.ts`
- locale/theme store ve normalization: `frontend/src/lib/ui/preferences.ts`
- ortak metin sozlugu ve formatter'lar: `frontend/src/lib/ui/copy.ts`
- varsayilan ve localized route'lar icin ortak page shell'leri: `frontend/src/components/Page/`
- ortak kontroller: `frontend/src/components/Layout/`
- locale-reactive shell copy: `frontend/src/components/Static/`
- diagnostics ve restore-drill yuzeyleri: `frontend/src/components/Interactive/`
- semantik token ve tema siniflari: `frontend/src/styles/global.scss`

Kurallar:

- yeni gorunur her metni `frontend/src/lib/ui/copy.ts` icinde iki locale ile birlikte ekle
- locale route'larini elle kurma; kanonik path uretimi icin `frontend/src/lib/ui/routes.ts` kullan
- ayni ihtiyaci ortak semantik sinifla ifade edebiliyorsan yeni light-only utility renk ekleme
- smoke/e2e icin onemli bir UI durumu varsa gorunur cevrilmis metin yerine stabil `data-testid` veya `data-*` isaretcisi ekle

## 4. Backend degisiklik haritasi

Yeni isleri su alanlarda gelistir:

- handler: `backend/services/api-gateway/internal/handlers/`
- middleware: `backend/services/api-gateway/internal/middleware/`
- runtime config: `backend/services/api-gateway/internal/runtimecfg/`
- incident monitor: `backend/services/api-gateway/internal/incidents/`
- logger kaliciligi: `backend/services/logger/internal/`

## 5. Admin diagnostics endpoint'leri

- `GET /api/v1/admin/runtime-config`
- `GET /api/v1/admin/runtime-metrics`
- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`

## 6. Pratik kural

Yeni bir operasyonel davranis eklediginde:

1. kodu yaz
2. testi yaz
3. ayni change set icinde dokumani guncelle
4. mimari anlamli sekilde degistiyse proje analizi ve gelistirme planini da guncelle

## 7. Sonraki okumalar

- [hizli-baslangic.md](/mnt/d/w/AppFoundryLab/docs/tr/hizli-baslangic.md)
- [operasyonlar.md](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- [incident-yonetimi.md](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- [deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
