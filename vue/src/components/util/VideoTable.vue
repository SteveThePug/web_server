<script setup>
/**
 * Grid of <video> players from a list of {link}. The MIME type is guessed from
 * the file extension. `preload="none"` keeps the page from downloading every clip
 * on load. Used by the Demoman shrine.
 */
defineProps({
  sourceArr: {
    type: Array,
    required: true,
  },
});

/** Guess a <source> MIME type from the file extension, e.g. "a.webm" -> "video/webm". */
function sourceType(link) {
  return "video/" + link.split(".").pop();
}
</script>

<template>
  <video
    v-for="(source, rowIndex) in sourceArr"
    :key="rowIndex"
    class="bdr-1"
    width="300"
    height="400"
    controls
    preload="none"
  >
    <source :src="source.link" :type="sourceType(source.link)" />
  </video>
</template>
