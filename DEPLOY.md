# Pi Deploy

## İlk kurulum

```bash
# Pi'de:
git clone <repo_url>
cd yol-asistani-api

cp .env.example .env
nano .env   # POSTGRES_PASSWORD, JWT_SECRET, JWT_REFRESH_SECRET doldur
```

`.env` içine şunu da ekle:
```
POSTGRES_PASSWORD=guclu_bir_sifre
```

```bash
docker compose up --build -d
docker compose logs -f api
```

## Güncelleme

```bash
git pull
docker compose up --build -d
```

## OSRM — Türkiye Haritası Preprocessing

OSRM ilk çalıştırmadan önce harita verisi hazırlanmalı. **Tek seferlik işlem.**

### 1. Disk kontrolü

```bash
df -h
# /data partition'ında en az 4 GB boş alan gerekli:
# turkey-latest.osm.pbf  ≈ 700 MB
# osrm-extract çıktısı   ≈ 2.5 GB
# osrm-partition/customize ≈ 500 MB ek
```

### 2. Harita indir + işle

```bash
cd yol-asistani-api

# Türkiye OSM verisi (Geofabrik)
wget -P osrm-data/ https://download.geofabrik.de/europe/turkey-latest.osm.pbf

# Veya hızlı script:
bash osrm-data/prepare.sh
```

`prepare.sh` yoksa adım adım:

```bash
# Extract (motorcycle.lua profili)
docker run --rm -t \
  -v "$(pwd)/osrm-data:/data" \
  osrm/osrm-backend \
  osrm-extract -p /data/motorcycle.lua /data/turkey-latest.osm.pbf

# Partition
docker run --rm -t \
  -v "$(pwd)/osrm-data:/data" \
  osrm/osrm-backend \
  osrm-partition /data/turkey-latest.osrm

# Customize
docker run --rm -t \
  -v "$(pwd)/osrm-data:/data" \
  osrm/osrm-backend \
  osrm-customize /data/turkey-latest.osrm
```

Her adım Pi 4'te ~10-20 dakika sürer.

### 3. Servis başlat

```bash
docker compose up -d osrm
docker compose logs -f osrm
# "running and waiting for requests" görününce hazır
```

### 4. OSRM smoke test

```bash
# Ankara → İzmir yaklaşık rota
curl "http://localhost:5000/route/v1/driving/32.85,39.92;27.14,38.41?alternatives=3&overview=false"
# "code":"Ok" ve routes dizisi dönmeli
```

---

## Smoke test

```bash
# Health
curl http://localhost:8080/health

# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"12345678","display_name":"Test"}'

# Login → access_token kopyala
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"12345678"}'

# Trip oluştur (TOKEN ile)
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Authorization: Bearer TOKEN_BURAYA" \
  -H "Content-Type: application/json" \
  -d '{"title":"Amasya - İzmir","description":"Test turu"}'
```
