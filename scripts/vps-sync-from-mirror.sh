#!/bin/bash
# Rebuild delivery reference tables from the local mirror on this host (no PostgreSQL).
set -euo pipefail

APP_ROOT="${APP_ROOT:-/opt/gallup-q14}"
export DB_PATH="${DB_PATH:-$APP_ROOT/data/gallup-q14.db}"
export DELIVERY_MIRROR_PATH="${DELIVERY_MIRROR_PATH:-$APP_ROOT/data/delivery_mirror.db}"

if [[ ! -f "$DELIVERY_MIRROR_PATH" ]]; then
  echo "Mirror not found: $DELIVERY_MIRROR_PATH" >&2
  exit 1
fi

python3 "$APP_ROOT/scripts/sync_delivery_reference.py"
