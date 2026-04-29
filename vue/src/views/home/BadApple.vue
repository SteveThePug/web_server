<script setup lang="ts">
import { useTemplateRef, ref, onMounted, onUnmounted } from "vue";

const display = useTemplateRef("display");
const displayText = ref("");

const charHeight: number = 14;
const charWidth: number = charHeight * 0.6;
let n: number;
let m: number;

function setup() {
  display.value.style.fontSize = `${charHeight}px`;
  display.value.style.lineHeight = `${charHeight}px`;
  fillDisplay();
}

function fillDisplay() {
  // M rows N columns
  m = Math.floor(display.value.offsetHeight / charHeight);
  n = Math.floor(display.value.offsetWidth / charWidth);
  const row = " ".repeat(n);
  displayText.value = (row + "\n").repeat(m - 1) + row;
}

function close() {
  displayText.value = "";
}

onMounted(() => {
  setup();
});

onUnmounted(() => {
  close();
});
</script>

<template>
  <pre
    class="overflow-scroll w-full h-full bg-black text-white m-0 p-0"
    id="container"
    ref="display"
    >{{ displayText }}</pre
  >
</template>
