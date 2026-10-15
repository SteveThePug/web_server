import { onUnmounted, ref, toValue } from "vue";

const SNAP = 5;
const MIN_SIZE = 40;

/** Owns pointer lifecycle and optimistic geometry while a box is manipulated. */
export function useBoxInteraction({ box, scale, emit }) {
  const draft = ref(null);
  let drag = null;

  const snap = (n) => Math.round(n / SNAP) * SNAP;

  function start(mode, event) {
    if (event.button !== 0) return;
    emit("select");
    event.preventDefault();
    event.target.setPointerCapture(event.pointerId);
    const { x, y, w, h } = toValue(box);
    drag = { mode, startX: event.clientX, startY: event.clientY, from: { x, y, w, h }, moved: false };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", end);
    window.addEventListener("pointercancel", end);
  }

  function move(event) {
    if (!drag) return;
    const factor = toValue(scale);
    const dx = (event.clientX - drag.startX) / factor;
    const dy = (event.clientY - drag.startY) / factor;
    if (!drag.moved && Math.hypot(dx, dy) < 3) return;
    drag.moved = true;

    const { x, y, w, h } = drag.from;
    const next = { x, y, w, h };
    if (drag.mode === "move") {
      next.x = Math.max(0, snap(x + dx));
      next.y = Math.max(0, snap(y + dy));
    } else {
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
    if (!drag) return;
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", end);
    window.removeEventListener("pointercancel", end);
    if (drag.moved && draft.value) emit("commit", { ...draft.value });
    drag = null;
    draft.value = null;
  }

  onUnmounted(end);
  return { draft, start };
}
