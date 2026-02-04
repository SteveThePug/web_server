<template>
    <div ref="container" class="overflow-y-auto">
        <slot />
    </div>
</template>

<script setup>
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";

const container = useTemplateRef("container");

const SPEED = 1; // px per frame
const PAUSE = 2000; // ms at top/bottom
const FRAME_TIME = 50; // ms at top/bottom

let direction = 1; // 1 = down, -1 = up
let paused = false;
let rafId;
let timeoutId;

function tick() {
    const el = container.value;

    el.scrollTop += SPEED * direction;

    const reachedBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1;
    const reachedTop = el.scrollTop <= 0;

    if (reachedBottom || reachedTop) {
        direction *= -1;

        timeoutId = setTimeout(() => {
            paused = false;
            rafId = setTimeout(tick, FRAME_TIME);
        }, PAUSE);

        return;
    }

    rafId = setTimeout(tick, FRAME_TIME);
}

onMounted(() => {
    rafId = setTimeout(tick, FRAME_TIME);
});

onBeforeUnmount(() => {
    clearTimeout(rafId);
    clearTimeout(timeoutId);
});
</script>
