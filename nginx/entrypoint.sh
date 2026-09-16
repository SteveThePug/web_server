#!/bin/sh
# =============================================================================
# nginx entrypoint — picks one of three nginx configs at container start.
#
# WHAT IT DOES
#   nginx.conf is not shipped in the image. Instead three *templates* are baked
#   in (see nginx/Dockerfile) and this script renders exactly one of them to
#   /etc/nginx/nginx.conf before starting nginx in the foreground.
#
# WHEN IT RUNS
#   Every time the `nginx` container starts (ENTRYPOINT). It is therefore also
#   the thing that re-evaluates "do we have certificates yet?" on each restart.
#
# WHICH CONFIG IS CHOSEN (the three-way decision)
#   1. DEV_MODE=true                  -> nginx_dev.conf.template
#        Local development. Certbot is disabled (docker-compose.dev.yml puts it
#        in the `disabled` profile), so there can never be a real certificate.
#        We mint a throwaway self-signed cert for CN=localhost so the :443
#        server block still has something to load, and proxy `/` to the Vite
#        dev server (vue:5173) instead of serving built files from disk.
#   2. Real cert exists for $DOMAIN   -> nginx.conf.template
#        Steady-state production: HTTP->HTTPS redirect, apex->www redirect,
#        SPA served from /etc/nginx/html, everything else reverse-proxied.
#   3. No cert yet                    -> nginx_setup.conf.template
#        The bootstrap case. See CHICKEN-AND-EGG below.
#
# CHICKEN-AND-EGG: nginx_setup.conf vs certbot first issuance
#   Let's Encrypt HTTP-01 validation requires a working HTTP server on port 80
#   that serves /.well-known/acme-challenge/. But nginx.conf.template references
#   /etc/letsencrypt/live/$DOMAIN/fullchain.pem in its `ssl_certificate`
#   directives, and nginx refuses to start if that file is missing. So on a
#   brand-new host nginx cannot use the production config, and certbot cannot
#   get a certificate without nginx. The setup config breaks the loop: it is a
#   minimal port-80-only server that serves the ACME webroot and 404s
#   everything else — no TLS, no upstreams, nothing that can fail to load.
#   Order of events on a fresh machine:
#     a. nginx starts with nginx_setup.conf (port 80, ACME webroot only).
#     b. certbot container runs `certbot certonly --webroot`, writes the cert
#        into the shared ./certbot/conf volume.
#     c. nginx is restarted (manually, or by a redeploy) — this script re-runs,
#        now finds the cert, and renders the production config.
#   Step (c) is NOT automatic. A first deploy on a new domain needs one manual
#   `docker compose restart nginx` after certbot succeeds.
#
# THE envsubst TEMPLATING PATTERN
#   The templates are plain nginx configs containing $VARIABLES. Two different
#   kinds of $VAR live in the same file:
#     - build-time vars we want substituted:  $DOMAIN, $BACKEND_HOST, ...
#     - runtime nginx vars that must survive: $host, $remote_addr, $uri,
#       $http_upgrade, $proxy_add_x_forwarded_for, $binary_remote_addr, ...
#   Bare `envsubst` would replace *every* $NAME it recognises and blank out the
#   nginx runtime variables, producing a config that silently forwards empty
#   headers. Passing an explicit SHELL-FORMAT list (the '${A} ${B} ...' string
#   below) restricts envsubst to exactly those names and leaves every other
#   $token untouched. That is why the list is in SINGLE quotes: it must reach
#   envsubst as a literal '${DOMAIN} ${BACKEND_HOST} ...' string, not be
#   expanded by the shell first.
#
# ENV VARS REQUIRED (all from ./.env via `env_file` in docker-compose.yml)
#   DOMAIN                         apex domain, e.g. example.com
#   BACKEND_HOST / BACKEND_PORT    Go API container name and port
#   BACKEND_ENDPOINT               URL prefix the API is mounted at (e.g. /api)
#   ICECAST_HOST / ICECAST_PORT    radio stream upstream
#   GITEA_HOST / GITEA_PORT        git server upstream
#   HASURA_HOST / HASURA_PORT      Hasura console upstream
#   QUARTZ_HOST / QUARTZ_PORT      notes site upstream
#   PYTHON_HOST / PYTHON_PORT      FastAPI upstream (defaulted below)
#   UPTIMEKUMA_* / WALLABAG_*      substituted but not currently referenced by
#                                  any template; kept so adding those services
#                                  back needs no entrypoint change.
#   DEV_MODE                       "true" only via docker-compose.dev.yml
#
# ASSUMES EXISTS
#   /etc/letsencrypt  (bind mount of ./certbot/conf, shared with certbot)
#   /var/www/certbot  (bind mount of ./certbot/www, the ACME webroot)
#   /uploads          (named volume `uploads`, shared with the backend)
#   /etc/nginx/html   (named volume `vue_dist`, written by the one-shot `vue`
#                     build container in production)
# =============================================================================
set -e

