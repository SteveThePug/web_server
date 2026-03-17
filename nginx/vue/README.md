# My Web - Frontend

Vue 3 SPA for [adam-french.co.uk](https://adam-french.co.uk). Built with Vite, Tailwind CSS v4, Pinia, and Vue Router.

## Setup

```sh
npm install
```

## Development

```sh
npm run dev
```

The Vite dev server proxies API requests:
- `/api` -> `http://localhost:8080` (Go backend)
- `/gitea` -> `http://localhost:3000` (Gitea)
- `/radio` -> `http://localhost:8000` (Icecast2)

## Production Build

```sh
npm run build
```

In production, the built `dist/` is served by Nginx inside a Docker container (see `../Dockerfile`).

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd)
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)
