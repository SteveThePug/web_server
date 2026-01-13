<script setup>
import MarkdownIt from "markdown-it";
import { katex } from "@mdit/plugin-katex";
import wiki from "markdown-it-wikilinks";

const mdIt = MarkdownIt().use(katex);
//.use(wiki);

const props = defineProps({
    source: String,
});
</script>

<template>
    <div
        v-html="mdIt.render(props.source)"
        class="flex-col center-content"
    ></div>
</template>

<style>
@import "katex/dist/katex.min.css";
</style>
