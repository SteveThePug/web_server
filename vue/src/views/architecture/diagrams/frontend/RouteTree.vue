<!--
  The route table as a tree: the router, the three top-level records (two
  layout shells and the bare /stp2), and every leaf route with its name, its
  component file and whether it is bundled in the entry or code-split.
  Reflects: vue/src/router/index.js, App.vue, layouts/DefaultLayout.vue,
  layouts/CVLayout.vue. (/architecture itself is left out on purpose.)
-->
<template>
  <DiagramFigure id="frontend-route-tree" title="The route tree and its layouts" :min-width="800">
    <svg class="dg" viewBox="0 0 1140 640" role="img" aria-labelledby="frontend-route-tree-t">
      <title id="frontend-route-tree-t">
        Vue Router table: DefaultLayout holding the landing page, admin routes, automata, hotels and
        the catch-all CanvasPage (which serves /stp); /stp2 with no layout; /cv inside CVLayout
      </title>

      <!-- Router -->
      <rect class="dg-box dg-box--tint" x="20" y="250" width="190" height="100" rx="4" />
      <text class="dg-title" x="115" y="276" text-anchor="middle">ROUTER</text>
      <text class="dg-code" x="115" y="298" text-anchor="middle">createWebHistory()</text>
      <text class="dg-small" x="115" y="318" text-anchor="middle">App.vue is only</text>
      <text class="dg-code" x="115" y="336" text-anchor="middle">&lt;RouterView /&gt;</text>

      <path class="dg-edge dg-edge--plain" d="M210,300 H230" />
      <path class="dg-edge dg-edge--plain" d="M230,211 V523" />
      <path class="dg-edge" d="M230,211 H246" />
      <path class="dg-edge dg-edge--muted" d="M230,453 H246" />
      <path class="dg-edge" d="M230,523 H246" />

      <!-- Layouts -->
      <rect class="dg-box" x="250" y="121" width="200" height="180" rx="4" />
      <text class="dg-title" x="350" y="146" text-anchor="middle">DEFAULTLAYOUT</text>
      <text class="dg-code" x="350" y="166" text-anchor="middle">path: "/"</text>
      <line class="dg-rule" x1="262" y1="178" x2="438" y2="178" />
      <text class="dg-small" x="264" y="198">Navbar · &lt;main&gt; · Footer</text>
      <text class="dg-small" x="264" y="218">dark theme, .halftone</text>
      <text class="dg-small" x="264" y="238">&lt;Transition name="slide"&gt;</text>
      <text class="dg-small" x="264" y="258">:key="$route.path" remounts</text>
      <text class="dg-small" x="264" y="278">paints background.style</text>

      <rect class="dg-box dg-box--muted" x="250" y="430" width="200" height="46" rx="4" />
      <text class="dg-text dg-text--muted" x="350" y="451" text-anchor="middle">(no layout)</text>
      <text class="dg-small" x="350" y="467" text-anchor="middle">top-level record</text>

      <rect class="dg-box" x="250" y="500" width="200" height="46" rx="4" />
      <text class="dg-title dg-title--sm" x="350" y="519" text-anchor="middle">CVLAYOUT</text>
      <text class="dg-small" x="350" y="537" text-anchor="middle">path: "/cv" · light tokens</text>

      <!-- Trunk to DefaultLayout children -->
      <path class="dg-edge dg-edge--plain" d="M450,211 H475" />
      <path class="dg-edge dg-edge--plain" d="M475,43 V323" />
      <path class="dg-edge" d="M475,43 H496" />
      <path class="dg-edge" d="M475,99 H496" />
      <path class="dg-edge dg-edge--pink" d="M475,155 H496" />
      <path class="dg-edge" d="M475,211 H496" />
      <path class="dg-edge" d="M475,267 H496" />
      <path class="dg-edge" d="M475,323 H496" />
      <path class="dg-edge dg-edge--muted" d="M450,453 H496" />
      <path class="dg-edge" d="M450,523 H496" />

      <!-- Leaves -->
      <g v-for="r in rows" :key="r.path">
        <rect
          class="dg-box"
          :class="{ 'dg-box--teal dg-box--tint': r.static, 'dg-box--pink': r.admin, 'dg-box--muted': r.bare }"
          x="500" :y="r.y" width="620" height="46" rx="4"
        />
        <text class="dg-code" x="514" :y="r.y + 19">{{ r.path }}</text>
        <text class="dg-text" x="634" :y="r.y + 19">{{ r.name }}</text>
        <text class="dg-code" x="760" :y="r.y + 19" style="fill: var(--color-secondary)">{{ r.file }}</text>
        <text
          :class="r.static ? 'dg-text dg-text--mint' : 'dg-small'"
          x="1106" :y="r.y + 19" text-anchor="end"
        >{{ r.static ? "STATIC" : "import()" }}</text>
        <text class="dg-small" x="514" :y="r.y + 37" :style="r.admin ? 'fill: var(--color-tertiary)' : ''">{{ r.note }}</text>
      </g>

      <!-- Column heads -->
      <text class="dg-small" x="514" y="12">path</text>
      <text class="dg-small" x="634" y="12">name (a contract)</text>
      <text class="dg-small" x="760" y="12">component</text>

      <!-- Legend -->
      <rect class="dg-box dg-box--teal dg-box--tint" x="20" y="586" width="30" height="18" rx="3" />
      <text class="dg-small" x="58" y="599">in the entry bundle (first paint)</text>
      <rect class="dg-box" x="270" y="586" width="30" height="18" rx="3" />
      <text class="dg-small" x="308" y="599">dynamic import(): its own Vite chunk</text>
      <rect class="dg-box dg-box--pink" x="540" y="586" width="30" height="18" rx="3" />
      <text class="dg-small" x="578" y="599">guarded by meta.requiresAdmin</text>
      <rect class="dg-box dg-box--muted" x="790" y="586" width="30" height="18" rx="3" />
      <text class="dg-small" x="828" y="599">outside both layouts</text>
      <text class="dg-small" x="20" y="628">Vue Router ranks the catch-all below every static path, so a static route always wins over a canvas page at the same address.</text>
    </svg>
    <template #caption>
      <code>App.vue</code> is an empty <code>&lt;RouterView /&gt;</code>; all chrome lives in the
      layouts the router mounts. Only <code>Landing.vue</code> is imported statically, because it is
      the first paint, and even it pulls <code>BoxCanvas</code> in as a separate chunk. Every other view
      is a dynamic <code>import()</code>. Pages made in the browser have no entry here: the catch-all
      looks the path up as a <code>Page</code>, which is why the backend's <code>validateSlug</code>
      refuses slugs whose first segment a static route or nginx already answers. Route
      <code>name</code>s are used by <code>RouterLink</code> and <code>router.push</code>, and
      <code>Login.vue</code> checks for the name <code>"404"</code>, so renaming one breaks links.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";

