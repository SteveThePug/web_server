# Backend

The Go API for adam-french.co.uk. It is one process behind nginx, serving a
GraphQL endpoint, a handful of REST endpoints, and the chat WebSocket, backed
by PostgreSQL and a few third-party APIs (Spotify, Steam, Gitea, Anthropic).

Nginx strips the `/api` prefix before proxying, so a browser request to
`/api/graphql` arrives here as `/graphql`. Every path below is as this service
sees it.

## Layout

```
main.go              wiring: env -> services -> Store -> routes
models/models.go     GORM entities, which are ALSO the GraphQL object types
handlers/            REST endpoints + the shared Store dependency container
  store.go           Store: DB, service handles, and the third-party caches
  handle_*.go        one file per domain
services/            stateful collaborators (DB, auth, chat hub, API clients)
graph/               GraphQL layer
  schema/*.graphql   source of truth for the schema
  *.resolvers.go     resolver implementations, one per schema file
  *_helpers.go       hand-written mapping helpers (never regenerated)
  context.go         reading identity out of the resolver context
  middleware.go      putting it in
  doc.go             package overview — read this before touching resolvers
  generated.go       GENERATED, do not edit
  model/models_gen.go GENERATED, do not edit
```

## Request lifecycle

1. Nginx terminates TLS, applies its own rate limits, strips `/api`, and
   proxies to this service.
2. Gin routes the request. `SetTrustedProxies` pins the Docker network CIDR so
   that `ClientIP()` — which the login rate limiter keys on — cannot be
   spoofed via `X-Forwarded-For`.
3. **REST**: the route's middleware chain runs. `AuthMiddlewear` requires a
   valid access token and stashes the claims in the Gin context;
   `AdminMiddleware` then requires `admin: true` in those claims. A handler
   reached through the `admin` group can assume both.
4. **GraphQL**: there is no guarding middleware. `AuthContextMiddleware` only
   *annotates* the request context with the Gin context and, if a valid access
   token is present, the JWT claims — and then lets everything through. Each
   resolver authorises itself with `IsAdminFromCtx` / `UserIDFromCtx`.
   **A resolver that forgets to check is public.**
5. Handlers and resolvers both reach the database and the caches through the
   single `handlers.Store` built in `main`.

## Auth

- `POST /auth/login` (or the `login` mutation) checks a bcrypt hash and sets
  two cookies, both `HttpOnly` + `Secure` + `SameSite=Lax`:
  - `access_token` (7 days) carries `id`, `username`, `admin`, so ordinary
    requests authorise without a database hit;
  - `refresh_token` (365 days) carries only `id`.
- `POST /auth/refresh` (or the `refreshToken` mutation) exchanges the refresh
  token for a new pair, re-reading the user — and therefore the admin flag —
  from the database.
- `GET /auth/validate-admin` answers with a bare status code (200 / 403 / 401)
  and will transparently refresh, so returning admins are recognised without a
  visible re-login.
- Tokens are stateless. **Logout only clears the browser's cookies**; nothing
  is revoked server-side, and a token that has already been issued stays valid
  until it expires. The same is true of demoting an admin: it takes effect on
  their next refresh, not immediately.

The two implementations (REST in `handlers/handle_auth.go`, GraphQL in
`graph/auth.resolvers.go`) are parallel by design, since the front end uses
both. They share `services.Auth` and the rate limiter, but each has its own
cookie-writing helper — change one and you must change the other.

## GraphQL vs REST

GraphQL is the default and covers essentially all the site's data. REST exists
for the things GraphQL cannot do:

| REST endpoint | why it is not GraphQL |
| --- | --- |
| `/auth/*` | cookie handling and the status-code-only admin probe |
| `/spotify/callback`, `/email/callback` | OAuth providers redirect a *browser* here |
| `/messages/upload`, `/radio/upload`, `/rowing` (POST) | multipart file uploads |
| `/notes/*path` | streams a file download, not JSON |
| `/ws` | WebSocket upgrade |
| `/rowing` (GET), `/spotify/listening`, `/spotify/recent` | predate the equivalent queries; kept working |

## Services

Everything is constructed in `main` from environment variables and handed to
`handlers.Store`. Nothing here fails start-up: a missing Spotify token, an
unconfigured mailbox or an absent Steam key all just leave that feature
dormant, which is why callers nil-check `Store.SpotifyClient` and friends.

- **database.go** — Postgres via GORM. There are no migration files:
  `AutoMigrate` is the whole schema story, so it adds tables/columns but never
  drops or renames, and a new model must be listed there to get a table.
- **auth.go** — HS256 JWT minting and verification.
- **ratelimit.go** — in-memory per-IP sliding window (5 logins/minute).
  Process-local and never evicts keys.
- **websocket.go** — the chat hub. A package-level singleton: one connection
  set, one mutex, one DB handle. Public, with admin connections additionally
  able to see/send private messages and delete.
