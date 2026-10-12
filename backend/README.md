# Backend

The Go API for adam-french.co.uk. It is one process behind nginx, serving a
GraphQL endpoint, a handful of REST endpoints, and the chat WebSocket, backed
by PostgreSQL and a few third-party APIs (Spotify, Steam, Gitea). Anything
that needs Claude is asked of the Python service, not called from here.

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
  schema/*.graphql   source of truth for the schema, including who may call what
  *.resolvers.go     resolver implementations, one per schema file
  *_helpers.go       hand-written mapping helpers (never regenerated)
  directives.go      @admin / @signedIn: the authorisation checks, written once
  directives_test.go fails when a mutation is added without one
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
   token is present, the JWT claims — and then lets everything through. A
   field is guarded by a directive on it in the schema, `@admin` or
   `@signedIn`, which runs before the resolver. **A field with neither is
   public.** See [GraphQL authorisation](#graphql-authorisation).
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

## GraphQL authorisation

Who may call a field is written in the schema, on the field:

```graphql
extend type Query {
  bookmarks: [Bookmark!]!                    # public
  places: [Place!]! @signedIn                # any signed-in user
  users: [User!]! @admin                     # admins only
}
```

| directive | passes when | error otherwise |
| --- | --- | --- |
| `@admin` | the access token's `admin` claim is true | `admin access required` |
| `@signedIn` | there is a valid access token at all | `authentication required` |
| *(none)* | always | — |

Both are declared in `graph/schema/schema.graphql`, implemented once in
`graph/directives.go`, and handed to gqlgen in `main.go`
(`graph.Config{Directives: graph.Directives()}`). gqlgen calls the directive
before the resolver; on an error the resolver never runs and the client gets
that message for the field. The two strings are shown as-is by the front end,
so they are part of the API. If the `Directives` line were ever dropped from
`main.go` the guarded fields would all fail (`directive admin is not
implemented`), not open.

This replaced an `if !IsAdminFromCtx(ctx) { return ... }` that was pasted into
the top of every guarded resolver, where leaving it out was invisible. Now a
guarded resolver has no check in it, and the absence of a guard is something a
test can see.

**What is guarded.** Queries are public and mutations are `@admin`, except:

- `login`, `logout`, `refreshToken` — public; they are how a session starts
  and ends.
- `createPlace`, `updatePlace`, `deletePlace` and the `places` query —
  `@signedIn`. Places are a shared list with no per-user ownership.
- `users`, `user`, `jobApplications`, `jobApplication`, `jobAppReferences` —
  queries, but `@admin`: private data.

**What a directive cannot say stays in the resolver**, using `IsAdminFromCtx` /
`UserIDFromCtx` from `graph/context.go`:

- `updatePost` / `deletePost` are `@admin` *and* require the caller to be the
  post's author; `setUserAdmin` refuses to change the caller's own flag.
- `me` has no directive: it returns its own `unauthorized` error.
- `messages` has no directive: it filters private messages out for non-admins
  instead of refusing.
- `spotifyNeedsReauth` has no directive: it is `false` for non-admins.

**The safety net** is `graph/directives_test.go`, which walks the parsed
schema:

- every `Mutation` field must carry `@admin` or `@signedIn`, unless it is in
  the `publicMutations` list in that file (the three above). A new mutation
  with no directive fails the build's tests until one or the other is done;
- the `@signedIn` mutations and the guarded queries are pinned by name, so
  loosening `@admin` to `@signedIn`, or dropping a guard from a private query,
  has to be done in the test as well;
- the directive functions are unit-tested, and a handful of operations are run
  through the real executable schema to pin the messages on the wire.

One caveat: the test reads the `.graphql` files, but enforcement is in
`generated.go`. Adding a directive to an existing field does nothing until
gqlgen is re-run.

## Adding an entity

1. **Model** — add the struct to `models/models.go` with `ID uint`
   (`gorm:"primarykey"`) and a `gorm.DeletedAt`, and list it in `AutoMigrate`
   in `services/database.go`. There are no migration files.
2. **Schema** — add `graph/schema/<name>.graphql` with the type, its inputs,
   and `extend type Query` / `extend type Mutation`. Put `@admin` or
   `@signedIn` on every mutation, and on any query that is not for the public.
   Use `#` comments, not `"""` descriptions.
3. **Binding** — map the type onto the model in `gqlgen.yml`, with
   `deletedAt: resolver: false`. Without the entry gqlgen generates a second,
   unrelated struct in `graph/model/`.
4. **Generate** — `go run github.com/99designs/gqlgen generate` (below). This
   creates `graph/<name>.resolvers.go` with a stub per query and mutation.
5. **Resolvers** — fill the stubs in. Do not write an auth check; the
   directive has run. Do not write an `ID` resolver; `ID uint` binds by itself
   (see [IDs](#gotchas)). Put commentary inside the function bodies, since
   gqlgen rewrites what is above the signature. A delete should load the row
   first and return an error for a missing id, as the existing ones do.
6. **Test** — `go test ./...`. `directives_test.go` will tell you if a
   mutation was left unguarded; if you added a guarded *query*, add it to the
   pinned list there.

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
| `/admin/env`, `/admin/logs`, `/admin/capture-token` | process environment and files on disk, not the data model |
| `/jobs/capture`, `/jobs/capture/status`, `/jobs/capture/extract` | called by a userscript on other sites' origins, where the `SameSite=Lax` auth cookies are never sent; they take a bearer token instead |

### Environment tab

`handlers/handle_env.go` and `services/env_manager.go` let the admin panel
change `.env` on the Pi without SSH. The backend never reads or writes the
file and has no Docker access: it validates an edit and leaves it for the
host, which is the part trusted to act.

- **`services/env_registry.json`** lists every variable in `.env.example`
  with its group, kind, whether it is editable, and the compose services that
  read it. It is embedded in the binary and read from the checkout by the host
  script, so one file decides what the panel offers and what the host accepts.
  A test fails if it and `.env.example` disagree.
- **The `python` service is a third place.** It is the one container given an
  explicit `environment:` list in `docker-compose.yml` instead of the whole
  `.env`, so that it never holds the database password or the JWT secret. A
  new Python variable therefore goes in the registry (with `python` in its
  `services`), in `.env.example`, *and* in that list.
  `TestEnvRegistryMatchesComposePythonEnvironment` reads the compose file and
  fails if a variable the registry says python reads is not passed to it, or
  if an editable variable passed to it does not list `python`. Locked
  variables carry no `services` and are exempt (`PYTHON_PORT` is one).
- `GET /admin/env` — the registry plus state from this process's own
  environment. A secret's value is never sent, only "set" and, for values of
  12+ characters, the last four.
- `PUT /admin/env` — `{"set": {NAME: value}, "generate": [NAME]}`; `""`
  clears. The whole batch is validated first. Control characters are refused
  because `.env` is line oriented: a newline would let an edit to an allowed
  variable write a second, arbitrary one. Generated values (32 random bytes,
  hex) are returned once.
- The change is written to `/backend/env-apply/request.json` (`./env-apply/run`
  on the host), `0600` and chowned to that directory's owner so the
  unprivileged host user can read it. `env-apply/apply.py` claims it, rewrites
  `.env`, recreates the affected services and writes `status.json`.
- `GET /admin/env/status` — `idle`, `pending`, `applying`, `applied` or
  `failed`. `watcher: false` means a request has sat unclaimed for 15 seconds,
  i.e. nothing on the host is watching. `DELETE /admin/env/pending` withdraws
  an unclaimed request.

The backend is always among the recreated services, because the values the
panel shows are the ones this process started with.

### Job capture

`handlers/handle_job_capture.go` is the server half of the browser userscript
(`vue/public/job-capture.user.js`, installed from the tracker page) that adds
the posting open in the browser to the job tracker.

- **Auth** is `Authorization: Bearer <token>`, checked by
  `CaptureTokenMiddleware`. The token opens these three routes and nothing
  else — it cannot read the tracker or reach GraphQL. **With no token
  generated the routes are off (503)**; that never means "no auth".
- **The token is generated in the admin panel**, not set in `.env`
  (`services/capture_token.go`; `GET`/`POST`/`DELETE /admin/capture-token`,
  cookie-authenticated like the rest of the admin group). It is 32 random
  bytes, shown once; only its SHA-256 is kept, in
  `/backend/token/job_capture_token.json` beside the OAuth tokens, so it
  survives restarts and rebuilds. There is one token at a time: generating
  again replaces it, which is how a leaked one is rotated.
- `GET /jobs/capture/status?url=` — is this posting already tracked, and at
  what status. Drives the script's badge.
- `POST /jobs/capture/extract` — page text in, `{jobTitle, company, location,
  notes}` out, read by Claude through the Python service
  (`/internal/jobs/extract`, see [Calling the Python service](#calling-the-python-service)).
  Stores nothing.
- `POST /jobs/capture` — an **upsert**. A posting is matched by its canonical
  URL (`canonicalJobURL` strips tracking parameters and folds LinkedIn's
  several URL shapes into one), then by `(company, job title)` the way the
  email pipeline matches. A repeat capture only ever adds: status moves
  forwards, empty fields are filled, nothing is overwritten.

The upsert is what keeps this and the email pipeline from making two rows for
one application. It is not watertight: the pipeline still matches on the exact
title, so a posting captured as "Senior Engineer" whose confirmation email
says "Senior Engineer (Platform)" becomes a second row.

## Calling the Python service

Every Claude call the site makes lives in the Python service; this backend has
no Anthropic client and does not read `CLAUDE_API_KEY`. It reaches three
internal routes directly over the Docker network
(`http://$PYTHON_HOST:$PYTHON_PORT`, default `python:8000`), never through
nginx, which answers 404 for `/py/internal/`:

| route | caller | sends | gets back |
| --- | --- | --- | --- |
| `/internal/rowing/read` | `POST /rowing` (`handlers/handle_rowing.go`) | `{media_type, data}` (image, base64) | `{timeMinutes, timeSeconds, distance}` |
| `/internal/jobs/extract` | `POST /jobs/capture/extract` (`handlers/handle_job_capture.go`) | `{url, title, text}` | `{jobTitle, company, location, notes}` |
| `/internal/email/classify` | the email pipeline (`services/email_sync.go`) | `{message}` | `EmailAnalysis` (below) |

All three go through one function, `services.CallPython[T]` in
`services/python_call.go`: marshal the body, POST it, decode a 200 into `T`.
It has the one HTTP client (two-minute timeout, above the Python side's own
worst case of a retried Claude call) and the one rule for failures:

- a **502 whose `detail` is a string** is a message the route wrote for a
  person ("failed to parse image data"). It comes back as a
  `*services.PythonError`, and the two HTTP handlers relay it to the browser;
- **anything else** — connection refused, a 422 (its `detail` is a list), a
  503 for an unset `CLAUDE_API_KEY`, a reply that will not decode — is an
  ordinary error. Callers log it and show something generic.

It lives in `services` because both `handlers` and `services` need it and
`handlers` already imports `services`. A new internal route needs only a
request struct, a reply struct and a one-line call.

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
  it, have each survivor classified, create or advance a `JobApplication`.
  Runs on a timer and can be triggered by an admin.

  The classification is the Python service's `/internal/email/classify`,
  which owns the prompt, the Claude call and the parsing of the model's
  reply. Go sends one string — `Subject:`, `From:` and `Date:` lines, a blank
  line, the cleaned body (capped at 4,000 bytes; the whole message at the
  200,000 characters Python accepts) — and gets back `EmailAnalysis`:
  `isJobEmail`, `action`, `company`, `jobTitle`, `status`, and the nullable
  `location`, `url`, `appliedAt`, `notes`. Everything either side of that
  call is unchanged and still here: the pre-filter, dedup, retry accounting,
  matching on (company, job title), and the forward-only status rule. Any
  failure of the call, whatever its status code, is a failed attempt at that
  email.
- **python_call.go** — `CallPython`, the one way to reach the Python service's
  internal routes. See [Calling the Python service](#calling-the-python-service).
- **job_status.go** — how the `JobStatus` GraphQL enum is read and written.
  See [Job status](#job-status).
- **notes.go** — path-traversal-safe file lookup for the notes directory.
- **seed.go** — dev-only test data. Creates an admin with the password
  `password`, so it must never run in production (`SEED_DB=true` is set only
  in `docker-compose.dev.yml`).

## Job status

A job application's status is one of eight lowercase strings: `saved`,
`applied`, `screening`, `assessment`, `interviewing`, `offer`, `rejected`,
`withdrawn`. The list is `statusOrder` in `services/email_sync.go`, which also
ranks them — the email pipeline and the capture endpoint only ever move a row
forwards through it (`services.StatusAdvances`).

GraphQL knows the same list as `enum JobStatus`
(`graph/schema/job_application.graphql`), used for `JobApplication.status` and
both inputs. `TestJobStatusEnumMatchesStatusOrder` in
`services/job_status_test.go` parses the schema file and fails unless the enum
and the map hold the same statuses in the same order, so adding a status means
editing both. (The classifier prompt in the Python service, the tracker page
and the userscript list them too, and nothing checks those.)

The enum is bound in `gqlgen.yml` to `services.MarshalJobStatus` /
`UnmarshalJobStatus` instead of a generated Go type, for two reasons:

- **The column stays a `string`.** `models.JobApplication.Status` binds as it
  is; there is no migration and no conversion in the pipeline.
- **Input is strict, output is lenient.** A mutation with a status outside the
  list is refused (`"hired" is not a valid JobStatus`) — nothing is coerced to
  `applied` any more. But a row already holding something else, such as the
  Title Case values the tracker once wrote, must not break reads: `status` is
  non-null, so a serialisation error on one row would null the entire
  `jobApplications` list. On the way out a value that matches once trimmed and
  lowercased is sent in its canonical form, and anything else is sent exactly
  as stored for the tracker to display and correct.

On the wire nothing changed from when the field was a `String`: the client
sends `"applied"` as a JSON string variable and reads back `"applied"`. (An
enum does differ for a status written *inline* in a query document, which must
now be `status: applied`, unquoted. The front end only uses variables.)

`services.NormalizeStatus`, which trims, lowercases and falls back to
`applied`, is still what the email pipeline and `POST /jobs/capture` use on
values that do not come through GraphQL.

## Regenerating gqlgen

The `.graphql` files are the source of truth.

```
cd backend && go run github.com/99designs/gqlgen generate
```

If that stops with `missing go.sum entry ... urfave/cli`, the generator's own
dependencies are not in `go.sum`. `go get github.com/99designs/gqlgen@<the
version in go.mod>` adds them.

This rewrites `graph/generated.go` and `graph/model/models_gen.go` and adds
stubs for new fields to the matching `*.resolvers.go`. Existing resolver bodies
are preserved, but hand-written code in a regenerated file can be relocated to
the end of it — which is why shared helpers live in `*_helpers.go` instead.

`gqlgen.yml` maps most GraphQL object types directly onto the GORM structs in
`models/`, so `graph/model/` holds only input and payload types. Adding a type
to the schema without a corresponding entry in `gqlgen.yml` gets you a fresh
generated struct rather than the model you expected.

Two scalar-level bindings in `gqlgen.yml` are worth knowing about: `ID` lists
both `graphql.IntID` and `graphql.UintID` (see IDs under Gotchas), and
`JobStatus` points at the hand-written marshaller described above.

If you delete a field or stop needing a resolver, delete the resolver function
by hand before regenerating; otherwise gqlgen keeps it, commented out, in a
warning block at the end of the file.

## Gotchas

- **Soft delete everywhere.** Every model except `ProcessedEmail` has a
  `gorm.DeletedAt`, so `Delete` sets a timestamp and later queries filter the
  row out. Rows are never actually removed, and a `uniqueIndex` still collides
  with deleted rows — a soft-deleted username cannot be reused.
- **`Preload` is not lazy.** `Post.Author` is nil unless the query preloads it;
  there is no error to tell you why the field came back null.
- **IDs.** Models use `uint`; the schema uses `ID`. `gqlgen.yml` binds `ID` to
  `graphql.IntID` *and* `graphql.UintID`, so a model's `ID uint` field is
  served directly — there are no `ID` resolvers, and a new type does not need
  one. `IntID` is listed first, which is why resolver *arguments* are still
  `id int`. Both write an id as a quoted decimal string (`"12"`) and accept a
  string or a number, so the wire format is the same either way. Other
  unsigned fields are not IDs and still have resolvers: `Message.authorId`,
  `Rowing.time`, `Rowing.distance`.
- **Deletes error on a missing id.** Every delete mutation loads the row
  first. `deleteJobApplication` and `deleteJobAppReference` return `Boolean!`
  rather than the row, so the only value they ever return is `true`; a wrong
  id is a `record not found` error, not `false`.
- **`Message.AuthorID` is not a user id.** It is a per-process counter handed
  out by the chat hub on connect, meaningless across restarts.
- **The chat table is capped.** Every insert soft-deletes all but the newest
  50 messages.
- **Store's caches are unsynchronised** and shared by all requests. The races
  are benign (at worst a duplicate upstream fetch), but do not add a field
  there that would be unsafe to tear.
- **Claude's output is never parsed here.** Prompts, the markdown-fence
  stripping and the JSON parsing are all in the Python service; this side
  receives typed JSON or an error (see
  [Calling the Python service](#calling-the-python-service)).
- **The email pipeline retries a failure up to `maxEmailAttempts` (3) times.**
  A failed email is recorded with action `"error"` and an attempt counter;
  while it has attempts left the next sync reprocesses it and updates the same
  row. Because the fetch window is anchored to *processing* time, which has
  already moved past the failure, the sync also rewinds the window to the
  oldest retryable row's `ReceivedAt` — otherwise the email could never be
  re-fetched. Once the cap is reached the row is terminal, so an email Claude
  can never parse stops costing API calls. A failure is any error from the
  classify call — Python down, a 502, a 503 because `CLAUDE_API_KEY` is unset
  — so an outage of the Python service uses up attempts just as an Anthropic
  outage does. The one exception is a cancelled sync, which records nothing.
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
| `PYTHON_HOST` / `PYTHON_PORT` | Python service, for the three Claude routes (default `python:8000`) |
| `GITEA_HOST` / `GITEA_PORT` | Gitea activity feed |
| `STEAM_API_KEY` / `STEAM_ID` | Steam presence |
| `EMAIL_SYNC_ENABLED` / `EMAIL_BACKEND` / `EMAIL_SYNC_INTERVAL` | email pipeline (`graph` or `imap`) |
| `MSGRAPH_CLIENT_ID` / `_SECRET` / `_TENANT_ID` / `_REDIRECT_URI` | Microsoft Graph backend |
| `IMAP_HOST` / `_PORT` / `_EMAIL` / `_PASSWORD` | IMAP backend |

Paths baked into the code, all inside mounted volumes so they survive a
container restart: `/backend/logs`, `/backend/notes`, `/backend/uploads`,
`/backend/fallback_music`, `/backend/token/` (Spotify and Microsoft Graph
OAuth tokens).

`CLAUDE_API_KEY` is not in the table: the backend does not read it. It is in
the registry with `services: ["python"]`, because the Python service is the
only container that uses it.
