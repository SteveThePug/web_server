#!/bin/sh
if [ ! -d /etc/letsencrypt/live/${DOMAIN} ]; then
    certbot certonly --webroot -w /var/www/certbot --email ${EMAIL} -d ${DOMAIN} -d www.${DOMAIN} --agree-tos --non-interactive;
fi;

trap exit TERM;

while :; do
    certbot renew --webroot -w /var/www/certbot;
    sleep 12h & wait $${!};
done
