#!/bin/sh
# =============================================================================
# certbot entrypoint — obtains the TLS certificate once, then renews forever.
#
# WHAT IT DOES
#   1. One blocking `certonly` run to get (or confirm) a certificate for the
#      apex domain and its www alias.
#   2. Then loops forever: `certbot renew` every 12 hours.
#
# WHEN IT RUNS
#   Production only. docker-compose.dev.yml puts this service in the `disabled`
#   profile, so `docker compose -f docker-compose.yml -f docker-compose.dev.yml
#   up` never starts it and dev uses nginx's self-signed localhost cert instead.
#
# HOW VALIDATION WORKS: DNS-01 (--dns-cloudflare) when CLOUDFLARE_DNS_API_TOKEN
# is set, otherwise HTTP-01 (--webroot)
#   DNS-01: certbot creates a _acme-challenge TXT record through the Cloudflare
#   API, Let's Encrypt looks it up, and certbot deletes it. Nothing has to
#   reach the Pi, which is the point: with the site behind the Cloudflare
#   tunnel, ports 80 and 443 need not be forwarded at all. The token needs
#   Zone -> DNS -> Edit on this zone only. certbot insists its credentials
#   arrive as a mode-600 ini file, not an environment variable, so the script
#   writes one to /tmp on every start; nothing is stored on disk outside the
#   container.
#
#   Switching method on an existing certificate: certbot records the method
#   in renewal/<DOMAIN>.conf and `renew` would keep using webroot. Flags on
#   the command line override the stored ones, and the next successful
#   renewal saves them, so the method flags are passed to BOTH commands below.
#   An unset token after a switch therefore puts it back on webroot, too.
#
#   HTTP-01, the fallback: certbot writes a challenge token into certbot writes a challenge token into
#   /var/www/certbot/.well-known/acme-challenge/ and Let's Encrypt fetches it
#   over plain HTTP on port 80. certbot itself listens on nothing — nginx is
#   the server that answers. /var/www/certbot is the ./certbot/www bind mount,
#   mounted into BOTH containers, which is what makes this work. Every nginx
#   template (including the bootstrap nginx_setup.conf) serves that path.
#   See nginx/entrypoint.sh for the first-issuance chicken-and-egg ordering.
#
# CERT STORAGE / SHARING
#   /etc/letsencrypt is the ./certbot/conf bind mount, also mounted read-write
#   into nginx. That directory is the private key material — it is gitignored
#   and must never be committed.
#
# RENEWAL TIMING AND THE RELOAD SIGNAL
#   `certbot renew` is a no-op unless the cert is within 30 days of expiry, so
#   running it every 12h is safe and well inside Let's Encrypt's rate limits —
#   it just means the renewal window is checked twice a day rather than once a
#   month. Certificates are valid 90 days, so a renewal lands ~day 60.
#
#   nginx loads the certificate files at startup and keeps them in memory, so a
#   renewed certificate is not served until nginx is reloaded. This container
#   cannot signal nginx directly (no docker socket, no shared PID namespace),
#   so --deploy-hook instead TOUCHES A SENTINEL FILE in /etc/letsencrypt, which
#   is bind-mounted read-write into BOTH containers. A background watcher in
#   nginx/entrypoint.sh polls for that file, deletes it and runs `nginx -s
#   reload`. certbot only fires a deploy hook when a certificate was actually
#   renewed, so the sentinel appears roughly once every 60 days, not twice a
#   day. --deploy-hook is passed to both `certonly` and `renew`.
#
#   The alternative — giving this container the docker socket so the hook could
#   `docker exec nginx nginx -s reload` — was rejected: the socket is
#   root-equivalent on the host, and the shared cert directory is a capability
#   both containers already have.
#
# ENV VARS (from ./.env via `env_file`)
#   DOMAIN  apex domain; both $DOMAIN and www.$DOMAIN go on one certificate,
#           filed under --cert-name $DOMAIN so the path nginx expects
#           (/etc/letsencrypt/live/$DOMAIN/) is stable regardless of which
#           name certbot would otherwise have chosen.
#   EMAIL   registration/expiry-notice address for the ACME account.
#   CLOUDFLARE_DNS_API_TOKEN
#           optional; selects DNS-01. See HOW VALIDATION WORKS.
#
# FLAGS WORTH KNOWING
#   --agree-tos --non-interactive  never prompt; required in a container.
#   --expand                       if a cert already exists under this name but
#                                  covers a different set of names, replace it
#                                  rather than erroring out.
#
# ASSUMES EXISTS
#   /etc/letsencrypt  (persistent cert store, shared with nginx)
#   Webroot method only: /var/www/certbot (shared ACME webroot, served by
#   nginx on port 80) and port 80 reachable from the public internet for
#   $DOMAIN and www.$DOMAIN.
# =============================================================================

# The reload sentinel: a file in the SHARED /etc/letsencrypt mount that nginx's
# entrypoint polls for. Touching it is the entire "tell nginx to reload" signal
# — see RELOAD SIGNAL above.
RELOAD_SENTINEL=/etc/letsencrypt/.nginx-reload

if [ -n "${CLOUDFLARE_DNS_API_TOKEN}" ]; then
    CREDENTIALS=/tmp/cloudflare.ini
    # umask rather than a chmod afterwards, so the token is never readable
    # by anyone else even for an instant.
    (umask 077 && printf 'dns_cloudflare_api_token = %s\n' "${CLOUDFLARE_DNS_API_TOKEN}" >"${CREDENTIALS}")
    # 30s rather than the plugin's 10: Let's Encrypt queries from several
    # vantage points, and a lookup that lands before Cloudflare has the record
    # everywhere fails the whole order.
    set -- --dns-cloudflare --dns-cloudflare-credentials "${CREDENTIALS}" \
        --dns-cloudflare-propagation-seconds 30
    echo "Validating over DNS (Cloudflare API)."
else
    set -- --webroot -w /var/www/certbot
    echo "Validating over HTTP (webroot); needs port 80 reachable."
fi

certbot certonly "$@" \
    --email "${EMAIL}" \
    -d "${DOMAIN}" -d "www.${DOMAIN}" \
    --cert-name "${DOMAIN}" \
    --deploy-hook "touch ${RELOAD_SENTINEL}" \
    --agree-tos --non-interactive --expand;

# Exit promptly on `docker compose stop` instead of sitting out the 12h sleep.
trap exit TERM;

while :; do
    certbot renew "$@" \
        --deploy-hook "touch ${RELOAD_SENTINEL}";
    sleep 12h;
done
