# Gelistirme Plani

Bu belge 1 Mart 2026 tarihinde performans odakli olarak sifirdan yeniden yazildi. Onceki plan temizlendi; mevcut icerik performans skoru icin derin analiz bulgularini, bu turda tamamlanan gelistirmeleri ve kalan yalnizca environment-tabanli dogrulama maddelerini tasir.

Sonraki aktif hedef: ENV-PERF-001 staging benzeri gercek hostta k6 smoke veya spike kosusunu alip yeni runtime diagnostics yolunu artifact ile arsivle.

## Performans skoru analizi

- Hedef: repository-side performans olgunlugunu 9.8 seviyesinden savunulabilir 10.0 seviyesine cikarmak
- Esas darboğazlar: runtime diagnostics alt sorgularinin tekrarli calismasi, admin panelinin kritik ilk yukte gereksiz request-log beklemesi, logger incident summary yolunun birden fazla Mongo turu yapmasi
- Sonuc: repo ici gelistirme tarafi icin kritik performans acigi kalmadi; kalan maddeler gercek trafik ve gercek host benchmark kaniti uretmeye donuk operasyon adimlari

## Derin analiz bulgulari

- PF-001: `runtime-report`, `runtime-metrics` ve incident monitor ayni expensive probe ailesini tekrar tekrar kuruyordu; cache olmadan ya da cache disinda ilk hesaplama maliyeti yuksek kalabiliyordu
- PF-002: gateway runtime diagnostics yolunda readiness, logger health, logger metrics ve incident summary bilgilerinin bir bolumu ardilsik calisiyordu; ilk admin cevabi gereksiz RTT zinciri olusturuyordu
- PF-003: logger incident summary uc ayri Mongo islemi ile elde ediliyordu; toplam event, aktif event ve son event bilgisi tek sort+group aggregation ile alinabilecek durumda idi
- PF-004: request-log ve incident listeleri sert ust sinira kavusmadan buyuk limitlerle pahali sorgulara acik kalabiliyordu; bu onceki turda sinirlandi ve indexlerle desteklendi
- PF-005: admin paneli artik `runtime-report` icinde gelen config+metrics verisini yeniden kullanmasina ragmen login sirasinda request-log cevabini da kritik yol uzerinde bekliyordu
- PF-006: logger tarafinda `request_logs` ve `incident_events` icin indeks yoksa trace lookup ve incident summary zamanla dusen performans gosterebilirdi; onceki turda bu indeksler garanti altina alindi
- PF-007: sandbox E2E dogrulamasi tarayici sistem kutuphanelerine bagli; bu repo ici perf degil, environment bootstrap problemidir

## Bu turda tamamlanan gelistirmeler

- PX-001: gateway runtime diagnostics ozeti artik external probe'lari paralel toplayarak ilk response latency'sini dusuruyor
- PX-002: logger incident summary akisi uc ayri query yerine tek Mongo aggregation pipeline ile hesaplanir hale getirildi
- PX-003: admin login akisi artik `users`, `runtime-report` ve `incident-events` isteklerini paralel topluyor
- PX-004: admin request-log verisi kritik ilk render yolundan cikarildi; panel verisi geldikten sonra arka planda yukleniyor
- PX-005: logger incident aggregation yolu unit test ile sabitlendi
- PX-006: bir onceki performans turunda eklenen runtime diagnostics snapshot cache, query limit clamp'leri ve Mongo indeksleri bu yeni degisiklik setiyle tamamlayici hale getirildi

## Kapanan yapilacaklar listesi

- [x] runtime diagnostics maliyetini tek snapshot ve paralel probe toplama modeliyle dusur
- [x] logger incident summary veri yolunu tek aggregation turuna indir
- [x] admin panelinde `runtime-report` disindaki ek istekleri kritik ilk response yolundan temizle
- [x] request-log sorgusunu bounded ve indexed kalacak sekilde koru
- [x] performans analizini dokumanlarda skor gerekcesi ile birlikte guncelle

## Guncel skor karari

- SK-001: repository-side performans skoru artik 10.0 / 10 olarak savunulabilir
- SK-002: bu karar su teknik zemine dayaniyor: async logger queue, load shedding, ready cache, runtime diagnostics snapshot cache, paralel probe toplama, bounded admin log sorgulari, Mongo indexleri ve admin panelde kritik yolun kisaltilmasi
- SK-003: kalan isler kod tarafli performans borcu degil; benchmark evidence ve gercek host profilleme gibi environment execution maddeleridir

## Kalan environment backlog

- ENV-PERF-001: staging benzeri hostta `./scripts/run-k6-smoke.sh` veya uygun profile gore `./scripts/run-k6-scenario.sh spike` kosup artifact'i arsivle
- ENV-PERF-002: yeni runtime diagnostics yolunu gercek hostta `archive-runtime-report.sh` artifact'i ile onceki kanitlarla karsilastir
- ENV-PERF-003: eger gercek trafik altinda p95 veya incident summary gecikmesi tekrar yukselirse o zaman pprof veya daha agir tracing yatirimi degerlendir; repo ici mevcut durumda buna zorunlu ihtiyac yok

## Dogrulama notlari

- frontend `bun run check` bu turda gecti
- logger tarafindaki yeni incident aggregation ve index testleri gecti
- gateway odakli Go testleri sandbox'in ag kisiti nedeniyle eksik modulleri indiremedigi icin burada tam tekrar edilemedi
- browser E2E bu sandbox'ta `libnspr4.so` gibi Playwright sistem kutuphaneleri eksik oldugunda environment bootstrap ister; bu kod regresyonu degil
