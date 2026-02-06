<template>
    <div ref="container" @mouseover="handleHover" class="overflow-y-auto">
        <slot />
    </div>
</template>

<script setup>
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";

const container = useTemplateRef("container");

const SPEED = 1; // px per frame
const PAUSE = 2000; // ms at top/bottom

let direction = 1; // 1 = down, -1 = up
let timeoutId;
let timeoutId2;

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
    el.scrollTop += SPEED * direction;

    const reachedBottom = el.scrollTop + el.clientHeight >= el.scrollHeight;
    const reachedTop = el.scrollTop <= 0;

    if (reachedBottom || reachedTop) {
        direction *= -1;
        handleHover();
        return;
    }

    timeoutId = requestAnimationFrame(tick);
}

onMounted(() => {
    timeoutId = requestAnimationFrame(tick);
});

onBeforeUnmount(() => {
    cancelAnimationFrame(timeoutId);
});
</script>
