import { onUnmounted, watch } from "vue";

const entries = [];
let listening = false;

function onKeydown(event) {
  if (event.key !== "Escape") return;
  entries.at(-1)?.close();
}

function updateListener() {
  if (entries.length && !listening) {
    window.addEventListener("keydown", onKeydown);
    listening = true;
  } else if (!entries.length && listening) {
    window.removeEventListener("keydown", onKeydown);
    listening = false;
  }
}

/** Registers one open overlay so Escape always closes only the topmost one. */
export function useOverlayStack(open, close) {
  let entry = null;

  function register() {
    if (entry) return;
    entry = { close };
    entries.push(entry);
    updateListener();
  }

  function unregister() {
    if (!entry) return;
    const index = entries.indexOf(entry);
    if (index !== -1) entries.splice(index, 1);
    entry = null;
    updateListener();
  }

  watch(open, (isOpen) => (isOpen ? register() : unregister()), { immediate: true });
  onUnmounted(unregister);
}
