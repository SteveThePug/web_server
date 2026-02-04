<script setup>
import Login from "@/components/admin/Login.vue";
import CreateUser from "@/components/admin/CreateUser.vue";
import CreatePost from "@/components/admin/CreatePost.vue";
</script>

<template>
    <main class="items-center flex flex-col">
        <div class="background halftone" />
        <div class="a5page-portrait bdr-1 flex flex-col">
            <Login class="bdr-2 bg-bg_primary" />
            <!--
            <CreateUser class="bdr-2 bg-bg_primary" />
            -->
            <CreatePost class="bdr-2 bg-bg_primary" />
        </div>
    </main>
</template>
