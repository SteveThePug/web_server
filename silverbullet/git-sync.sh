#!/bin/sh
# Commits SilverBullet's edits to the notes repo and syncs them with the remote.
#
# SilverBullet writes plain markdown straight into the space folder and has no
# concept of git, so this sidecar is what keeps the notes repo moving. It is
# deliberately a separate container rather than SilverBullet's own shell backend:
# the editor is protected only by the nginx admin gate, and enabling a shell
# backend would hand anyone past that gate command execution inside the container.
#
# ON-DEMAND SYNC
#   Waiting out the interval is annoying when you want a note on another device
#   now, so the editor can ask for an immediate sync. The "Sync notes" button in
#   SilverBullet (defined in the vault's own Sync.md) writes a timestamp to
#   TRIGGER_FILE; this script notices the modification time change and syncs on
#   the spot. The file is gitignored in the notes repo and is never deleted here
#   — deleting it would fight SilverBullet's client, which would re-upload its
#   cached copy and cause a delete/recreate loop. Only the mtime matters.

set -u

SPACE=/space
INTERVAL="${SYNC_INTERVAL:-300}"
TRIGGER_FILE="$SPACE/${SYNC_TRIGGER_FILE:-sync-request.md}"
POLL=2

log() { echo "[git-sync] $(date -Iseconds) $*"; }

git config --global --add safe.directory "$SPACE"

if [ ! -d "$SPACE/.git" ]; then
    log "ERROR: $SPACE is not a git repository."
    log "Clone the notes repo to the host path OBSIDIAN_DIR points at, then restart."
    # Idle rather than crash-loop, so the logs stay readable.
    while true; do sleep 3600; done
fi

cd "$SPACE" || exit 1

trigger_mtime() {
    [ -f "$TRIGGER_FILE" ] && stat -c %Y "$TRIGGER_FILE" 2>/dev/null || echo ""
}

sync_once() {
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
}

last_trigger=$(trigger_mtime)
log "watching $SPACE every ${INTERVAL}s (on-demand via ${TRIGGER_FILE#$SPACE/})"

while true; do
    sync_once

    # Sleep in short slices so an on-demand request is picked up quickly.
    waited=0
    while [ "$waited" -lt "$INTERVAL" ]; do
        sleep "$POLL"
        waited=$((waited + POLL))
        now=$(trigger_mtime)
        if [ -n "$now" ] && [ "$now" != "$last_trigger" ]; then
            last_trigger="$now"
            log "sync requested from the editor"
            break
        fi
    done
done
