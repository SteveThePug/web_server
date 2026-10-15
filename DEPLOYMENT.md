# Deployment

How this site is put together and how to operate it. Application code is
documented in `CLAUDE.md` and `README.md`; this file covers infrastructure only
— compose, nginx, TLS, the runner, and the runbook.

Everything runs on a single host, `stpvm` (an x86-64 Ubuntu VM on Proxmox,
reached over Tailscale as `stpvm.tailf20278.ts.net`; until 2026-10-15 it was
the Raspberry Pi `stppi`), as Docker containers on one bridge
network, behind one nginx.

---

## 1. Service topology

Only three things are reachable from outside the host: nginx on 80/443, Gitea on
3000/2222, and the liquidsoap harbor port. Everything else is internal. Web
traffic can also come in through the Cloudflare tunnel (`cloudflared`), which
dials out and publishes no port; see "Cloudflare Tunnel" in §3.

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
/api/    /radio/   /gitea/  /sb/     /notes/  /py/  |               |
backend  icecast2  gitea    sb       sb-pub  python |               |
  |         ^                                       |               |
  |         |                                       |               |
  |         +---------------------------------------|---------------+
  |                                                 |
  +----------+--------------------------------------+
             v
          +------+
          |  db  |  postgres:16
          +------+  (app database + gitea database)

  vue      one-shot build container -> writes vue_dist volume -> nginx serves it
  certbot  no ports; shares ./certbot/conf and ./certbot/www with nginx
```

### Who listens where

| Service                                       | Container name                | Internal port                 | Public route                                                                                            | Published to host    |
| --------------------------------------------- | ----------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------- |
| nginx                                         | `nginx`                       | 80, 443                       | —                                                                                                       | **80, 443**          |
| backend (Go)                                  | `${BACKEND_HOST}`             | `${BACKEND_PORT}` (8080)      | `${BACKEND_ENDPOINT}` = `/api/`                                                                         | no                   |
| db (Postgres 16)                              | `${POSTGRES_HOST}`            | 5432                          | none                                                                                                    | no                   |
| silverbullet (editor)                         | `${SILVERBULLET_HOST}`        | `${SILVERBULLET_PORT}`        | `/sb/` (admin-gated)                                                                                    | no                   |
| silverbullet-public (notes site)              | `${SILVERBULLET_PUBLIC_HOST}` | `${SILVERBULLET_PUBLIC_PORT}` | `/notes/` (**public**, read-only)                                                                       | no                   |
| silverbullet-git                              | n/a                           | n/a                           | none (timer loop)                                                                                       | no                   |
| python (FastAPI)                              | `${PYTHON_HOST}`              | `${PYTHON_PORT}`              | `/py/`, docs at `/py/docs`                                                                              | no                   |
| jobs (job tracker, published image)           | `jobs`                        | 8080                          | `/jobs/` (admin-gated by nginx `auth_request`; capture API and userscript public, bearer-token)       | no                   |
| activity-tracker (published image, adamf/ccp) | `activity-tracker`            | 8080                          | `/tracker/`, prefix stripped (pages admin-gated by nginx `auth_request`; `/api/v1/` bearer/app-checked) | no                   |
| icecast2 + liquidsoap                         | `${ICECAST_HOST}`             | `${ICECAST_PORT}`             | `/radio/`                                                                                               | **harbor port only** |
| gitea                                         | `${GITEA_HOST}`               | 3000, 2222                    | `/gitea/`                                                                                               | **3000, 2222**       |
| vue                                           | `vue`                         | 5173 (dev only)               | `/` (dev only)                                                                                          | no                   |
| certbot                                       | `certbot`                     | —                             | —                                                                                                       | no                   |
| cloudflared                                   | `cloudflared`                 | — (fixed IP `172.28.255.250`) | all of nginx, via Cloudflare                                                                            | no (outbound only)   |
| dozzle (log viewer)                           | `dozzle`                      | 8080                          | `/dozzle/` (admin-gated by nginx `auth_request`; reads the Docker socket)                               | no                   |
| traffic (tails the payload log)               | `traffic`                     | — (no network)                | none; its stdout is what Dozzle shows                                                                   | no                   |
| garage (S3 object store)                      | `garage`                      | 3900 S3, 3902 web, 3903 admin | none directly; `/uploads/` falls back to its web endpoint for files not on the volume                   | no                   |
| garage-webui (Garage admin UI)                | `garage-webui`                | 3909                          | `/garage/` (admin-gated by nginx `auth_request`)                                                        | no                   |

Container names come from `.env`, and nginx builds its upstreams from the same
variables. **Renaming a `*_HOST` value renames the container and repoints nginx
in one step — never change one without the other.**

### Things nginx does that are easy to miss

- **Canonical host.** `http://` → `https://`, and apex → `www`. The SPA, the
  cookies and the CORS rules all assume `www.<DOMAIN>`.
