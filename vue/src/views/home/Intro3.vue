<script setup lang="ts">
/**
 * Intro tile: Bad Apple rendered as ASCII text.
 *
 * Source is `vue/public/img/badapple.mp4`, bound via a const rather than written
 * as a literal `src` because Vite would otherwise try to resolve it as a bundled
 * asset import; files under public/ are served as-is and must stay unbundled.
 *
 * The <video> stays in the DOM (hidden,
 * muted, looping) because only a playing element decodes frames — drawing a
 * detached/paused video to a canvas yields nothing.
 *
 * Per frame we downscale the video straight into a tiny canvas whose pixel grid
 * IS the character grid, so the browser's scaler does the averaging for us and
 * we read back exactly one pixel per character. Sampling is throttled to ~24fps
 * (source framerate); the grid is only re-measured on resize, since reading
 * offsetWidth every frame forces a synchronous layout.
 */
import { ref, useTemplateRef, onMounted, onUnmounted } from "vue";

const VIDEO_SRC = "/img/badapple.mp4";
// Dark -> light. Video is black-and-white, so a short ramp is plenty.
const RAMP = " .:-=+*#%@";
const CHAR_HEIGHT = 8;
const CHAR_WIDTH = CHAR_HEIGHT * 0.55;
const FRAME_INTERVAL = 1000 / 24;

const display = useTemplateRef<HTMLPreElement>("display");
const video = useTemplateRef<HTMLVideoElement>("video");
const displayText = ref("");

const canvas = document.createElement("canvas");
const ctx = canvas.getContext("2d", { willReadFrequently: true })!;

let cols = 0;
let rows = 0;
let rafId = 0;
let lastFrame = 0;

function measure() {
  const el = display.value;
  if (!el) return;
  cols = Math.max(1, Math.floor(el.offsetWidth / CHAR_WIDTH));
  rows = Math.max(1, Math.floor(el.offsetHeight / CHAR_HEIGHT));
  canvas.width = cols;
  canvas.height = rows;
}

function draw(now: number) {
  rafId = requestAnimationFrame(draw);
  if (now - lastFrame < FRAME_INTERVAL) return;
  lastFrame = now;

  const v = video.value;
  if (!v || v.readyState < 2 || !cols || !rows) return;

  ctx.drawImage(v, 0, 0, cols, rows);
  const data = ctx.getImageData(0, 0, cols, rows).data;

  let out = "";
  for (let y = 0; y < rows; y++) {
    for (let x = 0; x < cols; x++) {
      const i = (y * cols + x) * 4;
      // Greyscale source, so the red channel alone is the luminance.
      out += RAMP[Math.min(RAMP.length - 1, (data[i] * RAMP.length) >> 8)];
    }
    if (y < rows - 1) out += "\n";
  }
  displayText.value = out;
}

let resizeObserver: ResizeObserver;

onMounted(() => {
  measure();
  resizeObserver = new ResizeObserver(measure);
  resizeObserver.observe(display.value!);
  // Muted autoplay is the only kind browsers allow without a gesture.
  video.value?.play().catch(() => {});
  rafId = requestAnimationFrame(draw);
});

onUnmounted(() => {
  cancelAnimationFrame(rafId);
  resizeObserver?.disconnect();
});
</script>

<template>
  <div class="w-full h-full overflow-hidden bg-black">
    <video
      ref="video"
      :src="VIDEO_SRC"
      muted
      loop
      playsinline
      preload="auto"
      class="hidden"
    />
    <pre
      ref="display"
      class="w-full h-full m-0 p-0 overflow-hidden bg-black text-white select-none"
      :style="{ fontSize: CHAR_HEIGHT + 'px', lineHeight: CHAR_HEIGHT + 'px' }"
      >{{ displayText }}</pre
    >
  </div>
</template>
