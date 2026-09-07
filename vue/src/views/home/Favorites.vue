<script setup>
import Header from "@/components/text/Header.vue";
import CreateToggle from "@/components/input/CreateToggle.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";

import { ref, defineAsyncComponent } from "vue";
import { useFavoritesStore } from "@/stores/favorites";

const CreateFavorite = defineAsyncComponent(
  () => import("@/views/admin/CreateFavorite.vue"),
);

const favoritesStore = useFavoritesStore();
const showCreate = ref(false);
</script>

<template>
  <div class="flex flex-col items-center">
    <Header>
      {{ showCreate ? "Create Favorite" : "favs" }}
      <template #action><CreateToggle v-model="showCreate" /></template>
    </Header>
    <CreateFavorite
      v-if="showCreate"
      class="w-full flex-1 p-1"
      @done="showCreate = false"
      @cancel="showCreate = false"
    />
    <AutoScroll v-if="!showCreate" class="w-full flex-1">
      <LinkTable
        variant="table"
        class="w-full"
        :items="favoritesStore.favorites"
      />
    </AutoScroll>
  </div>
</template>
