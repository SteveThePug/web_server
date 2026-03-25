<script setup>
import Header from "@/components/text/Header.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";

import { useFavoritesStore } from "@/stores/favorites";

const favoritesStore = useFavoritesStore();
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>favs</Header>
        <AutoScroll class="w-full flex-1">
            <LinkTable
                variant="table"
                class="w-full"
                :items="favoritesStore.favorites"
            />
        </AutoScroll>
    </div>
</template>
