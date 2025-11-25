<template>
    <div class="spotify-now-playing">
        <img :src="albumImage" alt="Album cover" v-if="albumImage" />
        <p v-if="songName"><strong>Song:</strong> {{ songName }}</p>
        <p v-if="artistName"><strong>Artist:</strong> {{ artistName }}</p>
        <p v-if="playing">Status: Playing</p>
        <p v-else>Status: Not playing</p>
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
        const playing = ref(false);

        const fetchSpotify = async () => {
            try {
                const res = await fetch("/api/spotify");
                if (!res.ok) throw new Error("Failed to fetch Spotify data");
                const data = await res.json();
                albumImage.value = data.album_image;
                artistName.value = data.artist_name;
                songName.value = data.song_name;
                playing.value = data.playing;
            } catch (err) {
                console.error(err);
            }
        };

        onMounted(() => {
            fetchSpotify();
            // Optional: refresh every 30 seconds
            // setInterval(fetchSpotify, 30000);
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
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
}

.spotify-now-playing img {
    width: 200px;
    height: 200px;
    object-fit: cover;
    border-radius: 8px;
    margin-bottom: 10px;
}
</style>
