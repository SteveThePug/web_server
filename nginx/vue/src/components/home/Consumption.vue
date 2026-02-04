<script setup>
import AutoScroll from "@/components/util/AutoScroll.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import Header from "@/components/text/Header.vue";

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
    <div class="flex flex-col items-center">
        <Header>Consumption</Header>
        <AutoScroll>
            <OptionalLinkTable :data="data" />
        </AutoScroll>
    </div>
</template>
