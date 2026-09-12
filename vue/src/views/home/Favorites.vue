<script setup>
import Header from "@/components/text/Header.vue";
import CreateToggle from "@/components/input/CreateToggle.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";
import Modal from "@/components/util/Modal.vue";

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
      favs
      <template #action><CreateToggle v-model="showCreate" /></template>
    </Header>
    <AutoScroll class="w-full flex-1">
      <LinkTable
        variant="table"
        class="w-full"
        :items="favoritesStore.favorites"
      />
    </AutoScroll>
    <Modal v-model="showCreate">
      <CreateFavorite
        @done="showCreate = false"
        @cancel="showCreate = false"
      />
    </Modal>
  </div>
</template>
