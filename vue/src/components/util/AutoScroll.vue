<template>
  <div
    ref="container"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
    class="overflow-y-auto"
  >
    <slot />
  </div>
</template>

<script setup>
/**
 * Scroll container that drifts its own content up and down forever, pausing at
 * each end. Used by the Favorites and Consumption widgets.
 *
 * Position is kept as a 0..1 fraction so it survives content resizing. On the
 * edges it is clamped to 0.999 / 0.001 rather than exactly 1 / 0 so the next
 * tick's `reachedBottom` / `reachedTop` test doesn't immediately fire again and
 * re-trigger the pause. scrollHeight is measured by a ResizeObserver instead of
 * being read every frame, which would force a layout 60 times a second.
 */
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";

const container = useTemplateRef("container");

const SPEED = 0.0005; // % per frame
const PAUSE = 2000; // ms at top/bottom

let pos = 0;
let direction = 1; // 1 = down, -1 = up
let hovered = false;
let rafId = null;
let pauseTimeoutId = null;
let cachedScrollHeight = 0;

function measureScrollHeight() {
  const el = container.value;
  // Scrollable range, not full content height: scrollTop saturates at
  // scrollHeight - clientHeight, so using scrollHeight would stop the drift
  // one viewport short of the end while pos still climbed to 1.
  if (el) cachedScrollHeight = Math.max(0, el.scrollHeight - el.clientHeight);
}

function stopLoop() {
  if (rafId !== null) {
    cancelAnimationFrame(rafId);
    rafId = null;
  }
  if (pauseTimeoutId !== null) {
    clearTimeout(pauseTimeoutId);
    pauseTimeoutId = null;
  }
}

function startLoop() {
  stopLoop();
  rafId = requestAnimationFrame(tick);
}

function onMouseEnter() {
  hovered = true;
  stopLoop();
}

function onMouseLeave() {
  hovered = false;
  const el = container.value;
  if (el && cachedScrollHeight > 0) {
    pos = el.scrollTop / cachedScrollHeight;
  }
  startLoop();
}

function schedulePause(callback) {
  stopLoop();
  pauseTimeoutId = setTimeout(callback, PAUSE);
}

function tick() {
  rafId = null;
  const el = container.value;
  if (hovered) return;

  if (!el || cachedScrollHeight === 0) {
    rafId = requestAnimationFrame(tick);
    return;
  }

  const reachedBottom = pos >= 1;
  const reachedTop = pos <= 0;

  if (reachedBottom) {
    pos = 0.999;
    direction = -1;
    schedulePause(startLoop);
    return;
  } else if (reachedTop && direction === -1) {
    pos = 0.001;
    direction = 1;
    schedulePause(startLoop);
    return;
  }

  pos += direction * SPEED;

  el.scrollTop = pos * cachedScrollHeight;

  rafId = requestAnimationFrame(tick);
}

let resizeObserver;

onMounted(() => {
  measureScrollHeight();
  schedulePause(startLoop);

  resizeObserver = new ResizeObserver(measureScrollHeight);
  resizeObserver.observe(container.value);
});

onBeforeUnmount(() => {
  stopLoop();
  resizeObserver?.disconnect();
});
</script>
