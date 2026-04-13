<script setup>
import { computed } from "vue";
import LinkTable from "@/components/util/LinkTable.vue";
import { useHomeDataStore } from "@/stores/homeData";

const homeData = useHomeDataStore();

const groupedBookmarks = computed(() => {
    const groups = {};
    for (const b of homeData.bookmarks) {
        if (!groups[b.category]) groups[b.category] = [];
        groups[b.category].push(b);
    }
    return Object.entries(groups);
});
</script>

<template>
    <main class="items-center flex flex-col">
        <div
            class="a4page-portrait bdr-1 flex flex-row flex-wrap overflow-x-auto gap-1"
        >
            <div class="w-full h-fit">
                <LinkTable
                    class="flex flex-col flex-wrap"
                    v-for="group in groupedBookmarks"
                    :title="group[0]"
                    :items="group[1]"
                />
            </div>
        </div>
    </main>
</template>
