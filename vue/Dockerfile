# =============================================================================
# Vue frontend image — builds the SPA.
#
# THIS IS A BUILD CONTAINER, NOT A SERVER. In production the container runs
# `vite build` into /output (the shared vue_dist volume) and EXITS. nginx then
# serves those files. Seeing `vue` as Exited(0) in `docker compose ps` is the
# success case, not a failure. The dev override replaces the CMD with the Vite
# dev server, and then it does stay running.
#
# package.json + package-lock.json are copied first so `npm ci` is cached
#   independently of source changes; the npm cache is also a BuildKit cache
#   mount. `npm ci` (not `npm install`) installs exactly the lockfile, so a
#   rebuild can never silently pick up a different dependency version.
#   `make` and `git` are installed because some npm dependencies build native
#   code or resolve git URLs during install.
#
# --emptyOutDir is required: /output is a mounted volume, and vite refuses to
#   clear a directory outside the project root without it. Without the flag,
#   stale assets from a previous build would accumulate and be served.
# =============================================================================
FROM node:22-slim
RUN apt-get update && apt-get install -y make git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY . .
CMD ["sh", "-c", "npx vite build --outDir /output --emptyOutDir"]
