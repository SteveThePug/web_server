<script setup>
import Login from "@/components/admin/Login.vue";
</script>

<template>
    <main class="center-content flex-col">
        <div class="a5page-portrait bdr-1 grid">
            <Login />
        </div>
    </main>
</template>

<style scoped>
.grid {
    display: grid;
    grid-gap: 10px;
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: repeat(2, 1fr);
}
</style>
