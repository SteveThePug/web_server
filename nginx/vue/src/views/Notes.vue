<script setup>
import Markdown from "@/components/quick/Markdown.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const note = ref(null);

function fixContents(contents) {
    // Obsidian notes have links in the form [[link|name]]
    // contents so that they are rendered correctly
    return contents.replace(/\[\[(.*?)\|(.*)\]\]/g, '<a href="/$1">$2</a>');
}

async function fetchNote() {
    const response = await axios.get("/api/notes/Index");
    response.data.contents = fixContents(response.data.contents);
    note.value = response.data;
}

onMounted(fetchNote);
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />

        <div
            v-if="note"
            class="a4page-portrait bdr-1 flex-col relative scroll-y gap bg-primary"
        >
            <h1>{{ note.title }}</h1>
            <small>{{ note.last_edited }}</small>
            <Markdown class="fill wrap" :source="note.contents" />
        </div>

        <div v-else>Loading…</div>
    </main>
</template>
