#!/usr/bin/env bash
# OSRM Türkiye harita preprocessing script.
# Pi'de yol-asistani-api/ dizininden çalıştır:
#   bash osrm-data/prepare.sh
#
# Gereksinim: Docker çalışıyor olmalı, osrm-data/ dizininde en az 4 GB boş alan.

set -euo pipefail

DATA_DIR="$(cd "$(dirname "$0")" && pwd)"
PBF="$DATA_DIR/turkey-latest.osm.pbf"
OSRM_IMAGE="osrm/osrm-backend"
PROFILE="$DATA_DIR/motorcycle.lua"

echo "==> Disk durumu:"
df -h "$DATA_DIR"
echo ""

# 1. PBF indir (yoksa)
if [[ ! -f "$PBF" ]]; then
  echo "==> Türkiye OSM verisi indiriliyor..."
  wget -O "$PBF" https://download.geofabrik.de/europe/turkey-latest.osm.pbf
else
  echo "==> PBF mevcut, indirme atlandı: $PBF"
fi

# 2. Extract
echo "==> osrm-extract (motorcycle profili)..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$OSRM_IMAGE" \
  osrm-extract -p /data/motorcycle.lua /data/turkey-latest.osm.pbf

# 3. Partition
echo "==> osrm-partition..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$OSRM_IMAGE" \
  osrm-partition /data/turkey-latest.osrm

# 4. Customize
echo "==> osrm-customize..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$OSRM_IMAGE" \
  osrm-customize /data/turkey-latest.osrm

echo ""
echo "==> Tamamlandı. Servis başlatmak için:"
echo "    docker compose up -d osrm"
