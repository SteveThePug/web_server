<!--
  How nginx/entrypoint.sh picks one of three configs at container start, then
  waits for the Vue build, starts the certificate watcher and execs nginx.
  Reflects nginx/entrypoint.sh, nginx/Dockerfile and the DEV_MODE switch in
  docker-compose.dev.yml. Re-check when the entrypoint's branches change.
-->
<template>
  <DiagramFigure id="edge-entrypoint-flow" title="Container start: which nginx.conf gets written" :min-width="770">
    <svg class="dg" viewBox="0 0 1100 960" role="img" aria-labelledby="edge-entrypoint-flow-t">
      <title id="edge-entrypoint-flow-t">Flowchart of nginx/entrypoint.sh choosing the dev, production or setup config</title>

      <!-- start -->
      <rect class="dg-box dg-box--solid" x="390" y="16" width="320" height="40" rx="20" />
      <text class="dg-title dg-text--ink" x="550" y="42" text-anchor="middle" style="fill: var(--color-surface-deep)">/ENTRYPOINT.SH STARTS</text>
      <path class="dg-edge" d="M550,56 V76" />

      <!-- 1: routes partial -->
      <rect class="dg-box dg-box--tint" x="250" y="80" width="600" height="58" rx="4" />
      <text class="dg-code" x="550" y="104" text-anchor="middle">envsubst "$ENVSUBST_VARS" &lt; nginx_routes.conf.template &gt; /etc/nginx/nginx_routes.conf</text>
      <text class="dg-small" x="550" y="124" text-anchor="middle">done once, before the choice: dev and production include it, and the watcher's nginx -t needs it later</text>
      <circle class="dg-badge" cx="250" cy="80" r="10" />
      <text class="dg-badge-text" x="250" y="80">1</text>
      <path class="dg-edge" d="M550,138 V156" />

      <!-- diamond 1: DEV_MODE -->
      <path class="dg-box" d="M550,160 L650,200 L550,240 L450,200 Z" />
      <text class="dg-code" x="550" y="197" text-anchor="middle">DEV_MODE</text>
      <text class="dg-code" x="550" y="212" text-anchor="middle">= "true"?</text>
      <circle class="dg-badge" cx="450" cy="200" r="10" />
      <text class="dg-badge-text" x="450" y="200">2</text>

      <path class="dg-edge" d="M440,200 H190 V366" />
      <text class="dg-edge-label dg-edge-label--mint" x="300" y="192" text-anchor="middle">yes</text>
      <path class="dg-edge" d="M550,240 V266" />
      <text class="dg-edge-label dg-edge-label--mint" x="566" y="257">no</text>

      <!-- diamond 2: cert present -->
      <path class="dg-box" d="M550,270 L680,312 L550,354 L420,312 Z" />
      <text class="dg-code" x="550" y="308" text-anchor="middle">fullchain.pem + privkey.pem</text>
      <text class="dg-code" x="550" y="323" text-anchor="middle">in live/$DOMAIN/ ?</text>
      <path class="dg-edge" d="M550,354 V366" />
      <text class="dg-edge-label dg-edge-label--mint" x="566" y="364">yes</text>
      <path class="dg-edge" d="M680,312 H910 V366" />
      <text class="dg-edge-label dg-edge-label--mint" x="795" y="304" text-anchor="middle">no</text>

      <!-- three configs -->
      <rect class="dg-box dg-box--muted" x="30" y="370" width="320" height="160" rx="4" />
      <text class="dg-title" x="46" y="396">DEV · NGINX_DEV.CONF.TEMPLATE</text>
      <text class="dg-small" x="46" y="416">self-signed cert, minted once, kept in certbot/conf</text>
      <text class="dg-code" x="46" y="438">openssl req -x509 … -subj /CN=localhost</text>
      <text class="dg-code" x="46" y="458">envsubst → nginx_dev_common.conf</text>
      <text class="dg-code" x="46" y="476">envsubst → nginx.conf</text>
      <text class="dg-small" x="46" y="496">:80 and :443 both serve; / → vue:5173</text>
      <text class="dg-code" x="46" y="518" style="fill: var(--color-primary)">CONFIG_MODE=dev</text>

      <rect class="dg-box" x="390" y="370" width="320" height="160" rx="4" />
      <text class="dg-title" x="406" y="396">PRODUCTION · NGINX.CONF.TEMPLATE</text>
      <text class="dg-small" x="406" y="416">steady state: the Let's Encrypt cert is on disk</text>
      <text class="dg-code" x="406" y="438">envsubst "$ENVSUBST_VARS"</text>
      <text class="dg-code" x="406" y="456">  &lt; nginx.conf.template &gt; nginx.conf</text>
      <text class="dg-small" x="406" y="478">:80 → 301 https · apex → www · TLS on :443</text>
      <text class="dg-small" x="406" y="496">includes nginx_routes.conf</text>
      <text class="dg-code" x="406" y="518" style="fill: var(--color-primary)">CONFIG_MODE=prod</text>

      <rect class="dg-box dg-box--pink" x="750" y="370" width="320" height="160" rx="4" />
      <text class="dg-title dg-title--pink" x="766" y="396">SETUP · NGINX_SETUP.CONF.TEMPLATE</text>
      <text class="dg-small" x="766" y="416">first boot on a new host: no certificate yet</text>
      <text class="dg-code" x="766" y="438">envsubst '${DOMAIN}'</text>
      <text class="dg-code" x="766" y="456">  &lt; nginx_setup.conf.template</text>
      <text class="dg-small" x="766" y="478">:80 only: ACME webroot, /nginx-health, 404</text>
      <text class="dg-small" x="766" y="496">no TLS, no upstreams: nothing can fail to load</text>
      <text class="dg-code" x="766" y="518" style="fill: var(--color-primary)">CONFIG_MODE=setup</text>

      <!-- merge -->
      <path class="dg-edge dg-edge--plain" d="M190,530 V556 H910 V530" />
      <path class="dg-edge" d="M550,530 V576" />

      <rect class="dg-box" x="370" y="580" width="360" height="44" rx="4" />
      <text class="dg-code" x="550" y="600" text-anchor="middle">chmod 755 /uploads</text>
      <text class="dg-small" x="550" y="616" text-anchor="middle">workers must traverse the volume the backend created</text>
      <path class="dg-edge" d="M550,624 V646" />

      <!-- diamond 3 -->
      <path class="dg-box" d="M550,650 L640,684 L550,718 L460,684 Z" />
      <text class="dg-code" x="550" y="681" text-anchor="middle">DEV_MODE</text>
      <text class="dg-code" x="550" y="696" text-anchor="middle">= "true"?</text>
      <path class="dg-edge" d="M550,718 V746" />
      <text class="dg-edge-label dg-edge-label--mint" x="566" y="738">no</text>
      <path class="dg-edge dg-edge--muted dg-edge--dashed" d="M640,684 H960 V922 H734" />
      <text class="dg-edge-label" x="800" y="676" text-anchor="middle">yes: Vite serves /, no watcher</text>

      <rect class="dg-box" x="370" y="750" width="360" height="52" rx="4" />
      <text class="dg-text" x="550" y="771" text-anchor="middle">wait up to 120 s for /etc/nginx/html/index.html</text>
      <text class="dg-small" x="550" y="790" text-anchor="middle">vue_dist is filled by the one-shot vue build; then start anyway</text>
      <path class="dg-edge" d="M550,802 V822" />

      <rect class="dg-box" x="370" y="826" width="360" height="52" rx="4" />
      <text class="dg-code" x="550" y="847" text-anchor="middle">cert_watcher &amp;</text>
      <text class="dg-small" x="550" y="866" text-anchor="middle">background loop, every $CERT_WATCH_INTERVAL (60 s)</text>
      <circle class="dg-badge" cx="370" cy="826" r="10" />
      <text class="dg-badge-text" x="370" y="826">3</text>
      <path class="dg-edge" d="M550,878 V898" />

      <rect class="dg-box dg-box--solid" x="370" y="902" width="360" height="40" rx="20" />
      <text class="dg-code" x="550" y="927" text-anchor="middle" style="fill: var(--color-surface-deep)">exec: nginx -g 'daemon off;'  (PID 1)</text>
    </svg>
    <template #caption>
      No <code>nginx.conf</code> is baked into the image; three templates are,
      and the entrypoint writes exactly one of them on every start. Step
      <strong>1</strong> renders the shared route table first, whatever happens
      next. At <strong>2</strong>, <code>DEV_MODE</code> (set only by
      <code>docker-compose.dev.yml</code>) wins outright; otherwise the presence
      of a certificate decides, because the production config's
      <code>ssl_certificate</code> lines would stop nginx from starting without
      one. The <code>envsubst</code> allow-list is what keeps nginx's own
      <code>$host</code> and <code>$remote_addr</code> intact: only the named
      <code>${DOMAIN} ${BACKEND_HOST} …</code> variables are replaced. The
      watcher at <strong>3</strong> moves a setup-mode nginx onto the
      production config once certbot's first certificate lands (render to a
      temp file, <code>nginx -t</code>, install, <code>nginx -s reload</code>),
      and reloads after a renewal when certbot's <code>--deploy-hook</code>
      touches <code>/etc/letsencrypt/.nginx-reload</code>. Grey dashed = local
      dev only.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
