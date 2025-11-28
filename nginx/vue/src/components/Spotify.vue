<template>
    <div v-if="playing" class="spotify-now-playing">
        <img :src="albumImage" />
        <p><strong>Song:</strong> {{ songName }}</p>
        <p><strong>Artist:</strong> {{ artistName }}</p>
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
    name: "SpotifyNowPlaying",
    setup() {
        const albumImage = ref("");
        const artistName = ref("");
        const songName = ref("");
        const songUrl = ref("");
        const playing = ref(false);

        async function fetchSpotify() {
            try {
                const res = await fetch("/api/spotify");
                if (!res.ok) throw new Error("Failed to fetch Spotify data");
                const data = await res.json();
                if (playing.value == false) {
                    return;
                } else {
                    albumImage.value = data.album_image;
                    artistName.value = data.artist_name;
                    songUrl.value = data.song_url;
                    songName.value = data.song_name;
                    playing.value = data.playing;
                    return;
                }
            } catch (err) {
                console.error(err);
            }
        }

        onMounted(() => {
            fetchSpotify();
            setInterval(fetchSpotify, 120000);
        });

        return {
            albumImage,
            artistName,
            songName,
            playing,
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
