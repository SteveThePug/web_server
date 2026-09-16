<script setup>
/**
 * Renders a list of {name, link, type} either as stacked link buttons
 * (variant="list") or as a two-column table (variant="table").
 *
 * Passing `title` wraps the whole thing in a collapsible <ToggleHeader> (and a
 * .h-fit.w-full div); without a title the items are rendered bare. The items
 * themselves live in LinkTableItems.vue so both branches share one copy.
 * Used by the Bookmarks, Favorites, Consumption and Links widgets.
 */
import { ref } from "vue";
import LinkTableItems from "@/components/util/LinkTableItems.vue";
import ToggleHeader from "@/components/text/ToggleHeader.vue";

const props = defineProps({
  items: {
    type: Array,
    required: true,
  },
  variant: {
    type: String,
    default: "list",
  },
  title: {
    type: String,
    default: "",
  },
});

const show = ref(false);
</script>

<template>
  <div v-if="title" class="h-fit w-full">
    <ToggleHeader v-model="show" class="justify-between flex items-center">
      {{ title }}
    </ToggleHeader>
    <LinkTableItems v-if="show" :items="items" :variant="variant" />
  </div>
  <LinkTableItems v-else :items="items" :variant="variant" />
</template>
