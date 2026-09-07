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

## Styling

Everything visual is defined in three places:

| File | What it holds |
| --- | --- |
| `src/assets/styles.css` | Fonts, **design tokens**, default element styles (dark theme), reusable classes (`bdr-1`, `bdr-2`, `page-a4`, `halftone`), the `slide` and `fade` transitions, print rules |
| `src/layouts/CVLayout.vue` | The **light theme** for `/cv` and `/cv/jobs`: re-points the colour tokens to ink-on-paper values and holds the shared `.cv-btn` styles |
| `src/views/CV/cv-shared.css` | Typography and page layout for the CV templates (`.cv-template`, `.cv-page`, `--cv-*` size variables) |

### Tokens

Tokens are declared once in a Tailwind `@theme` block, so each is available both as a utility class in templates and as a CSS variable in `<style>` blocks:

| Token | Utility | Variable | Role |
| --- | --- | --- | --- |
| primary | `text-primary` `border-primary` | `var(--color-primary)` | mint: headings, borders, link text |
| secondary | `text-secondary` | `var(--color-secondary)` | green: body text |
| tertiary | `text-tertiary` | `var(--color-tertiary)` | pink: accents, small print, lists |
| quaternary | `border-quaternary` | `var(--color-quaternary)` | teal: subtle borders |
| muted | `text-muted` | `var(--color-muted)` | grey: secondary labels |
| surface | `bg-surface` | `var(--color-surface)` | widget background |
| surface-deep | `bg-surface-deep` | `var(--color-surface-deep)` | page background |
| surface-tint | `bg-surface-tint` | `var(--color-surface-tint)` | halftone dots, list rows |
| link-bg | `bg-link-bg` | `var(--color-link-bg)` | background behind links and buttons |
| heading font | `font-heading` | `var(--font-heading)` | big_noodle_titling |
| body font | (default) | `var(--font-body)` | CreatoDisplay |

Note that Tailwind's spacing unit is set to **3px** (not the default 4px), so `p-1` is 3px, `gap-2` is 6px, `w-10` is 30px.

### Conventions

- Use Tailwind utilities for layout and one-off tweaks. Reach for a scoped `<style>` only for what utilities can't express: `grid-template-areas`, keyframe animations, media-query-specific layouts.
- Never hard-code a colour in a component; use a token. The one exception is `<canvas>` drawing (MobileAutomata), which can't read CSS variables.
- Home widgets: wrap the title in `<Header>`; admin-only create forms use `<CreateToggle v-model="showCreate">` in the header's `#action` slot.
- Crossfading content in place: `<Transition name="fade">` inside a `position: relative` parent. Route changes use `<Transition name="slide">`.

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
