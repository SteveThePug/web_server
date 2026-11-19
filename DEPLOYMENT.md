# Deployment

How this site is put together and how to operate it. Application code is
documented in `CLAUDE.md` and `README.md`; this file covers infrastructure only
— compose, the Caddy edge, TLS, the runner, and the runbook.

Everything runs on a single host, `stpvm` (an x86-64 Ubuntu VM on Proxmox,
reached over Tailscale as `stpvm.tailf20278.ts.net`; until 2026-10-15 it was
the Raspberry Pi `stppi`), as Docker containers on one bridge
network, behind one Caddy (nginx and certbot until 2026-11-17).

---

## 1. Service topology

Only three things are reachable from outside the host: Caddy on 80/443, Gitea on
3000/2222, and the liquidsoap harbor port. Everything else is internal. Web
traffic can also come in through the Cloudflare tunnel (`cloudflared`), which
dials out and publishes no port; see "Cloudflare Tunnel" in §3.

```
                 internet
                     |
            :80  :443|                        :3000 / :2222      :8001
                     v                              v               v
              +--------------+                 +--------+    +-------------+
              |    caddy     |                 | gitea  |    |  liquidsoap |
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

  vue      one-shot build container -> writes vue_dist volume -> caddy serves it
```

### Who listens where

| Service                                       | Container name                | Internal port                 | Public route                                                                                            | Published to host    |
| --------------------------------------------- | ----------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------- |
| caddy (the edge)                              | `caddy`                       | 80, 443 (fixed IP `172.28.255.251`) | everything below                                                                                  | **80, 443**          |
| backend (Go)                                  | `${BACKEND_HOST}`             | `${BACKEND_PORT}` (8080)      | `${BACKEND_ENDPOINT}` = `/api/`                                                                         | no                   |
| db (Postgres 16)                              | `${POSTGRES_HOST}`            | 5432                          | none                                                                                                    | no                   |
| silverbullet (editor)                         | `${SILVERBULLET_HOST}`        | `${SILVERBULLET_PORT}`        | `/sb/` (admin-gated)                                                                                    | no                   |
| silverbullet-public (notes site)              | `${SILVERBULLET_PUBLIC_HOST}` | `${SILVERBULLET_PUBLIC_PORT}` | `/notes/` (**public**, read-only)                                                                       | no                   |
| silverbullet-git                              | n/a                           | n/a                           | none (timer loop)                                                                                       | no                   |
| python (FastAPI)                              | `${PYTHON_HOST}`              | `${PYTHON_PORT}`              | `/py/`, docs at `/py/docs`                                                                              | no                   |
| jobs (job tracker, published image)           | `jobs`                        | 8080                          | `/jobs/` (admin-gated by Caddy `forward_auth`; capture API and userscript public, bearer-token)       | no                   |
| activity-tracker (published image, adamf/ccp) | `activity-tracker`            | 8080                          | `/tracker/`, prefix stripped (pages admin-gated by Caddy `forward_auth`; `/api/v1/` bearer/app-checked) | no                   |
| icecast2 + liquidsoap                         | `${ICECAST_HOST}`             | `${ICECAST_PORT}`             | `/radio/`                                                                                               | **harbor port only** |
| gitea                                         | `${GITEA_HOST}`               | 3000, 2222                    | `/gitea/`                                                                                               | **3000, 2222**       |
| vue                                           | `vue`                         | 5173 (dev only)               | `/` (dev only)                                                                                          | no                   |
| cloudflared                                   | `cloudflared`                 | — (fixed IP `172.28.255.250`) | all of Caddy, via Cloudflare                                                                            | no (outbound only)   |
| dozzle (log viewer)                           | `dozzle`                      | 8080                          | `/dozzle/` (admin-gated by Caddy `forward_auth`; reads the Docker socket)                               | no                   |
| receipt printer (`receipt serve`, adamf/receipt_maker) | — (runs on **stppi**, not in this stack) | 8765 on stppi's Tailscale IP (`RECEIPT_WEB_UPSTREAM`) | `/receipt/`, prefix stripped (permission-gated on `receipt` by Caddy `forward_auth`; Caddy adds `X-Receipt-Token` and `X-Remote-User`). The only upstream off this host; 503 "offline" when stppi is down | no |
| Partner Radar (adamf/cto, its own compose project) | `radar-web` (alias of that project's `web`, on app-network) | 80 | `radar.<DOMAIN>`, the whole host (its own login, no gate; zones `radar` and `radar_login`). Started from the cto checkout with `make stpvm`; 503 "offline" when that stack is down | no |
| traffic (tails the payload log)               | `traffic`                     | — (no network)                | none; its stdout is what Dozzle shows                                                                   | no                   |
| autoheal (restarts unhealthy caddy)           | `autoheal`                    | — (no network)                | none; reads the Docker socket                                                                           | no                   |
| garage (S3 object store)                      | `garage`                      | 3900 S3, 3902 web, 3903 admin | none directly; `/uploads/` falls back to its web endpoint for files not on the volume                   | no                   |
| garage-webui (Garage admin UI)                | `garage-webui`                | 3909                          | `/garage/` (admin-gated by Caddy `forward_auth`)                                                        | no                   |

Container names come from `.env`, and Caddy builds its upstreams from the same
variables (passed through the caddy service's `environment:` list). **Renaming
a `*_HOST` value renames the container and repoints Caddy in one step — never
change one without the other.**

### Things Caddy does that are easy to miss

The config is `caddy/Caddyfile` (global options and snippets), `caddy/routes.caddy`
(every proxied route, shared by dev and production) and `caddy/sites/`
(`prod.caddy`, `dev.caddy`). Each file's header explains it.

- **Canonical host.** `http://` → `https://www.<DOMAIN>`, and apex → `www`. The
  SPA, the cookies and the CORS rules all assume `www.<DOMAIN>`.
- **Route order is literal.** `routes.caddy` is one ordered list inside a
  `route` block, and the first matching `handle` wins: denials first, then each
  carve-out ahead of the gated prefix it is carved out of. A new route goes in
  the right place in that list, not just anywhere.
- **Prefix stripping.** `/api`, `/radio`, `/gitea`, `/py`, `/tracker` and
  `/receipt` are stripped with `uri strip_prefix` before proxying; each of
  those upstreams serves from `/`. `/sb/`, `/notes/`, `/jobs/`, `/usaco/`,
  `/dozzle/` and `/garage/` deliberately are NOT: those apps are told their own
  base path (SilverBullet's `SB_URL_PREFIX` builds its service-worker scope from
  it, so stripping the prefix serves a blank page with nothing in the log).
- **Upstreams resolve when dialled.** Caddy starts with any upstream container
  missing and answers 502 until it appears; it never refuses to start over one.
- **Admin gate.** `/sb/`, `/jobs/`, `/tracker/`, `/dozzle/`, `/garage/`, `/radio/admin/`, `/uploads/private/` and
  Gitea's history views and archives (`…/commit/`, `…/src/commit/`, `…/blame/commit/`, `…/archive/`, `…/compare/`; the
  branch views stay public) are
  protected by `forward_auth` to the backend's `/auth/validate-admin` (the
  `admin_gate` snippet); a 401/403 redirects to the SPA login. Most of these services have no auth of their own at that path; the
  job and activity trackers also check the session themselves. Their carve-outs, which a session
  cookie never reaches, stay public at the edge and are guarded by bearer tokens in the apps:
  `/jobs/api/capture` and `/jobs/job-capture.user.js`, `/tracker/api/v1/` and `/tracker/tracker.user.js`.
  Paths are matched cleaned and case-insensitively, so `/SB/` or `/uploads/x/../private/y`
  land on the gate too. With the backend down every gated path answers 502: it fails closed.
