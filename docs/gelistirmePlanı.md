# Gelistirme Plani

Bu belge stratejik faz siralamasini tutar.
Acik backlog'un kanonik kaynagi [PROGRESS.md](/mnt/d/w/AppFoundryLab/PROGRESS.md) dosyasidir.

Sonraki aktif hedef: Yok - repo-owned acik faz kalmadi

Son tamamlananlar:
- Faz 1 - Runtime Knob Transparency
- Faz 2 - Browser Coverage Depth
- Faz 3 - Live Smoke Cost Governance
- legacy Faz 0-5 backlog retirement ve canonical backlog reset
- Faz 1 - Admin Diagnostics Gorunurlugu
- Faz 2 - Governance Script Coverage
- Faz 3 - Host-Backed Release Confidence

## Faz 1 - Runtime Knob Transparency

Durum: tamamlandi (2026-03-06)

Odak:
- trusted proxy, logger ingest timing ve benzeri runtime knob'lari admin diagnostics ve runbook tarafinda ayni dille gorunur hale getirmek
- operator'un dependency matrix disindaki guvenlik/dayaniklilik ayarlarini da tarayicidan okuyabilmesini saglamak

Kapanis kaniti:
- `GET /api/v1/admin/runtime-config` artik request logging trusted proxy CIDR'larini ve logger timing degerlerini yayinlar
- admin diagnostics paneli runtime knob kartlariyla bu degerleri tarayicida gosterir
- EN/TR operator ve incident dokumanlari ayni kontrata cekildi

## Faz 2 - Browser Coverage Depth

Durum: tamamlandi (2026-03-06)

Odak:
- keyboard/a11y smoke ve degraded-state browser coverage'ini derinlestirmek
- operator panelleri icin davranissal regression assert'lerini arttirmak

Kapanis kaniti:
- mock-backed Playwright akislari keyboard/focus davranisini kapsar
- degraded admin diagnostics ve runtime knob fallback'lari regression zincirine baglandi
- live-stack smoke runtime knob panelini ve admin diagnostics gorunumunu dogrular

## Faz 3 - Live Smoke Cost Governance

Durum: tamamlandi (2026-03-06)

Odak:
- `e2e:live` kosusunun nightly/on-demand konumunu maliyet ve guven dengesiyle yeniden degerlendirmek
- policy, checklist ve workflow matrix'ini tek bir karar etrafinda sade tutmak

Kapanis kaniti:
- `e2e:live` sadece `RUN_LIVE_STACK_BROWSER_SMOKE=true` ile acilan nightly/on-demand release-confidence lane'i olarak sabitlendi
- branch protection ve duzenli PR kapilari mock-backed/browser-contract zincirinde tutuldu
- release policy, checklist ve workflow matrix'leri tam Docker-backed lane'in daha yuksek zaman/maliyet/flake yuzeyini ayni dille anlatir

## Stratejik Not

- Su anda yeni aktif repo fazi yoktur; acik backlog `PROGRESS.md` icinde bos durumdadir.
- Yeni bir faz acilacaksa once analiz, test kaniti ve dokuman drift kontroli ile acilmalidir.

## Tamamlanan Fazlar

- Faz 2 - Governance Script Coverage: `check-doc-drift.sh` icin fixture regresyonlari eklendi ve semantik truth kurallari test zincirine baglandi.
- Faz 3 - Host-Backed Release Confidence: `e2e:live` nightly ve on-demand release-confidence lane olarak policy, workflow ve checklist setine baglandi.
