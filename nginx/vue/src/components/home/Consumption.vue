<script setup>
import AutoScroll from "@/components/util/AutoScroll.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import Header from "@/components/text/Header.vue";

const data = [
    {
        type: "Anime",
        name: "Cowboy bebop",
    },
    {
        type: "Game",
        name: "Rounds",
        link: "https://store.steampowered.com/app/1557740/ROUNDS/",
    },
    {
        type: "Boardgame",
        name: "Root",
    },
    {
        type: "Game",
        name: "The farmer was replaced",
        link: "https://store.steampowered.com/app/2060160/The_Farmer_Was_Replaced/",
    },
    {
        type: "Game",
        name: "Higurashi When They Cry Hou - Ch.1",
        link: "https://store.steampowered.com/app/310360/Higurashi_When_They_Cry_Hou__Ch1_Onikakushi/",
    },
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
