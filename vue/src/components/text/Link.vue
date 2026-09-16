<script setup>
/**
 * Standard link. Renders a <RouterLink> when given `to`, a plain <a> when given
 * `href`. `bare` drops the styling (used where the parent styles the anchor).
 * External `target="_blank"` links get rel="noopener noreferrer" automatically.
 */
import { computed } from "vue";

const props = defineProps({
  href: { type: String, default: "" },
  to: { type: String, default: "" },
  target: { type: String, default: undefined },
  rel: { type: String, default: undefined },
  bare: { type: Boolean, default: false },
});

const computedRel = computed(() => {
  if (props.rel !== undefined) return props.rel;
  if (props.target === "_blank") return "noopener noreferrer";
  return undefined;
});
</script>

<template>
  <RouterLink v-if="to" :to="to" :class="{ link: !bare }">
    <slot />
  </RouterLink>
  <a
    v-else
    :href="href"
    :target="target"
    :rel="computedRel"
    :class="{ link: !bare }"
  >
    <slot />
  </a>
</template>

<style scoped>
.link {
  color: var(--color-primary);
  text-decoration: none;
  transition: color 0.15s ease;
}

.link:hover {
  color: var(--color-tertiary);
}
</style>
