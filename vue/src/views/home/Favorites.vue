<script setup>
/**
 * "Favorites" widget on /stp. All of the markup lives in ListWidget.vue; this
 * file only binds the favourites store and the (lazily imported) create form.
 */
import ListWidget from "@/views/home/ListWidget.vue";

import { defineAsyncComponent } from "vue";
import { useFavoritesStore } from "@/stores/favorites";

const CreateFavorite = defineAsyncComponent(
  () => import("@/views/admin/CreateFavorite.vue"),
);

const favoritesStore = useFavoritesStore();
</script>

<template>
  <ListWidget
    title="favs"
    :items="favoritesStore.favorites"
    :create-form="CreateFavorite"
  />
</template>
