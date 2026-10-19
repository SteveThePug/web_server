# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

### Check everything

```
make check
```

Runs `go vet` + `go test` (backend), pytest in a venv built from `python/requirements-dev.txt` (kept in `/.venv/`), a Vite build, and `docker compose config` for both compose files. Each part is also its own target (`check-backend`, `check-python`, `check-vue`, `check-compose`); `make help` lists them. The Gitea deploy workflow runs the backend and Python checks in containers on every push, and `main` only deploys if they pass — so run `make check` before pushing.

### Full stack (dev mode, HTTP only)

```
cp .env.dev .env
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

`.env.dev` is committed and every value in it is a dummy: local-only passwords, fake third-party keys, `DOMAIN=localhost`, background jobs that touch real accounts switched off. It has to be copied to `.env` because several services read `env_file: ./.env` directly. Never put a real credential in it. Dev mode seeds the database with test data (`SEED_DB=true`) and disables certbot/SSL. Visit `http://localhost`.

### Full stack (production, HTTPS)

```
docker compose up --build
```

### Frontend only (hot reload)

```
cd vue && npm run dev
```

Vite dev server proxies `/api`, `/gitea`, `/py` and `/radio`; the targets default to `localhost:8080`, `:3000`, `:8000` and `:8000` and are overridden with `VITE_PROXY_API`, `VITE_PROXY_GITEA`, `VITE_PROXY_PY` and `VITE_PROXY_RADIO` (shell or `vue/.env.local`). Compose publishes none of the backend, Python or Icecast ports to the host, and `/py` and `/radio` share a default, so set these to wherever the services really are — see the comment in `vue/vite.config.js`.

### Frontend build

```
cd vue && npm run build
```

`npm run build` is a plain `vite build`.

### Regenerate GraphQL (after editing schema files)

```
cd backend && go run github.com/99designs/gqlgen generate
```

This regenerates `graph/generated.go` and `graph/model/models_gen.go`. Resolver implementations in `*.resolvers.go` files are preserved.

## Architecture

Dockerized multi-service personal website self-hosted on `stpvm` (an x86-64 Ubuntu VM; formerly a Raspberry Pi).

**Backend** (`backend/`): Go with Gin router. GraphQL API via gqlgen at `POST /api/graphql`. REST endpoints for auth, file uploads, Spotify OAuth, and WebSockets. GORM for PostgreSQL with auto-migrations (no separate migration files). JWT auth stored in HTTP-only cookies.

**Frontend** (`vue/`): Vue 3 SPA with Vite, Tailwind CSS v4, Pinia stores, Vue Router. Built in a separate container; assets served through Nginx (production) or proxied to Vite dev server (dev mode).

**Python API** (`python/`): FastAPI app served by uvicorn on port 8000. Nginx proxies `/py/` to it and strips the prefix, so routes in `python/app/main.py` are written relative to `/`. Swagger UI at `/py/docs`. Dev mode mounts `python/app` and runs uvicorn with `--reload`. **The site's Claude call lives here**, behind an internal route: `/internal/rowing/read` (behind `POST /api/rowing`). The Go backend calls it directly over the Docker network through `services.CallPython`, and nginx 404s `/py/internal/`. The model is one setting, shared with the job tracker, `CLAUDE_MODEL` (default `claude-sonnet-5-5`), set in `.env`; `python/app/claude.py` shapes the request per model. Adding another route is a short recipe in `python/README.md` ("Internal routes").

**Job tracker** (`jobs` service, `/jobs/`): the job application tracker, leads inbox, email pipeline and capture userscript now live in their own repo (`adamf/job-tracker` on Gitea) and run here as a published image, with their own Postgres database and role. It has no login: it asks the backend's `/auth/validate-admin` with the visitor's cookies. Nothing of it is left in `backend/`, `python/` or `vue/`; nginx redirects the old `/cv/jobs`, `/cv/leads` and userscript URLs to `/jobs/`. A push to that repo's `main` publishes the image and rolls this stack's `jobs` container onto it; pin `JOBS_IMAGE_TAG` in `.env` to hold a version. Its variables are passed through the `jobs` service's explicit `environment:` list in `docker-compose.yml`.

