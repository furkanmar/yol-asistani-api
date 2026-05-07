#!/usr/bin/env bash
# OSRM Türkiye harita preprocessing — ARM64 (Raspberry Pi)
# Çalıştır: bash osrm-data/prepare.sh  (yol-asistani-api/ dizininden)
#
# Adımlar:
#   1. ARM64 OSRM image build et (ubuntu:22.04 apt, ~2 dk)
#   2. Turkey PBF indir (yoksa, ~600 MB)
#   3. osrm-extract → osrm-partition → osrm-customize (~30-60 dk)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$SCRIPT_DIR"
PBF="$DATA_DIR/turkey-latest.osm.pbf"
IMAGE="yol-asistani-osrm"

echo "==> Disk durumu:"
df -h "$DATA_DIR"
echo ""

# 1. ARM64 OSRM image build
echo "==> ARM64 OSRM image build ediliyor (ilk seferinde ~2-3 dk)..."
docker build -f "$PROJECT_DIR/Dockerfile.osrm" -t "$IMAGE" "$PROJECT_DIR"
echo "==> Image hazır: $IMAGE"
echo ""

# 2. PBF indir (daha önce indirildiyse atla)
if [[ ! -f "$PBF" ]]; then
  echo "==> Türkiye OSM verisi indiriliyor (~600 MB)..."
  wget -O "$PBF" https://download.geofabrik.de/europe/turkey-latest.osm.pbf
else
  echo "==> PBF mevcut, indirme atlandı: $(du -sh "$PBF" | cut -f1)"
fi
echo ""

# 3. Extract (~10-20 dk Pi 5'te)
echo "==> osrm-extract (motorcycle profili)..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$IMAGE" \
  osrm-extract -p /data/motorcycle.lua /data/turkey-latest.osm.pbf

# 4. Partition (~5-10 dk)
echo "==> osrm-partition..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$IMAGE" \
  osrm-partition /data/turkey-latest.osrm

# 5. Customize (~2-5 dk)
echo "==> osrm-customize..."
docker run --rm -t \
  -v "$DATA_DIR:/data" \
  "$IMAGE" \
  osrm-customize /data/turkey-latest.osrm

echo ""
echo "==> Preprocessing tamamlandı!"
echo ""
echo "Servis başlat:"
echo "  docker compose up -d osrm"
echo ""
echo "Smoke test:"
echo "  curl 'http://localhost:5000/route/v1/driving/32.85,39.92;27.14,38.41?overview=false'"
