# Gelistirme Plani

Bu belge stratejik faz siralamasini tutar.
Acik backlog'un kanonik kaynagi [PROGRESS.md](/mnt/d/w/AppFoundryLab/PROGRESS.md) dosyasidir.

Sonraki aktif hedef: Faz 1 - Runtime ve Toolchain Sertlestirme

## Faz 1 - Runtime ve Toolchain Sertlestirme

Odak:
- kullanilabilirlik sinyali ile process liveness sinyalini ayirmak
- toolchain dogrulugunu repo gercegiyle hizalamak
- dependency-backed route'lar icin net degrade/fail-fast davranisi tanimlamak

## Faz 2 - Frontend Bakim ve Test Kapsami

Odak:
- ana diagnostics yuzeyini daha kucuk parcali hale getirmek
- auth/runtime hata durumlarini testlerle guvenceye almak
- gorsel ve davranissal smoke kapsamını genisletmek

## Faz 3 - Security ve Evidence Hijyeni

Odak:
- hassas artifact export'larini sinirlamak
- signed attestation'i environment bazli zorunlu hale getirmek
- operator akislari icin daha guvenli credential kullanimi saglamak

## Faz 4 - Ortam Sahipli Kapanislar

Odak:
- staging ve production signed evidence harvest
- secret rotation sahipligi
- canli ortam kanit toplama disiplinini runbook seviyesinde tamamlamak
