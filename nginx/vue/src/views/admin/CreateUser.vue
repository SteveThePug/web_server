<script setup>
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import axios from "axios";

const auth = useAuthStore();
const username = ref("");
const password = ref("");
const message = ref("");
const error = ref("");

async function handleCreate() {
    message.value = "";
    error.value = "";
    try {
        const res = await axios.post("/api/user", {
            username: username.value,
            password: password.value,
        });
        message.value = `User "${res.data.username}" created successfully.`;
        username.value = "";
        password.value = "";
    } catch (err) {
        error.value = err.response?.data?.message || "Failed to create user.";
    }
}
</script>

<template>
    <div v-if="auth.loggedIn && auth.user.admin" class="flex flex-col">
        <h1>Create User</h1>
        <p v-if="message" class="text-green-500">{{ message }}</p>
        <p v-if="error" class="text-red-500">{{ error }}</p>
        <input type="text" v-model="username" placeholder="Username" @keyup.enter="handleCreate" />
        <input type="password" v-model="password" placeholder="Password" @keyup.enter="handleCreate" />
        <Button @click="handleCreate">Create Account</Button>
    </div>
    <div v-else-if="auth.loggedIn" class="flex flex-col">
        <p>You do not have permission to create users.</p>
    </div>
    <div v-else class="flex flex-col">
        <p>You must be logged in as an admin to create users.</p>
    </div>
</template>
