# =============================================================================
# Go backend image — multi-stage: build with the full toolchain, ship a slim
# runtime containing one static binary.
#
# STAGE 1 (golang:1.26)
#   go.mod/go.sum are copied and downloaded BEFORE the rest of the source, so
#   the dependency layer is cached and only re-runs when the module files
#   change — a meaningful saving when building on a Raspberry Pi.
#
#   The `--mount=type=cache` lines are BuildKit cache mounts: the module cache
#   and the Go build cache persist BETWEEN builds on the host without being
#   baked into any image layer. This is the single biggest reason a rebuild on
#   the Pi takes minutes rather than much longer. They require BuildKit, which
#   is the default in modern `docker compose build`.
#
#   CGO_ENABLED=0 produces a fully static binary with no libc dependency, which
#   is what lets a golang-built binary run on a different, slimmer base image.
#
# STAGE 2 (debian:bookworm-slim)
#   Carries only the binary plus ca-certificates — needed because the backend
#   makes outbound HTTPS calls (Spotify, Claude, Steam, the Gitea feed) and
#   would fail certificate verification without a trust store.
#
# EXPOSE 8080 is documentation only; the port is never published to the host.
# nginx reaches it over app-network as BACKEND_HOST:BACKEND_PORT.
#
# Base images are multi-arch, so this builds unchanged on arm64 (the Pi) and
# amd64 (a dev laptop) with no platform pin.
# =============================================================================
FROM golang:1.26 AS builder

WORKDIR /backend

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o main .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --yes ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /backend

COPY --from=builder /backend/main ./main

EXPOSE 8080

CMD ["./main"]
