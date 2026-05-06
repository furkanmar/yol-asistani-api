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
