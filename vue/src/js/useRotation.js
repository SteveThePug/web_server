// Rotate through a list on a timer, as the Listening and Steam widgets do.
//
// A self-rearming setTimeout rather than setInterval, so clicking to advance
// manually restarts the full delay instead of leaving whatever was left of the
// current tick. An optional periodic refresh callback covers the "re-poll the
// source every few minutes" half of the same pattern. Both timers are cleared
// on unmount.

import { ref, unref, onMounted, onUnmounted } from "vue";

/**
 * @param {import("vue").Ref<Array>|(() => Array)} items  the list being rotated
 * @param {number} ms  delay between rotations
 * @param {object} [options]
 * @param {() => void} [options.refresh]     called every refreshMs
 * @param {number} [options.refreshMs]       refresh interval
 * @returns {{ idx: import("vue").Ref<number>, next: () => void }}
 */
export function useRotation(items, ms, { refresh, refreshMs } = {}) {
  const idx = ref(0);
  let nextId = null;
  let refreshId = null;

  function list() {
    return (typeof items === "function" ? items() : unref(items)) ?? [];
  }

  /** Advance one step and re-arm. Leaves the index alone on an empty list. */
  function next() {
    clearTimeout(nextId);
    nextId = setTimeout(next, ms);
    const count = list().length;
    if (count) idx.value = (idx.value + 1) % count;
  }

  onMounted(() => {
    nextId = setTimeout(next, ms);
    if (refresh && refreshMs) refreshId = setInterval(refresh, refreshMs);
  });

  onUnmounted(() => {
    clearTimeout(nextId);
    clearInterval(refreshId);
  });

  return { idx, next };
}
