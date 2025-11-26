<template>
    <div>
        <audio v-if="streamLive" controls :src="streamUrl" ref="audio"></audio>
        <p v-else>Stream is currently offline.</p>
        <button @click="checkStream()">Check Stream</button>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";

const streamMount = ref("");
const streamUrl = ref("");
const streamLive = ref(false);
const audio = ref(null);

async function checkStream() {
    try {
        const res = await fetch("/radio/status-json.xsl"); // Icecast JSON status
        const data = await res.json();
        // Replace 'mounts' and '/stream' with your Icecast mountpoint
        streamMount.value = data.icestats.source.listenurl.split("/").pop();

        if (streamMount.value) {
            streamLive.value = true;
            streamUrl.value = "/radio/" + streamMount.value;

            if (audio.value) audio.value.load(); // reload audio if it was offline before
        }
    } catch (err) {
        streamLive.value = false;
    }
}
</script>

<style scoped></style>
