#!/bin/bash
# =============================================================================
# icecast2 entrypoint — runs the streaming server AND its audio source together.
#
# WHAT IT DOES
#   This one container holds two cooperating daemons:
#     icecast2   - the HTTP streaming server listeners connect to. nginx
#                  proxies /radio/ to it.
#     liquidsoap - the *source*: it decides what audio to play and pushes a
#                  single MP3 stream into icecast's mount point.
#   They are colocated because liquidsoap connects to icecast over
#   host="localhost" (see stream.liq.template) — no network hop, and the source
#   password never leaves the container.
#
# WHEN IT RUNS
#   Container start (ENTRYPOINT), in both dev and production.
#
# THE TEMPLATING STEP
#   Both configs are rendered by bare `envsubst` (no allow-list) because unlike
#   the nginx templates, neither icecast.xml nor stream.liq has any legitimate
#   dollar-token of its own — every ${VAR} in them is meant to be substituted,
#   so an unrestricted envsubst is safe here. NOTE the consequence: any
#   referenced variable that is unset silently becomes an empty string, which
#   for a password field means "empty password" rather than an error.
#
# STARTUP ORDER AND THE sleep 2
#   liquidsoap's output.icecast connects to icecast as a client on startup and
#   there is no readiness check available, so the script backgrounds icecast,
#   waits a fixed 2 seconds for it to bind its port, then starts liquidsoap.
#   It is a race, not a guarantee — but `fallible=true` in stream.liq means
#   liquidsoap retries the connection rather than dying if it loses.
#
# PROCESS SUPERVISION
#   `wait -n` returns as soon as EITHER daemon exits. We then kill whatever is
#   still running and exit 1, so a half-dead container (icecast up, liquidsoap
#   dead = silence forever) becomes a full crash that docker's
#   `restart: always` will recycle. Exiting non-zero is deliberate.
#
# ENV VARS REQUIRED (from ./.env)
#   ICECAST_SOURCE_PASSWORD  shared by icecast (accepts) and liquidsoap (sends)
#   ICECAST_RELAY_PASSWORD   for relay clients
#   ICECAST_ADMIN_USER / ICECAST_ADMIN_PASSWORD   icecast web admin
#   ICECAST_HOST / ICECAST_PORT / ICECAST_MOUNT   listener-facing stream
#   LIQUIDSOAP_HARBOR_MOUNT / LIQUIDSOAP_HARBOR_PORT   live DJ input
#
# ASSUMES EXISTS
#   /music  - read-only bind mount of ./icecast2/fallback_music, the fallback
#             playlist. If it is empty the stream falls through to silence.
# =============================================================================
set -e

envsubst < /etc/icecast2/icecast.xml.template > /etc/icecast2/icecast.xml
envsubst < /etc/liquidsoap/stream.liq.template > /etc/liquidsoap/stream.liq

icecast2 -c /etc/icecast2/icecast.xml &
# Give icecast time to bind before liquidsoap tries to connect as a source.
sleep 2
liquidsoap /etc/liquidsoap/stream.liq &
# Exit as soon as either daemon dies, then take the other down with it.
wait -n
kill $(jobs -p) 2>/dev/null || true
exit 1
