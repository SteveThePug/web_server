#!/bin/sh
set -e

# Check if dev mode, certificate exists, or setup mode
if [ "$DEV_MODE" = "true" ]; then
  echo "Dev mode. Using HTTP-only nginx config."
  envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT}' \
    </etc/nginx/nginx_dev.conf.template \
    >/etc/nginx/nginx.conf
elif [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
  echo "Certificates found. Using production nginx config."
  envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT}' \
    </etc/nginx/nginx.conf.template \
    >/etc/nginx/nginx.conf
else
  echo "Certificates NOT found. Using setup nginx config."
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
