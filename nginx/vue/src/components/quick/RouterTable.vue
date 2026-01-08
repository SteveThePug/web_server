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
    <RouterLink
        class="bdr-2"
        v-for="(row, rowIndex) in linkArr"
        :key="rowIndex"
        :to="row.link"
    >
        <h2>{{ row.name }}</h2>
    </RouterLink>
</template>
