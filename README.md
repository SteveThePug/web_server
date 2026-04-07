# My Web

![screenshot](vue/public/img/screenshot.png)

Welcome to the source code for my website! Please contact me if you would like to collaborate and thank you for visiting.

This website is self-hosted on my Raspberry Pi. Any interference and the killswitch will activate and stop the UK national grid power system so please don't tamper with my domain :).

## Architecture

All services run in Docker containers orchestrated by Docker Compose behind Nginx as a reverse proxy on a single bridge network.

```
vue              ── Frontend build (outputs dist to shared volume)
nginx (80, 443)  ── Frontend SPA + Reverse Proxy
backend (8080)   ── Go API (GraphQL + REST)
db (5432)        ── PostgreSQL 16
icecast2 (8000)  ── Audio Streaming (Icecast2 + Liquidsoap)
gitea (3000)     ── Self-Hosted Git
quartz (8080)    ── Obsidian Notes Publisher (Quartz v4.4.0)
searxng (8080)   ── Meta Search Engine
hasura (8080)    ── Hasura GraphQL Engine
certbot          ── SSL Certificate Management (disabled in dev)
```

## Tech Stack

**Frontend** - Vue 3, Vite, Tailwind CSS v4, Pinia, Vue Router, markdown-it (wikilinks + KaTeX), Rust/WASM

**Backend** - Go (Gin), gqlgen (GraphQL), GORM, PostgreSQL, JWT auth, WebSockets

**Integrations** - Spotify API, Steam API, Anthropic Claude API, Icecast2

**Infrastructure** - Docker Compose, Nginx, Let's Encrypt (Certbot), Gitea + Act Runner

## Features

- Spotify integration (currently playing, recently played)
- Steam integration (online status, recent games)
- Obsidian note viewer via Quartz
- Live radio streaming via Icecast2 + Liquidsoap
- Real-time chat over WebSockets with image/video uploads
- Blog with admin panel (CRUD)
- Activity and rowing session tracking
- Fan shrines (GTO, Evangelion, Demoman, Skip Skip Benben)
- Self-hosted Git (Gitea) with CI/CD and commit feed on homepage
- Claude AI integration
- Printable CV with role-specific sections
- SearXNG meta search engine
- Hasura GraphQL console
- Landing page with animated stamps section
- Route transitions (slide/fade) and performance optimizations (gzip, WOFF2 fonts, lazy loading)

## Pages

| Route          | Description                           |
| -------------- | ------------------------------------- |
| `/`            | Landing page                          |
| `/stp`         | Home dashboard with grid layout       |
| `/admin`       | Admin panel (authenticated)           |
| `/cv`          | Curriculum Vitae (printable)          |
| `/bookmarks`   | Bookmarks                             |
| `/notes/:path` | Obsidian note viewer (via Quartz)     |
| `/shrines`     | Fan shrine index + individual shrines |

## API

### GraphQL

**Endpoint:** `POST /api/graphql`
**Playground:** `GET /api/graphql` (when `GQL_PLAYGROUND=true`)

| Operation | Type | Description |
| --- | --- | --- |
| `users` | Query | Get all users |
| `user(id)` | Query | Get user by ID |
| `me` | Query | Get authenticated user |
| `posts` | Query | Get all posts |
| `post(id)` | Query | Get post by ID |
| `activities` | Query | Get all activities |
| `favorites` | Query | Get all favorites |
| `rowingSessions` | Query | Get all rowing sessions |
| `messages` | Query | Get all messages |
| `spotifyListening` | Query | Currently playing track |
| `spotifyRecent` | Query | Recently played tracks |
| `giteaFeed` | Query | Latest Gitea activity |
| `steamStatus` | Query | Steam online status |
| `login` | Mutation | Authenticate user |
| `logout` | Mutation | Logout |
| `refreshToken` | Mutation | Refresh auth token |
| `createPost` / `updatePost` / `deletePost` | Mutation | Post CRUD (admin) |
| `createUser` / `deleteUser` | Mutation | User management (admin) |
| `setUserAdmin` | Mutation | Toggle admin status |
| `createFavorite` | Mutation | Add favorite (admin) |
| `createActivity` | Mutation | Add activity (admin) |

