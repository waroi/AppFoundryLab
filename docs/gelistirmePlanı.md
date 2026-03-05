# Gelistirme Plani (Acik Maddeler)

Bu belge 5 Mart 2026 itibariyla sadece acik teknik ve operasyonel backlog maddelerini tutar.
Tamamlanan maddeler bu dosyada tekrar listelenmez.

## Faz 4 - Canli Kanit Kapanisi

1. ENV-REL-001 (Pending): GitHub staging ortaminda signed release evidence harvest kosusu al ve artifact'leri release ledger ile eslestir.
2. ENV-REL-002 (Pending): Production ortaminda ayni signed evidence zincirini kontrollu pencerede tekrar et.
3. ENV-REL-003 (Pending): Signed evidence kosulari icin secret rotasyonu ve sahiplik modelini runbook'a bagla.

## Faz 5 - Gelistirici Deneyimi

1. DEV-OPS-001 (Pending): `scripts/dev-down.sh` icin opsiyonel `--volumes` modu ekleyerek credential drift recovery adimini tek komuta indir.
2. DEV-OPS-002 (Pending): `dev-up` credential check hatasinda otomatik rehber ciktisini platforma gore (WSL/Linux) daha yonlendirici hale getir.

## Faz 6 - Kalite Kapisi Sertlestirme

1. QA-001 (Pending): `scripts/test-dev-scripts.sh` icin `openssl` olmayan minimal shell imajlarinda portable fallback stratejisi tasarla.
2. QA-002 (Pending): Frontend lint warning-level kurallari asamali olarak strict seviyeye cekilecek sekilde Biome policy plani cikart.

## Cikis Kriterleri

- staging ve production signed evidence zinciri en az birer kez basariyla calismis olacak
- kritik local dev komutlari (`dev-doctor`, `dev-up`, `dev-down`) ek manuel adim olmadan calisacak
- CI ve local kalite kapilarinda skip gerektiren cevresel bagimliliklar net ve minimum seviyede olacak
