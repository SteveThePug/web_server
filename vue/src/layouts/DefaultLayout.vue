<script setup>
/**
 * Shell for everything under `/`: sticky navbar, scrolling <main>, sticky footer.
 *
 * The RouterView is wrapped in the global `slide` transition (defined in
 * assets/styles.css). `:key="$route.path"` forces a full remount on every path
 * change so the transition actually runs even between two routes that share a
 * component; `mode="out-in"` stops the outgoing and incoming pages overlapping.
 *
 * The background is the halftone unless the page on screen has one of its
 * own (stores/background.js — set per canvas page, landing page and shrines
 * included). The inline style overrides the halftone's layers.
 */
import { RouterView } from "vue-router";
import { useBackgroundStore } from "@/stores/background";
import Navbar from "@/components/Navbar.vue";
import Footer from "@/components/Footer.vue";

const background = useBackgroundStore();
</script>

<template>
  <div class="default-layout halftone" :style="background.style">
    <Navbar class="no-print sticky top-0 z-50" />
    <main class="default-content">
      <RouterView v-slot="{ Component }">
        <Transition name="slide" mode="out-in">
          <component :is="Component" :key="$route.path" />
        </Transition>
      </RouterView>
    </main>
    <Footer class="no-print sticky bottom-0 z-50" />
  </div>
</template>

<style scoped>
.default-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.default-content {
  flex: 1;
  overflow-y: auto;
}
</style>
