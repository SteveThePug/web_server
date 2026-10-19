<!--
  Request lifecycle through the Go backend: nginx strips /api, Gin matches the
  route, every chain starts with the traffic logger and Recovery, then the
  route group's middleware (open / protected / admin) or, for /graphql,
  AuthContextMiddleware -> gqlgen -> directive -> resolver. Everything ends at
  handlers.Store.
  Reflects: backend/main.go (r.Use, route groups, gqlSrv setup),
  handlers/handle_auth.go (AuthMiddlewear, AdminMiddleware),
  graph/middleware.go, graph/directives.go, nginx/nginx_routes.conf.template.
-->
<template>
  <DiagramFigure id="backend-request-lifecycle" title="Request lifecycle: from nginx to a handler" :min-width="770">
    <svg class="dg" viewBox="0 0 1100 660" role="img" aria-labelledby="backend-request-lifecycle-t">
      <title id="backend-request-lifecycle-t">
        A request passes nginx, the Gin router, the logger and Recovery, then one of four
        middleware chains: open routes, the protected group, the admin group, or GraphQL.
      </title>

      <!-- Top row: into the engine -->
      <rect class="dg-box dg-box--pink" x="20" y="50" width="100" height="75" rx="4" />
      <text class="dg-title dg-title--pink" x="70" y="82" text-anchor="middle">BROWSER</text>
      <text class="dg-small" x="70" y="102" text-anchor="middle">+ cookies</text>

      <rect class="dg-box" x="150" y="50" width="200" height="75" rx="4" />
      <text class="dg-title" x="250" y="74" text-anchor="middle">NGINX</text>
      <text class="dg-code" x="250" y="94" text-anchor="middle">/api/graphql → /graphql</text>
      <text class="dg-small" x="250" y="112" text-anchor="middle">TLS · limit_req · strips /api</text>

      <rect class="dg-group dg-group--fill" x="370" y="20" width="720" height="120" rx="6" />
      <text class="dg-group-label" x="384" y="38">GIN ENGINE · EVERY CHAIN STARTS LOGGER → RECOVERY</text>

      <rect class="dg-box dg-box--tint" x="385" y="50" width="270" height="75" rx="4" />
      <text class="dg-title" x="520" y="74" text-anchor="middle">ROUTER</text>
      <text class="dg-code" x="520" y="94" text-anchor="middle">SetTrustedProxies(172.28.0.0/16)</text>
      <text class="dg-small" x="520" y="112" text-anchor="middle">match route → pick its handler chain</text>

      <rect class="dg-box" x="685" y="50" width="190" height="75" rx="4" />
      <text class="dg-title" x="780" y="74" text-anchor="middle">LOGGER</text>
      <text class="dg-code" x="780" y="94" text-anchor="middle">TrafficLogFormatter</text>
      <text class="dg-small" x="780" y="112" text-anchor="middle">stdout + /backend/logs/go.log</text>

      <rect class="dg-box" x="905" y="50" width="170" height="75" rx="4" />
      <text class="dg-title" x="990" y="74" text-anchor="middle">RECOVERY</text>
      <text class="dg-code" x="990" y="94" text-anchor="middle">gin.Recovery()</text>
      <text class="dg-small" x="990" y="112" text-anchor="middle">panic → 500</text>

      <path class="dg-edge dg-edge--pink" d="M120,88 H146" />
      <path class="dg-edge" d="M350,88 H381" />
      <path class="dg-edge" d="M655,88 H681" />
      <path class="dg-edge" d="M875,88 H901" />

      <!-- Trunk down to the four chains -->
      <path class="dg-edge dg-edge--plain" d="M990,125 V158 H60 V600" />
      <path class="dg-edge" d="M60,240 H96" />
      <path class="dg-edge" d="M60,360 H96" />
      <path class="dg-edge" d="M60,480 H96" />
      <path class="dg-edge dg-edge--flow" d="M60,600 H96" />

      <!-- Lane 1: open -->
      <text class="dg-group-label" x="100" y="186">OPEN ROUTES · r.GET / r.POST</text>
      <rect class="dg-box dg-box--muted" x="100" y="210" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm" x="195" y="236" text-anchor="middle">NO MIDDLEWARE</text>
      <text class="dg-small" x="195" y="254" text-anchor="middle">handler checks cookies itself</text>
      <path class="dg-edge" d="M290,240 H316" />
      <rect class="dg-box" x="320" y="195" width="610" height="90" rx="4" />
      <text class="dg-title dg-title--sm" x="340" y="216">HANDLER</text>
      <text class="dg-code" x="340" y="237">GET /rowing · POST /auth/login · POST /auth/refresh · POST /auth/logout</text>
      <text class="dg-code" x="340" y="256">GET /auth/check · GET /auth/validate-admin · GET /spotify/callback</text>
      <text class="dg-code" x="340" y="275">GET /spotify/listening · GET /spotify/recent · GET /ws · GET /</text>

      <!-- Lane 2: protected -->
      <text class="dg-group-label" x="100" y="316">PROTECTED GROUP · r.Group("/", AuthMiddlewear)</text>
      <rect class="dg-box dg-box--pink" x="100" y="330" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm dg-title--pink" x="195" y="354" text-anchor="middle">AUTHMIDDLEWEAR</text>
      <text class="dg-text dg-text--pink" x="195" y="374" text-anchor="middle">bad token → 401</text>
      <path class="dg-edge" d="M290,360 H536" />
      <text class="dg-edge-label" x="415" y="352" text-anchor="middle">no admin check</text>
      <rect class="dg-box" x="540" y="325" width="390" height="70" rx="4" />
      <text class="dg-title dg-title--sm" x="560" y="346">HANDLER</text>
      <text class="dg-code" x="560" y="366">POST /uploads</text>
      <text class="dg-small" x="560" y="384">private=true checks the admin claim itself (requestIsAdmin)</text>

      <!-- Lane 3: admin -->
      <text class="dg-group-label" x="100" y="426">ADMIN GROUP · AuthMiddlewear, AdminMiddleware</text>
      <rect class="dg-box dg-box--pink" x="100" y="450" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm dg-title--pink" x="195" y="474" text-anchor="middle">AUTHMIDDLEWEAR</text>
      <text class="dg-text dg-text--pink" x="195" y="494" text-anchor="middle">bad token → 401</text>
      <path class="dg-edge" d="M290,480 H316" />
      <rect class="dg-box dg-box--pink" x="320" y="450" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm dg-title--pink" x="415" y="474" text-anchor="middle">ADMINMIDDLEWARE</text>
      <text class="dg-text dg-text--pink" x="415" y="494" text-anchor="middle">admin ≠ true → 403</text>
      <path class="dg-edge" d="M510,480 H536" />
      <rect class="dg-box" x="540" y="435" width="390" height="90" rx="4" />
      <text class="dg-title dg-title--sm" x="560" y="456">HANDLER</text>
      <text class="dg-code" x="560" y="477">POST /rowing · GET /spotify/auth · GET /notes/*path</text>
      <text class="dg-code" x="560" y="496">POST /radio/upload · GET /radio/songs</text>
      <text class="dg-code" x="560" y="515">DELETE /radio/songs/:f · PATCH …/disable|enable</text>

      <!-- Lane 4: GraphQL -->
      <text class="dg-group-label" x="100" y="556">POST /graphql · NO GUARDING MIDDLEWARE</text>
      <rect class="dg-box dg-box--tint" x="100" y="570" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm" x="195" y="594" text-anchor="middle">AUTHCONTEXTMIDDLEWARE</text>
      <text class="dg-small" x="195" y="614" text-anchor="middle">never rejects; adds claims</text>
      <path class="dg-edge dg-edge--flow" d="M290,600 H316" />
      <rect class="dg-box" x="320" y="570" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm" x="415" y="594" text-anchor="middle">GQLGEN SERVER</text>
      <text class="dg-small" x="415" y="614" text-anchor="middle">complexity ≤ 200 · LRU 1000</text>
      <path class="dg-edge dg-edge--flow" d="M510,600 H536" />
      <rect class="dg-box dg-box--pink" x="540" y="570" width="170" height="60" rx="4" />
      <text class="dg-title dg-title--sm dg-title--pink" x="625" y="594" text-anchor="middle">DIRECTIVE</text>
      <text class="dg-code" x="625" y="614" text-anchor="middle">@admin @signedIn</text>
      <path class="dg-edge dg-edge--flow" d="M710,600 H736" />
      <rect class="dg-box" x="740" y="570" width="190" height="60" rx="4" />
      <text class="dg-title dg-title--sm" x="835" y="594" text-anchor="middle">RESOLVER</text>
      <text class="dg-code" x="835" y="614" text-anchor="middle">graph/*.resolvers.go</text>

      <!-- Store -->
      <rect class="dg-box dg-box--teal dg-box--deep" x="960" y="195" width="120" height="435" rx="4" />
      <text class="dg-title" x="1020" y="225" text-anchor="middle">STORE</text>
      <text class="dg-code" x="1020" y="245" text-anchor="middle">handlers.Store</text>
      <line class="dg-rule" x1="972" y1="258" x2="1068" y2="258" />
      <text class="dg-small" x="1020" y="280" text-anchor="middle">DB *gorm.DB</text>
      <text class="dg-small" x="1020" y="298" text-anchor="middle">Auth · Notes</text>
      <text class="dg-small" x="1020" y="316" text-anchor="middle">Uploads</text>
      <text class="dg-small" x="1020" y="334" text-anchor="middle">LoginLimiter</text>
      <text class="dg-small" x="1020" y="352" text-anchor="middle">SpotifyClient</text>
      <text class="dg-small" x="1020" y="370" text-anchor="middle">PythonURL</text>
      <text class="dg-small" x="1020" y="388" text-anchor="middle">TTL caches</text>
      <text class="dg-small" x="1020" y="406" text-anchor="middle">(cacheMu)</text>
      <path class="dg-edge" d="M930,240 H956" />
      <path class="dg-edge" d="M930,360 H956" />
      <path class="dg-edge" d="M930,480 H956" />
      <path class="dg-edge dg-edge--flow" d="M930,600 H956" />

      <circle class="dg-badge" cx="100" cy="210" r="10" />
      <text class="dg-badge-text" x="100" y="210">1</text>
      <circle class="dg-badge" cx="100" cy="330" r="10" />
      <text class="dg-badge-text" x="100" y="330">2</text>
      <circle class="dg-badge" cx="100" cy="450" r="10" />
      <text class="dg-badge-text" x="100" y="450">3</text>
      <circle class="dg-badge" cx="100" cy="570" r="10" />
      <text class="dg-badge-text" x="100" y="570">4</text>
    </svg>
    <template #caption>
      nginx strips <code>/api</code>, so the backend's routes are written from <code>/</code>.
      Gin matches the route first, then runs its chain, which always begins with
      <code>TrafficLogFormatter</code> and <code>Recovery</code>. After that there are four kinds of chain:
      open routes ① that check cookies themselves if they need to, the <code>protected</code> group ②,
      the <code>admin</code> group ③ (whose <code>AdminMiddleware</code> reads the claims
      <code>AuthMiddlewear</code> stored, so the order matters), and GraphQL ④ (the animated path),
      which has no guard in front of it at all: each field's directive decides. Pink marks the auth
      checks. <code>SetTrustedProxies</code> pins the Docker network so <code>ClientIP()</code>, which
      the login rate limiter keys on, can't be spoofed with <code>X-Forwarded-For</code>.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
