<script setup>
import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";

import Login from "./Login.vue";
import CreateUser from "./CreateUser.vue";
import CreatePost from "./CreatePost.vue";
import CreateFavorite from "./CreateFavorite.vue";
import CreateActivity from "./CreateActivity.vue";
import CreateRowing from "./CreateRowing.vue";
import ManageUsers from "./ManageUsers.vue";

const auth = useAuthStore();
</script>

<template>
    <main class="halftone justify-center flex flex-row w-full h-full">
        <div class="bdr-1 flex flex-col">
            <Login class="bdr-2 bg-bg_primary" />
            <CreateUser class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreatePost class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreateFavorite class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreateActivity class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <CreateRowing class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
            <ManageUsers class="bdr-2 bg-bg_primary" v-if="auth.loggedIn" />
        </div>
    </main>
</template>