- **Prefix stripping.** `/api`, `/radio`, `/gitea`, `/py` are stripped
  — but `/sb/` and `/notes/` deliberately are NOT: SilverBullet is told its own base path via
  `SB_URL_PREFIX` and builds its service-worker scope from it, so stripping the
  prefix serves a blank page with nothing in the log.
  by a `rewrite ... break` before proxying; each upstream serves from `/`.
- **Per-request DNS.** Upstreams go through `set $upstream_* ...` variables with
  `resolver 127.0.0.11`, so names resolve per request. nginx boots fine with a
  backend down (502s until it appears) instead of refusing to start.
- **Admin gate.** `/sb/`, `/jobs/`, `/tracker/`, `/dozzle/`, `/garage/`, `/radio/admin/` and `/uploads/private/` are
  protected by an `auth_request` subrequest to the backend's `/auth/validate-admin`; a failure
  redirects to the SPA login. Most of these services have no auth of their own at that path; the
  job and activity trackers also check the session themselves. Their carve-outs, which a session
  cookie never reaches, stay public at nginx and are guarded by bearer tokens in the apps:
  `/jobs/api/capture` and `/jobs/job-capture.user.js`, `/tracker/api/v1/` and `/tracker/tracker.user.js`.
- **Public notes privacy boundary.** `/notes/` is public, and the only thing
  keeping `private/` (and `templates/`, `.git`, `.obsidian`) off it is a pair of
  nginx deny blocks. SilverBullet's `SB_SPACE_IGNORE` only hides paths from its
  file _list_; direct `/.fs/` fetches and server-rendered page URLs still serve
  them. Adding a folder to `SB_SPACE_IGNORE` on `silverbullet-public` without
  adding it to the deny regex in `nginx_routes.conf.template` publishes it. See the
  comment above `location /notes/` in `nginx_routes.conf.template`.
- **Uploads: disk first, then the bucket.** `/uploads/<f>` is looked for on
  the shared `uploads` volume and, if it is not there, fetched from Garage's
  website endpoint (`@uploads_bucket`). Both paths add `nosniff`,
  `Content-Disposition: inline` and a deny-all CSP, so user-supplied files
  cannot execute anything on our origin. Disk first means Garage being down
  only affects files that exist nowhere else; see "Garage" in §7.
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
   form, which waits only for _start_, not readiness.

On the host this is triggered automatically by `.gitea/workflows/deploy.yaml`,
which has two jobs:

- **`test`** runs on every push to any branch. It exports the pushed commit
  (never the deploy checkout) and runs the backend's `go vet` + `go test` in
  `golang:1.26` and the Python tests in `python:3.12-slim`. The Vue build is
  not repeated here: the image build in the deploy step does it, and a failed
  build leaves the old containers running.
- **`deploy`** runs only on `main`, only after `test` passes, and only if the
  tested commit is still the tip of `main` (so an older run cannot deploy a
  newer, untested commit). It fetches over `ssh://git@localhost:2222`,
  hard-resets the deploy checkout, and runs
  `docker compose up -d --build --remove-orphans` under the
  `.compose.lock` flock, shared with the weekly maintenance workflow.

