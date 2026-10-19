<!--
  Frame for one diagram on /architecture: numbered title, the SVG (default
  slot), and a caption (#caption slot).

  The SVG scales to the column; below `minWidth` it stops shrinking and the
  frame scrolls sideways instead, so small text stays readable on a phone.
  Clicking "expand" lifts the figure into a full-screen overlay for the big
  ones. Escape or the button closes it.
-->
<template>
  <figure :id="id" class="figure" :class="{ expanded }">
    <figcaption class="head">
      <span class="title">
        <a v-if="id" :href="`#${id}`" class="anchor">#</a>
        {{ title }}
      </span>
      <button type="button" class="expand" @click="expanded = !expanded">
        {{ expanded ? "close ✕" : "expand ⤢" }}
      </button>
    </figcaption>
    <div class="canvas">
      <div class="inner" :style="{ minWidth: `${minWidth}px` }">
        <slot />
      </div>
    </div>
    <div v-if="$slots.caption" class="caption">
      <slot name="caption" />
    </div>
  </figure>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from "vue";

defineProps({
  title: { type: String, required: true },
  // Anchor id, so a figure can be linked to directly.
  id: { type: String, default: "" },
  minWidth: { type: Number, default: 640 },
});

const expanded = ref(false);

function onKey(e) {
  if (e.key === "Escape") expanded.value = false;
}
watch(expanded, (open) => {
  if (open) window.addEventListener("keydown", onKey);
  else window.removeEventListener("keydown", onKey);
});
onBeforeUnmount(() => window.removeEventListener("keydown", onKey));
</script>

<style scoped>
.figure {
  margin: 0;
  border: 1px solid var(--color-quaternary);
  background: var(--color-surface-deep);
  scroll-margin-top: 4rem;
}
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  border-bottom: 1px solid var(--color-quaternary);
  background: var(--color-surface);
}
.title {
  font-family: var(--font-heading);
  font-size: 1.15rem;
  letter-spacing: 0.04em;
  color: var(--color-primary);
}
.anchor {
  background: none;
  color: var(--color-tertiary);
  text-decoration: none;
  margin-right: 0.25rem;
}
.expand {
  font-family: var(--font-heading);
  letter-spacing: 0.05em;
  color: var(--color-tertiary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-quaternary);
  padding: 0.1rem 0.5rem;
  cursor: pointer;
  white-space: nowrap;
}
.expand:hover {
  border-color: var(--color-tertiary);
}
.canvas {
  overflow-x: auto;
  padding: 0.75rem;
}
.caption {
  padding: 0.5rem 0.75rem 0.75rem;
  border-top: 1px dotted var(--color-quaternary);
  color: var(--color-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}
.caption :deep(code) {
  font-size: 0.85em;
}

.expanded {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.expanded .canvas {
  flex: 1;
  overflow: auto;
  display: flex;
  align-items: flex-start;
  justify-content: center;
}
.expanded .inner {
  width: min(100%, 1600px);
}
.expanded .caption {
  max-height: 25vh;
  overflow-y: auto;
}
</style>
