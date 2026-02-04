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
    <main class="items-center flex flex-col">
        <div class="background halftone" />
        <div
            class="a4page-portrait bdr-1 flex flex-col relative overflow-scroll gap-1"
        >
            <RouterTable :linkArr="shrine_links" />
        </div>
    </main>
</template>
