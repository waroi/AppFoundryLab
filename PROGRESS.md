# PROGRESS (Acik Isler) - 2026-03-05

Bu dosyada sadece acik ve gelistirme bekleyen maddeler tutulur.
Tamamlanan maddeler bu dosyadan kaldirilir ve commit gecmisinde izlenir.

## x12 Orkestrasyon Dagitimi (Guncel)

1. Team Lead: Fazlar arasi koordinasyon ve PR cikti birlestirme
2. Principal Architect: Mimari risk degerlendirmesi ve teknik kararlar
3. Research Analyst 1: Runtime/ops script analizi
4. Research Analyst 2: Frontend kalite kapilari analizi
5. Research Analyst 3: CI ve governance kontrol analizi
6. QA Guardian 1: Unit ve integration test kapilari
7. QA Guardian 2: Browser/e2e regresyon kapilari
8. Security Reviewer 1: Evidence, ledger, attestation zinciri
9. Security Reviewer 2: mTLS/operator erisim kontrolleri
10. Visual Researcher 1: UI smoke ve localized route dogrulamasi
11. Visual Researcher 2: Selector dayanikliligi ve tema gecisleri
12. Team Lead 2: Dokumantasyon senkronizasyonu

## Faz Backlogu

### Faz 4 - Canli Evidence Kapanisi

- [ ] ENV-REL-001: Staging ortaminda signed evidence harvest kos ve artifact zincirini raporla
- [ ] ENV-REL-002: Production ortaminda signed evidence harvest kos ve staging ciktilariyla karsilastir
- [ ] ENV-REL-003: Signed evidence secret/rotation sahipligini runbook seviyesinde finalize et

### Faz 5 - Developer UX Sertlestirme

- [ ] DEV-OPS-001: `dev-down` icin `--volumes` secenegi ekleyip credential drift recovery adimini kisalt
- [ ] DEV-OPS-002: `dev-up` credential check fail durumunda platforma gore otomatik kurtarma komutlari sun

### Faz 6 - Kalite Kapisi Sertlestirme

- [ ] QA-001: `test-dev-scripts` icin openssl olmayan minimal shell ortamlarda portable fallback modeli
- [ ] QA-002: Frontend lint policy'sinde warning-level kurallari asamali strict plana baglama

## Dogrulama Beklentisi

- Her faz sonunda: `frontend lint/test/e2e`, `dev-doctor`, `dev-up`, `check-doc-drift --mode strict`, `check-release-policy-drift`
- Faz 4 kapanisinda: signed evidence artifact seti ve runbook kayitlari zorunlu
