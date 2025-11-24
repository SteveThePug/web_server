#!/bin/sh
set -e

# Check if certificate exists
if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
    echo "Certificates found. Using production nginx config."
    envsubst '${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${ICECAST_HOST} ${ICECAST_PORT}' \
        < /etc/nginx/nginx.conf.template \
        > /etc/nginx/nginx.conf
else
    echo "Certificates NOT found. Using setup nginx config."
    envsubst '${DOMAIN}' < /etc/nginx/nginx_setup.conf.template > /etc/nginx/nginx.conf
fi

# Start nginx
nginx -g 'daemon off;'
