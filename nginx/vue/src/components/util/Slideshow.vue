<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import Header from "@/components/text/Header.vue";

const props = defineProps({
    images: {
        type: Array,
        required: true,
    },
    interval: {
        type: Number,
        default: 10000,
    },
});

const currentIndex = ref(0);
const currentComment = computed(() => props.images[currentIndex.value].comment);
const currentUrl = computed(() => props.images[currentIndex.value].url);

let nextId;

function nextImage() {
    clearTimeout(nextId);
    currentIndex.value = (currentIndex.value + 1) % props.images.length;
    nextId = setTimeout(nextImage, props.interval);
}

onMounted(() => {
    nextId = setTimeout(nextImage, props.interval);
});

onUnmounted(() => {
    clearTimeout(nextId);
});
</script>

<template>
    <div class="slideshow-wrapper">
        <Transition name="fade">
            <div class="image-viewer" @click="nextImage" :key="currentIndex">
                <Header v-if="currentComment">
                    {{ currentComment }}
                </Header>
                <img :src="currentUrl" alt="Image Viewer" loading="lazy" />
            </div>
        </Transition>
    </div>
</template>

<style scoped>
.slideshow-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
}

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
.fade-leave-active {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
