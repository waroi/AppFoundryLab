# Gelistirme Plani

Bu belge stratejik faz siralamasini tutar.
Acik backlog'un kanonik kaynagi [PROGRESS.md](/mnt/d/w/AppFoundryLab/PROGRESS.md) dosyasidir.

Sonraki aktif hedef: Faz 1 - Admin Diagnostics Gorunurlugu

## Faz 1 - Admin Diagnostics Gorunurlugu

Odak:
- dependency policy matrix'ini yalnizca API ve runbook'ta degil, admin diagnostics UI'da da gorunur hale getirmek
- operator'un `STRICT_DEPENDENCIES` ve dependency degradation kontratini tarayicidan okuyabilmesini saglamak

## Faz 2 - Governance Script Coverage

Odak:
- `check-doc-drift.sh` icin fixture tabanli regresyon testleri eklemek
- semantik doc truth kurallarini script seviyesinde daha guclu bir safety net'e cevirmek

## Faz 3 - Host-Backed Release Confidence

Odak:
- `e2e:live` kosusunun nightly-only mi, merge-blocking mi, yoksa on-demand mi olacagini yazili policy'ye baglamak
- secilen policy'yi workflow, quality gate ve dokuman seti boyunca tek bir kontrata indirmek
