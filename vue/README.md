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

The Vite dev server stands in for nginx and proxies four prefixes. Each target
is an environment variable, read in `vite.config.js` from the shell or from
`vue/.env.local`:

| Prefix | Variable | Default | Service |
| --- | --- | --- | --- |
| `/api` | `VITE_PROXY_API` | `http://localhost:8080` | Go backend |
| `/gitea` | `VITE_PROXY_GITEA` | `http://localhost:3000` | Gitea |
| `/py` | `VITE_PROXY_PY` | `http://localhost:8000` | FastAPI (the `/py` prefix is stripped, as nginx does in production) |
| `/radio` | `VITE_PROXY_RADIO` | `http://localhost:8000` | Icecast2 |

**The defaults are not "where the compose stack is".** The stack publishes only
nginx (80/443), Gitea (3000) and the liquidsoap harbor port to the host; the
backend, the Python service and Icecast are reachable only inside the Docker
network. The defaults are each service's own port, i.e. where it listens when
you run it directly on the host (`go run .`, `uvicorn app.main:app`). `/py` and
`/radio` both default to 8000 because both services do, so at most one of them
can be right at a time — normally `/py`, with the radio simply reading as
offline.

To use a running compose stack instead, point the prefix at its nginx, e.g.
`VITE_PROXY_API=http://localhost npm run dev`. That works for `/api`, `/gitea`
and `/radio`, which are forwarded with the path intact. It does not work for
`/py`, because the proxy strips the prefix nginx routes on — `VITE_PROXY_PY`
has to be a bare uvicorn.

## Structure

```
src/
  main.js          entry point — mounts <App>
  App.vue          empty shell; all chrome lives in the layouts
  router/index.js  route table + the single admin route guard
  graphql.js       the entire GraphQL client (one axios POST)
  api.js           the REST client for the backend's JSON routes under /api
  layouts/         DefaultLayout (dark, navbar+footer) and CVLayout (light, print)
  stores/          Pinia stores
  components/      reusable UI — text/, input/, util/, borders/, elle/
  views/           page components, one directory per area
  assets/styles.css design tokens + base styles (see Styling)
  js/              small shared helpers
```

## Routing

Two layout shells, defined in `src/router/index.js`:

