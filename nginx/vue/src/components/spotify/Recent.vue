<template>
    <div
        v-for="(song, idx) in played"
        :key="song.item.id || idx"
        class="spotify-now-playing"
    >
        <img :src="song.item.album.images[0].url" />
        <p><strong>Song:</strong> {{ song.item.name }}</p>
        <p><strong>Artist:</strong> {{ song.item.artists[0].name }}</p>
        <p>Is what im currently listening to rnrnrn ^_^</p>
    </div>
</template>

<script>
import { ref, onMounted } from "vue";

export default {
    name: "spotify-recent",
    setup() {
        const played = ref([]);

        async function fetchRecent() {
            try {
                const res = await fetch("/api/spotify/recent");
                if (!res.ok) throw new Error("Failed to fetch Spotify data");
                played.value = await res.json();
            } catch (err) {
                console.error(err);
            }
        }

        onMounted(() => {
            fetchRecent();
            setInterval(fetchRecent, 120000);
        });

        return {
            played,
        };
    },
};
</script>

<style scoped>
.spotify-now-playing {
    width: fit-content;
    height: fit-content;
    flex-direction: column;
    align-items: center;
    text-align: center;
    box-shadow: 3px;
}

.spotify-not-playing {
    border: 2px solid black;
    align-items: center;
    text-align: center;
    background: white;
}
</style>
