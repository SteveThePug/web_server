#!/bin/sh
set -e

# Check if certificate exists
if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
    echo "Certificates found. Using production nginx config."
    envsubst '$BACKEND_HOST $BACKEND_PORT $DOMAIN $ICECAST_PORT $ICECAST_HOST' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf
else
    echo "Certificates NOT found. Using setup nginx config."
    envsubst '$BACKEND_HOST $BACKEND_PORT $DOMAIN' < /etc/nginx/nginx_setup.conf.template > /etc/nginx/nginx.conf
fi

# Start nginx
nginx -g 'daemon off;'