# Defaults for optional services so envsubst never leaves an empty upstream
export PYTHON_HOST="${PYTHON_HOST:-python}"
export PYTHON_PORT="${PYTHON_PORT:-8000}"

# The envsubst allow-list, defined once and reused by both full-config branches
# so the two lists cannot drift apart. It MUST stay single-quoted here and be
# passed as "$ENVSUBST_VARS" (one argument) below.
ENVSUBST_VARS='${DOMAIN} ${BACKEND_HOST} ${BACKEND_PORT} ${BACKEND_ENDPOINT} ${ICECAST_HOST} ${ICECAST_PORT} ${GITEA_HOST} ${GITEA_PORT} ${HASURA_HOST} ${HASURA_PORT} ${QUARTZ_HOST} ${QUARTZ_PORT} ${UPTIMEKUMA_HOST} ${UPTIMEKUMA_PORT} ${WALLABAG_HOST} ${WALLABAG_PORT} ${PYTHON_HOST} ${PYTHON_PORT}'

# Check if DEV_MODE
if [ "$DEV_MODE" = "true" ]; then
  echo "Dev mode. Generating self-signed certificate for HTTPS."
  CERT_DIR="/etc/letsencrypt/live/localhost"
  # Generated once and kept in the certbot/conf bind mount, so restarts reuse
  # the same cert and the browser's "accept this exception" sticks.
  if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
    mkdir -p "$CERT_DIR"
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
      -keyout "$CERT_DIR/privkey.pem" \
      -out "$CERT_DIR/fullchain.pem" \
      -subj "/CN=localhost" 2>/dev/null
  fi
  # In dev mode, so use nginx_dev.conf.template
  envsubst "$ENVSUBST_VARS" \
    </etc/nginx/nginx_dev.conf.template \
    >/etc/nginx/nginx.conf
elif [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
  echo "Certificates found. Using production nginx config."
  # In production with certificates already existing, so use nginx.conf.template
  envsubst "$ENVSUBST_VARS" \
    </etc/nginx/nginx.conf.template \
    >/etc/nginx/nginx.conf
else
  echo "Certificates NOT found. Using setup nginx config."
  # In production with no certificates, so use nginx_setup.conf.template and will need restart after generation
  # Only ${DOMAIN} is substituted: the setup config has no upstreams at all.
  envsubst '${DOMAIN}' </etc/nginx/nginx_setup.conf.template >/etc/nginx/nginx.conf
fi

# Ensure upload directory is traversable by nginx worker.
# `uploads` is a named volume shared with the backend, which creates it with
# the backend's umask; nginx runs its workers as a different user and needs
# +x on the directory to serve files out of it via the /uploads/ alias.
# `|| true` because on a read-only or foreign-owned mount this is not fatal.
chmod 755 /uploads 2>/dev/null || true

# Wait for Vue assets in production mode.
# The `vue` service is a ONE-SHOT build container: it runs `vite build` into the
# shared `vue_dist` volume and exits. `depends_on: vue` only waits for it to
# start, not to finish, so on a cold build nginx can come up before index.html
# exists and would 404 the whole site. Poll for up to 120s, then start anyway
# (a running nginx that 404s is still better than a crash loop, and the assets
# appear moments later). In dev mode there are no built assets at all — `/` is
# proxied to the Vite dev server — so the wait is skipped.
if [ "$DEV_MODE" != "true" ]; then
  echo "Waiting for Vue assets..."
  elapsed=0
  vue_build_timeout=120
  while [ ! -f /etc/nginx/html/index.html ] && [ "$elapsed" -lt "$vue_build_timeout" ]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done
  if [ ! -f /etc/nginx/html/index.html ]; then
    echo "WARNING: Vue assets not found after ${vue_build_timeout}s, starting nginx anyway"
  else
    echo "Vue assets ready."
  fi
fi

# Start nginx in the foreground so it is PID 1 and docker can signal/restart it.
nginx -g 'daemon off;'
