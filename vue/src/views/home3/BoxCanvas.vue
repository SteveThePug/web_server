<script setup>
/**
 * A free-form canvas of boxes the admin arranges in the browser, for one page
 * key (see stores/boxes.js). Home3.vue (/stp), the landing page and every
 * admin-created page (views/canvas/CanvasPage.vue) are this component.
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
 * The canvas also owns its page's background (stores/background.js), which
 * DefaultLayout paints behind the whole window while the canvas is mounted;
 * admins change it with the Background button.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useMediaQuery } from "@vueuse/core";
import { useAuthStore } from "@/stores/auth";
import { useBoxesStore } from "@/stores/boxes";
import { useBackgroundStore } from "@/stores/background";
import BackgroundEditor from "./BackgroundEditor.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";
import BoxFrame from "./BoxFrame.vue";
import BoxEditor from "./BoxEditor.vue";
import HtmlBox from "./HtmlBox.vue";
import WidgetBox from "./WidgetBox.vue";
import { WIDGETS } from "./widgetRegistry";
import { TEMPLATES, imageHtml, videoHtml } from "./templates";
import { mediaSize, uploadFile } from "./upload";

const props = defineProps({
  page: { type: String, required: true },
  // Layout offered by the "Import starter layout" button in edit mode: a
  // function returning box inputs, or null for no button.
  starter: { type: Function, default: null },
  starterLabel: { type: String, default: "Import starter layout" },
});

const CANVAS_W = 1400;
const MOBILE_QUERY = "(max-width: 700px)";

const auth = useAuthStore();
const store = useBoxesStore(props.page);
// The standalone `me` query, not the home query: a canvas of only HTML boxes
// mounts no widget, and the landing page must not fan out to Spotify and
// Steam just to learn whether to show an Edit button.
auth.ensureLoaded();

store.fetch();

const background = useBackgroundStore();
background.show(props.page);
onBeforeUnmount(() => background.hide(props.page));
const backgroundOpen = ref(false);
const confirmation = ref(null);

const isMobile = useMediaQuery(MOBILE_QUERY);
const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);
// adminOnly boxes come and go with the session, and so does an adminOnly
// page's background.
watch(isAdmin, () => {
  store.fetch();
  background.load(props.page);
});

const editing = ref(false);
const canEdit = computed(() => isAdmin.value && !isMobile.value);
const inEditMode = computed(() => editing.value && canEdit.value);

const selectedId = ref(null);
const selected = computed(() => store.boxes.find((b) => b.id === selectedId.value) ?? null);
const preview = ref(null); // unsaved HTML for the selected box, from BoxEditor
const errorMsg = ref("");

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
// drag to.
const canvasH = computed(() => Math.max(400, contentH.value + (inEditMode.value ? 400 : 40)));

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
    askConfirmation({
      title: "Delete box",
      message: `Delete “${labelOf(box)}”? This can't be undone from here.`,
      confirmLabel: "Delete",
    }).then((confirmed) => {
      if (!confirmed) return;
      if (selectedId.value === box.id) selectedId.value = null;
      run(() => store.remove(box.id));
    });
  }
}

function askConfirmation(request) {
  return new Promise((resolve) => {
    confirmation.value = { ...request, open: true, resolve };
  });
}

function finishConfirmation(confirmed) {
  const request = confirmation.value;
  confirmation.value = null;
  request?.resolve(confirmed);
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

// Offered on a canvas that already has boxes too — the landing page had notes
// pinned over its old content before that content became boxes — but then
// it asks first, since importing twice doubles every box.
const importing = ref(false);
async function importStarter() {
  const count = store.boxes.length;
  if (count) {
    const confirmed = await askConfirmation({
      title: "Import starter layout",
      message: `This page already has ${count} box${count === 1 ? "" : "es"}. Add the starter layout on top?`,
      confirmLabel: "Import",
    });
    if (!confirmed) return;
  }
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
// Leaving a box saves its unsaved HTML; BoxEditor does that itself as the
// selection changes or it unmounts.
function select(box) {
  backgroundOpen.value = false;
  selectedId.value = box.id;
}

function toggleBackground() {
  backgroundOpen.value = !backgroundOpen.value;
  if (backgroundOpen.value) selectedId.value = null;
}

function toggleEdit() {
  editing.value = !editing.value;
  if (!editing.value) selectedId.value = null;
}
</script>

<template>
  <div
    class="box-canvas"
    :class="{ 'with-editor': inEditMode && selected }"
  >
    <div v-if="canEdit" class="admin-bar">
      <button type="button" class="bar-btn" :class="{ on: editing }" @click="toggleEdit">
        {{ editing ? "Done" : "Edit layout" }}
      </button>
       <button type="button" class="bar-btn" :class="{ on: backgroundOpen }" @click="toggleBackground">
        Background
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
          v-if="starter && store.loaded"
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
    <p v-else-if="store.loaded && store.boxes.length === 0 && !canEdit" class="state">
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

    <BackgroundEditor v-if="backgroundOpen && canEdit" :page="page" @close="backgroundOpen = false" />

    <BoxEditor
      v-if="inEditMode && selected"
      :box="selected"
      @preview="preview = $event"
      @close="selectedId = null"
      @error="report"
    />

    <ConfirmDialog
      v-if="confirmation"
      v-model="confirmation.open"
      :title="confirmation.title"
      :message="confirmation.message"
      :confirm-label="confirmation.confirmLabel"
      @confirm="finishConfirmation(true)"
      @cancel="finishConfirmation(false)"
    />
  </div>
</template>

<style scoped>
.box-canvas {
  width: 100%;
  box-sizing: border-box;
}
/* Leave room for the fixed editor panel so the canvas rescales beside it
   instead of disappearing under it. */
.box-canvas.with-editor {
  padding-right: calc(min(460px, 100vw) + 16px);
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
