<!--
  The whole production route table of the www server block: every location
  nginx matches, what it does to the request, and the container:port it lands
  on. Reflects nginx/nginx_routes.conf.template (shared routes),
  nginx/nginx.conf.template (/api/graphql, /uploads/, static blocks, apex /v2/)
  and the ports in docker-compose.yml / .env.example. Redraw when a location
  is added, gated or moved.
-->
<template>
  <DiagramFigure id="edge-route-map" title="Every route nginx serves, and where it goes" :min-width="770">
    <svg class="dg" viewBox="0 0 1100 1086" role="img" aria-labelledby="edge-route-map-t">
      <title id="edge-route-map-t">Table of nginx locations, grouped by the upstream container each one proxies to, coloured by public, admin-gated, 404 and disk-served</title>
      <text class="dg-title dg-title--sm" x="36" y="46">LOCATION (MATCH TYPE)</text>
      <text class="dg-title dg-title--sm" x="390" y="46">WHAT NGINX DOES ON THE WAY</text>
      <text class="dg-title dg-title--sm" x="812" y="46">UPSTREAM CONTAINER:PORT</text>
      <line class="dg-rule" x1="16" y1="58" x2="1084" y2="58" />
      <rect class="dg-band" x="16" y="70" width="1068" height="140" opacity="0.18" />
      <rect class="dg-box" x="800" y="73" width="280" height="134" rx="3" />
      <text class="dg-code" x="812" y="144" style="fill: var(--color-primary)">backend:8080</text>
      <text class="dg-small" x="1070" y="144" text-anchor="end">Go API</text>
      <circle class="dg-dot" cx="24" cy="85" r="4" />
      <text class="dg-code" x="36" y="89" style="fill: var(--color-secondary)">/api/ws</text>
      <text class="dg-small" x="390" y="89">strip /api · Upgrade · no rate limit</text>
      <path class="dg-edge" d="M712,85 H796" />
      <circle class="dg-dot" cx="24" cy="107" r="4" />
      <text class="dg-code" x="36" y="111" style="fill: var(--color-secondary)">/api/auth/login</text>
      <text class="dg-small" x="390" y="111">strip /api · zone login</text>
      <path class="dg-edge" d="M712,107 H796" />
      <circle class="dg-dot" cx="24" cy="129" r="4" />
      <text class="dg-code" x="36" y="133" style="fill: var(--color-secondary)">/api/uploads</text>
      <text class="dg-small" x="390" y="133">strip /api · zone upload</text>
      <path class="dg-edge" d="M712,129 H796" />
      <circle class="dg-dot" cx="24" cy="151" r="4" />
      <text class="dg-code" x="36" y="155" style="fill: var(--color-secondary)">/api/graphql</text>
      <text class="dg-small" x="390" y="155">strip /api · zone graphql · production only</text>
      <path class="dg-edge" d="M712,151 H796" />
      <circle class="dg-dot" cx="24" cy="173" r="4" />
      <text class="dg-code" x="36" y="177" style="fill: var(--color-secondary)">/api/</text>
      <text class="dg-small" x="390" y="177">strip /api · zone api</text>
      <path class="dg-edge" d="M712,173 H796" />
      <circle class="dg-dot dg-dot--pink" cx="24" cy="195" r="4" />
      <text class="dg-code" x="36" y="199" style="fill: var(--color-secondary)">= /internal/auth/admin-validate</text>
      <text class="dg-small" x="390" y="199">internal; rewritten to /auth/validate-admin</text>
      <path class="dg-edge dg-edge--pink dg-edge--dashed" d="M712,195 H796" />
      <rect class="dg-box" x="800" y="223" width="280" height="46" rx="3" />
      <text class="dg-code" x="812" y="250" style="fill: var(--color-primary)">icecast2:8000</text>
      <text class="dg-small" x="1070" y="250" text-anchor="end">radio stream</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="235" r="4" />
      <text class="dg-code" x="36" y="239" style="fill: var(--color-secondary)">/radio/admin/</text>
      <text class="dg-small" x="390" y="239">strip /radio · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,235 H796" />
      <circle class="dg-dot" cx="24" cy="257" r="4" />
      <text class="dg-code" x="36" y="261" style="fill: var(--color-secondary)">/radio/</text>
      <text class="dg-small" x="390" y="261">strip /radio</text>
      <path class="dg-edge" d="M712,257 H796" />
      <rect class="dg-band" x="16" y="282" width="1068" height="74" opacity="0.18" />
      <rect class="dg-box" x="800" y="285" width="280" height="68" rx="3" />
      <text class="dg-code" x="812" y="323" style="fill: var(--color-primary)">gitea:3000</text>
      <text class="dg-small" x="1070" y="323" text-anchor="end">git server</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="297" r="4" />
      <text class="dg-code" x="36" y="301" style="fill: var(--color-secondary)">~ /gitea/&lt;o&gt;/&lt;r&gt;/(…commit/|archive/|compare/)</text>
      <text class="dg-small" x="390" y="301">strip /gitea · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,297 H796" />
      <circle class="dg-dot" cx="24" cy="319" r="4" />
      <text class="dg-code" x="36" y="323" style="fill: var(--color-secondary)">/gitea/</text>
      <text class="dg-small" x="390" y="323">strip /gitea</text>
      <path class="dg-edge" d="M712,319 H796" />
      <circle class="dg-dot" cx="24" cy="341" r="4" />
      <text class="dg-code" x="36" y="345" style="fill: var(--color-secondary)">/v2/   (apex server block only)</text>
      <text class="dg-small" x="390" y="345">registry · no body limit · 900s</text>
      <path class="dg-edge" d="M712,341 H796" />
      <rect class="dg-box" x="800" y="369" width="280" height="90" rx="3" />
      <text class="dg-code" x="812" y="418" style="fill: var(--color-primary)">silverbullet-public:3000</text>
      <text class="dg-small" x="1070" y="418" text-anchor="end">public notes</text>
      <circle class="dg-dot" cx="24" cy="381" r="4" />
      <text class="dg-code" x="36" y="385" style="fill: var(--color-secondary)">/notes/</text>
      <text class="dg-small" x="390" y="385">prefix kept · Upgrade</text>
      <path class="dg-edge" d="M712,381 H796" />
      <circle class="dg-dot dg-dot--muted" cx="24" cy="403" r="4" />
      <text class="dg-code" x="36" y="407" style="fill: var(--color-secondary)">~* /notes/…/(private|templates)/</text>
      <text class="dg-small" x="390" y="407">privacy boundary</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M712,403 H742" />
      <text class="dg-text dg-text--muted" x="750" y="407">404</text>
      <circle class="dg-dot dg-dot--muted" cx="24" cy="425" r="4" />
      <text class="dg-code" x="36" y="429" style="fill: var(--color-secondary)">~ /notes/&lt;x&gt;/.&lt;dot-segment&gt;</text>
      <text class="dg-small" x="390" y="429">hides .git, .obsidian …</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M712,425 H742" />
      <text class="dg-text dg-text--muted" x="750" y="429">404</text>
      <circle class="dg-dot dg-dot--muted" cx="24" cy="447" r="4" />
      <text class="dg-code" x="36" y="451" style="fill: var(--color-secondary)">^~ /notes/.revisions</text>
      <text class="dg-small" x="390" y="451">full history incl. private/</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M712,447 H742" />
      <text class="dg-text dg-text--muted" x="750" y="451">404</text>
      <rect class="dg-band" x="16" y="472" width="1068" height="30" opacity="0.18" />
      <rect class="dg-box" x="800" y="475" width="280" height="24" rx="3" />
      <text class="dg-code" x="812" y="491" style="fill: var(--color-primary)">silverbullet:3000</text>
      <text class="dg-small" x="1070" y="491" text-anchor="end">notes editor</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="487" r="4" />
      <text class="dg-code" x="36" y="491" style="fill: var(--color-secondary)">/sb/</text>
      <text class="dg-small" x="390" y="491">prefix kept · Upgrade · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,487 H796" />
      <rect class="dg-box" x="800" y="515" width="280" height="68" rx="3" />
      <text class="dg-code" x="812" y="553" style="fill: var(--color-primary)">python:8000</text>
      <text class="dg-small" x="1070" y="553" text-anchor="end">FastAPI</text>
      <circle class="dg-dot" cx="24" cy="527" r="4" />
      <text class="dg-code" x="36" y="531" style="fill: var(--color-secondary)">~ ^/py/hotels/(search|scan)</text>
      <text class="dg-small" x="390" y="531">strip /py · zone hotel_search · unbuffered</text>
      <path class="dg-edge" d="M712,527 H796" />
      <circle class="dg-dot" cx="24" cy="549" r="4" />
      <text class="dg-code" x="36" y="553" style="fill: var(--color-secondary)">/py/</text>
      <text class="dg-small" x="390" y="553">strip /py · zone api</text>
      <path class="dg-edge" d="M712,549 H796" />
      <circle class="dg-dot dg-dot--muted" cx="24" cy="571" r="4" />
      <text class="dg-code" x="36" y="575" style="fill: var(--color-secondary)">^~ /py/internal/</text>
      <text class="dg-small" x="390" y="575">Go calls it directly instead</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M712,571 H742" />
      <text class="dg-text dg-text--muted" x="750" y="575">404</text>
      <rect class="dg-band" x="16" y="596" width="1068" height="74" opacity="0.18" />
      <rect class="dg-box" x="800" y="599" width="280" height="68" rx="3" />
      <text class="dg-code" x="812" y="637" style="fill: var(--color-primary)">jobs:8080</text>
      <text class="dg-small" x="1070" y="637" text-anchor="end">job tracker</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="611" r="4" />
      <text class="dg-code" x="36" y="615" style="fill: var(--color-secondary)">^~ /jobs/</text>
      <text class="dg-small" x="390" y="615">prefix kept · zone api · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,611 H796" />
      <circle class="dg-dot" cx="24" cy="633" r="4" />
      <text class="dg-code" x="36" y="637" style="fill: var(--color-secondary)">^~ /jobs/api/capture</text>
      <text class="dg-small" x="390" y="637">zone api · bearer token checked by app</text>
      <path class="dg-edge" d="M712,633 H796" />
      <circle class="dg-dot" cx="24" cy="655" r="4" />
      <text class="dg-code" x="36" y="659" style="fill: var(--color-secondary)">= /jobs/job-capture.user.js</text>
      <text class="dg-small" x="390" y="659">userscript update URL</text>
      <path class="dg-edge" d="M712,655 H796" />
      <rect class="dg-box" x="800" y="683" width="280" height="68" rx="3" />
      <text class="dg-code" x="812" y="721" style="fill: var(--color-primary)">activity-tracker:8080</text>
      <text class="dg-small" x="1070" y="721" text-anchor="end">activity tracker</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="695" r="4" />
      <text class="dg-code" x="36" y="699" style="fill: var(--color-secondary)">^~ /tracker/</text>
      <text class="dg-small" x="390" y="699">strip /tracker · zone api · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,695 H796" />
      <circle class="dg-dot" cx="24" cy="717" r="4" />
      <text class="dg-code" x="36" y="721" style="fill: var(--color-secondary)">^~ /tracker/api/v1/</text>
      <text class="dg-small" x="390" y="721">strip /tracker · device keys checked by app</text>
      <path class="dg-edge" d="M712,717 H796" />
      <circle class="dg-dot" cx="24" cy="739" r="4" />
      <text class="dg-code" x="36" y="743" style="fill: var(--color-secondary)">= /tracker/tracker.user.js</text>
      <text class="dg-small" x="390" y="743">strip /tracker</text>
      <path class="dg-edge" d="M712,739 H796" />
      <rect class="dg-band" x="16" y="764" width="1068" height="30" opacity="0.18" />
      <rect class="dg-box" x="800" y="767" width="280" height="24" rx="3" />
      <text class="dg-code" x="812" y="783" style="fill: var(--color-primary)">usaco:8080</text>
      <text class="dg-small" x="1070" y="783" text-anchor="end">USACO log</text>
      <circle class="dg-dot" cx="24" cy="779" r="4" />
      <text class="dg-code" x="36" y="783" style="fill: var(--color-secondary)">^~ /usaco/</text>
      <text class="dg-small" x="390" y="783">prefix kept · zone api · 120s</text>
      <path class="dg-edge" d="M712,779 H796" />
      <rect class="dg-box" x="800" y="807" width="280" height="24" rx="3" />
      <text class="dg-code" x="812" y="823" style="fill: var(--color-primary)">dozzle:8080</text>
      <text class="dg-small" x="1070" y="823" text-anchor="end">log viewer</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="819" r="4" />
      <text class="dg-code" x="36" y="823" style="fill: var(--color-secondary)">/dozzle/</text>
      <text class="dg-small" x="390" y="823">prefix kept · SSE unbuffered · 1h</text>
      <path class="dg-edge dg-edge--pink" d="M712,819 H796" />
      <rect class="dg-band" x="16" y="844" width="1068" height="30" opacity="0.18" />
      <rect class="dg-box" x="800" y="847" width="280" height="24" rx="3" />
      <text class="dg-code" x="812" y="863" style="fill: var(--color-primary)">garage-webui:3909</text>
      <text class="dg-small" x="1070" y="863" text-anchor="end">bucket admin</text>
      <circle class="dg-dot dg-dot--pink" cx="24" cy="859" r="4" />
      <text class="dg-code" x="36" y="863" style="fill: var(--color-secondary)">/garage/</text>
      <text class="dg-small" x="390" y="863">prefix kept · body up to 2G · unbuffered</text>
      <path class="dg-edge dg-edge--pink" d="M712,859 H796" />
      <rect class="dg-box" x="800" y="887" width="280" height="24" rx="3" />
      <text class="dg-code" x="812" y="903" style="fill: var(--color-primary)">garage:3902</text>
      <text class="dg-small" x="1070" y="903" text-anchor="end">S3 website endpoint</text>
      <circle class="dg-dot" cx="24" cy="899" r="4" />
      <text class="dg-code" x="36" y="903" style="fill: var(--color-secondary)">/uploads/</text>
      <text class="dg-small" x="390" y="903">strip · Host uploads.web.garage · 2s connect</text>
      <path class="dg-edge" d="M712,899 H796" />
      <rect class="dg-band" x="16" y="924" width="1068" height="52" opacity="0.18" />
      <rect class="dg-box dg-box--green" x="800" y="927" width="280" height="46" rx="3" />
      <text class="dg-code" x="812" y="954" style="fill: var(--color-primary)">uploads volume</text>
      <text class="dg-small" x="1070" y="954" text-anchor="end">/uploads on disk</text>
      <circle class="dg-dot dg-dot--green" cx="24" cy="939" r="4" />
      <text class="dg-code" x="36" y="943" style="fill: var(--color-secondary)">@uploads_disk</text>
      <text class="dg-small" x="390" y="943">named location · Garage 404/5xx lands here</text>
      <path class="dg-edge dg-edge--green" d="M712,939 H796" />
      <circle class="dg-dot dg-dot--pink" cx="24" cy="961" r="4" />
      <text class="dg-code" x="36" y="965" style="fill: var(--color-secondary)">/uploads/private/</text>
      <text class="dg-small" x="390" y="965">alias /uploads/private/ · auth_request</text>
      <path class="dg-edge dg-edge--pink" d="M712,961 H796" />
      <rect class="dg-box dg-box--green" x="800" y="989" width="280" height="46" rx="3" />
      <text class="dg-code" x="812" y="1016" style="fill: var(--color-primary)">vue_dist volume</text>
      <text class="dg-small" x="1070" y="1016" text-anchor="end">/etc/nginx/html</text>
      <circle class="dg-dot dg-dot--green" cx="24" cy="1001" r="4" />
      <text class="dg-code" x="36" y="1005" style="fill: var(--color-secondary)">/</text>
      <text class="dg-small" x="390" y="1005">try_files $uri $uri/ /index.html</text>
      <path class="dg-edge dg-edge--green" d="M712,1001 H796" />
      <circle class="dg-dot dg-dot--green" cx="24" cy="1023" r="4" />
      <text class="dg-code" x="36" y="1027" style="fill: var(--color-secondary)">/assets/ /fonts/ /img/ /sound/</text>
      <text class="dg-small" x="390" y="1027">expires 1y / 30d / 7d / 7d</text>
      <path class="dg-edge dg-edge--green" d="M712,1023 H796" />
      <circle class="dg-dot" cx="24" cy="1066" r="5" />
      <text class="dg-small" x="36" y="1070">public, proxied</text>
      <circle class="dg-dot dg-dot--pink" cx="169.0" cy="1066" r="5" />
      <text class="dg-small" x="181.0" y="1070">admin-gated (auth_request)</text>
      <circle class="dg-dot dg-dot--muted" cx="382.20000000000005" cy="1066" r="5" />
      <text class="dg-small" x="394.20000000000005" y="1070">answered 404 by nginx</text>
      <circle class="dg-dot dg-dot--green" cx="564.4000000000001" cy="1066" r="5" />
      <text class="dg-small" x="576.4000000000001" y="1070">served by nginx from a volume</text>
    </svg>
    <template #caption>
      Read it right to left: each box is one container on <code>app-network</code>
      (or a volume nginx reads itself), and the rows beside it are the
      locations that reach it. Pink rows run the admin gate first; grey rows
      never leave nginx. Upstreams are written as <code>set $upstream_* …</code>
      variables, so nginx resolves container names per request through Docker's
      DNS (<code>resolver 127.0.0.11 valid=10s</code>) and still boots when one of
      them is down. Every proxied row also gets the bare-prefix twin
      (<code>/gitea</code> → 301 <code>/gitea/</code>) that is left out here; in
      dev, <code>/</code> goes to <code>vue:5173</code> instead of
      <code>vue_dist</code>, and <code>/api/graphql</code> falls into
      <code>/api/</code>.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
