# Hizli Baslangic

## 1. Yerel stack'i kaldir

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Eger `dev-doctor` WSL icinde `docker compose unavailable` diyorsa Docker Desktop WSL integration'i acin veya su sekilde tekrar deneyin:

```bash
DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe" ./scripts/dev-doctor.sh
DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe" ./scripts/dev-up.sh standard
```

## 2. Yerel saglik kontratini anla

- `GET /health/live` sadece gateway prosesinin ayakta oldugunu soyler
- `GET /health/ready` dependency-backed stack'in kullanilabilir oldugunu soyler
- Frontend tarafindaki `GET /healthz` hafif shell health endpoint'idir
- `dev-up` artik basari demeden once readiness, logger erisimi ve bir authenticated admin endpoint'i dogrular

Varsayilan adresler:
- Frontend: `http://127.0.0.1:4321/`
- Frontend health: `http://127.0.0.1:4321/healthz`
- API live: `http://127.0.0.1:8080/health/live`
- API ready: `http://127.0.0.1:8080/health/ready`
- Logger metrics: `http://127.0.0.1:8090/metrics`

## 3. Ilk tarayici smoke dogrulamasi

- `http://127.0.0.1:4321/` adresini acin
- `admin` kullanicisi ile giris yapin
- Sifre olarak `./scripts/bootstrap.sh` cikisindaki veya `.env.docker.local` icindeki `BOOTSTRAP_ADMIN_PASSWORD` degerini kullanin
- Runtime ozeti, trace lookup paneli ve request log listesi yukleniyorsa ilk dogrulama tamamdir

Gercek stack browser smoke:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e:live
```

Mock-backed hizli UI regresyonu:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e
```

## 4. Credential drift durumunda yerel reset

Persist edilmis Postgres veya Mongo volume'leri yeni credential'lari kabul etmiyorsa:

```bash
./scripts/dev-down.sh standard --volumes
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

## 5. Sonraki dokumanlar

- [Gelistirme Rehberi](/mnt/d/w/AppFoundryLab/docs/tr/gelistirme-rehberi.md)
- [Operasyonlar](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- [Test ve Kalite](/mnt/d/w/AppFoundryLab/docs/tr/test-ve-kalite.md)
- [Proje Analizi](/mnt/d/w/AppFoundryLab/docs/tr/proje-analizi.md)
