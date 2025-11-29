<template>
    <div v-if="streamLive">
        <img src="/img/tmpen31z3pe.PNG" />
        <audio controls :src="streamUrl" ref="audio"></audio>
    </div>
    <div v-else>
        <img src="/img/tmpen31z3pe.PNG" />
        <p>Stream is currently offline.</p>
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

onMounted(() => {
    checkStream();
    setInterval(checkStream, 30000);
});
</script>
