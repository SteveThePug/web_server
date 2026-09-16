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
# HOW VALIDATION WORKS (--webroot)
#   Let's Encrypt HTTP-01: certbot writes a challenge token into
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
# ENV VARS REQUIRED (from ./.env via `env_file`)
#   DOMAIN  apex domain; both $DOMAIN and www.$DOMAIN go on one certificate,
#           filed under --cert-name $DOMAIN so the path nginx expects
#           (/etc/letsencrypt/live/$DOMAIN/) is stable regardless of which
#           name certbot would otherwise have chosen.
#   EMAIL   registration/expiry-notice address for the ACME account.
#
# FLAGS WORTH KNOWING
#   --agree-tos --non-interactive  never prompt; required in a container.
#   --expand                       if a cert already exists under this name but
#                                  covers a different set of names, replace it
#                                  rather than erroring out.
#
# ASSUMES EXISTS
#   /var/www/certbot  (shared ACME webroot, served by nginx on port 80)
#   /etc/letsencrypt  (persistent cert store, shared with nginx)
#   Port 80 reachable from the public internet for $DOMAIN and www.$DOMAIN.
# =============================================================================

# The reload sentinel: a file in the SHARED /etc/letsencrypt mount that nginx's
# entrypoint polls for. Touching it is the entire "tell nginx to reload" signal
# — see RELOAD SIGNAL above.
RELOAD_SENTINEL=/etc/letsencrypt/.nginx-reload

certbot certonly --webroot -w /var/www/certbot \
    --email "${EMAIL}" \
    -d "${DOMAIN}" -d "www.${DOMAIN}" \
    --cert-name "${DOMAIN}" \
    --deploy-hook "touch ${RELOAD_SENTINEL}" \
    --agree-tos --non-interactive --expand;

# Exit promptly on `docker compose stop` instead of sitting out the 12h sleep.
trap exit TERM;

while :; do
    certbot renew --webroot -w /var/www/certbot \
        --deploy-hook "touch ${RELOAD_SENTINEL}";
    sleep 12h;
done
