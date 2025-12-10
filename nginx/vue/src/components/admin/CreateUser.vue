<script setup>
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const username = ref("");
const password = ref("");
const loggedIn = computed(() => !!auth.user.username);

function handleLogin() {
    auth.createUser(username.value, password.value);
}
</script>

<template>
    <div v-if="loggedIn">
        <h1>Logged in</h1>
        <p>{{ auth.user.ID }}</p>
        <p>{{ auth.user.username }}</p>
    </div>
    <div v-else>
        <h1>Create User</h1>
        <input type="text" v-model="username" placeholder="Username" />
        <input type="password" v-model="password" placeholder="Password" />
        <button @click="handleLogin">Create Account</button>
    </div>
</template>
