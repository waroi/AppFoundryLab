# PROGRESS.md -- Fullstack Boilerplate Analiz & Iyilestirme Plani

> Olusturma: 2026-03-06
> Son guncelleme: 2026-03-06
> Durum: Faz 0 tamamlandi, Faz 1-5 planlanmis

---

## Proje Mimarisi Ozeti

| Katman | Teknoloji | Konum |
|---|---|---|
| Frontend | Astro 5 + Svelte 5 + Tailwind 4 | `frontend/` |
| API Gateway | Go 1.23 (chi router) | `backend/services/api-gateway/` |
| Logger Service | Go 1.23 (MongoDB) | `backend/services/logger/` |
| Calculator (gRPC) | Rust (tonic) | `backend/core/calculator/` |
| Veritabani | PostgreSQL 16 + Redis 7 + MongoDB 7 | `docker-compose.yml` |
| CI/CD | GitHub Actions (13 workflow) | `.github/workflows/` |
| Deploy | Docker Compose + Caddy | `deploy/` |

---

## FAZ 0 -- Tamamlandi (Kritik Hatalar)

### 0.1 Resource Leak: PostgreSQL Pool ve Redis Client Kapatilmiyor
- **Dosya**: `backend/services/api-gateway/cmd/api-gateway/bootstrap.go:33-43`
- **Sorun**: `initDependencies` fonksiyonundaki `cleanup` kapanis fonksiyonu yalnizca `workerClient.Close()` cagiriyordu. PostgreSQL connection pool (`pgxpool.Pool.Close()`) ve Redis client (`redis.Client.Close()`) hic kapatilmiyordu.
- **Etki**: Graceful shutdown sirasinda baglanti havuzu sizintisi; `initRedis` basarisiz olursa zaten olusturulmus postgres pool asla kapatilmaz.
- **Cozum**: Cleanup fonksiyonuna `deps.pool.Close()` ve `deps.redisClient.Close()` eklendi (nil-guarded).
- **Durum**: TAMAMLANDI

### 0.2 Olü Kod: JWT Expiry Cift Kontrol
- **Dosya**: `backend/services/api-gateway/internal/middleware/auth.go:63-66`
- **Sorun**: JWT parser `jwt.WithLeeway(leeway)` ile yapilandirilmis durumda; token suresi dolmussa `ParseWithClaims` zaten `err != nil` doner. Satirlar 63-66'daki manuel kontrol hicbir zaman tetiklenemez -- olü koddur.
- **Cozum**: Redundant 4 satir kaldirildi.
- **Durum**: TAMAMLANDI

---

## FAZ 1 -- Guvenlik ve Veri Butunlugu (Oncelik: Yuksek)

### 1.1 AdminPing Handler Nil Pointer Riski
- **Dosya**: `backend/services/api-gateway/internal/handlers/admin.go:11`
- **Sorun**: `claims, _ := middleware.ClaimsFromContext(r.Context())` -- `ok` degeri yok sayiliyor. Auth middleware calismadiysa `claims` nil olur ve `claims.Role` panik atar.
- **Cozum**: `ok` degerini kontrol et; false ise 403 don.

### 1.2 Incident Monitor Thread-Safety Sorunu
- **Dosya**: `backend/services/api-gateway/internal/incidents/monitor.go:60-61, 253-254, 268-269`
- **Sorun**: `lastDispatchAt` ve `lastDispatchError` string alanlari `mu` mutex'i altinda degil; `dispatchEvent` icinde (mutex disinda) yaziliyor, `publishStats` icinde (yine mutex disinda) okunuyor.
- **Etki**: Race condition -- `go test -race` ile tespit edilebilir.
- **Cozum**: Bu alanlari `mu` mutex korumasi altina al veya `atomic.Value` kullan.

