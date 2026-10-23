<script setup>
/**
 * The admin's components panel, opened from BoxCanvas's "Components" button:
 * place a saved component on this page, linked or as a copy, and rename or
 * delete them (stores/components.js).
 *
 * Placing is BoxCanvas's job (it knows where the admin is looking), so this
 * only emits `place`. A component is made by "Save as component" in the box
 * editor, or by "Import built-ins", which turns the starters in templates.js
 * into components. Its html is edited through any box linked to it.
 *
 * The list is refetched on opening, because `uses` changes whenever a linked
 * box is placed or deleted, on any page.
 */
import { ref } from "vue";
import { useComponentsStore } from "@/stores/components";
import OverlayShell from "@/components/util/OverlayShell.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";

const emit = defineEmits(["close", "place"]);
const open = ref(true);

const components = useComponentsStore();
components.fetch();

const busy = ref(false);
const error = ref("");
const notice = ref("");
const deleteTarget = ref(null);
const renaming = ref(null); // id of the component whose name is being edited
const renameDraft = ref("");

function close() {
  open.value = false;
  emit("close");
}

async function act(action, failure) {
  error.value = "";
  notice.value = "";
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

function startRename(component) {
  renaming.value = component.id;
  renameDraft.value = component.name;
}

async function rename(component) {
  const ok = await act(
    () => components.update(component.id, { name: renameDraft.value.trim() }),
    "Failed to rename",
  );
  if (ok) renaming.value = null;
}

function confirmDelete() {
  const component = deleteTarget.value;
  deleteTarget.value = null;
  act(() => components.remove(component.id), "Failed to delete");
}

function importBuiltins() {
  act(async () => {
    const added = await components.importBuiltins();
    notice.value = added ? `Imported ${added}.` : "All the built-ins are already here.";
  }, "Failed to import");
}

function describe(component) {
  const parts = [`${component.w}×${component.h}`];
  parts.push(component.uses ? `${component.uses} linked box${component.uses === 1 ? "" : "es"}` : "not linked anywhere");
  if (component.adminOnly) parts.push("admin only");
  return parts.join(" · ");
}

function deleteMessage(component) {
  const base = `Delete the component “${component.name}”?`;
  if (!component.uses) return base;
  return `${base} Boxes linked to it (${component.uses}) keep its html as their own copy, so no page changes.`;
}
</script>

<template>
  <OverlayShell v-model="open" title="Components" @close="close">
    <div class="component-panel bdr-2 bg-surface">
      <h3>Components</h3>
      <p class="hint">
        Linked boxes share the component's html: editing one edits them all. A copy is an ordinary box.
        Save a box as a component from the box editor.
      </p>

      <p v-if="components.error" class="text-tertiary">Couldn't load the components.</p>
      <p v-else-if="components.loaded && !components.components.length" class="hint">
        No components yet. Import the built-in starters, or save a box from the box editor.
      </p>
      <ul v-else class="list">
        <li v-for="component in components.components" :key="component.id">
          <form v-if="renaming === component.id" class="row" @submit.prevent="rename(component)">
            <input v-model="renameDraft" type="text" required class="flex-1" />
            <button type="submit" class="bar-btn small" :disabled="busy">Save</button>
            <button type="button" class="bar-btn small" @click="renaming = null">Cancel</button>
          </form>
          <template v-else>
            <span class="name">{{ component.name }}</span>
            <span class="hint">{{ describe(component) }}</span>
            <span class="row">
              <button type="button" class="bar-btn small" @click="emit('place', component, true)">Place linked</button>
              <button type="button" class="bar-btn small" @click="emit('place', component, false)">Place copy</button>
              <button type="button" class="bar-btn small" @click="startRename(component)">Rename</button>
              <button type="button" class="bar-btn small danger" @click="deleteTarget = component">Delete</button>
            </span>
          </template>
        </li>
      </ul>

      <div class="row">
        <button type="button" class="bar-btn" @click="close">Close</button>
        <button
          type="button"
          class="bar-btn"
          :disabled="busy || !components.loaded"
          title="Add the starters from templates.js that aren't here yet"
          @click="importBuiltins"
        >
          Import built-ins
        </button>
        <span v-if="busy" class="hint">Working…</span>
        <span v-else-if="notice" class="hint">{{ notice }}</span>
      </div>
      <p v-if="error" class="text-tertiary">{{ error }}</p>
    </div>
  </OverlayShell>

  <ConfirmDialog
    v-if="deleteTarget"
    :model-value="true"
    title="Delete component"
    :message="deleteMessage(deleteTarget)"
    confirm-label="Delete"
    @confirm="confirmDelete"
    @cancel="deleteTarget = null"
  />
</template>

<style scoped>
.component-panel {
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
  margin: 0;
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
