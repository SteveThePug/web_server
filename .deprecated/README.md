# .deprecated

STATUS: nothing in here is live. No compose file, nginx template or script
references this directory. It is kept, not deleted, in line with the repo's
"legacy is kept" convention; the full history is also in git.

## quartz/

The static-site build that used to serve `/notes/`. Replaced by the read-only
`silverbullet-public` service (PR #6). Its compose service and nginx route were
removed at that point; only the build context survives here.

## hasura/

Admin-only DB console that was proxied at `/hasura/`.

- `docker-compose.hasura.yml` — the service, verbatim.
- `nginx-locations.conf` — the upstream variable and `location` blocks that
  were in both `nginx.conf.template` and `nginx_dev_common.conf.template`.

To bring it back: layer the compose file
(`docker compose -f docker-compose.yml -f .deprecated/hasura/docker-compose.hasura.yml up`),
paste the nginx blocks back into both templates, re-add `${HASURA_HOST}
${HASURA_PORT}` to `ENVSUBST_VARS` in `nginx/entrypoint.sh`, and set
`HASURA_HOST`, `HASURA_PORT` and `HASURA_GRAPHQL_ADMIN_SECRET` in `.env`. Dev
mode additionally set `HASURA_GRAPHQL_ENABLE_CONSOLE` and
`HASURA_GRAPHQL_DEV_MODE` to `"true"`.

## autoheal/

Watchdog that restarted any container docker marked `unhealthy` (only `backend`
and `python` define healthchecks). It mounts the docker socket, which is
effectively root on the host. To bring it back, layer
`docker-compose.autoheal.yml` the same way as above.