**Uploads / Garage** (`garage/`): every stored file goes through one path. Frontend: `vue/src/uploads.js` (`uploadFile(file)` → URL, plus `ACCEPT` for pickers, `kindOf`, `isUploadUrl`, `tagFor`). Backend: `POST /api/uploads` (`handlers/handle_upload.go`), a thin wrapper over `services.Uploads.Save` (`services/uploads.go`), which owns the type table (`UploadTypes`), the 10MB limit, random naming and storage. A new feature that takes a file calls those, never its own endpoint. Public files go to a bucket in a self-hosted Garage container (the `uploads` volume when `S3_ENDPOINT` is unset, or when Garage refuses a write). nginx serves `/uploads/` from Garage's anonymous website endpoint first and the volume second, so the bucket is the source of truth and a Garage outage only affects bucket-only files. Private attachments (`/uploads/private/`) always stay on disk. `garage-webui` at `/garage/` (admin-gated) manages the bucket. First run on a fresh volume needs `garage/bootstrap.sh`; the full runbook is "Garage" in `DEPLOYMENT.md` §7.

**Nginx** (`nginx/`): Reverse proxy + SPA server. Config is templated (`nginx.conf.template`) and selected at runtime by `entrypoint.sh` based on `DEV_MODE` and certificate presence. Rate limiting on login (5/min), API (30/sec), uploads (30/min, burst 20). Proxied routes are written once, in `nginx/nginx_routes.conf.template`, which both the production and dev configs include — add a new route there. Production-only pieces (TLS, static asset caching, the GraphQL rate-limit zone, security headers) stay in `nginx.conf.template`; a new rate-limit zone must be declared in both `http` blocks. Every request is also logged as a JSON line, body included (redacted if it mentions a password, token or secret), to `logs/nginx/payloads.log` via `nginx/payload_log.conf`. The `traffic` container tails that file so it shows up in Dozzle, which is served at `/dozzle/` behind the admin login; the admin panel no longer has a log view. nginx runs one worker per CPU and has a healthcheck (`/nginx-health` on :80); the `autoheal` container restarts it when that fails, since Docker never restarts an unhealthy container itself. Gitea's commit-pinned history views and archives are admin-gated in nginx because a crawler swarm hammers them; the branch views stay public.

## Backend Structure

- `main.go` — entry point: wires up DB, services, router
- `handlers/store.go` — `Store` struct holds DB, SpotifyAuth, Auth, Notes, etc. Passed to all handlers
- `handlers/handle_*.go` — REST handlers grouped by domain
- `graph/schema/*.graphql` — GraphQL schema files (source of truth)
- `graph/*.resolvers.go` — resolver implementations (one per schema file, `follow-schema` layout)
- `graph/generated.go` — auto-generated by gqlgen, do not edit
- `graph/model/models_gen.go` — auto-generated GraphQL models, do not edit
- `models/models.go` — GORM database models (User, Post, Message, Activity, Favorite, Rowing, Bookmark, Place, Box, Page, Background)
- `graph/directives.go` — the `@admin` / `@signedIn` schema directives (see Key Patterns)
- `services/` — database init, JWT auth, WebSocket server, Spotify OAuth, Gitea feed, the Python-service call helper (`python_call.go`), uploads (`uploads.go`: validation, naming, Garage bucket with disk fallback), DB seeding. There is no Claude client in Go any more

## Frontend Structure

- `src/graphql.js` — thin axios-based GraphQL client (`POST /api/graphql`)
- `src/api.js` — thin REST client for the `/api` JSON endpoints; failures become an `ApiError` carrying the backend's `error` message and the status
- `src/uploads.js` — storing and showing uploaded files: `uploadFile(file)` → public `/uploads/...` URL, `ACCEPT` for file pickers, `kindOf`/`tagFor`/`isUploadUrl`/`mediaSize`. Every upload goes through it
- `src/stores/` — Pinia stores for auth, posts, favorites, activities, songs, messages, homeData, places, boxes, pages. `createCrudStore.js` is the factory behind places, boxes and pages; `createMirrorStore.js` is the one behind the home-page slices
- `src/views/` — page components (Home, Admin, CV, Notes, Bookmarks, shrines). `/stp` is `views/home3/Home3.vue`: a canvas of admin-arranged boxes (built-in widgets or sandboxed-iframe HTML) stored as the `Box` GraphQL type. The canvas (`views/home3/BoxCanvas.vue`) is reusable per page key: the landing page (`/`) is one too, and admin-created `Page`s (e.g. shrines) are served at `/<slug>` by the router's catch-all, `views/canvas/CanvasPage.vue`. Boxes and pages can be admin-only, filtered server-side. Each canvas page key also has an optional `Background`, set from the canvas's Background button and painted by `DefaultLayout`. See "The /stp canvas" in `vue/README.md`
- `src/components/` — reusable UI components
- `src/assets/styles.css` — design tokens (Tailwind `@theme`), base element styles, shared classes and transitions. `layouts/CVLayout.vue` re-points the colour tokens for the light CV theme. See "Styling" in `vue/README.md`.

