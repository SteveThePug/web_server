#!/bin/bash
set -e

envsubst < /etc/icecast2/icecast.xml.template > /etc/icecast2/icecast.xml
envsubst < /etc/liquidsoap/stream.liq.template > /etc/liquidsoap/stream.liq

icecast2 -c /etc/icecast2/icecast.xml &
sleep 2
liquidsoap /etc/liquidsoap/stream.liq &
wait -n
kill $(jobs -p) 2>/dev/null || true
exit 1
