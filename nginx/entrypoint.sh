#!/bin/sh
set -e

# Check if DEV_MODE
if [ "$DEV_MODE" = "true" ]; then
  echo "Dev mode. Generating self-signed certificate for HTTPS."
  CERT_DIR="/etc/letsencrypt/live/localhost"
  if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
    mkdir -p "$CERT_DIR"
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
      -keyout "$CERT_DIR/privkey.pem" \
      -out "$CERT_DIR/fullchain.pem" \
      -subj "/CN=localhost" 2>/dev/null
  fi
  # In dev mode, so use nginx_dev.conf.template
  envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT} ${HASURA_HOST} ${HASURA_PORT} ${QUARTZ_HOST} ${QUARTZ_PORT} ${UPTIMEKUMA_HOST} ${UPTIMEKUMA_PORT} ${SEARXNG_HOST} ${SEARXNG_PORT} ${WALLABAG_HOST} ${WALLABAG_PORT} ${OPENWEBUI_HOST} ${OPENWEBUI_PORT}' \
    </etc/nginx/nginx_dev.conf.template \
    >/etc/nginx/nginx.conf
elif [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
  echo "Certificates found. Using production nginx config."
  # In production with certificates already existing, so use nginx.conf.template
  envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT} ${HASURA_HOST} ${HASURA_PORT} ${QUARTZ_HOST} ${QUARTZ_PORT} ${UPTIMEKUMA_HOST} ${UPTIMEKUMA_PORT} ${SEARXNG_HOST} ${SEARXNG_PORT} ${WALLABAG_HOST} ${WALLABAG_PORT} ${OPENWEBUI_HOST} ${OPENWEBUI_PORT}' \
    </etc/nginx/nginx.conf.template \
    >/etc/nginx/nginx.conf
else
  echo "Certificates NOT found. Using setup nginx config."
  # In production with no certificates, so use nginx_setup.conf.template and will need restart after generation
  envsubst '${DOMAIN}' </etc/nginx/nginx_setup.conf.template >/etc/nginx/nginx.conf
fi

# Ensure upload directory is traversable by nginx worker
chmod 755 /uploads 2>/dev/null || true

# Wait for Vue assets in production mode
if [ "$DEV_MODE" != "true" ]; then
  echo "Waiting for Vue assets..."
  elapsed=0
  while [ ! -f /etc/nginx/html/index.html ] && [ $elapsed -lt 120 ]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done
  if [ ! -f /etc/nginx/html/index.html ]; then
    echo "WARNING: Vue assets not found after 120s, starting nginx anyway"
  else
    echo "Vue assets ready."
  fi
fi

# Start nginx
nginx -g 'daemon off;'
