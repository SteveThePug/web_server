<script setup>
import { RouterView } from "vue-router";
</script>

<template>
    <div class="cv-layout">
        <RouterView />
    </div>
</template>

<style scoped>
.cv-layout {
    min-height: 100vh;
    background: white;
    color: #111;
}
</style>
