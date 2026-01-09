<script setup>
import Markdown from "@/components/quick/Markdown.vue";

const source = `
The gaussian is $\\frac{1}{\\sqrt{2\\pi}}e^{-\\frac{x^2}{2}}$.
`;
</script>

<template>
    <main class="center-content flex-col">
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y">
            <h1>These are my notes about Vue.js</h1>
            <h1>Pinia</h1>
            <Markdown :source="source" />
            <h1>Composition API</h1>
        </div>
    </main>
</template>
