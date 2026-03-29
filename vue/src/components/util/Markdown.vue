<script setup>
import MarkdownIt from "markdown-it";
import { katex } from "@mdit/plugin-katex";
import DOMPurify from "dompurify";

const mdIt = MarkdownIt().use(katex);

const props = defineProps({
    source: String,
});

function renderMarkdown(source) {
    return DOMPurify.sanitize(mdIt.render(source));
}
</script>

<template>
    <div
        v-html="renderMarkdown(props.source)"
        class="flex flex-col items-center"
    ></div>
</template>

<style>
@import "katex/dist/katex.min.css";
</style>
