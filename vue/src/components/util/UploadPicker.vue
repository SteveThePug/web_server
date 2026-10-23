<script setup>
/**
 * Choose a file that is already uploaded, instead of uploading it again: a
 * modal grid of every public upload (listUploads in @/uploads), newest first.
 * Admin only, because the listing is. Used by the canvas's Background panel,
 * Box editor and "+ From uploads" button.
 *
 *   <UploadPicker v-if="picking" v-model="picking" :kinds="['image']" @pick="use" />
 *
 * `kinds` filters by kindOf(url). The bucket holds hundreds of megabytes, so
 * thumbnails load lazily and videos fetch only their metadata.
 */
import { computed, ref } from "vue";
import { kindOf, listUploads } from "@/uploads";
import OverlayShell from "./OverlayShell.vue";

const open = defineModel({ type: Boolean, default: false });

const props = defineProps({
  kinds: { type: Array, default: () => ["image"] },
  title: { type: String, default: "Choose an upload" },
});

const emit = defineEmits(["pick", "close"]);

const files = ref([]);
const loading = ref(true);
const error = ref("");
const filter = ref("");

listUploads()
  .then((all) => (files.value = all.filter((f) => props.kinds.includes(kindOf(f.url)))))
  .catch((err) => (error.value = err?.message || "Couldn't list uploads"))
  .finally(() => (loading.value = false));

const shown = computed(() => {
  const q = filter.value.trim().toLowerCase();
  return q ? files.value.filter((f) => f.name.toLowerCase().includes(q)) : files.value;
});

function describe(f) {
  const kb = f.size / 1024;
  const size = kb < 1024 ? `${Math.round(kb)} KB` : `${(kb / 1024).toFixed(1)} MB`;
  return `${f.name} · ${size} · ${new Date(f.modified).toLocaleString()}`;
}

function close() {
  open.value = false;
  emit("close");
}

function pick(f) {
  emit("pick", f.url);
  close();
}
</script>

<template>
  <OverlayShell v-model="open" :title="title" backdrop close-on-backdrop @close="close">
    <div class="picker bdr-2 bg-surface">
      <header class="flex items-center gap-2">
        <h3 class="m-0 flex-1">{{ title }}</h3>
        <input v-model="filter" type="search" placeholder="Filter by name" class="filter" />
        <button type="button" class="close" title="Close" @click="close">✕</button>
      </header>

      <p v-if="loading" class="state">Loading…</p>
      <p v-else-if="error" class="state text-tertiary">{{ error }}</p>
      <p v-else-if="!shown.length" class="state">
        {{ files.length ? "Nothing matches." : "No uploads yet." }}
      </p>
      <div v-else class="grid">
        <button v-for="f in shown" :key="f.name" type="button" class="tile" :title="describe(f)" @click="pick(f)">
          <img v-if="kindOf(f.url) === 'image'" :src="f.url" alt="" loading="lazy" />
          <video v-else-if="kindOf(f.url) === 'video'" :src="f.url" preload="metadata" muted></video>
          <span v-else class="name">{{ f.name }}</span>
        </button>
      </div>
    </div>
  </OverlayShell>
</template>

<style scoped>
.picker {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  width: min(760px, calc(100vw - 24px));
  height: min(600px, calc(100vh - 24px));
  padding: 0.75rem;
  color: var(--color-secondary);
}
.filter {
  min-width: 0;
  width: 12rem;
  padding: 3px 5px;
  color: var(--color-secondary);
  background: var(--color-surface-deep);
  border: 1px solid var(--color-quaternary);
}
.close {
  color: var(--color-primary);
  background: var(--color-link-bg);
  border: 1px solid var(--color-primary);
  padding: 0 0.5rem;
  cursor: pointer;
}
.state {
  color: var(--color-muted);
}
.grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  grid-auto-rows: 120px;
  gap: 0.5rem;
}
.tile {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  overflow: hidden;
  background: var(--color-surface-deep);
  border: 2px solid var(--color-quaternary);
  cursor: pointer;
}
.tile:hover,
.tile:focus-visible {
  border-color: var(--color-primary);
}
.tile img,
.tile video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
}
.name {
  padding: 0.25rem;
  font-size: 0.75rem;
  word-break: break-all;
  color: var(--color-muted);
}
</style>
