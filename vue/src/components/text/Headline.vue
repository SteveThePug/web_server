<script setup>
/**
 * Infinite horizontal marquee. Renders its slot twice side by side and
 * translates the pair left; when the offset passes one copy's width it wraps by
 * adding that width back, so the seam is never visible.
 *
 * The offset is a plain `let`, not a ref, and the transform is written straight
 * to element.style — a reactive ref here would queue a Vue re-render on every one
 * of the 60 frames per second. The width is measured once and re-measured by a
 * ResizeObserver rather than read each frame, because reading offsetWidth in the
 * rAF callback forces a layout on every frame.
 */
import { onMounted, useTemplateRef, onUnmounted } from "vue";

const container = useTemplateRef("container");
const item1 = useTemplateRef("item1");

let offset = 0;
let cachedWidth = 0;

let rafId;

const speed = 0.5; // pixels per frame

function measureWidth() {
  const ctnr = container.value;
  const it1 = item1.value;
  if (ctnr && it1) {
    cachedWidth = Math.max(ctnr.offsetWidth, it1.scrollWidth);
  }
}

function animate() {
  const ctnr = container.value;
  if (!ctnr || cachedWidth === 0) {
    rafId = requestAnimationFrame(animate);
    return;
  }

  offset -= speed;

  if (offset <= -cachedWidth) {
    offset += cachedWidth;
  }

  ctnr.style.transform = `translateX(${offset}px)`;

  rafId = requestAnimationFrame(animate);
}

let resizeObserver;

onMounted(() => {
  measureWidth();
  rafId = requestAnimationFrame(animate);

  resizeObserver = new ResizeObserver(measureWidth);
  resizeObserver.observe(container.value);
});

onUnmounted(() => {
  cancelAnimationFrame(rafId);
  resizeObserver?.disconnect();
});
</script>

<template>
  <!-- The slot is rendered twice: the second copy fills the gap the first
       leaves as it scrolls off, which is what makes the loop seamless. -->
  <div class="root">
    <div class="container" ref="container">
      <div ref="item1">
        <slot />
      </div>
      <div>
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
.root {
  overflow: hidden;
}

.container {
  width: fit-content;
  height: fit-content;
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: max-content;
  /* Each column fits its content */
  overflow-x: visible;
  will-change: transform;
  gap: 10em;
}
</style>
