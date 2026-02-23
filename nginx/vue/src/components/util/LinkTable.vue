<script setup>
import { computed } from "vue";

const props = defineProps({
    linkArr: {
        type: Array,
        required: true,
    },
});

const keys = ["name", "link"];
</script>

<template>
    <a v-for="(row, rowIndex) in linkArr" :key="rowIndex" :href="row.link">
        <p class="bdr-2 bg-bg_tertiary">
            {{ row.name }}
        </p>
    </a>
</template>
