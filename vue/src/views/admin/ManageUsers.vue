<script setup>
/**
 * Admin panel: list users and toggle their admin flag. Mutates the local `user`
 * object in place from the mutation's response, so the row updates without a
 * refetch. A failed load or toggle is shown above the table rather than only
 * logged, and a row's button is disabled while its toggle is in flight.
 */
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { gql } from "@/graphql";

const auth = useAuthStore();
const users = ref([]);
const loaded = ref(false);
const error = ref("");
const pending = ref(null);

async function fetchUsers() {
  try {
    const data = await gql(`query { users { id username admin } }`);
    users.value = data.users;
  } catch (err) {
    console.error(err);
    error.value = err.message || "Couldn't load the users";
  } finally {
    loaded.value = true;
  }
}

async function toggleAdmin(user) {
  error.value = "";
  pending.value = user.id;
  try {
    const data = await auth.setUserAdmin(user.id, !user.admin);
    user.admin = data.admin;
  } catch (err) {
    console.error(err);
    error.value = err.message || `Couldn't change ${user.username}`;
  } finally {
    pending.value = null;
  }
}

onMounted(fetchUsers);
</script>

<template>
  <div class="flex flex-col gap-2">
    <h1>Manage Users</h1>
    <p v-if="error" class="text-tertiary" role="alert">{{ error }}</p>
    <p v-if="!loaded" class="text-muted">Loading…</p>
    <table v-else-if="users.length" class="w-full">
      <tr v-for="user in users" :key="user.id">
        <td class="py-1">{{ user.username }}</td>
        <td class="py-1">{{ user.admin ? "admin" : "user" }}</td>
        <td class="py-1 text-right">
          <span v-if="user.id === auth.user.id" class="text-muted">you</span>
          <Button
            v-else
            :disabled="pending === user.id"
            @click="toggleAdmin(user)"
          >
            {{ user.admin ? "Demote" : "Promote" }}
          </Button>
        </td>
      </tr>
    </table>
    <p v-else-if="!error" class="text-muted">No users.</p>
  </div>
</template>
