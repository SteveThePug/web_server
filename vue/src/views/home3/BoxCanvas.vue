<script setup>
/**
 * A free-form canvas of boxes the admin arranges in the browser, for one page
 * key (see stores/boxes.js). The landing page and every admin-created page
 * (views/canvas/CanvasPage.vue, /stp included) are this component.
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
 * changes save on release, and the selected box opens in BoxEditor. A dragged
 * box snaps its edges and centre to other boxes' and to the canvas's sides
 * and centre, drawing a guide along each line it shares (./alignment.js); the
 * Align button turns that off, and Ctrl/⌘ while dragging skips it once. Boxes
 * marked adminOnly are never sent to anyone else (the server filters them);
 * admins see them outlined. Image and video files dropped onto the canvas in
 * edit mode are uploaded and each becomes a box where it was dropped; "+ From
 * uploads" does the same with a file that is already uploaded.
 *
 * The canvas also owns its page's background (stores/background.js), which
 * DefaultLayout paints behind the whole window while the canvas is mounted;
 * admins change it with the Background button.
 *
 * The Layouts button saves the page as a named layout, or loads a saved one
 * onto it (LayoutPanel.vue, stores/layouts.js) — which is how a layout moves
 * from one page to another.
 *
 * "+ New box" offers the saved components (stores/components.js) as starters,
 * each placed as an ordinary box with a copy of its html, plus the widgets.
 * The Components button (ComponentPanel.vue) places them linked as well: a
 * linked box shows its component's html, so editing it in BoxEditor edits
 * every box linked to that component. Linked boxes are labelled with ⛓.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useMediaQuery } from "@vueuse/core";
import { useAuthStore } from "@/stores/auth";
import { useBoxesStore } from "@/stores/boxes";
import { useBackgroundStore } from "@/stores/background";
import { useComponentsStore } from "@/stores/components";
import BackgroundEditor from "./BackgroundEditor.vue";
import LayoutPanel from "./LayoutPanel.vue";
import ComponentPanel from "./ComponentPanel.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";
import UploadPicker from "@/components/util/UploadPicker.vue";
import BoxFrame from "./BoxFrame.vue";
import BoxEditor from "./BoxEditor.vue";
import { panelWidth, snapping } from "./editorPrefs";
import { snapTargets } from "./alignment";
import HtmlBox from "./HtmlBox.vue";
import WidgetBox from "./WidgetBox.vue";
import { WIDGETS } from "./widgetRegistry";
import { imageHtml, videoHtml } from "./templates";
import { isVideo, mediaSize, uploadFile } from "@/uploads";

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
// Which of the left-hand panels is open: "background", "layouts",
// "components" or null. They sit in the same place, so only one at a time.
const panel = ref(null);
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

// Admin only, so fetched on first entering edit mode, for "+ New box".
const components = useComponentsStore();
watch(inEditMode, (on) => {
  if (on && !components.loaded && !components.loading) components.fetch();
});

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

// ---- Alignment -----------------------------------------------------------
// What a dragged box can line up with: every other box, and the canvas. Only
// the dragged box reports guides, so one list serves the whole canvas. One
// function for every BoxFrame, so redrawing the guides on each pointermove
// doesn't hand every frame a new prop and redraw it too.
const guides = ref([]);
function targetsFor(box) {
  return snapping.value ? snapTargets(store.boxes.filter((b) => b.id !== box.id), CANVAS_W) : null;
}

// A guide is a 1-screen-px line, so its thickness undoes the canvas scale.
function guideStyle(g) {
  const thickness = `${1 / scale.value}px`;
  return g.axis === "x"
    ? { left: `${g.at}px`, top: `${g.from}px`, width: thickness, height: `${g.to - g.from}px` }
    : { top: `${g.at}px`, left: `${g.from}px`, height: thickness, width: `${g.to - g.from}px` };
}

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
    run(async () => {
      selectedId.value = (await store.duplicate(box)).id;
      if (box.componentId) components.countUse(box.componentId, 1);
    });
  } else if (action === "delete") {
    askConfirmation({
      title: "Delete box",
      message: `Delete “${labelOf(box)}”? This can't be undone from here.`,
      confirmLabel: "Delete",
    }).then((confirmed) => {
      if (!confirmed) return;
      if (selectedId.value === box.id) selectedId.value = null;
      run(async () => {
        await store.remove(box.id);
        if (box.componentId) components.countUse(box.componentId, -1);
      });
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
  let name = box.title || (box.kind === "widget" ? box.widget : "HTML");
  if (box.componentId) name = `⛓ ${name}`;
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
  if (choice === "import") {
    run(() => components.importBuiltins());
    return;
  }
  const [kind, key] = choice.split(":");
  if (kind === "component") {
    const component = components.components.find((c) => c.id === key);
    if (component) placeComponent(component, false);
    return;
  }
  const { x, y } = visibleOrigin();
  const [w, h] = WIDGETS[key].size;
  run(async () => {
    const created = await store.create({ kind: "widget", widget: key, w, h, x, y, z: store.topZ() });
    selectedId.value = created.id;
  });
}

// From "+ New box" (always a copy) or the Components panel (either). The
// panel is closed, like selecting a box closes it, since the new box opens in
// the editor and lands where the panel would cover it.
function placeComponent(component, linked) {
  panel.value = null;
  run(async () => {
    const created = await components.place(component, props.page, { ...visibleOrigin(), z: store.topZ() }, linked);
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

// A box showing one stored image or video at (x, y), sized to the file's own
// aspect ratio so nothing is cropped. Selected once made.
async function placeMedia(url, size, title, x, y) {
  const fit = size ? Math.min(1, DROP_MAX / Math.max(size.w, size.h)) : 1;
  const created = await store.create({
    kind: "html",
    title,
    html: isVideo(url) ? videoHtml(url) : imageHtml(url),
    framed: false,
    x,
    y,
    w: size ? Math.max(40, Math.round(size.w * fit)) : 300,
    h: size ? Math.max(40, Math.round(size.h * fit)) : 300,
    z: store.topZ(),
  });
  selectedId.value = created.id;
}

// Each file is uploaded and placed at the drop point, then cascaded like
// visibleOrigin().
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
      await placeMedia(url, size, file.name, x, y);
      x += 20;
      y += 20;
    } catch (err) {
      report(err);
    } finally {
      uploading.value--;
    }
  }
}

// A file picked from what is already uploaded goes where a new box would.
const picking = ref(false);
async function placePicked(url) {
  const { x, y } = visibleOrigin();
  try {
    await placeMedia(url, await mediaSize(url), url.split("/").pop(), x, y);
  } catch (err) {
    report(err);
  }
}

// ---- Selection -----------------------------------------------------------
// Leaving a box saves its unsaved HTML; BoxEditor does that itself as the
// selection changes or it unmounts.
function select(box) {
  panel.value = null;
  selectedId.value = box.id;
}

// Opening a panel closes the others and the box editor.
function togglePanel(name) {
  panel.value = panel.value === name ? null : name;
  if (panel.value) selectedId.value = null;
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
    :style="{ '--editor-width': `${panelWidth}px` }"
  >
    <div v-if="canEdit" class="admin-bar">
      <button type="button" class="bar-btn" :class="{ on: editing }" @click="toggleEdit">
        {{ editing ? "Done" : "Edit layout" }}
      </button>
      <button type="button" class="bar-btn" :class="{ on: panel === 'background' }" @click="togglePanel('background')">
        Background
      </button>
      <button type="button" class="bar-btn" :class="{ on: panel === 'layouts' }" @click="togglePanel('layouts')">
        Layouts
      </button>
      <button type="button" class="bar-btn" :class="{ on: panel === 'components' }" @click="togglePanel('components')">
        Components
      </button>
      <template v-if="inEditMode">
        <select v-model="newChoice" class="bar-btn" @change="addBox">
          <option value="" disabled>+ New box…</option>
          <optgroup label="HTML">
            <option v-if="!components.loaded" value="" disabled>Loading…</option>
            <option v-else-if="!components.components.length" value="import">Import built-in templates</option>
            <option v-for="c in components.components" :key="c.id" :value="`component:${c.id}`">{{ c.name }}</option>
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
        <button type="button" class="bar-btn" @click="picking = true">+ From uploads…</button>
        <button
          type="button"
          class="bar-btn"
          :class="{ on: snapping }"
          :aria-pressed="snapping"
          title="Snap dragged boxes to other boxes' edges and centres (hold Ctrl/⌘ to skip once)"
          @click="snapping = !snapping"
        >
          Align: {{ snapping ? "on" : "off" }}
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
            :targets="targetsFor"
            @select="select(box)"
            @guides="guides = $event"
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
          <div v-for="g in guides" :key="`${g.axis}${g.at}`" class="guide" :style="guideStyle(g)" />
        </div>
      </div>
    </div>

    <BackgroundEditor v-if="panel === 'background' && canEdit" :page="page" @close="panel = null" />
    <LayoutPanel v-if="panel === 'layouts' && canEdit" :page="page" @close="panel = null" />
    <ComponentPanel v-if="panel === 'components' && canEdit" @close="panel = null" @place="placeComponent" />

    <BoxEditor
      v-if="inEditMode && selected"
      :box="selected"
      @preview="preview = $event"
      @close="selectedId = null"
      @error="report"
    />

    <UploadPicker v-if="picking" v-model="picking" :kinds="['image', 'video']" @pick="placePicked" />

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
   instead of disappearing under it. The width is the admin's chosen one
   (./editorPrefs), so the canvas follows the panel's edge as it is dragged. */
.box-canvas.with-editor {
  padding-right: calc(min(var(--editor-width), 100vw) + 16px);
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
/* Above every box, the lifted selected one (BoxFrame) included. */
.guide {
  position: absolute;
  z-index: 100001;
  background: var(--color-tertiary);
  pointer-events: none;
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
