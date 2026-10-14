<script setup>
/**
 * One positioned box on the canvas, plus its edit-mode drag/resize overlay.
 *
 * Geometry is canvas px (see models.Box). The canvas is drawn with
 * transform: scale(), so pointer deltas, which are screen px, are divided by
 * `scale` before they are applied.
 *
 * In edit mode an overlay covers the whole box. That is what makes dragging
 * work at all over an HTML box: pointer events that land on an iframe go to
 * the iframe's document and never reach this one. It also means the box
 * contents can't be clicked while editing, which is intended — leave edit mode
 * to use them.
 *
 * While a drag is in progress the box draws from a local `draft`. On release
 * the final geometry is emitted once as `commit`; the parent saves it. A press
 * that never moves is a click (select only) and commits nothing.
 */
import { computed, ref } from "vue";

const props = defineProps({
  box: { type: Object, required: true },
  scale: { type: Number, default: 1 },
  editing: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  label: { type: String, default: "" },
});

const emit = defineEmits(["select", "commit", "action"]);

const SNAP = 5;
const MIN_SIZE = 40; // matches minBoxSize in backend/graph/box_helpers.go
const HANDLES = ["n", "s", "e", "w", "ne", "nw", "se", "sw"];

const draft = ref(null);
const geom = computed(() => draft.value ?? props.box);

const style = computed(() => ({
  left: `${geom.value.x}px`,
  top: `${geom.value.y}px`,
  width: `${geom.value.w}px`,
  height: `${geom.value.h}px`,
  // The selected box is lifted above everything while editing: its resize
  // handles sit just outside its edges, where an overlapping neighbour with a
  // higher z would otherwise cover them and take the press.
  zIndex: props.selected ? 100000 : props.box.z,
}));

const snap = (n) => Math.round(n / SNAP) * SNAP;

let drag = null;

function start(mode, event) {
  if (event.button !== 0) return;
  emit("select");
  event.preventDefault();
  event.target.setPointerCapture(event.pointerId);
  const { x, y, w, h } = props.box;
  drag = { mode, startX: event.clientX, startY: event.clientY, from: { x, y, w, h }, moved: false };
  window.addEventListener("pointermove", move);
  window.addEventListener("pointerup", end);
  window.addEventListener("pointercancel", end);
}

function move(event) {
  const dx = (event.clientX - drag.startX) / props.scale;
  const dy = (event.clientY - drag.startY) / props.scale;
  // A few px of jitter on a click shouldn't count as a move.
  if (!drag.moved && Math.hypot(dx, dy) < 3) return;
  drag.moved = true;

  const { x, y, w, h } = drag.from;
  const next = { x, y, w, h };
  if (drag.mode === "move") {
    next.x = Math.max(0, snap(x + dx));
    next.y = Math.max(0, snap(y + dy));
  } else {
    // Each handle moves the edges named in it. Moving the left or top edge
    // changes position and size together, and is clamped so the opposite
    // edge stays put once the box reaches its minimum size.
    if (drag.mode.includes("e")) next.w = Math.max(MIN_SIZE, snap(w + dx));
    if (drag.mode.includes("s")) next.h = Math.max(MIN_SIZE, snap(h + dy));
    if (drag.mode.includes("w")) {
      next.x = Math.max(0, Math.min(snap(x + dx), x + w - MIN_SIZE));
      next.w = x + w - next.x;
    }
    if (drag.mode.includes("n")) {
      next.y = Math.max(0, Math.min(snap(y + dy), y + h - MIN_SIZE));
      next.h = y + h - next.y;
    }
  }
  draft.value = next;
}

function end() {
  window.removeEventListener("pointermove", move);
  window.removeEventListener("pointerup", end);
  window.removeEventListener("pointercancel", end);
  if (drag?.moved && draft.value) emit("commit", { ...draft.value });
  drag = null;
  draft.value = null;
}
</script>

<template>
  <div
    class="box"
    :class="{ framed: box.framed, editing, selected }"
    :style="style"
  >
    <slot />
    <div
      v-if="editing"
      class="overlay"
      @pointerdown="start('move', $event)"
    >
      <span class="label">{{ label }}</span>
      <div v-if="selected" class="toolbar" @pointerdown.stop>
        <button type="button" title="Bring to front" @click="emit('action', 'front')">▲</button>
        <button type="button" title="Send to back" @click="emit('action', 'back')">▼</button>
        <button type="button" title="Duplicate" @click="emit('action', 'duplicate')">⧉</button>
        <button type="button" title="Delete" @click="emit('action', 'delete')">✕</button>
      </div>
      <template v-if="selected">
        <div
          v-for="h in HANDLES"
          :key="h"
          class="handle"
          :class="`h-${h}`"
          @pointerdown.stop="start(h, $event)"
        />
      </template>
    </div>
  </div>
</template>

<style scoped>
.box {
  position: absolute;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.box.framed {
  background-color: var(--color-surface);
  border: 2px solid var(--color-quaternary);
}

.overlay {
  position: absolute;
  inset: 0;
  cursor: move;
  touch-action: none;
  outline: 1px dashed var(--color-quaternary);
  background: color-mix(in srgb, var(--color-surface-deep) 15%, transparent);
}
.box.selected .overlay {
  outline: 2px solid var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 6%, transparent);
}

.label {
  position: absolute;
  left: 0;
  top: 0;
  padding: 0 4px;
  font-size: 12px;
  color: var(--color-surface-deep);
  background: var(--color-primary);
  pointer-events: none;
  opacity: 0.85;
}

.toolbar {
  position: absolute;
  right: 0;
  top: -28px;
  display: flex;
  gap: 2px;
  cursor: default;
}
.toolbar button {
  width: 26px;
  height: 24px;
  font-size: 13px;
  line-height: 1;
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-primary);
  cursor: pointer;
}
.toolbar button:hover {
  color: var(--color-tertiary);
  border-color: var(--color-tertiary);
}

.handle {
  position: absolute;
  width: 10px;
  height: 10px;
  background: var(--color-primary);
  border: 1px solid var(--color-surface-deep);
  touch-action: none;
}
.h-n,
.h-s {
  left: calc(50% - 5px);
  cursor: ns-resize;
}
.h-e,
.h-w {
  top: calc(50% - 5px);
  cursor: ew-resize;
}
.h-n,
.h-ne,
.h-nw {
  top: -5px;
}
.h-s,
.h-se,
.h-sw {
  bottom: -5px;
}
.h-e,
.h-ne,
.h-se {
  right: -5px;
}
.h-w,
.h-nw,
.h-sw {
  left: -5px;
}
.h-ne,
.h-sw {
  cursor: nesw-resize;
}
.h-nw,
.h-se {
  cursor: nwse-resize;
}
</style>
