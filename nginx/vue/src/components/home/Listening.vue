<script setup>
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
    <Transition name="fade" mode="out-in">
        <div
            @click="nextSong"
            :key="song.track.id"
            class="flex-col center-content center-text"
        >
            <h2>Listening To</h2>
            <img :src="song.track.album.images[0].url" />
            <p><strong>Song:</strong> {{ song.track.name }}</p>
            <p><strong>Artist:</strong> {{ song.track.artists[0].name }}</p>
        </div>
    </Transition>
</template>

<style scoped>
img {
    width: 70%;
}
p {
    width: 100%;
    margin: 0 auto;
}

.fade-enter-active {
    transition: opacity 0.5s ease;
}
.fade-leave-active {
    transition: opacity 0.5s ease;
}
.fade-enter-from {
    opacity: 0;
}
.fade-leave-to {
    opacity: 0;
}
</style>
