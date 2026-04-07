#!/bin/sh
set -e

sed -e "s|\${BASE_URL}|${BASE_URL}|g" \
    -e "s|\${SEARXNG_SECRET_KEY}|${SEARXNG_SECRET_KEY}|g" \
    /searxng/settings.yml.template > /etc/searxng/settings.yml

exec /usr/local/searxng/entrypoint.sh "$@"
