<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";

const song = ref(null);
const fetched = ref(false);
let songs = [];
const len = 3;
let idx = 0;

async function fetchRecent() {
    try {
        const res = await axios.get("/api/spotify/recent");
        songs = res.data;
        fetched.value = true;
        song.value = songs[0];
    } catch (err) {
        console.error(err);
    }
}

function nextSong() {
    song.value = songs[idx];
    idx = (idx + 1) % len;
}

onMounted(() => {
    fetchRecent();
    setInterval(fetchRecent, 120000);
    setInterval(nextSong, 5000);
});
</script>

<template>
    <Transition name="fade" mode="out-in">
        <div
            v-if="fetched"
            :key="song"
            v-on:click="nextSong"
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
