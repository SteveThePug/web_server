<script setup>
import RouterTable from "@/components/quick/RouterTable.vue";
const res = await axios.post("/api/notes/Welcome");
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <RouterTable :linkArr="note_links" />
        </div>
    </main>
</template>
