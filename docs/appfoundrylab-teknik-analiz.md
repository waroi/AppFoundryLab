# AppFoundryLab Teknik Analiz

Bu belge AppFoundryLab'in guncel kanonik teknik analiz ozeti ve karar notudur.

## Genel durum

 Proje; sadece calisan uygulama iskeleti degil, operasyon yuzeyini de acik eden bir polyglot full-stack boilerplate tabani sunuyor. Frontend, API gateway, logger, Rust worker, veri servisleri, kalite kapilari ve iki dilli dokumantasyon ayni temelde birlesiyor. Son turda odak sadece incident diagnostics degil; frontend sunum katmani da artik route-tabanli locale girisi, acik/koyu tema modeli, SSR-correct `html[lang]` ve test isaretcileri netlestirilmis bir yapiya kavustu.

## Bu turdaki ana gelistirmeler

- single-host deploy artik `build` ve `image` modlarini ayri destekliyor
- `deploy/docker-compose.single-host.ghcr.yml` ile digest-pinned GHCR image deploy varyanti eklendi
- staging ve production SSH workflow'lari pinned `known_hosts` ile sertlestirildi
- deploy artifact zinciri artik runtime archive, deploy manifest ve compose image snapshot'i birlikte uretiyor
- `scripts/backup-single-host.sh` ile Postgres ve Mongo backup'lari bundle manifest, sha256 ve opsiyonel encryption ile ayni paket haline getirildi
- off-host backup sync, versioned target catalogu ve retention temizligi tek script/env modeli uzerine tasindi
- `scripts/restore-drill-single-host.sh` artik tek marker yerine bundle'a yazilan deterministik business fixture ve kanonik verification artifact'lari ureten disposable restore drill akisi sunuyor
- `restore-drill-single-host.yml` ile bu akisin repo icinde tekrar edilebilir CI karsiligi tanimlandi
- logger servisine `GET /request-logs` ve `GET /metrics/prometheus` eklendi
- gateway artik `GET /api/v1/admin/request-logs` ile traceId odakli request log sorgusu sunuyor
- `deploy/docker-compose.observability.yml` ve `deploy/observability/prometheus.yml` ile private Prometheus scrape katmani eklendi
- incident webhook akisi HMAC ve allowlist kurallariyla sertlestirildi
- `publish-ghcr-images.yml` ile GHCR publish + image-mode validation workflow'u eklendi
- `publish-ghcr-images.yml` artik ayni immutable manifesti staging deploy, staging rollback yolu, staging restore drill ve production deploy zincirine bagliyor
- `scripts/release-catalog.sh` ile release manifest katalogu, selector tabanli rollback ve release-ledger JSON export akisi eklendi
- `scripts/collect-release-evidence.sh` ile katalog + ledger yapisi Markdown/JSON evidence summary formatina baglandi
- `scripts/attest-release-ledger.sh` ve `scripts/verify-release-ledger-attestation.sh` ile release-ledger kanit zincirine attestation katmani eklendi
- signed attestation zorunlulugu artik `LEDGER_ATTESTATION_REQUIRE_SIGNED=true` ile enforce edilebiliyor
- frontend production image'i `PUBLIC_API_BASE_URL` degerini runtime'da `/runtime-config.js` uzerinden aliyor
- frontend admin paneli artik traceId ile request-log sorgusu ve son log akislarini gosterebiliyor
- frontend tarafina Playwright tabanli browser regression katmani ve restore-drill artifact preview sayfasi eklendi
- `scripts/bootstrap-playwright-linux.sh` ile Linux ortamlarinda Playwright Chromium bootstrap akisi resmi hale getirildi
- frontend unit testleri `bun:test` bagimliligindan cikarilip `vitest` kosusuna hizalandi ve `@/*` alias cozumlemesi `vitest.config.ts` icinde netlestirildi
- frontend e2e kosusu `bun run e2e` ile Windows ve Linux'ta ayni sekilde calisacak hale getirildi; Playwright `webServer` komutu shell-spesifik yol cagrisindan arindirildi
- API gateway Docker build katmani `golang:1.24-alpine` seviyesine cekildi; `go.mod` ile CI/runtime toolchain drift'i kapatildi
- `toolchain.versions.json` ile CI workflow'larindaki Go versiyonlari `1.24.x` ile senkronlandi
- `internal/incidents/monitor.go` icinde derlemeyi kiran sink secim yolu yeniden tanimlandi (`sinkEnabled` + fallback parse)
- kok seviyesinde `.gitattributes` ile `*.sh` dosyalari LF standardina alindi
- `scripts/single-host-common.sh` icinde env degerleri okunurken trailing `CR` temizlenerek Windows checkout kaynakli endpoint kirilmasi giderildi
- `frontend/bun.lock` normalize edilerek `Dockerfile.prod` icindeki `bun install --frozen-lockfile` adimi tekrar gecilir hale getirildi
- `scripts/rehearse-release-evidence-local.sh` ile local release evidence zinciri (catalog + ledger + attestation + verify) tekrarli sekilde dogrulandi
- `scripts/check-operator-mtls-readiness.sh` sertifika gecerlilik kontrolu GNU `date -d` bagimliligindan cikarilip `openssl -checkend` uzerine tasindi
- off-host backup sync profili S3/object-storage hedefleri icin genisletildi
- operator erisimi gerekiyorsa kullanilmak uzere Prometheus basic-auth ve mTLS proxy overlay'leri eklendi
- `release-evidence-harvest.yml` ile staging ve production evidence kataloglarini periyodik toplama akisi tanimlandi
- `scripts/export-release-evidence.sh` ile release evidence paketleri yerel klasor, SCP ve S3 audit hedeflerine aktarilabilir hale geldi
- `scripts/rehearse-release-evidence-local.sh` ile staging/production-benzeri evidence zinciri local single-host deploy uzerinden prova edilebilir hale geldi
- `backup-lifecycle-drift.yml`, `scripts/check-s3-lifecycle-policy.sh` ve `deploy/backups/s3-lifecycle-policy.example.json` ile S3 retention davranisi drift check seviyesine cikarildi
- `scripts/generate-operator-mtls-certs.sh`, `scripts/check-operator-mtls-readiness.sh` ve `docs/operator-observability-runbook.md` ile operator mTLS overlay'i runbook seviyesinde tamamlandi
- local operasyon scriptleri `DOCKER_BIN` degiskeni ile Docker Desktop `docker.exe` gibi ortamlara uyumlu hale getirildi
- `scripts/single-host-common.sh` ve `scripts/dev-doctor.sh`, varsayilan `docker` komutu Podman shim'e gidip compose daemon'a baglanamiyorsa Docker Desktop `docker.exe` binary'sini otomatik sececek sekilde sertlestirildi
- local Docker dev akisi artik ayri `*_HOST_PORT` degiskenleri ve `DOCKER_HOST_BIND_ADDRESS=127.0.0.1` varsayilani ile host publish yuzeyini service portlarindan ayiriyor; `APP_ENV_FILE` uzerinden `.env.docker.local` degerleri konteynerlere dogru tasiniyor
- `scripts/dev-up.sh`, compose `up` sonrasinda Postgres ve Mongo icin gercek auth kontrolu yaparak persist edilen volume credential drift'ini erken yakalar hale getirildi
- `scripts/check-doc-drift.sh`, `git` binary'si olmayan strict shell'lerde `DOC_DRIFT_CHANGED_FILES` fallback'i ile deterministic denetim yapacak sekilde guncellendi
- runtime diagnostics artik cache'lenmis snapshot uzerinden yeniden kullaniliyor ve readiness/logger/incident probe'lari paralel toplanarak ilk admin response latency'si kisaltiliyor
- logger incident summary artik tek Mongo aggregation pipeline'i ile hesaplaniyor; admin paneli de request-log yukunu kritik ilk render yolundan cikariyor
- frontend artik `frontend/src/lib/ui/preferences.ts` ile merkezi locale/theme state katmani, `frontend/src/lib/ui/copy.ts` ile EN/TR shell/admin metin sozlugu, `frontend/src/lib/ui/routes.ts` ile localized path kontrati ve `BaseLayout.astro` icinde route-correct `html[lang]` + pre-paint theme bootstrap kazandi
- `global.scss` semantik surface/text/control token'lari uzerinden acik/koyu tema katmanina tasindi; koyu tema artik charcoal zemin + canli turuncu CTA aksani kullanirken hero, test template, diagnostics ve restore-drill yuzeyleri ayni token modelini paylasiyor
- frontend `HomePage.astro` ve `TestPage.astro` shell'leriyle `/`, `/test`, `/tr` ve `/tr/test` sayfalarini ayni composition uzerinden uretiyor; language switch ise ayni mantiksal sayfanin localized route'una navigation yapiyor
- frontend smoke ve Playwright dogrulama katmani artik gorunur Ingilizce metinlere baglanmak yerine `data-testid`, `data-role`, `data-mode`, `data-status`, localized URL ve `html[lang]`/`html[data-theme]` odakli selector'larla locale/theme degisimine daha dayanikli hale getirildi
- `scripts/bootstrap-playwright-linux.sh`, Ubuntu `apt download` aday paketi 404 donerse `apt-cache madison` uzerinden bilinen onceki paket surumlerine geri cekilerek Playwright runtime kutuphanelerini local toolchain altinda toparliyor

