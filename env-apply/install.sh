#!/usr/bin/env bash
# =============================================================================
# install.sh — one-time setup, on the Pi, of the watcher behind the admin
# panel's Environment tab.
#
# WHAT IT DOES
#   Installs two systemd units:
#     web-server-env-apply.path     watches env-apply/run/request.json
#     web-server-env-apply.service  runs apply.py when that file appears
#   The units are generated here rather than committed because they need this
#   checkout's absolute path and the user that owns it.
#
# HOW TO RUN
#   From the deploy checkout, as the user that runs docker compose there:
#     ./env-apply/install.sh
#   It calls sudo for the three steps that need root. Re-running is safe and
#   is how to pick up a change to the units below.
#
# WHY A PATH UNIT
#   PathExists fires as soon as the backend writes the request, and also at
#   boot if a request was queued while the Pi was down. apply.py renames the
#   request away before doing anything else, so the unit cannot retrigger on
#   the same file.
#
# The service runs as the checkout's owner, not root: it needs exactly what
# the deploy workflow has (write .env, run docker compose) and no more.
# DOCKER_API_VERSION and COMPOSE_BAKE mirror .gitea/workflows/deploy.yaml.
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_USER="$(id -un)"
RUN_DIR="${ROOT}/env-apply/run"

if [ "$RUN_USER" = "root" ]; then
    echo "Run this as the user that owns the checkout, not root." >&2
    exit 1
fi
if ! command -v python3 >/dev/null; then
    echo "python3 is required." >&2
    exit 1
fi

# The backend bind-mounts this directory. If Docker creates it first it is
# root-owned and apply.py cannot write its status there.
mkdir -p "$RUN_DIR"
if [ "$(stat -c %U "$RUN_DIR")" != "$RUN_USER" ]; then
    sudo chown "$RUN_USER" "$RUN_DIR"
fi

sudo tee /etc/systemd/system/web-server-env-apply.path >/dev/null <<EOF
[Unit]
Description=Watch for .env changes queued from the admin panel

[Path]
PathExists=${RUN_DIR}/request.json

[Install]
WantedBy=multi-user.target
EOF

sudo tee /etc/systemd/system/web-server-env-apply.service >/dev/null <<EOF
[Unit]
Description=Apply a .env change queued from the admin panel
After=docker.service

[Service]
Type=oneshot
User=${RUN_USER}
WorkingDirectory=${ROOT}
Environment=DOCKER_API_VERSION=1.41
Environment=COMPOSE_BAKE=false
ExecStart=/usr/bin/env python3 ${ROOT}/env-apply/apply.py
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now web-server-env-apply.path

echo "Installed. Watching ${RUN_DIR}/request.json as ${RUN_USER}."
echo "Logs: journalctl -u web-server-env-apply.service"
