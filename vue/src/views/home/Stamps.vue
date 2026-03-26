<script setup>
import { ref, onMounted, onUnmounted } from "vue";

import Touchscreen from "@/components/util/Touchscreen.vue";
import Link from "@/components/text/Link.vue";
import { shuffleArray } from "@/js/utils.js";

let srcs = [
    "/img/stamps/portal.gif",
    "/img/stamps/miku.gif",
    "/img/stamps/utau.gif",
    "/img/stamps/teto.webp",
    "/img/stamps/3ds.jpg",
    "/img/stamps/fry.png",
    "/img/stamps/ai.png",
    "/img/stamps/rei.png",
    "/img/stamps/tetris.gif",
    "/img/stamps/tf2.gif",
    "/img/stamps/demo.gif",
    "/img/stamps/demo.gif",
    "/img/stamps/demo.gif",
    "/img/stamps/demo.gif",
];
shuffleArray(srcs);

const touchscreen = ref(null);
let animId = null;
let posX = 0;
let posY = 0;
let dx = 0.2;
let dy = 0.12;

function bounce() {
    const el = touchscreen.value?.$el;
    if (!el) return;

    const maxX = el.scrollWidth - el.clientWidth;
    const maxY = el.scrollHeight - el.clientHeight;

    if (maxX > 0) {
        posX += dx;
        if (posX <= 0) {
            posX = 0;
            dx = -dx;
        } else if (posX >= maxX) {
            posX = maxX;
            dx = -dx;
        }
        el.scrollLeft = posX;
    }
    if (maxY > 0) {
        posY += dy;
        if (posY <= 0) {
            posY = 0;
            dy = -dy;
        } else if (posY >= maxY) {
            posY = maxY;
            dy = -dy;
        }
        el.scrollTop = posY;
    }

    animId = requestAnimationFrame(bounce);
}

onMounted(() => {
    animId = requestAnimationFrame(bounce);
});

onUnmounted(() => {
    if (animId) cancelAnimationFrame(animId);
});
</script>

<template>
    <Touchscreen ref="touchscreen">
        <div class="flex flex-wrap tst">
            <Link bare href="https://www.adam-french.co.uk">
                <img src="https://www.adam-french.co.uk/img/stamps/mine.gif" alt="adam-french.co.uk" />
            </Link>
            <Link bare href="https://jacobbarron.xyz">
                <img
                    src="https://jacobbarron.xyz/Banneh.gif"
                    alt="jacobbarron.xyz"
                />
            </Link>
            <img v-for="src in srcs" :src="src" :alt="src.split('/').pop().split('.')[0]" loading="lazy" />
        </div>
    </Touchscreen>
</template>

<style scoped>
img {
    width: 89px;
    height: 59px;
}

.tst {
    min-width: calc(89px * 4);
}
</style>
