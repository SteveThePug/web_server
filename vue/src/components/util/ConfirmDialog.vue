<script setup>
import OverlayShell from "./OverlayShell.vue";

const open = defineModel({ type: Boolean, default: false });

defineProps({
  title: { type: String, default: "Are you sure?" },
  message: { type: String, required: true },
  confirmLabel: { type: String, default: "Confirm" },
  // A second, milder choice beside the confirm button, e.g. "Add on top"
  // next to "Replace". Omitted, there are only Confirm and Cancel.
  secondaryLabel: { type: String, default: "" },
});

const emit = defineEmits(["confirm", "secondary", "cancel"]);

function cancel() {
  open.value = false;
  emit("cancel");
}

function confirm() {
  open.value = false;
  emit("confirm");
}

function secondary() {
  open.value = false;
  emit("secondary");
}
</script>

<template>
  <OverlayShell v-model="open" :title="title" backdrop close-on-backdrop @close="cancel">
    <div class="confirm bdr-1">
      <h2>{{ title }}</h2>
      <p>{{ message }}</p>
      <div class="actions">
        <button type="button" class="confirm-button" @click="confirm">{{ confirmLabel }}</button>
        <button v-if="secondaryLabel" type="button" @click="secondary">{{ secondaryLabel }}</button>
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
