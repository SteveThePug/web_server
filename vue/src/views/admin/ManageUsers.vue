<script setup>
/**
 * Admin panel: list users, toggle their admin flag, and grant non-admins the
 * gated services in `permissionTypes` (the receipt printer and so on; see
 * backend/services/permissions.go). Mutates the local `user` object in place
 * from each mutation's response, so the row updates without a refetch. A
 * failed load or change is shown above the table rather than only logged, and
 * a row's controls are disabled while a change to it is in flight.
 *
 * Admins pass every permission check on the server, so their checkboxes are
 * replaced with "all" rather than shown with grants that would mean nothing.
 * A grant or revoke applies on the user's next request (the gate reads the
 * database, not their token), so there is nothing to tell them to do.
 */
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import { useAuthStore, USER_FIELDS } from "@/stores/auth";
import { gql } from "@/graphql";

const auth = useAuthStore();
const users = ref([]);
const permissionTypes = ref([]);
const loaded = ref(false);
const error = ref("");
const pending = ref(null);

async function fetchUsers() {
  try {
    const data = await gql(
      `query { users { ${USER_FIELDS} } permissionTypes { key label path } }`,
    );
    users.value = data.users;
    permissionTypes.value = data.permissionTypes;
  } catch (err) {
    console.error(err);
    error.value = err.message || "Couldn't load the users";
  } finally {
    loaded.value = true;
  }
}

// Shared by both kinds of change: one in flight per row, the row replaced
// from the server's answer, failures shown above the table.
async function change(user, mutate) {
  error.value = "";
  pending.value = user.id;
  try {
    Object.assign(user, await mutate());
  } catch (err) {
    console.error(err);
    error.value = err.message || `Couldn't change ${user.username}`;
  } finally {
    pending.value = null;
  }
}

function toggleAdmin(user) {
  return change(user, () => auth.setUserAdmin(user.id, !user.admin));
}

// The mutation replaces the whole list, so send the row's list with this one
// key flipped. Afterwards the box is set from the row: on a failure the row
// is unchanged, so Vue sees nothing to patch and would leave the box showing
// the click that did not take.
async function togglePermission(user, key, event) {
  const current = user.permissions ?? [];
  const next = current.includes(key)
    ? current.filter((k) => k !== key)
    : [...current, key];
  await change(user, () => auth.setUserPermissions(user.id, next));
  event.target.checked = (user.permissions ?? []).includes(key);
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
        <td class="py-1">
          <span v-if="user.admin" class="text-muted">all services</span>
          <div v-else class="grants">
            <label
              v-for="p in permissionTypes"
              :key="p.key"
              class="grant"
              :title="p.path"
            >
              <input
                type="checkbox"
                :checked="(user.permissions ?? []).includes(p.key)"
                :disabled="pending === user.id"
                @change="togglePermission(user, p.key, $event)"
              />
              {{ p.label }}
            </label>
          </div>
        </td>
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

<style scoped>
.grants {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem 0.8rem;
}
.grant {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}
.grant input {
  width: auto;
}
</style>
