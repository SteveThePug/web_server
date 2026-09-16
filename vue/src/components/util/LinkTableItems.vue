<script setup>
/**
 * The body of LinkTable.vue: the items rendered either as stacked link buttons
 * (variant="list") or as a two-column table (variant="table").
 *
 * Split out only so LinkTable can render the same markup inside and outside its
 * collapsible <ToggleHeader> wrapper without duplicating it. A component adds no
 * DOM of its own, so the rendered output is unchanged.
 */
import Link from "@/components/text/Link.vue";

defineProps({
  items: {
    type: Array,
    required: true,
  },
  variant: {
    type: String,
    default: "list",
  },
});
</script>

<template>
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
