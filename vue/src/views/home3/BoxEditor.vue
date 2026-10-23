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
 *     store) and is saved by Save or Ctrl/Cmd+S, or automatically on leaving
 *     the box: selecting another one, clicking empty canvas, Done, or closing
 *     the panel. Saving every keystroke would push half-written markup to
 *     every visitor; leaving is the point it is finished.
 *
 * "Insert upload" uploads a file (@/uploads) and puts a tag for it at the
 * cursor — the way to add pictures to an existing box, e.g. a gallery.
 * "Choose upload" does the same with a file that is already uploaded.
 *
 * The panel is teleported to <body>, so it is fixed to the window whatever
 * transform or stacking context the page puts around the canvas. Its width
 * (drag the left edge; double-click it to reset) and the code's text size
 * (A−/A+, or Ctrl/Cmd with +, - or 0 in the editor) are remembered per
 * browser through ./editorPrefs.
 */
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useBoxesStore } from "@/stores/boxes";
import { ACCEPT, tagFor, uploadFile } from "@/uploads";
import CodeEditor from "./CodeEditor.vue";
import { FONT_SIZE, PANEL_WIDTH, fontSize, panelWidth, setPanelWidth, zoom } from "./editorPrefs";
import { WIDGETS } from "./widgetRegistry";
import OverlayShell from "@/components/util/OverlayShell.vue";
import UploadPicker from "@/components/util/UploadPicker.vue";

const props = defineProps({
  box: { type: Object, required: true },
});
const emit = defineEmits(["preview", "close", "error"]);

// A box only ever moves within its own page's store.
const store = useBoxesStore(props.box.page);
const codeEditor = ref(null);
const fileInput = ref(null);
const uploading = ref(false);
const picking = ref(false);
const open = ref(true);

const draftHtml = ref(props.box.html ?? "");
const saving = ref(false);
const dirty = computed(() => props.box.kind === "html" && draftHtml.value !== (props.box.html ?? ""));

// Save the draft of box `id` if it differs from what is stored. Fire and
// forget: it runs as the editor moves on or unmounts, with nothing left to
// show progress in, so a failure is only reported through `error`.
function flush(id) {
  const box = store.boxes.find((b) => b.id === id);
  if (!box || box.kind !== "html" || draftHtml.value === (box.html ?? "")) return;
  store.update(id, { html: draftHtml.value }).catch((err) => emit("error", err));
}

// Switching to another box: save the old one's draft, drop its preview, start
// from the new box's saved html.
watch(
  () => props.box.id,
  (_, oldId) => {
    flush(oldId);
    emit("preview", null);
    draftHtml.value = props.box.html ?? "";
  },
);

// Deselecting, Done and ✕ all unmount the panel.
onBeforeUnmount(() => {
  clearTimeout(previewTimer);
  flush(props.box.id);
  emit("preview", null);
});

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

function insertFile(url) {
  const tag = tagFor(url);
  if (!codeEditor.value?.insert(tag)) draftHtml.value += `\n${tag}`;
}

async function upload(event) {
  const file = event.target.files?.[0];
  event.target.value = "";
  if (!file) return;
  uploading.value = true;
  try {
    insertFile(await uploadFile(file));
  } catch (err) {
    emit("error", err);
  } finally {
    uploading.value = false;
  }
}

function close() {
  open.value = false;
  emit("close");
}

// The panel is pinned to the right of the window, so its width is simply the
// distance from the pointer to that edge. Pointer capture keeps the drag
// alive when the pointer crosses a box's iframe, which would otherwise take
// the events.
function startResize(event) {
  if (event.button !== 0) return;
  event.preventDefault();
  const handle = event.currentTarget;
  handle.setPointerCapture(event.pointerId);
  const move = (e) => setPanelWidth(window.innerWidth - e.clientX);
  const end = () => {
    handle.removeEventListener("pointermove", move);
    handle.removeEventListener("pointerup", end);
    handle.removeEventListener("pointercancel", end);
  };
  handle.addEventListener("pointermove", move);
  handle.addEventListener("pointerup", end);
  handle.addEventListener("pointercancel", end);
}

