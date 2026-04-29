<script setup>
import { ref, onMounted, useTemplateRef, onUnmounted } from "vue";
import { getRandomColor } from "@/js/utils";

const container = ref(null);

// List of (offset, width)
function generateOffsets(width = 100, step = 10, n = 20) {
  return Array.from({ length: n }, (_, i) => ({
    width,
    offset: step * i,
    color: getRandomColor(),
  }));
}
const offsets = ref(generateOffsets((150, 15, 10)));
let rafId;

const speed = 0.5; // pixels per frame

function animate() {
  const ctnr = container.value;
  for (const item of offsets.value) {
    const width = Math.max(ctnr.offsetWidth, item.width);

    console.log(ctnr.offsetWidth);

    item.offset -= speed;
    if (item.offset <= -width) {
      item.color = getRandomColor();
      item.offset = 0;
    }
  }
  rafId = requestAnimationFrame(animate);
}

onMounted(() => {
  rafId = requestAnimationFrame(animate);
});

onUnmounted(() => {
  cancelAnimationFrame(rafId);
});
</script>

<template>
  <div class="bg-primary container" ref="container">
    <div :key="index" v-for="(item, index) in offsets">
      <div
        :style="{
          width: item.width + 'px',
          translate: item.offset + 'px',
          backgroundColor: item.color,
        }"
        class="item item1"
      />
      <div
        :style="{
          width: item.width + 'px',
          right: -item.width + 'px',
          translate: item.offset + 'px',
          backgroundColor: item.color,
        }"
        class="item item2"
      />
    </div>
  </div>
</template>

<style scoped>
.container {
  position: relative;
  overflow: hidden;
  width: 100%;
  will-change: transform;
}
.item {
  opacity: 40%;
  height: 100%;
  position: absolute;
}
.item1 {
  left: 0px;
  top: 0px;
}
.item2 {
  top: 0px;
}
</style>
