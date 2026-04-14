<script setup>
import { onMounted, useTemplateRef, onUnmounted } from "vue";
import { HeadlineScroller } from "@/wasm/stp_wasm.js";

const container = useTemplateRef("container");
const item1 = useTemplateRef("item1");

let scroller = null;

onMounted(() => {
    if (!container.value || !item1.value) return;
    scroller = new HeadlineScroller(container.value, item1.value);
    scroller.start();
});

onUnmounted(() => {
    scroller?.destroy();
    scroller?.free();
    scroller = null;
});
</script>

<template>
    <div class="root">
        <div class="container" ref="container">
            <div ref="item1">
                <slot />
            </div>
            <div>
                <slot />
            </div>
        </div>
    </div>
</template>

<style scoped>
.root {
    overflow: hidden;
}

.container {
    width: fit-content;
    height: fit-content;
    display: grid;
    grid-auto-flow: column;
    grid-auto-columns: max-content;
    /* Each column fits its content */
    overflow-x: visible;
    will-change: transform;
    gap: 10em;
}
</style>
