<script setup>
/**
 * The admin's layouts panel for one canvas page, opened from BoxCanvas's
 * "Layouts" button: save the page as a named layout, or load a saved one
 * onto it (stores/layouts.js).
 *
 * Loading onto a page that already has boxes asks whether to replace them or
 * add the layout on top. A replace is undoable: the server saves the page
 * first, as a layout named "<page> before <layout>, <time>", which shows up
 * at the top of this list. Onto an empty page it simply replaces, so the
 * layout's background comes with it.
 */
import { ref } from "vue";
import { pagePath, useLayoutsStore } from "@/stores/layouts";
import { useBoxesStore } from "@/stores/boxes";
import OverlayShell from "@/components/util/OverlayShell.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";

const props = defineProps({
  page: { type: String, required: true },
});
const emit = defineEmits(["close"]);
const open = ref(true);

const layouts = useLayoutsStore();
layouts.fetch();
const boxes = useBoxesStore(props.page);

const name = ref("");
const busy = ref(false);
const error = ref("");
const loadTarget = ref(null); // the layout waiting on "replace or add?"
const deleteTarget = ref(null);
const renaming = ref(null); // id of the layout whose name is being edited
const renameDraft = ref("");

function close() {
  open.value = false;
  emit("close");
}

async function act(action, failure) {
  error.value = "";
  busy.value = true;
  try {
    await action();
    return true;
  } catch (err) {
    error.value = err?.message || failure;
    return false;
  } finally {
    busy.value = false;
  }
}

async function save() {
  const ok = await act(() => layouts.create({ name: name.value.trim(), page: props.page }), "Failed to save");
  if (ok) name.value = "";
}

function load(layout) {
  if (boxes.boxes.length) loadTarget.value = layout;
  else apply(layout, true);
}

function apply(layout, replace) {
  loadTarget.value = null;
  act(() => layouts.apply(layout.id, props.page, replace), "Failed to load");
}

function startRename(layout) {
  renaming.value = layout.id;
  renameDraft.value = layout.name;
}

async function rename(layout) {
  const ok = await act(() => layouts.update(layout.id, { name: renameDraft.value.trim() }), "Failed to rename");
  if (ok) renaming.value = null;
}

function confirmDelete() {
  const layout = deleteTarget.value;
  deleteTarget.value = null;
  act(() => layouts.remove(layout.id), "Failed to delete");
}

function describe(layout) {
  const parts = [`from ${pagePath(layout.sourcePage)}`, `${layout.boxCount} box${layout.boxCount === 1 ? "" : "es"}`];
  if (layout.hasBackground) parts.push("background");
  return parts.join(" · ");
}
</script>

<template>
  <OverlayShell v-model="open" title="Layouts" @close="close">
    <div class="layout-panel bdr-2 bg-surface">
      <h3>Layouts</h3>

      <form class="row" @submit.prevent="save">
        <input v-model="name" type="text" placeholder="Name for this page's layout" required class="flex-1" />
        <button type="submit" class="bar-btn" :disabled="busy || !boxes.boxes.length">Save</button>
      </form>

      <p v-if="layouts.error" class="text-tertiary">Couldn't load the layouts.</p>
      <p v-else-if="layouts.loaded && !layouts.layouts.length" class="hint">
        No saved layouts yet. Save this page to start one.
      </p>
      <ul v-else class="list">
        <li v-for="layout in layouts.layouts" :key="layout.id">
          <form v-if="renaming === layout.id" class="row" @submit.prevent="rename(layout)">
            <input v-model="renameDraft" type="text" required class="flex-1" />
            <button type="submit" class="bar-btn small" :disabled="busy">Save</button>
            <button type="button" class="bar-btn small" @click="renaming = null">Cancel</button>
          </form>
          <template v-else>
            <span class="name">{{ layout.name }}</span>
            <span class="hint">{{ describe(layout) }}</span>
            <span class="row">
              <button type="button" class="bar-btn small" :disabled="busy" @click="load(layout)">Load here</button>
              <button type="button" class="bar-btn small" @click="startRename(layout)">Rename</button>
              <button type="button" class="bar-btn small danger" @click="deleteTarget = layout">Delete</button>
            </span>
          </template>
        </li>
      </ul>

      <div class="row">
        <button type="button" class="bar-btn" @click="close">Close</button>
        <span v-if="busy" class="hint">Working…</span>
      </div>
      <p v-if="error" class="text-tertiary">{{ error }}</p>
    </div>
  </OverlayShell>

  <ConfirmDialog
    v-if="loadTarget"
    :model-value="true"
    title="Load layout"
    :message="`This page already has ${boxes.boxes.length} box${boxes.boxes.length === 1 ? '' : 'es'}. Replace them with “${loadTarget.name}”, or add it on top? Replacing saves the page as a layout first.`"
    confirm-label="Replace"
    secondary-label="Add on top"
    @confirm="apply(loadTarget, true)"
    @secondary="apply(loadTarget, false)"
    @cancel="loadTarget = null"
  />
  <ConfirmDialog
    v-if="deleteTarget"
    :model-value="true"
    title="Delete layout"
    :message="`Delete the layout “${deleteTarget.name}”? Pages it was loaded onto keep their boxes.`"
    confirm-label="Delete"
    @confirm="confirmDelete"
    @cancel="deleteTarget = null"
  />
</template>

<style scoped>
.layout-panel {
  position: fixed;
  top: 80px;
  left: 16px;
  z-index: 65;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: min(420px, calc(100vw - 32px));
  max-height: calc(100vh - 100px);
  overflow-y: auto;
  padding: 0.75rem;
  box-sizing: border-box;
}
h3 {
  margin: 0;
}
.row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.list li {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--color-quaternary);
}
.name {
  color: var(--color-primary);
}
.hint {
  color: var(--color-muted);
  font-size: 0.85rem;
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
.bar-btn:hover {
  border-color: var(--color-primary);
}
.bar-btn:disabled {
  opacity: 0.5;
}
.bar-btn.small {
  padding: 0.1rem 0.5rem;
  font-size: 0.9rem;
}
.bar-btn.danger:hover {
  border-color: var(--color-tertiary);
  color: var(--color-tertiary);
}
</style>
