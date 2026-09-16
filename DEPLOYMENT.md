# Deployment

How this site is put together and how to operate it. Application code is
documented in `CLAUDE.md` and `README.md`; this file covers infrastructure only
— compose, nginx, TLS, the runner, and the runbook.

Everything runs on a single Raspberry Pi, as Docker containers on one bridge
network, behind one nginx.

---

## 1. Service topology

Only three things are reachable from outside the Pi: nginx on 80/443, Gitea on
3000/2222, and the liquidsoap harbor port. Everything else is internal.

```
                 internet
                     |
            :80  :443|                        :3000 / :2222      :8001
                     v                              v               v
              +--------------+                 +--------+    +-------------+
              |    nginx     |                 | gitea  |    |  liquidsoap |
              | (TLS, SPA,   |                 | (git + |    |   harbor    |
              |  reverse     |                 |  CI)   |    |  (live DJ)  |
              |  proxy)      |                 +--------+    +------+------+
              +------+-------+                      |               |
                     |                              |               |
   ---------------------------------------------    |        (same container
   |        |         |        |       |       |    |         as icecast2)
   v        v         v        v       v       v    |               |
/api/    /radio/   /gitea/  /hasura/ /notes/  /py/  |               |
backend  icecast2  gitea    hasura   quartz  python |               |
  |         ^                  |                    |               |
  |         |                  |                    |               |
  |         +------------------|--------------------|---------------+
  |                            |                    |
  +----------+-----------------+--------------------+
             v
          +------+
          |  db  |  postgres:16
          +------+  (app database + gitea database)

  vue      one-shot build container -> writes vue_dist volume -> nginx serves it
  certbot  no ports; shares ./certbot/conf and ./certbot/www with nginx
  autoheal no network; watches the docker socket and restarts unhealthy containers
```

### Who listens where

| Service | Container name | Internal port | Public route | Published to host |
|---|---|---|---|---|
| nginx | `nginx` | 80, 443 | — | **80, 443** |
| backend (Go) | `${BACKEND_HOST}` | `${BACKEND_PORT}` (8080) | `${BACKEND_ENDPOINT}` = `/api/` | no |
| db (Postgres 16) | `${POSTGRES_HOST}` | 5432 | none | no |
| hasura | `${HASURA_HOST}` | `${HASURA_PORT}` | `/hasura/` (admin-gated) | no |
| quartz (notes) | `${QUARTZ_HOST}` | `${QUARTZ_PORT}` | `/notes/` (admin-gated) | no |
| python (FastAPI) | `${PYTHON_HOST}` | `${PYTHON_PORT}` | `/py/`, docs at `/py/docs` | no |
| icecast2 + liquidsoap | `${ICECAST_HOST}` | `${ICECAST_PORT}` | `/radio/` | **harbor port only** |
| gitea | `${GITEA_HOST}` | 3000, 2222 | `/gitea/` | **3000, 2222** |
| vue | `vue` | 5173 (dev only) | `/` (dev only) | no |
| certbot | `certbot` | — | — | no |
| autoheal | `autoheal` | — | — | no |

Container names come from `.env`, and nginx builds its upstreams from the same
variables. **Renaming a `*_HOST` value renames the container and repoints nginx
in one step — never change one without the other.**

### Things nginx does that are easy to miss

- **Canonical host.** `http://` → `https://`, and apex → `www`. The SPA, the
  cookies and the CORS rules all assume `www.<DOMAIN>`.
- **Prefix stripping.** `/api`, `/radio`, `/gitea`, `/notes`, `/py` are stripped
  by a `rewrite ... break` before proxying; each upstream serves from `/`.
- **Per-request DNS.** Upstreams go through `set $upstream_* ...` variables with
  `resolver 127.0.0.11`, so names resolve per request. nginx boots fine with a
  backend down (502s until it appears) instead of refusing to start.
- **Admin gate.** `/hasura/` and `/notes/` are protected by an `auth_request`
  subrequest to the backend's `/auth/validate-admin`; a failure redirects to the
  SPA login. Neither service has any auth of its own at that path.
- **Uploads served directly.** `/uploads/` is an alias onto the shared `uploads`
  volume with `nosniff`, `Content-Disposition: inline` and a deny-all CSP, so
  user-supplied files cannot execute anything on our origin.
- **Rate limits** — see §6.

---

## 2. Dev vs prod boot paths

### Production

```
docker compose up --build
```

