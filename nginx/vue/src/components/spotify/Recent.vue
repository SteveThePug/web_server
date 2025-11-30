<template>
    <h2 class="center-content">Listening to RECENTLY</h2>
    <div v-if="fetched" class="flex-row">
        <div
            v-if="fetched"
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
    <div v-else class="flex-row">
        <div class="bg-white border2 shadow1 tile1">
            <img src="/img/Untitled.png" />
            <p>I ain't listenin to nofin rn :/</p>
        </div>
        <div class="bg-white border2 shadow1 tile1">
            <img src="/img/Untitled.png" />
            <p>I ain't listenin to nofin rn :/</p>
        </div>
        <div class="bg-white border2 shadow1 tile1">
            <img src="/img/Untitled.png" />
            <p>I ain't listenin to nofin rn :/</p>
        </div>
    </div>
</template>

<script>
import { ref, onMounted } from "vue";
import axios from "axios";

export default {
    name: "spotify-recent",
    setup() {
        const played = ref([]);
        const fetched = ref(false);

        async function fetchRecent() {
            try {
                const res = await axios.get("/api/spotify/recent");
                played.value = res.data;
                fetched.value = true;
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
            fetched,
        };
    },
};
</script>
