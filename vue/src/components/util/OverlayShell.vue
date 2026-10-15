<script setup>
import { nextTick, onUnmounted, ref, watch } from "vue";
import { useOverlayStack } from "@/js/useOverlayStack";

const open = defineModel({ type: Boolean, default: false });
const emit = defineEmits(["close"]);

const props = defineProps({
  role: { type: String, default: "dialog" },
  title: { type: String, default: "" },
  panelClass: { type: String, default: "" },
  backdrop: { type: Boolean, default: false },
  closeOnBackdrop: { type: Boolean, default: false },
});

const panel = ref(null);
let returnFocus = null;
let bodyLocks = 0;
let previousOverflow = "";
let bodyLocked = false;

function lockBody() {
  if (!bodyLocks++) previousOverflow = document.body.style.overflow;
  document.body.style.overflow = "hidden";
  bodyLocked = true;
}

function unlockBody() {
  if (!bodyLocked) return;
  bodyLocked = false;
  if (--bodyLocks === 0) document.body.style.overflow = previousOverflow;
}

function close() {
  open.value = false;
  emit("close");
}

useOverlayStack(open, close);

function focusPanel() {
  const target = panel.value?.querySelector("input, select, textarea, button, [tabindex]:not([tabindex='-1'])");
  (target ?? panel.value)?.focus();
}

function trapFocus(event) {
  if (event.key !== "Tab" || !panel.value) return;
  const focusable = [...panel.value.querySelectorAll("input, select, textarea, button, [tabindex]:not([tabindex='-1'])")];
  if (!focusable.length) {
    event.preventDefault();
    panel.value.focus();
    return;
  }
  const first = focusable[0];
  const last = focusable.at(-1);
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

watch(open, async (isOpen) => {
  if (isOpen) {
    lockBody();
    returnFocus = document.activeElement;
    await nextTick();
    focusPanel();
  } else if (returnFocus?.isConnected) {
    unlockBody();
    returnFocus.focus();
    returnFocus = null;
  } else {
    unlockBody();
  }
}, { immediate: true });

onUnmounted(() => {
  unlockBody();
  if (returnFocus?.isConnected) returnFocus.focus();
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="overlay-shell"
      :class="{ 'has-backdrop': props.backdrop }"
      @click.self="props.closeOnBackdrop && close()"
    >
      <section
        ref="panel"
        class="overlay-panel"
        :class="props.panelClass"
        :role="props.role"
        :aria-modal="props.role === 'dialog' ? 'true' : undefined"
        :aria-label="props.title || undefined"
        tabindex="-1"
        @keydown="trapFocus"
      >
        <slot />
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.overlay-shell {
  position: fixed;
  inset: 0;
  z-index: 60;
  pointer-events: none;
}
.overlay-shell.has-backdrop {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: max(12px, env(safe-area-inset-top)) max(12px, env(safe-area-inset-right))
    max(12px, env(safe-area-inset-bottom)) max(12px, env(safe-area-inset-left));
  background: rgb(0 0 0 / 65%);
  pointer-events: auto;
}
.overlay-panel {
  max-width: 100%;
  max-height: 100%;
  pointer-events: auto;
  outline: none;
}
</style>