| Path | Layout | Notes |
| --- | --- | --- |
| `/` | DefaultLayout | Landing page (statically imported — it's the first paint): a box canvas, page key `index`; the old fixed page is its starter layout, `views/landing/landingLayout.js` |
| `/stp` | DefaultLayout | The personal home page: an ordinary canvas page (the `Page` with slug `stp`) served by the catch-all, see "The /stp canvas" below |
| `/admin`, `/admin/login` | DefaultLayout | `/admin` is guarded by `meta.requiresAdmin` |
| `/automata`, `/hotels` | DefaultLayout | |
| `/shrines`, `/shrines/*` | DefaultLayout | All canvas pages via the catch-all: `/shrines` is an index whose links are Link boxes placed by hand |
| `/cv` | CVLayout | Light print theme. Its Jobs button goes to `/jobs/`, the separate job tracker app (`adamf/job-tracker`); nginx redirects the old `/cv/jobs` and `/cv/leads` there |
| `/stp2` | (none) | Bare full-bleed mock, deliberately outside both layouts |
| `:pathMatch(.*)*` | DefaultLayout | `views/canvas/CanvasPage.vue`: an admin-created canvas page if one exists at the path, else the 404 (with a "create a page here" form for admins) |

Everything except the landing page is a dynamic `import()`, so Vite gives each
view its own chunk.

Route `name`s are contracts used by `RouterLink`/`router.push` — renaming one
breaks every link to it.

**The admin guard** (`router.beforeEach`) has one subtlety: who is signed in is
learned asynchronously. Navigating straight to `/admin` on a cold load would
otherwise read an empty user and bounce a real admin to the login page, so the
guard first awaits `auth.ensureLoaded()` and then reads `auth.user.admin`.

`ensureLoaded()` (in `stores/auth.js`) resolves once the session user is known:

- If nothing has supplied it yet, it issues a standalone
  `query { me { id username admin } }` — at most one in flight, and not again
  once answered.
- On `/stp` it usually has nothing to do: the home query also returns `me`, and
  `homeData` hands it to the auth store (`auth.adoptSession()`), so that page
  does not fire a second request.
- `logIn()`, `logOut()` and `refreshToken()` set the user directly and mark it
  known, so a change of session mid-visit is what the next navigation sees. An
  answer that was requested before such a change is discarded.
- It **never rejects and never hangs**. A failed `me` query settles as "not
  signed in" (the guard redirects to the login page) and is retried on the next
  call.

The guard deliberately does not wait on the home query. It used to, via
`homeData.loaded`, which is only set on success — so when that query failed
(it fans out to Spotify, Steam and Gitea) navigation to `/admin`
never resolved. The auth store does not import `homeData` at all; the
dependency runs the other way.

`/admin/login` is not guarded, so `Login.vue` calls `ensureLoaded()` itself to
show its "Already signed in" card.

## Data flow

There is no GraphQL cache or codegen. `src/graphql.js` is a single function:

```js
const data = await gql(`query { ... }`, variables);
```

It POSTs to `/api/graphql` and throws on an error response with no data. Auth
rides along in **HTTP-only cookies** set by the Go backend, so nothing here
touches tokens — the request just has to be same-origin (nginx in production,
the Vite proxy in dev).

The backend's REST routes go through `src/api.js`, the same idea for plain JSON:

```js
const songs = await api.get("/radio/songs", { fallback: "Failed to load songs" });
await api.post("/radio/upload", formData, { fallback: "Upload failed" });
```

It prefixes `/api`, resolves to the response body, and on failure throws an
`ApiError` whose `message` is the backend's `{"error": "..."}` string when there
is one, otherwise the `fallback` the caller passed. `err.status` carries the
HTTP status for callers that need to branch on it. Any other option is
passed to axios, so multipart bodies and `onUploadProgress` work unchanged.

### Which one to use

| You are… | Use |
| --- | --- |
| reading or writing the site's data model | `gql()`, from a store |
| doing list/create/update/delete on one GraphQL type | a store built with `createCrudStore()` |
| storing a file, or showing one that was stored | `uploadFile()` and friends from `@/uploads` |
| calling a backend REST route (admin tooling, OAuth start) | `api` |
| reading the Python service's NDJSON streams | `views/hotels/hotelsStream.js` — incremental `fetch`, not `api` |
| probing `/radio/stream` | bare `axios.head` — it is Icecast, not the API |

### Uploading files

`src/uploads.js` is the one way to store a file:

```js
import { uploadFile, ACCEPT, kindOf, tagFor } from "@/uploads";

// <input type="file" :accept="ACCEPT.images" @change="onPick">
const url = await uploadFile(file); // "/uploads/3f9c….png", public and permanent
```

It posts to `/api/uploads`; the backend (`backend/services/uploads.go`) owns
the allowed types, the 10MB limit and the file name, and stores the file in
Garage. A refused file rejects with an `ApiError` whose message can be shown
as-is. `kindOf(url)` says image/video/audio/file, `tagFor(url)` gives the
markup for a box, `isUploadUrl(url)` checks a URL from someone else before
rendering it, and `mediaSize(file)` reads an image or video's dimensions
(from a `File` or a stored URL). `listUploads()` (admin only) lists every
public upload, newest first; `components/util/UploadPicker.vue` is the modal
that shows them as a grid to choose from.
`{ private: true }` (admin only) is for private chat attachments; never use
it for anything shown in a canvas box, whose sandboxed frame has no cookie.

New views should not call `gql()` or `axios` directly (a few small admin forms
still call `gql()` for a one-off mutation). A view that needs data gets a
store; a view that needs a REST route calls `api`.

### Stores

**`stores/homeData.js` is the hub.** It issues *one* query for everything the
`/stp` page needs (posts, favourites, activities, Spotify, rowing, bookmarks,
Gitea feed, Steam status, current user) instead of letting a dozen widgets each
fire their own request. That fetch is kicked off as a side effect of the first
`useHomeDataStore()` call, so no component has to start it.

`homeData` exposes two flags, because "finished" and "succeeded" are different
questions. **`settled`** turns true when a fetch completes either way — a
loading state or anything that waits must key off this one. **`loaded`** turns
true only once a fetch has succeeded, meaning the data is real. Waiting on
`loaded` alone waits forever when the query fails.

| Store | Role |
| --- | --- |
| `homeData` | the one query; owns `settled` / `loaded` / `error`, and pushes `me` to `auth` |
| `auth` | current user (`ensureLoaded()`, see Routing), plus login/logout mutations |
| `posts`, `favorites`, `activity`, `songs`, `steam` | mirror one slice of `homeData` via an `immediate` watch, and keep a placeholder value until real data arrives |
| `places` | CRUD via `createCrudStore`; sign-in only, so it is fetched lazily |
| `boxes` | CRUD via `createCrudStore`, one store per canvas page: `useBoxesStore(page)`. Public read (admin-only boxes excepted), admin write |
| `pages` | CRUD via `createCrudStore`; the admin-created canvas pages. Public read (admin-only pages excepted), admin write |
| `messages` | the chat WebSocket — not GraphQL at all |
| `models` | dormant; not imported anywhere |

The mirroring watches are `immediate` on purpose: a store created *after*
`homeData` has already loaded would never see the existing value otherwise.

Two factories keep the repeated shapes in one place:

- **`createMirrorStore`** — the body of a store that mirrors a `homeData` slice.
- **`createCrudStore(id, { type, list, fields, ... })`** — a whole store for one
  GraphQL type that has the schema's usual `create<Type>` / `update<Type>` /
  `delete<Type>` mutations and a list query. It gives `fetch`, `create`,
  `update`, `remove` and `loaded` / `loading` / `error`, and patches the local
  array after each mutation rather than refetching. `fields` is interpolated
  into every operation so the selection sets cannot drift. Options cover the
  differences between types: `itemsKey` (the name the array is
  exposed under), `insert` (`"prepend"` or `"append"`), `deleteReturns`
  (`"object"` for `deletePlace`, `"boolean"` for a delete that returns a
  scalar — a scalar must not be given a selection set), `args` (fixed
  arguments for the list query, e.g. the boxes store's `page`) and `extend`
  for extra actions.

  The error contract is deliberate: `fetch()` never throws and records the
  failure on `error`; the mutations throw, so the form that called them can
  show the message and keep what was typed.

## The /stp canvas

`/stp` is a set of **boxes** (`Box` in GraphQL, `stores/boxes.js`) that an admin
creates, drags, resizes and edits in the browser. Changes save to the database,
so every visitor sees the saved layout. The canvas is `views/home3/BoxCanvas.vue`,
and it is not only `/stp`: every box has a `page` key, and the same component
renders any page's boxes.

- **Pages.** `"index"` is the landing page
  (`Landing.vue`), which is nothing but the canvas. Any other key is the slug of a **`Page`** (GraphQL
  type, `stores/pages.js`), an admin-created page served at `/<slug>` by the
  router's catch-all (`views/canvas/CanvasPage.vue`). Pages are created from
  the admin panel's Pages tab (which also has a "New shrine" form, a blank page
  under `shrines/`), or from the 404 of the address you want; each can start
  from a layout in `views/canvas/pageStarters.js` (Blank, Notes board, Old /stp grid). Title, address and visibility are under "Page
  settings" on the page; renaming moves its boxes, deleting takes them too.
  `/stp` is one of these pages too (slug `stp`): it used to have its own route
  (`Home3.vue`, now legacy), and the backend creates its `Page` row once at
  start-up (`ensureStpPage` in `backend/services/database.go`), so it can be
  renamed, hidden or deleted like any other.
  The server refuses slugs that an SPA or nginx route would answer first
  (`validateSlug` in `backend/graph/page_helpers.go`) — add to that list when
  adding a route.
- **Backgrounds.** Each page key can have its own background (`Background` in
  GraphQL, `stores/background.js`): a colour, an image, and whether the image
  fills the window, fits inside it or tiles. Admins set it with the
  "Background" button next to "Edit layout" (`BackgroundEditor.vue`), which
  previews live and takes an uploaded, dropped or already-uploaded image
  (`UploadPicker`). `DefaultLayout` paints
  it behind the whole window, navbar included, while that canvas is mounted,
  and falls back to the halftone. Renaming or deleting a page carries its
  background along; an admin-only page's background is hidden with it.
- **Saved layouts.** A **`Layout`** (GraphQL type, `stores/layouts.js`) is a
  named copy of one page's boxes and background, made with the "Layouts"
  button next to "Background" (`LayoutPanel.vue`). "Load here" copies a layout
  onto the page on screen; the admin panel's Pages tab copies one onto any page
  without opening it. Loading onto a page with boxes asks whether to replace
  them or add on top (copies stacked above). A replace first saves the page
  as a layout called "<page> before <layout>, <time>", so it can be undone by
  loading that back. They are copies, not links: editing a page never changes
  a layout or another page. That is also how a layout moves: save it on one
  page, load it on another. The copying happens on the server
  (`backend/graph/layout_helpers.go`); the browser only sees names and box
  counts. All of it is admin-only, since a layout keeps admin-only boxes.
- **Admin only.** A box or a page can be marked admin-only. The server never
  sends those to anyone else, so they are fit for private notes; admins see
  such boxes outlined, with a 🔒 in edit mode. The "Note" template starts
  admin-only.
- **Coordinates.** `x`/`y`/`w`/`h` are px on a fixed 1400px-wide design canvas.
  `BoxCanvas.vue` draws the canvas at that width and shrinks it with
  `transform: scale()` to fit the window, so the arrangement is identical at
  every desktop width. `BoxFrame.vue` divides pointer deltas by that scale.
  `z` is the CSS z-index.
- **Two kinds of box.** `kind: "widget"` shows a built-in Vue widget by name
  from `widgetRegistry.js`. The registry keys are stored data, so don't rename
  them. `kind: "html"` shows admin-written HTML (`HtmlBox.vue`) in an
  `<iframe sandbox="allow-scripts">` **without `allow-same-origin`**. Its
  scripts run, but they have an opaque origin: they can't read cookies, call
  `/api` as the visitor, or touch the page. That sandbox is what makes storing
  script in the database acceptable, so never add `allow-same-origin`.
  `fetch()` from a box only reaches keyless APIs that answer
  `Access-Control-Allow-Origin: *` (the Weather template uses Open-Meteo).
- **Site look inside the iframe.** `frameStyles.js` copies the colour tokens in
  and inlines the two fonts as `data:` URLs. An opaque origin can't load
  `/fonts/` without CORS headers.
- **Editing.** Admins get an "Edit layout" toggle at desktop width. Geometry
  saves on pointer release, and the box is moved locally first and put back if
  the save fails. `BoxEditor.vue` is the side panel. Settings save on change;
  HTML is a draft that previews live and saves with Save or Ctrl/Cmd+S, or
  automatically on leaving the box (selecting another, clicking empty canvas,
  Done, ✕).
  CodeMirror is loaded only when that panel opens; the panel is teleported to
  `<body>`. Its width (drag the left edge, double-click to reset) and the
  code's text size (A−/A+, or Ctrl/Cmd +/−/0 in the editor) are kept per
  browser in localStorage by `editorPrefs.js`; the canvas pads itself by the
  same width so it rescales beside the panel.
  Image and video files dropped onto the canvas in edit mode are uploaded and
  each becomes an unframed box at the drop point. "Insert upload" stores a
  file through `@/uploads` (see "Uploading files") and puts an `<img>`/`<video>`
  tag for it at the cursor; "Choose upload" does the same with a file already
  stored, and "+ From uploads…" in the admin bar makes a box from one, as a
  drop would (`UploadPicker`). Starter HTML is in `templates.js`; links in a box
  need `target="_top"` or `_blank` (the Link templates set it), or they open
  inside the frame. A canvas given a `starter` has an import button in edit
  mode (it confirms first if the page isn't empty): only `/` has one, "Import
  old landing page", which recreates the old bio and links. The old `Home.vue`
  grid (`defaultLayout.js`) is the "Old /stp grid" starter for a new page.
- **Phones (≤700px)** get no canvas: boxes stack full-width in `(y, x)` order at
  their saved height, skipping any with `hideOnMobile`. There is no separate
  mobile layout. To change the order, move boxes on the desktop canvas.

## Component conventions

- Layout and one-off tweaks go in Tailwind utilities; reach for a scoped
  `<style>` only for what utilities can't express (`grid-template-areas`,
  keyframes, media-query-specific layouts).
- Never hard-code a colour — use a token. The one exception is `<canvas>`
  drawing (`views/automata/MobileAutomata.vue`), which cannot read CSS
  variables.
- Home widgets: title goes in `<Header>`; admin create forms use
  `<CreateToggle v-model="showCreate">` in the header's `#action` slot and
  render the form inside `<Modal v-model="showCreate">`.
- Admin create forms are loaded with `defineAsyncComponent`, so a visitor who
  is not signed in never downloads them.
- Create forms emit `done` / `cancel` and never close themselves — the host
  decides what to dismiss and refresh.
- Animation loops (`Headline`, `Intro2`, `Stamps`, `AutoScroll`, `Elle`) keep
  their state in plain `let`s and write to the DOM directly. Putting per-frame
  state in a `ref` would queue a Vue re-render 30–60 times a second. They also
  cache measurements in a `ResizeObserver` rather than reading `offsetWidth` in
  the rAF callback, which forces a synchronous layout every frame.

## Dev vs production

- **Dev** (`npm run dev`): Vite serves the SPA and proxies `/api`, `/gitea`,
  `/radio` and `/py` to the `VITE_PROXY_*` targets (see Development above and
  `vite.config.js`). The `/py` proxy strips the prefix, matching what nginx
  does in production.
- **Production** (`npm run build`): `dist/` is built into the nginx container
  and served as static files, with nginx doing the same reverse-proxying.

`npm run build` is a plain `vite build`. It is also the local check that the
app still compiles — there is no test runner or linter.

## Live vs legacy views

Several files are kept around but not reachable. Each carries a `STATUS:` note
in its header comment; the summary:

| File | Status |
| --- | --- |
| `views/home3/Home3.vue` | legacy: `/stp`'s old route, before it became a canvas `Page`; no route. The rest of `views/home3/` (`BoxCanvas.vue` and friends) is live |
| `views/home/shrines/{Shrines,Demoman,GTO,Evangelion,Skipskipbenben}.vue` | the old `/shrines` index and hand-written shrines (Demoman a video wall, the rest `<Wip />`), unrouted; shrines are canvas pages now |
| `views/home/Home.vue` | legacy fixed grid that `/stp` used to be; no route. Its widgets are still live through `home3/widgetRegistry.js` |
| `views/home/Intro.vue` | superseded by `Intro2.vue`; not imported |
| `views/home/Gym.vue` | superseded by `Gym2.vue`; not imported |
| `views/home2/Home2.vue` | reachable at `/stp2`, but a static mock and unlinked |
| `views/home/BadApple.vue` | unfinished ASCII video player; not imported |
| `views/CV/CVProgramming.vue` | live — the only CV template listed in `CV.vue` |
| `views/CV/CV{General,Backend,Frontend,Security,Electrical,Hospitality,Analyst,Temp}.vue` | commented out of `CV.vue`'s list; unreachable |
| `views/unused/CoverLetters.vue` | **live** despite the directory — it is the "Cover Letter" tab on `/cv` |
| `views/unused/Notes.vue` | no route points at it; not imported |
| `components/util/{MusicPlayer,ObjectTable}.vue` | not imported |
| `components/elle/Elle.vue`, `components/borders/UtenaFrame.vue` | decorative; not imported |
| `stores/models.js`, `js/mobile-automata.mjs` | not imported |

Nothing in this list is dead weight the build ships — unimported files are
tree-shaken out. They are kept deliberately; don't delete without asking.


## Styling

Everything visual is defined in a handful of places:

| File | What it holds |
| --- | --- |
| `src/assets/styles.css` | Fonts, **design tokens**, default element styles (dark theme), reusable classes (`bdr-1`, `bdr-2`, `page-a4`, `halftone`), the `slide` and `fade` transitions, print rules |
| `src/layouts/CVLayout.vue` | The **light theme** for `/cv`: re-points the colour tokens to ink-on-paper values and holds the shared `.cv-btn` styles |
| `src/views/CV/cv-shared.css` | Typography and page layout for the CV templates (`.cv-template`, `.cv-page`, `--cv-*` size variables) |

| `src/views/hotels/hotels-shared.css` | Panels, forms, tables and tab strip for `/hotels`, shared by the single-stay and cheapest-night panels; every selector is prefixed with the page's `.hotels` root class |

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
- Home widgets: wrap the title in `<Header>`; admin-only create forms use `<CreateToggle v-model="showCreate">` in the header's `#action` slot and render the form inside a `<Modal v-model="showCreate">` so it pops up over the page instead of replacing the widget.
- Crossfading content in place: `<Transition name="fade">` inside a `position: relative` parent. Route changes use `<Transition name="slide">`.
- Keyboard focus and reduced motion are handled globally in `styles.css`: every focusable element gets a pink `:focus-visible` ring (text fields a pink border instead), and `prefers-reduced-motion` collapses transitions and CSS animations. Both are low-specificity, so a component's own focus or outline style wins. A JS animation loop has to check `matchMedia("(prefers-reduced-motion: reduce)")` itself, as `Headline.vue` does.
- Forms are real `<form @submit.prevent>` elements with `type="submit"` / `type="button"` on their buttons, not `@keyup.enter` on each input: Enter then submits from any field and password managers recognise the login. A form that saves shows its error inline and disables its button while the request is out (`useCreateForm` returns `error` and `submitting` for this).

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