### REST Endpoints

**Auth**
| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/auth/login` | Login (rate limited: 5/min) |
| POST | `/api/auth/refresh` | Refresh token |
| GET | `/api/auth/check` | Check token validity |
| POST | `/api/auth/logout` | Logout |

**Spotify**
| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/spotify/callback` | OAuth callback |
| GET | `/api/spotify/listening` | Currently playing |
| GET | `/api/spotify/recent` | Recently played |

**Public**
| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/favorites` | Get all favorites |
| GET | `/api/rowing` | Get all rowing sessions |
| GET | `/api/activity` | Get all activities |
| GET | `/api/posts` | Get all posts |
| GET | `/api/posts/:id` | Get post by ID |
| GET | `/api/user` | Get all users |
| GET | `/api/user/:id` | Get user by ID |

**Protected (auth required)**
| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/messages/upload` | Upload message file (rate limited: 5/min) |

**Admin (auth + admin required)**
| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/favorites` | Create favorite |
| POST | `/api/rowing` | Create rowing session |
| POST | `/api/activity` | Create activity |
| POST | `/api/posts` | Create post |
| PUT | `/api/posts/:id` | Update post |
| DELETE | `/api/posts/:id` | Delete post |
| POST | `/api/user` | Create user |
| PUT | `/api/user/:id` | Update user |
| DELETE | `/api/user/:id` | Delete user |
| PATCH | `/api/user/:id/admin` | Set/unset admin |
| POST | `/api/radio/upload` | Upload radio song |
| GET | `/api/radio/songs` | List radio songs |
| DELETE | `/api/radio/songs/:filename` | Delete radio song |
| PATCH | `/api/radio/songs/:filename/disable` | Disable radio song |
| PATCH | `/api/radio/songs/:filename/enable` | Enable radio song |

**WebSocket**
| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/ws` | WebSocket chat (10s ping keepalive) |

### Nginx Proxy Routes

| Route | Target | Notes |
| --- | --- | --- |
| `/api` | backend:8080 | API (rate limited: 30r/s) |
| `/radio` | icecast:8000 | Audio streaming |
| `/gitea` | gitea:3000 | Git service |
| `/hasura` | hasura:8080 | GraphQL console + WebSocket |
| `/notes` | quartz:8080 | Obsidian notes |
| `/searxng` | searxng:8080 | Search engine |
| `/uploads` | local alias | User-uploaded files |

### Deprecated Endpoints

**Backend note API** (`GET /api/notes/*path`) - The backend has a REST endpoint that serves note files directly from the mounted Obsidian vault. This is superseded by the Quartz service which now handles note rendering at `/notes/`. The backend endpoint still exists in code but is no longer the primary note serving path.

## Setup

### `.env`

Create a `.env` file in the project root. All services read from this file.

