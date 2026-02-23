<script setup>
import { ref } from "vue";

import Touchscreen from "@/components/util/Touchscreen.vue";
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
];
shuffleArray(srcs);
</script>

<template>
    <Touchscreen>
        <div class="flex flex-wrap tst">
            <a href="https://www.adam-french.co.uk">
                <img src="https://www.adam-french.co.uk/img/stamps/mine.gif" />
            </a>
            <a href="https://jacobbarron.xyz">
                <img
                    src="https://jacobbarron.xyz/Banneh.gif"
                    alt="jacobbarron.xyz"
                />
            </a>
            <img v-for="src in srcs" :src="src" />
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
