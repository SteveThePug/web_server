<script setup>
import Markdown from "@/components/util/Markdown.vue";
import { ref, onMounted } from "vue";
import axios from "axios";
import { useRoute } from "vue-router";

const file = ref(null);
const filename = ref("");
const last_edited = ref(null);

// if the address is https://www.adam-french.co.uk/notes/PATH
// request from https://www.adam-french.co.uk/api/notes/PATH
const route = useRoute();
const pathArray = route.params.path;
const path = Array.isArray(pathArray) ? pathArray.join("/") : pathArray;
const url = `/api/notes/${path}`;

function getFilename(headers) {
    const disposition = headers["content-disposition"];
    if (!disposition) return null;

    const match = disposition.match(/filename="?([^"]+)"?/);
    return match ? match[1] : null;
}

async function fetchFile() {
    const response = await axios.get(url, { responseType: "blob" });
    filename.value = getFilename(response.headers);

    const lastModified = response.headers["last-modified"];
    last_edited.value = lastModified ? new Date(lastModified) : null;

    if (filename.value.toLowerCase().endsWith(".md")) {
        const text = await response.data.text();
        file.value = fixLinks(text);
    } else {
        file.value = response.data;
    }
}

function fixLinks(filedata) {
    return filedata.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (match, text, url) => {
        if (
            url.startsWith("http://") ||
            url.startsWith("https://") ||
            url.startsWith("#") ||
            url.startsWith("./") ||
            url.startsWith("../") ||
            url.startsWith("//")
        ) {
            return match;
        }

        return `[${text}](/notes/${url})`;
    });
}

onMounted(fetchFile);
</script>

<style>
/* ── Obsidian-style note container ───────────────────────────────────── */
.obsidian-note {
    position: relative;
    width: min(720px, 100%);
    min-height: 60vh;
    margin: 0 auto;
    background-color: var(--bg_primary);
    border: 1px solid var(--quaternary);
    box-shadow: 0 0 0 1px var(--quaternary), 4px 4px 24px rgba(0, 0, 0, 0.6);
    display: flex;
    flex-direction: column;
}

/* Ruled-line paper texture on the body area */
.obsidian-body {
    display: flex;
    flex: 1;
    background-image: repeating-linear-gradient(
        to bottom,
        transparent 0px,
        transparent calc(1.6rem - 1px),
        var(--quaternary) calc(1.6rem - 1px),
        var(--quaternary) 1.6rem
    );
    background-attachment: local;
}

/* ── Title bar ───────────────────────────────────────────────────────── */
.obsidian-titlebar {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 12px 16px 10px;
    border-bottom: 1px solid var(--quaternary);
    background-color: var(--bg_secondary);
}

.obsidian-breadcrumb {
    font-family: var(--font_heading);
    font-size: 0.75rem;
    letter-spacing: 0.08em;
    color: var(--primary);
    opacity: 0.7;
    text-transform: lowercase;
}

.obsidian-meta {
    font-size: 0.7rem;
    color: var(--tertiary);
    opacity: 0.8;
}

/* ── Left gutter (Obsidian fold-indicator column) ────────────────────── */
.obsidian-gutter {
    width: 32px;
    flex-shrink: 0;
    border-right: 1px solid var(--quaternary);
    background-color: color-mix(in srgb, var(--bg_secondary) 60%, transparent);
}

/* ── Content area ────────────────────────────────────────────────────── */
.obsidian-content {
    flex: 1;
    padding: 24px 32px 40px 24px;
    line-height: 1.6rem;
    overflow-wrap: break-word;
    min-width: 0;
}

/* Headings sit on the ruled baseline */
.obsidian-content h1,
.obsidian-content h2,
.obsidian-content h3,
.obsidian-content h4 {
    line-height: 1.6rem;
    margin-bottom: 0;
    padding-top: 1.6rem;
    border-bottom: 1px dotted var(--quaternary);
}

.obsidian-content h1 {
    font-size: 1.6rem;
    line-height: 3.2rem;
    padding-top: 0;
}

/* Paragraphs align to grid */
.obsidian-content p {
    margin: 0;
    padding: 0;
    line-height: 1.6rem;
}

/* Code blocks */
.obsidian-content pre {
    background-color: var(--bg_secondary);
    border-left: 3px solid var(--primary);
    padding: 4px 12px;
    margin: 0;
    overflow-x: auto;
    line-height: 1.6rem;
}

.obsidian-content pre code {
    background: none;
    padding: 0;
}

.obsidian-content code {
    background-color: var(--bg_secondary);
    padding: 0 4px;
    border-radius: 2px;
}

/* Blockquotes */
.obsidian-content blockquote {
    border-left: 3px solid var(--tertiary);
    margin: 0;
    padding-left: 16px;
    color: var(--tertiary);
}

/* Links */
.obsidian-content a {
    text-decoration: underline;
    text-decoration-style: dotted;
}

/* Lists stay on grid */
.obsidian-content ul,
.obsidian-content ol {
    margin: 0;
    padding-left: 1.5rem;
}

.obsidian-content li {
    line-height: 1.6rem;
}
</style>

<template>
    <main class="items-center flex flex-col">
        <div class="background halftone" />
        <div v-if="file" class="obsidian-note">
            <div class="obsidian-titlebar">
                <div class="obsidian-breadcrumb">notes / {{ filename }}</div>
                <small class="obsidian-meta">last edited: {{ last_edited?.toLocaleString() }}</small>
            </div>
            <div class="obsidian-body">
                <div class="obsidian-gutter" aria-hidden="true" />
                <Markdown class="obsidian-content" :source="file" />
            </div>
        </div>

        <div v-else class="text-secondary">Loading…</div>
    </main>
</template>
