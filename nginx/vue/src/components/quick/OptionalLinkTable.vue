<script setup>
// Array will have the form
// [ {type: string, name: string, link?: string}]

const props = defineProps({
    data: {
        type: Array,
        required: true,
    },
});

const keys = ["type", "name", "link"];
</script>

<template>
    <table>
        <tbody>
            <tr v-for="item in data" :key="item.id">
                <th>{{ item.type }}</th>
                <td v-if="item.link">
                    <a :href="item.link">{{ item.name }}</a>
                </td>
                <td v-else>{{ item.name }}</td>
            </tr>
        </tbody>
    </table>
</template>
