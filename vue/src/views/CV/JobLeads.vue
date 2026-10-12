<script setup>
/**
 * Route `/cv/leads` (guarded by `meta.requiresAdmin`) — the leads inbox:
 * postings the server found on its own, waiting to be tracked or dismissed.
 *
 * The server side is backend/services/job_leads.go. On a timer, or when
 * "Check now" is pressed here, it polls the sources (JobSources.vue), keeps
 * the postings that pass the keyword filter and has Claude score them against
 * the profile (JobSearchProfile.vue). Nothing reaches the tracker at /cv/jobs
 * until **Track** is pressed on a lead, which adds it there as Saved.
 *
 * This view owns the page frame, the single error banner, the run button and
 * the list. Data access is in stores/jobLeads.js; nothing here calls gql().
 *
 * A run takes from seconds to a few minutes, so "Check now" does not wait for
 * it: the server starts the run and this page polls its status, then reloads
 * the list when it finishes.
 */
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { RouterLink } from "vue-router";
import { storeToRefs } from "pinia";
import { useJobLeadsStore } from "@/stores/jobLeads";
import { useJobSearchProfileStore } from "@/stores/jobSearchProfile";
import JobSources from "./JobSources.vue";
import JobSearchProfile from "./JobSearchProfile.vue";
import "./jobs-shared.css";

const TABS = [
  { value: "new", label: "Inbox" },
  { value: "promoted", label: "Tracked" },
  { value: "dismissed", label: "Dismissed" },
];
// Labels for JobLead.sourceKind; a COPY of the kinds in JobSources.vue.
const SOURCE_LABELS = { greenhouse: "Greenhouse", lever: "Lever", ashby: "Ashby", adzuna: "Adzuna", reed: "Reed" };
const POLL_MS = 3000;

const store = useJobLeadsStore();
const { leads, status, loading } = storeToRefs(store);
const { profile } = storeToRefs(useJobSearchProfileStore());

const error = ref("");
const tab = ref("new");
const search = ref("");
// Two separate "show me the rest" switches: leads scored below the profile's
// minimum (already loaded, hidden here), and leads the keyword filter
// rejected (not loaded at all until asked for — that is most of every board).
const showLowScores = ref(false);
const showUnmatched = ref(false);
const busyId = ref(null);
const sourcesBox = ref(null);

// The posting text for the one lead whose description is open.
// `undefined` in the cache means "not fetched", `""` means "fetched, none".
const descriptionId = ref(null);
const descriptions = ref({});

function fail(err) {
  console.error(err);
  error.value = String(err?.message ?? err);
}

function clearError() {
  error.value = "";
}

async function fetchLeads() {
  // The store's fetch() records its failure instead of throwing.
  await store.fetch({ includeUnmatched: showUnmatched.value });
  if (store.error) fail(store.error);
}

const minScore = computed(() => profile.value?.minScore ?? 0);

/** An unscored lead is never "low": no score is not a bad score. */
const isLow = (lead) => lead.score != null && lead.score < minScore.value;

const counts = computed(() => {
  const out = { new: 0, promoted: 0, dismissed: 0 };
  for (const lead of leads.value) if (lead.state in out) out[lead.state]++;
  return out;
});

const inTab = computed(() => leads.value.filter((lead) => lead.state === tab.value));

const hiddenLow = computed(() =>
  tab.value === "new" && !showLowScores.value ? inTab.value.filter(isLow).length : 0,
);

// The server already sorts best-first, unscored after scored, newest first
// within a score, so this only filters.
const visibleLeads = computed(() => {
  const needle = search.value.trim().toLowerCase();
  return inTab.value.filter((lead) => {
    if (tab.value === "new" && !showLowScores.value && isLow(lead)) return false;
    if (!needle) return true;
    return [lead.jobTitle, lead.company, lead.location, lead.scoreReasons]
      .some((field) => field?.toLowerCase().includes(needle));
  });
});

function scoreClass(lead) {
  if (lead.score == null) return "jl-score-none";
  if (lead.score >= 80) return "jl-score-high";
  if (lead.score >= 60) return "jl-score-mid";
  return "jl-score-low";
}

