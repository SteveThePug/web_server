<script setup>
import Header from "@/components/text/Header.vue";
import OptionalLinkTable from "@/components/util/OptionalLinkTable.vue";
const gym = [
    { name: "Row", type: "30 min" },
    { name: "Run", type: "5k" },
    { name: "Pushup & Squat", type: "50" },
];
</script>

<template>
    <div class="flex flex-col place-content-between items-center">
        <Header>Gym</Header>
        <p>I'm not a gym geek</p>
        <p>4/7 days I do:</p>
        <div class="overflow-scroll w-full border-box">
            <OptionalLinkTable class="w-full" :data="gym" />
        </div>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