## Guclu yonler

- deploy, rollback, backup, restore drill ve observability artik ayni compose topolojisi etrafinda toplanmis durumda
- image-mode deploy checkout tabanli akisi bozmadan eklenmis oldugu icin gecis riski dusuk
- request log trace backend'i mevcut logger/Mongo mimarisini yeniden kullanarak ek operasyonel servis maliyeti olusturmuyor
- runtime diagnostics ve incident response yolu artik ayni veriyi tekrar tekrar toplamiyor; cache + paralel probe modeli operatör panelini daha dusuk maliyetle besliyor
- Prometheus overlay 127.0.0.1 ile sinirli kalirken Caddy ornegi de onu varsayilan public surface'e eklemedigi icin metrics exposure riski azaltilmis durumda
- local Docker publish katmani artik varsayilan olarak `127.0.0.1` ile sinirli ve host port degerleri container servis portlarindan ayrildigi icin yerel port cakismalarini gidermek operasyonel olarak daha guvenli hale geldi
- backup bundle manifest'i operasyonel kanit toplamayi kolaylastiriyor
- restore drill fixture'i bundle ile birlikte tasindigi icin ayni backup paketi daha sonra deterministik sekilde yeniden dogrulanabiliyor
- ledger attestation katmani release kanit zincirini sadece JSON export seviyesinde birakmiyor; dogrulanabilir hash/signature seviyesine tasiyor
- Playwright bootstrap akisi artik sadece dokumantasyon tavsiyesi degil; CI frontend-check zincirinin resmi parcasi
- release evidence export zinciri sayesinde katalog, ledger, attestation ve summary ailesi uzun omurlu audit storage'a ayni semantik ile aktarilabiliyor
- local rehearsal komutu repo ici olgunluk ile ortam bagimli canli rollout arasindaki siniri netlestiriyor; eksik olan kod degil, canli host icrasi oluyor
- frontend locale/theme mantigi tek store + tek sozluk etrafinda toplandigi icin yeni gorunur metin veya yeni tema davranisi eklemek artik dosya daginikligi yerine belirgin kontratlar uzerinden ilerliyor
- URL-tabanli locale girisi sayesinde non-default dil artik hydration sonrasi degil, ilk statik boya anindan itibaren dogru shell/title/lang ile geliyor
- pre-paint theme bootstrap koyu tema flash'ini azaltirken, semantic CSS token'lari sayesinde ayni kart/toolbar/form dili tum frontend yuzeyine tutarli yayiliyor
- koyu temanin cool-blue tabandan charcoal tabana alinmasi ve CTA aksaninin turuncuya cekilmesi, admin agirlikli aksiyonlar ile warning/alert semantiklerini ayni token sistemi icinde daha net ayiriyor
- smoke ve e2e selector'lari metin bagimliligindan ciktigi icin locale degisiklikleri regression sinyalini kirletmeden test edilebiliyor
- Playwright bootstrap fallback'i root yetkisi gerektirmeden Linux browser regression yolunu daha dayanikli hale getiriyor
- frontend script lint sinyali iyilestirildi; e2e/smoke helper ciktilari `process.stdout.write` ile noConsole uyarilarini kirletmeden korunuyor

