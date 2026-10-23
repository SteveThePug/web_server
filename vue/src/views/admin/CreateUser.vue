<script setup>
/**
 * Admin form: create an account. Reports success or failure inline.
 */
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { gql } from "@/graphql";

const auth = useAuthStore();
const username = ref("");
const password = ref("");
const message = ref("");
const error = ref("");
const submitting = ref(false);

async function handleCreate() {
  if (submitting.value) return;
  message.value = "";
  error.value = "";
  submitting.value = true;
  try {
    const data = await gql(
      `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id username } }`,
      { input: { username: username.value, password: password.value } },
    );
    message.value = `User "${data.createUser.username}" created successfully.`;
    username.value = "";
    password.value = "";
  } catch (err) {
    error.value = err.message || "Failed to create user.";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <form
    v-if="auth.loggedIn && auth.user.admin"
    class="flex flex-col"
    @submit.prevent="handleCreate"
  >
    <h1>Create User</h1>
    <p v-if="message" class="text-secondary" role="status">{{ message }}</p>
    <p v-if="error" class="text-tertiary" role="alert">{{ error }}</p>
    <input
      type="text"
      v-model="username"
      placeholder="Username"
      autocomplete="off"
    />
    <input
      type="password"
      v-model="password"
      placeholder="Password"
      autocomplete="new-password"
    />
    <Button type="submit" :disabled="submitting">
      {{ submitting ? "Creating…" : "Create Account" }}
    </Button>
  </form>
  <div v-else-if="auth.loggedIn" class="flex flex-col">
    <p>You do not have permission to create users.</p>
  </div>
  <div v-else class="flex flex-col">
    <p>You must be logged in as an admin to create users.</p>
  </div>
</template>
