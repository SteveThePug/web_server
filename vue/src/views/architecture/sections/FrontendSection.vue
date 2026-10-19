<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import RouteTree from "../diagrams/frontend/RouteTree.vue";
import BootSequence from "../diagrams/frontend/BootSequence.vue";
</script>

<template>
    <section id="frontend" class="flex flex-col gap-6">
        <Header>Vue frontend</Header>
        <Paragraph>
            A Vue 3 single-page app: a small route table with two layout
            shells, a catch-all that serves admin-made canvas pages, and Pinia
            stores that load data over GraphQL.
        </Paragraph>
        <RouteTree />
        <BootSequence />
    </section>
</template>
