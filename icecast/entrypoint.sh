#!/bin/sh
set -e

# Substitute environment variables into template
envsubst < /etc/icecast2/icecast.xml.template > /etc/icecast2/icecast.xml
# envsubst < /etc/liquidsoap/stream.liq.template > /etc/liquidsoap/stream.liq

# Run icecast with the generated config
exec icecast2 -c /etc/icecast2/icecast.xml
# exec liquidsoap /etc/liquidsoap/stream.liq
# wait -n
