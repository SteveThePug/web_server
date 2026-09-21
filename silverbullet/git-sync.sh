#!/bin/sh
# Commits SilverBullet's edits to the notes repo and syncs them with the remote.
#
# SilverBullet writes plain markdown straight into the space folder and has no
# concept of git, so this sidecar is what keeps the notes repo moving. It is
# deliberately a separate container rather than SilverBullet's own shell backend:
# the editor is protected only by the nginx admin gate, and enabling a shell
# backend would hand anyone past that gate command execution inside the container.

set -u

SPACE=/space
INTERVAL="${SYNC_INTERVAL:-300}"

log() { echo "[git-sync] $(date -Iseconds) $*"; }

git config --global --add safe.directory "$SPACE"

if [ ! -d "$SPACE/.git" ]; then
    log "ERROR: $SPACE is not a git repository."
    log "Clone the notes repo to the host path OBSIDIAN_DIR points at, then restart."
    # Idle rather than crash-loop, so the logs stay readable.
    while true; do sleep 3600; done
fi

cd "$SPACE" || exit 1
log "watching $SPACE every ${INTERVAL}s"

while true; do
    git add -A
    if ! git diff --cached --quiet; then
        if git commit -q -m "notes: $(date -Iseconds)"; then
            log "committed local changes"
        else
            log "WARN: commit failed"
        fi
    fi

    if git remote get-url origin >/dev/null 2>&1; then
        git pull --rebase --autostash -q || log "WARN: pull failed (conflict, auth or network)"
        git push -q || log "WARN: push failed (conflict, auth or network)"
    fi

    sleep "$INTERVAL"
done
