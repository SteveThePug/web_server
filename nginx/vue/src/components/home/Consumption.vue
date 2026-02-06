<script setup>
import { onBeforeMount } from "vue";

import AutoScroll from "@/components/util/AutoScroll.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import Header from "@/components/text/Header.vue";

import { useActivityStore } from "@/stores/activity";

const activityStore = useActivityStore();

onBeforeMount(() => {
    activityStore.fetchActivity();
});
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>Consumption</Header>
        <AutoScroll class="flex-1 w-full">
            <OptionalLinkTable class="w-full" :data="activityStore.activity" />
        </AutoScroll>
    </div>
</template>
