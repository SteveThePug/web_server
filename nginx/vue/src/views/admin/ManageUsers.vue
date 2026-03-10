<script setup>
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import axios from "axios";

const auth = useAuthStore();
const users = ref([]);

async function fetchUsers() {
    try {
        const res = await axios.get("/api/user");
        users.value = res.data;
    } catch (err) {
        console.error(err);
    }
}

async function toggleAdmin(user) {
    try {
        const res = await auth.setUserAdmin(user.id, !user.admin);
        user.admin = res.admin;
    } catch (err) {
        console.error(err);
    }
}

onMounted(fetchUsers);
</script>

<template>
    <div class="flex flex-col">
        <h1>Manage Users</h1>
        <div v-for="user in users" :key="user.id" class="flex flex-row items-center gap-2">
            <span>{{ user.username }}</span>
            <span v-if="user.admin">(admin)</span>
            <Button
                v-if="user.id !== auth.user.id"
                @click="toggleAdmin(user)"
            >
                {{ user.admin ? "Demote" : "Promote" }}
            </Button>
        </div>
    </div>
</template>
