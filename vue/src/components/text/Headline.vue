<script setup>
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
