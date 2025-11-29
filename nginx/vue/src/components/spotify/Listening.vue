<template>
    <div v-if="song.is_playing" class="spotify-now-playing">
        <img :src="song.item.album.images[0].url" />
        <p><strong>Song:</strong> {{ song.item.name }}</p>
        <p><strong>Artist:</strong> {{ song.item.artists[0].name }}</p>
        <p>Is what im currently listening to rnrnrn ^_^</p>
    </div>
    <div v-else class="spotify-not-playing">
        <img src="/img/Untitled.png" />
        <p>I ain't listenin to nofin</p>
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
                console.log(data);
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