| Variable | Description |
| --- | --- |
| `POSTGRES_USER` | PostgreSQL username |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `POSTGRES_DB` | Main app database name |
| `POSTGRES_PORT` | PostgreSQL port (typically `5432`) |
| `POSTGRES_HOST` | PostgreSQL hostname (use `db` for Docker) |
| `GITEA_HOST` | Gitea hostname (use `gitea` for Docker) |
| `GITEA_PORT` | Gitea HTTP port (typically `3000`) |
| `GITEA_INTERNAL_TOKEN` | Gitea internal API token (generate with `gitea generate secret INTERNAL_TOKEN`) |
| `GITEA_LFS_JWT_SECRET` | Gitea LFS JWT secret (generate with `gitea generate secret LFS_JWT_SECRET`) |
| `GITEA_OAUTH2_JWT_SECRET` | Gitea OAuth2 JWT secret (generate with `gitea generate secret JWT_SECRET`) |
| `POSTGRES_GITEA_DB` | Gitea database name |
| `UPTIMEKUMA_HOST` | Uptime Kuma hostname (planned) |
| `UPTIMEKUMA_PORT` | Uptime Kuma port (planned) |
| `SEARXNG_HOST` | SearXNG hostname (use `searxng` for Docker) |
| `SEARXNG_PORT` | SearXNG port (typically `8080`) |
| `SEARXNG_SECRET_KEY` | SearXNG secret key (random hex string) |
| `WALLABAG_HOST` | Wallabag hostname (planned) |
| `WALLABAG_PORT` | Wallabag port (planned) |
| `QUARTZ_HOST` | Quartz hostname (use `quartz` for Docker) |
| `QUARTZ_PORT` | Quartz port (typically `8080`) |
| `GITEA_RUNNER_HOST` | Gitea runner hostname |
| `GITEA_RUNNER_NAME` | Gitea runner display name |
| `GITEA_RUNNER_REGISTRATION_TOKEN` | Token to register Gitea Actions runner |
| `BACKEND_PORT` | Backend port (typically `8080`) |
| `BACKEND_HOST` | Backend hostname (use `backend` for Docker) |
| `BACKEND_SECRET` | JWT signing secret |
| `BACKEND_ENDPOINT` | API path prefix (typically `/api`) |
| `OBSIDIAN_DIR` | Absolute path to Obsidian vault on host machine |
| `SPOTIFY_CLIENT_ID` | Spotify app client ID |
| `SPOTIFY_CLIENT_SECRET` | Spotify app client secret |
| `SPOTIFY_REDIRECT_URI` | Spotify OAuth redirect (e.g. `https://www.<DOMAIN>/api/spotify/callback`) |
| `SPOTIFY_AUTH_STATE` | Arbitrary state string for Spotify OAuth |
| `ICECAST_SOURCE_PASSWORD` | Icecast source connection password |
| `ICECAST_RELAY_PASSWORD` | Icecast relay password |
| `ICECAST_ADMIN_USER` | Icecast admin username |
| `ICECAST_ADMIN_PASSWORD` | Icecast admin password |
| `ICECAST_HOST` | Icecast hostname (use `icecast` for Docker) |
| `ICECAST_PORT` | Icecast port (typically `8000`) |
| `ICECAST_MOUNT` | Icecast mount point (e.g. `/stream`) |
| `LIQUIDSOAP_HARBOR_MOUNT` | Liquidsoap live input mount (e.g. `/live`) |
| `LIQUIDSOAP_HARBOR_PORT` | Liquidsoap harbor port (e.g. `8005`) |
| `DOMAIN` | Production domain name |
| `EMAIL` | Email for Let's Encrypt registration |
| `CLAUDE_API_KEY` | Anthropic Claude API key |
| `STEAM_API_KEY` | Steam Web API key |
| `STEAM_ID` | Steam user ID |
| `HASURA_GRAPHQL_ADMIN_SECRET` | Hasura admin secret |
| `HASURA_HOST` | Hasura hostname (use `hasura` for Docker) |
| `HASURA_PORT` | Hasura port (typically `8080`) |
| `SEED_DB` | Set to `true` to seed test data on startup |

### Gitea Config

Copy from the template and fill in secrets:

```sh
cp gitea/config/app.ini.template gitea/config/app.ini
```

Populate `LFS_JWT_SECRET`, `SECRET_KEY`, `INTERNAL_TOKEN`, `JWT_SECRET`, and the database `PASSWD`. Alternatively, the Gitea entrypoint generates `app.ini` from the template using environment variables.

### SearXNG Config

Copy from the template:

```sh
cp searxng/settings.yml.template searxng/settings.yml
```

