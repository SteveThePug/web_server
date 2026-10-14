<script setup>
/**
 * A box showing one of the built-in Vue widgets from widgetRegistry.js.
 *
 * The widgets were written for Home.vue's grid cells and sidebars, and a few
 * size their own root (Steam is 54mm tall). On the canvas the box decides the
 * size, so the root is forced to fill it — hence the !important, which has to
 * beat each widget's own scoped rule regardless of stylesheet order.
 */
import { computed } from "vue";
import { WIDGETS } from "./widgetRegistry";

const props = defineProps({
  name: { type: String, required: true },
});

const entry = computed(() => WIDGETS[props.name]);
</script>

<template>
  <div class="widget-host">
    <component :is="entry.component" v-if="entry" />
    <p v-else class="missing">Unknown widget “{{ name }}”</p>
  </div>
</template>

<style scoped>
.widget-host {
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.widget-host > :deep(*) {
  width: 100% !important;
  height: 100% !important;
  max-height: none !important;
  flex: 1 1 auto;
  min-height: 0;
}
.missing {
  margin: auto;
  color: var(--color-tertiary);
}
</style>
