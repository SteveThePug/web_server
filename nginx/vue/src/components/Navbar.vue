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
</script>

<template>
    <nav class="flex-row">
        <RouterLink class="bdr-2" to="/"><h1>HOME</h1></RouterLink>
        <RouterLink class="bdr-2" v-if="parentPath" :to="parentPath">
            <h1>UP</h1>
        </RouterLink>
    </nav>
</template>

<style scoped>
h1 {
    padding: 10px;
}
</style>
