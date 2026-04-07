#!/usr/bin/env bash
set -euo pipefail

VERSION="0.2.11"
BASE_URL="https://gitea.com/gitea/act_runner/releases/download/v${VERSION}"

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ASSET="act_runner-${VERSION}-linux-amd64" ;;
    aarch64) ASSET="act_runner-${VERSION}-linux-arm64" ;;
    armv7l)  ASSET="act_runner-${VERSION}-linux-armv7" ;;
    *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEST="${SCRIPT_DIR}/act_runner"

echo "Downloading act_runner v${VERSION} for ${ARCH}..."
curl -fSL "${BASE_URL}/${ASSET}" -o "$DEST"
chmod +x "$DEST"
echo "Downloaded to $DEST"
