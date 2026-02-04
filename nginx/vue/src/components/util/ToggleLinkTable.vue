<script setup>
import { ref } from "vue";
import LinkTable from "@/components/util/LinkTable.vue";

const props = defineProps({
    linkArr: {
        type: Array,
        required: true,
    },
    title: {
        type: String,
        required: true,
    },
});

const show_links = ref(false);
</script>

<template>
    <div class="flex flex-row">
        <h2>{{ title }}</h2>
        <button @click="show_links = !show_links">Toggle</button>
    </div>
    <LinkTable v-if="show_links" :linkArr="props.linkArr" />
</template>
