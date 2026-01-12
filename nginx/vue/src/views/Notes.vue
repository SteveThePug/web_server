<script setup>
import Markdown from "@/components/quick/Markdown.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const note = ref(await axios.post("/api/notes/Welcome"));

function fetchNote() {}
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />
        <div class="a4page-portrait bdr-1 flex-col relative scroll-y gap">
            <h1>{{ note.title }}</h1>
            <small>{{ note.last_edited }}</small>
            <Markdown class="fill wrap" :source="note.contents" />
        </div>
    </main>
</template>
