# .deprecated

STATUS: nothing in here is live. No compose file, Caddyfile or script
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

Since 2026-10-19 it is back in `docker-compose.yml`, watching only containers
labelled `autoheal=true` (caddy; nginx until 2026-11-17). This copy is the old
`all` variant, kept for reference.

## nginx/ and certbot/

The public edge until 2026-11-17, when Caddy (`caddy/`) replaced both:
`nginx/` is the image (three templates chosen at start by `entrypoint.sh`,
the shared route table, the payload log and security-header includes), and
`certbot/entrypoint.sh` obtained and renewed the certificate that nginx
reloaded through a sentinel file. Their compose services were removed in
`daa2a46`. The last certbot certificate is still in `./certbot/conf` on the
host, valid until 2027-01-20.

To bring them back: revert the commit that moved them here and then `daa2a46`;
DEPLOYMENT.md §7 ("Roll back to nginx") has the commands.
