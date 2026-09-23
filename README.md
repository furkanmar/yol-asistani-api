# Yol Asistanı API

Backend for [Yol Asistanı](https://github.com/furkanmar/yol-asistani), a multi-day motorcycle trip planner.

## What it does

- User accounts with JWT access/refresh tokens
- Trips and ordered waypoints (create, reorder, edit)
- Route calculation and alternatives via a self-hosted **OSRM** instance with a custom **motorcycle** profile (`osrm-data/motorcycle.lua`)

## Stack

Go · Fiber · PostgreSQL (pgx) · Redis · OSRM · Docker

## Structure

```
cmd/api/          entry point
internal/
  auth/           register, login, refresh, logout
  trip/           trips CRUD
  waypoint/       waypoints CRUD + reorder
  route/          OSRM routing, alternatives, segments
  db/, util/
migrations/       SQL migrations
```

## Running

```bash
cp .env.example .env     # POSTGRES_PASSWORD, JWT_SECRET, JWT_REFRESH_SECRET
docker compose up -d --build
```

See [DEPLOY.md](DEPLOY.md) for deployment notes.
