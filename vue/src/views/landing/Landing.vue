<script setup>
/**
 * Route `/` — the front door, as a box canvas (page key "index") the admin
 * arranges in the browser like /stp. The fixed bio-and-links page it
 * replaced is kept as boxes in landingLayout.js, behind the canvas's import
 * button.
 *
 * Statically imported by the router (it is the first paint), but the canvas
 * is loaded in its own chunk: it brings the box editor with it.
 */
import { defineAsyncComponent } from "vue";
import { landingLayout } from "./landingLayout";

const BoxCanvas = defineAsyncComponent(() =>
  import("@/views/home3/BoxCanvas.vue"),
);
</script>

<template>
  <main class="landing">
    <BoxCanvas page="index" :starter="landingLayout" starter-label="Import old landing page" />
  </main>
</template>

<style scoped>
.landing {
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
  padding: 10px 16px;
}
</style>
