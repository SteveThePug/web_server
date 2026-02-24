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

const items = ref<Item[]>(
    phrases.map((text, i) => ({
        x: i * 20,
        y: i * 20,
        dx: rand(0, 30) / 100,
        dy: 0.5,
        content: text,
    })),
);

let rafId = 0;

function animate() {
    const c = container.value;
    if (!c) return;

    const cw = c.clientWidth;
    const ch = c.clientHeight;

    items.value.forEach((item, i) => {
        const el = itemEls.value[i];
        if (!el) return;

        const ew = el.offsetWidth;
        const eh = el.offsetHeight;

        item.x += item.dx;
        item.y += item.dy;

        if (item.x < 0 || item.x > cw - ew) item.dx *= -1;
        if (item.y < 0 || item.y > ch - eh) item.dy *= -1;
    });

    rafId = requestAnimationFrame(animate);
}

onMounted(async () => {
    await nextTick();
    rafId = requestAnimationFrame(animate);
});

onUnmounted(() => {
    cancelAnimationFrame(rafId);
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
            :style="{
                transform: `translate(${item.x}px, ${item.y}px)`,
            }"
        >
            <h1>
                {{ item.content }}
            </h1>
        </div>
    </div>
</template>
