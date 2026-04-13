<script setup>
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";

import Button from "@/components/input/Button.vue";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();
const username = ref("");
const password = ref("");

async function handleLogin() {
    await auth.logIn(username.value, password.value);
    if (auth.loggedIn && route.query.redirect) {
        router.push(route.query.redirect);
    }
}

function handleLogout() {
    auth.logOut();
}
</script>

<template>
    <div v-if="auth.loggedIn" class="flex flex-col">
        <h1>Logged in</h1>
        <p>{{ auth.user.id }}</p>
        <p>{{ auth.user.username }}</p>
        <p>{{ auth.user.admin }}</p>
        <Button @click="handleLogout">Log Out</Button>
    </div>
    <div v-else class="flex flex-col">
        <h1>Login</h1>
        <input type="text" v-model="username" placeholder="Username" @keyup.enter="handleLogin" />
        <input type="password" v-model="password" placeholder="Password" @keyup.enter="handleLogin" />
        <Button @click="handleLogin">Log In</Button>
    </div>
</template>
