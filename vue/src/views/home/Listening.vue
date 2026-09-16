<script setup>
/**
 * "Listening" widget on /stp: rotates through recent Spotify plays.
 *
 * Two timers: a 5s rotation (self-rearming setTimeout, so a manual change would
 * reset the full interval) and a 2-minute re-poll of Spotify. Both are cleared on
 * unmount.
 */
import Header from "@/components/text/Header.vue";
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useSongsStore } from "@/stores/songs";

const songsStore = useSongsStore();
const idx = ref(0);
const song = computed(() => songsStore.songs[idx.value]);

let nextId = null;
let refreshId = null;

// Self-rearming timeout rather than setInterval, so the full 5s is restarted
// whenever the song changes rather than leaving a short remainder.
function nextSong() {
  clearTimeout(nextId);
  nextId = setTimeout(nextSong, 5000);
  idx.value = (idx.value + 1) % songsStore.songsCount;
}

onMounted(() => {
  songsStore.fetchSongs();
  nextId = setTimeout(nextSong, 5000);
  refreshId = setInterval(songsStore.fetchSongs, 120000);
});

onUnmounted(() => {
  clearTimeout(nextId);
  clearInterval(refreshId);
});
</script>

<template>
  <div class="listening-wrapper">
    <div class="header-wrapper">
      <Header>Listening To</Header>
    </div>
    <div class="content-scroll">
      <Transition name="fade">
        <div
          @click="nextSong"
          :key="song.track.name"
          class="flex flex-col items-center"
        >
          <img
            :src="song.track.album.images[0].url"
            :alt="song.track.album.name + ' album art'"
          />
          <p class="text-center"><strong>Song:</strong> {{ song.track.name }}</p>
          <p class="text-center">
            <strong>Artist:</strong> {{ song.track.artists[0].name }}
          </p>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.listening-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.header-wrapper {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  padding-top: 0.5rem;
}

.content-scroll {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

img {
  width: 70%;
  max-width: 100%;
  height: auto;
}
p {
  width: 100%;
  margin: 0 auto;
}
</style>
