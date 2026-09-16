<script setup>
/**
 * Renders a list of {name, link, type} either as stacked link buttons
 * (variant="list") or as a two-column table (variant="table").
 *
 * Passing `title` wraps the whole thing in a collapsible <ToggleHeader>.
 * Used by the Bookmarks, Favorites, Consumption and Links widgets.
 */
import { ref } from "vue";
import Link from "@/components/text/Link.vue";
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
    <template v-if="show">
      <Link
        v-if="variant === 'list'"
        v-for="(item, i) in items"
        :key="i"
        :href="item.link"
      >
        <p class="bdr-2 bg-surface-tint">{{ item.name }}</p>
      </Link>
      <table class="w-full" v-else>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <th>{{ item.type }}</th>
            <td v-if="item.link">
              <Link :href="item.link">{{ item.name }}</Link>
            </td>
            <td v-else>{{ item.name }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </div>
  <template v-else>
    <template v-if="variant === 'list'">
      <Link v-for="(item, i) in items" :key="i" :href="item.link">
        <p class="bdr-2 bg-surface-tint">{{ item.name }}</p>
      </Link>
    </template>
    <table class="w-full" v-else>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <th>{{ item.type }}</th>
          <td v-if="item.link">
            <Link :href="item.link">{{ item.name }}</Link>
          </td>
          <td v-else>{{ item.name }}</td>
        </tr>
      </tbody>
    </table>
  </template>
</template>
