<script setup>
import { computed } from "vue";

const props = defineProps({
    href: { type: String, default: "" },
    to: { type: String, default: "" },
    target: { type: String, default: undefined },
    rel: { type: String, default: undefined },
});

const computedRel = computed(() => {
    if (props.rel !== undefined) return props.rel;
    if (props.target === "_blank") return "noopener noreferrer";
    return undefined;
});
</script>

<template>
    <RouterLink v-if="to" :to="to" class="inline-link">
        <slot />
    </RouterLink>
    <a v-else :href="href" :target="target" :rel="computedRel" class="inline-link">
        <slot />
    </a>
</template>

<style scoped>
.inline-link {
    color: var(--primary);
    font-weight: bold;
    font-style: italic;
    text-decoration: none;
    transition: color 0.15s ease;
}

.inline-link:hover {
    color: var(--tertiary);
}
</style>
