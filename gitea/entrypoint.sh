#!/bin/sh
set -e

# Generate app.ini from template if it doesn't already exist
if [ ! -f /etc/gitea/app.ini ]; then
    cp /etc/gitea/app.ini.template /etc/gitea/app.ini
    echo "Generated app.ini from template"
fi

exec /usr/local/bin/docker-entrypoint.sh "$@"
