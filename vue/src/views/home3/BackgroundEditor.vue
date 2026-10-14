<script setup>
/**
 * The admin's background panel for one canvas page, opened from BoxCanvas's
 * "Background" button: a colour, an image (uploaded, dropped on the panel or
 * given as a URL) and how the image fills the window.
 *
 * Every change previews live through the background store; Cancel (or
 * closing) puts the saved one back. Saving with neither colour nor image is
 * the same as "Theme default", which removes the page's background.
 */
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { FITS, useBackgroundStore } from "@/stores/background";
import { uploadFile, isImage } from "./upload";

const props = defineProps({
  page: { type: String, required: true },
});
const emit = defineEmits(["close"]);

const bg = useBackgroundStore();
const saved = bg.byPage[props.page];
const draft = ref({
  color: saved?.color ?? "",
  image: saved?.image ?? "",
  fit: saved?.fit ?? "cover",
});
watch(draft, (d) => (bg.preview = { page: props.page, ...d }), { deep: true, immediate: true });
onBeforeUnmount(() => {
  if (bg.preview?.page === props.page) bg.preview = null;
});

// <input type="color"> only takes #rrggbb; a colour typed as a name or rgb()
// still works, the swatch just can't show it.
const HEX = /^#[0-9a-f]{6}$/i;
const swatch = computed(() => (HEX.test(draft.value.color) ? draft.value.color : "#04080f"));

const busy = ref(false);
const error = ref("");
const dragOver = ref(false);

async function upload(file) {
  if (!file) return;
  if (!file.type.startsWith("image/")) {
    error.value = "That isn't an image";
    return;
  }
  error.value = "";
  busy.value = true;
  try {
    draft.value.image = await uploadFile(file);
  } catch (err) {
    error.value = err?.message || "Upload failed";
  } finally {
    busy.value = false;
  }
}

function onDrop(event) {
  dragOver.value = false;
  upload(event.dataTransfer?.files?.[0]);
}

async function save() {
  error.value = "";
  busy.value = true;
  try {
    const { color, image, fit } = draft.value;
    if (!color.trim() && !image.trim()) await bg.clear(props.page);
    else await bg.save(props.page, { color, image, fit });
    emit("close");
  } catch (err) {
    error.value = err?.message || "Failed to save";
  } finally {
    busy.value = false;
  }
}

async function resetToDefault() {
  if (!confirm("Go back to the site's default background on this page?")) return;
  error.value = "";
  busy.value = true;
  try {
    await bg.clear(props.page);
    emit("close");
  } catch (err) {
    error.value = err?.message || "Failed to reset";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <form
    class="bg-editor bdr-2 bg-surface"
    :class="{ 'drag-over': dragOver }"
    @submit.prevent="save"
    @dragover.prevent="dragOver = true"
    @dragleave.self="dragOver = false"
    @drop.prevent="onDrop"
  >
    <h3>Page background</h3>

    <label>
      Colour
      <span class="row">
        <input type="color" :value="swatch" @input="draft.color = $event.target.value" />
        <input v-model="draft.color" type="text" placeholder="theme colour" class="flex-1" />
        <button v-if="draft.color" type="button" class="bar-btn small" @click="draft.color = ''">✕</button>
      </span>
    </label>

    <label>
      Image
      <span class="row">
        <input v-model="draft.image" type="text" placeholder="/uploads/… or https://…" class="flex-1" />
        <button v-if="draft.image" type="button" class="bar-btn small" @click="draft.image = ''">✕</button>
      </span>
    </label>
    <span class="row">
      <label class="bar-btn upload">
        {{ busy ? "Working…" : "Upload image…" }}
        <input type="file" accept="image/*" hidden :disabled="busy" @change="upload($event.target.files[0])" />
      </label>
      <span class="hint">or drop one on this panel</span>
    </span>
    <img v-if="draft.image && isImage(draft.image)" :src="draft.image" alt="" class="thumb" />

    <label v-if="draft.image">
      Image fit
      <select v-model="draft.fit">
        <option v-for="f in FITS" :key="f.id" :value="f.id">{{ f.label }}</option>
      </select>
    </label>

    <div class="row">
      <button type="submit" class="bar-btn" :disabled="busy">Save</button>
      <button type="button" class="bar-btn" @click="emit('close')">Cancel</button>
      <button v-if="saved" type="button" class="bar-btn" :disabled="busy" @click="resetToDefault">
        Theme default
      </button>
    </div>
    <p v-if="error" class="text-tertiary">{{ error }}</p>
  </form>
</template>

<style scoped>
.bg-editor {
  position: fixed;
  top: 80px;
  left: 16px;
  z-index: 65;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: min(380px, calc(100vw - 32px));
  padding: 0.75rem;
  box-sizing: border-box;
}
.bg-editor.drag-over {
  outline: 3px dashed var(--color-primary);
}
h3 {
  margin: 0;
}
label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  color: var(--color-muted);
}
.row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}
input[type="color"] {
  width: 2.5rem;
  height: 2.2rem;
  padding: 0;
  border: 1px solid var(--color-quaternary);
  background: none;
  cursor: pointer;
}
select {
  padding: 0.4rem;
  color: var(--color-primary);
  background: var(--color-surface-deep);
  border: 1px solid var(--color-primary);
}
.thumb {
  max-height: 100px;
  max-width: 100%;
  align-self: flex-start;
  border: 1px solid var(--color-quaternary);
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
/* The label is the button; the file input inside it is hidden. */
.bar-btn.upload {
  display: inline-block;
  flex-direction: row;
  color: var(--color-primary);
}
</style>
