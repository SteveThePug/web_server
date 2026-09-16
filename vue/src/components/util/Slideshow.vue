<script setup>
/**
 * Crossfading image rotator. Clicking advances immediately.
 *
 * Each advance clears and re-arms its own timeout rather than using setInterval,
 * so a manual click restarts the full interval instead of leaving a short
 * remainder. The fade works because `:key="currentIndex"` makes Vue treat each
 * image as a new element for the global `fade` transition, and the wrapper is
 * `position: relative` so the outgoing copy can be absolutely positioned over it.
 */
import { ref, computed, onMounted, onUnmounted } from "vue";
import Header from "@/components/text/Header.vue";

const props = defineProps({
  images: {
    type: Array,
    required: true,
  },
  interval: {
    type: Number,
    default: 10000,
  },
});

const currentIndex = ref(0);
const currentComment = computed(() => props.images[currentIndex.value].comment);
const currentUrl = computed(() => props.images[currentIndex.value].url);

let nextId;

function nextImage() {
  clearTimeout(nextId);
  currentIndex.value = (currentIndex.value + 1) % props.images.length;
  nextId = setTimeout(nextImage, props.interval);
}

onMounted(() => {
  nextId = setTimeout(nextImage, props.interval);
});

onUnmounted(() => {
  clearTimeout(nextId);
});
</script>

<template>
  <div class="slideshow-wrapper">
    <Transition name="fade">
      <div class="image-viewer" @click="nextImage" :key="currentIndex">
        <Header v-if="currentComment">
          {{ currentComment }}
        </Header>
        <img :src="currentUrl" alt="Image Viewer" fetchpriority="high" />
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.slideshow-wrapper {
  position: relative; /* for the global .fade transition */
  width: 100%;
  overflow: hidden;
}

.image-viewer {
  width: 100%;
  overflow: hidden;
}

img {
  width: 100%;
  object-fit: cover;
  display: block;
}
</style>