- **Permission gate.** `/receipt/` asks `/auth/validate/receipt` instead (the
  `permission_gate` snippet), which also lets in a
  non-admin account granted `receipt` in the admin panel's Manage Users tab. Not signed in
  redirects to the login; signed in without the grant redirects to `/no-access`. To let a
  friend print: create their user (admin panel → User), then tick "Receipt printer" for them
  under Manage Users. Revoking is the same tick, and applies on their next request.
- **Public notes privacy boundary.** `/notes/` is public, and the only thing
  keeping `private/` (and `templates/`, `.git`, `.obsidian`) off it is a pair of
  `path_regexp` deny rules at the top of `caddy/routes.caddy`. SilverBullet's `SB_SPACE_IGNORE` only hides paths from its
  file _list_; direct `/.fs/` fetches and server-rendered page URLs still serve
  them. Adding a folder to `SB_SPACE_IGNORE` on `silverbullet-public` without
  adding it to the deny regex publishes it. `caddy/edge_test.sh` checks these
  denials against a running edge.
- **Uploads: the bucket first, then disk.** `/uploads/<f>` is fetched from
  Garage's website endpoint and, if Garage has no such file or is down, from
  the shared `uploads` volume (the `uploads_disk` snippet, which hands the
  request to a loopback-only site inside the container so that 206 and 304
  answers keep their status). Both paths add `nosniff`,
  `Content-Disposition: inline` and a deny-all CSP, so user-supplied files
  cannot execute anything on our origin. The bucket is the source of truth;
  a Garage outage only affects files that exist nowhere else. See "Garage"
  in §7.
