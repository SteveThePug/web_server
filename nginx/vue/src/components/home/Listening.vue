<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import axios from "axios";

let intervalId = null;
let refreshId = null;

const song = computed(() => songs.value[idx.value]);
const songs = ref([
    {
        id: 1,
        track: {
            id: 1,
            name: "^_^",
            album: { images: [{ url: "/img/Untitled.png" }] },
            artists: [{ name: ">_<" }],
        },
    },
]);
const idx = ref(0);

async function fetchRecent() {
    try {
        const res = await axios.get("/api/spotify/recent");
        if (Array.isArray(res.data)) {
            songs.value = res.data;
            idx.value = 0;
        } else {
            throw new Error("Invalid response from Spotify API");
        }
    } catch (err) {
        console.error("Cannot connect to Spotify API", err);
    }
}

function nextSong() {
    if (!songs.value.length) return;
    idx.value = (idx.value + 1) % songs.value.length;
}

onMounted(async () => {
    await fetchRecent();
    intervalId = setInterval(nextSong, 5000);
    refreshId = setInterval(fetchRecent, 120000);
});

onUnmounted(() => {
    clearInterval(intervalId);
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
