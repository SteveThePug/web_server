<script setup>
/**
 * "Consumption" widget on /stp (books/films/games logged as activities). Same
 * shape as Favorites.vue — both are ListWidget.vue with a different store.
 */
import ListWidget from "@/views/home/ListWidget.vue";

import { defineAsyncComponent } from "vue";
import { useActivityStore } from "@/stores/activity";

const CreateActivity = defineAsyncComponent(
  () => import("@/views/admin/CreateActivity.vue"),
);

const activityStore = useActivityStore();
</script>

<template>
  <ListWidget
    title="Consumption"
    :items="activityStore.activity"
    :create-form="CreateActivity"
  />
</template>