- **Rate limits** — see §6.

---

## 2. Dev vs prod boot paths

### Production

```
docker compose up --build
```

1. `vue` builds the SPA into the `vue_dist` volume and **exits 0** (expected).
2. `caddy` starts with the image's own command (`caddy run` on
   `/etc/caddy/Caddyfile`). `CADDY_SITE` is unset, so it loads
   `sites/prod.caddy`. It serves at once: the SPA is whatever is in `vue_dist`,
   and upstreams that are not up yet are 502s. If `./caddy/data` holds no
   certificate it obtains one first (§3), and TLS fails for those seconds.
3. Everything else comes up in parallel; `depends_on` here is the plain list
   form, which waits only for _start_, not readiness.

On the host this is triggered automatically by `.gitea/workflows/deploy.yaml`,
which has two jobs:

- **`test`** runs on every push to any branch. It exports the pushed commit
  (never the deploy checkout) and runs the backend's `go vet` + `go test` in
  `golang:1.26` and the Python tests in `python:3.12-slim`. The Vue build and
  the Caddy config check are not repeated here: the image builds in the deploy
  step do them (`caddy/Dockerfile` runs `caddy validate`), and a failed build
  leaves the old containers running.
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
cp .env.dev .env
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

`.env.dev` is committed and holds only dummy values, so the stack starts
without any real credential and cannot reach a real account. Features that
need one fail at the provider, which is the point. It must become `.env`
rather than be passed with `--env-file`, because several services read
`env_file: ./.env` directly. Its Postgres passwords only match a `dbdata`
volume created with them, and as in production only `POSTGRES_DB` is created
automatically.

The override layer changes these services and disables `cloudflared`:

- **vue** — `npm run dev` instead of a one-shot build; `./vue` is bind-mounted
  for hot reload, with an anonymous-volume mask on `/app/node_modules` so the
  image's installed deps are not hidden by the bind mount. You don't need
  `node_modules` on the host.
- **caddy** — `CADDY_SITE=dev`, which loads `caddy/sites/dev.caddy`: `/`
  proxies to the Vite dev server, both `:80` and `:443` serve the whole site,
  no HSTS or other security headers, no HTTP→HTTPS redirect. HTTPS uses a
  certificate from Caddy's own local CA (kept in `./caddy/data`), which the
  browser warns about. Every proxied route, gate and limit is the same
  `routes.caddy` as production.
- **backend** — `DEV_MODE`, GraphQL playground and introspection on, a localhost
  Spotify redirect URI, and **`SEED_DB=true`**.
- **python** — `uvicorn --reload` with `python/app` bind-mounted.
- **jobs** — seeded, with the GraphQL playground on.
- **cloudflared** — put in the `disabled` profile, so it never starts: a dev
  stack must never join the production tunnel.

Visit `http://localhost`.

**`SEED_DB`** makes the Go backend populate the database with test posts, users
and messages at startup. It is set only in the dev override and must never be
set on the host — it would inject fake content into the live site.

Frontend-only work can skip Docker: `cd vue && npm run dev`, which proxies
`/api`, `/gitea` and `/radio` to locally running services.

