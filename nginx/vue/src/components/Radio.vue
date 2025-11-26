<template>
    <div>
        <audio v-if="streamLive" controls :src="streamUrl" ref="audio"></audio>
        <p v-else>Stream is currently offline.</p>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";

const streamUrl = "/radio/stream";
const streamLive = ref(false);
const audio = ref(null);

const checkStream = async () => {
    try {
        const res = await fetch("/radio/status-json.xsl"); // Icecast JSON status
        const data = await res.json();
        // Replace 'mounts' and '/stream' with your Icecast mountpoint
        streamLive.value = !!data.icestats.source.find((src) =>
            src.listenurl.includes(streamUrl),
        );
        if (streamLive.value && audio.value) {
            audio.value.load(); // reload audio if it was offline before
        }
    } catch (err) {
        streamLive.value = false;
    }
};

// Check on mount
onMounted(() => {
    checkStream();
    // Poll every 10 seconds
    setInterval(checkStream, 10000);
});
</script>
