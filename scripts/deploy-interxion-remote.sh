#!/bin/bash
set -euo pipefail

ADMIN_PASSWORD="${ADMIN_PASSWORD:?ADMIN_PASSWORD required}"
VPS_HOST="${VPS_HOST:?VPS_HOST required}"
APP_ROOT=/opt/gallup-q14
SRC="$APP_ROOT/src"

mkdir -p "$SRC" "$APP_ROOT/bin" "$APP_ROOT/data" "$APP_ROOT/frontend" "$APP_ROOT/backups" "$APP_ROOT/logs"
rm -rf "$SRC"/*
tar -xzf /tmp/deploy-src.tgz -C "$SRC"

# Build
cd "$SRC/backend"
export PATH="/usr/lib/go-1.24/bin:/usr/bin:$PATH"
go mod download
CGO_ENABLED=0 go build -o "$APP_ROOT/bin/gallup-q14" ./cmd/server
chmod +x "$APP_ROOT/bin/gallup-q14"

# Frontend
rm -rf "$APP_ROOT/frontend/dist"
mkdir -p "$APP_ROOT/frontend"
cp -a "$SRC/frontend/dist" "$APP_ROOT/frontend/"

# Env
cat > "$APP_ROOT/.env" <<EOF
ADMIN_PASSWORD=${ADMIN_PASSWORD}
PORT=8080
DB_PATH=${APP_ROOT}/data/gallup-q14.db
CORS_ORIGIN=https://${VPS_HOST}:8443
DELIVERY_SYNC_INTERVAL_HOURS=0
EOF
chmod 600 "$APP_ROOT/.env"

chown -R gallup:gallup "$APP_ROOT"

# TLS self-signed
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/gallup-q14.key \
  -out /etc/ssl/certs/gallup-q14.crt \
  -subj "/CN=${VPS_HOST}" 2>/dev/null

cat > /etc/nginx/sites-available/gallup-q14 <<NGINX
server {
    listen 8443 ssl;
    listen [::]:8443 ssl;
    server_name ${VPS_HOST};

    ssl_certificate     /etc/ssl/certs/gallup-q14.crt;
    ssl_certificate_key /etc/ssl/private/gallup-q14.key;

    client_max_body_size 2m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGINX

ln -sf /etc/nginx/sites-available/gallup-q14 /etc/nginx/sites-enabled/gallup-q14
rm -f /etc/nginx/sites-enabled/default
nginx -t

cat > /etc/systemd/system/gallup-q14.service <<'UNIT'
[Unit]
Description=Gallup Q14 engagement survey
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=gallup
Group=gallup
WorkingDirectory=/opt/gallup-q14/bin
EnvironmentFile=/opt/gallup-q14/.env
ExecStart=/opt/gallup-q14/bin/gallup-q14
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable gallup-q14
systemctl restart gallup-q14
systemctl reload nginx

sleep 2
systemctl is-active gallup-q14
systemctl is-active nginx
curl -sk https://127.0.0.1:8443/api/health
echo
curl -s http://127.0.0.1:8080/api/health
echo
