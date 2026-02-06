<script setup>
import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";

import Login from "@/components/admin/Login.vue";
import CreateUser from "@/components/admin/CreateUser.vue";
import CreatePost from "@/components/admin/CreatePost.vue";
import CreateFavorite from "@/components/admin/CreateFavorite.vue";
import CreateActivity from "@/components/admin/CreateActivity.vue";

const auth = useAuthStore();
</script>

<template>
    <main class="halftone justify-center flex flex-row w-full h-full">
        <div class="a5page-portrait bdr-1 flex flex-col">
            <Login class="bdr-2 bg-bg_primary" />
            <!--
            <CreateUser class="bdr-2 bg-bg_primary" />
            -->
            <CreatePost class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreateFavorite class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreateActivity class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
        </div>
    </main>
</template>
