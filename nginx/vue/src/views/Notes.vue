<script setup>
import RouterTable from "@/components/quick/RouterTable.vue";
const note_links = [
    { name: "ILSP", link: "/notes/islp" },
    { name: "Vue", link: "/notes/vue" },
];
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <RouterTable :linkArr="note_links" />
        </div>
    </main>
</template>
