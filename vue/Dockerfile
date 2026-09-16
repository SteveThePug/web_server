# =============================================================================
# Vue frontend image — compiles Rust to WASM, then builds the SPA.
#
# THIS IS A BUILD CONTAINER, NOT A SERVER. In production the container runs
# `vite build` into /output (the shared vue_dist volume) and EXITS. nginx then
# serves those files. Seeing `vue` as Exited(0) in `docker compose ps` is the
# success case, not a failure. The dev override replaces the CMD with the Vite
# dev server, and then it does stay running.
#
# STAGE 1 — WASM
#   Builds crates/stp_wasm with wasm-pack into src/wasm. Only that crate
#   directory is copied in, so touching Vue source never invalidates the (slow)
#   Rust layer. The two cache mounts persist the cargo registry and the Rust
#   target directory between builds on the host — on a Raspberry Pi this turns
#   a very long compile into an incremental one. src/wasm is gitignored: it is
#   generated, never committed.
#
# STAGE 2 — Vue
#   package.json + package-lock.json are copied first so `npm ci` is cached
#   independently of source changes; the npm cache is also a BuildKit cache
#   mount. `npm ci` (not `npm install`) installs exactly the lockfile, so a
#   rebuild can never silently pick up a different dependency version.
#   `make` and `git` are installed because some npm dependencies build native
#   code or resolve git URLs during install.
#
# --emptyOutDir is required: /output is a mounted volume, and vite refuses to
#   clear a directory outside the project root without it. Without the flag,
#   stale assets from a previous build would accumulate and be served.
#
# Note stage 2 is not named and is the final stage; both base images are
# multi-arch so this builds on the Pi unchanged.
# =============================================================================
# Stage 1: Build WASM from Rust
FROM rust:slim AS wasm-builder
RUN rustup target add wasm32-unknown-unknown \
    && cargo install wasm-pack
WORKDIR /wasm
COPY crates/stp_wasm/ crates/stp_wasm/
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/wasm/crates/stp_wasm/target \
    wasm-pack build crates/stp_wasm --target web --out-dir ../../src/wasm

# Stage 2: Build Vue frontend
FROM node:22-slim
RUN apt-get update && apt-get install -y make git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY . .
COPY --from=wasm-builder /wasm/src/wasm/ src/wasm/
CMD ["sh", "-c", "npx vite build --outDir /output --emptyOutDir"]
