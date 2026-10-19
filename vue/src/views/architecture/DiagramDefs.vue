<!--
  The arrowhead markers every diagram on /architecture uses, declared once.
  diagram.css points the .dg-edge classes at these ids, so a diagram never
  declares its own <defs> (fifty inline SVGs repeating the same ids would
  collide).

  Hidden by size rather than display:none — Firefox won't render markers
  that live in a display:none SVG.
-->
<template>
  <svg class="defs" aria-hidden="true" focusable="false">
    <defs>
      <marker
        v-for="m in markers"
        :id="m.id"
        :key="m.id"
        viewBox="0 0 10 10"
        refX="9"
        refY="5"
        markerWidth="7"
        markerHeight="7"
        :orient="m.start ? 'auto-start-reverse' : 'auto'"
      >
        <path d="M0,0 L10,5 L0,10 z" :style="{ fill: m.colour }" />
      </marker>
    </defs>
  </svg>
</template>

<script setup>
const colours = {
  "": "var(--color-primary)",
  "-pink": "var(--color-tertiary)",
  "-green": "var(--color-secondary)",
  "-muted": "var(--color-muted)",
};

const markers = Object.entries(colours).flatMap(([suffix, colour]) => [
  { id: `dg-arrow${suffix}`, colour, start: false },
  { id: `dg-arrow-start${suffix}`, colour, start: true },
]);
</script>

<style scoped>
.defs {
  position: absolute;
  width: 0;
  height: 0;
  overflow: hidden;
}
</style>
