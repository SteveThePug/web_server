<template>
    <div ref="container" @mouseover="handleHover" class="overflow-y-auto">
        <slot />
    </div>
</template>

<script setup>
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";

const container = useTemplateRef("container");

const SPEED = 0.0005; // % per frame
const PAUSE = 2000; // ms at top/bottom

let pos = 0;
let direction = 1; // 1 = down, -1 = up
let timeoutId;
let timeoutId2;
let cachedScrollHeight = 0;

function measureScrollHeight() {
    const el = container.value;
    if (el) cachedScrollHeight = el.scrollHeight;
}

function handleHover() {
    cancelAnimationFrame(timeoutId);
    clearTimeout(timeoutId2);
    timeoutId2 = setTimeout(
        () => (timeoutId = requestAnimationFrame(tick)),
        PAUSE,
    );
}

function tick() {
    const el = container.value;
    if (!el || cachedScrollHeight === 0) {
        timeoutId = requestAnimationFrame(tick);
        return;
    }

    const reachedBottom = pos <= 0;
    const reachedTop = pos >= 1;

    if (reachedBottom) {
        pos = 0.001;
        direction = 1;
        handleHover();
        return;
    } else if (reachedTop) {
        pos = 0.999;
        direction = -1;
        handleHover();
        return;
    }

    pos += direction * SPEED;

    el.scrollTop = pos * cachedScrollHeight;

    timeoutId = requestAnimationFrame(tick);
}

let resizeObserver;

onMounted(() => {
    measureScrollHeight();
    timeoutId = requestAnimationFrame(tick);

    resizeObserver = new ResizeObserver(measureScrollHeight);
    resizeObserver.observe(container.value);
});

onBeforeUnmount(() => {
    cancelAnimationFrame(timeoutId);
    resizeObserver?.disconnect();
});
</script>
