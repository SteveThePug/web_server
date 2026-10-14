<script setup>
/**
 * Side panel for the selected box in edit mode.
 *
 * Two save rhythms, matching how each kind of field is used:
 *   - Settings (title, geometry, frame, mobile, widget) save as soon as they
 *     change — on blur/Enter for numbers, on change for the rest — the same
 *     "instant" model as dragging.
 *   - HTML is a draft. Typing updates the box on the canvas live (through the
 *     `preview` event, so the parent can render the draft without touching the
 *     store) and is saved only by Save or Ctrl/Cmd+S. Saving every keystroke
 *     would push half-written markup to every visitor.
 *
 * "Insert upload" stores a file through the chat's upload endpoint (public,
 * unguessable URLs under /uploads/) and puts a tag for it at the cursor —
 * the quick way to fill a shrine with images. It is never the private upload
 * directory: that one is behind a cookie check, and the sandboxed frame
 * loads as an opaque origin, so its requests may not carry the cookie.
 */
import { computed, ref, watch } from "vue";
import { useBoxesStore } from "@/stores/boxes";
import { api } from "@/api";
import CodeEditor from "./CodeEditor.vue";
import { WIDGETS } from "./widgetRegistry";

const props = defineProps({
  box: { type: Object, required: true },
});
const emit = defineEmits(["preview", "close", "error"]);

// A box only ever moves within its own page's store.
const store = useBoxesStore(props.box.page);
const codeEditor = ref(null);
const fileInput = ref(null);
const uploading = ref(false);

const draftHtml = ref(props.box.html ?? "");
const saving = ref(false);
const dirty = computed(() => props.box.kind === "html" && draftHtml.value !== (props.box.html ?? ""));

// Switching to another box: drop this one's preview, start from the new
// box's saved html.
watch(
  () => props.box.id,
  () => {
    emit("preview", null);
    draftHtml.value = props.box.html ?? "";
  },
);

let previewTimer = null;
watch(draftHtml, (html) => {
  clearTimeout(previewTimer);
  previewTimer = setTimeout(() => emit("preview", dirty.value ? html : null), 300);
});

async function save(patch) {
  try {
    await store.update(props.box.id, patch);
  } catch (err) {
    emit("error", err);
  }
}

async function saveHtml() {
  if (!dirty.value || saving.value) return;
  saving.value = true;
  try {
    await store.update(props.box.id, { html: draftHtml.value });
    emit("preview", null);
  } catch (err) {
    emit("error", err);
  } finally {
    saving.value = false;
  }
}

function saveNumber(field, event) {
  const value = Math.round(Number(event.target.value));
  if (!Number.isFinite(value) || value === props.box[field]) {
    event.target.value = props.box[field];
    return;
  }
  save({ [field]: value });
}

// The markup an upload is inserted as, by file type. Anything else (a PDF,
// a text file) becomes a link.
function tagFor(url) {
  if (/\.(jpe?g|png|gif|webp)$/i.test(url)) return `<img src="${url}" alt="">`;
  if (/\.(mp4|webm)$/i.test(url)) return `<video src="${url}" controls loop muted></video>`;
  if (/\.(mp3|ogg)$/i.test(url)) return `<audio src="${url}" controls></audio>`;
  return `<a href="${url}" target="_blank">${url.split("/").pop()}</a>`;
}

async function upload(event) {
  const file = event.target.files?.[0];
  event.target.value = "";
  if (!file) return;
  uploading.value = true;
  try {
    const body = new FormData();
    body.append("file", file);
    const { url } = await api.post("/messages/upload", body, { fallback: "Upload failed" });
    const tag = tagFor(url);
    if (!codeEditor.value?.insert(tag)) draftHtml.value += `\n${tag}`;
  } catch (err) {
    emit("error", err);
  } finally {
    uploading.value = false;
  }
}

