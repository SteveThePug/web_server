import { onUnmounted, ref, toValue } from "vue";
import { NO_TARGETS, SNAP_PX, guidesFor, place } from "./alignment";

/**
 * Owns pointer lifecycle and optimistic geometry while a box is manipulated.
 *
 * `targets` returns what the box can line up with (alignment.snapTargets), or
 * null when snapping is off; it is asked once per drag, since nothing else
 * moves meanwhile. The guides for each position are emitted as `guides`, and
 * an empty list on release. Holding Ctrl/⌘ places the box on the grid alone.
 */
export function useBoxInteraction({ box, scale, targets, emit }) {
  const draft = ref(null);
  let drag = null;

  function start(mode, event) {
    if (event.button !== 0) return;
    emit("select");
    event.preventDefault();
    event.target.setPointerCapture(event.pointerId);
    const { x, y, w, h } = toValue(box);
    drag = {
      mode,
      startX: event.clientX,
      startY: event.clientY,
      from: { x, y, w, h },
      targets: targets?.() ?? NO_TARGETS,
      moved: false,
    };
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

    const free = event.ctrlKey || event.metaKey;
    const lines = free ? NO_TARGETS : drag.targets;
    const next = place(drag.from, drag.mode, dx, dy, lines, SNAP_PX / factor);
    draft.value = next;
    emit("guides", guidesFor(next, drag.mode, lines));
  }

  function end() {
    if (!drag) return;
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", end);
    window.removeEventListener("pointercancel", end);
    if (drag.moved) emit("guides", []);
    if (drag.moved && draft.value) emit("commit", { ...draft.value });
    drag = null;
    draft.value = null;
  }

  onUnmounted(end);
  return { draft, start };
}
