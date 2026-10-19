<!--
  The big picture: every container docker-compose.yml runs on stpvm, grouped
  by role, and who talks to whom. Reflects docker-compose.yml (services,
  networks, ports, network_mode, ipv4_address, aliases, environment URLs),
  nginx/nginx_routes.conf.template (the `set $upstream_*` targets and their
  locations), nginx/nginx.conf.template (/v2/ on the apex), DEPLOYMENT.md §1
  and gitea-runner/run.sh. Re-check when a service is added or removed, a
  route moves, or a port is published.
-->
<template>
  <DiagramFigure id="infra-host-overview" title="Every container on stpvm, and who talks to whom" :min-width="840">
    <svg class="dg" viewBox="0 0 1200 920" role="img" aria-labelledby="infra-host-overview-t">
      <title id="infra-host-overview-t">
        Twenty containers on one host: nginx fans requests out to twelve services; backend, gitea and the trackers
        share one Postgres; garage holds uploads; the CI runner and two socket-holding containers sit off the network.
      </title>

      <!-- ===== Outside the host ===== -->
      <text class="dg-group-label" x="16" y="56">INTERNET</text>
      <rect class="dg-box dg-box--pink" x="16" y="76" width="150" height="66" rx="4" />
      <text class="dg-title dg-title--pink" x="91" y="102" text-anchor="middle">CLOUDFLARE</text>
      <text class="dg-small" x="91" y="120" text-anchor="middle">edge · tunnel</text>
      <text class="dg-small" x="91" y="134" text-anchor="middle">DNS API</text>

      <rect class="dg-box dg-box--pink" x="16" y="220" width="150" height="56" rx="4" />
      <text class="dg-title dg-title--pink" x="91" y="245" text-anchor="middle">VISITOR</text>
      <text class="dg-small" x="91" y="263" text-anchor="middle">browser</text>

      <rect class="dg-box dg-box--pink" x="16" y="546" width="150" height="56" rx="4" />
      <text class="dg-title dg-title--pink" x="91" y="571" text-anchor="middle">LET'S ENCRYPT</text>
      <text class="dg-small" x="91" y="589" text-anchor="middle">ACME CA</text>

      <!-- ===== Host ===== -->
      <rect class="dg-group dg-group--pink" x="186" y="16" width="998" height="846" rx="8" />
      <text class="dg-group-label" x="200" y="36" style="fill: var(--color-tertiary)">HOST · STPVM — X86-64 UBUNTU VM</text>

      <rect class="dg-group dg-group--fill" x="200" y="46" width="800" height="800" rx="6" />
      <text class="dg-group-label" x="214" y="64">APP-NETWORK · BRIDGE 172.28.0.0/16</text>

      <!-- ---- Column A: the edge ---- -->
      <rect class="dg-box" x="216" y="76" width="160" height="66" rx="4" />
      <text class="dg-title" x="296" y="100" text-anchor="middle">CLOUDFLARED</text>
      <text class="dg-code" x="296" y="117" text-anchor="middle">172.28.255.250</text>
      <text class="dg-small" x="296" y="133" text-anchor="middle">dials out, no ports</text>

      <rect class="dg-box dg-box--teal dg-box--tint" x="216" y="176" width="160" height="232" rx="4" />
      <text class="dg-title dg-title--lg" x="296" y="204" text-anchor="middle">NGINX</text>
      <text class="dg-code" x="296" y="224" text-anchor="middle">:80 :443 → host</text>
      <text class="dg-text" x="296" y="245" text-anchor="middle">TLS · SPA · proxy</text>
      <line class="dg-rule" x1="226" y1="258" x2="366" y2="258" />
      <text class="dg-small" x="296" y="276" text-anchor="middle">network aliases</text>
      <text class="dg-code" x="296" y="293" text-anchor="middle">DOMAIN</text>
      <text class="dg-code" x="296" y="309" text-anchor="middle">www.DOMAIN</text>
      <line class="dg-rule" x1="226" y1="322" x2="366" y2="322" />
      <text class="dg-small" x="296" y="340" text-anchor="middle">auth_request on admin</text>
      <text class="dg-small" x="296" y="354" text-anchor="middle">routes → backend</text>
      <text class="dg-code" x="296" y="370" text-anchor="middle">/auth/validate-admin</text>
      <text class="dg-small" x="296" y="394" text-anchor="middle">label autoheal=true</text>

      <rect class="dg-box" x="216" y="444" width="160" height="56" rx="4" />
      <text class="dg-title" x="296" y="468" text-anchor="middle">VUE</text>
      <text class="dg-small" x="296" y="486" text-anchor="middle">one-shot vite build, exits 0</text>

      <rect class="dg-box" x="216" y="540" width="160" height="66" rx="4" />
      <text class="dg-title" x="296" y="564" text-anchor="middle">CERTBOT</text>
      <text class="dg-small" x="296" y="581" text-anchor="middle">certonly, then renew</text>
      <text class="dg-small" x="296" y="595" text-anchor="middle">every 12h</text>

      <rect class="dg-box dg-box--muted" x="216" y="640" width="160" height="70" rx="4" />
      <text class="dg-small" x="296" y="662" text-anchor="middle">nginx and certbot share</text>
      <text class="dg-code" x="296" y="680" text-anchor="middle">./certbot/conf</text>
      <text class="dg-code" x="296" y="697" text-anchor="middle">./certbot/www</text>

      <!-- ingress edges -->
      <path class="dg-edge dg-edge--pink dg-edge--flow" d="M91,220 V146" />
      <text class="dg-edge-label dg-edge-label--pink" x="98" y="190">HTTPS</text>
      <path class="dg-edge dg-edge--flow" d="M166,109 H212" />
      <text class="dg-edge-label dg-edge-label--mint" x="190" y="101" text-anchor="middle">tunnel</text>
      <path class="dg-edge dg-edge--flow" d="M296,142 V172" />
      <text class="dg-edge-label dg-edge-label--mint" x="302" y="162">https://www.DOMAIN</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M166,248 H212" />
      <text class="dg-edge-label" x="190" y="240" text-anchor="middle">direct</text>
      <path class="dg-edge dg-edge--green" d="M296,444 V412" />
      <text class="dg-edge-label" x="302" y="432">vue_dist</text>
      <path class="dg-edge dg-edge--muted" d="M216,573 H170" />
      <text class="dg-edge-label" x="191" y="565" text-anchor="middle">ACME</text>

      <!-- ---- The proxy trunk ---- -->
      <path class="dg-edge dg-edge--plain" d="M376,292 H400" />
      <line class="dg-edge dg-edge--plain" x1="400" y1="92" x2="400" y2="812" />

      <!-- ---- Column C: everything nginx proxies to ---- -->
      <g>
        <rect class="dg-box dg-box--tint" x="480" y="60" width="180" height="64" rx="4" />
        <text class="dg-title" x="570" y="82" text-anchor="middle">BACKEND</text>
        <text class="dg-code" x="570" y="99" text-anchor="middle">:8080 · Go / Gin</text>
        <text class="dg-small" x="570" y="115" text-anchor="middle">GraphQL + REST</text>
        <path class="dg-edge" d="M400,92 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="85" text-anchor="middle">/api/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="150" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="170" text-anchor="middle">PYTHON</text>
        <text class="dg-code" x="570" y="187" text-anchor="middle">:8000 · FastAPI</text>
        <path class="dg-edge" d="M400,173 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="166" text-anchor="middle">/py/</text>
        <path class="dg-edge" d="M510,124 V146" />
        <text class="dg-edge-label dg-edge-label--mint" x="516" y="140">CallPython</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="218" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="238" text-anchor="middle">JOBS</text>
        <text class="dg-small" x="570" y="255" text-anchor="middle">published image · :8080</text>
        <path class="dg-edge" d="M400,241 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="234" text-anchor="middle">/jobs/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="280" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="300" text-anchor="middle">ACTIVITY-TRACKER</text>
        <text class="dg-small" x="570" y="317" text-anchor="middle">published image · :8080</text>
        <path class="dg-edge" d="M400,303 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="296" text-anchor="middle">/tracker/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="342" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="362" text-anchor="middle">USACO</text>
        <text class="dg-small" x="570" y="379" text-anchor="middle">published image · :8080</text>
        <path class="dg-edge" d="M400,365 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="358" text-anchor="middle">/usaco/</text>
      </g>
      <g>
        <rect class="dg-box dg-box--tint" x="480" y="414" width="180" height="56" rx="4" />
        <text class="dg-title" x="570" y="434" text-anchor="middle">GITEA</text>
        <text class="dg-code" x="570" y="451" text-anchor="middle">:3000 http · :2222 ssh</text>
        <text class="dg-small" x="570" y="465" text-anchor="middle" style="fill: var(--color-tertiary)">both published</text>
        <path class="dg-edge" d="M400,437 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="430" text-anchor="middle">/gitea/</text>
        <path class="dg-edge dg-edge--green dg-edge--both-green" d="M520,392 V410" />
        <text class="dg-edge-label" x="528" y="405">clone · push webhook</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="490" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="510" text-anchor="middle">ICECAST2</text>
        <text class="dg-small" x="570" y="527" text-anchor="middle">+ liquidsoap · harbor 8001 published</text>
        <path class="dg-edge" d="M400,513 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="506" text-anchor="middle">/radio/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="552" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="572" text-anchor="middle">SILVERBULLET</text>
        <text class="dg-small" x="570" y="589" text-anchor="middle">editor · vault read-write</text>
        <path class="dg-edge" d="M400,575 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="568" text-anchor="middle">/sb/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="614" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="634" text-anchor="middle">SILVERBULLET-PUBLIC</text>
        <text class="dg-small" x="570" y="651" text-anchor="middle">public · vault read-only</text>
        <path class="dg-edge" d="M400,637 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="630" text-anchor="middle">/notes/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="676" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="696" text-anchor="middle">GARAGE-WEBUI</text>
        <text class="dg-code" x="570" y="713" text-anchor="middle">:3909</text>
        <path class="dg-edge" d="M400,699 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="692" text-anchor="middle">/garage/</text>
      </g>
      <g>
        <rect class="dg-box" x="480" y="738" width="180" height="46" rx="4" />
        <text class="dg-title" x="570" y="758" text-anchor="middle">DOZZLE</text>
        <text class="dg-small" x="570" y="775" text-anchor="middle">:8080 · docker.sock :ro</text>
        <path class="dg-edge" d="M400,761 H476" />
        <text class="dg-edge-label dg-edge-label--mint" x="440" y="754" text-anchor="middle">/dozzle/</text>
      </g>
      <!-- /uploads/ passes under column C straight to garage's website port -->
      <path class="dg-edge" d="M400,812 H736" />
      <text class="dg-edge-label dg-edge-label--mint" x="570" y="805" text-anchor="middle">/uploads/ → :3902 (then the uploads volume)</text>

      <!-- ---- Column D: state ---- -->
      <rect class="dg-box dg-box--deep" x="740" y="96" width="200" height="380" rx="4" style="stroke: var(--color-secondary)" />
      <text class="dg-title" x="840" y="122" text-anchor="middle" style="fill: var(--color-secondary)">DB · POSTGRES:16</text>
      <text class="dg-small" x="840" y="138" text-anchor="middle">one instance, four databases</text>
      <line class="dg-rule" x1="750" y1="146" x2="930" y2="146" />
      <rect class="dg-band" x="742" y="226" width="196" height="30" />
      <text class="dg-code" x="756" y="245">jobs</text>
      <text class="dg-small" x="924" y="245" text-anchor="end">own role</text>
      <rect class="dg-band" x="742" y="288" width="196" height="30" />
      <text class="dg-code" x="756" y="307">tracker</text>
      <text class="dg-small" x="924" y="307" text-anchor="end">own role</text>
      <rect class="dg-band" x="742" y="422" width="196" height="30" />
      <text class="dg-code" x="756" y="441">gitea</text>
      <text class="dg-small" x="924" y="441" text-anchor="end">made by hand</text>
      <rect class="dg-band" x="742" y="160" width="196" height="30" />
      <text class="dg-code" x="756" y="179">POSTGRES_DB (site)</text>
      <text class="dg-small" x="840" y="368" text-anchor="middle">healthcheck pg_isready</text>
      <text class="dg-small" x="840" y="383" text-anchor="middle">volume dbdata</text>

      <!-- database edges (green = data) -->
      <path class="dg-edge dg-edge--green" d="M660,108 H700 V175 H736" />
      <path class="dg-edge dg-edge--green" d="M660,241 H736" />
      <path class="dg-edge dg-edge--green" d="M660,303 H736" />
      <path class="dg-edge dg-edge--green" d="M660,437 H736" />

      <rect class="dg-box" x="740" y="540" width="200" height="60" rx="4" />
      <text class="dg-title" x="840" y="562" text-anchor="middle">SILVERBULLET-GIT</text>
      <text class="dg-small" x="840" y="578" text-anchor="middle">commit · pull --rebase · push</text>
      <text class="dg-small" x="840" y="592" text-anchor="middle">every 300s, vault = notes repo</text>
      <path class="dg-edge dg-edge--green dg-edge--dashed" d="M740,570 H700 V456 H664" />
      <text class="dg-edge-label" x="706" y="520">ssh</text>

      <rect class="dg-box dg-box--tint" x="740" y="676" width="200" height="154" rx="4" style="stroke: var(--color-secondary)" />
      <text class="dg-title" x="840" y="700" text-anchor="middle" style="fill: var(--color-secondary)">GARAGE</text>
      <text class="dg-small" x="840" y="716" text-anchor="middle">S3 object store</text>
      <text class="dg-code" x="756" y="740">:3900 S3 API</text>
      <text class="dg-code" x="756" y="760">:3903 admin API</text>
      <text class="dg-code" x="756" y="816">:3902 website</text>
      <text class="dg-small" x="924" y="760" text-anchor="end">← webui</text>
      <text class="dg-small" x="924" y="740" text-anchor="end">← backend</text>
      <path class="dg-edge dg-edge--green" d="M660,699 H700 V756 H736" />

      <!-- backend → garage over S3, round the right of the db -->
      <path class="dg-edge dg-edge--green" d="M660,74 H970 V736 H944" />
      <text class="dg-edge-label" x="820" y="68" text-anchor="middle">S3 PUT · S3_ENDPOINT=http://garage:3900</text>

      <!-- ---- Column E: host level, off the network ---- -->
      <rect class="dg-group dg-group--pink" x="1014" y="60" width="158" height="210" rx="6" />
      <text class="dg-group-label" x="1024" y="80" style="fill: var(--color-tertiary)">NETWORK_MODE: NONE</text>
      <rect class="dg-box" x="1024" y="92" width="138" height="62" rx="4" />
      <text class="dg-title" x="1093" y="114" text-anchor="middle">TRAFFIC</text>
      <text class="dg-small" x="1093" y="130" text-anchor="middle">tail -F the nginx</text>
      <text class="dg-small" x="1093" y="144" text-anchor="middle">payload log</text>
      <rect class="dg-box dg-box--pink" x="1024" y="172" width="138" height="86" rx="4" />
      <text class="dg-title dg-title--pink" x="1093" y="194" text-anchor="middle">AUTOHEAL</text>
      <text class="dg-small" x="1093" y="210" text-anchor="middle">restarts containers</text>
      <text class="dg-small" x="1093" y="224" text-anchor="middle">labelled autoheal</text>
      <text class="dg-small" x="1093" y="238" text-anchor="middle">pinned by digest</text>

      <rect class="dg-box dg-box--pink dg-box--deep" x="1014" y="310" width="158" height="46" rx="4" />
      <text class="dg-code" x="1093" y="338" text-anchor="middle">/var/run/docker.sock</text>
      <path class="dg-edge dg-edge--pink" d="M1093,258 V306" />
      <text class="dg-edge-label dg-edge-label--pink" x="1099" y="288">rw</text>

      <rect class="dg-box dg-box--pink" x="1014" y="400" width="158" height="104" rx="4" />
      <text class="dg-title dg-title--pink" x="1093" y="422" text-anchor="middle">ACT_RUNNER</text>
      <text class="dg-small" x="1093" y="439" text-anchor="middle">outside Docker, host mode</text>
      <text class="dg-code" x="1093" y="457" text-anchor="middle">gitea-runner/run.sh</text>
      <text class="dg-small" x="1093" y="475" text-anchor="middle">polls localhost:3000,</text>
      <text class="dg-small" x="1093" y="489" text-anchor="middle">runs docker compose</text>
      <path class="dg-edge dg-edge--pink" d="M1093,400 V360" />
      <text class="dg-edge-label dg-edge-label--pink" x="1099" y="384">compose up</text>

      <!-- published ports table -->
      <rect class="dg-box dg-box--deep" x="1014" y="548" width="158" height="164" rx="4" />
      <text class="dg-title dg-title--sm" x="1093" y="568" text-anchor="middle">PUBLISHED TO HOST</text>
      <line class="dg-rule" x1="1022" y1="578" x2="1164" y2="578" />
      <text class="dg-code" x="1024" y="598">80, 443</text>
      <text class="dg-small" x="1162" y="598" text-anchor="end">nginx</text>
      <text class="dg-code" x="1024" y="620">3000</text>
      <text class="dg-small" x="1162" y="620" text-anchor="end">gitea http</text>
      <text class="dg-code" x="1024" y="642">2222</text>
      <text class="dg-small" x="1162" y="642" text-anchor="end">gitea ssh</text>
      <text class="dg-code" x="1024" y="664">8001</text>
      <text class="dg-small" x="1162" y="664" text-anchor="end">liquidsoap harbor</text>
      <text class="dg-small" x="1093" y="694" text-anchor="middle">all else: app-network only</text>

      <!-- legend -->
      <g transform="translate(200,886)">
        <line class="dg-edge dg-edge--plain" x1="0" y1="0" x2="34" y2="0" />
        <text class="dg-small" x="42" y="4">request path (proxied)</text>
        <line class="dg-edge dg-edge--green dg-edge--plain" x1="200" y1="0" x2="234" y2="0" />
        <text class="dg-small" x="242" y="4">data: SQL, S3, git</text>
        <line class="dg-edge dg-edge--pink dg-edge--plain" x1="380" y1="0" x2="414" y2="0" />
        <text class="dg-small" x="422" y="4">visitor / host control</text>
        <line class="dg-edge dg-edge--muted dg-edge--dashed dg-edge--plain" x1="570" y1="0" x2="604" y2="0" />
        <text class="dg-small" x="612" y="4">optional path</text>
        <line class="dg-edge dg-edge--flow dg-edge--plain" x1="720" y1="0" x2="754" y2="0" />
        <text class="dg-small" x="762" y="4">the live ingress (tunnel)</text>
      </g>
    </svg>
    <template #caption>
      Twenty containers on one bridge network, <code>app-network</code>, plus the CI runner on the host itself.
      Requests come in through the Cloudflare tunnel (<code>cloudflared</code> dials out, so nothing has to be
      forwarded) or directly on 80/443. Either way they land on <strong>nginx</strong>, the only thing that sees the
      internet, which fans them out by path prefix to twelve upstreams it finds by container name. Behind it,
      <code>backend</code>, <code>gitea</code>, <code>jobs</code> and <code>activity-tracker</code> share one
      Postgres with a database each. Only four ports are published: 80/443 for nginx, 3000/2222 for Gitea (the runner and
      git clients need them), and liquidsoap's harbor for live DJs. The two containers that hold the Docker socket
      read-write (<code>autoheal</code>) or tail a file (<code>traffic</code>) get no network at all.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
