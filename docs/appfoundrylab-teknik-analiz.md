# AppFoundryLab Teknik Analiz

Bu belge 2026-03-06 itibariyla repo gercegine dayanan kisa teknik analiz ozetidir.

## Guclu Yonler

- Topoloji dogru: frontend, gateway, logger, worker ve data servisleri rollerini net ayiriyor
- Local lifecycle scriptleri (`dev-doctor`, `bootstrap`, `dev-up`, `dev-down`) starter deneyiminin merkezinde
- Mock-backed UI regresyonu ile gercek stack browser smoke'u artik birbirinden ayrilabiliyor
- EN/TR dokuman seti, stack'i anlamak ve extension noktalarini bulmak icin yeterli tabani sagliyor

## Bu Dongude Kapatilan Basliklar

- legacy ve yeni backlog listelerinin ayni dosyada birlikte durmasi kaldirildi; `PROGRESS.md` yeniden kanonik backlog olarak kuruldu
- `dev-up` yalnizca liveness degil; readiness, logger ve authenticated admin smoke dogrular
- `dev-down --volumes` credential drift recovery icin resmi reset yolu oldu
- `quality-gate.sh ci-full` artik tam release gate'i calistirir
- frontend Biome config'i ve `bun run test` yolu tekrar calisir hale getirildi
- gercek stack icin ayri Playwright smoke yolu eklendi
- mock-backed ve live Playwright kosulari Linux bootstrap env'ini otomatik yukler hale getirildi
- invalid `LOCAL_AUTH_MODE` degerleri fail-safe davranisa cekildi
- Fibonacci validation gateway ile worker arasinda ayni limite hizalandi
- dependency-backed route davranisi kanonik bir matrix olarak belgelendi ve runtime config uzerinden yayinlandi
- `archive-runtime-report.sh` env/stdin-first hale geldi; signed evidence beklentisi dokumanlarla hizalandi
- `.env.docker.local` yeniden bootstrap ile uretilen ve ignore edilen yerel artifact olarak konumlandi
- backlog ve analiz belgeleri tek maturity anlatisi etrafinda temizlendi
- runtime config ve admin diagnostics trusted proxy CIDR'lari ile logger timing knob'larini ayni kontratta yayinlar hale geldi
- operator yuzeyi icin keyboard/focus, degraded-state ve runtime knob browser coverage'i regression zincirine baglandi
- `e2e:live` tam Docker-backed, 45 saniye test timeout'lu ve 60 dakikalik nightly butceye oturan on-demand/nightly release-confidence lane'i olarak sabitlendi

## Guncel Durus

- Bu repo release-oriented bir boilerplate olarak temiz ve calisir durumda.
- Onceki PROGRESS fazlarindaki toolchain, runtime recovery, logger health, `SystemStatus`, evidence hijyeni ve signed-attestation drift basliklari repo tarafinda kapanmis durumda.
- Runtime knob transparency, browser coverage depth ve live smoke governance fazlari da repo tarafinda kapanmis durumda.
- Kalan islerin buyuk kismi ortam sahipli operasyon icrasidir; repo backlog'u yalnizca yeni analizle tekrar acilan maddeleri tasimalidir.

## Opsiyonel Gelecek Alanlari

- daha genis operator runbook otomasyonu veya remote evidence toplama katmanlari
- daha ileri single-host observability ve rollback deneyleri
- repo-disina ait ortam sahipli release/recovery tatbikatlari

## Mimari Sign-Off

Topoloji degistirilmemeli.
Bu repoda sonraki deger, yeni operator ozellikleri eklemekten degil; seam truth, runtime hardening ve onboarding durustlugunu korumaktan geliyor.
`PROGRESS.md` acik backlog'un, `docs/gelistirmePlanı.md` ise stratejik fazlamanin tek referansi olarak kullanilmalidir.
Yeni repo-owned faz acilmadikca aktif backlog yok kabul edilmelidir.
