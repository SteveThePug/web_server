<script setup>
import { ref, onMounted, onUnmounted } from "vue";

const container = ref(null);
const item1 = ref(null);
const item2 = ref(null);
const offset = ref(0);

let rafId;

const speed = 0.5; // pixels per frame

function animate() {
    const ctnr = container.value;
    const it1 = item1.value;
    const it2 = item2.value;

    const width = Math.max(ctnr.offsetWidth, it1.scrollWidth);

    offset.value -= speed;

    if (offset.value <= -width) {
        offset.value += width;
    }

    it1.style.transform = `translateX(${offset.value}px)`;
    it2.style.transform = `translateX(${width + offset.value}px)`;

    rafId = requestAnimationFrame(animate);
}

onMounted(() => {
    rafId = requestAnimationFrame(animate);
});

onUnmounted(() => {
    cancelAnimationFrame(rafId);
});
</script>

<template>
    <div class="marquee" ref="container">
        <p class="item" ref="item1"><slot /></p>
        <p class="item item2" ref="item2"><slot /></p>
    </div>
</template>

<style scoped>
.marquee {
    overflow: hidden;
    width: 100%;
    will-change: transform;
}

.item {
    top: 0px;
    padding-right: 3em;
    width: fit-content;
    white-space: nowrap;
}

.item2 {
    position: absolute;
}
</style>
