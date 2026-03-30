<template>
    <div v-if="streamLive" class="overflow-auto">
        <Header>Radio</Header>
        <img src="/img/tmpen31z3pe.PNG" alt="Radio" width="176" height="177" />
        <audio controls :src="streamUrl" ref="audio"></audio>
    </div>
    <div v-else class="overflow-auto">
        <Header>Radio</Header>
        <img src="/img/tmpen31z3pe.PNG" alt="Radio" width="176" height="177" />
        <div class="m-1 text-center">
            <p>Radio is offline. Message for info!</p>
            <Button class="w-full" @click="checkStream()">Check Stream</Button>
        </div>
    </div>
</template>

<script setup>
import Button from "@/components/input/Button.vue";
import Header from "@/components/text/Header.vue";
import { ref, useTemplateRef, onMounted, nextTick } from "vue";
import axios from "axios";

const streamUrl = ref("");
const streamLive = ref(false);
const audio = useTemplateRef("audio");

async function checkStream() {
    try {
        await axios.head("/radio/stream");
        if (!streamLive.value) {
            streamLive.value = true;
            streamUrl.value = "/radio/stream";
            await nextTick();
            if (audio.value) {
                audio.value.load();
                audio.value.volume = 0.2;
            }
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

<style scoped>
img {
    width: 100%;
    max-height: 150px;
    object-fit: cover;
}
</style>
