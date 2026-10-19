<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import PythonRouteMap from "../diagrams/services/PythonRouteMap.vue";
import RowingClaudeCall from "../diagrams/services/RowingClaudeCall.vue";
</script>

<template>
    <section id="services" class="flex flex-col gap-6">
        <Header>Python &amp; the other services</Header>
        <Paragraph>
            The FastAPI service holds the scrapers and the site's one Claude
            call. Public routes are reached through nginx at
            <code>/py/</code>; internal ones only from the Go backend.
        </Paragraph>
        <PythonRouteMap />
        <RowingClaudeCall />
    </section>
</template>
