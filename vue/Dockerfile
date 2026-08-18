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
