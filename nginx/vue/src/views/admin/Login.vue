<script setup>
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "@/stores/auth";

import Button from "@/components/input/Button.vue";

const auth = useAuthStore();
const username = ref("");
const password = ref("");

function handleLogin() {
    auth.logIn(username.value, password.value);
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
