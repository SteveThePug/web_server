<script setup>
import OptionalLinkTable from "@/components/quick/OptionalLinkTable.vue";

const data = [
    {
        type: "Substack",
        name: "By Ellie",
        link: "https://oceandrops.substack.com/p/japan-is-what-late-stage-capitalist?utm_source=%2Finbox&utm_medium=reader2",
    },
    {
        type: "Substack",
        name: "By Cricket Guest",
        link: "https://finalgirldigital.substack.com/p/the-diet-coke-essay?utm_source=%2Finbox&utm_medium=reader2",
    },
    {
        type: "Book",
        name: "The Defining Decade",
    },
    {
        type: "Movie",
        name: "Funny Pages",
    },
    {
        type: "Series",
        name: "Curb Your Enthusiasm",
    },
    {
        type: "Game",
        name: "Hatsune Miku: Project Mirai DX",
    },
    {
        type: "Movie",
        name: "Redline",
    },
    {
        type: "Game",
        name: "Persona 5 Royal",
    },
    {
        type: "Anime",
        name: "GTO",
    },
    {
        type: "Anime",
        name: "Revolutionary Girl Utena",
    },
    {
        type: "Anime",
        name: "Chainsaw man",
    },
    {
        type: "Anime",
        name: "Azumanga Daioh",
    },
];
</script>

<template>
    <div class="flex-col center-content">
        <h2>Consumption</h2>
        <div class="scroll fill">
            <OptionalLinkTable :data="data" />
        </div>
    </div>
</template>
