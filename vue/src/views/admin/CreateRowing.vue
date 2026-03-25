<script setup>
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import axios from "axios";

const images = ref([]);
const results = ref([]);

function onFileChange(e) {
    images.value = Array.from(e.target.files);
    results.value = [];
}

async function submit() {
    if (!images.value.length) return;
    results.value = images.value.map((f) => ({ name: f.name, status: "Uploading..." }));

    await Promise.all(
        images.value.map(async (file, i) => {
            const formData = new FormData();
            formData.append("image", file);
            try {
                const res = await axios.post("/api/rowing", formData, {
                    headers: { "Content-Type": "multipart/form-data" },
                });
                const mins = Math.floor(res.data.Time / 1e9 / 60);
                const secs = String(Math.floor((res.data.Time / 1e9) % 60)).padStart(2, "0");
                results.value[i].status = `${res.data.Distance}m in ${mins}:${secs}`;
                results.value[i].ok = true;
            } catch (err) {
                results.value[i].status = err.response?.data?.error || "Upload failed";
                results.value[i].ok = false;
            }
        })
    );

    images.value = [];
}
</script>

<template>
    <div class="flex flex-col gap-2">
        <h1>Create Rowing</h1>
        <input type="file" accept="image/jpeg,image/png,image/gif,image/webp" multiple @change="onFileChange" />
        <Button @click="submit">Upload</Button>
        <div v-for="r in results" :key="r.name">
            <span class="text-primary">{{ r.name }}: </span>
            <span :class="r.ok ? 'text-secondary' : 'text-red-500'">{{ r.status }}</span>
        </div>
    </div>
</template>
