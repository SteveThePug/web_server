<script setup>
import OptionalLinkTable from "@/components/quick/OptionalLinkTable.vue";
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
    <div class="flex-col center-content">
        <h2>favs</h2>
        <div class="scroll fill">
            <OptionalLinkTable class="scroll" :data="favs" />
        </div>
    </div>
</template>
