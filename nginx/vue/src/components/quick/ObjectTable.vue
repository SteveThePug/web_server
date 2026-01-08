<script setup>
import { computed } from "vue";

const props = defineProps({
    objArr: {
        type: Array,
        required: true,
    },
});

const resolvedColumns = computed(() => {
    const keys = new Set();

    for (const obj of props.objArr) {
        Object.keys(obj).forEach((key) => keys.add(key));
    }

    return Array.from(keys).map((key) => ({
        key,
        label: key,
    }));
});
</script>

<template>
    <table>
        <thead>
            <tr>
                <th v-for="col in resolvedColumns" :key="col.key">
                    {{ col.label }}
                </th>
            </tr>
        </thead>

        <tbody>
            <tr v-for="(row, rowIndex) in objArr" :key="rowIndex">
                <td v-for="col in resolvedColumns" :key="col.key">
                    {{ row[col.key] ?? "" }}
                </td>
            </tr>
        </tbody>
    </table>
</template>
