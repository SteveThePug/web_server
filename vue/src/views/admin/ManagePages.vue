<script setup>
/**
 * Admin tab: the canvas pages made in the browser (views/canvas/). Lists them
 * all, admin-only ones included, and creates new ones. A page's title,
 * address and visibility are edited from "Page settings" on the page itself.
 *
 * "New shrine" is the same form with the address fixed under /shrines/. It
 * starts blank and adds no link anywhere: /shrines is itself a canvas page,
 * and its links are boxes put there by hand.
 *
 * "Layouts" lists the saved layouts (stores/layouts.js) and copies one onto
 * any page without opening it — the same as "Load here" in that page's
 * Layouts panel, where layouts are also saved and renamed.
 */
import { usePagesStore } from "@/stores/pages";
import { pagePath, useLayoutsStore } from "@/stores/layouts";
import PageForm from "@/views/canvas/PageForm.vue";
import { ref } from "vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";

const pages = usePagesStore();
pages.fetch();
const removeTarget = ref(null);
const confirmationOpen = ref(false);
const removeError = ref("");

const layouts = useLayoutsStore();
layouts.fetch();
const layoutError = ref("");
const layoutNotice = ref("");
// { layout, page } waiting on "replace or add?", and the layout to delete.
const copyRequest = ref(null);
const deleteLayoutTarget = ref(null);

function askCopy(layout, event) {
  const page = event.target.value;
  event.target.value = "";
  if (page) copyRequest.value = { layout, page };
}

async function layoutAction(action, done) {
  layoutError.value = "";
  layoutNotice.value = "";
  try {
    await action();
    layoutNotice.value = done;
  } catch (err) {
    layoutError.value = err?.message || "Failed";
  }
}

function copy(replace) {
  const { layout, page } = copyRequest.value;
  copyRequest.value = null;
  layoutAction(
    () => layouts.apply(layout.id, page, replace),
    `“${layout.name}” ${replace ? "now fills" : "was added to"} ${pagePath(page)}.`,
  );
}

function confirmDeleteLayout() {
  const layout = deleteLayoutTarget.value;
  deleteLayoutTarget.value = null;
  layoutAction(() => layouts.remove(layout.id), `Deleted “${layout.name}”.`);
}

async function remove(page) {
  removeTarget.value = page;
  confirmationOpen.value = true;
}

async function confirmRemove() {
  const page = removeTarget.value;
  removeTarget.value = null;
  if (!page) return;
  removeError.value = "";
  try {
    await pages.remove(page.id);
  } catch (err) {
    removeError.value = err?.message || "Failed to delete";
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <section>
      <h1>Pages</h1>
      <p class="text-muted">
        The landing page (/) is a canvas too, but not listed here — edit it on
        the page itself.
      </p>
      <p v-if="removeError" class="text-tertiary" role="alert">
        {{ removeError }}
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
      <h1>Layouts</h1>
      <p class="text-muted">
        Saved copies of a page's boxes and background. Save one from a page's
        Layouts button; copy it onto another page here or from that page.
        Replacing a page's contents saves them as a layout first.
      </p>
      <p v-if="layoutError" class="text-tertiary" role="alert">{{ layoutError }}</p>
      <p v-else-if="layoutNotice" class="text-muted" role="status">{{ layoutNotice }}</p>
      <p v-if="layouts.error" class="text-tertiary">Couldn't load the layouts.</p>
      <table v-else-if="layouts.layouts.length" class="w-full">
        <tr v-for="layout in layouts.layouts" :key="layout.id">
          <td>{{ layout.name }}</td>
          <td class="text-muted">
            from {{ pagePath(layout.sourcePage) }} · {{ layout.boxCount }} box{{ layout.boxCount === 1 ? "" : "es" }}
          </td>
          <td>
            <select class="copy-select" aria-label="Copy onto page" @change="askCopy(layout, $event)">
              <option value="">Copy onto…</option>
              <option value="index">/ (landing)</option>
              <option v-for="page in pages.pages" :key="page.id" :value="page.slug">/{{ page.slug }}</option>
            </select>
          </td>
          <td class="text-right">
            <button type="button" class="text-tertiary cursor-pointer" @click="deleteLayoutTarget = layout">
              Delete
            </button>
          </td>
        </tr>
      </table>
      <p v-else-if="layouts.loaded" class="text-muted">No saved layouts yet.</p>
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
    <ConfirmDialog
      v-if="copyRequest"
      :model-value="true"
      title="Copy layout"
      :message="`Replace everything on ${pagePath(copyRequest.page)} with “${copyRequest.layout.name}”, or add it on top? Replacing saves the page as a layout first.`"
      confirm-label="Replace"
      secondary-label="Add on top"
      @confirm="copy(true)"
      @secondary="copy(false)"
      @cancel="copyRequest = null"
    />
    <ConfirmDialog
      v-if="deleteLayoutTarget"
      :model-value="true"
      title="Delete layout"
      :message="`Delete the layout “${deleteLayoutTarget.name}”? Pages it was copied onto keep their boxes.`"
      confirm-label="Delete"
      @confirm="confirmDeleteLayout"
      @cancel="deleteLayoutTarget = null"
    />
  </div>
</template>

<style scoped>
.copy-select {
  padding: 0.25rem;
  color: var(--color-primary);
  background: var(--color-surface-deep);
  border: 1px solid var(--color-primary);
}
</style>
