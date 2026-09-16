# =============================================================================
# Quartz image — builds the Obsidian notes site.
#
# HOW THE SOURCE IS OBTAINED: Quartz is not an npm package you depend on, it is
# an application you clone and configure. So the image `git clone`s the
# upstream repo at build time, which is why `git` is installed. --depth 1
# --branch pins to an exact tag (QUARTZ_VERSION, overridable as a build ARG)
# and fetches no history — a shallow clone is much faster on a Pi and the
# history is of no use here.
#
# Because this clone happens at build time, upstream Quartz is only updated
# when the image is rebuilt with a bumped QUARTZ_VERSION. The tag pin is the
# thing stopping an upstream change from breaking the notes site on a routine
# redeploy.
#
# gettext-base supplies `envsubst` for rendering quartz.config.ts from its
# template at container start (see quartz/entrypoint.sh).
#
# quartz.config.ts.template and quartz.layout.ts are copied in AFTER the clone
# so they land on top of the upstream defaults — layout is used as-is, config
# is templated at runtime because it needs DOMAIN.
#
# node:22-slim is multi-arch, so this builds unchanged on the Pi.
# =============================================================================
FROM node:22-slim

RUN apt-get update \
    && apt-get install --yes git gettext-base \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /quartz

ARG QUARTZ_VERSION=v4.4.0
RUN git clone --depth 1 --branch ${QUARTZ_VERSION} \
    https://github.com/jackyzha0/quartz.git . \
    && npm ci

COPY quartz.config.ts.template /quartz/quartz.config.ts.template
COPY quartz.layout.ts /quartz/quartz.layout.ts
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
