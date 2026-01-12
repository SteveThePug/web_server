<script setup>
import Markdown from "@/components/quick/Markdown.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const note = ref(null);

const fetchNote = async () => {
    const response = await axios.post("/api/notes/Welcome");
    note.value = response.data;
};

onMounted(fetchNote);
</script>

<template>
    <main class="center-content flex-col">
        <div class="background halftone" />

        <div
            v-if="note"
            class="a4page-portrait bdr-1 flex-col relative scroll-y gap"
        >
            <h1>{{ note.title }}</h1>
            <small>{{ note.last_edited }}</small>
            <Markdown class="fill wrap" :source="note.contents" />
        </div>

        <div v-else>Loading…</div>
    </main>
</template>
