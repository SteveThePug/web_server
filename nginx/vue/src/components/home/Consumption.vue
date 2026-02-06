<script setup>
import AutoScroll from "@/components/util/AutoScroll.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import Header from "@/components/text/Header.vue";

import { useActivityStore } from "@/stores/activity";

const activityStore = useActivityStore();
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>Consumption</Header>
        <AutoScroll>
            <OptionalLinkTable :data="activityStore.activity" />
        </AutoScroll>
    </div>
</template>