1. `vue` builds the SPA into the `vue_dist` volume and **exits 0** (expected).
2. `nginx` starts, runs `nginx/entrypoint.sh`, which:
   - is not in `DEV_MODE`, so it looks for `/etc/letsencrypt/live/$DOMAIN/`;
   - renders `nginx.conf.template` if the cert is there, else
     `nginx_setup.conf.template`;
   - polls up to 120s for `index.html` to appear in `vue_dist`, then starts
     nginx either way.
3. `certbot` obtains or confirms the certificate, then loops renewing.
4. Everything else comes up in parallel; `depends_on` here is the plain list
   form, which waits only for *start*, not readiness.

On the Pi this is triggered automatically: `.gitea/workflows/deploy.yaml` runs on
every push to `main`, fetches over `ssh://git@localhost:2222`, hard-resets the
deploy checkout, and runs `docker compose up -d --build --remove-orphans`.

### Development

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

The override layer changes five services and disables one:

- **vue** — `npm run dev` instead of a one-shot build; `./vue` is bind-mounted
  for hot reload, with anonymous-volume masks on `/app/node_modules` and
  `/app/src/wasm` so the image's installed deps and compiled WASM are not hidden
  by the bind mount. You need neither `node_modules` nor a Rust toolchain on the
  host.
- **nginx** — `DEV_MODE=true`, which makes the entrypoint self-sign a
  `CN=localhost` certificate (once, kept in `certbot/conf`) and render
  `nginx_dev.conf.template`: `/` proxies to the Vite dev server, both `:80` and
  `:443` serve the whole site, no HSTS, no HTTP→HTTPS redirect.
- **backend** — `DEV_MODE`, GraphQL playground and introspection on, a localhost
  Spotify redirect URI, and **`SEED_DB=true`**.
- **hasura** — console and dev mode on.
- **python** — `uvicorn --reload` with `python/app` bind-mounted.
- **certbot** — put in the `disabled` profile, so it never starts. Let's Encrypt
  cannot validate localhost anyway.

Visit `http://localhost`.

**`SEED_DB`** makes the Go backend populate the database with test posts, users
and messages at startup. It is set only in the dev override and must never be
set on the Pi — it would inject fake content into the live site.

Frontend-only work can skip Docker: `cd vue && npm run dev`, which proxies
`/api`, `/gitea` and `/radio` to locally running services.

---

## 3. TLS: first issuance and renewal

### The chicken-and-egg

The production nginx config references
`/etc/letsencrypt/live/$DOMAIN/fullchain.pem`, and nginx refuses to start if that
file is missing. Certbot's HTTP-01 challenge needs a working HTTP server on port
80. On a brand-new host neither can go first.

`nginx_setup.conf.template` breaks the loop: port 80 only, no TLS directives, no
upstreams, nothing that can fail to load. It serves
`/.well-known/acme-challenge/` from `/var/www/certbot` and 404s everything else,
so a half-configured site is never partially exposed.

**First run on a new domain:**

1. `docker compose up -d`. nginx logs `Certificates NOT found. Using setup nginx
   config.`
2. certbot runs `certbot certonly --webroot` and writes the certificate into
   `./certbot/conf` (bind-mounted into both containers — that sharing is what
   makes the whole thing work).
3. **`docker compose restart nginx`.** This step is manual. Nothing watches the
   cert directory. On restart the entrypoint re-checks, finds the cert, and
   renders the production config.

### Renewal

`certbot/entrypoint.sh` loops `certbot renew` every 12 hours. Renewal is a no-op
until the cert is inside its 30-day window, so this is well within Let's Encrypt
rate limits — it just means the window is checked twice a day. Certs are valid 90
days, so renewal lands around day 60.

**Caveat worth knowing:** there is no `--deploy-hook`, and the certbot container
cannot signal nginx (no docker socket, no shared PID namespace). nginx reads the
certificate at startup and holds it in memory, so **a renewed certificate is not
served until nginx is reloaded or restarted.** In practice pushes to `main`
recreate nginx often enough to hide this. If the site ever serves an expired
cert, `docker compose restart nginx` is the fix. See proposal 1 in §8.

---

## 4. What persists across restarts

**Named volumes** — survive `docker compose down` and rebuilds; destroyed only by
`docker compose down -v` or `docker volume rm`:

| Volume | Holds | Losing it means |
|---|---|---|
| `dbdata` | Postgres: the app database **and** the Gitea database | All posts, users, messages, and all Gitea metadata |
| `uploads` | User-uploaded files (backend writes, nginx serves) | All uploaded images |
| `vue_dist` | The built SPA | Nothing — rebuilt on next deploy |

