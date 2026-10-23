<script setup>
/**
 * Routes `/account` and `/no-access` (both `meta.requiresSignIn`): the landing
 * page for a signed-in account that is not an admin, listing the gated
 * services it has been granted (backend/services/permissions.go) as links.
 *
 * `/no-access` is where nginx's permission gate sends a signed-in visitor who
 * lacks the permission (`@auth_forbidden`, with the refused address in
 * `?path=`). It is this same page with a notice on top, since "what can I
 * open instead" is the useful answer to "you can't open that".
 *
 * The links are plain <a href>, not RouterLinks: every service lives in
 * another container behind nginx, so it needs a real page load.
 */
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import Button from "@/components/input/Button.vue";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const permissionTypes = ref([]);
const loaded = ref(false);
const error = ref("");

const services = computed(() =>
  permissionTypes.value.filter((p) => auth.can(p.key)),
);

// Only shown on /no-access, and only as text, so an odd value is harmless.
const refused = computed(() =>
  route.name === "no-access" && typeof route.query.path === "string"
    ? route.query.path
    : "",
);

onMounted(async () => {
  try {
    permissionTypes.value = await auth.fetchPermissionTypes();
  } catch (err) {
    console.error(err);
    error.value = err.message || "Couldn't load your services";
  } finally {
    loaded.value = true;
  }
});

// The guard only runs on navigation, so logging out in place would leave the
// page up for a signed-out visitor; leave it instead.
async function handleLogout() {
  await auth.logOut();
  router.push("/admin/login");
}
</script>

<template>
  <main class="account-page">
    <div class="account-card bdr-1">
      <p v-if="refused" class="account-refused" role="alert">
        You don't have access to <code>{{ refused }}</code>.
      </p>

      <h1>Your services</h1>
      <small>Signed in as <code>{{ auth.user.username }}</code>.</small>

      <p v-if="error" class="text-tertiary" role="alert">{{ error }}</p>
      <p v-else-if="!loaded" class="text-muted">Loading…</p>
      <ul v-else-if="services.length" class="account-services">
        <li v-for="s in services" :key="s.key">
          <a :href="s.path">{{ s.label }}</a>
        </li>
      </ul>
      <p v-else class="text-muted">Nothing has been shared with you yet.</p>

      <div class="account-actions">
        <Button v-if="auth.user.admin" @click="router.push('/admin')">
          Open Admin
        </Button>
        <Button @click="handleLogout">Log Out</Button>
      </div>
    </div>
  </main>
</template>

<style scoped>
.account-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  padding: 16px;
}

.account-card {
  width: 100%;
  max-width: 380px;
  background-color: var(--color-surface);
  padding: 18px 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.account-refused {
  color: var(--color-tertiary);
  border: 1px dashed var(--color-tertiary);
  padding: 4px 8px;
}

.account-services {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-left: 1.2rem;
  list-style: square;
}

.account-actions {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}
</style>
