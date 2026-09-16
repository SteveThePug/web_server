#!/bin/sh
# =============================================================================
# gitea entrypoint wrapper — materialise app.ini on first boot, then hand off.
#
# WHAT IT DOES
#   Gitea refuses to start without /etc/gitea/app.ini, but app.ini is a live
#   file: Gitea rewrites it when you change settings in the web UI, and it ends
#   up holding generated secrets (SECRET_KEY, INTERNAL_TOKEN, ...). So it is
#   gitignored and NOT committed. Only app.ini.template is in git, with the
#   secret fields left blank. This script copies the template into place the
#   first time the container ever runs on a host, then gets out of the way.
#
#   The `if [ ! -f ]` guard is the important part: on every subsequent boot the
#   real app.ini already exists in the ./gitea/config bind mount and is left
#   completely alone. Editing app.ini.template therefore has NO effect on an
#   existing install — you must edit ./gitea/config/app.ini directly (or delete
#   it to re-seed, which loses any UI-made settings).
#
# WHEN IT RUNS
#   Container start. docker-compose.yml overrides gitea's entrypoint to
#   ["/usr/bin/dumb-init", "--", "/etc/gitea/entrypoint.sh"], keeping dumb-init
#   as PID 1 for signal handling exactly as the stock image does, and inserting
#   this script ahead of Gitea's own docker-entrypoint.sh.
#
# WHERE THE SECRETS ACTUALLY COME FROM
#   Not from the template. docker-compose.yml passes GITEA__section__KEY
#   environment variables (database credentials, LFS_JWT_SECRET,
#   INTERNAL_TOKEN, OAUTH2 JWT_SECRET) sourced from ./.env; Gitea applies those
#   over whatever app.ini says. That is why the template can safely ship with
#   those fields empty.
#
# ASSUMES EXISTS
#   /etc/gitea  - bind mount of ./gitea/config (persistent, gitignored)
#   /etc/gitea/app.ini.template  - committed, lives in that same directory
#   /usr/local/bin/docker-entrypoint.sh  - Gitea's real entrypoint
#
# `exec` replaces this shell so Gitea inherits PID and receives signals
# directly instead of through a lingering parent.
# =============================================================================
set -e

# Generate app.ini from template if it doesn't already exist
if [ ! -f /etc/gitea/app.ini ]; then
    cp /etc/gitea/app.ini.template /etc/gitea/app.ini
    echo "Generated app.ini from template"
fi

exec /usr/local/bin/docker-entrypoint.sh "$@"