**Bind mounts** — live in the repo checkout on the Pi, survive everything short
of deleting the directory:

| Path | Holds | Notes |
|---|---|---|
| `./certbot/conf` | TLS certificates and **private keys** | Gitignored. Never commit. |
| `./certbot/www` | ACME challenge webroot | Transient by nature |
| `./backend/token/` | **Spotify OAuth refresh token** (`spotify_token.json`) | Gitignored. Without this mount, Spotify would need re-authorising after every rebuild. |
| `./gitea/data` | Git repositories, LFS objects, avatars, sessions | The actual git data |
| `./gitea/config` | Gitea's live `app.ini` | Gitignored; generated from the template on first boot and then owned by Gitea |
| `./logs` | Backend logs | |
| `./icecast2/fallback_music` | Radio fallback MP3s | Synced by `sync-secrets.sh`, not git |
| `${OBSIDIAN_DIR}` | The Obsidian vault | Read-write to backend, **read-only** to quartz |

**Ephemeral:** anything else inside a container. Notably the Vue build, the
rendered nginx/icecast/quartz configs, and the Gitea runner's downloaded binary
are all regenerated.

---

## 5. Environment variables

All configuration lives in **`./.env`** at the repo root, which is gitignored.
`.env.example` documents every variable with blank values — start from that.

Two distinct mechanisms read it, and services use both:

- **`${VAR}` in `docker-compose.yml`** — substituted by the compose CLI at parse
  time. Builds container names, published ports, the Hasura database URL.
- **`env_file: ./.env`** — injects the whole file into the container. Used by
  nginx, backend, db, icecast2, quartz, certbot.

`gitea` instead receives explicit `GITEA__section__KEY` variables (Gitea's own
convention for overriding `app.ini` keys), and `python` receives an explicit
three-variable allow-list, so it never sees the database password or any API key.

**`.env` reaches the Pi via `./sync-secrets.sh`**, which syncs exactly two
things between this checkout and `adamf@stppi.local:~/deploy/web_server`:
`.env`, and `icecast2/fallback_music/*.mp3`. The sync is **bidirectional and
resolved by modification time — newer wins, with no merge.** Editing `.env` on
both machines silently loses one side's edits. Run it before and after editing.

An unset variable is not an error: `envsubst` renders it as an empty string,
which for a password field means *no password*. After editing `.env`, confirm
with `docker compose config`.

---

## 6. Rate limits

Defined once per config in the `http` block, applied per location. Keyed on
`$binary_remote_addr` — the client IP in packed form, so each entry is ~64 bytes
and a `10m` zone holds roughly 160k IPs. `10m` is a safe round number, not a
tuned value.

`rate` is a **smoothed leaky bucket**, not a per-window counter: `5r/m` means one
request every 12 seconds, not five at once. `burst=N` lets N requests queue ahead
of that drip, and `nodelay` (used everywhere here) serves those N immediately.
So the effective behaviour is *"N instant requests, then throttled to the drip
rate"*. Anything beyond the burst gets an immediate **503**.

| Zone | Rate | Burst | Applied to | Why |
|---|---|---|---|---|
| `login` | 5/min | 3 | `/api/auth/login` | Brute-force defence. A human mistyping gets 3 tries then one per 12s; a credential-stuffing script gets nothing useful. |
| `api` | 30/sec | 20 | `/api/`, `/py/` | Generous on purpose — one SPA page load fans out to several calls. An abuse ceiling, not a quota. |
| `graphql` | 10/sec | 10 | `/api/graphql` | Tighter than REST because one query can be arbitrarily expensive. **Production config only**; dev folds GraphQL into `api`. |
| `upload` | 5/min | 3 | `/api/messages/upload` | Uploads can be 50MB and hit disk; bounds how fast one IP fills the volume. |
| `hotel_search` | 3/min | 2 | `/py/hotels/(search\|scan)` | Each request holds an upstream call for minutes and the Python side has one search slot — concurrency protection more than abuse protection. |

The WebSocket endpoint is deliberately unlimited: a socket is one long-lived
request, and a limit there would drop reconnect storms rather than abuse.

These zones key on the direct client IP, which is correct because nginx is the
sole edge. **If a CDN or proxy is ever put in front, they must switch to a
trusted `X-Forwarded-For` or every visitor shares one bucket.**

---

## 7. Runbook

All commands run from the deploy checkout on the Pi
(`~/deploy/web_server`) unless stated.

