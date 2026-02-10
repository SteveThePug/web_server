<script setup>
import { ref, onMounted, useTemplateRef, onUnmounted } from "vue";

const container = useTemplateRef("container");
const item1 = useTemplateRef("item1");
const item2 = useTemplateRef("item2");

let offset = 0;

let rafId;

const speed = 0.5; // pixels per frame

function animate() {
    const ctnr = container.value;
    const it1 = item1.value;
    const it2 = item2.value;

    const width = Math.max(ctnr.offsetWidth, it1.scrollWidth);

    offset -= speed;

    if (offset <= -width) {
        offset += width;
    }

    it1.style.transform = `translateX(${offset}px)`;
    it2.style.transform = `translateX(${width + offset}px)`;

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
    <div class="marquee">
        <div class="container" ref="container">
            <div class="item" ref="item1"><slot /></div>
            <div class="item item2" ref="item2"><slot /></div>
        </div>
    </div>
</template>

<style scoped>
.marquee {
    overflow: hidden;
    width: 100%;
}

.container {
    width: 100%;
    height: fit-content;
    position: relative;
    will-change: transform;
}

.item {
    height: fit-content;
    top: 0px;
    padding-right: 3em;
    width: fit-content;
    white-space: nowrap;
}

.item1 {
    left: 0px;
}

.item2 {
    position: absolute;
}
</style>
