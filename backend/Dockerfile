FROM golang:1.25 AS builder

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
