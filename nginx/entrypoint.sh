#!/bin/sh
set -e

# Check if dev mode, certificate exists, or setup mode
if [ "$DEV_MODE" = "true" ]; then
    echo "Dev mode. Using HTTP-only nginx config."
    envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT}' \
        < /etc/nginx/nginx_dev.conf.template \
        > /etc/nginx/nginx.conf
elif [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
    echo "Certificates found. Using production nginx config."
    envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT}' \
        < /etc/nginx/nginx.conf.template \
        > /etc/nginx/nginx.conf
else
    echo "Certificates NOT found. Using setup nginx config."
    envsubst '${DOMAIN}' < /etc/nginx/nginx_setup.conf.template > /etc/nginx/nginx.conf
fi

# Start nginx
nginx -g 'daemon off;'
