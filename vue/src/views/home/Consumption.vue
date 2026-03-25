<script setup>
import AutoScroll from "@/components/util/AutoScroll.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Header from "@/components/text/Header.vue";

import { useActivityStore } from "@/stores/activity";

const activityStore = useActivityStore();
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>Consumption</Header>
        <AutoScroll class="flex-1 w-full">
            <LinkTable variant="table" class="w-full" :items="activityStore.activity" />
        </AutoScroll>
    </div>
</template>
