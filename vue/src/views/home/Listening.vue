<script setup>
import Header from "@/components/text/Header.vue";
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useSongsStore } from "@/stores/songs";

const songsStore = useSongsStore();
const idx = ref(0);
const song = computed(() => songsStore.songs[idx.value]);

let nextId = null;
let refreshId = null;

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
    <Transition name="fade">
        <div
            @click="nextSong"
            :key="song.track.name"
            class="flex flex-col items-center pt-2"
        >
            <Header>Listening To</Header>
            <img :src="song.track.album.images[0].url" :alt="song.track.album.name + ' album art'" />
            <p class="text-center">
                <strong>Song:</strong> {{ song.track.name }}
            </p>
            <p class="text-center">
                <strong>Artist:</strong> {{ song.track.artists[0].name }}
            </p>
        </div>
    </Transition>
    </div>
</template>

<style scoped>
.listening-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
    min-width: 0;
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

.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.5s ease;
}
.fade-leave-active {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