function close() {
  if (dirty.value && !confirm("Discard unsaved HTML changes?")) return;
  emit("preview", null);
  emit("close");
}

defineExpose({ close });
</script>

<template>
  <aside class="box-editor" @keydown.esc="close">
    <header class="flex items-center justify-between gap-2">
      <h3 class="m-0">
        {{ box.kind === "widget" ? `Widget · ${box.widget}` : "HTML box" }}
      </h3>
      <button type="button" class="close" title="Close" @click="close">✕</button>
    </header>

    <label class="field">
      <span>Title</span>
      <input
        type="text"
        :value="box.title ?? ''"
        placeholder="(label in edit mode)"
        @change="save({ title: $event.target.value })"
      />
    </label>

    <label v-if="box.kind === 'widget'" class="field">
      <span>Widget</span>
      <select :value="box.widget" @change="save({ widget: $event.target.value })">
        <option v-for="name in Object.keys(WIDGETS)" :key="name" :value="name">{{ name }}</option>
      </select>
    </label>

    <div class="geometry">
      <label v-for="field in ['x', 'y', 'w', 'h', 'z']" :key="field" class="field">
        <span>{{ field }}</span>
        <input
          type="number"
          :value="box[field]"
          @change="saveNumber(field, $event)"
          @keydown.enter="$event.target.blur()"
        />
      </label>
    </div>

    <div class="flex flex-wrap gap-x-4 gap-y-1">
      <label class="check">
        <input type="checkbox" :checked="box.framed" @change="save({ framed: $event.target.checked })" />
        Frame
      </label>
      <label class="check">
        <input
          type="checkbox"
          :checked="box.hideOnMobile"
          @change="save({ hideOnMobile: $event.target.checked })"
        />
        Hide on mobile
      </label>
      <label class="check" title="Only signed-in admins are sent this box">
        <input type="checkbox" :checked="box.adminOnly" @change="save({ adminOnly: $event.target.checked })" />
        Admin only
      </label>
    </div>

    <template v-if="box.kind === 'html'">
      <CodeEditor ref="codeEditor" v-model="draftHtml" class="flex-1" @save="saveHtml" />
      <footer class="flex items-center justify-between gap-2">
        <small class="text-muted">
          Runs sandboxed: scripts work, cookies and /api don't. fetch() only reaches
          keyless CORS-open APIs. Ctrl+S saves.
        </small>
        <input
          ref="fileInput"
          type="file"
          accept="image/*,video/mp4,video/webm,audio/mpeg,audio/ogg"
          hidden
          @change="upload"
        />
        <button
          type="button"
          class="save"
          :disabled="uploading"
          title="Upload a file and insert it at the cursor"
          @click="fileInput.click()"
        >
          {{ uploading ? "Uploading…" : "Insert upload" }}
        </button>
        <button type="button" class="save" :disabled="!dirty || saving" @click="saveHtml">
          {{ saving ? "Saving…" : dirty ? "Save" : "Saved" }}
        </button>
      </footer>
    </template>
  </aside>
</template>

<style scoped>
.box-editor {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 60;
  width: min(460px, 100vw);
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  padding: 0.75rem;
  box-sizing: border-box;
  background: var(--color-surface);
  border-left: 2px solid var(--color-primary);
  color: var(--color-secondary);
}

.close {
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-primary);
  padding: 0 0.5rem;
  cursor: pointer;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 0.85rem;
  color: var(--color-muted);
}
.field input,
.field select {
  padding: 3px 5px;
  color: var(--color-secondary);
  background: var(--color-surface-deep);
  border: 1px solid var(--color-quaternary);
  min-width: 0;
}

.geometry {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0.4rem;
}

.check {
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.9rem;
}

.save {
  flex: none;
  padding: 0.3rem 1rem;
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 2px solid var(--color-primary);
  font-family: var(--font-heading);
  font-size: 1.1rem;
  cursor: pointer;
}
.save:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
