<!--
  A credential shown exactly once: the value, a copy button, and the warning
  that it cannot be read back. Used wherever the server mints a secret and
  keeps only a hash of it, or no copy the browser can ask for again.
-->
<script setup>
import { ref, watch } from "vue";
import Button from "@/components/input/Button.vue";

const props = defineProps({
  value: { type: String, required: true },
  label: { type: String, default: "" },
});

const copied = ref(false);
const failed = ref(false);

// A regenerated value is a new thing to copy.
watch(
  () => props.value,
  () => {
    copied.value = false;
    failed.value = false;
  },
);

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value);
    copied.value = true;
  } catch {
    // Clipboard access can be refused; the value is on screen to select.
    failed.value = true;
  }
}
</script>

<template>
  <div class="secret-reveal">
    <p class="secret-note">
      <code v-if="label">{{ label }}</code>
      Copy this now. It will not be shown again.
    </p>
    <code class="secret-value">{{ value }}</code>
    <Button class="px-2" @click="copy">{{ copied ? "Copied" : "Copy" }}</Button>
    <p v-if="failed" class="secret-note secret-failed">
      Could not copy — select the value and copy it by hand.
    </p>
  </div>
</template>

<style scoped>
.secret-reveal {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--color-tertiary);
  background: var(--color-surface-tint);
}
.secret-note {
  flex-basis: 100%;
  margin: 0;
  font-size: 0.85rem;
  opacity: 0.85;
}
.secret-failed {
  color: var(--color-error, #b91c1c);
}
.secret-value {
  flex: 1;
  min-width: 0;
  font-size: 0.8rem;
  word-break: break-all;
  user-select: all;
}
</style>