---

## 3. TLS: certificates

Caddy obtains and renews its own certificates (`prod_tls` in
`caddy/sites/prod.caddy`): one per name, `<DOMAIN>` and `www.<DOMAIN>`, from
Let's Encrypt over **DNS-01**. It creates the `_acme-challenge` TXT record
through the Cloudflare API with `CLOUDFLARE_DNS_API_TOKEN`, so nothing has to
reach the host: the router forwards no ports. The certificates, their keys and
the ACME account live in `./caddy/data` (gitignored, private key material).

**First run on a new host or domain.** Set `CLOUDFLARE_DNS_API_TOKEN` (the
token needs **Zone → Zone → Read** and **Zone → DNS → Edit** on this zone; see
`.env.example`), then `docker compose up -d`. Caddy serves straight away and
obtains both certificates in the background, usually within seconds; until
they land the TLS handshake fails, so the tunnel answers 502. There is no
bootstrap config and nothing to restart.

```
docker compose logs caddy | grep -E 'obtaining certificate|certificate obtained|error'
```

**Renewal** happens in-process when a third of a certificate's lifetime is
left, and the new one is served immediately, with no reload. A failed attempt
is retried with backoff and logged. To see what is being served:

```
for n in www.$DOMAIN $DOMAIN; do echo | openssl s_client -connect 127.0.0.1:443 -servername $n 2>/dev/null | openssl x509 -noout -subject -enddate; done
```

**`resolvers 1.1.1.1 1.0.0.1`** in `prod_tls` matters: on `app-network`
`<DOMAIN>` is an alias for the caddy container itself, so Docker's DNS would
answer the plugin's zone lookup and the order would fail with "expected 1
zone, got 0".

**Losing `./caddy/data`** costs only a re-issue on the next start. Let's
Encrypt allows five certificates for the same name a week, so do not make a
habit of it.

Until 2026-11-17 this was certbot plus a bootstrap nginx config, a reload
sentinel and a watcher loop, now in `.deprecated/`. The last certbot
certificate is still in `./certbot/conf` on the host as the rollback (§7); it
expires on 2027-01-20.

### Cloudflare Tunnel

The `cloudflared` container holds an outbound connection to Cloudflare and
forwards requests for the site's hostnames down it to Caddy. It is a
_remotely managed_ tunnel: the only thing in this repo is the token
(`CLOUDFLARE_TUNNEL_TOKEN` in `.env`); hostnames and
origin settings live in the dashboard.

**Origin: `https://www.<DOMAIN>`, the site's own name.** Cloudflare
terminates the visitor's TLS at its edge, and cloudflared opens a second TLS
connection to Caddy. That is deliberate. Pointed at `http://caddy:80`, the
tunnel would hit the HTTP→HTTPS redirect and loop. Going to :443 keeps every
redirect, security header and the `X-Forwarded-Proto: https` the backend's
Secure cookies depend on, and needs no change to the edge.

The origin is the public name, not `caddy`, because Caddy carries `<DOMAIN>`
and `www.<DOMAIN>` as network aliases on `app-network` (`docker-compose.yml`).
Inside the network those names resolve to the caddy container, so the hop
never leaves the host, and cloudflared verifies the Let's Encrypt certificate
against a name it covers. `https://caddy` would fail that check and need No
TLS Verify. The flip side: **an expired certificate takes the tunnel down
with a 502**, which is why renewal over DNS-01 (step 5) matters.

An alias only exists while Caddy is running, so cloudflared also pins both
names to Caddy's fixed address (`172.28.255.251`) with `extra_hosts`. Without
that, while Caddy is stopped for a deploy or the weekly maintenance, the name
falls through to public DNS, resolves to Cloudflare itself, and visitors see
Cloudflare error 1000 instead of a 502.

