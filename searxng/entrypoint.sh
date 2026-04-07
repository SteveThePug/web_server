#!/bin/sh
set -e

envsubst < /searxng/settings.yml.template > /etc/searxng/settings.yml

exec /usr/local/searxng/dockerfiles/docker-entrypoint.sh "$@"
