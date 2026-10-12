<script setup>
/**
 * The "Sources" box on `/cv/leads`: the watch list the lead fetcher polls.
 * Rendered only by JobLeads.vue; data comes from stores/jobSources.js.
 *
 * A source is either one company's job board (Greenhouse, Lever, Ashby —
 * identified by the slug in the board's URL) or a keyword search (Adzuna,
 * Reed — which need API keys set on the admin panel's Environment tab). The
 * kinds are a COPY of the list in backend/services/job_sources.go; the server
 * rejects any other.
 *
 * Like JobReferences.vue it has no error display of its own: failures are
 * emitted as `fail` and successes as `ok`, for the page's single banner.
 */
import { ref, computed, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useJobSourcesStore } from "@/stores/jobSources";
import "./jobs-shared.css";

const emit = defineEmits(["fail", "ok"]);

const KINDS = [
  { value: "greenhouse", label: "Greenhouse", board: true, hint: "boards.greenhouse.io/<slug>" },
  { value: "lever", label: "Lever", board: true, hint: "jobs.lever.co/<slug>" },
  { value: "ashby", label: "Ashby", board: true, hint: "jobs.ashbyhq.com/<slug>" },
  { value: "adzuna", label: "Adzuna search", board: false },
  { value: "reed", label: "Reed search", board: false },
];
const kindOf = (value) => KINDS.find((k) => k.value === value) ?? { label: value, board: true };

const store = useJobSourcesStore();
const { sources } = storeToRefs(store);

const blank = (kind = "greenhouse") => ({ kind, label: "", token: "", query: "", location: "" });
const form = ref(blank());
const formKind = computed(() => kindOf(form.value.kind));

async function fetchSources() {
  // The store's fetch() records its failure instead of throwing.
  await store.fetch();
  if (store.error) emit("fail", store.error);
}
defineExpose({ refresh: fetchSources });

async function createSource() {
  try {
    await store.create({ ...form.value });
    form.value = blank(form.value.kind);
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

async function toggleEnabled(source) {
  try {
    await store.update(source.id, { enabled: !source.enabled });
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

async function deleteSource(source) {
  // Its leads stay: the server keeps them so that re-adding the source does
  // not bring back everything already dismissed.
  if (!confirm(`Stop watching ${source.label}? Leads already found are kept.`)) return;
  try {
    await store.remove(source.id);
    emit("ok");
  } catch (err) {
    emit("fail", err);
  }
}

function describe(source) {
  if (kindOf(source.kind).board) return source.token;
  return source.location ? `"${source.query}" in ${source.location}` : `"${source.query}"`;
}

function lastPoll(source) {
  if (!source.lastFetchedAt) return "not polled yet";
  const when = new Date(source.lastFetchedAt).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
  return `${source.lastCount} listed · ${when}`;
}

onMounted(fetchSources);
</script>

<template>
  <details class="jl-box" :open="!sources.length">
    <summary class="jl-box-heading">
      Sources
      <span class="jl-hint">{{ sources.filter((s) => s.enabled).length }} watched</span>
      <span v-if="sources.some((s) => s.lastError)" class="jl-warn">one or more failing</span>
    </summary>

    <div v-for="source in sources" :key="source.id" class="jl-source" :class="{ 'jl-source-off': !source.enabled }">
      <span class="jl-source-kind">{{ kindOf(source.kind).label }}</span>
      <span class="jl-source-label">{{ source.label }}</span>
      <span class="jl-source-what" :title="describe(source)">{{ describe(source) }}</span>
      <span v-if="source.lastError" class="jl-warn jl-source-status" :title="source.lastError">{{ source.lastError }}</span>
      <span v-else class="jl-hint jl-source-status">{{ lastPoll(source) }}</span>
      <button class="cv-btn cv-btn-sm" @click="toggleEnabled(source)">
        {{ source.enabled ? "Pause" : "Resume" }}
      </button>
      <button class="cv-btn cv-btn-sm cv-btn-danger" @click="deleteSource(source)">Delete</button>
    </div>
    <p v-if="!sources.length" class="jl-hint">
      Nothing is being watched. Add a company's job board, or a keyword search, below.
    </p>

    <form class="jl-source-form" @submit.prevent="createSource">
      <select v-model="form.kind" class="ja-input ja-select">
        <option v-for="k in KINDS" :key="k.value" :value="k.value">{{ k.label }}</option>
      </select>
      <template v-if="formKind.board">
        <input v-model="form.label" class="ja-input" placeholder="Company name *" required />
        <input v-model="form.token" class="ja-input" :placeholder="`Board slug * — ${formKind.hint}`" required />
      </template>
      <template v-else>
        <input v-model="form.query" class="ja-input" placeholder="Keywords * — e.g. golang backend" required />
        <input v-model="form.location" class="ja-input" placeholder="Location — e.g. London" />
      </template>
      <button type="submit" class="cv-btn cv-btn-primary">Add</button>
    </form>
    <p v-if="!formKind.board" class="jl-hint">
      {{ formKind.label }} needs its API key set under Admin → Environment → Job leads.
    </p>
  </details>
</template>

<style scoped>
/* The box and its heading (.jl-box, .jl-box-heading, .jl-hint, .jl-warn) are
   in JobLeads.vue's unscoped block, shared with JobSearchProfile.vue. */
.jl-source {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.3rem 0;
  font-size: 0.85rem;
}

.jl-source:first-of-type {
  margin-top: 0.75rem;
}

.jl-source-off {
  opacity: 0.55;
}

.jl-source-kind {
  flex: 0 0 6.5rem;
  color: var(--color-ink-muted);
}

.jl-source-label {
  flex: 0 0 9rem;
  font-weight: 500;
  color: var(--color-ink-soft);
}

.jl-source-what {
  flex: 1 1 6rem;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-ink-muted);
}

.jl-source-status {
  flex: 1 1 10rem;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
}

.jl-source-form {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .jl-source {
    flex-wrap: wrap;
  }

  .jl-source-status {
    flex-basis: 100%;
    text-align: left;
  }
}
</style>