The test job keeps state on the host outside the deploy checkout: a bare
repository cache at `/home/adamf/.cache/web_server-ci/` and three named
volumes (`web_server_ci_gomod`, `web_server_ci_gobuild`, `web_server_ci_pip`).
The weekly maintenance workflow's `docker image prune -af` removes the two
toolchain images, so the first run after it pulls them again; the volumes
survive. Maintenance takes the same flock around its compose steps.

`make check` runs the same checks locally, plus the Vue build and compose
validation — run it before pushing, since a red `test` job blocks the deploy.

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
- **python** — `uvicorn --reload` with `python/app` bind-mounted.
- **certbot** — put in the `disabled` profile, so it never starts. Let's Encrypt
  cannot validate localhost anyway.

Visit `http://localhost`.

**`SEED_DB`** makes the Go backend populate the database with test posts, users
and messages at startup. It is set only in the dev override and must never be
set on the host — it would inject fake content into the live site.

Frontend-only work can skip Docker: `cd vue && npm run dev`, which proxies
`/api`, `/gitea` and `/radio` to locally running services.

---

## 3. TLS: first issuance and renewal

### The chicken-and-egg

The production nginx config references
`/etc/letsencrypt/live/$DOMAIN/fullchain.pem`, and nginx refuses to start if that
file is missing. Certbot's HTTP-01 challenge needs a working HTTP server on port 80. On a brand-new host neither can go first.

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

**Validation method.** With `CLOUDFLARE_DNS_API_TOKEN` set, certbot proves
control of the domain by creating a `_acme-challenge` TXT record through the
Cloudflare API (DNS-01), so renewal needs no inbound port at all. That is
what lets the router's 80/443 forwards close behind the tunnel. Without the
token it uses the webroot (HTTP-01) on port 80, as it always did. Switching
either way takes effect at the next renewal; see `certbot/entrypoint.sh`.

**Caveat worth knowing:** there is no `--deploy-hook`, and the certbot container
cannot signal nginx (no docker socket, no shared PID namespace). nginx reads the
certificate at startup and holds it in memory, so **a renewed certificate is not
served until nginx is reloaded or restarted.** In practice pushes to `main`
recreate nginx often enough to hide this. If the site ever serves an expired
cert, `docker compose restart nginx` is the fix. See proposal 1 in §8.

### Cloudflare Tunnel

The `cloudflared` container holds an outbound connection to Cloudflare and
forwards requests for the site's hostnames down it to nginx. It is a
_remotely managed_ tunnel: the only thing in this repo is the token
(`CLOUDFLARE_TUNNEL_TOKEN` in `.env`); hostnames and
origin settings live in the dashboard.

**Origin: `https://www.<DOMAIN>`, the site's own name.** Cloudflare
terminates the visitor's TLS at its edge, and cloudflared opens a second TLS
connection to nginx. That is deliberate. Pointed at `http://nginx:80`, the
tunnel would hit the HTTP→HTTPS redirect and loop. Going to :443 keeps every
redirect, security header and the `X-Forwarded-Proto: https` the backend's
Secure cookies depend on, and needs no nginx change.

The origin is the public name, not `nginx`, because nginx carries `<DOMAIN>`
and `www.<DOMAIN>` as network aliases on `app-network` (`docker-compose.yml`).
Inside the network those names resolve to the nginx container, so the hop
never leaves the host, and cloudflared verifies the Let's Encrypt certificate
against a name it covers. `https://nginx` would fail that check and need No
TLS Verify. The flip side: **an expired certificate now takes the tunnel down
with a 502**, which is why renewal over DNS-01 (step 5) matters.

Setup, once:

1. Cloudflare dashboard → the domain must be on Cloudflare DNS.
2. Zero Trust → Networks → Tunnels → Create a tunnel → Cloudflared. Copy the
   token from the install command (the long string after `--token`).
3. Put it in `.env` as `CLOUDFLARE_TUNNEL_TOKEN`, then
   `docker compose up -d cloudflared`. The tunnel
   shows as Healthy in the dashboard.
