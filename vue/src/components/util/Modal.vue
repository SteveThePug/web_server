<script setup>
import OverlayShell from "./OverlayShell.vue";

// Popup modal: a dimmed backdrop over the page with the content floated on top.
const open = defineModel({ type: Boolean, default: false });

const props = defineProps({
  maxWidth: { type: String, default: "420px" },
});

</script>

<template>
  <OverlayShell v-model="open" title="Dialog" backdrop close-on-backdrop>
    <div class="modal bdr-1" :style="{ maxWidth: props.maxWidth }">
      <slot />
    </div>
  </OverlayShell>
</template>

<style scoped>
.modal {
  box-sizing: border-box;
  background-color: var(--color-surface);
  padding: 16px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}
</style>
