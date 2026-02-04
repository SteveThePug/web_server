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
        <p>I'm not a gym geek but I have a consistent routine:</p>
        <div class="overflow-scroll flex-1 border-box">
            <OptionalLinkTable :data="gym" />
        </div>
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
