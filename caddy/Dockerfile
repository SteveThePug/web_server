# =============================================================================
# caddy image — the public edge. Caddy with two plugins, plus this
# directory's config, checked while the image is built.
#
# WHY A CUSTOM BUILD
#   Stock Caddy has neither of these, and the config needs both:
#     caddy-dns/cloudflare     DNS-01 certificates through the Cloudflare API
#                              (sites/prod.caddy), since the router forwards
#                              nothing to obtain them over HTTP.
#     mholt/caddy-ratelimit    the `rate_limit` directive (the `limit` snippet
#                              in the Caddyfile).
#   xcaddy compiles them in. The cloudflare plugin is pinned to a release;
#   caddy-ratelimit has had one tagged release (v0.1.0, 2024) and is pinned
#   to a commit, so Renovate leaves it alone and it is bumped by hand.
#
# BASE: caddy:2.11.9, the builder and the runtime from the same release
#   (the builder's xcaddy builds the CADDY_VERSION its image sets). Both
#   tags are bumped together. Official images are multi-arch.
#
# THE CONFIG IS VALIDATED AT BUILD TIME
#   The last RUN parses and provisions the config with `caddy validate`, for
#   the production sites and the dev ones, against placeholder values for
#   every variable the config reads. A config that would not load therefore
#   fails the image build, and a failed build aborts `docker compose up
#   --build` BEFORE any container is replaced, so the running edge keeps
#   serving. `make check-caddy` is the same build. The placeholders exist
#   only for that RUN; the real values come from the container's
#   environment (docker-compose.yml).
#
# NO ENTRYPOINT. The base image's CMD runs
#   `caddy run --config /etc/caddy/Caddyfile --adapter caddyfile`, which is all
#   that is needed: Caddy reads the environment itself, and obtains and renews
#   its own certificates.
# =============================================================================
FROM caddy:2.11.9-builder AS builder
RUN xcaddy build \
    --with github.com/caddy-dns/cloudflare@v0.2.4 \
    --with github.com/mholt/caddy-ratelimit@5625512f24f6f59d6f64fb3aafe5eecff0b286db

FROM caddy:2.11.9
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile routes.caddy /etc/caddy/
COPY sites/ /etc/caddy/sites/
# Mount points (docker-compose.yml). Created so Caddy still starts without
# them: the log directory especially, since a log file that cannot be opened
# stops the config loading.
RUN mkdir -p /var/log/caddy /srv/www /srv/uploads
# Validation writes Caddy's data and the payload log, so it is pointed at a
# throwaway directory and the log is removed afterwards, leaving the image as
# it was.
RUN set -e; \
    for site in prod dev; do \
        echo "caddy validate (CADDY_SITE=$site)"; \
        env XDG_DATA_HOME=/tmp/validate XDG_CONFIG_HOME=/tmp/validate \
            CADDY_SITE=$site DOMAIN=example.com EMAIL=validate@example.com \
            BACKEND_HOST=backend BACKEND_PORT=8080 BACKEND_ENDPOINT=/api \
            ICECAST_HOST=icecast2 ICECAST_PORT=8000 \
            GITEA_HOST=gitea GITEA_PORT=3000 \
            SILVERBULLET_HOST=silverbullet SILVERBULLET_PORT=3000 \
            SILVERBULLET_PUBLIC_HOST=silverbullet-public SILVERBULLET_PUBLIC_PORT=3000 \
            PYTHON_HOST=python PYTHON_PORT=8000 \
            RECEIPT_WEB_UPSTREAM=127.0.0.1:9 RECEIPT_WEB_TOKEN=validate \
            CLOUDFLARE_DNS_API_TOKEN=validate-placeholder-token-0000000000000 \
            caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile; \
    done; \
    rm -rf /tmp/validate /var/log/caddy/*