4. Add a public hostname per name, each with service `HTTPS` and its own
   name as the URL, and no other settings:
    - `www.<DOMAIN>` → `https://www.<DOMAIN>`: the site.
    - `<DOMAIN>` → `https://<DOMAIN>`: nginx redirects it to www, and it serves Gitea's registry
      at `/v2/`.
      Creating each hostname replaces the existing A record for that name with
      a CNAME to the tunnel. **That is the cutover**: until then the tunnel
      carries nothing.

5. Create a Cloudflare API token (My Profile → API Tokens → "Edit zone
   DNS", this zone only) and set it as `CLOUDFLARE_DNS_API_TOKEN`. certbot
   then renews over DNS-01 (see Renewal above) and needs no inbound port.
6. On the host, point the apex at nginx locally (next section).
7. Close the router's 80 and 443 forwards. Once 5 and 6 are done nothing
   depends on them.

#### What the tunnel does not carry, and where it goes instead

- **Certificate renewal.** Over DNS-01 with `CLOUDFLARE_DNS_API_TOKEN` set;
  without it certbot falls back to HTTP-01, which needs port 80 forwarded.
  The certificate matters behind the tunnel: cloudflared verifies it, as
  do the registry path below and any direct LAN access. A fresh machine with no
  certificate at all boots the setup config (§3), which serves nothing over
  443, so the tunnel carries nothing until the first issuance.
- **Registry pushes and pulls on the host.** Cloudflare caps request bodies at
  100 MB on the free plan, so a push of a larger image layer through the
  tunnel fails. The runner on the host pushes to `<DOMAIN>/adamf/...` and the
  stack pulls from it, both through the host's Docker daemon, which resolves
  names through the host's `/etc/hosts`. Pin the apex there to the loopback:

  ```
  127.0.0.1   <DOMAIN>
  ```

  so registry traffic goes straight to nginx on :443, with the real
  certificate, and never leaves the host. (Without it the name resolves to
  Cloudflare once DNS moves there.) Only the apex: www stays on Cloudflare for anything on the host
  that wants the public site. Pushing from another machine still goes
  through Cloudflare and its cap; use the host's Tailscale name for that.
- **Gitea SSH (2222).** Not HTTP, so not tunnelled. Already Tailscale-only:
  `GITEA_SSH_DOMAIN` in `.env` (Gitea's `SSH_DOMAIN`) is the host's MagicDNS name
  and 2222 has never been forwarded. Nothing changes.
- **The radio harbor (8001).** Also a raw source protocol, so a DJ client
  pointed at `<DOMAIN>:8001` stops connecting once the domain resolves to
  Cloudflare. Point it at the host's Tailscale name
  (`stpvm.tailf20278.ts.net:8001`) and close the 8001 forward. A guest DJ
  off the tailnet would need either a Tailscale share of the host or a
  Cloudflare TCP route used with `cloudflared access tcp` on their side.

#### Other limits
- Cloudflare returns a 524 if the origin sends no response bytes for 100
  seconds. Streaming endpoints are fine while they keep writing; a long
  silent request is not.
- Dev never runs cloudflared (`disabled` profile in `docker-compose.dev.yml`),
  so a laptop cannot join the production tunnel.

---

## 4. What persists across restarts

**Named volumes** — survive `docker compose down` and rebuilds; destroyed only by
`docker compose down -v` or `docker volume rm`:

| Volume     | Holds                                                 | Losing it means                                    |
| ---------- | ----------------------------------------------------- | -------------------------------------------------- |
| `dbdata`   | Postgres: the app database **and** the Gitea database | All posts, users, messages, and all Gitea metadata |
| `uploads`  | Uploads from before Garage, private attachments, and any upload Garage refused | Those files                               |
| `garage_meta` | Garage's metadata: buckets, keys, object index, layout | Everything in the bucket (the bytes in `garage_data` are unreadable without it) |
| `garage_data` | The bucket's object bytes                          | Every upload stored in Garage                      |
| `vue_dist` | The built SPA                                         | Nothing — rebuilt on next deploy                   |

**Bind mounts** — live in the repo checkout on the host, survive everything short
of deleting the directory:

| Path                        | Holds                                                                                               | Notes                                                                                  |
| --------------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `./certbot/conf`            | TLS certificates and **private keys**                                                               | Gitignored. Never commit.                                                              |
| `./certbot/www`             | ACME challenge webroot                                                                              | Transient by nature                                                                    |
| `./backend/token/`          | **Spotify OAuth refresh token** (`spotify_token.json`)                                              | Gitignored. Without this mount, Spotify would need re-authorising after every rebuild. |
| `./jobs-data`               | The job tracker's **capture token hash** and Graph OAuth session                                    | Gitignored. Must be owned by uid 1000, the user the `jobs` image runs as.              |
| `./gitea/data`              | Git repositories, LFS objects, avatars, sessions                                                    | The actual git data                                                                    |
| `./gitea/config`            | Gitea's live `app.ini`                                                                              | Gitignored; generated from the template on first boot and then owned by Gitea          |
| `./logs`                    | Backend logs (`go.log`); nginx's JSON request-body log in `logs/nginx/payloads.log`                  | The payload log grows without bound; rotate it on the host (§7, "Incoming request bodies") |
| `./icecast2/fallback_music` | Radio fallback MP3s                                                                                 | Copied to the host by hand, not git                                                    |
| `${OBSIDIAN_DIR}`           | The Obsidian vault (a git clone of the notes repo)                                                  | Read-write to backend and silverbullet, **read-only** to silverbullet-public           |

**Ephemeral:** anything else inside a container. Notably the Vue build, the
rendered nginx/icecast configs, and the Gitea runner's downloaded binary
are all regenerated.

---

## 5. Environment variables

All configuration lives in **`./.env`** at the repo root, which is gitignored.
`.env.example` documents every variable with blank values — start from that.

Two distinct mechanisms read it, and services use both:

- **`${VAR}` in `docker-compose.yml`** — substituted by the compose CLI at parse
  time. Builds container names and published ports.
- **`env_file: ./.env`** — injects the whole file into the container. Used by
  nginx, backend, db, icecast2, certbot.

`gitea` instead receives explicit `GITEA__section__KEY` variables (Gitea's own
convention for overriding `app.ini` keys), and `python` receives an explicit
allow-list, so it never sees the database password or any API key other than
the two it uses (`TFL_APP_KEY`, and `CLAUDE_API_KEY`, plus `CLAUDE_MODEL`,
which picks the model). The Python service
makes the site's one Claude call (the rowing photo reader); the `jobs` container
(the job tracker, built from its own repo) makes its own, and gets a similar
allow-list with its own database role. The backend receives `CLAUDE_API_KEY`
through `env_file` but does not use it. A new variable for either service has
to be added to its allow-list as well as
`.env.example`.

**`.env` is edited by hand on the host**, in
`~/deploy/web_server`. Nothing syncs it from a laptop checkout, so a laptop's
`.env` (often with dev values) never reaches production by accident.

An unset variable is not an error: `envsubst` renders it as an empty string,
which for a password field means _no password_. After editing `.env`, confirm
with `docker compose config`.

### Changing a value on the host

Edit `.env` on the host, and recreate only the services that read the variable:
`docker compose up -d <service>`. A plain `up -d` would also restart Postgres,
since nginx, db, icecast2 and certbot load the whole file.

---

## 6. Rate limits

Defined once per config in the `http` block, applied per location. Keyed on
`$binary_remote_addr` — the client IP in packed form, so each entry is ~64 bytes
and a `10m` zone holds roughly 160k IPs. `10m` is a safe round number, not a
tuned value.

`rate` is a **smoothed leaky bucket**, not a per-window counter: `5r/m` means one
request every 12 seconds, not five at once. `burst=N` lets N requests queue ahead
of that drip, and `nodelay` (used everywhere here) serves those N immediately.
So the effective behaviour is _"N instant requests, then throttled to the drip
rate"_. Anything beyond the burst gets an immediate **503**.

| Zone           | Rate   | Burst | Applied to                  | Why                                                                                                                                                                                                    |
| -------------- | ------ | ----- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `login`        | 5/min  | 3     | `/api/auth/login`           | Brute-force defence. A human mistyping gets 3 tries then one per 12s; a credential-stuffing script gets nothing useful.                                                                                |
| `api`          | 30/sec | 20    | `/api/`, `/py/`             | Generous on purpose — one SPA page load fans out to several calls. An abuse ceiling, not a quota.                                                                                                      |
| `graphql`      | 10/sec | 10    | `/api/graphql`              | Tighter than REST because one query can be arbitrarily expensive. **Production config only**; dev folds GraphQL into `api`.                                                                            |
| `upload`       | 30/min | 20    | `/api/messages/upload`      | Uploads hit disk; bounds how fast one IP fills the volume. Generous because uploaders must be signed in (accounts are admin-created) and a batch of pictures dropped on a canvas page uploads at once. |
| `hotel_search` | 3/min  | 2     | `/py/hotels/(search\|scan)` | Each request holds an upstream call for minutes and the Python side has one search slot — concurrency protection more than abuse protection.                                                           |

The WebSocket endpoint is deliberately unlimited: a socket is one long-lived
request, and a limit there would drop reconnect storms rather than abuse.

These zones key on `$remote_addr`. For tunnel traffic nginx rewrites that
from Cloudflare's `CF-Connecting-IP` header (`set_real_ip_from` /
`real_ip_header` in `nginx.conf.template`), trusting the header only from
cloudflared's fixed address. **If that address changes in
`docker-compose.yml` without the nginx line, every tunnel visitor shares
cloudflared's one bucket**, and the login limit locks everyone out together.

---

## 7. Runbook

All commands run from the deploy checkout on the host
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

Or browse every container's logs at `https://www.<DOMAIN>/dozzle/` (admin
login required).

