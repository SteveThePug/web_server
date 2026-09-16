<script setup>
/**
 * Root component. Deliberately empty: all chrome lives in the layout components
 * (layouts/DefaultLayout.vue, layouts/CVLayout.vue) that the router mounts.
 */
import { RouterView } from "vue-router";
</script>

<template>
  <RouterView />
</template>