function ageLabel(lead) {
  const at = lead.postedAt ?? lead.createdAt;
  if (!at) return "";
  const days = Math.floor((Date.now() - new Date(at).getTime()) / 86_400_000);
  const age = days <= 0 ? "today" : days === 1 ? "1 day ago" : `${days} days ago`;
  return lead.postedAt ? `posted ${age}` : `found ${age}`;
}

async function act(lead, action) {
  busyId.value = lead.id;
  try {
    await store[action](lead.id);
    if (descriptionId.value === lead.id) descriptionId.value = null;
    error.value = "";
  } catch (err) {
    fail(err);
  } finally {
    busyId.value = null;
  }
}

async function toggleDescription(lead) {
  if (descriptionId.value === lead.id) {
    descriptionId.value = null;
    return;
  }
  descriptionId.value = lead.id;
  if (descriptions.value[lead.id] !== undefined) return;
  try {
    descriptions.value[lead.id] = await store.fetchDescription(lead.id);
  } catch (err) {
    descriptionId.value = null;
    fail(err);
  }
}

async function toggleUnmatched() {
  showUnmatched.value = !showUnmatched.value;
  await fetchLeads();
}

// Run and poll

let pollTimer = null;

function stopPolling() {
  clearTimeout(pollTimer);
  pollTimer = null;
}

/** Poll until the run finishes, then reload what it changed. */
async function pollStatus() {
  const current = await store.fetchStatus();
  if (current?.running) {
    pollTimer = setTimeout(pollStatus, POLL_MS);
    return;
  }
  stopPolling();
  if (current?.lastError) fail(current.lastError);
  await Promise.all([fetchLeads(), sourcesBox.value?.refresh()]);
}

async function checkNow() {
  try {
    // false means a run was already going (the scheduler's, say); either way
    // there is now one to wait for.
    await store.runSync();
    error.value = "";
    status.value = { ...(status.value ?? {}), running: true };
    stopPolling();
    pollTimer = setTimeout(pollStatus, POLL_MS);
  } catch (err) {
    fail(err);
  }
}

const running = computed(() => !!status.value?.running);

const lastRun = computed(() => {
  if (running.value) return "Checking sources and scoring…";
  if (!status.value?.lastRunAt) return "";
  const when = new Date(status.value.lastRunAt).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
  return `Last checked ${when}`;
});

onMounted(async () => {
  await fetchLeads();
  // Arriving mid-run (the scheduler's, or one started before a reload) should
  // look the same as having pressed the button.
  const current = await store.fetchStatus();
  if (current?.running) pollTimer = setTimeout(pollStatus, POLL_MS);
});

onBeforeUnmount(stopPolling);
</script>

