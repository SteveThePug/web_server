<script setup>
/**
 * STATUS: legacy — no route points here. `/stp` is now an ordinary canvas
 * Page (slug "stp", created by ensureStpPage in backend/services/database.go)
 * served by views/canvas/CanvasPage.vue; the "Old /stp grid" starter in
 * views/canvas/pageStarters.js replaces this view's import button. Kept for
 * reference.
 *
 * Was route `/stp` — the home page as a free-form canvas of boxes the admin
 * arranges in the browser (page key "stp"). Replaces views/home/Home.vue's
 * fixed grid. The canvas itself is BoxCanvas.vue, shared with the landing
 * page's notes and the admin-created pages.
 */
import BoxCanvas from "./BoxCanvas.vue";
import { defaultLayout } from "./defaultLayout";
</script>

<template>
  <main class="home3">
    <BoxCanvas page="stp" :starter="defaultLayout" starter-label="Import current layout" />
  </main>
</template>

<style scoped>
.home3 {
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
  padding: 10px 16px;
}
</style>