**Incoming request bodies.** nginx writes one JSON line per request, body
included, to `logs/nginx/payloads.log` (`nginx/payload_log.conf`). The
`traffic` container tails it, so in Dozzle pick **traffic** and each request
renders as fields you can search. From a shell:

```
tail -f logs/nginx/payloads.log | jq 'select(.method=="POST")'
jq -r 'select(.uri=="/api/graphql") | .body | fromjson? | .operationName' logs/nginx/payloads.log
```

Bodies mentioning `password`, `token` or `secret` are logged as `[redacted]`
(that covers every login and the refresh call). Bodies over 64 KB, i.e.
uploads, are logged with an empty `body`, and responses are never logged.

The file is never rotated by the stack. On the host, add
`/etc/logrotate.d/web_server`:

```
/home/adamf/deploy/web_server/logs/nginx/*.log {
    weekly
    rotate 4
    compress
    missingok
    notifempty
    copytruncate
}
```

`copytruncate` because nginx keeps the file open; `tail -F` in `traffic`
follows the truncation.

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

**Garage (the `/uploads/` bucket).** Public uploads can live in a Garage
bucket instead of the `uploads` volume. It is opt-in and every step can be
undone; nothing breaks while Garage is down or not set up:

- the backend writes to disk until `S3_ENDPOINT` is set, and falls back to
  disk (with a log line) whenever Garage refuses a write;
