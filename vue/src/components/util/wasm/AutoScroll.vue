<template>
  <div ref="container" class="overflow-y-auto">
    <slot />
  </div>
</template>

<script setup>
/**
 * Rust/wasm port of components/util/AutoScroll.vue — the whole rAF loop lives in
 * the AutoScroller struct from src/wasm/.
 *
 * Note the two-step teardown: destroy() detaches the loop and listeners, free()
 * releases the Rust-side memory. Skipping free() leaks the wasm allocation, which
 * JS garbage collection cannot reclaim.
 *
 * STATUS: not imported anywhere; the JS AutoScroll is what the widgets use.
 */
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";
import { AutoScroller } from "@/wasm/stp_wasm.js";

const container = useTemplateRef("container");

let scroller = null;

onMounted(() => {
  if (!container.value) return;
  scroller = new AutoScroller(container.value);
  scroller.start();
});

onBeforeUnmount(() => {
  scroller?.destroy();
  scroller?.free();
  scroller = null;
});
</script>
