<script setup>
import OverlayShell from "./OverlayShell.vue";

const open = defineModel({ type: Boolean, default: false });

defineProps({
  title: { type: String, default: "Are you sure?" },
  message: { type: String, required: true },
  confirmLabel: { type: String, default: "Confirm" },
});

const emit = defineEmits(["confirm", "cancel"]);

function cancel() {
  open.value = false;
  emit("cancel");
}

function confirm() {
  open.value = false;
  emit("confirm");
}
</script>

<template>
  <OverlayShell v-model="open" :title="title" backdrop close-on-backdrop @close="cancel">
    <div class="confirm bdr-1">
      <h2>{{ title }}</h2>
      <p>{{ message }}</p>
      <div class="actions">
        <button type="button" class="confirm-button" @click="confirm">{{ confirmLabel }}</button>
        <button type="button" class="cancel-button" @click="cancel">Cancel</button>
      </div>
    </div>
  </OverlayShell>
</template>

<style scoped>
.confirm {
  box-sizing: border-box;
  width: min(420px, calc(100vw - 24px));
  padding: 16px;
  color: var(--color-secondary);
  background: var(--color-surface);
}
.confirm h2 {
  margin-top: 0;
}
.actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
.actions button {
  padding: 0.35rem 0.75rem;
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-primary);
  cursor: pointer;
}
.confirm-button {
  color: var(--color-tertiary) !important;
  border-color: var(--color-tertiary) !important;
}
</style>
