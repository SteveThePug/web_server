# My Web

## Important TODO

- Get a new background

![screenshot](vue/public/img/screenshot.png)

Welcome to the source code for my website! Please contact me if you would like to collaborate and thank you for visiting.

This website is currently self hosted on my Raspberry Pi. Any interference and the killswitch will activate and stop the UK national grid power system so please don't tamper with my domain :).

## Architecture

All services run in Docker containers orchestrated by Docker Compose:

```
vue             ── Frontend build (outputs dist to shared volume)
nginx (80, 443) ── Frontend SPA + Reverse Proxy
backend (8080)  ── Go API
db (5432)       ── PostgreSQL 16
icecast2 (8000) ── Audio Streaming
gitea (3000)    ── Self-Hosted Git
gitea-runner    ── CI/CD Runner
certbot         ── SSL Certificate Management
```

## Tech Stack

**Frontend** - Vue 3, Vite, Tailwind CSS v4, Pinia, Vue Router, markdown-it (wikilinks + KaTeX), Rust/WASM

**Backend** - Go (Gin), gqlgen (GraphQL), GORM, PostgreSQL, JWT auth, WebSockets

**Integrations** - Spotify API, Steam API, Anthropic Claude API, Icecast2

**Infrastructure** - Docker Compose, Nginx, Let's Encrypt (Certbot), Gitea + Act Runner

## Features

- Spotify integration (currently playing, recently played)
- Steam integration (online status, recent games)
- Obsidian note viewer with wikilink and LaTeX support
- Live radio streaming via Icecast2
- Real-time chat over WebSockets with image/video uploads
- Blog with admin panel (CRUD)
- Activity and rowing session tracking
- Fan shrines (GTO, Evangelion, Demoman, Skip Skip Benben)
- Self-hosted Git (Gitea) with CI/CD and commit feed on homepage
- Claude AI integration
- Printable CV with role-specific sections
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
| `/notes/:path` | Obsidian note viewer                  |
| `/shrines`     | Fan shrine index + individual shrines |

## API

The primary API is **GraphQL** at `POST /api/graphql` (with a playground at `GET /api/graphql`). Queries cover posts, users, favorites, activities, rowing sessions, Spotify (currently playing, recently played), Gitea feed, and messages.

REST endpoints handle auth (`/auth/*`), Spotify OAuth (`/spotify/*`), file uploads (`/messages/upload`), note serving (`/notes/*`), and WebSocket chat (`/api/ws`). Steam data (online status, recent games) is also available via the GraphQL API.

Protected endpoints require JWT authentication via `/auth/login` (tokens set as HTTP-only cookies).

## Local Testing (Dev Mode)

Run the full stack over plain HTTP without SSL certificates:

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

This uses an HTTP-only nginx config with all routing (SPA, backend proxy, radio, gitea) and disables certbot. Visit `http://localhost` to test.

## Future Ideas

- More Rust to WASM
- ML for chatboards
- Cache requests
- Design more webpages
- Calendar to show radio times
- Nice smooth function background and transitions
- Design shrines
- Redis (not really but practical experience)

## .env

These environment variables are found in the `.env` file. The use of environment variables can be found by reading the code so the security of the variable names are not significant.

```
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
POSTGRES_PORT=
POSTGRES_HOST=

GITEA_HOST=
GITEA_PORT=
POSTGRES_GITEA_DB=

GITEA_RUNNER_HOST=
GITEA_RUNNER_NAME=
GITEA_RUNNER_REGISTRATION_TOKEN=

BACKEND_PORT=
BACKEND_HOST=
BACKEND_SECRET=
BACKEND_ENDPOINT=

CLAUDE_API_KEY=

SEED_DB=

OBSIDIAN_DIR=

SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
SPOTIFY_REDIRECT_URI=
SPOTIFY_AUTH_STATE=

STEAM_API_KEY=
STEAM_ID=

ICECAST_SOURCE_PASSWORD=
ICECAST_RELAY_PASSWORD=
ICECAST_ADMIN_USER=
ICECAST_ADMIN_PASSWORD=
ICECAST_HOST=
ICECAST_PORT=
ICECAST_MOUNT=

DOMAIN=
EMAIL=

GITEA_LFS_JWT_SECRET=
GITEA_INTERNAL_TOKEN=
GITEA_OAUTH2_JWT_SECRET=
```
