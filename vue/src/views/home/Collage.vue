<script setup>
/**
 * "鸟" widget on /stp — a <Slideshow>. The Places list that used to share it
 * as a second tab is its own widget now, views/home/Places.vue.
 */
import Slideshow from "@/components/util/Slideshow.vue";

const images = [{ url: "/img/memes/pidgeon.gif", comment: "鸟" }];
</script>

<template>
  <div class="collage-cell">
    <Slideshow :images="images" />
  </div>
</template>

<style scoped>
.collage-cell {
  position: relative;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
