<script setup>
import Header from "@/components/text/Header.vue";
import LinkTable from "@/components/util/LinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";

import { ref, defineAsyncComponent } from "vue";
import { useFavoritesStore } from "@/stores/favorites";
import { useAuthStore } from "@/stores/auth";

const CreateFavorite = defineAsyncComponent(
  () => import("@/views/admin/CreateFavorite.vue"),
);

const favoritesStore = useFavoritesStore();
const authStore = useAuthStore();
const showCreate = ref(false);
</script>

<template>
  <div class="flex flex-col items-center">
    <Header>
      <span class="flex items-center justify-between w-full">
        {{ showCreate ? "Create Favorite" : "favs" }}
        <button
          v-if="authStore.user.admin"
          class="text-sm px-1"
          @click="showCreate = !showCreate"
        >
          {{ showCreate ? "x" : "+" }}
        </button>
      </span>
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
