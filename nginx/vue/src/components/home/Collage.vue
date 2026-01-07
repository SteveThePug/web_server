<script setup>
import { ref } from "vue";

const images = [
    "/img/memes/fire-woman.gif",
    "/img/memes/pidgeon.gif",
    "/img/memes/no_slip.png",
    "/img/memes/epic.jpeg",
];

const currentIndex = ref(0);

function nextImage() {
    currentIndex.value = Math.floor(Math.random() * 100) % images.length;
}

setInterval(nextImage, 10000);
</script>

<template>
    <div class="image-viewer" @click="nextImage">
        <img
            :src="images[currentIndex]"
            v-on:click="nextImage"
            alt="Image Viewer"
        />
    </div>
</template>

<style scoped>
img {
    width: 100%;
    height: 100%;
}
</style>