function resizeByKey(event) {
  const step = event.shiftKey ? 100 : 20;
  if (event.key === "ArrowLeft") setPanelWidth(panelWidth.value + step);
  else if (event.key === "ArrowRight") setPanelWidth(panelWidth.value - step);
  else return;
  event.preventDefault();
}
</script>

<template>
  <OverlayShell v-model="open" title="Box editor" @close="close">
    <aside class="box-editor" :style="{ '--editor-width': `${panelWidth}px` }">
      <header class="flex items-center justify-between gap-2">
        <h3 class="m-0">
          {{ box.kind === "widget" ? `Widget · ${box.widget}` : "HTML box" }}
        </h3>
        <div class="flex items-center gap-2">
          <div v-if="box.kind === 'html'" class="zoom" role="group" aria-label="Code text size">
            <button
              type="button"
              title="Smaller code (Ctrl/Cmd -)"
              :disabled="fontSize <= FONT_SIZE.min"
              @click="zoom(-1)"
            >
              A−
            </button>
            <button type="button" title="Reset code size (Ctrl/Cmd 0)" @click="zoom(0)">
              {{ fontSize }}px
            </button>
            <button
              type="button"
              title="Larger code (Ctrl/Cmd +)"
              :disabled="fontSize >= FONT_SIZE.max"
              @click="zoom(1)"
            >
              A+
            </button>
          </div>
          <button type="button" class="close" title="Close" @click="close">✕</button>
        </div>
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
        <CodeEditor
          ref="codeEditor"
          v-model="draftHtml"
          :font-size="fontSize"
          class="flex-1"
          @save="saveHtml"
          @zoom="zoom"
        />
        <footer class="flex flex-wrap items-center justify-between gap-2">
          <small class="text-muted">
            Runs sandboxed: scripts work, cookies and /api don't. fetch() only reaches
            keyless CORS-open APIs. Ctrl+S saves; so does leaving the box.
          </small>
          <input
            ref="fileInput"
            type="file"
            :accept="ACCEPT.media"
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
          <button
            type="button"
            class="save"
            title="Insert a file that is already uploaded"
            @click="picking = true"
          >
            Choose upload
          </button>
          <button type="button" class="save" :disabled="!dirty || saving" @click="saveHtml">
            {{ saving ? "Saving…" : dirty ? "Save" : "Saved" }}
          </button>
        </footer>
      </template>

      <!-- Last in the panel so the opening focus lands on a header button, not here. -->
      <div
        class="resize-handle"
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize editor"
        :aria-valuenow="panelWidth"
        :aria-valuemin="PANEL_WIDTH.min"
        tabindex="0"
        title="Drag to resize · double-click to reset"
        @pointerdown="startResize"
        @dblclick="setPanelWidth(PANEL_WIDTH.default)"
        @keydown="resizeByKey"
      />
    </aside>
  </OverlayShell>
  <UploadPicker
    v-if="picking"
    v-model="picking"
    :kinds="['image', 'video', 'audio']"
    @pick="insertFile"
  />
</template>

<style scoped>
.box-editor {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(var(--editor-width), 100vw);
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  padding: 0.75rem;
  box-sizing: border-box;
  background: var(--color-surface);
  border-left: 2px solid var(--color-primary);
  color: var(--color-secondary);
}

.resize-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -5px;
  width: 9px;
  cursor: ew-resize;
  touch-action: none;
}
.resize-handle:hover,
.resize-handle:focus-visible,
.resize-handle:active {
  background: linear-gradient(to right, transparent 3px, var(--color-primary) 3px 6px, transparent 6px);
  outline: none;
}

.zoom {
  display: flex;
}
.zoom button {
  min-width: 2rem;
  padding: 0 0.4rem;
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-quaternary);
  font-size: 0.85rem;
  cursor: pointer;
}
.zoom button + button {
  border-left: none;
}
.zoom button:hover:not(:disabled) {
  border-color: var(--color-primary);
}
.zoom button:disabled {
  opacity: 0.5;
  cursor: default;
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
