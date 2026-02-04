<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Transition } from "vue";
import Header from "@/components/text/Header.vue";

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

let nextId;

function nextImage() {
    clearTimeout(nextId);
    let newIndex;
    do {
        newIndex = Math.floor(Math.random() * images.length);
    } while (newIndex === currentIndex.value);
    currentIndex.value = newIndex;
    nextId = setTimeout(nextImage, 10000);
}

onMounted(() => {
    nextId = setTimeout(nextImage, 10000);
});

onUnmounted(() => {
    clearTimeout(nextId);
});
</script>

<template>
    <Transition name="fade" mode="out-in">
        <div
            class="image-viewer text-center"
            @click="nextImage"
            :key="currentIndex"
        >
            <Header v-if="currentComment">
                {{ currentComment }}
            </Header>
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
