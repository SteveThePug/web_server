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
