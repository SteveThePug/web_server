<script setup>
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "@/stores/auth";

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
    <div v-if="auth.loggedIn">
        <h1>Logged in</h1>
        <p>{{ auth.user.ID }}</p>
        <p>{{ auth.user.username }}</p>
        <button @click="handleLogout">Log Out</button>
    </div>
    <div v-else>
        <h1>Login</h1>
        <input type="text" v-model="username" placeholder="Username" />
        <input type="password" v-model="password" placeholder="Password" />
        <button @click="handleLogin">Log In</button>
    </div>
</template>
