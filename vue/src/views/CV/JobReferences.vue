<script setup>
/**
 * The "Quick Reference" box on `/cv/jobs`: reusable profile/experience snippets
 * to copy into application forms, with inline editing. Rendered only by
 * JobApplications.vue; data comes from stores/jobAppReferences.js.
 *
 * It has no error display of its own — the page has one banner, in the parent,
 * so failures are emitted as `fail` (with the error) and every successful
 * mutation emits `ok` so the parent can clear a stale message.
 *
 * Its root sits inside the parent's `.ja-root`, which jobs-shared.css relies on.
 */
import { ref, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useJobAppReferencesStore } from "@/stores/jobAppReferences";
import "./jobs-shared.css";

const emit = defineEmits(["fail", "ok"]);

const store = useJobAppReferencesStore();
const { references } = storeToRefs(store);

const refForm = ref({ category: "profile", label: "", value: "" });
const editingRefId = ref(null);
const editRefForm = ref({});
const REF_CATEGORIES = ["profile", "experience"];

async function fetchReferences() {
  // The store's fetch() records its failure instead of throwing.
  await store.fetch();
  if (store.error) emit("fail", store.error);
}

async function createReference() {
  if (!refForm.value.label || !refForm.value.value) return;
  try {
    await store.create({
      category: refForm.value.category,
      label: refForm.value.label,
      value: refForm.value.value,
    });
    refForm.value = { category: refForm.value.category, label: "", value: "" };
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

function startRefEdit(ref) {
  editingRefId.value = ref.id;
  editRefForm.value = {
    category: ref.category,
    label: ref.label,
    value: ref.value,
  };
}

function cancelRefEdit() {
  editingRefId.value = null;
  editRefForm.value = {};
}

async function saveRefEdit(id) {
  try {
    await store.update(id, {
      category: editRefForm.value.category || undefined,
      label: editRefForm.value.label || undefined,
      value: editRefForm.value.value || undefined,
    });
    editingRefId.value = null;
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

async function deleteReference(ref) {
  if (!confirm(`Delete the reference "${ref.label}"?`)) return;
  try {
    await store.remove(ref.id);
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

function refsByCategory(category) {
  return references.value.filter((r) => r.category === category);
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text);
}

onMounted(fetchReferences);
</script>

<template>
  <div class="ja-ref-section">
    <h3 class="ja-ref-heading">Quick Reference</h3>
    <div v-for="cat in REF_CATEGORIES" :key="cat" class="ja-ref-category">
      <h4 class="ja-ref-cat-label">{{ cat }}</h4>
      <div
        v-for="ref in refsByCategory(cat)"
        :key="ref.id"
        class="ja-ref-item"
      >
        <template v-if="editingRefId !== ref.id">
          <span class="ja-ref-label">{{ ref.label }}</span>
          <span class="ja-ref-value" :title="ref.value">{{ ref.value }}</span>
          <button
            class="cv-btn cv-btn-sm"
            @click="copyToClipboard(ref.value)"
            title="Copy"
          >
            Copy
          </button>
          <button class="cv-btn cv-btn-sm" @click="startRefEdit(ref)">
            Edit
          </button>
          <button
            class="cv-btn cv-btn-sm cv-btn-danger"
            @click="deleteReference(ref)"
          >
            Delete
          </button>
        </template>
        <template v-else>
          <select
            v-model="editRefForm.category"
            class="ja-input ja-input-sm ja-select"
          >
            <option v-for="c in REF_CATEGORIES" :key="c" :value="c">
              {{ c }}
            </option>
          </select>
          <input
            v-model="editRefForm.label"
            class="ja-input ja-input-sm"
            placeholder="Label"
          />
          <input
            v-model="editRefForm.value"
            class="ja-input ja-input-sm"
            placeholder="Value"
          />
          <button
            class="cv-btn cv-btn-sm cv-btn-primary"
            @click="saveRefEdit(ref.id)"
          >
            Save
          </button>
          <button class="cv-btn cv-btn-sm" @click="cancelRefEdit">
            Cancel
          </button>
        </template>
      </div>
      <p v-if="!refsByCategory(cat).length" class="ja-ref-empty">
        No {{ cat }} items yet.
      </p>
    </div>
    <form class="ja-ref-form" @submit.prevent="createReference">
      <select v-model="refForm.category" class="ja-input ja-select">
        <option v-for="c in REF_CATEGORIES" :key="c" :value="c">
          {{ c }}
        </option>
      </select>
      <input
        v-model="refForm.label"
        class="ja-input"
        placeholder="Label *"
        required
      />
      <input
        v-model="refForm.value"
        class="ja-input"
        placeholder="Value *"
        required
      />
      <button type="submit" class="cv-btn cv-btn-primary">Add</button>
    </form>
  </div>
</template>

<style scoped>
/* Colours are the paper palette from CVLayout.vue; buttons are .cv-btn from
   there too, and the inputs are styled by jobs-shared.css. */
.ja-ref-section {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 6px;
  background: var(--color-paper);
}

.ja-ref-heading {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--color-ink-soft);
}

.ja-ref-category {
  margin-bottom: 0.75rem;
}

.ja-ref-cat-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-ink-muted);
  text-transform: capitalize;
  margin-bottom: 0.35rem;
}

.ja-ref-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.3rem 0;
}

.ja-ref-label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-ink-soft);
  min-width: 80px;
}

.ja-ref-value {
  font-size: 0.85rem;
  color: var(--color-ink-muted);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ja-ref-empty {
  font-size: 0.8rem;
  color: var(--color-ink-faint);
  margin: 0.2rem 0;
}

.ja-ref-form {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
  flex-wrap: wrap;
}
</style>
