<script setup>
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
    <a v-else :href="href" :target="target" :rel="computedRel" :class="{ link: !bare }">
        <slot />
    </a>
</template>

<style scoped>
.link {
    color: var(--primary);
    text-decoration: none;
    transition: color 0.15s ease;
}

.link:hover {
    color: var(--tertiary);
}
</style>
