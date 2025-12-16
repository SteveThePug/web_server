<script setup>
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const username = ref("");
const password = ref("");

function handleLogin() {
    auth.createUser(username.value, password.value);
}
</script>

<template>
    <div v-if="auth.loggedIn" class="flex-col">
        <h1>Logged in</h1>
        <p>{{ auth.user.id }}</p>
        <p>{{ auth.user.username }}</p>
        <p>{{ auth.user.admin }}</p>
    </div>
    <div v-else class="flex-col">
        <h1>Create User</h1>
        <input type="text" v-model="username" placeholder="Username" />
        <input type="password" v-model="password" placeholder="Password" />
        <button @click="handleLogin">Create Account</button>
    </div>
</template>
