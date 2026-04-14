<template>
    <div ref="container" class="overflow-y-auto">
        <slot />
    </div>
</template>

<script setup>
import { useTemplateRef, onMounted, onBeforeUnmount } from "vue";
import { AutoScroller } from "@/wasm/stp_wasm.js";

const container = useTemplateRef("container");

let scroller = null;

onMounted(() => {
    if (!container.value) return;
    scroller = new AutoScroller(container.value);
    scroller.start();
});

onBeforeUnmount(() => {
    scroller?.destroy();
    scroller?.free();
    scroller = null;
});
</script>
