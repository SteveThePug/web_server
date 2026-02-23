<template>
    <div v-if="streamLive">
        <img src="/img/tmpen31z3pe.PNG" />
        <audio controls :src="streamUrl" ref="audio"></audio>
    </div>
    <div v-else>
        <img src="/img/tmpen31z3pe.PNG" />
        <div class="m-1">
            <p>Radio is offline. Message for info!</p>
            <Button class="w-full" @click="checkStream()">Check Stream</Button>
        </div>
    </div>
</template>

<script setup>
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const streamMount = ref("");
const streamUrl = ref("");
const streamLive = ref(false);
const audio = ref(null);

async function checkStream() {
    try {
        const res = await axios.get("/radio/status-json.xsl");
        const data = res.data;

        streamMount.value = data.icestats.source.listenurl.split("/").pop();
        if (streamMount.value) {
            streamLive.value = true;
            streamUrl.value = "/radio/" + streamMount.value;

            if (audio.value) audio.value.load();
        }
    } catch (err) {
        streamLive.value = false;
    }
}

onMounted(() => {
    checkStream();
    setInterval(checkStream, 120000);
});
</script>
