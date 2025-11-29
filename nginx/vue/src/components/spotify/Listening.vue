<template>
    <h2 class="center-content">Listening to RN! ^_^</h2>
    <div v-if="song.is_playing" class="center-content">
        <img :src="song.item.album.images[0].url" />
        <p><strong>Song:</strong> {{ song.item.name }}</p>
        <p><strong>Artist:</strong> {{ song.item.artists[0].name }}</p>
    </div>
    <div v-else class="center-content bg-white border2 shadow1">
        <img src="/img/Untitled.png" />
        <p>I ain't listenin to nofin rn :/</p>
    </div>
</template>

<script>
import { ref, onMounted } from "vue";

export default {
    name: "spotify-listening",
    setup() {
        const song = ref({});

        async function fetchSpotify() {
            try {
                const res = await fetch("/api/spotify/listening");
                if (!res.ok) throw new Error("Failed to fetch Spotify data");
                song.value = await res.json();
            } catch (err) {
                console.error(err);
            }
        }

        onMounted(() => {
            fetchSpotify();
            setInterval(fetchSpotify, 120000);
        });

        return {
            song,
        };
    },
};
</script>
