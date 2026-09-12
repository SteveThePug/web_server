<script setup>
import { watch, onUnmounted } from "vue";

// Popup modal: a dimmed backdrop over the page with the content floated on top.
// Closes on backdrop click or Escape. Used by the home widgets so their create
// forms pop up instead of replacing the widget's contents.
const open = defineModel({ type: Boolean, default: false });

const props = defineProps({
  maxWidth: { type: String, default: "420px" },
});

function onKeydown(e) {
  if (e.key === "Escape") open.value = false;
}

watch(
  open,
  (isOpen) => {
    if (isOpen) {
      window.addEventListener("keydown", onKeydown);
    } else {
      window.removeEventListener("keydown", onKeydown);
    }
  },
  { immediate: true },
);

onUnmounted(() => window.removeEventListener("keydown", onKeydown));
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="modal-backdrop" @click.self="open = false">
      <div class="modal bdr-1" :style="{ maxWidth: props.maxWidth }">
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.65);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  padding: 12px;
}

.modal {
  background-color: var(--color-surface);
  padding: 16px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}
</style>
