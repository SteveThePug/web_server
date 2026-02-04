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
    <a
        class="bdr-2 bg-bg_primary"
        v-for="(row, rowIndex) in linkArr"
        :key="rowIndex"
        :href="row.link"
    >
        <p>
            {{ row.name }}
        </p>
    </a>
</template>
