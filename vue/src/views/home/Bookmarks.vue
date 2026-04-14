<script setup>
import { computed } from "vue";
import LinkTable from "@/components/util/LinkTable.vue";
import Header from "@/components/text/Header.vue";
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
    <div class="bookmarks-wrapper">
        <h3 class="text-left">Bookmarks</h3>
        <div class="bookmarks-scroll">
            <LinkTable
                v-for="group in groupedBookmarks"
                :key="group[0]"
                :title="group[0]"
                :items="group[1]"
            />
        </div>
    </div>
</template>

<style scoped>
.bookmarks-wrapper {
    display: flex;
    flex-direction: column;
    min-height: 0;
    flex: 1;
}

.bookmarks-scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
}
</style>