The aliases have a side effect for anything else on `app-network` that uses
the public name for a port Caddy does not serve. `silverbullet-git` pushes to
Gitea's SSH on `<DOMAIN>:2222`, so it maps `<DOMAIN>` to the host gateway
instead, where Gitea publishes that port.

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
    - `<DOMAIN>` → `https://<DOMAIN>`: Caddy redirects it to www, and it serves Gitea's registry
      at `/v2/`.
    - `radar.<DOMAIN>` → `https://radar.<DOMAIN>`: the Partner Radar demo
      (Caddy's `radar.{$DOMAIN}` site, `sites/prod.caddy`).
      Creating each hostname replaces the existing A record for that name with
      a CNAME to the tunnel. **That is the cutover**: until then the tunnel
      carries nothing.

5. Create a Cloudflare API token (My Profile → API Tokens → "Edit zone
   DNS", this zone only), add **Zone → Zone → Read** to it, and set it as
   `CLOUDFLARE_DNS_API_TOKEN`. Caddy then obtains and renews over DNS-01
   (§3) and needs no inbound port.
6. On the host, point the apex at Caddy locally (next section).
7. Close the router's 80 and 443 forwards. Once 5 and 6 are done nothing
   depends on them.

#### What the tunnel does not carry, and where it goes instead

- **Certificate renewal.** Over DNS-01 only: Caddy is given no HTTP
  challenge to fall back to, so without a working `CLOUDFLARE_DNS_API_TOKEN`
  nothing is issued or renewed. The certificate matters behind the tunnel:
  cloudflared verifies it, as do the registry path below and any direct LAN
  access. A fresh machine with no certificate serves nothing over 443 until
  the first issuance (§3), so the tunnel carries nothing for those seconds.
- **Registry pushes and pulls on the host.** Cloudflare caps request bodies at
  100 MB on the free plan, so a push of a larger image layer through the
  tunnel fails. The runner on the host pushes to `<DOMAIN>/adamf/...` and the
  stack pulls from it, both through the host's Docker daemon, which resolves
  names through the host's `/etc/hosts`. Pin the apex there to the loopback:

  ```
  127.0.0.1   <DOMAIN>
  ```

  so registry traffic goes straight to Caddy on :443, with the real
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
| `caddy_config` | Caddy's autosaved JSON config                     | Nothing — rewritten on every start                 |

**Bind mounts** — live in the repo checkout on the host, survive everything short
of deleting the directory:

| Path                        | Holds                                                                                               | Notes                                                                                  |
| --------------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `./caddy/data`              | Caddy's TLS certificates, **private keys** and ACME account                                         | Gitignored. Never commit. Losing it means a re-issue on the next start (§3).           |
| `./certbot/conf`            | The old certbot certificate and keys                                                                | Unused since 2026-11-17; kept as the rollback until it is cleaned up (§7).             |
| `./backend/token/`          | **Spotify OAuth refresh token** (`spotify_token.json`)                                              | Gitignored. Without this mount, Spotify would need re-authorising after every rebuild. |
| `./jobs-data`               | The job tracker's **capture token hash** and Graph OAuth session                                    | Gitignored. Must be owned by uid 1000, the user the `jobs` image runs as.              |
| `./gitea/data`              | Git repositories, LFS objects, avatars, sessions                                                    | The actual git data                                                                    |
| `./gitea/config`            | Gitea's live `app.ini`                                                                              | Gitignored; generated from the template on first boot and then owned by Gitea          |
| `./logs`                    | Backend logs (`go.log`); Caddy's JSON request-body log in `logs/caddy/payloads.log`                  | Caddy rotates the payload log itself (50 MiB, five kept). `logs/nginx/` is nginx's, finished. |
| `./icecast2/fallback_music` | Radio fallback MP3s                                                                                 | Copied to the host by hand, not git                                                    |
| `${OBSIDIAN_DIR}`           | The Obsidian vault (a git clone of the notes repo)                                                  | Read-write to backend and silverbullet, **read-only** to silverbullet-public           |

**Ephemeral:** anything else inside a container. Notably the Vue build, the
rendered icecast config, and the Gitea runner's downloaded binary
are all regenerated.

---

## 5. Environment variables

All configuration lives in **`./.env`** at the repo root, which is gitignored.
`.env.example` documents every variable with blank values — start from that.

Two distinct mechanisms read it, and services use both:

- **`${VAR}` in `docker-compose.yml`** — substituted by the compose CLI at parse
  time. Builds container names and published ports.
- **`env_file: ./.env`** — injects the whole file into the container. Used by
  backend, db and icecast2.

`gitea` instead receives explicit `GITEA__section__KEY` variables (Gitea's own
convention for overriding `app.ini` keys), and `python` receives an explicit
allow-list (as do `jobs`, `caddy` and the rest), so it never sees the database password or any API key other than
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

An unset variable is not an error: compose substitutes an empty string,
which for a password field means _no password_. After editing `.env`, confirm
with `docker compose config`.

### Changing a value on the host

Edit `.env` on the host, and recreate only the services that read the variable:
`docker compose up -d <service>`. A plain `up -d` would also restart Postgres,
since backend, db and icecast2 load the whole file.

---

## 6. Rate limits

The `rate_limit` directive comes from the `mholt/caddy-ratelimit` plugin
(compiled in by `caddy/Dockerfile`) and is applied per route through the
`limit` snippet: `import limit <zone> <events> <window>`. Each zone is a
**sliding window** keyed on `{client_ip}`: at most `<events>` requests per
`<window>` per IP, then the request is refused. There is no separate burst, so
the numbers are nginx's old rates, which keeps the sustained rates the same;
the instant allowance comes out close to nginx's burst + 1.

A refused request gets **503**, not the plugin's 429: `handle_errors 429` in
the Caddyfile rewrites it, because `vue/src/views/hotels/hotelsStream.js`
tells a rate-limit 503 ("too many searches from your address") apart from the
Python service's own 429 ("someone else is searching").

A zone's state is shared by every route that names it (`/api/`, `/py/`,
`/jobs/`, `/usaco/` and `/tracker/` share one `api` bucket per IP), so a zone
name must always be given the same numbers. A route that wants different ones
gets its own zone, as `/receipt/` does.

| Zone           | Limit     | Applied to                                   | Why                                                                                                                                                                                                    |
| -------------- | --------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `login`        | 5 / 1m    | `/api/auth/login`                            | Brute-force defence, on top of the backend's own 5/min. A human mistyping gets five tries a minute; a credential-stuffing script gets nothing useful.                                                  |
| `api`          | 30 / 1s   | `/api/`, `/py/`, `/jobs/`, `/usaco/`, `/tracker/` | Generous on purpose — one SPA page load fans out to several calls. An abuse ceiling, not a quota.                                                                                                 |
| `graphql`      | 10 / 1s   | `/api/graphql`                               | Tighter than REST because one query can be arbitrarily expensive. Dev has it too now.                                                                                                                  |
| `upload`       | 30 / 1m   | `/api/uploads`                               | Bounds how fast one IP fills the bucket or the volume. Generous because uploaders must be signed in (accounts are admin-created) and a batch of pictures dropped on a canvas page uploads at once. |
| `hotel_search` | 3 / 1m    | `/py/hotels/(search\|scan)`                  | Each request holds an upstream call for minutes and the Python side has one search slot — concurrency protection more than abuse protection.                                                           |
| `receipt`      | 10 / 1s   | `/receipt/`                                  | Prints are slow and physical; ten a second is far above any real use.                                                                                                                                  |
| `radar`        | 30 / 1s   | all of `radar.<DOMAIN>`                      | The Partner Radar demo: an abuse ceiling like `api`, in its own zone because it is a different app. |
| `radar_login`  | 5 / 1m    | `radar.<DOMAIN>/api/auth/login`              | Brute-force defence on the demo's login, on top of `radar` and the app's own login throttling. |

The WebSocket endpoint is deliberately unlimited: a socket is one long-lived
request, and a limit there would drop reconnect storms rather than abuse.

`{client_ip}` is the visitor's address only because of `trusted_proxies` and
`client_ip_headers CF-Connecting-IP` in the Caddyfile's global options, which
trust that header from cloudflared's fixed address alone. **If that address
changes in `docker-compose.yml` without the Caddyfile, every tunnel visitor
shares cloudflared's one bucket**, and the login limit locks everyone out
together.

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

**Apply a Caddy config change.** Everything in `caddy/` is copied into the
image, so a change needs a rebuild, which also validates it (the Dockerfile's
last step runs `caddy validate` for the production and dev sites, and a
config that would not load fails the build before any container is
replaced). Only the first build on a host, or a Caddy or plugin version bump,
compiles Caddy itself (about 25 minutes on stpvm's two CPUs); a config-only
change rebuilds in seconds. The recreate drops connections for a second or
two. Certificates need nothing.

```
make check-caddy                         # on a laptop, before pushing
docker compose up -d --build caddy       # on the host
```

**Check the edge.** `caddy/edge_test.sh` sends a fixed set of requests and
checks the privacy denials and every gate (non-zero exit on a FAIL), then
prints redirects, headers and uploads for comparison. Pass an admin session's
`Cookie` header value as `EDGE_COOKIE` to cover the signed-in cases. On the
host, pin the name to the machine or curl goes out over the internet:

```
export "$(grep '^DOMAIN=' .env)"
caddy/edge_test.sh https://www.$DOMAIN --resolve www.$DOMAIN:443:127.0.0.1
```

**Roll back to nginx** (possible while the certbot certificate in
`./certbot/conf` is valid, until 2027-01-20). The old edge is in
`.deprecated/nginx/` and `.deprecated/certbot/`. Revert the commit that moved
it there and then the cutover (`daa2a46`), which restores both directories
and nginx's compose services. Push, or apply it by hand on the host under the
lock so a deploy cannot interleave:

```
git revert --no-edit <the .deprecated move> daa2a46
flock .compose.lock sh -c 'docker compose stop caddy && docker compose rm -f caddy && docker compose up -d --build nginx certbot cloudflared traffic'
```

**Check logs:**

```
docker compose logs -f caddy           # follow one service (one line per request)
docker compose logs --tail=200 backend
docker compose ps                      # note: `vue` Exited(0) is correct
docker compose logs caddy | grep -E 'tls|certificate'   # issuance / renewal
```

Or browse every container's logs at `https://www.<DOMAIN>/dozzle/` (admin
login required).

**Incoming request bodies.** Caddy writes one JSON line per request, body
included, to `logs/caddy/payloads.log` (the `payloads_file` logger and the
`logging` snippet in `caddy/Caddyfile`). The `traffic` container tails it, so
in Dozzle pick **traffic** and each request renders as fields you can search.
From a shell:

```
tail -f logs/caddy/payloads.log | jq 'select(.request.method=="POST")'
jq -r 'select(.request.uri=="/api/graphql") | .req_body | fromjson? | .operationName' logs/caddy/payloads.log
jq -r '[.ts, .request.client_ip, .request.method, .request.uri, .status] | @tsv' logs/caddy/payloads.log
```

Bodies mentioning `password`, `token` or `secret` are logged as `[redacted]`
(that covers every login and the refresh call), and `Cookie`/`Authorization`
headers as `REDACTED`. A body is only logged when the client declared a
length under 100 KB on a route that is not an upload (`/api/uploads`,
`/garage/`, `/receipt/`, `/v2/`); everything else has no `req_body`, and
responses are never logged.

Caddy rotates the file itself (50 MiB, five old files kept), and `tail -F`
in `traffic` follows the rotation, so it needs no logrotate entry. The
`logs/nginx/*.log` stanza in `/etc/logrotate.d/web_server` is nginx's and
can go when `logs/nginx/` does.

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

**Garage (the `/uploads/` bucket).** Public uploads live in a Garage bucket,
with the `uploads` volume as a safety net. Nothing breaks while Garage is down
or not set up:

- the backend writes to disk until `S3_ENDPOINT` is set, and falls back to
  disk (with a log line) whenever Garage refuses a write;
- Caddy serves `/uploads/` from Garage and falls back to the volume when
  Garage has no such file or is unreachable, and it starts without the
  `garage` container.

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

Moving the old files in (so they show in garage-webui, and so the bucket
holds everything):

```
export $(grep -E '^S3_(ACCESS_KEY_ID|SECRET_ACCESS_KEY)=' .env)
docker run --rm --network web_server_app-network \
  -v web_server_uploads:/data:ro \
  -e AWS_ACCESS_KEY_ID="$S3_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$S3_SECRET_ACCESS_KEY" \
  -e AWS_DEFAULT_REGION=garage \
  amazon/aws-cli:2.38.0 s3 sync /data s3://uploads --exclude 'private/*' \
  --endpoint-url http://garage:3900
```

The copies left on the volume are then only a fallback for when Garage is
down, and the rollback. One catch: a file deleted in garage-webui is still
served from its disk copy, if it has one. Once the bucket has been trusted for
a while, remove everything but `private/` from the volume.

**Rollback:** blank `S3_ENDPOINT` and `docker compose up -d backend`. Files
written to the bucket in the meantime stay reachable as long as `garage` runs;
to drop Garage entirely, first copy them back
(`aws s3 sync s3://uploads /data` with the volume mounted read-write).

**Managing files:** garage-webui at `/garage/` (site admin login). Files
uploaded there are public at `/uploads/<path>` and skip the backend's type
checks; keep them under a prefix such as `media/`. Uploads through it are
capped at 2 GB by Caddy.

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

**Dependency updates (Renovate).** Dependabot is GitHub-only, so
`.gitea/workflows/renovate.yaml` runs Renovate on the same runner every Monday
at 04:00 UTC (or by hand from the Actions tab), configured by
`renovate.json5`. It opens PRs from `renovate/*` branches, which get the
normal `test` job; nothing automerges, because merging to `main` deploys. It
needs a `RENOVATE_TOKEN` repo Actions secret: create a `renovate` Gitea user,
add it to this repo as a collaborator with write access, and mint a token for
it with `read:user`, `read:organization`, `write:repository` and
`write:issue`. `RENOVATE_GITHUB_COM_TOKEN` (any github.com token, no scopes)
is optional and lifts GitHub's rate limit on release-note lookups. The first
run opens a "Dependency Dashboard" issue listing everything pending.

**Go live on the radio:** point a broadcasting client (e.g. butt) at the host's
Tailscale name (not the domain, which resolves to Cloudflare) on
`${LIQUIDSOAP_HARBOR_PORT}` with mount `${LIQUIDSOAP_HARBOR_MOUNT}` and the
icecast source password. liquidsoap pre-empts the fallback playlist immediately
(`track_sensitive=false`). Disconnect and it falls back to the playlist, then to
silence. Adding MP3s to `icecast2/fallback_music/` is picked up live
(`reload_mode="watch"`) with no restart.

**Only caddy is restarted when unhealthy.** Docker marks a container
`unhealthy` but never acts on it; `restart: always` only brings back a process
that exits. The `autoheal` container restarts anything labelled
`autoheal=true`, which today is just caddy (its healthcheck fetches
`http://127.0.0.1/caddy-health` every 30s; three misses, about 90s, and it is
restarted). `backend` and `python` also define healthchecks, but carry no
label, so for them `unhealthy` is only visible in `docker ps`. Add the label to
cover another service.

**Site down but every container `Up`.** That is what a stalled edge looks
like (2026-10-19, under nginx: 4.6 hours, every tunnel request a `TLS
handshake timeout` in `docker logs cloudflared`, and the Cloudflare dashboard
still showing the tunnel healthy, since that only covers the tunnel's own
link). The healthcheck should catch it. If it happens, capture what Caddy is
doing before restarting, since a restart destroys the only evidence of why.
Its admin API (inside the container only) serves Go's profiler:

```
docker inspect caddy --format '{{.State.Health.Status}}'
docker compose exec caddy wget -q -O - 'http://localhost:2019/debug/pprof/goroutine?debug=2' > caddy-goroutines.txt
docker compose restart caddy
```

---

## 8. Known rough edges

Documented, not fixed — each would change deploy behaviour.

1. **The old edge's host directories are still on stpvm** as the rollback
   from the 2026-11-17 cutover: `./certbot/` (the last certbot certificate)
   and `./logs/nginx/`, plus the `logs/nginx` stanza in
   `/etc/logrotate.d/web_server`. The code is in `.deprecated/`. Once the
   rollback is no longer wanted, delete the two directories and the stanza.
2. **`gitea/config/entrypoint.sh` is an empty, unused file** that would be
   mounted to the same in-container path as the real wrapper. It is currently
   shadowed by an explicit file mount — fragile. Deleting it would be safe but
   is a file removal.