- nginx serves `/uploads/` from the volume first and asks Garage only for
  files that are not there, and it boots without the `garage` container.

Turning it on:

1. Fill in `GARAGE_RPC_SECRET`, `GARAGE_ADMIN_TOKEN`, `S3_ACCESS_KEY_ID` and
   `S3_SECRET_ACCESS_KEY` in `.env` **before** the deploy that adds the
   containers (otherwise `garage` crash-loops on a blank secret until they
   are set; nothing else is affected). `.env.example` says how to generate
   each.
2. Deploy, then `garage/bootstrap.sh` (idempotent; it creates the layout, the
   `uploads` bucket and the backend's key, and opens the bucket to the
   website endpoint).
3. Set `S3_ENDPOINT=http://garage:3900` and `docker compose up -d backend`.
   The backend logs `[upload] storing uploads in bucket uploads ...` on start.
4. Check: upload something through a canvas page or chat, then
   `docker compose exec garage /garage bucket info uploads` shows the object
   count going up, and the returned `/uploads/...` URL loads.

Moving the old files in (optional — they are served from the volume anyway):

```
export $(grep -E '^S3_(ACCESS_KEY_ID|SECRET_ACCESS_KEY)=' .env)
docker run --rm --network web_server_app-network \
  -v web_server_uploads:/data:ro \
  -e AWS_ACCESS_KEY_ID="$S3_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$S3_SECRET_ACCESS_KEY" \
  -e AWS_DEFAULT_REGION=garage \
  amazon/aws-cli:2.38.0 s3 sync /data s3://uploads --exclude 'private/*' \
  --endpoint-url http://garage:3900
```