## Key Patterns

- **GraphQL models vs GORM models**: `gqlgen.yml` maps GraphQL types directly to GORM models in `backend/models`. The `graph/model/` package has only generated input/payload types.
- **Auth flow**: Login sets `access_token` (7 days) and `refresh_token` (365 days) as HTTP-only cookies (`backend/main.go:96`). `AuthMiddleware` validates tokens and injects user into Gin context. `AuthContextMiddleware` passes Gin context into GraphQL resolver context. Revocation is by `User.TokenVersion`: every token carries the version it was minted with, and logout or a password change bumps it, invalidating all of that user's tokens. On the frontend the router guard awaits `auth.ensureLoaded()` (a standalone `me` query), not the home query; `homeData` exposes `settled` (finished either way) as well as `loaded` (succeeded).
- **GraphQL auth is declared in the schema.** The `/api/graphql` endpoint has no guarding middleware; `AuthContextMiddleware` only makes the user *available*. A field is protected by annotating it `@admin` or `@signedIn` in `graph/schema/*.graphql` (implemented in `graph/directives.go`). `graph/directives_test.go` fails if a mutation has neither directive and is not on its short public allow-list (`login`, `logout`, `refreshToken`), so a new mutation cannot ship unguarded by accident; a new *query* is public unless you annotate it. Finer rules (post ownership, hiding private messages) stay in the resolver. See `backend/graph/doc.go`.
- **Adding a GraphQL entity**: model in `models/models.go` (and the AutoMigrate list), schema file with directives, mapping in `gqlgen.yml`, `make generate`, fill in the resolvers; on the frontend a store built with `createCrudStore`. Model `ID uint` binds directly via `graphql.UintID` — there are no hand-written ID resolvers. Full recipe in `backend/README.md`.
- **Spotify tokens**: Persisted to `/backend/token/spotify_token.json` inside the container, surviving restarts.
- **Environment variables**: every variable goes in `.env.example`, and a dummy value for it in `.env.dev`. One the Python service or the job tracker reads must also be added to the `python` or `jobs` service's explicit `environment:` list in `docker-compose.yml`, or the container never receives it. `.env` is edited by hand on the host; there is no admin-panel editor.
- **Gitea feed**: Backend proxies and caches (1 min TTL) the Gitea activity feed API.
- **All GORM models use soft delete** (`gorm.DeletedAt` field). Note that a `uniqueIndex` still collides with soft-deleted rows.

## Further documentation

Each area has its own README with the detail that doesn't belong here:

- `backend/README.md` — request lifecycle, auth model, GraphQL-vs-REST split and why each REST endpoint isn't GraphQL, service rundown, gotchas.
- `vue/README.md` — app structure, routing, store/GraphQL data flow, the design-token styling system, and a live-vs-legacy map of the views.
- `python/README.md` — endpoints, the NDJSON streaming contract, the hotel pricing method, and the scraper's fragility points.
- `DEPLOYMENT.md` — service topology and ports, dev vs prod boot paths, TLS first-run and renewal, what persists, env vars, and an ops runbook.
- `.env.example` — every environment variable the stack reads, documented, with blank values.

## Conventions

- **Comments explain *why*, not *what*.** The codebase is deliberately heavily commented at the points where behaviour is non-obvious (concurrency, caching, auth, scraping, nginx templating). Preserve these when editing; if you change the behaviour they describe, update them.
- **Resolver commentary lives inside function bodies**, because gqlgen relocates or strips anything above the signature when regenerating.
- **Schema files use `#` comments**, not triple-quoted description blocks — descriptions would change GraphQL introspection output.
- **Legacy views are kept, not deleted.** `vue/src/views/unused/`, `Intro.vue`, `Gym.vue`, the unlisted CV templates and similar carry a `STATUS:` header comment saying whether they are live. They are tree-shaken out of the build. Check the header before assuming a file is dead — e.g. `views/unused/CoverLetters.vue` is actually live on `/cv`.
