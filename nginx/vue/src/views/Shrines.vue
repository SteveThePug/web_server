<script setup>
import RouterTable from "@/components/quick/RouterTable.vue";
const shrine_links = [
    { name: "Demoman", link: "/shrines/demoman" },
    { name: "Evangelion", link: "/shrines/evangelion" },
    { name: "GTO", link: "/shrines/gto" },
    { name: "Skipskipbenben", link: "/shrines/skipskipbenben" },
];
</script>

<template>
    <main class="center-content flex-col">
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <RouterTable :linkArr="shrine_links" />
        </div>
    </main>
</template>
