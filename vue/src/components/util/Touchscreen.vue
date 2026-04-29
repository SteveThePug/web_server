<template>
  <div
    ref="container"
    class="overflow-auto w-full h-full"
    @mousedown="handleMouseDown"
    @mousemove="handleMouseMove"
    @mouseup="handleMouseUp"
    @mouseleave="handleMouseLeave"
    :style="{ cursor: isDragging ? 'grabbing' : 'grab' }"
  >
    <slot />
  </div>
</template>

<script setup>
import { ref } from "vue";

const container = ref(null);
const isDragging = ref(false);
const startX = ref(0);
const startY = ref(0);
const scrollLeft = ref(0);
const scrollTop = ref(0);

const handleMouseDown = (e) => {
  isDragging.value = true;
  startX.value = e.pageX - container.value.offsetLeft;
  startY.value = e.pageY - container.value.offsetTop;
  scrollLeft.value = container.value.scrollLeft;
  scrollTop.value = container.value.scrollTop;

  // Prevent text selection while dragging
  e.preventDefault();
};

const handleMouseMove = (e) => {
  if (!isDragging.value) return;

  e.preventDefault();

  const x = e.pageX - container.value.offsetLeft;
  const y = e.pageY - container.value.offsetTop;
  const walkX = (x - startX.value) * 1; // Multiply by scroll speed factor
  const walkY = (y - startY.value) * 1;

  container.value.scrollLeft = scrollLeft.value - walkX;
  container.value.scrollTop = scrollTop.value - walkY;
};

const handleMouseUp = () => {
  isDragging.value = false;
};

const handleMouseLeave = () => {
  isDragging.value = false;
};
</script>

<style scoped>
/* Prevent text selection while dragging */
div[style*="cursor: grabbing"] {
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}
</style>
