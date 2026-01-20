<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { Transition } from "vue";

const images = [
    "/img/memes/pidgeon.gif",
    "/img/memes/no_slip.png",
    "/img/memes/epic.jpeg",
    "/img/bedroom/img2.png",
    "/img/bedroom/img1.png",
];

const currentIndex = ref(0);

function nextImage() {
    let newIndex;
    do {
        newIndex = Math.floor(Math.random() * images.length);
    } while (newIndex === currentIndex.value); // prevent same image repeating
    currentIndex.value = newIndex;
}

let intervalId;

onMounted(() => {
    intervalId = setInterval(nextImage, 10000);
});

onUnmounted(() => {
    clearInterval(intervalId);
});
</script>

<template>
    <div class="image-viewer" @click="nextImage">
        <Transition name="fade" mode="out-in">
            <img
                :src="images[currentIndex]"
                :key="currentIndex"
                alt="Image Viewer"
            />
        </Transition>
    </div>
</template>

<style scoped>
.image-viewer {
    width: 100%;
    height: 100%;
    overflow: hidden;
}

img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.5s ease;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
