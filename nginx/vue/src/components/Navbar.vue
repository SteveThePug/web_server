<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const parentPath = computed(() => {
    const segments = route.path.split("/").filter(Boolean);
    if (segments.length == 1) {
        return "/";
    } else {
        segments.pop();
        return segments.length ? "/" + segments.join("/") : null;
    }
});

const inHome = computed(() => {
    return route.path == "/";
});
</script>

<template>
    <nav class="flex flex-row w-fit">
        <RouterLink class="bdr-2 bg-bg_primary" to="/" v-if="!inHome">
            <a>HOME</a>
        </RouterLink>
        <RouterLink
            class="bdr-2 bg-bg_primary"
            v-if="parentPath"
            :to="parentPath"
        >
            <a>UP</a>
        </RouterLink>
    </nav>
</template>

<style scoped>
.left {
    position: fixed;
    top: 0;
    left: 0;
}
</style>
