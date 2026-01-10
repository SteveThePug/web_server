<script setup>
import Login from "@/components/admin/Login.vue";
import CreateUser from "@/components/admin/CreateUser.vue";
import CreatePost from "@/components/admin/CreatePost.vue";
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div class="a5page-portrait bdr-1 flex-col">
            <Login class="bdr-2 bg-secondary" />
            <!--
            <CreateUser class="bdr-2 bg-secondary" />
            -->
            <CreatePost class="bdr-2 bg-secondary" />
        </div>
    </main>
</template>