**Deploy.** Normally automatic — push to `main` and the Gitea Actions runner does
it. By hand:

```
git pull && docker compose up -d --build --remove-orphans
```

**Rebuild one service** (e.g. after changing only the backend):

```
docker compose up -d --build backend
```

Note: editing `certbot/entrypoint.sh` needs no rebuild (it is bind-mounted over
the stock image's entrypoint) — just `docker compose restart certbot`. Editing
an nginx **template** does need a rebuild, since templates are `COPY`'d into the
image.

**Apply a new nginx config / pick up a renewed certificate:**

```
docker compose up -d --build nginx     # template changed
docker compose restart nginx           # certificate changed
```

**Renew certificates manually** (then restart nginx so it is actually served):

```
docker compose exec certbot certbot renew --webroot -w /var/www/certbot
docker compose restart nginx
```

Add `--dry-run` to test without burning rate limit.

**Check logs:**

```
docker compose logs -f nginx           # follow one service
docker compose logs --tail=200 backend
docker compose ps                      # note: `vue` Exited(0) is correct
docker compose logs certbot            # issuance / renewal outcomes
```

**Reset the database** — destructive; wipes the app database **and Gitea's**:

```
docker compose down
docker volume rm web_server_dbdata      # confirm the name with `docker volume ls`
docker compose up -d
```

The app schema is recreated by GORM auto-migration on backend start (there are no
migration files). **Gitea's database is not auto-created** — the Postgres image
only creates `POSTGRES_DB` on first init, so create `${POSTGRES_GITEA_DB}`
manually before Gitea will start:

```
docker compose exec db createdb -U "$POSTGRES_USER" "$POSTGRES_GITEA_DB"
```

**Sync secrets / fallback music** (from the laptop checkout, not the Pi):

```
./sync-secrets.sh
```

**Set up the CI runner on a fresh Pi:**

```
cd gitea-runner
bash download.sh                                  # pulls the arm64 binary
# mint a registration token: Gitea -> Site Administration -> Actions -> Runners
GITEA_RUNNER_REGISTRATION_TOKEN=... ./run.sh      # registers once, then daemonises
```

The runner lives **outside** Docker on purpose: it runs `docker compose up`
against this stack, so it cannot be a container in it. Registration is one-shot —
`.runner` is the lasting credential; re-registering needs that file deleted *and*
a fresh token.

**Go live on the radio:** point a broadcasting client (e.g. butt) at the Pi on
`${LIQUIDSOAP_HARBOR_PORT}` with mount `${LIQUIDSOAP_HARBOR_MOUNT}` and the
icecast source password. liquidsoap pre-empts the fallback playlist immediately
(`track_sensitive=false`). Disconnect and it falls back to the playlist, then to
silence. Adding MP3s to `icecast2/fallback_music/` is picked up live
(`reload_mode="watch"`) with no restart.

**Unhealthy containers** are restarted automatically by `autoheal`, which polls
the docker socket every 30s. Only `backend` and `python` define healthchecks, so
only they are covered.

---

## 8. Known rough edges

Documented, not fixed — each would change deploy behaviour.

1. **Renewed certificates are not served until nginx restarts.** No
   `--deploy-hook` in `certbot/entrypoint.sh`. Fix: add
   `--deploy-hook "touch /var/www/certbot/.reloadme"` plus a watcher, or give
   certbot the docker socket to send nginx a `HUP`.
2. **First TLS issuance needs a manual `docker compose restart nginx`.** The
   entrypoint only re-evaluates cert presence at container start.
3. **`gitea/config/entrypoint.sh` is an empty, unused file** that would be
   mounted to the same in-container path as the real wrapper. It is currently
   shadowed by an explicit file mount — fragile. Deleting it would be safe but
   is a file removal.
4. **The dev nginx template duplicates its `:80` and `:443` server blocks**
   almost verbatim (~190 lines), so a routing fix must be made twice.
5. **`nginx.conf.template` has a `/img/stamps/mine.gif` CORS header allowing
   only our own origin**, which defeats the point of having the header at all.
6. **`certbot/entrypoint.sh` uses unquoted `${EMAIL}` and `${DOMAIN}`.** Inert
   today (neither contains whitespace) but a latent quoting bug.
7. **`UPTIMEKUMA_*` and `WALLABAG_*` are substituted by nginx but no compose
   service provides them** — leftovers from removed services.
8. **`autoheal` mounts the docker socket**, which is effectively root on the
   host, and uses the `:latest` tag.
