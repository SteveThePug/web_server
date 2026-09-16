<script setup>
/**
 * Link styled to sit inside running prose (bold italic, no underline). Same
 * `to` vs `href` and auto-`rel` behaviour as Link.vue, shared through
 * useLinkAttrs.js.
 */
import { linkProps, useComputedRel } from "@/components/text/useLinkAttrs";

const props = defineProps({ ...linkProps });

const computedRel = useComputedRel(props);
</script>

<template>
  <RouterLink v-if="to" :to="to" class="inline-link">
    <slot />
  </RouterLink>
  <a
    v-else
    :href="href"
    :target="target"
    :rel="computedRel"
    class="inline-link"
  >
    <slot />
  </a>
</template>

<style scoped>
.inline-link {
  color: var(--color-primary);
  font-weight: bold;
  font-style: italic;
  text-decoration: none;
  transition: color 0.15s ease;
}

.inline-link:hover {
  color: var(--color-tertiary);
}
</style>
