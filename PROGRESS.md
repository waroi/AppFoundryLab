# PROGRESS

Bu dosya yalnizca repo icinde halen acik olan backlog'u tutar.
Tamamlanan maddeler burada tekrar listelenmez; README, analiz dokumanlari ve commit gecmisi kapatilan isi anlatir.

## Faz 1 - Runtime ve Toolchain Sertlestirme

- [ ] `backend/go.mod` ile repo-ici Go toolchain surumunu ayni baseline'a cek
- [ ] Postgres, Redis ve Mongo istemcileri icin yeniden deneme/iyilesme davranisini kalici hale getir
- [ ] Logger `/health` sinyalini Mongo erisilebilirligiyle daha dogru hizala
- [ ] `STRICT_DEPENDENCIES=false` icin tum dependency-backed route'larda net degrade veya fail-fast politikasi tanimla

## Faz 2 - Frontend Bakim ve Test Kapsami

- [ ] `frontend/src/components/Interactive/SystemStatus.svelte` dosyasini daha kucuk, test edilebilir bolumlere ayir
- [ ] Auth hata durumlari, runtime hata durumlari ve trace lookup bos/hatali akislari icin component veya integration testleri ekle
- [ ] Locale/theme akislarini ve ana sayfa hiyerarsisini gorsel regression tabanli smoke ile destekle

## Faz 3 - Security ve Evidence Hijyeni

- [ ] Release evidence export akisinda request log ve benzeri hassas artifact'ler icin redaction/minimization politikasi ekle
- [ ] Staging/production akislarinda signed ledger attestation'i zorunlu hale getir
- [ ] Repo icindeki operator scriptlerinde positional admin password kullanimini env/stdin tabanli modele tasiyarak shell history sizintisini azalt

## Faz 4 - Ortam Sahipli Kapanislar

- [ ] Staging ortaminda signed release evidence harvest kos ve artifact zincirini dogrula
- [ ] Production ortaminda ayni signed evidence zincirini kontrollu pencereyle tekrar et
- [ ] Signed evidence secret rotation ve sahiplik modelini runbook seviyesinde netlestir

## Kabul Kriterleri

- Yerel stack icin `dev-up` yalnizca kullanilabilir runtime durumunu dogruladiginda basarili kabul edilir
- Mock-backed UI regresyonu ile gercek stack browser smoke'u birbirinden net ayrilir
- README, quick-start, testing docs ve analiz belgeleri ayni maturity sinyalini verir
- `PROGRESS.md` repo backlog'unun tek kanonik kaynagi olarak kalir
