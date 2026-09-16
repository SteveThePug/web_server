<script setup>
/**
 * Admin form: upload photos of an erg monitor. The Go backend OCRs each image and
 * returns the parsed session, so this posts multipart/form-data to /api/rowing
 * rather than using GraphQL. Uploads run in parallel and each file's row reports
 * its own outcome.
 *
 * The returned Time is a Go duration in nanoseconds, hence the /1e9 before
 * splitting into minutes and seconds.
 */
import Button from "@/components/input/Button.vue";
import { ref } from "vue";
import axios from "axios";

const emit = defineEmits(["done", "cancel"]);

const images = ref([]);
const results = ref([]);

function onFileChange(e) {
  images.value = Array.from(e.target.files);
  results.value = [];
}

async function submit() {
  if (!images.value.length) return;
  results.value = images.value.map((f) => ({
    name: f.name,
    status: "Uploading...",
  }));

  await Promise.all(
    images.value.map(async (file, i) => {
      const formData = new FormData();
      formData.append("image", file);
      try {
        const res = await axios.post("/api/rowing", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        // Time comes back as a Go time.Duration, i.e. nanoseconds.
        const mins = Math.floor(res.data.Time / 1e9 / 60);
        const secs = String(Math.floor((res.data.Time / 1e9) % 60)).padStart(
          2,
          "0",
        );
        results.value[i].status = `${res.data.Distance}m in ${mins}:${secs}`;
        results.value[i].ok = true;
      } catch (err) {
        results.value[i].status = err.response?.data?.error || "Upload failed";
        results.value[i].ok = false;
      }
    }),
  );

  images.value = [];
  emit("done");
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <h1>Create Rowing</h1>
    <input
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      multiple
      @change="onFileChange"
    />
    <Button @click="submit">Upload</Button>
    <Button @click="emit('cancel')">Cancel</Button>
    <div v-for="r in results" :key="r.name">
      <span class="text-primary">{{ r.name }}: </span>
      <span :class="r.ok ? 'text-secondary' : 'text-red-500'">{{
        r.status
      }}</span>
    </div>
  </div>
</template>
