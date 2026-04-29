<script setup>
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { gql } from "@/graphql";

const auth = useAuthStore();
const users = ref([]);

async function fetchUsers() {
  try {
    const data = await gql(`query { users { id username admin } }`);
    users.value = data.users;
  } catch (err) {
    console.error(err);
  }
}

async function toggleAdmin(user) {
  try {
    const data = await auth.setUserAdmin(user.id, !user.admin);
    user.admin = data.admin;
  } catch (err) {
    console.error(err);
  }
}

onMounted(fetchUsers);
</script>

<template>
  <div class="flex flex-col">
    <h1>Manage Users</h1>
    <div
      v-for="user in users"
      :key="user.id"
      class="flex flex-row items-center gap-2"
    >
      <span>{{ user.username }}</span>
      <span v-if="user.admin">(admin)</span>
      <Button v-if="user.id !== auth.user.id" @click="toggleAdmin(user)">
        {{ user.admin ? "Demote" : "Promote" }}
      </Button>
    </div>
  </div>
</template>
