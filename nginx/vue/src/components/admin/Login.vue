<script setup>
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const username = ref("");
const password = ref("");
const loggedIn = computed(() => Object.keys(auth.user.ID || {}).length > 0);

function handleLogin() {
    auth.logIn(username.value, password.value);
}
</script>

<template>
    <div v-if="loggedIn">
        <p>{{ auth.user.ID }}</p>
        <p>{{ auth.user.username }}</p>
    </div>
    <div v-else>
        <h1>login</h1>
        <textarea v-model="username"></textarea>
        <textarea type="password" v-model="password"></textarea>
        <button @click="handleLogin">Log In</button>
    </div>
</template>