<template>
  <div class="ja-root jl-root">
    <div class="jl-header">
      <div class="jl-header-left">
        <RouterLink to="/cv/jobs" class="jl-back">← Tracker</RouterLink>
        <h2 class="jl-heading">Job Leads</h2>
      </div>
      <div class="jl-header-right">
        <span class="jl-hint">{{ lastRun }}</span>
        <button
          class="cv-btn cv-btn-primary"
          :disabled="running"
          title="Poll every source now, then score new leads. Profile and keyword changes take effect on a run."
          @click="checkNow"
        >
          {{ running ? "Checking…" : "Check now" }}
        </button>
      </div>
    </div>

    <div v-if="error" class="jl-error" role="alert">
      <span>{{ error }}</span>
      <button class="cv-btn cv-btn-sm" @click="error = ''">Dismiss</button>
    </div>

    <JobSearchProfile @fail="fail" @ok="clearError" />
    <JobSources ref="sourcesBox" @fail="fail" @ok="clearError" />

    <div class="jl-tabs">
      <button
        v-for="t in TABS"
        :key="t.value"
        class="jl-tab"
        :class="{ 'jl-tab-active': tab === t.value }"
        @click="tab = t.value"
      >
        <span class="jl-tab-count">{{ counts[t.value] }}</span>
        <span class="jl-tab-label">{{ t.label }}</span>
      </button>
    </div>

    <div class="jl-filterbar">
      <input v-model="search" class="ja-input" type="search" placeholder="Search title, company, location, reasons…" />
      <label v-if="tab === 'new'" class="jl-check">
        <input v-model="showLowScores" type="checkbox" />
        Below {{ minScore }}<template v-if="hiddenLow"> ({{ hiddenLow }} hidden)</template>
      </label>
      <label v-if="tab === 'new'" class="jl-check" title="Postings the keyword filter rejected. They are never scored.">
        <input :checked="showUnmatched" type="checkbox" @change="toggleUnmatched" />
        Filtered out
      </label>
    </div>

    <article
      v-for="lead in visibleLeads"
      :key="lead.id"
      class="jl-lead"
      :class="{ 'jl-lead-unmatched': !lead.matched && lead.state === 'new' }"
    >
      <div class="jl-score" :class="scoreClass(lead)" :title="lead.score == null ? 'Not scored' : `Fit score ${lead.score} out of 100`">
        {{ lead.score ?? "–" }}
      </div>
      <div class="jl-lead-body">
        <div class="jl-lead-title">
          <a v-if="lead.url" :href="lead.url" target="_blank" rel="noopener" class="jl-link">{{ lead.jobTitle }}</a>
          <span v-else>{{ lead.jobTitle }}</span>
          <span class="jl-lead-company">{{ lead.company }}</span>
        </div>
        <div class="jl-lead-meta">
          <span v-if="lead.location">{{ lead.location }}</span>
          <span>{{ SOURCE_LABELS[lead.sourceKind] ?? lead.sourceKind }}</span>
          <span>{{ ageLabel(lead) }}</span>
          <span v-if="!lead.matched && lead.state === 'new'">filtered out</span>
        </div>
        <p v-if="lead.scoreReasons" class="jl-lead-reasons">{{ lead.scoreReasons }}</p>
        <pre v-if="descriptionId === lead.id" class="jl-desc">{{
          descriptions[lead.id] === undefined ? "Loading…" : descriptions[lead.id] || "The source gave no description."
        }}</pre>
      </div>
      <div class="jl-lead-actions">
        <button class="cv-btn cv-btn-sm" @click="toggleDescription(lead)">
          {{ descriptionId === lead.id ? "Hide" : "Posting" }}
        </button>
        <template v-if="lead.state === 'new'">
          <button
            class="cv-btn cv-btn-sm cv-btn-primary"
            :disabled="busyId === lead.id"
            title="Add to the tracker as Saved"
            @click="act(lead, 'promote')"
          >
            Track
          </button>
          <button class="cv-btn cv-btn-sm" :disabled="busyId === lead.id" @click="act(lead, 'dismiss')">Dismiss</button>
        </template>
        <button v-else-if="lead.state === 'dismissed'" class="cv-btn cv-btn-sm" :disabled="busyId === lead.id" @click="act(lead, 'restore')">
          Restore
        </button>
        <RouterLink v-else to="/cv/jobs" class="cv-btn cv-btn-sm jl-btn-link">In tracker</RouterLink>
      </div>
    </article>

    <p v-if="!visibleLeads.length" class="jl-empty">
      <template v-if="loading">Loading…</template>
      <template v-else-if="tab !== 'new'">Nothing here.</template>
      <template v-else-if="hiddenLow">Nothing at or above {{ minScore }}. Tick "Below {{ minScore }}" to see the rest.</template>
      <template v-else-if="!leads.length">No leads yet. Add a source above, then press "Check now".</template>
      <template v-else>The inbox is empty.</template>
    </p>
  </div>
</template>

<!-- Unscoped on purpose, for the same reason as jobs-shared.css: the two child
     components render the boxes these classes style, and a scoped rule stops
     at a child's root element. Every selector is under .jl-root. -->
<style>
.jl-root .jl-box {
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 6px;
  background: var(--color-paper);
}

.jl-root .jl-box-heading {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-ink-soft);
  cursor: pointer;
}

.jl-root .jl-box-heading > span {
  margin-left: 0.6rem;
  font-weight: 400;
}

.jl-root .jl-hint {
  font-size: 0.8rem;
  color: var(--color-ink-faint);
}

