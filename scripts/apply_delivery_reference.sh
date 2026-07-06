#!/bin/bash
# Apply pre-exported Delivery reference SQL to production SQLite (no live PG sync).
set -euo pipefail

APP_ROOT="${APP_ROOT:-/opt/gallup-q14}"
DB_PATH="${DB_PATH:-$APP_ROOT/data/gallup-q14.db}"
SEED_SQL="${1:-$APP_ROOT/scripts/delivery_reference_seed.sql}"

if [[ ! -f "$SEED_SQL" ]]; then
  echo "Seed SQL not found: $SEED_SQL" >&2
  exit 1
fi

if [[ ! -f "$DB_PATH" ]]; then
  echo "Database not found: $DB_PATH" >&2
  exit 1
fi

command -v sqlite3 >/dev/null || { echo "sqlite3 required" >&2; exit 1; }

BACKUP="$APP_ROOT/backups/gallup-pre-delivery-seed-$(date +%Y%m%d-%H%M%S).db"
mkdir -p "$APP_ROOT/backups"
cp "$DB_PATH" "$BACKUP"
echo "Backup: $BACKUP"

systemctl stop gallup-q14
sqlite3 "$DB_PATH" < "$SEED_SQL"
chown gallup:gallup "$DB_PATH"
systemctl start gallup-q14

echo "Applied Delivery reference from $SEED_SQL"
sqlite3 "$DB_PATH" "SELECT 'org_options', COUNT(*) FROM delivery_org_options UNION ALL SELECT 'context_stats', COUNT(*) FROM delivery_context_stats UNION ALL SELECT 'sync_meta', COUNT(*) FROM delivery_sync_meta;"
