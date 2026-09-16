<script setup>
/**
 * Renders a markdown string, with KaTeX maths support, into sanitised HTML.
 *
 * The v-html is safe only because DOMPurify strips scripts and event handlers
 * from markdown-it's output first — post content is admin-authored but still
 * untrusted input. Used by the Feed widget.
 */
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
