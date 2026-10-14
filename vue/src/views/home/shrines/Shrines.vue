<script setup>
/**
 * Route `/shrines` — index of the shrine pages. Demoman is a hand-written
 * view; every other shrine is a canvas page (views/canvas/) whose address
 * starts with "shrines/", listed from the pages query. Admins get a form to
 * start a new one from the Shrine layout.
 */
import { computed } from "vue";
import RouterTable from "@/components/util/RouterTable.vue";
import { useAuthStore } from "@/stores/auth";
import { usePagesStore } from "@/stores/pages";
import PageForm from "@/views/canvas/PageForm.vue";

const auth = useAuthStore();
const pages = usePagesStore();
auth.ensureLoaded();
pages.fetch();

const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);

const shrine_links = computed(() =>
  [
    { name: "Demoman", link: "/shrines/demoman" },
    ...pages.pages
      .filter((p) => p.slug.startsWith("shrines/"))
      .map((p) => ({ name: p.adminOnly ? `${p.title} 🔒` : p.title, link: `/${p.slug}` })),
  ].sort((a, b) => a.name.localeCompare(b.name)),
);
</script>

<template>
  <main class="items-center flex flex-col">
    <div
      class="page-a4 bdr-1 flex flex-col relative overflow-scroll gap-1"
    >
      <RouterTable :linkArr="shrine_links" />
      <section v-if="isAdmin" class="mt-4">
        <h3>New shrine</h3>
        <PageForm prefix="shrines/" starter="shrine" />
      </section>
    </div>
  </main>
</template>
