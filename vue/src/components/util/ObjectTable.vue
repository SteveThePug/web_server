<script setup>
/**
 * Generic table that derives its columns from the union of keys across all rows,
 * so heterogeneous objects still line up. STATUS: not imported anywhere.
 */
import { computed } from "vue";

const props = defineProps({
  objArr: {
    type: Array,
    required: true,
  },
});

// Union of keys across every row, in first-seen order, so rows with different
// shapes still line up under one header.
const resolvedColumns = computed(() => {
  const keys = new Set();

  for (const obj of props.objArr) {
    Object.keys(obj).forEach((key) => keys.add(key));
  }

  return Array.from(keys);
});
</script>

<template>
  <table>
    <thead>
      <tr>
        <th v-for="col in resolvedColumns" :key="col">
          {{ col }}
        </th>
      </tr>
    </thead>

    <tbody>
      <tr v-for="(row, rowIndex) in objArr" :key="rowIndex">
        <td v-for="col in resolvedColumns" :key="col">
          {{ row[col] ?? "" }}
        </td>
      </tr>
    </tbody>
  </table>
</template>
