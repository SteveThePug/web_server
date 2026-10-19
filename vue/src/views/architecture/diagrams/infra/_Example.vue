<!--
  Reference example for the diagram kit. Not used on the page.
-->
<template>
  <DiagramFigure id="example" title="Example: a request through two services">
    <svg class="dg" viewBox="0 0 760 260" role="img" aria-labelledby="ex-t">
      <title id="ex-t">Browser to nginx to backend, with a database</title>
      <rect class="dg-group dg-group--fill" x="200" y="20" width="540" height="220" rx="6" />
      <text class="dg-group-label" x="214" y="40">DOCKER NETWORK · app-network</text>

      <rect class="dg-box dg-box--pink" x="20" y="100" width="130" height="60" rx="4" />
      <text class="dg-title dg-title--pink" x="85" y="126" text-anchor="middle">BROWSER</text>
      <text class="dg-small" x="85" y="144" text-anchor="middle">visitor</text>

      <rect class="dg-box" x="240" y="100" width="150" height="60" rx="4" />
      <text class="dg-title" x="315" y="126" text-anchor="middle">NGINX</text>
      <text class="dg-code" x="315" y="144" text-anchor="middle">:80 / :443</text>

      <rect class="dg-box dg-box--tint" x="470" y="70" width="230" height="120" rx="4" />
      <text class="dg-title" x="585" y="96" text-anchor="middle">BACKEND</text>
      <text class="dg-text" x="490" y="122">Gin router</text>
      <line class="dg-rule" x1="480" y1="132" x2="690" y2="132" />
      <text class="dg-code" x="490" y="152">POST /api/graphql</text>
      <text class="dg-code" x="490" y="172">GET  /api/ws</text>

      <path class="dg-edge dg-edge--pink" d="M150,130 H236" />
      <text class="dg-edge-label dg-edge-label--pink" x="193" y="122" text-anchor="middle">HTTPS</text>
      <path class="dg-edge dg-edge--flow" d="M390,130 H466" />
      <text class="dg-edge-label" x="428" y="122" text-anchor="middle">proxy</text>

      <circle class="dg-badge" cx="240" cy="100" r="10" />
      <text class="dg-badge-text" x="240" y="100">1</text>
    </svg>
    <template #caption>
      Captions explain what the picture shows and why it is that way, with
      <code>code</code> names where they help.
    </template>
  </DiagramFigure>
</template>

<script setup>
import DiagramFigure from "../../DiagramFigure.vue";
</script>
