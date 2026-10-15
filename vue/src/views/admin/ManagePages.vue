<script setup>
/**
 * Admin tab: the canvas pages made in the browser (views/canvas/). Lists them
 * all, admin-only ones included, and creates new ones. A page's title,
 * address and visibility are edited from "Page settings" on the page itself.
 *
 * "New shrine" is the same form with the address fixed under /shrines/. It
 * starts blank and adds no link anywhere: /shrines is itself a canvas page,
 * and its links are boxes put there by hand.
 */
import { usePagesStore } from "@/stores/pages";
import PageForm from "@/views/canvas/PageForm.vue";
import { ref } from "vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";

const pages = usePagesStore();
pages.fetch();
const removeTarget = ref(null);
const confirmationOpen = ref(false);

async function remove(page) {
  removeTarget.value = page;
  confirmationOpen.value = true;
}

async function confirmRemove() {
  const page = removeTarget.value;
  removeTarget.value = null;
  if (!page) return;
  try {
    await pages.remove(page.id);
  } catch (err) {
    alert(err?.message || "Failed to delete");
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <section>
      <h1>Pages</h1>
      <p class="text-muted">
        The home page (/stp) and the landing page's notes are canvases too —
        edit those on the pages themselves.
      </p>
      <p v-if="pages.error" class="text-tertiary">Couldn't load the pages.</p>
      <table v-else-if="pages.pages.length" class="w-full">
        <tr v-for="page in pages.pages" :key="page.id">
          <td>
            <RouterLink :to="`/${page.slug}`">/{{ page.slug }}</RouterLink>
          </td>
          <td>{{ page.title }}</td>
          <td>{{ page.adminOnly ? "🔒 admin only" : "public" }}</td>
          <td class="text-right">
            <button type="button" class="text-tertiary cursor-pointer" @click="remove(page)">Delete</button>
          </td>
        </tr>
      </table>
      <p v-else-if="pages.loaded" class="text-muted">No pages yet.</p>
    </section>

    <section>
      <h1>New shrine</h1>
      <p class="text-muted">
        A blank page under /shrines/. Link to it from the /shrines page with a
        Link box.
      </p>
      <PageForm prefix="shrines/" placeholder="address, e.g. evangelion" submit-label="Create shrine" />
    </section>

    <section>
      <h1>New page</h1>
      <PageForm />
    </section>

    <ConfirmDialog
      v-if="removeTarget"
      v-model="confirmationOpen"
      title="Delete page"
      :message="`Delete the page “${removeTarget.title}” and everything on it?`"
      confirm-label="Delete"
      @confirm="confirmRemove"
      @cancel="removeTarget = null"
    />
  </div>
</template>
