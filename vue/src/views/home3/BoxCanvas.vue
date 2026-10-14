<script setup>
/**
 * A free-form canvas of boxes the admin arranges in the browser, for one page
 * key (see stores/boxes.js). Home3.vue (/stp), the landing page's overlay and
 * every admin-created page (views/canvas/CanvasPage.vue) are this component.
 *
 * Every box (the `Box` GraphQL type) has a position, size and z-index in px
 * on a fixed CANVAS_W-wide design canvas. The canvas is drawn at that width
 * and shrunk with transform: scale() to fit narrower windows, so what the
 * admin arranges is exactly what every desktop visitor sees, just smaller.
 * The canvas is as tall as its lowest box.
 *
 * At phone width the canvas would be unreadably small, so the boxes are laid
 * out in a single column instead: reading order is top-to-bottom then
 * left-to-right by their canvas position, each keeps its height, and any box
 * flagged hideOnMobile is skipped. There is no separate mobile layout to
 * maintain, and no editing at that width.
 *
 * Admins get an Edit toggle. In edit mode boxes drag and resize (BoxFrame),
 * changes save on release, and the selected box opens in BoxEditor. Boxes
 * marked adminOnly are never sent to anyone else (the server filters them);
 * admins see them outlined. Image and video files dropped onto the canvas in
 * edit mode are uploaded and each becomes a box where it was dropped.
 *
 * `overlay` lays the canvas over whatever the parent already renders instead
 * of taking space of its own — the landing page uses it to pin notes over its
 * fixed content. The parent must be position: relative; the canvas covers it
 * from the top, lets clicks through everywhere except on boxes (until edit
 * mode, which owns the whole area), and reports its height through `height`
 * so the parent can grow to contain it. On a phone the overlay's boxes stack
 * below the parent's content like any other canvas.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useMediaQuery } from "@vueuse/core";
import { useAuthStore } from "@/stores/auth";
import { useBoxesStore } from "@/stores/boxes";
import BoxFrame from "./BoxFrame.vue";
import BoxEditor from "./BoxEditor.vue";
import HtmlBox from "./HtmlBox.vue";
import WidgetBox from "./WidgetBox.vue";
import { WIDGETS } from "./widgetRegistry";
import { TEMPLATES, imageHtml, videoHtml } from "./templates";
import { mediaSize, uploadFile } from "./upload";

const props = defineProps({
  page: { type: String, required: true },
  overlay: { type: Boolean, default: false },
  // Layout offered by "Import starter layout" on an empty canvas: a function
  // returning box inputs, or null for no button.
  starter: { type: Function, default: null },
  starterLabel: { type: String, default: "Import starter layout" },
});
const emit = defineEmits(["height"]);

const CANVAS_W = 1400;
const MOBILE_QUERY = "(max-width: 700px)";

const auth = useAuthStore();
const store = useBoxesStore(props.page);
// The standalone `me` query, not the home query: a canvas of only HTML boxes
// mounts no widget, and the landing page must not fan out to Spotify and
// Steam just to learn whether to show an Edit button.
auth.ensureLoaded();

store.fetch();

const isMobile = useMediaQuery(MOBILE_QUERY);
const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);
// adminOnly boxes come and go with the session.
watch(isAdmin, () => store.fetch());

const editing = ref(false);
const canEdit = computed(() => isAdmin.value && !isMobile.value);
const inEditMode = computed(() => editing.value && canEdit.value);

const selectedId = ref(null);
const selected = computed(() => store.boxes.find((b) => b.id === selectedId.value) ?? null);
const preview = ref(null); // unsaved HTML for the selected box, from BoxEditor
const errorMsg = ref("");
const editor = ref(null);

// ---- Scaling -------------------------------------------------------------
const viewport = ref(null);
const availableW = ref(CANVAS_W);
let observer = null;
onMounted(() => {
  observer = new ResizeObserver(([entry]) => (availableW.value = entry.contentRect.width));
  if (viewport.value) observer.observe(viewport.value);
});
onBeforeUnmount(() => observer?.disconnect());
// The viewport is only rendered on desktop, so it appears and disappears as
// the window crosses the phone breakpoint.
watch(viewport, (el, old) => {
  if (old) observer?.unobserve(old);
  if (el) observer?.observe(el);
});

const scale = computed(() => Math.min(1, availableW.value / CANVAS_W));
const contentH = computed(() => Math.max(0, ...store.boxes.map((b) => b.y + b.h)));
// Spare room below the lowest box while editing, so there is somewhere to
// drag to. An overlay has no floor of its own: empty, it should take no room.
const canvasH = computed(() => {
  const floor = props.overlay && !inEditMode.value ? 0 : 400;
  return Math.max(floor, contentH.value + (inEditMode.value ? 400 : 40));
});

watch(
  () => (props.overlay && !isMobile.value ? canvasH.value * scale.value : 0),
  (h) => emit("height", h),
  { immediate: true },
);

// ---- Mobile ordering -----------------------------------------------------
const stacked = computed(() =>
  store.boxes.filter((b) => !b.hideOnMobile).sort((a, b) => a.y - b.y || a.x - b.x),
);

// ---- Saving --------------------------------------------------------------
function report(err) {
  errorMsg.value = err?.message ?? String(err);
  setTimeout(() => (errorMsg.value = ""), 6000);
}

// The new geometry is applied locally before the request so the box stays
// where it was dropped instead of snapping back until the server answers;
// on failure it is put back.
async function commitGeometry(box, geom) {
  const before = { x: box.x, y: box.y, w: box.w, h: box.h };
  Object.assign(box, geom);
  try {
    await store.update(box.id, geom);
  } catch (err) {
    Object.assign(box, before);
    report(err);
  }
}

async function run(action) {
  try {
    await action();
  } catch (err) {
    report(err);
  }
}

function onAction(box, action) {
  if (action === "front") run(() => store.bringToFront(box));
  else if (action === "back") run(() => store.sendToBack(box));
  else if (action === "duplicate") {
    run(async () => (selectedId.value = (await store.duplicate(box)).id));
  } else if (action === "delete") {
    if (!confirm(`Delete “${labelOf(box)}”? This can't be undone from here.`)) return;
    if (selectedId.value === box.id) selectedId.value = null;
    run(() => store.remove(box.id));
  }
}

function labelOf(box) {
  const name = box.title || (box.kind === "widget" ? box.widget : "HTML");
  return box.adminOnly ? `🔒 ${name}` : name;
}

// ---- Creating ------------------------------------------------------------
const canvasEl = ref(null);

// Top-left of the part of the canvas currently on screen, so a new box
// appears where the admin is looking rather than at the top of the page.
function visibleOrigin() {
  const rect = canvasEl.value?.getBoundingClientRect();
  const top = rect ? Math.max(0, -rect.top / scale.value) : 0;
  const origin = { x: 40, y: Math.round(top / 5) * 5 + 40 };
  // Cascade, so adding several boxes in a row doesn't stack them exactly on
  // top of one another.
  while (store.boxes.some((b) => b.x === origin.x && b.y === origin.y)) {
    origin.x += 20;
    origin.y += 20;
  }
  return origin;
}

const newChoice = ref("");
function addBox() {
  const choice = newChoice.value;
  newChoice.value = "";
  if (!choice) return;
  const [kind, key] = choice.split(":");
  const { x, y } = visibleOrigin();
  let input;
  if (kind === "widget") {
    const [w, h] = WIDGETS[key].size;
    input = { kind: "widget", widget: key, w, h };
  } else {
    const t = TEMPLATES[Number(key)];
    input = {
      kind: "html",
      html: t.html,
      title: t.title,
      w: t.size[0],
      h: t.size[1],
      framed: t.framed ?? true,
      adminOnly: t.adminOnly ?? false,
    };
  }
  run(async () => {
    const created = await store.create({ ...input, x, y, z: store.topZ() });
    selectedId.value = created.id;
  });
}

const importing = ref(false);
async function importStarter() {
  importing.value = true;
  try {
    // Sequential, so the boxes come back in z order without a refetch.
    for (const box of props.starter()) await store.create(box);
  } catch (err) {
    report(err);
  } finally {
    importing.value = false;
  }
}

// ---- Dropping files -----------------------------------------------------
const DROP_MAX = 400; // longest side of a dropped picture's box, canvas px
const dragOver = ref(false);
const uploading = ref(0);

function isMedia(file) {
  return file.type.startsWith("image/") || file.type.startsWith("video/");
}

// Each file is uploaded and placed at the drop point, sized to its own
// aspect ratio so nothing is cropped, then cascaded like visibleOrigin().
async function onDrop(event) {
  dragOver.value = false;
  const files = [...(event.dataTransfer?.files ?? [])].filter(isMedia);
  if (!files.length) return;
  const rect = canvasEl.value.getBoundingClientRect();
  const snap = (n) => Math.max(0, Math.round(n / 5) * 5);
  let x = snap((event.clientX - rect.left) / scale.value);
  let y = snap((event.clientY - rect.top) / scale.value);
  uploading.value += files.length;
  for (const file of files) {
    try {
      const [url, size] = await Promise.all([uploadFile(file), mediaSize(file)]);
      const fit = size ? Math.min(1, DROP_MAX / Math.max(size.w, size.h)) : 1;
      const w = size ? Math.max(40, Math.round(size.w * fit)) : 300;
      const h = size ? Math.max(40, Math.round(size.h * fit)) : 300;
      const video = file.type.startsWith("video/");
      const created = await store.create({
        kind: "html",
        title: file.name,
        html: video ? videoHtml(url) : imageHtml(url),
        framed: false,
        x,
        y,
        w,
        h,
        z: store.topZ(),
      });
      selectedId.value = created.id;
      x += 20;
      y += 20;
    } catch (err) {
      report(err);
    } finally {
      uploading.value--;
    }
  }
}

// ---- Selection -----------------------------------------------------------
function select(box) {
  if (selectedId.value === box.id) return;
  // Leaving a box with unsaved HTML goes through the editor's own confirm.
  if (editor.value && selected.value) {
    editor.value.close();
    if (selectedId.value !== null) return; // they chose to keep editing
  }
  selectedId.value = box.id;
}

function toggleEdit() {
  if (editing.value && selected.value) {
    editor.value?.close();
    if (selectedId.value !== null) return;
  }
  editing.value = !editing.value;
}
</script>

<template>
  <div
    class="box-canvas"
    :class="{
      'with-editor': inEditMode && selected && !overlay,
      overlay: overlay && !isMobile,
      'overlay-editing': overlay && inEditMode,
    }"
  >
    <div v-if="canEdit" class="admin-bar">
      <button type="button" class="bar-btn" :class="{ on: editing }" @click="toggleEdit">
        {{ editing ? "Done" : overlay ? "Edit notes" : "Edit layout" }}
      </button>
      <template v-if="inEditMode">
        <select v-model="newChoice" class="bar-btn" @change="addBox">
          <option value="" disabled>+ New box…</option>
          <optgroup label="HTML">
            <option v-for="(t, i) in TEMPLATES" :key="t.label" :value="`html:${i}`">{{ t.label }}</option>
          </optgroup>
          <optgroup label="Widget">
            <option v-for="name in Object.keys(WIDGETS)" :key="name" :value="`widget:${name}`">{{ name }}</option>
          </optgroup>
        </select>
        <button
          v-if="starter && store.loaded && store.boxes.length === 0"
          type="button"
          class="bar-btn"
          :disabled="importing"
          @click="importStarter"
        >
          {{ importing ? "Importing…" : starterLabel }}
        </button>
      </template>
      <span v-if="inEditMode" class="hint">
        {{ uploading ? `Uploading ${uploading}…` : "Drop images or videos onto the canvas to add them" }}
      </span>
      <slot name="admin-bar" :editing="inEditMode" />
      <span v-if="errorMsg" class="error">{{ errorMsg }}</span>
    </div>

    <p v-if="store.error" class="state">Couldn't load the page layout.</p>
    <p v-else-if="!overlay && store.loaded && store.boxes.length === 0 && !canEdit" class="state">
      Nothing here yet.
    </p>

    <!-- Phone: one column, no editing -->
    <div v-if="isMobile" class="stack">
      <div
        v-for="box in stacked"
        :key="box.id"
        class="stack-item"
        :class="{ framed: box.framed, 'admin-only': box.adminOnly }"
        :style="{ height: `${box.h}px` }"
      >
        <WidgetBox v-if="box.kind === 'widget'" :name="box.widget" />
        <HtmlBox v-else :html="box.html" :title="box.title" />
      </div>
    </div>

    <!-- Desktop: the scaled canvas -->
    <div v-else ref="viewport" class="viewport">
      <div
        class="canvas-size"
        :style="{ width: `${CANVAS_W * scale}px`, height: `${canvasH * scale}px` }"
      >
        <div
          ref="canvasEl"
          class="canvas"
          :class="{ editing: inEditMode, 'drag-over': dragOver }"
          :style="{ width: `${CANVAS_W}px`, height: `${canvasH}px`, transform: `scale(${scale})` }"
          @pointerdown.self="selectedId = null"
          @dragover.prevent="dragOver = inEditMode"
          @dragleave.self="dragOver = false"
          @drop.prevent="inEditMode && onDrop($event)"
        >
          <BoxFrame
            v-for="box in store.boxes"
            :key="box.id"
            :box="box"
            :scale="scale"
            :editing="inEditMode"
            :selected="inEditMode && box.id === selectedId"
            :label="labelOf(box)"
            @select="select(box)"
            @commit="commitGeometry(box, $event)"
            @action="onAction(box, $event)"
          >
            <WidgetBox v-if="box.kind === 'widget'" :name="box.widget" />
            <HtmlBox
              v-else
              :html="box.id === selectedId && preview !== null ? preview : box.html"
              :title="box.title"
            />
          </BoxFrame>
        </div>
      </div>
    </div>

    <BoxEditor
      v-if="inEditMode && selected"
      ref="editor"
      :box="selected"
      @preview="preview = $event"
      @close="selectedId = null"
      @error="report"
    />
  </div>
</template>

<style scoped>
.box-canvas {
  width: 100%;
  box-sizing: border-box;
}
/* Leave room for the fixed editor panel so the canvas rescales beside it
   instead of disappearing under it. Not for an overlay: rescaling it would
   pull its boxes off the content they are pinned over, so the panel just
   covers the right of the page instead. */
