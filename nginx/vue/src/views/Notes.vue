<script setup>
import Markdown from "@/components/quick/Markdown.vue";
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

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div
            v-if="file"
            class="a4page-portrait bdr-primary flex-col relative scroll-y gap bg-primary"
        >
            <h1>{{ filename }}</h1>
            <small>{{ last_edited }}</small>
            <Markdown class="fill wrap" :source="file" />
        </div>

        <div v-else>Loading…</div>
    </main>
</template>
