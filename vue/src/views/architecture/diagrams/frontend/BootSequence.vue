<!--
  Sequence of a cold load of "/": index.html → main.js → router + Pinia →
  Landing → BoxCanvas, and the three GraphQL requests the canvas fires
  (me, boxes, background), with the home query only when a widget asks.
  Reflects: vue/index.html, src/main.js, router/index.js,
  views/landing/Landing.vue, views/home3/BoxCanvas.vue, stores/auth.js,
  stores/boxes.js, stores/background.js, stores/homeData.js, graphql.js.
-->
<template>
  <DiagramFigure id="frontend-boot-sequence" title="Cold load of the landing page" :min-width="770">
    <svg class="dg" viewBox="0 0 1100 800" role="img" aria-labelledby="frontend-boot-sequence-t">
      <title id="frontend-boot-sequence-t">
        Sequence diagram: index.html loads main.js, which installs the router and Pinia and mounts the
        app; the router passes the landing route, Landing loads BoxCanvas, which asks the auth, boxes and
        background stores for data over three GraphQL requests
      </title>

      <!-- Lifelines -->
      <g v-for="l in lanes" :key="l.label">
        <line class="dg-lifeline" :x1="l.x" y1="66" :x2="l.x" y2="780" />
        <rect class="dg-box" :class="l.cls" :x="l.x - l.w / 2" y="18" :width="l.w" height="48" rx="4" />
        <text class="dg-title dg-title--sm" :class="l.titleCls" :x="l.x" y="38" text-anchor="middle">{{ l.label }}</text>
        <text class="dg-code" :x="l.x" y="56" text-anchor="middle">{{ l.sub }}</text>
      </g>

      <!-- Activations -->
      <rect class="dg-activation" x="265" y="132" width="10" height="140" />
      <rect class="dg-activation" x="435" y="172" width="10" height="150" />
      <rect class="dg-activation" x="615" y="332" width="10" height="400" />
      <rect class="dg-activation" x="795" y="412" width="10" height="330" />
      <rect class="dg-activation" x="985" y="442" width="10" height="200" />

      <!-- Messages -->
      <g v-for="(m, i) in messages" :key="i">
        <template v-if="m.self">
          <path
            class="dg-edge" :class="m.cls"
            :d="`M${m.from + 5},${m.y - 8} H${m.from + 34} V${m.y + 6} H${m.from + 9}`"
          />
          <text class="dg-edge-label" :class="m.labelCls" :x="m.from + 40" :y="m.y + 2">{{ m.label }}</text>
        </template>
        <template v-else>
          <path
            class="dg-edge" :class="m.cls"
            :d="`M${m.from + (m.to > m.from ? 5 : -5)},${m.y} H${m.to + (m.to > m.from ? -9 : 9)}`"
          />
          <text class="dg-edge-label" :class="m.labelCls" :x="(m.from + m.to) / 2" :y="m.y - 6" text-anchor="middle">{{ m.label }}</text>
        </template>
        <circle class="dg-badge" cx="26" :cy="m.y" r="10" />
        <text class="dg-badge-text" x="26" :y="m.y">{{ i + 1 }}</text>
      </g>

      <!-- Notes -->
      <rect class="dg-band" x="40" y="298" width="1050" height="24" />
      <text class="dg-small" x="460" y="314">blank until here: RouterView renders nothing while the initial navigation is pending</text>
      <rect class="dg-band dg-band--pink" x="40" y="688" width="1050" height="84" />
      <text class="dg-small" x="830" y="768">only if a widget box mounts</text>
    </svg>
    <template #caption>
      <code>main.js</code> installs the router (3) <em>before</em> Pinia (4), and the router's install
      starts the first navigation at once. That works only because navigation guards run in a promise
      chain, so by the time <code>beforeEach</code> calls <code>useAuthStore()</code> (6) Pinia is active.
      On <code>/</code> the guard has nothing to wait for. The canvas then fires three independent
      GraphQL POSTs (9–14). It asks for <code>me</code> on its own rather than through the home query,
      so the landing page never fans out to Spotify, Steam and Gitea just to decide whether to show an
      Edit button. The nine-field <code>HomeData</code> query (pink band, 17–18) is sent only as a side
      effect of the first <code>useHomeDataStore()</code> call, which happens when a widget box mounts.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";

const X = { html: 110, main: 270, router: 440, view: 620, stores: 800, gql: 990 };

const lanes = [
  { x: X.html, w: 130, label: "INDEX.HTML", sub: "<body id=app>", cls: "dg-box--pink", titleCls: "dg-title--pink" },
  { x: X.main, w: 130, label: "MAIN.JS", sub: "createApp" },
  { x: X.router, w: 140, label: "ROUTER", sub: "router/index.js" },
  { x: X.view, w: 170, label: "LANDING → CANVAS", sub: "BoxCanvas.vue" },
  { x: X.stores, w: 160, label: "PINIA STORES", sub: "auth · boxes · bg" },
  { x: X.gql, w: 180, label: "GQL()", sub: "POST /api/graphql" },
];

const messages = [
  { from: X.html, to: X.main, y: 100, label: "loads /src/main.js" },
  { from: X.main, y: 132, self: true, label: "createApp(App)" },
  { from: X.main, to: X.router, y: 172, label: "app.use(router)" },
  { from: X.main, to: X.stores, y: 210, label: "app.use(createPinia())" },
  { from: X.main, to: X.html, y: 250, label: "mount('#app')", cls: "dg-edge--pink", labelCls: "dg-edge-label--pink" },
  { from: X.router, y: 282, self: true, label: "beforeEach: no requiresAdmin → pass" },
  { from: X.router, to: X.view, y: 340, label: "render Landing" },
  { from: X.view, y: 372, self: true, label: "import() BoxCanvas" },
  { from: X.view, to: X.stores, y: 420, label: "auth.ensureLoaded()" },
  { from: X.stores, to: X.gql, y: 450, label: "query Me { me }", cls: "dg-edge--flow" },
  { from: X.view, to: X.stores, y: 490, label: "boxes:index .fetch()" },
  { from: X.stores, to: X.gql, y: 520, label: "boxes(page: \"index\")", cls: "dg-edge--flow" },
  { from: X.view, to: X.stores, y: 560, label: "background.show(\"index\")" },
  { from: X.stores, to: X.gql, y: 590, label: "background(page)", cls: "dg-edge--flow" },
  { from: X.gql, to: X.stores, y: 630, label: "3 answers, any order", cls: "dg-edge--green dg-edge--dashed", labelCls: "" },
  { from: X.stores, to: X.view, y: 660, label: "BoxFrames render", cls: "dg-edge--green dg-edge--dashed" },
  { from: X.view, to: X.stores, y: 710, label: "useHomeDataStore()", cls: "dg-edge--pink dg-edge--dashed", labelCls: "dg-edge-label--pink" },
  { from: X.stores, to: X.gql, y: 742, label: "query HomeData", cls: "dg-edge--pink dg-edge--dashed", labelCls: "dg-edge-label--pink" },
];
</script>
