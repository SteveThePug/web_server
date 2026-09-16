#!/usr/bin/env bash
# =============================================================================
# gitea-runner/download.sh — fetch the act_runner binary for this machine.
#
# WHAT IT DOES
#   Downloads the correct act_runner release asset for the host CPU and drops
#   it next to this script as ./act_runner (gitignored — a binary, not source).
#
# WHY IT IS A SCRIPT AND NOT A CONTAINER
#   The runner executes the deploy workflow, which itself runs
#   `docker compose up -d --build` against this stack. Running it inside the
#   stack would mean a container restarting its own parent mid-job. So the
#   runner lives OUTSIDE docker, directly on the Pi, and is the one moving part
#   of this repo that docker compose does not manage.
#
# ARCHITECTURE SELECTION (the Raspberry Pi bit)
#   uname -m decides the asset. aarch64 is the 64-bit Pi OS case and is what
#   this deployment actually uses; armv7l covers a 32-bit Pi OS install and
#   x86_64 covers running the same script on a dev laptop. An unknown arch is a
#   hard error rather than a silent wrong download.
#
# VERSION is pinned so a runner upgrade is an explicit commit, never a surprise
#   on a re-run. Bump it here, delete ./act_runner, and re-run to upgrade.
#
# SCRIPT_DIR is resolved so the binary always lands beside the script even when
#   invoked from another directory.
#
# USAGE: bash gitea-runner/download.sh   (run.sh calls it automatically if the
#        binary is missing)
# =============================================================================
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