The Docker entrypoint handles environment variable substitution (`${BASE_URL}`, `${SEARXNG_SECRET_KEY}`) automatically, so manual setup is only needed when running outside Docker.

### Spotify Token Setup

1. Create a Spotify app at [developer.spotify.com](https://developer.spotify.com/dashboard) and set the redirect URI to `https://www.<DOMAIN>/api/spotify/callback`
2. Set `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`, `SPOTIFY_REDIRECT_URI`, and `SPOTIFY_AUTH_STATE` in `.env`
3. Start the stack, then visit the Spotify auth URL logged by the backend on startup to authorize the app
4. After authorization, Spotify redirects to the callback endpoint which stores tokens at `/backend/token/spotify_token.json`
5. Tokens are refreshed automatically; the file persists across container restarts via volume mount

### Obsidian Notes Setup

1. Set `OBSIDIAN_DIR` in `.env` to the absolute path of your Obsidian vault on the host machine
2. The vault is mounted read-only into the Quartz container at `/quartz/content`
3. Quartz builds a static site from the vault on startup and serves it at `/notes/`
4. The backend also mounts the vault at `/backend/notes` (legacy, see deprecated endpoints above)

### SSL Certificates (Certbot)

**Initial setup (production):**

1. Set `DOMAIN` and `EMAIL` in `.env`
2. On first run, Nginx starts with `nginx_setup.conf.template` which only serves the ACME challenge route at `/.well-known/acme-challenge/`
3. Certbot requests a certificate via `certbot certonly --webroot`
4. Once the certificate is issued to `certbot/conf/live/<DOMAIN>/`, restart Nginx — it will detect the certs and switch to the full `nginx.conf.template`
5. Certbot checks for renewal every 12 hours automatically

**Dev mode:** Nginx generates a self-signed certificate for localhost automatically.

### Icecast Radio

Place at least one `.mp3` file in `icecast2/fallback_music/`. Liquidsoap plays these as fallback when no live source is connected. Connect a live source to port 8001 on the `LIQUIDSOAP_HARBOR_MOUNT`.

### Gitea Runner

1. Download the `act_runner` binary from [Gitea releases](https://gitea.com/gitea/act_runner/releases) and place in `gitea-runner/`
2. Set `GITEA_RUNNER_REGISTRATION_TOKEN` in `.env`
3. The runner registers automatically on first startup

## Dev Mode

Run the full stack over plain HTTP without SSL certificates:

```sh
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

This:
- Uses an HTTP-only Nginx config with all routing (SPA, backend proxy, radio, gitea, etc.)
- Generates a self-signed certificate for localhost
- Disables certbot
- Seeds the database with test data (`SEED_DB=true`)
- Enables GraphQL playground and introspection
- Enables Hasura console and dev mode

Visit `http://localhost` to test.

### Frontend only (hot reload)

```sh
cd vue && npm run dev
```

Vite dev server proxies `/api` to `localhost:8080`, `/gitea` to `localhost:3000`, `/radio` to `localhost:8000`.

## Untracked Files

These files are git-ignored and must be created manually:

| File | Notes |
| --- | --- |
| `.env` | See setup section above |
| `gitea/config/app.ini` | Copy from `app.ini.template` or let entrypoint generate it |
| `searxng/settings.yml` | Copy from `settings.yml.template` or let entrypoint generate it |
| `certbot/conf/`, `certbot/www/` | Created automatically by certbot; use dev mode to skip |
| `backend/token/` | Created automatically by Docker volume mount |
| `icecast2/fallback_music/*.mp3` | Place at least one MP3 file |
| `gitea-runner/act_runner` | Download from Gitea releases |
| `gitea-runner/.runner` | Generated on first runner startup |

## Future Ideas

- More Rust to WASM
- ML for chatboards
- Cache requests
- Design more webpages
- Calendar to show radio times
- Nice smooth function background and transitions
- Design shrines
- Redis (not really but practical experience)
