<script setup>
/**
 * STATUS: unused — unrouted. /shrines is now an admin-made canvas page whose
 * links are boxes; new shrines are made on the admin panel's Pages tab.
 *
 * Was route `/shrines` — index of the shrine pages: every canvas page
 * (views/canvas/) whose address starts with "shrines/", from the pages query.
 * Admins — and only admins; the server refuses createPage to anyone else
 * anyway — get a button that opens a form to start one from the Shrine
 * layout (since removed).
 */
import { computed, ref } from "vue";
import RouterTable from "@/components/util/RouterTable.vue";
import { useAuthStore } from "@/stores/auth";
import { usePagesStore } from "@/stores/pages";
import PageForm from "@/views/canvas/PageForm.vue";

const auth = useAuthStore();
const pages = usePagesStore();
auth.ensureLoaded();
pages.fetch();

const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);
const creating = ref(false);

const shrine_links = computed(() =>
  pages.pages
    .filter((p) => p.slug.startsWith("shrines/"))
    .map((p) => ({ name: p.adminOnly ? `${p.title} 🔒` : p.title, link: `/${p.slug}` }))
    .sort((a, b) => a.name.localeCompare(b.name)),
);
</script>

<template>
  <main class="items-center flex flex-col">
    <div
      class="page-a4 bdr-1 flex flex-col relative overflow-scroll gap-1"
    >
      <RouterTable :linkArr="shrine_links" />
      <p v-if="pages.loaded && !shrine_links.length" class="text-muted text-center">
        No shrines yet.
      </p>
      <template v-if="isAdmin">
        <button type="button" class="bdr-2 bg-surface mt-4 cursor-pointer" @click="creating = !creating">
          {{ creating ? "Cancel" : "+ New shrine" }}
        </button>
        <PageForm v-if="creating" prefix="shrines/" />
      </template>
    </div>
  </main>
</template>
