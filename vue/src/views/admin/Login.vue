<script setup>
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";

import Button from "@/components/input/Button.vue";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();
const username = ref("");
const password = ref("");
const submitting = ref(false);
const errorMsg = ref("");

async function handleLogin() {
  if (submitting.value) return;
  submitting.value = true;
  errorMsg.value = "";
  try {
    await auth.logIn(username.value, password.value);
    if (auth.loggedIn) {
      const dest = route.query.redirect || "/admin";
      router.push(dest);
    } else {
      errorMsg.value = "Invalid username or password";
    }
  } finally {
    submitting.value = false;
  }
}

function handleLogout() {
  auth.logOut();
}
</script>

<template>
  <main class="login-page">
    <div v-if="auth.loggedIn" class="login-card bdr-1">
      <h1>Already signed in</h1>
      <p>You're logged in as <code>{{ auth.user.username }}</code>.</p>
      <div class="login-actions">
        <Button @click="router.push('/admin')">Open Admin</Button>
        <Button @click="handleLogout">Log Out</Button>
      </div>
    </div>

    <div v-else class="login-card bdr-1">
      <h1>Login</h1>
      <small class="login-hint">
        Admin access is required beyond this point.
      </small>

      <label class="login-field">
        <span>Username</span>
        <input
          type="text"
          v-model="username"
          autocomplete="username"
          placeholder="Username"
          @keyup.enter="handleLogin"
        />
      </label>

      <label class="login-field">
        <span>Password</span>
        <input
          type="password"
          v-model="password"
          autocomplete="current-password"
          placeholder="Password"
          @keyup.enter="handleLogin"
        />
      </label>

      <small v-if="errorMsg" class="login-error">{{ errorMsg }}</small>

      <Button @click="handleLogin" :disabled="submitting">
        {{ submitting ? "Signing in…" : "Log In" }}
      </Button>
    </div>
  </main>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  padding: 16px;
}

.login-card {
  width: 100%;
  max-width: 380px;
  background-color: var(--color-surface);
  padding: 18px 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.login-hint {
  margin-bottom: 6px;
}

.login-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.login-field span {
  font-family: var(--font-heading);
  color: var(--color-primary);
  letter-spacing: 0.05em;
  padding-left: 2px;
}
.login-field input {
  transition:
    border-color 120ms ease,
    box-shadow 120ms ease,
    background-color 120ms ease;
}
.login-field input:focus {
  outline: none;
  border-color: var(--color-tertiary);
  box-shadow: 0 0 0 2px var(--color-quaternary);
  background-color: var(--color-surface-tint);
}

.login-error {
  color: var(--color-tertiary);
  border: 1px dashed var(--color-tertiary);
  padding: 4px 8px;
}

.login-actions {
  display: flex;
  gap: 6px;
}
</style>
