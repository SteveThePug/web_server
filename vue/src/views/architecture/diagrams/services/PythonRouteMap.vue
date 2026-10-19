<!--
  The Python API's route map: which /py/ paths nginx proxies (and how), the
  public routes FastAPI serves after the prefix strip, the one internal route
  and its two locks, and which module each route hands work to.
  Reflects: nginx/nginx_routes.conf.template (the /py locations),
  python/app/main.py (routes, _internal_only, _internal_claude_route),
  python/app/{travelodge,rowing,claude}.py, docker-compose.yml (python service),
  backend/services/python_call.go.
-->
<template>
  <DiagramFigure id="services-python-route-map" title="The Python API: routes, prefix strip and the internal door" :min-width="790">
    <svg class="dg" viewBox="0 0 1120 560" role="img" aria-labelledby="services-python-route-map-t">
      <title id="services-python-route-map-t">nginx proxies /py/ to FastAPI with the prefix stripped; /py/internal/ is a 404 at nginx, and the internal rowing route is only reachable by the Go backend directly</title>

      <!-- Callers -->
      <rect class="dg-box dg-box--pink" x="20" y="90" width="150" height="60" rx="4" />
      <text class="dg-title dg-title--pink" x="95" y="116" text-anchor="middle">BROWSER</text>
      <text class="dg-small" x="95" y="134" text-anchor="middle">/hotels page, /py/docs</text>

      <rect class="dg-box dg-box--green" x="20" y="440" width="170" height="70" rx="4" />
      <text class="dg-title" x="105" y="466" text-anchor="middle">GO BACKEND</text>
      <text class="dg-code" x="105" y="484" text-anchor="middle">CreateRowing</text>
      <text class="dg-small" x="105" y="500" text-anchor="middle">admin-only POST /api/rowing</text>

      <!-- nginx -->
      <rect class="dg-box" x="220" y="50" width="300" height="350" rx="4" />
      <text class="dg-title" x="370" y="76" text-anchor="middle">NGINX</text>
      <text class="dg-small" x="370" y="94" text-anchor="middle">nginx_routes.conf.template</text>

      <rect class="dg-box dg-box--tint" x="234" y="110" width="272" height="62" rx="3" />
      <text class="dg-code" x="246" y="130">~ ^/py/hotels/(search|scan)</text>
      <text class="dg-small" x="246" y="148">zone hotel_search 3r/m, burst 2</text>
      <text class="dg-small" x="246" y="164">proxy_buffering off · read 300s</text>

      <rect class="dg-box dg-box--pink dg-box--tint" x="234" y="188" width="272" height="54" rx="3" />
      <text class="dg-code" x="246" y="208">^~ /py/internal/</text>
      <text class="dg-text dg-text--pink" x="246" y="228">return 404 (never proxied)</text>

      <rect class="dg-box dg-box--tint" x="234" y="258" width="272" height="48" rx="3" />
      <text class="dg-code" x="246" y="278">/py/</text>
      <text class="dg-small" x="246" y="296">zone api 30r/s, burst 20 · read 300s</text>

      <line class="dg-rule" x1="234" y1="324" x2="506" y2="324" />
      <text class="dg-small" x="246" y="344">both proxied blocks:</text>
      <text class="dg-code" x="246" y="362">rewrite ^/py/(.*)$ /$1 break</text>
      <text class="dg-small" x="246" y="382">sets X-Real-IP, X-Forwarded-For</text>

      <path class="dg-edge dg-edge--pink" d="M170,120 H195 V141 H230" />
      <text class="dg-edge-label dg-edge-label--pink" x="176" y="112">/py/…</text>
      <path class="dg-edge dg-edge--pink" d="M195,141 V214 H230" />
      <path class="dg-edge dg-edge--pink" d="M195,214 V282 H230" />

      <!-- Python container -->
      <rect class="dg-group dg-group--fill" x="560" y="40" width="320" height="400" rx="6" />
      <text class="dg-group-label" x="574" y="62">CONTAINER · python</text>
      <text class="dg-code" x="574" y="82">uvicorn :8000 --root-path /py</text>

      <rect class="dg-box" x="574" y="96" width="292" height="196" rx="3" />
      <text class="dg-text dg-text--mint" x="588" y="116">PUBLIC (written relative to /)</text>
      <text class="dg-code" x="588" y="138">GET  /  · /health · /hello/{name}</text>
      <text class="dg-code" x="588" y="158">GET  /docs · /openapi.json</text>
      <text class="dg-code" x="588" y="178">GET  /hotels/origins</text>
      <line class="dg-rule" x1="584" y1="190" x2="856" y2="190" />
      <text class="dg-code" x="588" y="210">POST /hotels/search</text>
      <text class="dg-small" x="588" y="226">blocking: 429 busy, 502 upstream</text>
      <text class="dg-code" x="588" y="248">POST /hotels/search/stream</text>
      <text class="dg-code" x="588" y="268">POST /hotels/scan/stream</text>
      <text class="dg-small" x="588" y="284">NDJSON, always 200</text>

      <rect class="dg-box dg-box--pink" x="574" y="310" width="292" height="116" rx="3" />
      <text class="dg-text dg-text--pink" x="588" y="330">INTERNAL · _internal_claude_route</text>
      <text class="dg-code" x="588" y="352">POST /internal/rowing/read</text>
      <text class="dg-small" x="588" y="372">Depends(_internal_only): 404 if the request</text>
      <text class="dg-small" x="588" y="388">carries X-Real-IP or X-Forwarded-For</text>
      <text class="dg-small" x="588" y="410">ReadError → 502 · NotConfigured → 503</text>

      <path class="dg-edge" d="M506,141 H570" />
      <path class="dg-edge" d="M506,282 H540 V160 H570" />

      <!-- Backend direct call -->
      <path class="dg-edge dg-edge--green dg-edge--flow" d="M190,475 H545 V380 H570" />
      <text class="dg-edge-label" x="370" y="466" text-anchor="middle">services.CallPython → http://python:8000 (no proxy headers)</text>

      <!-- Modules -->
      <rect class="dg-box dg-box--tint" x="910" y="110" width="190" height="96" rx="4" />
      <text class="dg-title dg-title--sm" x="1005" y="132" text-anchor="middle">TRAVELODGE.PY</text>
      <text class="dg-small" x="922" y="154">run_search · run_scan</text>
      <text class="dg-small" x="922" y="172">→ Travelodge /api/v2/hotel</text>
      <text class="dg-small" x="922" y="190">→ TfL Journey Planner</text>

      <rect class="dg-box dg-box--tint" x="910" y="300" width="190" height="50" rx="4" />
      <text class="dg-title dg-title--sm" x="1005" y="322" text-anchor="middle">ROWING.PY</text>
      <text class="dg-code" x="1005" y="340" text-anchor="middle">read_display</text>

      <rect class="dg-box dg-box--tint" x="910" y="380" width="190" height="62" rx="4" />
      <text class="dg-title dg-title--sm" x="1005" y="402" text-anchor="middle">CLAUDE.PY</text>
      <text class="dg-code" x="1005" y="420" text-anchor="middle">ask_json</text>
      <text class="dg-small" x="1005" y="436" text-anchor="middle">model = CLAUDE_MODEL</text>

      <rect class="dg-box dg-box--muted" x="910" y="474" width="190" height="44" rx="4" />
      <text class="dg-text" x="1005" y="501" text-anchor="middle">Anthropic Messages API</text>

      <path class="dg-edge" d="M866,230 H888 V158 H906" />
      <path class="dg-edge dg-edge--pink" d="M866,352 H888 V325 H906" />
      <path class="dg-edge dg-edge--pink" d="M1005,350 V376" />
      <path class="dg-edge dg-edge--pink" d="M1005,442 V470" />
      <text class="dg-small" x="910" y="546">none of the three modules imports FastAPI</text>
    </svg>
    <template #caption>
      nginx strips <code>/py</code> with <code>rewrite ^/py/(.*)$ /$1 break</code>, so FastAPI's
      routes are written relative to <code>/</code> and <code>--root-path /py</code> only exists to
      keep <code>/py/docs</code> pointing at the right OpenAPI URL. The hotel streams get their own
      location because nginx would otherwise buffer the whole NDJSON response. The pink route spends
      a paid Claude call and has no auth of its own, so it has two locks: nginx answers
      <code>/py/internal/</code> with a 404, and <code>_internal_only</code> 404s anything carrying
      the <code>X-Real-IP</code>/<code>X-Forwarded-For</code> headers nginx always adds. The Go
      backend (green) calls the container directly and sends neither. The guard is a dependency, so
      a request from outside gets the 404 before its body is validated.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
