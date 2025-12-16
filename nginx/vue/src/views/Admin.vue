<script setup>
import Login from "@/components/admin/Login.vue";
import CreateUser from "@/components/admin/CreateUser.vue";
import CreatePost from "@/components/admin/CreatePost.vue";

import Halftone from "@/components/backgrounds/Halftone.vue";
</script>

<template>
    <main class="center-content flex-col">
        <Halftone />
        <div class="a5page-portrait bdr-1 grid">
            <Login class="bdr-2 bg-secondary" />
            <CreateUser class="bdr-2 bg-secondary" />
            <CreatePost class="bdr-2 bg-secondary" />
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
