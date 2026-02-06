<script setup>
import Header from "@/components/text/Header.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";

import { useFavoritesStore } from "@/stores/favorites";

const favoritesStore = useFavoritesStore();
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>favs</Header>
        <AutoScroll class="w-full flex-1">
            <OptionalLinkTable
                class="w-full"
                :data="favoritesStore.favorites"
            />
        </AutoScroll>
    </div>
</template>
