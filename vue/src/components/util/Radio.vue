<template>
  <div v-if="streamLive" class="overflow-auto">
    <Header>Radio</Header>
    <img src="/img/tmpen31z3pe.PNG" alt="Radio" width="176" height="177" />
    <audio controls :src="streamUrl" ref="audio"></audio>
  </div>
  <div v-else class="overflow-auto">
    <Header>Radio</Header>
    <img src="/img/tmpen31z3pe.PNG" alt="Radio" width="176" height="177" />
    <div class="m-1 text-center">
      <p>Radio is offline. Message for info!</p>
      <Button class="w-full" @click="checkStream()">Check Stream</Button>
    </div>
  </div>
</template>

<script setup>
/**
 * Icecast radio widget. Probes /radio/stream with a HEAD request on mount and
 * every 2 minutes, and only mounts the <audio> element once the stream answers.
 *
 * The nextTick() before load() matters: the <audio> only exists after `streamLive`
 * flips and Vue has re-rendered, so calling load() any earlier would hit null.
 * Volume is knocked down to 20% because the stream is loud.
 */
import Button from "@/components/input/Button.vue";
import Header from "@/components/text/Header.vue";
import { ref, useTemplateRef, onMounted, onUnmounted, nextTick } from "vue";
import axios from "axios";

const POLL_INTERVAL_MS = 120000; // re-probe the stream every 2 minutes
const INITIAL_VOLUME = 0.2; // the stream is loud

const streamUrl = ref("");
const streamLive = ref(false);
const audio = useTemplateRef("audio");

/** HEAD the stream; on the first success mount and prime the <audio>.
 *  A rejected request just means the station is off air. */
async function checkStream() {
  try {
    await axios.head("/radio/stream");
    if (!streamLive.value) {
      streamLive.value = true;
      streamUrl.value = "/radio/stream";
      await nextTick();
      if (audio.value) {
        audio.value.load();
        audio.value.volume = INITIAL_VOLUME;
      }
    }
  } catch (err) {
    streamLive.value = false;
  }
}

let pollId;

onMounted(() => {
  checkStream();
  pollId = setInterval(checkStream, POLL_INTERVAL_MS);
});

onUnmounted(() => {
  clearInterval(pollId);
});
</script>

<style scoped>
img {
  width: 100%;
  max-height: 150px;
  object-fit: cover;
}
</style>