const rows = [
  { y: 20, path: "/", name: "landing", file: "views/landing/Landing.vue", static: true,
    note: "BoxCanvas page=\"index\"; the canvas itself is a defineAsyncComponent chunk" },
  { y: 76, path: "/admin/login", name: "admin-login", file: "views/admin/Login.vue",
    note: "not guarded: calls auth.ensureLoaded() itself for its \"Already signed in\" card" },
  { y: 132, path: "/admin", name: "admin", file: "views/admin/Admin.vue", admin: true,
    note: "meta: { requiresAdmin: true }, the only flag the beforeEach guard acts on" },
  { y: 188, path: "/automata", name: "mobile-automata", file: "views/automata/MobileAutomata.vue",
    note: "Turing machine / mobile automaton drawn to a <canvas>" },
  { y: 244, path: "/hotels", name: "cheap-hotels", file: "views/hotels/CheapHotels.vue",
    note: "Travelodge price scanner over /py/hotels (NDJSON streams via fetch)" },
  { y: 300, path: "/:pathMatch(.*)*", name: "404", file: "views/canvas/CanvasPage.vue",
    note: "any other path, /stp included: query page(slug) → BoxCanvas for that slug, else the 404 view" },
  { y: 430, path: "/stp2", name: "home2", file: "views/home2/Home2.vue", bare: true,
    note: "static full-bleed CSS mock: no navbar, no footer, unlinked" },
  { y: 500, path: "/cv", name: "cv", file: "views/CV/CV.vue",
    note: "template picker + Print; Jobs button leaves the SPA for /jobs/" },
];
</script>
