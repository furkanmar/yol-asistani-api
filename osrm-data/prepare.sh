#!/usr/bin/env bash
# OSRM Türkiye harita preprocessing — ARM64 (Raspberry Pi 5)
# Çalıştır: bash osrm-data/prepare.sh  (yol-asistani-api/ dizininden)
#
# Adımlar:
#   1. ARM64 OSRM image build et (kaynaktan, ~20-30 dk)
#   2. Turkey PBF indir (mevcut ise atla)
#   3. extract → partition → customize

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$SCRIPT_DIR"
PBF="$DATA_DIR/turkey-latest.osm.pbf"
IMAGE="yol-asistani-osrm"

# ── Disk kontrolü ──────────────────────────────────────────────────────────
echo "==> Disk durumu:"
df -h "$DATA_DIR"
echo ""

FREE_KB=$(df "$DATA_DIR" | awk 'NR==2 {print $4}')
if (( FREE_KB < 2500000 )); then
  echo "UYARI: Build için ~2.5 GB boş alan gerekli, şu an az görünüyor."
  echo "       Devam etmeden önce Docker cache temizle:"
  echo "           docker system prune -f"
  echo ""
  echo "Devam etmek için Enter, iptal için Ctrl+C:"
  read -r
fi

# ── 1. ARM64 image build ───────────────────────────────────────────────────
echo "==> OSRM ARM64 image build ediliyor (kaynaktan, ~20-30 dk)..."
echo "    Coffee time ☕"
docker build -f "$PROJECT_DIR/Dockerfile.osrm" -t "$IMAGE" "$PROJECT_DIR"
echo "==> Image hazır: $IMAGE"
echo ""

# ── 2. PBF indir ──────────────────────────────────────────────────────────
if [[ ! -f "$PBF" ]]; then
  echo "==> Türkiye OSM verisi indiriliyor (~600 MB)..."
  wget -O "$PBF" https://download.geofabrik.de/europe/turkey-latest.osm.pbf
else
  echo "==> PBF mevcut: $(du -sh "$PBF" | cut -f1) — indirme atlandı."
fi
echo ""

# ── 3. Preprocessing ──────────────────────────────────────────────────────
# NOT: Docker image ENTRYPOINT=osrm-routed olduğu için
#      preprocessing araçları --entrypoint ile override ediliyor.

echo "==> osrm-extract (motorcycle profili)..."
docker run --rm -t --entrypoint osrm-extract \
  -v "$DATA_DIR:/data" "$IMAGE" \
  -p /data/motorcycle.lua /data/turkey-latest.osm.pbf

echo "==> osrm-partition..."
docker run --rm -t --entrypoint osrm-partition \
  -v "$DATA_DIR:/data" "$IMAGE" \
  /data/turkey-latest.osrm

echo "==> osrm-customize..."
docker run --rm -t --entrypoint osrm-customize \
  -v "$DATA_DIR:/data" "$IMAGE" \
  /data/turkey-latest.osrm

# ── Bitti ─────────────────────────────────────────────────────────────────
echo ""
echo "==> Preprocessing tamamlandı!"
echo ""
echo "Build cache temizle (~2 GB kazanır):"
echo "    docker builder prune -f"
echo ""
echo "Servisleri başlat:"
echo "    docker compose up -d"
echo ""
echo "OSRM smoke test:"
echo "    curl 'http://localhost:5000/route/v1/driving/32.85,39.92;27.14,38.41?overview=false'"
