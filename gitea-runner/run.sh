#!/usr/bin/env bash

echo "Waiting for localhost:3000 to respond..." >&2

while ! curl -sf http://localhost:3000 > /dev/null 2>&1; do
    sleep 2
done

echo "localhost:3000 is up. Starting act_runner daemon..." >&2

exec ./act_runner daemon

