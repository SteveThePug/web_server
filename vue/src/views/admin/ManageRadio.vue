<script setup>
/**
 * Admin panel: list and upload tracks for the Icecast radio. REST
 * (multipart/form-data) rather than GraphQL because it moves files.
 * Deleting removes the file from the radio's library for good, so it asks
 * first; a failed list, enable/disable or delete is shown above the list.
 */
import Button from "@/components/input/Button.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";
import { ref, onMounted } from "vue";
import { api } from "@/api";

const songs = ref([]);
const files = ref([]);
const results = ref([]);
const loading = ref(false);
const error = ref("");
const deleteTarget = ref(null);
const confirmOpen = ref(false);

async function fetchSongs() {
  try {
    const res = await api.get("/radio/songs", {
      fallback: "Couldn't load the songs",
    });
    songs.value = res.songs;
  } catch (err) {
    console.error(err);
    error.value = err.message;
  }
}

function onFileChange(e) {
  files.value = Array.from(e.target.files);
  results.value = [];
}

async function upload() {
  if (!files.value.length) return;
  loading.value = true;
  results.value = files.value.map((f) => ({
    name: f.name,
    status: "Uploading...",
  }));

  await Promise.all(
    files.value.map(async (file, i) => {
      const formData = new FormData();
      formData.append("file", file);
      try {
        await api.post("/radio/upload", formData, {
          headers: { "Content-Type": "multipart/form-data" },
          fallback: "Upload failed",
        });
        results.value[i].status = "Uploaded";
        results.value[i].ok = true;
      } catch (err) {
        results.value[i].status = err.message;
        results.value[i].ok = false;
      }
    }),
  );

  files.value = [];
  loading.value = false;
  await fetchSongs();
}

function askDelete(song) {
  deleteTarget.value = song;
  confirmOpen.value = true;
}

async function deleteSong() {
  const song = deleteTarget.value;
  deleteTarget.value = null;
  if (!song) return;
  error.value = "";
  try {
    await api.delete(`/radio/songs/${encodeURIComponent(song.name)}`, {
      fallback: "Delete failed",
    });
    await fetchSongs();
  } catch (err) {
    console.error(err);
    error.value = err.message;
  }
}

async function toggleSong(song) {
  const action = song.disabled ? "enable" : "disable";
  error.value = "";
  try {
    await api.patch(
      `/radio/songs/${encodeURIComponent(song.name)}/${action}`,
      undefined,
      { fallback: `Couldn't ${action} the song` },
    );
    await fetchSongs();
  } catch (err) {
    console.error(err);
    error.value = err.message;
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
    <input
      type="file"
      accept=".mp3,.ogg,.flac,.wav,.m4a,.opus"
      multiple
      @change="onFileChange"
    />
    <Button @click="upload" :disabled="loading || !files.length">
      {{ loading ? "Uploading…" : "Upload" }}
    </Button>
    <p v-if="error" class="text-tertiary" role="alert">{{ error }}</p>
    <div v-for="r in results" :key="r.name">
      <span class="text-primary">{{ r.name }}: </span>
      <span :class="r.ok ? 'text-secondary' : 'text-tertiary'">{{
        r.status
      }}</span>
    </div>
    <div
      v-for="song in songs"
      :key="song.name"
      class="flex flex-row flex-wrap items-center gap-2"
      :class="{ 'opacity-50': song.disabled }"
    >
      <span class="text-primary" :class="{ 'line-through': song.disabled }">
        {{ song.name }}
      </span>
      <span class="text-secondary text-sm">{{ formatSize(song.size) }}</span>
      <span v-if="song.disabled" class="text-tertiary text-xs">disabled</span>
      <Button @click="toggleSong(song)">{{
        song.disabled ? "Enable" : "Disable"
      }}</Button>
      <Button @click="askDelete(song)">Delete</Button>
    </div>

    <ConfirmDialog
      v-if="deleteTarget"
      v-model="confirmOpen"
      title="Delete song"
      :message="`Delete “${deleteTarget.name}” from the radio? The file is removed for good.`"
      confirm-label="Delete"
      @confirm="deleteSong"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
