#!/bin/sh
# =============================================================================
# One-time setup of the `garage` container. Run from the repo root, on the
# host, once the container is up:
#
#     garage/bootstrap.sh
#
# Idempotent: every step checks before acting, so re-running it is harmless.
# Garage keeps all of this in the garage_meta volume, so it is only needed on
# a fresh volume — once on stpvm, and again whenever a dev stack starts from
# nothing.
#
# WHY A HOST SCRIPT
#   The dxflrs/garage image is FROM scratch: a single binary, no shell. There
#   is nothing to run an init script with inside it, so this drives the
#   binary through `docker compose exec`.
#
# WHAT IT DOES
#   1. Gives the single node a role in the cluster layout. Until it has one,
#      Garage accepts connections but refuses to store anything.
#   2. Creates the bucket (S3_BUCKET, default `uploads`).
#   3. Imports the backend's access key from .env (S3_ACCESS_KEY_ID /
#      S3_SECRET_ACCESS_KEY). Importing rather than letting Garage generate
#      one keeps .env the single source of truth: a wiped volume comes back
#      with the same credentials and the backend needs no change.
#   4. Grants that key read+write on the bucket. garage-webui also browses
#      with it: it borrows the first key that has access to a bucket.
#   5. Opens the bucket on the website endpoint, which is what nginx's
#      /uploads/ fallback reads from, anonymously.
#
# ENV / .env
#   S3_ACCESS_KEY_ID       required; "GK" + 24 hex chars
#   S3_SECRET_ACCESS_KEY   required; 64 hex chars
#   S3_BUCKET              optional, default uploads
#   GARAGE_CAPACITY        optional, default 20G. Only weights data placement
#                          between nodes; with one node it is no quota.
#   COMPOSE                optional, default `docker compose`. Set it to add
#                          the dev override, e.g.
#                          COMPOSE="docker compose -f docker-compose.yml -f docker-compose.dev.yml"
# =============================================================================
set -eu

COMPOSE=${COMPOSE:-docker compose}

# Read one KEY=value from .env without sourcing it: .env is not guaranteed to
# be valid shell (unquoted spaces, `$` in secrets). A value already in the
# environment wins, so the script also runs without a .env.
envval() {
    eval "current=\${$1:-}"
    if [ -n "$current" ]; then
        printf '%s' "$current"
        return
    fi
    [ -f .env ] || return 0
    grep -E "^$1=" .env | tail -n 1 | cut -d= -f2- | sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'$/\1/"
}

garage() {
    $COMPOSE exec -T garage /garage "$@"
}

KEY_ID=$(envval S3_ACCESS_KEY_ID)
SECRET=$(envval S3_SECRET_ACCESS_KEY)
BUCKET=$(envval S3_BUCKET)
BUCKET=${BUCKET:-uploads}
CAPACITY=$(envval GARAGE_CAPACITY)
CAPACITY=${CAPACITY:-20G}
KEY_NAME=site-backend

# Garage rejects other shapes at import time with a less helpful message.
if ! printf '%s' "$KEY_ID" | grep -Eq '^GK[0-9a-f]{24}$'; then
    echo "S3_ACCESS_KEY_ID must be GK followed by 24 hex chars (try: echo GK\$(openssl rand -hex 12))" >&2
    exit 1
fi
if ! printf '%s' "$SECRET" | grep -Eq '^[0-9a-f]{64}$'; then
    echo "S3_SECRET_ACCESS_KEY must be 64 hex chars (try: openssl rand -hex 32)" >&2
    exit 1
fi

# 1. Layout. `node id -q` prints "<full id>@<addr>"; the id is what
#    `layout assign` takes.
if garage layout show 2>/dev/null | grep -q "No nodes currently have a role"; then
    NODE_ID=$(garage node id -q 2>/dev/null | cut -d@ -f1)
    echo "Assigning node $(printf '%s' "$NODE_ID" | cut -c1-16) (capacity $CAPACITY)"
    garage layout assign -z dc1 -c "$CAPACITY" "$NODE_ID"
    garage layout apply --version 1
else
    echo "Layout already applied"
fi

# 2. Bucket.
if garage bucket info "$BUCKET" >/dev/null 2>&1; then
    echo "Bucket $BUCKET exists"
else
    garage bucket create "$BUCKET"
fi

# 3. Key.
if garage key info "$KEY_ID" >/dev/null 2>&1; then
    echo "Key $KEY_ID exists"
else
    garage key import -n "$KEY_NAME" --yes "$KEY_ID" "$SECRET"
fi

# 4 and 5 are idempotent in Garage itself.
garage bucket allow --read --write "$BUCKET" --key "$KEY_ID"
garage bucket website --allow "$BUCKET"

echo "Garage ready: bucket $BUCKET, key $KEY_ID"
