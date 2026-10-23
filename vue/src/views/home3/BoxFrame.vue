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
 * that never moves is a click (select only) and commits nothing. Drags snap
 * to the lines `targets` returns (see ./alignment.js), and the guides for
 * what the box currently lines up with are emitted as `guides` for the canvas
 * to draw, since a guide runs outside this box.
 */
import { computed, toRef } from "vue";
import { useBoxInteraction } from "./useBoxInteraction";

const props = defineProps({
  box: { type: Object, required: true },
  scale: { type: Number, default: 1 },
  editing: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  label: { type: String, default: "" },
  // (box) => alignment.snapTargets(...) for that box, or null for the grid alone.
  targets: { type: Function, default: null },
});

const emit = defineEmits(["select", "commit", "action", "guides"]);

const HANDLES = ["n", "s", "e", "w", "ne", "nw", "se", "sw"];

const { draft, start } = useBoxInteraction({
  box: toRef(props, "box"),
  scale: toRef(props, "scale"),
  targets: () => props.targets?.(props.box),
  emit,
});
const geom = computed(() => draft.value ?? props.box);

function nudge(mode, event) {
  if (!props.selected || !["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight"].includes(event.key)) return;
  event.preventDefault();
  const step = event.shiftKey ? 10 : 5;
  const next = { x: props.box.x, y: props.box.y, w: props.box.w, h: props.box.h };
  if (mode === "move") {
    if (event.key === "ArrowLeft") next.x -= step;
    if (event.key === "ArrowRight") next.x += step;
    if (event.key === "ArrowUp") next.y -= step;
    if (event.key === "ArrowDown") next.y += step;
  } else {
    if (mode.includes("w") && event.key === "ArrowLeft") {
      next.x -= step;
      next.w += step;
    }
    if (mode.includes("w") && event.key === "ArrowRight") {
      next.x += step;
      next.w = Math.max(40, next.w - step);
    }
    if (mode.includes("e") && event.key === "ArrowRight") next.w += step;
    if (mode.includes("e") && event.key === "ArrowLeft") next.w = Math.max(40, next.w - step);
    if (mode.includes("n") && event.key === "ArrowUp") {
      next.y -= step;
      next.h += step;
    }
    if (mode.includes("n") && event.key === "ArrowDown") {
      next.y += step;
      next.h = Math.max(40, next.h - step);
    }
    if (mode.includes("s") && event.key === "ArrowDown") next.h += step;
    if (mode.includes("s") && event.key === "ArrowUp") next.h = Math.max(40, next.h - step);
  }
  next.x = Math.max(0, next.x);
  next.y = Math.max(0, next.y);
  emit("commit", next);
}

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

</script>

<template>
  <div
    class="box"
    :class="{ framed: box.framed, editing, selected, 'admin-only': box.adminOnly }"
    :style="style"
  >
    <slot />
    <div
      v-if="editing"
      class="overlay"
      :tabindex="selected ? 0 : -1"
      @pointerdown="start('move', $event)"
      @keydown="nudge('move', $event)"
    >
      <span class="label">{{ label }}</span>
      <div v-if="selected" class="toolbar" @pointerdown.stop>
        <button type="button" title="Bring to front" aria-label="Bring to front" @click="emit('action', 'front')">▲</button>
        <button type="button" title="Send to back" aria-label="Send to back" @click="emit('action', 'back')">▼</button>
        <button type="button" title="Duplicate" aria-label="Duplicate" @click="emit('action', 'duplicate')">⧉</button>
        <button type="button" title="Delete" aria-label="Delete" @click="emit('action', 'delete')">✕</button>
      </div>
      <template v-if="selected">
        <div
          v-for="h in HANDLES"
          :key="h"
          class="handle"
          :class="`h-${h}`"
          :tabindex="selected ? 0 : -1"
          role="slider"
          :aria-label="`Resize ${h}`"
          @pointerdown.stop="start(h, $event)"
          @keydown.stop="nudge(h, $event)"
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
/* Only admins are ever sent an adminOnly box; the outline reminds them that
   nobody else can see it. */
.box.admin-only:not(.editing) {
  outline: 2px dashed var(--color-tertiary);
  outline-offset: -2px;
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
