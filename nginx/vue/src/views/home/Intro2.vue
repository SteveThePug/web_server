<script setup lang="ts">
import { rand } from "@vueuse/core";
import { ref, onMounted, onUnmounted, nextTick } from "vue";

interface Item {
    x: number;
    y: number;
    dx: number;
    dy: number;
    content: string;
}

const container = ref<HTMLDivElement | null>(null);
const itemEls = ref<HTMLDivElement[]>([]);

const phrases = [
    "Welcome to my website",
    "I'm looking to do a big revamp",
    "A4 sheets of paper are what I'm used to",
    "I'd love to know your recommendations",
    "for very short books",
    "Message me by any means necessary",
    "I like anime, all kinds of music and sci fic",
];

// Non-reactive animation state to avoid triggering Vue re-renders every frame
const animState = phrases.map((text, i) => ({
    x: i * 20,
    y: i * 20,
    dx: rand(0, 60) / 100,
    dy: 1.0,
    content: text,
    cachedW: 0,
    cachedH: 0,
}));

// Reactive items only for initial render
const items = ref<Item[]>(
    animState.map((s) => ({
        x: s.x,
        y: s.y,
        dx: s.dx,
        dy: s.dy,
        content: s.content,
    })),
);

let rafId = 0;
let cachedCW = 0;
let cachedCH = 0;
let lastFrameTime = 0;
const FRAME_INTERVAL = 1000 / 30;

function measureSizes() {
    const c = container.value;
    if (c) {
        cachedCW = c.clientWidth;
        cachedCH = c.clientHeight;
    }
    itemEls.value.forEach((el, i) => {
        if (el && animState[i]) {
            animState[i].cachedW = el.offsetWidth;
            animState[i].cachedH = el.offsetHeight;
        }
    });
}

function animate(timestamp: number) {
    if (!cachedCW || !cachedCH) {
        rafId = requestAnimationFrame(animate);
        return;
    }

    if (timestamp - lastFrameTime < FRAME_INTERVAL) {
        rafId = requestAnimationFrame(animate);
        return;
    }
    lastFrameTime = timestamp;

    for (let i = 0; i < animState.length; i++) {
        const s = animState[i];
        const el = itemEls.value[i];
        if (!el) continue;

        s.x += s.dx;
        s.y += s.dy;

        if (s.x < 0 || s.x > cachedCW - s.cachedW) s.dx *= -1;
        if (s.y < 0 || s.y > cachedCH - s.cachedH) s.dy *= -1;

        el.style.transform = `translate(${s.x}px, ${s.y}px)`;
    }

    rafId = requestAnimationFrame(animate);
}

let resizeObserver: ResizeObserver;

onMounted(async () => {
    await nextTick();
    measureSizes();
    rafId = requestAnimationFrame(animate);

    resizeObserver = new ResizeObserver(measureSizes);
    resizeObserver.observe(container.value!);
});

onUnmounted(() => {
    cancelAnimationFrame(rafId);
    resizeObserver?.disconnect();
});
</script>

<template>
    <div
        ref="container"
        class="w-full h-full min-h-125 relative overflow-hidden"
    >
        <div
            v-for="(item, i) in items"
            :key="i"
            ref="itemEls"
            class="absolute w-fit h-fit"
        >
            <h1>
                {{ item.content }}
            </h1>
        </div>
    </div>
</template>