.jl-root p.jl-hint {
  margin: 0.5rem 0 0;
}

.jl-root .jl-warn {
  font-size: 0.8rem;
  color: var(--color-ink-danger);
}
</style>

<style scoped>
/* Colours are the paper palette from CVLayout.vue; buttons are .cv-btn from
   there too, and the inputs are styled by jobs-shared.css. The frame matches
   the tracker page's (.ja-root and friends in JobApplications.vue), whose
   rules are scoped to that component and so are repeated here. */
.jl-root {
  padding: 1.5rem;
  border-top: 2px solid var(--color-ink-soft);
  background: var(--color-paper-tint);
}

.jl-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.jl-header-left,
.jl-header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.jl-back {
  font-size: 0.85rem;
  color: var(--color-ink-muted);
  text-decoration: none;
}

.jl-back:hover {
  color: var(--color-ink);
}

.jl-heading {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-ink-soft);
}

.jl-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid #b91c1c;
  border-radius: 4px;
  background: #fef2f2;
  color: #7f1d1d;
  font-size: 0.85rem;
}

.jl-tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.jl-tab {
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  padding: 0.4rem 0.9rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 6px;
  background: var(--color-paper);
  color: var(--color-ink-muted);
  cursor: pointer;
}

.jl-tab-active {
  border-color: var(--color-ink-soft);
  color: var(--color-ink);
}

.jl-tab-count {
  font-size: 1.05rem;
  font-weight: 600;
}

.jl-tab-label {
  font-size: 0.8rem;
}

.jl-filterbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.jl-check {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: var(--color-ink-muted);
  white-space: nowrap;
  cursor: pointer;
}

.jl-lead {
  display: flex;
  align-items: flex-start;
  gap: 0.9rem;
  margin-bottom: 0.5rem;
  padding: 0.7rem 0.9rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 6px;
  background: var(--color-paper);
}

.jl-lead-unmatched {
  opacity: 0.6;
}

/* The score reads as a number first; the tint only groups it. Same tints as
   the tracker's status pills. */
.jl-score {
  flex: 0 0 2.6rem;
  padding: 0.35rem 0;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  text-align: center;
  color: var(--color-ink-soft);
  background: var(--color-paper-shade);
}

.jl-score-high {
  background: #dcfce7;
  color: #14532d;
}

.jl-score-mid {
  background: #fef9c3;
  color: #713f12;
}

.jl-score-none {
  color: var(--color-ink-faint);
}

.jl-lead-body {
  flex: 1;
  min-width: 0;
}

.jl-lead-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-ink);
}

.jl-lead-company {
  margin-left: 0.5rem;
  font-weight: 400;
  color: var(--color-ink-muted);
}

.jl-link {
  color: var(--color-ink-link);
  text-decoration: none;
}

.jl-link:hover {
  text-decoration: underline;
}

.jl-lead-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 0.9rem;
  margin-top: 0.15rem;
  font-size: 0.8rem;
  color: var(--color-ink-faint);
}

.jl-lead-reasons {
  margin: 0.4rem 0 0;
  font-size: 0.85rem;
  color: var(--color-ink-soft);
}

/* The posting text is long, so it scrolls inside the card. */
.jl-desc {
  max-height: 22rem;
  overflow: auto;
  margin: 0.6rem 0 0;
  padding: 0.6rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 4px;
  background: var(--color-paper-tint);
  font-family: inherit;
  font-size: 0.8rem;
  white-space: pre-wrap;
}

.jl-lead-actions {
  display: flex;
  gap: 0.4rem;
  white-space: nowrap;
}

.jl-btn-link {
  text-decoration: none;
}

.jl-empty {
  color: var(--color-ink-faint);
  font-size: 0.9rem;
}

@media (max-width: 640px) {
  .jl-header,
  .jl-filterbar {
    flex-wrap: wrap;
  }

  .jl-lead {
    flex-wrap: wrap;
  }

  .jl-lead-body {
    flex-basis: calc(100% - 3.5rem);
  }

  .jl-lead-actions {
    flex-basis: 100%;
  }
}
</style>
