<script setup>
/**
 * Clock and date widget. updateDateTime() runs once immediately and then every
 * minute. Note the interval is module-level and never cleared — harmless because
 * the widget lives for the life of the home page, but it would leak if reused.
 */
import Header from "@/components/text/Header.vue";
import { ref } from "vue";

const time = ref("");
const weekday = ref("");
const day = ref("");
const month = ref("");

function updateDateTime() {
  const date = new Date();
  day.value = date.getDate();
  time.value = date.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
  });
  weekday.value = date.toLocaleDateString("en-GB", { weekday: "long" });
  month.value = date.toLocaleDateString("en-GB", { month: "long" });
}

updateDateTime();

setInterval(updateDateTime, 60000);
</script>

<template>
  <div class="flex flex-col">
    <Header>{{ weekday }} {{ day }}, {{ month }}</Header>
    <h1>{{ time }}</h1>
  </div>
</template>

<style scoped>
div {
  text-align: center;
  padding: 4px;
}
</style>