### 1.3 X-Forwarded-For Header Spoofing
- **Dosya**: `backend/services/api-gateway/internal/middleware/request_logger.go:294-303`
- **Sorun**: `clientIP()` fonksiyonu `X-Forwarded-For` header'ini dogrulama yapmadan kabul ediyor. Saldirgan IP adresini sahteleştirebilir.
- **Etki**: Log kayitlarinda yaniltici IP bilgisi. Rate limiter `RemoteAddr` kullandigi icin dogrudan guvenlik acigi degil.
- **Cozum**: Trusted proxy kontrolu ekle veya yalnizca ilk/son XFF degerini al.

### 1.4 Logger Ingest Timestamp Penceresi Cok Genis
- **Dosya**: `backend/services/logger/cmd/logger/main.go:205`
- **Sorun**: Timestamp dogrulama penceresi 30 saniye gelecege izin veriyor -- replay saldiri penceresi genis.
- **Cozum**: 5 saniyeye dusur.

### 1.5 Varsayilan Kimlik Bilgileri Kaynak Kodda
- **Dosya**: `backend/services/api-gateway/internal/runtimecfg/config.go:13-18`
- **Sorun**: `admin_dev_password` ve `developer_dev_password` kaynak kodda sabit.
- **Not**: Tespit ve uyari mekanizmasi mevcut (`DefaultCredentialsInUse`). Ancak kaynak kodda gorunurlugu hala bir risk.
- **Cozum**: Varsayilan sifreleri kaldir; tum ortamlarda cevre degiskeni zorunlu kil.

### 1.6 TLS Sertifika Yollari Goreceli
- **Dosya**: `backend/services/api-gateway/internal/worker/client.go:75, 86-87`
- **Sorun**: Varsayilan sertifika yollari `backend/infrastructure/certs/dev/...` seklinde goreceli. Calisma dizinine bagli.
- **Cozum**: Mutlak yol kullan veya sertifika bulunamazsa acik hata mesaji ver.

---

## FAZ 2 -- Hata Yonetimi ve Dayaniklilik (Oncelik: Yuksek)

### 2.1 WriteJSON Encode Hatasi Sessizce Yutulur
- **Dosya**: `backend/services/api-gateway/pkg/httpx/write_json.go:11`
- **Sorun**: `_ = json.NewEncoder(w).Encode(payload)` -- encode hatasi yok sayiliyor. Kismi JSON yaniti istemciye gidebilir.
- **Cozum**: Hatayi logla. Header zaten gonderildiginden HTTP durumu degistirilemez ancak en azindan sunucu tarafinda kayit olmali.

### 2.2 Logger Service Response Hatalari Yutulur
- **Dosyalar**: `backend/services/logger/cmd/logger/main.go:53, 57, 76, 89, 104, 131`
- **Sorun**: `_, _ = w.Write()` ve `json.NewEncoder(w).Encode()` hatalari yok sayiliyor.
- **Cozum**: En azindan hatalari logla.

### 2.3 Health Check context.Background() Kullanimi
- **Dosya**: `backend/services/api-gateway/internal/handlers/health.go:116`
- **Sorun**: `context.WithTimeout(context.Background(), 2*time.Second)` -- istemci baglanti timeout'unu yok sayar.
- **Cozum**: `r.Context()` uzerinden turetilmis context kullan.

### 2.4 PostgreSQL Singleton Baslangic Hata Kurtarma Yok
- **Dosya**: `backend/services/api-gateway/internal/database/postgres.go:19-35`
- **Sorun**: `sync.Once` kalibinda PostgreSQL baglanti havuzu bir kez basarisiz olursa, hata kalici olarak kaydedilir ve asla yeniden denenmez.
- **Etki**: Baslangicta gecici bir sorun olursa (DNS, ag gecikmesi) uygulama yeniden baslatilana kadar DB'ye eriseemez.
- **Cozum**: `sync.Once` yerine yeniden deneme mekanizmali singleton kullan veya `MustConnect` kalibina gec.

