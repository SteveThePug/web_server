<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import EntrypointFlow from "../diagrams/edge/EntrypointFlow.vue";
import RouteMap from "../diagrams/edge/RouteMap.vue";
</script>

<template>
    <section id="edge" class="flex flex-col gap-6">
        <Header>The edge: nginx</Header>
        <Paragraph>
            Every request enters through nginx. Its config is chosen when the
            container starts, and one shared routes file decides which
            container each path is handed to.
        </Paragraph>
        <EntrypointFlow />
        <RouteMap />
    </section>
</template>
