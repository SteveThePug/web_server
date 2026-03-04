<script setup>
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import axios from "axios";

const image = ref(null);
const status = ref("");
const error = ref("");

function onFileChange(e) {
    image.value = e.target.files[0] || null;
}

async function submit() {
    if (!image.value) {
        error.value = "Please select an image";
        return;
    }
    error.value = "";
    status.value = "Uploading...";

    const formData = new FormData();
    formData.append("image", image.value);

    try {
        const res = await axios.post("/api/rowing", formData, {
            headers: { "Content-Type": "multipart/form-data" },
        });
        status.value = `Saved: ${res.data.Distance}m in ${Math.floor(res.data.Time / 1e9 / 60)}:${String(Math.floor((res.data.Time / 1e9) % 60)).padStart(2, "0")}`;
        image.value = null;
    } catch (err) {
        status.value = "";
        error.value = err.response?.data?.error || "Upload failed";
    }
}
</script>

<template>
    <div class="flex flex-col gap-2">
        <h1>Create Rowing</h1>
        <input type="file" accept="image/jpeg,image/png,image/gif,image/webp" @change="onFileChange" />
        <Button @click="submit">Upload</Button>
        <p v-if="status">{{ status }}</p>
        <p v-if="error" class="text-red-500">{{ error }}</p>
    </div>
</template>
