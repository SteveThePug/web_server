<script setup>
import Header from "@/components/text/Header.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
import AutoScroll from "@/components/util/AutoScroll.vue";
const favs = [
    {
        type: "Daioh",
        name: "Tomo",
    },
    {
        type: "Color",
        name: "Teal",
    },
    {
        type: "Pet",
        name: "Cat",
    },
    {
        type: "Subject",
        name: "Math",
    },
    {
        type: "Drink",
        name: "Monster",
    },
    {
        type: "Sport",
        name: "Running",
    },
    {
        type: "Anime",
        name: "Evangelion",
    },
    {
        type: "Ingredient",
        name: "Eggs",
    },
    {
        type: "OS",
        name: "Linux",
    },
    {
        type: "Fruit",
        name: "Satsumas",
    },
    {
        type: "Vegetable",
        name: "Cucumber",
    },
    {
        type: "Hobby",
        name: "Climbing",
    },
    {
        type: "Season",
        name: "Summer",
    },
    {
        type: "Language",
        name: "Rust",
    },
];
</script>

<template>
    <div class="flex flex-col items-center">
        <Header>favs</Header>
        <AutoScroll class="w-full flex-1">
            <OptionalLinkTable class="w-full" :data="favs" />
        </AutoScroll>
    </div>
</template>
