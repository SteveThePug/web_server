#!/usr/bin/env bash
# =============================================================================
# gitea-runner/run.sh — register (once) and then run the Gitea Actions runner.
#
# WHAT IT DOES
#   1. Downloads the act_runner binary if it is not already present.
#   2. Blocks until Gitea answers on http://localhost:3000.
#   3. Registers with Gitea the first time only.
#   4. execs the runner daemon in the foreground.
#
# WHEN IT RUNS
#   On the Pi, outside docker, as a long-lived process (systemd unit or nohup).
#   It must be started from THIS directory: the script uses relative paths
#   (./act_runner, .runner) and `act_runner daemon` looks for .runner in the
#   current working directory.
#
# WHY IT WAITS FOR PORT 3000
#   On a Pi reboot this process and the docker stack race. Registering or
#   starting the daemon against a Gitea that is not up yet fails permanently
#   rather than retrying, so the loop polls (forever, by design — there is no
#   sensible timeout when the answer is "wait longer") until Gitea responds.
#   It targets localhost:3000, the port docker-compose.yml publishes directly,
#   NOT the public /gitea/ path through nginx — so the runner keeps working
#   even while TLS or nginx is broken.
#
# THE REGISTRATION TOKEN (one-shot, and easy to get wrong)
#   GITEA_RUNNER_REGISTRATION_TOKEN comes from ./.env and is generated in the
#   Gitea web UI (Site Administration -> Actions -> Runners -> Create new
#   runner). It is a single-use enrolment token, NOT a lasting credential:
#   registration exchanges it for a permanent runner identity written to the
#   .runner file (gitignored — that file IS the real credential).
#
#   The `if [ ! -f .runner ]` guard is what makes this idempotent. Once .runner
#   exists the token is never used again and may safely be stale in .env. To
#   re-register a runner you must delete .runner AND mint a FRESH token in the
#   UI — reusing the old one will fail.
#
#   --labels self-hosted must match `runs-on: self-hosted` in
#   .gitea/workflows/*.yaml, or jobs queue forever with no runner picking them up.
#
# ENV VARS
#   GITEA_RUNNER_REGISTRATION_TOKEN  required for first registration only
#   GITEA_RUNNER_NAME                optional, defaults to "pi-runner"
#   (this script does not source .env itself — export them or use an
#    EnvironmentFile in the systemd unit)
#
# Diagnostics go to stderr so stdout stays clean for the daemon's own output.
# =============================================================================
if [ ! -f ./act_runner ]; then
    echo "act_runner binary not found. Downloading..." >&2
    bash "$(dirname "$0")/download.sh"
fi

echo "Waiting for localhost:3000 to respond..." >&2

while ! curl -sf http://localhost:3000 > /dev/null 2>&1; do
    sleep 2
done

echo "localhost:3000 is up." >&2

if [ ! -f .runner ]; then
    echo "No .runner file found. Registering runner..." >&2
    ./act_runner register --no-interactive \
        --instance http://localhost:3000 \
        --token "${GITEA_RUNNER_REGISTRATION_TOKEN}" \
        --name "${GITEA_RUNNER_NAME:-pi-runner}" \
        --labels self-hosted
fi

echo "Starting act_runner daemon..." >&2
exec ./act_runner daemon

