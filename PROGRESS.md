# PROGRESS

Bu dosya yalnizca repo icinde halen acik olan backlog'u tutar.
Tamamlanan maddeler burada tekrar listelenmez; README, analiz dokumanlari ve commit gecmisi kapatilan isi anlatir.

## Faz 1 - Admin Diagnostics Gorunurlugu

- [ ] `GET /api/v1/admin/runtime-config` icindeki `dependencyPolicies` alanini admin diagnostics UI'da okunur bir ozet olarak goster
- [ ] Bu yeni diagnostics yuzeyi icin EN/TR copy ve component veya integration coverage ekle

## Faz 2 - Governance Script Coverage

- [ ] `scripts/check-doc-drift.sh` icin fixture tabanli shell testleri ekle
- [ ] Semantic truth checks'i archive usage, signed evidence ve mock/live smoke ayrimi senaryolariyla regression altina al

## Faz 3 - Host-Backed Release Confidence

- [ ] `e2e:live` kosusunun nightly-only, merge-blocking veya on-demand modellerinden hangisinde tutulacagina karar ver
- [ ] Secilen policy'yi workflow, `quality-gate.sh` semantigi ve dokuman seti boyunca tek kontrat haline getir

## Kabul Kriterleri

- `PROGRESS.md` yalnizca gercek repo-owned aciklari listeler
- Ortam sahipli staging/production evidence icralari runbook seviyesinde kalir; backlog'u kirletmez
- Mock-backed UI regresyonu ile gercek stack browser smoke'u ayri fakat ayni dokuman gercegine bagli kalir
- Toolchain, kalite ve deployment dokumanlari ayni maturity sinyalini verir
