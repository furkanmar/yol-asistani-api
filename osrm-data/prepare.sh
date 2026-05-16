#!/usr/bin/env bash
# OSRM Türkiye harita preprocessing — Pi native (Docker yok)
# Çalıştır: bash osrm-data/prepare.sh  (yol-asistani-api/ dizininden)
#
# 1. osrm-backend kur (apt veya kaynaktan)
# 2. Turkey PBF indir (mevcut ise atla)
# 3. extract → partition → customize
# 4. Systemd servisi kur

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR"
PBF="$DATA_DIR/turkey-latest.osm.pbf"
OSRM="$DATA_DIR/turkey-latest.osrm"

# ── 1. OSRM kurulum ────────────────────────────────────────────────────────
if command -v osrm-extract &>/dev/null; then
  echo "==> osrm-backend zaten kurulu: $(osrm-extract --version 2>&1 | head -1)"
else
  echo "==> osrm-backend kuruluyor..."
  sudo apt-get update -qq

  if sudo apt-get install -y osrm-backend 2>/dev/null; then
    echo "==> apt ile kuruldu."
  else
    echo "==> apt paketi yok — kaynaktan derleniyor (~20-30 dk)..."
    sudo apt-get install -y --no-install-recommends \
      build-essential cmake git pkg-config \
      libbz2-dev libxml2-dev libzip-dev zlib1g-dev \
      libboost-filesystem-dev libboost-iostreams-dev \
      libboost-regex-dev libboost-thread-dev \
      libboost-date-time-dev libboost-program-options-dev \
      libboost-system-dev libboost-test-dev libtbb-dev \
      lua5.4 liblua5.4-dev \
      libprotobuf-dev protobuf-compiler ca-certificates

    TMP=$(mktemp -d)
    git clone --depth 1 --branch v5.27.1 \
      https://github.com/Project-OSRM/osrm-backend.git "$TMP/osrm-src"

    cmake -S "$TMP/osrm-src" -B "$TMP/build" \
      -DCMAKE_BUILD_TYPE=Release \
      -DENABLE_MASON=OFF \
      -DENABLE_NODE_BINDINGS=OFF \
      -DENABLE_UNIT_TESTS=OFF

    cmake --build "$TMP/build" --parallel "$(nproc)"
    # cmake --install tüm hedefleri arar; sadece ihtiyacımız olanları kopyala
    sudo cp "$TMP/build/osrm-extract"   /usr/local/bin/
    sudo cp "$TMP/build/osrm-partition" /usr/local/bin/
    sudo cp "$TMP/build/osrm-customize" /usr/local/bin/
    sudo cp "$TMP/build/osrm-routed"    /usr/local/bin/
    rm -rf "$TMP"
    echo "==> Kaynaktan derleme tamamlandı."
  fi
fi

# ── 2. PBF indir ──────────────────────────────────────────────────────────
echo ""
echo "==> Disk durumu:"
df -h "$DATA_DIR"
echo ""

if [[ ! -f "$PBF" ]]; then
  echo "==> Türkiye OSM verisi indiriliyor (~600 MB)..."
  wget -O "$PBF" https://download.geofabrik.de/europe/turkey-latest.osm.pbf
else
  echo "==> PBF mevcut: $(du -sh "$PBF" | cut -f1) — indirme atlandı."
fi

# ── 3. Preprocessing ──────────────────────────────────────────────────────
echo ""
echo "==> osrm-extract (motorcycle profili)..."
osrm-extract -p "$DATA_DIR/motorcycle.lua" "$PBF"

echo "==> osrm-partition..."
osrm-partition "$OSRM"

echo "==> osrm-customize..."
osrm-customize "$OSRM"

# ── 4. Systemd servisi ────────────────────────────────────────────────────
SERVICE_FILE="/etc/systemd/system/osrm.service"

echo ""
echo "==> Systemd servisi kuruluyor: $SERVICE_FILE"

sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=OSRM Routing Engine
After=network.target

[Service]
ExecStart=/usr/bin/osrm-routed --algorithm mld $OSRM --max-table-size 10000
Restart=always
RestartSec=5
User=$USER

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable osrm
sudo systemctl restart osrm

sleep 2
if systemctl is-active --quiet osrm; then
  echo "==> OSRM servisi aktif."
else
  echo "HATA: OSRM servisi başlamadı. Log: sudo journalctl -u osrm -n 20"
  exit 1
fi

# ── Bitti ─────────────────────────────────────────────────────────────────
echo ""
echo "==> Tamamlandı!"
echo ""
echo "Smoke test:"
echo "    curl 'http://localhost:5000/route/v1/driving/32.85,39.92;27.14,38.41?overview=false'"
echo ""
echo "API'yi yeniden başlat (config değişti):"
echo "    docker compose up -d --build api"