The copies on the volume keep winning (disk first) until they are deleted,
which is also the rollback: leave them until the bucket has been trusted for a
while, then remove everything but `private/` from the volume. Until then,
deleting a file in garage-webui does not unpublish it if a disk copy exists.

**Rollback:** blank `S3_ENDPOINT` and `docker compose up -d backend`. Files
written to the bucket in the meantime stay reachable as long as `garage` runs;
to drop Garage entirely, first copy them back
(`aws s3 sync s3://uploads /data` with the volume mounted read-write).

**Managing files:** garage-webui at `/garage/` (site admin login). Files
uploaded there are public at `/uploads/<path>` and skip the backend's type
checks; keep them under a prefix such as `media/`. Uploads through it are
capped at 2G by nginx.

**Backups:** Garage here is one node on the same disk — it is not a backup.
Back up `garage_meta` and `garage_data` together (meta without data, or the
reverse, is useless); Garage also snapshots its metadata every 6h inside
`garage_meta`.

**Set up the CI runner on a fresh host:**

```
cd gitea-runner
bash download.sh                                  # pulls the binary for this CPU
# mint a registration token: Gitea -> Site Administration -> Actions -> Runners
GITEA_RUNNER_REGISTRATION_TOKEN=... ./run.sh      # registers once, then daemonises
```

The runner lives **outside** Docker on purpose: it runs `docker compose up`
against this stack, so it cannot be a container in it. Registration is one-shot —
`.runner` is the lasting credential; re-registering needs that file deleted _and_
a fresh token.

**Go live on the radio:** point a broadcasting client (e.g. butt) at the host's
Tailscale name (not the domain, which resolves to Cloudflare) on
`${LIQUIDSOAP_HARBOR_PORT}` with mount `${LIQUIDSOAP_HARBOR_MOUNT}` and the
icecast source password. liquidsoap pre-empts the fallback playlist immediately
(`track_sensitive=false`). Disconnect and it falls back to the playlist, then to
silence. Adding MP3s to `icecast2/fallback_music/` is picked up live
(`reload_mode="watch"`) with no restart.

**Unhealthy containers are not restarted automatically.** `backend` and `python`
define healthchecks, but the `autoheal` watchdog that acted on them was retired
(`.deprecated/autoheal`), so `unhealthy` is only visible in `docker ps`.
`restart: always` still brings back a container whose process exits.

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
4. **The `/uploads/` location blocks (with `@uploads_bucket`) are still written twice** — once in
   `nginx.conf.template` and once in `nginx_dev_common.conf.template` —
   because production adds security headers inside an auth-gated block. Every
   other proxied route lives once in `nginx_routes.conf.template`, which both
   configs include; the `auth_request` lines on `/uploads/private/` are the
   one thing that must be kept in step by hand.
5. **`nginx.conf.template` has a `/img/stamps/mine.gif` CORS header allowing
   only our own origin**, which defeats the point of having the header at all.
6. **`certbot/entrypoint.sh` uses unquoted `${EMAIL}` and `${DOMAIN}`.** Inert
   today (neither contains whitespace) but a latent quoting bug.
7. **`UPTIMEKUMA_*` and `WALLABAG_*` are substituted by nginx but no compose
   service provides them** — leftovers from removed services.
