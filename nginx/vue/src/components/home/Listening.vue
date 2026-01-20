<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import axios from "axios";

const song = ref(null);
const intervalId = ref(null);
const refreshId = ref(null);

let songs = [];
let idx = 0;

async function fetchRecent() {
    try {
        const res = await axios.get("/api/spotify/recent");
        songs = res.data || [];

        idx = 0;
        song.value = songs[0] ?? null;
    } catch (err) {
        console.error("Cannot connect to Spotify API", err);
    }
}

function nextSong() {
    if (!songs.length) return;

    song.value = songs[idx];
    idx = (idx + 1) % songs.length;
}

onMounted(async () => {
    await fetchRecent();

    intervalId.value = setInterval(nextSong, 5000);
    refreshId.value = setInterval(fetchRecent, 120000);
});

onUnmounted(() => {
    clearInterval(intervalId.value);
    clearInterval(refreshId.value);
});
</script>

<template>
    <Transition name="fade" mode="out-in">
        <div
            v-if="song"
            @click="nextSong"
            class="flex-col center-content center-text"
        >
            <h2>Listening To</h2>
            <img :src="song.track.album.images[0].url" />
            <p><strong>Song:</strong> {{ song.track.name }}</p>
            <p><strong>Artist:</strong> {{ song.track.artists[0].name }}</p>
        </div>
        <div v-else class="flex-col center-content center-text">
            <h2>Listening To</h2>
            <img src="/img/Untitled.png" />
            <p><strong>Song:</strong> >_<</p>
            <p><strong>Artist:</strong> ^_^</p>
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