## Kalan bosluklar

- repo ici boilerplate acisindan kritik bir bosluk kalmadi; kalan maddeler canli staging/production ortamlarinda kanit toplama ve secret dagitimi gibi environment sahipligindeki operasyonlar
- signed ledger modu icin `RELEASE_LEDGER_ATTESTATION_KEY` secret'inin hedef ortamlarda yuklenmesi gerekiyor; `LEDGER_ATTESTATION_REQUIRE_SIGNED=true` ise secret eksigini artik sessizce gecmek yerine fail ederek gorunur kiliyor
- observability katmani su an Prometheus + logger trace query baseline'i ve basic-auth/mTLS operator proxy sunuyor; tam OTLP/collector topolojisi halen ancak hacim gerektirirse anlamli olacak bir sonraki asama
- performans tarafinda repo ici belirgin darboğaz kalmadi; kalan calisma gercek host benchmark ve kanit toplama asamasidir
- locale gecisi artik SSR-correct route navigation uzerinden yapiliyor; bunun trade-off'u canli dil degisiminde tam sayfa navigation olmasi. Bu secim non-default locale icin ilk boya dogrulugunu, temiz deep-link davranisini ve test netligini onceledigi icin bilincli olarak secildi
- LF zorlamasi (`.gitattributes`) ve CR trim duzeltmeleri yapildi; kalan risk daha cok farkli shell dagitimlarinda (`bash` + git kurulu degil gibi) komut bulunabilirligi
- frontend lint kapisi artik Biome 2.x ile calisiyor; sonraki hedef warning-level kurallarin kademeli strict seviyeye alinmasi
- `scripts/test-dev-scripts.sh` bootstrap fixture'lari, `openssl` olmayan minimal shell imajlarinda fail ediyor; test harness icin portable fallback stratejisi halen acik bir iyilestirme maddesi

## Sonraki odak

1. canli GitHub environment'larda ilk signed evidence harvest turunu kosmak
2. operator mTLS sertifika rotasyonunu runbook disiplininde periyodik hale getirmek
3. ihtiyac dogarsa collector topolojisine gecmek
