<script setup>
import Button from "@/components/input/Button.vue";
import { ref, onMounted } from "vue";
import axios from "axios";

const songs = ref([]);
const files = ref([]);
const results = ref([]);
const loading = ref(false);

async function fetchSongs() {
    try {
        const res = await axios.get("/api/radio/songs");
        songs.value = res.data.songs;
    } catch (err) {
        console.error(err);
    }
}

function onFileChange(e) {
    files.value = Array.from(e.target.files);
    results.value = [];
}

async function upload() {
    if (!files.value.length) return;
    loading.value = true;
    results.value = files.value.map((f) => ({ name: f.name, status: "Uploading..." }));

    await Promise.all(
        files.value.map(async (file, i) => {
            const formData = new FormData();
            formData.append("file", file);
            try {
                await axios.post("/api/radio/upload", formData, {
                    headers: { "Content-Type": "multipart/form-data" },
                });
                results.value[i].status = "Uploaded";
                results.value[i].ok = true;
            } catch (err) {
                results.value[i].status = err.response?.data?.error || "Upload failed";
                results.value[i].ok = false;
            }
        })
    );

    files.value = [];
    loading.value = false;
    await fetchSongs();
}

async function deleteSong(name) {
    try {
        await axios.delete(`/api/radio/songs/${encodeURIComponent(name)}`);
        await fetchSongs();
    } catch (err) {
        console.error(err);
    }
}

async function toggleSong(song) {
    const action = song.disabled ? "enable" : "disable";
    try {
        await axios.patch(`/api/radio/songs/${encodeURIComponent(song.name)}/${action}`);
        await fetchSongs();
    } catch (err) {
        console.error(err);
    }
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

onMounted(fetchSongs);
</script>

<template>
    <div class="flex flex-col gap-2">
        <h1>Manage Radio</h1>
        <input type="file" accept=".mp3,.ogg,.flac,.wav,.m4a,.opus" multiple @change="onFileChange" />
        <Button @click="upload" :disabled="loading">Upload</Button>
        <div v-for="r in results" :key="r.name">
            <span class="text-primary">{{ r.name }}: </span>
            <span :class="r.ok ? 'text-secondary' : 'text-red-500'">{{ r.status }}</span>
        </div>
        <div
            v-for="song in songs"
            :key="song.name"
            class="flex flex-row items-center gap-2"
            :class="{ 'opacity-50': song.disabled }"
        >
            <span :class="{ 'line-through': song.disabled }">{{ song.name }}</span>
            <span class="text-secondary text-sm">{{ formatSize(song.size) }}</span>
            <span v-if="song.disabled" class="text-red-400 text-xs">disabled</span>
            <Button @click="toggleSong(song)">{{ song.disabled ? "Enable" : "Disable" }}</Button>
            <Button @click="deleteSong(song.name)">Delete</Button>
        </div>
    </div>
</template>
