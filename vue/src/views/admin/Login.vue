<script setup>
/**
 * Route `/admin/login`. On success redirects to `?redirect=` or to /admin.
 *
 * The `?redirect=` is set by two different gates: the SPA router guard
 * (`meta.requiresAdmin`) and nginx's `auth_request` gate in front of the
 * proxied containers (/notes/, /sb/, /hasura/). The latter means the
 * destination is frequently NOT a Vue route, so see goTo() below.
 *
 * auth.logIn() never throws — it swallows the error — so success is detected by
 * re-reading `auth.loggedIn` afterwards rather than by a try/catch.
 */
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

// Send the browser to `dest`, choosing SPA navigation vs. a real page load.
//
// nginx bounces /notes/, /sb/ and /hasura/ here when the admin cookie is
// missing; those paths belong to other containers, so router.push() would just
// match the catch-all 404 route and never leave the SPA. Resolving the path
// first tells us which it is: the catch-all matches everything, so landing on
// the route named "404" means no real Vue route exists and the browser needs a
// full document request for nginx to proxy it.
//
// `dest` comes from the query string, so it is only followed when it is a
// site-relative path — a leading "//" or a scheme would be an open redirect.
function goTo(dest) {
  if (typeof dest !== "string" || !dest.startsWith("/") || dest.startsWith("//")) {
    router.push("/admin");
    return;
  }
  if (router.resolve(dest).name !== "404") {
    router.push(dest);
  } else {
    window.location.assign(dest);
  }
}

async function handleLogin() {
  if (submitting.value) return;
  submitting.value = true;
  errorMsg.value = "";
  try {
    // logIn() swallows its error, so success is detected by re-reading
    // `loggedIn` rather than by catching.
    await auth.logIn(username.value, password.value);
    if (auth.loggedIn) {
      goTo(route.query.redirect || "/admin");
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
