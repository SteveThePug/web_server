<template>
    <div class="spotify-now-playing">
        <img :src="playing ? albumImage : '/img/Untitled.png'" />
        <p><strong v-if="playing">Song:</strong> {{ songName }}</p>
        <p><strong v-if="playing">Artist:</strong> {{ artistName }}</p>
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
        const songUrl = ref("");
        const playing = ref(false);

        const fetchSpotify = async () => {
            try {
                const res = await fetch("/api/spotify");
                if (!res.ok) throw new Error("Failed to fetch Spotify data");
                const data = await res.json();
                albumImage.value = data.album_image;
                artistName.value = data.artist_name;
                songUrl.value = data.song_url;
                songName.value = data.song_name;
                playing.value = data.playing;
            } catch (err) {
                console.error(err);
            }
        };

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

.spotify-now-playing img {
    width: 200px;
    height: 200px;
    object-fit: cover;
    box-shadow: 3px;
    margin-bottom: 10px;
}
</style>
