#!/usr/bin/env bash

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

