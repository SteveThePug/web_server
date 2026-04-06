#!/bin/sh
set -e

envsubst '${DOMAIN}' \
  < /quartz/quartz.config.ts.template \
  > /quartz/quartz.config.ts

exec npx quartz build --serve --port "${QUARTZ_PORT:-8080}"