### 2.5 Circuit Breaker Failure Counter Sifirlanmiyor
- **Dosya**: `frontend/src/lib/api/fetcher.ts:94-100`
- **Sorun**: `failureCount` yalnizca basari durumunda sifirlanir; zamana bagli pencere yok. Uzun sureli sayfalarda saatler onceki hatalar birikerek circuit'i acar.
- **Cozum**: Zaman pencereli (orn. son 5 dakika) hata sayimi uygula.

### 2.6 Frontend Validator Array Icerikleri Dogrulanmiyor
- **Dosya**: `frontend/src/lib/api/validators.ts` (birden fazla satir)
- **Sorun**: `Array.isArray()` kontrolleri array icindeki elemanlarin tipini dogrulamiyor. Ornegin `loadShedExemptPrefixes` array icindeki degerlerin `string` olup olmadigi kontrol edilmiyor.
- **Cozum**: Array elemanlarini tip kontrolunden gecir.

---

## FAZ 3 -- Kod Kalitesi ve Tutarlilik (Oncelik: Orta)

### 3.1 Logger Service Tutarsiz Hata Formati
- **Dosyalar**: `backend/services/logger/cmd/logger/main.go:71, 84, 99, 116, 122, 143, 150`
- **Sorun**: Logger service `http.Error()` ile duz metin doner; API Gateway JSON hata zarfi kullanir.
- **Cozum**: Logger service'e de JSON hata zarfi ekle.

