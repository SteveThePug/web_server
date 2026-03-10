<template>
    <div v-if="streamLive">
        <Header>Radio</Header>
        <img src="/img/tmpen31z3pe.PNG" />
        <audio controls :src="streamUrl" ref="audio"></audio>
    </div>
    <div v-else>
        <Header>Radio</Header>
        <img src="/img/tmpen31z3pe.PNG" />
        <div class="m-1 text-center">
            <p>Radio is offline. Message for info!</p>
            <Button class="w-full" @click="checkStream()">Check Stream</Button>
        </div>
    </div>
</template>

<script setup>
import Button from "@/components/input/Button.vue";
import Header from "@/components/text/Header.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const streamUrl = ref("");
const streamLive = ref(false);
const audio = ref(null);

async function checkStream() {
    try {
        await axios.head("/radio/stream");
        streamLive.value = true;
        streamUrl.value = "/radio/stream";

        if (audio.value) audio.value.load();
    } catch (err) {
        streamLive.value = false;
    }
}

onMounted(() => {
    checkStream();
    setInterval(checkStream, 120000);
});
</script>
