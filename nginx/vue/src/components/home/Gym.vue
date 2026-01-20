<script setup>
import OptionalLinkTable from "@/components/quick/OptionalLinkTable.vue";
const gym = [
    { name: "Row", type: "30 min" },
    { name: "Run", type: "5k" },
    { name: "Pushup & Squat", type: "50" },
];
</script>

<template>
    <div class="flex-col center-content">
        <h2>Gym</h2>
        <p>I'm not a gym geek but I have a consistent routine:</p>
        <OptionalLinkTable class="scroll fill center-text" :data="gym" />
    </div>
</template>

<style scoped>
img {
    width: 100%;
}
</style>
