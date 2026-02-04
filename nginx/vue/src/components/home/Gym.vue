<script setup>
import OptionalLinkTable from "@/components/quick/OptionalLinkTable.vue";
const gym = [
    { name: "Row", type: "30 min" },
    { name: "Run", type: "5k" },
    { name: "Pushup & Squat", type: "50" },
];
</script>

<template>
    <div class="flex flex-col items-center">
        <h2>Gym</h2>
        <p class="m-3">I'm not a gym geek but I always do:</p>
        <div class="overflow-scroll w-full flex-1 border-box">
            <OptionalLinkTable class="w-full" :data="gym" />
        </div>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
