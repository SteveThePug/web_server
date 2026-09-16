<script setup>
/**
 * A column of in-app <RouterLink> buttons from a list of {name, link}. Used by
 * the Links widget and the shrines index.
 */
defineProps({
  linkArr: {
    type: Array,
    required: true,
  },
});
</script>

<template>
  <RouterLink
    class="bdr-2 bg-surface"
    v-for="(row, rowIndex) in linkArr"
    :key="rowIndex"
    :to="row.link"
  >
    {{ row.name }}
  </RouterLink>
</template>
