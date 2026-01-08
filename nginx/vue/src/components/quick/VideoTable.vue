<script setup>
import { computed } from "vue";

const props = defineProps({
    sourceArr: {
        type: Array,
        required: true,
    },
});

function sourceType(link) {
    return "video/" + link.split(".").pop();
}

const keys = ["name", "link"];
</script>

<template>
    <video
        v-for="(source, rowIndex) in sourceArr"
        :key="rowIndex"
        class="bdr-1"
        width="300"
        height="400"
        controls
        preload="none"
    >
        <source :src="source.link" :type="sourceType(source.link)" />
    </video>
</template>