- **spotify.go**, **steam.go**, **gitea.go** — third-party reads. Results are
  cached on `Store` with TTLs of 1 min (Spotify recent, Gitea) and 5 min
  (Steam).
- **email_sync.go** / **email_imap.go** — the job-application pipeline: fetch
  recent mail (Microsoft Graph *or* a hand-rolled IMAP client), keyword-filter
  it, ask Claude to extract structured data, create or advance a
  `JobApplication`. Runs on a timer and can be triggered by an admin.
- **claude.go** — the shared Anthropic client, used by the email pipeline and
  by the rowing photo reader.
- **notes.go** — path-traversal-safe file lookup for the notes directory.
- **seed.go** — dev-only test data. Creates an admin with the password
  `password`, so it must never run in production (`SEED_DB=true` is set only
  in `docker-compose.dev.yml`).

## Regenerating gqlgen

The `.graphql` files are the source of truth.

```
cd backend && go run github.com/99designs/gqlgen generate
```

This rewrites `graph/generated.go` and `graph/model/models_gen.go` and adds
stubs for new fields to the matching `*.resolvers.go`. Existing resolver bodies
are preserved, but hand-written code in a regenerated file can be relocated to
the end of it — which is why shared helpers live in `*_helpers.go` instead.

`gqlgen.yml` maps most GraphQL object types directly onto the GORM structs in
`models/`, so `graph/model/` holds only input and payload types. Adding a type
to the schema without a corresponding entry in `gqlgen.yml` gets you a fresh
generated struct rather than the model you expected.

## Gotchas

- **Soft delete everywhere.** Every model except `ProcessedEmail` has a
  `gorm.DeletedAt`, so `Delete` sets a timestamp and later queries filter the
  row out. Rows are never actually removed, and a `uniqueIndex` still collides
  with deleted rows — a soft-deleted username cannot be reused.
- **`Preload` is not lazy.** `Post.Author` is nil unless the query preloads it;
  there is no error to tell you why the field came back null.
- **IDs.** Models use `uint`, GraphQL uses `Int`. That is why nearly every type
  has a hand-written `ID` resolver doing `int(obj.ID)`.
- **`Message.AuthorID` is not a user id.** It is a per-process counter handed
  out by the chat hub on connect, meaningless across restarts.
- **The chat table is capped.** Every insert soft-deletes all but the newest
  50 messages.
- **Store's caches are unsynchronised** and shared by all requests. The races
  are benign (at worst a duplicate upstream fetch), but do not add a field
  there that would be unsafe to tear.
- **Claude output is parsed as JSON** after stripping a possible markdown code
  fence, in both `handle_rowing.go` and `email_sync.go`.
- **The email pipeline never retries.** A failed email is recorded with action
  `"error"` so the next sync skips it.
- **Authenticating email after start-up does not start the scheduler**, which
  only checks readiness once — restart, or trigger syncs by hand.
- **Introspection and the playground are off** unless `DEV_MODE=true` *and*
  `GQL_INTROSPECTION` / `GQL_PLAYGROUND` are set.
- **No graceful shutdown.** `r.Run` blocks forever and its error is ignored;
  the container is simply killed.

## Environment

| var | purpose |
| --- | --- |
| `POSTGRES_USER` / `_PASSWORD` / `_DB` / `_HOST` / `_PORT` | database connection |
| `BACKEND_PORT` | listen port |
| `BACKEND_SECRET` | JWT signing secret |
| `BACKEND_ENDPOINT` | this service's public base URL |
| `DOMAIN` | auth cookie domain and WebSocket origin allow-list |
| `DEV_MODE` | release vs debug Gin; gates the two GraphQL dev flags |
| `SEED_DB` | seed dev test data (never in production) |
| `GQL_INTROSPECTION` / `GQL_PLAYGROUND` | dev-only GraphQL extras |
| `SPOTIFY_CLIENT_ID` / `_SECRET` / `_REDIRECT_URI` / `SPOTIFY_AUTH_STATE` | Spotify OAuth |
| `CLAUDE_API_KEY` | Anthropic API |
| `GITEA_HOST` / `GITEA_PORT` | Gitea activity feed |
| `STEAM_API_KEY` / `STEAM_ID` | Steam presence |
| `EMAIL_SYNC_ENABLED` / `EMAIL_BACKEND` / `EMAIL_SYNC_INTERVAL` | email pipeline (`graph` or `imap`) |
| `MSGRAPH_CLIENT_ID` / `_SECRET` / `_TENANT_ID` / `_REDIRECT_URI` | Microsoft Graph backend |
| `IMAP_HOST` / `_PORT` / `_EMAIL` / `_PASSWORD` | IMAP backend |

Paths baked into the code, all inside mounted volumes so they survive a
container restart: `/backend/logs`, `/backend/notes`, `/backend/uploads`,
`/backend/fallback_music`, `/backend/token/` (Spotify and Microsoft Graph
OAuth tokens).
