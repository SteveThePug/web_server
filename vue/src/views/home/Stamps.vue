<script setup>
/**
 * 88x31 webring "stamps" widget on /stp, inside a drag-scrollable <Touchscreen>.
 *
 * The order is shuffled on module load so the wall looks different each visit.
 * The rAF loop then drifts the scroll position diagonally and bounces off each
 * edge — it drives the DOM's scrollLeft/scrollTop directly rather than any
 * reactive state, and reads scrollWidth/clientWidth each frame because the
 * wall is re-sized whenever the cell is.
 *
 * The wall is a fixed-track grid whose column and row counts are derived from
 * the cell's own size (see fitWall), so it always overflows on both axes. A
 * wrapping flex row only overflowed when the cell happened to be narrower than
 * the wall, which left it motionless at tablet widths.
 */
import { ref, computed, onMounted, onUnmounted } from "vue";

import Touchscreen from "@/components/util/Touchscreen.vue";
import Link from "@/components/text/Link.vue";
import { shuffleArray } from "@/js/utils.js";

let srcs = [
  "/img/stamps/portal.gif",
  "/img/stamps/miku.gif",
  "/img/stamps/utau.gif",
  "/img/stamps/teto.webp",
  "/img/stamps/3ds.jpg",
  "/img/stamps/fry.png",
  "/img/stamps/ai.png",
  "/img/stamps/rei.png",
  "/img/stamps/tetris.gif",
  "/img/stamps/tf2.gif",
  "/img/stamps/demo.gif",
  "/img/stamps/demo.gif",
  "/img/stamps/demo.gif",
  "/img/stamps/demo.gif",
];
// Shuffled once at module evaluation, so the wall is ordered differently each
// page load but stable while the widget is mounted. Mutates in place.
shuffleArray(srcs);

// The two linked stamps stay first; the shuffled local ones follow.
const stamps = [
  {
    src: "https://www.adam-french.co.uk/img/stamps/mine.gif",
    alt: "adam-french.co.uk",
    href: "https://www.adam-french.co.uk",
  },
  {
    src: "https://jacobbarron.xyz/Banneh.gif",
    alt: "jacobbarron.xyz",
    href: "https://jacobbarron.xyz",
  },
  ...srcs.map((src) => ({ src, alt: src.split("/").pop().split(".")[0] })),
];

// Must match the grid track sizes in the <style> block.
const STAMP_W = 89;
const STAMP_H = 59;

const touchscreen = ref(null);
const cols = ref(4);
const rows = ref(4);

// Every grid slot is filled, cycling through the stamps when the wall needs
// more slots than there are stamps — a ragged last row would leave a blank
// strip drifting into view.
const cells = computed(() =>
  Array.from(
    { length: cols.value * rows.value },
    (_, i) => stamps[i % stamps.length],
  ),
);

/**
 * Size the wall to between one and two stamps larger than the cell on each
 * axis, so there is always something to bounce across whatever the page size.
 * Rows are also raised, if need be, until every stamp appears at least once.
 */
function fitWall() {
  const el = touchscreen.value?.$el;
  if (!el) return;

  cols.value = Math.ceil(el.clientWidth / STAMP_W) + 1;
  rows.value = Math.max(
    Math.ceil(el.clientHeight / STAMP_H) + 1,
    Math.ceil(stamps.length / cols.value),
  );
}

let resizeObserver = null;
let animId = null;
let posX = 0;
let posY = 0;
let dx = 0.2;
let dy = 0.12;

/**
 * rAF loop drifting the scroll position diagonally and bouncing off the edges.
 *
 * Writes scrollLeft/scrollTop on the DOM node directly (no reactive state, so
 * no re-render per frame), and re-reads scrollWidth/clientWidth every frame
 * because fitWall changes the bounds whenever the cell is resized.
 */
function bounce() {
  const el = touchscreen.value?.$el;
  if (!el) return;

  const maxX = el.scrollWidth - el.clientWidth;
  const maxY = el.scrollHeight - el.clientHeight;

  if (maxX > 0) {
    posX += dx;
    if (posX <= 0) {
      posX = 0;
      dx = -dx;
    } else if (posX >= maxX) {
      posX = maxX;
      dx = -dx;
    }
    el.scrollLeft = posX;
  }
  if (maxY > 0) {
    posY += dy;
    if (posY <= 0) {
      posY = 0;
      dy = -dy;
    } else if (posY >= maxY) {
      posY = maxY;
      dy = -dy;
    }
    el.scrollTop = posY;
  }

  animId = requestAnimationFrame(bounce);
}

onMounted(() => {
  fitWall();
  // The cell's size follows the home grid's breakpoints (and the scrollbars
  // appearing inside it), not just window resizes, so observe the element.
  resizeObserver = new ResizeObserver(fitWall);
  resizeObserver.observe(touchscreen.value.$el);
  animId = requestAnimationFrame(bounce);
});

onUnmounted(() => {
  if (animId) cancelAnimationFrame(animId);
  resizeObserver?.disconnect();
});
</script>

<template>
  <Touchscreen ref="touchscreen">
    <div class="wall" :style="{ '--cols': cols }">
      <template v-for="(stamp, i) in cells" :key="i">
        <Link v-if="stamp.href" bare :href="stamp.href">
          <img :src="stamp.src" :alt="stamp.alt" />
        </Link>
        <img v-else :src="stamp.src" :alt="stamp.alt" loading="lazy" />
      </template>
    </div>
  </Touchscreen>
</template>

<style scoped>
img {
  display: block;
  width: 89px;
  height: 59px;
}

/* Fixed tracks rather than flex-wrap: the wall's size comes from --cols (set
   by fitWall), never from how much room the cell happens to have. */
.wall {
  display: grid;
  grid-template-columns: repeat(var(--cols), 89px);
  grid-auto-rows: 59px;
  width: max-content;
}
</style>
