<template>
    <h2 class="center-content">What've been listening to recently</h2>
    <div class="flex-row">
        <div
            v-for="(song, idx) in played"
            :key="song.track.id || idx"
            class="bg-white border2 shadow1"
            style="width: 280px"
        >
            <img :src="song.track.album.images[0].url" />
            <p><strong>Song:</strong> {{ song.track.name }}</p>
            <p><strong>Artist:</strong> {{ song.track.artists[0].name }}</p>
        </div>
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
