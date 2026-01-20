<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Transition } from "vue";

const images = [
    { url: "/img/memes/pidgeon.gif", comment: "鸟" },
    //{ url: "/img/memes/no_slip.png" },
    //{ url: "/img/memes/epic.jpeg" },
    { url: "/img/bedroom/img2.png", comment: "办公桌" },
    { url: "/img/bedroom/img1.png", comment: "床" },
];

const currentIndex = ref(0);
const currentComment = computed(() => images[currentIndex.value].comment);
const currentUrl = computed(() => images[currentIndex.value].url);

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
    <Transition name="fade" mode="out-in">
        <div
            class="image-viewer center-text"
            @click="nextImage"
            :key="currentIndex"
        >
            <p v-if="currentComment">
                {{ currentComment }}
            </p>
            <img :src="currentUrl" alt="Image Viewer" />
        </div>
    </Transition>
</template>

<style scoped>
.image-viewer {
    width: 100%;
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
