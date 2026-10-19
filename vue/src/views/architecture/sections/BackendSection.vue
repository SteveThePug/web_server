<script setup>
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import RequestLifecycle from "../diagrams/backend/RequestLifecycle.vue";
</script>

<template>
    <section id="backend" class="flex flex-col gap-6">
        <Header>Go backend</Header>
        <Paragraph>
            A Gin server with a GraphQL API at <code>/api/graphql</code> and a
            handful of REST endpoints. Every request passes through the same
            middleware chain before it reaches a handler or resolver.
        </Paragraph>
        <RequestLifecycle />
    </section>
</template>
