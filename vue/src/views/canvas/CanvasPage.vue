<script setup>
/**
 * The router's catch-all: any path no other route claims is looked up as an
 * admin-created canvas page (the `Page` GraphQL type) and, if there is one,
 * rendered as a BoxCanvas whose page key is the slug. Otherwise it is the
 * site's 404 — with, for an admin, a form to create a page right here.
 *
 * Static routes always win, which is why the server refuses slugs under the
 * SPA's and nginx's own paths (validateSlug in backend/graph/page_helpers.go).
 *
 * An adminOnly page answers null to anyone else, so to them it is simply a
 * 404.
 */
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { gql } from "@/graphql";
import { useAuthStore } from "@/stores/auth";
import { PAGE_FIELDS, usePagesStore } from "@/stores/pages";
import BoxCanvas from "@/views/home3/BoxCanvas.vue";
import NotFound from "@/views/404/404.vue";
import PageForm from "./PageForm.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const pages = usePagesStore();

const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);

// pathMatch is the path split on "/"; a trailing slash leaves an empty last
// part.
const slug = computed(() => {
  const parts = route.params.pathMatch;
  return (Array.isArray(parts) ? parts : [parts]).filter(Boolean).join("/");
});

const page = ref(null);
const state = ref("loading"); // loading | found | missing | error

async function load() {
  const wanted = slug.value;
  state.value = "loading";
  try {
    const data = await gql(`query Page($slug: String!) { page(slug: $slug) { ${PAGE_FIELDS} } }`, {
      slug: wanted,
    });
    if (wanted !== slug.value) return; // navigated on while it loaded
    page.value = data.page;
    state.value = data.page ? "found" : "missing";
  } catch {
    if (wanted === slug.value) state.value = "error";
  }
}
// The session matters too: signing in may reveal an adminOnly page.
watch([slug, isAdmin], load, { immediate: true });
auth.ensureLoaded();

// ---- Page settings (admin) -----------------------------------------------
const settingsOpen = ref(false);
const draft = ref({ title: "", slug: "", adminOnly: false });
const settingsError = ref("");
const deleteConfirmationOpen = ref(false);

function openSettings() {
  draft.value = { title: page.value.title, slug: page.value.slug, adminOnly: page.value.adminOnly };
  settingsError.value = "";
  settingsOpen.value = !settingsOpen.value;
}

async function saveSettings() {
  settingsError.value = "";
  try {
    const updated = await pages.update(page.value.id, {
      title: draft.value.title.trim(),
      slug: draft.value.slug.trim().replace(/^\/+|\/+$/g, ""),
      adminOnly: draft.value.adminOnly,
    });
    settingsOpen.value = false;
    if (updated.slug !== page.value.slug) router.replace(`/${updated.slug}`);
    else page.value = updated;
  } catch (err) {
    settingsError.value = err?.message || "Failed to save";
  }
}

async function deletePage() {
  deleteConfirmationOpen.value = true;
}

async function confirmDeletePage() {
  deleteConfirmationOpen.value = false;
  try {
    await pages.remove(page.value.id);
    router.push("/");
  } catch (err) {
    settingsError.value = err?.message || "Failed to delete";
  }
}
</script>

<template>
  <main v-if="state === 'found'" class="canvas-page">
    <BoxCanvas :key="page.slug" :page="page.slug">
      <template #admin-bar>
        <button type="button" class="bar-btn" :class="{ on: settingsOpen }" @click="openSettings">
          Page settings
        </button>
        <span v-if="page.adminOnly" class="text-tertiary" title="Only admins can see this page">🔒 admin only</span>
      </template>
    </BoxCanvas>

    <form v-if="settingsOpen" class="settings bdr-2 bg-surface" @submit.prevent="saveSettings">
      <label>Title <input v-model="draft.title" type="text" required /></label>
      <label>Address <input v-model="draft.slug" type="text" required /></label>
      <label class="check"><input v-model="draft.adminOnly" type="checkbox" /> Only admins can see this page</label>
      <div class="flex gap-2">
        <button type="submit" class="bar-btn">Save</button>
        <button type="button" class="bar-btn" @click="settingsOpen = false">Cancel</button>
        <button type="button" class="bar-btn danger" @click="deletePage">Delete page</button>
      </div>
      <p v-if="settingsError" class="text-tertiary">{{ settingsError }}</p>
    </form>
  </main>

  <p v-else-if="state === 'error'" class="state">Couldn't load this page.</p>

  <template v-else-if="state === 'missing'">
    <section v-if="isAdmin" class="create bdr-2 bg-surface">
      <h3>No page here yet — make one?</h3>
      <PageForm :slug="slug" @created="load" />
    </section>
    <NotFound />
  </template>

  <ConfirmDialog
    v-if="deleteConfirmationOpen"
    v-model="deleteConfirmationOpen"
    title="Delete page"
    :message="`Delete the page “${page.title}” and everything on it?`"
    confirm-label="Delete"
    @confirm="confirmDeletePage"
  />
</template>

<style scoped>
.canvas-page {
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
  padding: 10px 16px;
}
.state {
  text-align: center;
  color: var(--color-muted);
}

.bar-btn {
  padding: 0.25rem 0.75rem;
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 2px solid var(--color-quaternary);
  font-family: var(--font-heading);
  font-size: 1.1rem;
  letter-spacing: 0.025em;
  cursor: pointer;
}
.bar-btn:hover,
.bar-btn.on {
  border-color: var(--color-primary);
}
.bar-btn.danger:hover {
  border-color: var(--color-tertiary);
  color: var(--color-tertiary);
}

.settings,
.create {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 420px;
  margin: 0 auto 10px;
  padding: 0.75rem;
}
.settings {
  position: fixed;
  top: 80px;
  left: 16px;
  z-index: 65;
}
.settings label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  color: var(--color-muted);
}
.settings label.check {
  flex-direction: row;
  align-items: center;
  gap: 0.4rem;
}
.settings label.check input {
  width: auto;
}
</style>