.box-canvas.with-editor {
  padding-right: calc(min(460px, 100vw) + 16px);
}

/* Over the parent's content: only the boxes take clicks, so the page
   underneath stays usable. In edit mode the whole area does, so a click on
   empty canvas deselects rather than following a link underneath. */
.box-canvas.overlay {
  position: absolute;
  inset: 0 0 auto 0;
  z-index: 20;
  pointer-events: none;
}
.box-canvas.overlay :deep(.box),
.box-canvas.overlay .admin-bar,
.box-canvas.overlay-editing .canvas {
  pointer-events: auto;
}
/* An overlay's bar must not push the canvas down off the content it is
   aligned with, so it floats in the corner of the window — above the
   layout's footer bar. */
.box-canvas.overlay .admin-bar {
  position: fixed;
  left: 16px;
  bottom: 56px;
  z-index: 70;
  margin: 0;
}

.admin-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin: 0 auto 10px;
  max-width: 1400px;
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
.bar-btn:hover,
.bar-btn.on {
  border-color: var(--color-primary);
}
.bar-btn:disabled {
  opacity: 0.5;
}
.error {
  color: var(--color-tertiary);
}
.hint {
  color: var(--color-muted);
  font-size: 0.85rem;
}
.state {
  text-align: center;
  color: var(--color-muted);
}

.viewport {
  width: 100%;
}
.canvas-size {
  position: relative;
  margin: 0 auto;
}
.canvas {
  position: absolute;
  left: 0;
  top: 0;
  transform-origin: 0 0;
}
/* 20px dot grid while editing, to line things up by eye. */
.canvas.editing {
  outline: 1px dashed var(--color-quaternary);
  background-image: radial-gradient(var(--color-quaternary) 1px, transparent 1px);
  background-size: 20px 20px;
}
.canvas.drag-over {
  outline: 3px dashed var(--color-primary);
}

.stack {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.stack-item {
  box-sizing: border-box;
  width: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.stack-item.framed {
  background-color: var(--color-surface);
  border: 2px solid var(--color-quaternary);
}
.stack-item.admin-only {
  outline: 2px dashed var(--color-tertiary);
  outline-offset: -2px;
}
</style>
