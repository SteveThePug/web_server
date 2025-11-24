#!/bin/sh
set -e

# Substitute environment variables into template
envsubst < /etc/icecast2/icecast.xml.template > /etc/icecast2/icecast.xml

echo "test"

# Run icecast with the generated config
exec icecast -c /etc/icecast2/icecast.xml