### 3.2 Magic Number'lar Sabit Olarak Tanimlanmamis
- **Dosyalar**: Birden fazla Go dosyasi
- **Ornekler**:
  - `800 * time.Millisecond` (logger HTTP client timeout) -- 4+ farkli yerde
  - `2 * time.Second` (handler timeout'lari) -- tutarsiz
  - `30 * time.Second` (timestamp dogrulama penceresi)
- **Cozum**: Paket seviyesinde sabitler tanimla.

### 3.3 Frontend Form Erisilebilirlik Eksiklikleri
- **Dosya**: `frontend/src/components/Interactive/SystemStatus.svelte:327-342`
- **Sorun**: Kullanici adi ve sifre alanlarina `<label>` veya `aria-label` atanmamis.
- **Cozum**: Tum form alanlarina uygun etiketler ekle.

### 3.4 Frontend Store Aboneligi Memory Leak
- **Dosya**: `frontend/src/components/Interactive/SystemStatus.svelte:59`
- **Sorun**: `$locale` store aboneligi `$:` reaktif ifadede kullaniliyor ancak component unmount oldugunda temizlenmiyor.
- **Cozum**: `onDestroy` ile store aboneligini temizle veya Svelte auto-unsubscribe kalibini kullan.

### 3.5 Fetch Islemleri Iptal Edilmiyor
- **Dosyalar**: `frontend/src/components/Interactive/SystemStatus.svelte`, `RestoreDrillArtifactPreview.svelte`
- **Sorun**: Hicbir fetch isleminde `AbortController` kullanilmiyor; component unmount olursa devam eden istekler havada kalir.
- **Cozum**: `AbortController` ile fetch islemlerini component yaşam dongusune bagla.

### 3.6 Dil Tercihi Kalici Depolanmiyor
- **Dosya**: `frontend/src/lib/ui/preferences.ts:61`
- **Sorun**: Tema `localStorage`'a kaydediliyor ancak dil tercihi kaydedilmiyor; sayfa yenilendiginde sifirlanir.
- **Cozum**: Dil tercihini de `localStorage`'a kaydet.

---

## FAZ 4 -- Altyapi ve DevOps (Oncelik: Orta)

### 4.1 Calculator Service Docker Health Check Eksik
- **Dosya**: `docker-compose.yml:69-77`
- **Sorun**: `calculator` servisi `condition: service_started` ile baslatiliyor, `service_healthy` degil. Health check tanimlanmamis.
- **Etki**: API Gateway, calculator hazir olmadan baglanti kurmaya calisabilir.
- **Cozum**: gRPC health check veya TCP port kontrolu ekle.

### 4.2 Frontend Service Docker Health Check Eksik
- **Dosya**: `docker-compose.yml:105-118`
- **Sorun**: Frontend servisi icin health check tanimlanmamis.
- **Cozum**: HTTP health check ekle.

### 4.3 Docker Build Cache Optimizasyonu
- **Dosya**: `frontend/Dockerfile:4`
- **Sorun**: `bun install` komutu `--frozen-lockfile` olmadan calistiriliyor; surum farklilasmasi riski var.
- **Cozum**: `--frozen-lockfile` ekle.

### 4.4 Env Dosyalarinda Varsayilan Sifreler
- **Dosyalar**: `.env.example`, `.env.docker`, `.env.docker.local`
- **Sorun**: `JWT_SECRET=replace_with_long_random_secret`, `POSTGRES_PASSWORD=appfoundrylab_secure_password` gibi varsayilan degerler mevcut.
- **Not**: Bu `example` dosyalar icin beklenen bir durum ancak `.env.docker` dosyasi calistirilabilir varsayilan degerler iceriyor.
- **Cozum**: `.env.docker` icindeki varsayilanlari bos birak veya acik uyari ekle.

### 4.5 CI Pipeline Go Test Race Flag Eksik
- **Dosya**: `.github/workflows/appfoundrylab-ci.yml:112`
- **Sorun**: `go test ./...` komutu `-race` flag'i olmadan calistiriliyor.
- **Cozum**: `go test -race ./...` olarak guncelle.

### 4.6 Docker Compose Security Overlay Network Izolasyonu
- **Dosya**: `docker-compose.security.yml:33-35`
- **Sorun**: Frontend servisi hem `backend_internal` hem `edge` aginda. Frontend'in veritabani agina dogrudan erisimi olmamali.
- **Cozum**: Frontend'i yalnizca `edge` agina bagla; API Gateway uzerinden erisim sagla.

### 4.7 Docker Container Guvenlik Sertlestirme Eksik
- **Dosya**: `docker-compose.yml` (tum servisler)
- **Sorun**: Ana compose dosyasinda `cap_drop: [ALL]`, `read_only: true`, `security_opt: [no-new-privileges:true]` yok. Starter template (`starter/clean-service-template/docker-compose.security.yml`) bu direktifleri iceriyor ancak ana projeye uygulanmamis.
- **Cozum**: Tum uygulama servislerine guvenlik direktiflerini ekle.

### 4.8 Backup Sifreleme Parolasi Bos
- **Dosyalar**: `.env.docker:88`, `.env.docker.local:80`
- **Sorun**: `BACKUP_ENCRYPTION_PASSPHRASE=` bos; yedekleme sifreleme olmadan calisir.
- **Cozum**: Bootstrap scriptinde rastgele deger olustur; backup scriptinde bos parola uyarisi ekle.

### 4.9 Prometheus Operator Parola Hash Dogrulamasi Eksik
- **Dosyalar**: `.env.docker:96`, `deploy/observability/Caddyfile.prometheus-operator:13`
- **Sorun**: `PROMETHEUS_OPERATOR_PASSWORD_HASH=` bos; Caddy basic_auth bos hash ile baslatilirsa guvenlik acigi olusabilir.
- **Cozum**: Operator erisimi aktifse hash dogrulamasi zorunlu kilinin.

---

## FAZ 5 -- Test ve Dokumantasyon (Oncelik: Dusuk)

### 5.1 E2E Test Kapsami Yetersiz
- **Dosya**: `frontend/e2e/system-status.spec.ts`
- **Sorun**: Tum frontend icin yalnizca 2 test senaryosu var.
- **Eksikler**:
  - Hata durumlari ve hata isleme
  - API basarisizliklari ve circuit breaker davranisi
  - Form dogrulama
  - Klavye navigasyonu ve ekran okuyucu destegi
- **Hedef**: Kritik yollar icin en az %50 test kapsami.

### 5.2 E2E/Smoke Testlerinde Sabit Kimlik Bilgileri
- **Dosyalar**: `frontend/e2e/system-status.spec.ts:13-14`, `frontend/scripts/smoke.mjs:38-39`
- **Sorun**: Test dosyalarinda kimlik bilgileri kodda sabit yazilmis (`admin/admin_dev_password`).
- **Cozum**: Cevre degiskenleri uzerinden oku.

### 5.3 Go Test Dosyalarinda Eksik Edge Case'ler
- **Dosyalar**: Birden fazla `_test.go` dosyasi
- **Eksikler**:
  - `request_logger_test.go` -- async log sender retry ve drop senaryolari
  - `load_shed_test.go` -- esik deger sinir kosullari
  - `health_test.go` -- stale cache senaryolari
- **Cozum**: Edge case testleri ekle.

### 5.4 Config Dogrulama Kapsami Yetersiz
- **Dosya**: `backend/services/api-gateway/internal/runtimecfg/config.go:92-136`
- **Sorun**: `Validate()` yalnizca logger ve webhook yapilandirmalarini dogruluyor. Eksikler:
  - `JWT_SECRET` varligi
  - Rate limit degerlerinin negatif olmamasi
  - TLS sertifika dosyalarinin varligi
  - Port numarasi gecerliligi
- **Cozum**: Dogrulama fonksiyonunu genislet.

### 5.5 Incomplete Frontend Type Definitions
- **Dosya**: `frontend/src/lib/api/types.ts`
- **Sorun**: Hata yanitlari icin tip tanimi yok; opsiyonel alanlar tutarsiz isretlenmis.
- **Cozum**: `ApiErrorResponse` tipi ekle; opsiyonel alan isaretlerini tutarli hale getir.

---

## Olcum Ozeti

| Kategori | Bulgu Sayisi | Tamamlanan |
|---|---|---|
| Kritik (Faz 0) | 2 | 2 |
| Guvenlik (Faz 1) | 6 | 0 |
| Hata Yonetimi (Faz 2) | 6 | 0 |
| Kod Kalitesi (Faz 3) | 6 | 0 |
| Altyapi/DevOps (Faz 4) | 9 | 0 |
| Test/Dokumantasyon (Faz 5) | 5 | 0 |
| **Toplam** | **34** | **2** |

---

## Pozitif Bulgular

Proje genel olarak olgun bir boilerplate yapisina sahip:

- JWT kimlik dogrulama leeway, issuer/audience dogrulamasi ile dogru uygulanmis
- CORS, CSP, X-Frame-Options gibi guvenlik header'lari mevcut
- Request body limit (1MB) middleware ile zorunlu kiliniyor
- HMAC-SHA256 imzalama logger ingest icin dogru uygulanmis
- Rate limiter Redis basarisizliginda fail-open/fail-closed secenegi sunuyor
- Load shedding atomic counter'lar ile dogru uygulanmis
- gRPC mTLS destegi mevcut ve yapilandirmasi esnek
- MongoDB indexleri dogru sekilde olusturuluyor
- Incident event webhook HTTPS zorunlulugu ve host allowlist mevcut
- CI pipeline kapsamli: lint, type check, build, test, security scan, SBOM, perf benchmark
- Runtime diagnostik paneli admin icin kapsamli operasyonel gorunum sagliyor
- Coklu dil destegi (EN/TR) tutarli anahtar yapisi ile uygulanmis
- Docker compose'da tum veritabani portlari `127.0.0.1` bind ile localhost'a sinirlandirilmis
- Single-host modunda `backend_internal` agi `internal: true` ile dogru izole edilmis
- Dev sertifikalari `.gitignore` ile VCS disinda tutulmus; `certs-dev.sh` ile otomatik uretiliyor
- Deployment pipeline release evidence, ledger attestation ve SBOM uretimi iceriyor
- Backup/restore altyapisi drill otomasyonu ile test edilebilir durumda
- Non-root container calistirma tum uygulama servislerinde uygulanmis
- Post-deploy kontrol scripti frontend, API ve logger endpoint'lerini dogruluyor
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
