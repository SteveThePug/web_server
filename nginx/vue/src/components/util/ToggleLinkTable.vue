<script setup>
import ToggleHeader from "@/components/text/ToggleHeader.vue";

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
    <div class="h-fit w-fit">
        <ToggleHeader v-model="show_links" class="justify-between flex"
            >{{ title }}
        </ToggleHeader>
        <LinkTable v-if="show_links" :linkArr="props.linkArr" />
    </div>
</template>
