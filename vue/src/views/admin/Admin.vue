<script setup>
import { ref, computed } from "vue";
import { useAuthStore } from "@/stores/auth";
import Button from "@/components/input/Button.vue";

import CreateUser from "./CreateUser.vue";
import CreatePost from "./CreatePost.vue";
import CreateFavorite from "./CreateFavorite.vue";
import CreateActivity from "./CreateActivity.vue";
import CreateRowing from "./CreateRowing.vue";
import CreateBookmark from "./CreateBookmark.vue";
import CreatePlace from "./CreatePlace.vue";
import ManageUsers from "./ManageUsers.vue";
import ManageRadio from "./ManageRadio.vue";

const auth = useAuthStore();

const sections = [
  { id: "user", label: "User", component: CreateUser },
  { id: "post", label: "Post", component: CreatePost },
  { id: "favorite", label: "Favorite", component: CreateFavorite },
  { id: "activity", label: "Activity", component: CreateActivity },
  { id: "rowing", label: "Rowing", component: CreateRowing },
  { id: "bookmark", label: "Bookmark", component: CreateBookmark },
  { id: "place", label: "Place / Thing To Do", component: CreatePlace },
  { id: "users", label: "Manage Users", component: ManageUsers },
  { id: "radio", label: "Manage Radio", component: ManageRadio },
];

const active = ref("place");
const activeSection = computed(() =>
  sections.find((s) => s.id === active.value),
);

function handleLogout() {
  auth.logOut();
}
</script>

<template>
  <main class="admin-page">
    <div class="admin-shell bdr-1">
      <header class="admin-header">
        <div class="flex flex-col">
          <h1>Admin</h1>
          <small v-if="auth.user?.username">
            signed in as {{ auth.user.username }}
          </small>
        </div>
        <Button @click="handleLogout">Log Out</Button>
      </header>

      <div class="admin-body">
        <nav class="admin-nav">
          <button
            v-for="section in sections"
            :key="section.id"
            class="admin-nav-link"
            :class="{ 'is-active': active === section.id }"
            @click="active = section.id"
          >
            {{ section.label }}
          </button>
        </nav>

        <section class="admin-panel bdr-2 bg-surface">
          <component :is="activeSection.component" />
        </section>
      </div>
    </div>
  </main>
</template>

<style scoped>
.admin-page {
  display: flex;
  justify-content: center;
  width: 100%;
  padding: 8px;
}

.admin-shell {
  width: 100%;
  max-width: 900px;
  background-color: var(--color-surface);
  display: flex;
  flex-direction: column;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  border-bottom: 1px solid var(--color-quaternary);
}

.admin-body {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 8px;
  padding: 8px;
}

.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.admin-nav-link {
  text-align: left;
  padding: 6px 10px;
  color: var(--color-primary);
  background-color: var(--color-link-bg);
  border: 1px solid transparent;
  font-family: var(--font-heading);
  letter-spacing: 0.025em;
  cursor: pointer;
  transition:
    background-color 120ms ease,
    border-color 120ms ease,
    color 120ms ease;
}
.admin-nav-link:hover {
  border-color: var(--color-primary);
}
.admin-nav-link.is-active {
  border-color: var(--color-primary);
  color: var(--color-tertiary);
  background-color: var(--color-surface-tint);
}

.admin-panel {
  padding: 12px;
  min-height: 300px;
}
.admin-panel :deep(h1) {
  margin-top: 0;
}

@media (max-width: 700px) {
  .admin-body {
    grid-template-columns: 1fr;
  }
}
</style>
