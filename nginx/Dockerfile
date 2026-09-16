# =============================================================================
# nginx image — the public edge. Stock nginx plus two tools and three templates.
#
# BASE: nginx:1.27, pinned to a minor version. Official nginx images are
#   multi-arch, so the same tag resolves to the arm64 build on the Raspberry Pi
#   with no platform pin needed — that is why there is no `--platform` anywhere
#   in this repo.
#
# WHY THE EXTRA PACKAGES
#   gettext-base  provides `envsubst`, which entrypoint.sh uses to render a
#                 config template. Not in the base image.
#   openssl       used by entrypoint.sh to mint the self-signed localhost
#                 certificate in DEV_MODE only. Installed unconditionally
#                 because there is one image for both dev and production.
#
# `rm -rf /etc/nginx/html/*` clears the stock "Welcome to nginx" page, since
#   that directory is where the vue_dist volume gets mounted and a leftover
#   index.html there would be served if the Vue build ever failed.
#
# apt lists are deleted in the SAME RUN layer so they never end up in the image.
#
# NO nginx.conf IS COPIED IN. All three templates are, and entrypoint.sh picks
# one at container start. See nginx/entrypoint.sh for that decision.
# =============================================================================
FROM nginx:1.27
RUN rm -rf /etc/nginx/html/* && \
    apt-get update && apt-get install -y gettext-base openssl && \
    rm -rf /var/lib/apt/lists/*
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template
COPY nginx_dev.conf.template /etc/nginx/nginx_dev.conf.template
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
