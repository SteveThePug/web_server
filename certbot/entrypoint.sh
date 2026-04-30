#!/bin/sh
certbot certonly --webroot -w /var/www/certbot \
    --email ${EMAIL} \
    -d ${DOMAIN} -d www.${DOMAIN} -d chat.${DOMAIN} \
    --agree-tos --non-interactive --expand;

trap exit TERM;

while :; do
    certbot renew --webroot -w /var/www/certbot;
    sleep 12h;
done
