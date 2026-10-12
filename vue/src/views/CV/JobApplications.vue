<script setup>
/**
 * Route `/cv/jobs` (guarded by `meta.requiresAdmin`) — a private tracker for job
 * applications and for the reusable reference snippets that feed the CV copy.
 *
 * This view owns the page frame, the single error banner and the applications
 * table (pipeline bar, filter/sort, inline editing via an `editingId` plus a
 * scratch copy of the row). The two other pieces are components beside it:
 * JobReferences.vue (the Quick Reference box, a separate table) and
 * JobApplicationForm.vue (the add form). Data access is in
 * stores/jobApplications.js and stores/jobAppReferences.js; nothing here calls
 * gql() directly.
 *
 * Rows also arrive from two places that are not this page: the email pipeline,
 * and the browser userscript in `public/job-capture.user.js`, which the header
 * links to for installing.
 */
import { ref, computed, onMounted } from "vue";
import { RouterLink } from "vue-router";
import { storeToRefs } from "pinia";
import { useJobApplicationsStore } from "@/stores/jobApplications";
import { STATUS_OPTIONS, OPEN_STATUSES } from "./jobStatus";
import JobReferences from "./JobReferences.vue";
import JobApplicationForm from "./JobApplicationForm.vue";
import "./jobs-shared.css";

const store = useJobApplicationsStore();
const { applications } = storeToRefs(store);
const editingId = ref(null);
const editForm = ref({});
const error = ref("");

const STALE_DAYS = 14;

// The captured posting text for the one row whose description is open.
// `undefined` in the cache means "not fetched", `""` means "fetched, none saved".
const descriptionId = ref(null);
const descriptions = ref({});

// Filter/sort state. All of it runs over the already-loaded array — the query
// has no arguments and the dataset is one person's applications.
const search = ref("");
const statusFilter = ref(null);
const sortKey = ref("updatedAt");
const sortDir = ref("desc");

function fail(err) {
  console.error(err);
  // The realistic failure here is a lapsed 7-day cookie: the router guard only
  // checks the in-memory admin flag, so every mutation returns "admin access
  // required" while the page still looks signed in.
  error.value = String(err?.message ?? err);
}

function clearError() {
  error.value = "";
}

async function fetchApplications() {
  // The store's fetch() records its failure instead of throwing.
  await store.fetch();
  if (store.error) fail(store.error);
}

function startEdit(app) {
  // Switching rows mid-edit would throw away unsaved typing silently, and the
  // Edit button sits in every row, so the misclick is easy to make.
  if (editingId.value !== null && editingId.value !== app.id && isEditDirty()) {
    if (!confirm("Discard unsaved changes to the row you are editing?")) return;
  }
  editingId.value = app.id;
  editForm.value = {
    jobTitle: app.jobTitle,
    company: app.company,
    location: app.location ?? "",
    url: app.url ?? "",
    status: app.status,
    notes: app.notes ?? "",
    appliedAt: app.appliedAt ? app.appliedAt.substring(0, 10) : "",
  };
}

function isEditDirty() {
  const app = applications.value.find((a) => a.id === editingId.value);
  if (!app) return false;
  const e = editForm.value;
  return (
    e.jobTitle !== app.jobTitle ||
    e.company !== app.company ||
    e.location !== (app.location ?? "") ||
    e.url !== (app.url ?? "") ||
    e.status !== app.status ||
    e.notes !== (app.notes ?? "") ||
    e.appliedAt !== (app.appliedAt ? app.appliedAt.substring(0, 10) : "")
  );
}

function cancelEdit() {
  editingId.value = null;
  editForm.value = {};
}

async function saveEdit(id) {
  try {
    // Optional fields send `''` rather than `undefined` when emptied: the
    // resolver skips fields that are absent, so `|| undefined` would turn
    // "delete this wrong URL" into a silent no-op that re-renders the old
    // value. appliedAt has no empty representation, so clearing it is not
    // offered — the field stays as it was.
    await store.update(id, {
      jobTitle: editForm.value.jobTitle,
      company: editForm.value.company,
      status: editForm.value.status,
      location: editForm.value.location ?? "",
      url: editForm.value.url ?? "",
      notes: editForm.value.notes ?? "",
      appliedAt: editForm.value.appliedAt
        ? new Date(editForm.value.appliedAt).toISOString()
        : undefined,
    });
    editingId.value = null;
    error.value = "";
  } catch (err) {
    // editingId is deliberately left set: the row stays open with the typing
    // intact so the save can be retried.
    fail(err);
  }
}

/** Status is the field that changes most, so it is editable straight from the pill. */
async function setStatus(app, status) {
  if (status === app.status) return;
  try {
    await store.update(app.id, { status });
    error.value = "";
  } catch (err) {
    fail(err);
  }
}

async function toggleDescription(app) {
  if (descriptionId.value === app.id) {
    descriptionId.value = null;
    return;
  }
  descriptionId.value = app.id;
  if (descriptions.value[app.id] !== undefined) return;
  try {
    descriptions.value[app.id] = await store.fetchDescription(app.id);
  } catch (err) {
    descriptionId.value = null;
    fail(err);
  }
}

async function deleteApplication(app) {
  // Delete sits next to Edit in a 0.4rem-gap row and the notes it takes with
  // it are not recoverable through any UI path.
  if (!confirm(`Delete the ${app.jobTitle} application at ${app.company}?`))
    return;
  try {
    await store.remove(app.id);
    error.value = "";
  } catch (err) {
    fail(err);
  }
}

function exportCsv() {
  const headers = [
    "Job Title",
    "Company",
    "Status",
    "Location",
    "URL",
    "Applied",
    "Notes",
    "Created",
  ];
  const rows = applications.value.map((a) => [
    a.jobTitle,
    a.company,
    a.status,
    a.location ?? "",
    a.url ?? "",
    a.appliedAt ? a.appliedAt.substring(0, 10) : "",
    a.notes ?? "",
    a.createdAt ? a.createdAt.substring(0, 10) : "",
  ]);
  const escape = (v) => `"${String(v).replace(/"/g, '""')}"`;
  // CRLF joins and a BOM because the destination for an exported tracker is
  // Excel, which otherwise reads the UTF-8 as the local ANSI codepage (so "£"
  // and accents arrive mangled) and mis-splits LF-only rows.
  const csv =
    "\ufeff" +
    [headers, ...rows].map((r) => r.map(escape).join(",")).join("\r\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "job_applications.csv";
  a.click();
  URL.revokeObjectURL(url);
}

function statusLabel(status) {
  return (
    STATUS_OPTIONS.find((s) => s.value === status?.toLowerCase())?.label ??
    status
  );
}

function statusClass(status) {
  return `ja-badge-${status?.toLowerCase() ?? "unknown"}`;
}

/** Whole days between a timestamp and now; null when there is no date. */
function daysSince(iso) {
  if (!iso) return null;
  return Math.floor((Date.now() - new Date(iso).getTime()) / 86400000);
}

function ageLabel(app) {
  // createdAt is the fallback: an application logged without an applied date
  // still has a meaningful age.
  const days = daysSince(app.appliedAt ?? app.createdAt);
  return days === null ? "—" : days === 0 ? "today" : `${days}d`;
}

/** A row worth chasing: still open, and nothing has moved it in a fortnight. */
function isStale(app) {
  if (!OPEN_STATUSES.includes(app.status?.toLowerCase())) return false;
  const days = daysSince(app.updatedAt ?? app.createdAt);
  return days !== null && days >= STALE_DAYS;
}

const statusCounts = computed(() => {
  const counts = {};
  for (const app of applications.value) {
    const key = app.status?.toLowerCase();
    counts[key] = (counts[key] ?? 0) + 1;
  }
  return counts;
});

const visibleApplications = computed(() => {
  const q = search.value.trim().toLowerCase();
  const rows = applications.value.filter((app) => {
    if (statusFilter.value && app.status?.toLowerCase() !== statusFilter.value)
      return false;
    if (!q) return true;
    // Notes are searched because the email pipeline appends its findings
    // there, making them the de-facto history of the application.
    return [app.jobTitle, app.company, app.location, app.notes]
      .filter(Boolean)
      .some((field) => field.toLowerCase().includes(q));
  });

  const dir = sortDir.value === "asc" ? 1 : -1;
  return [...rows].sort((a, b) => {
    const av = a[sortKey.value] ?? "";
    const bv = b[sortKey.value] ?? "";
    // Rows missing the sort field go last in either direction rather than
    // clustering at whichever end empty-string happens to sort to.
    if (!av && bv) return 1;
    if (av && !bv) return -1;
    if (av === bv) return 0;
    return av > bv ? dir : -dir;
  });
});

function toggleSort(key) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
  } else {
    sortKey.value = key;
    sortDir.value = "desc";
  }
}

function toggleStatusFilter(status) {
  statusFilter.value = statusFilter.value === status ? null : status;
}

onMounted(fetchApplications);
</script>

<template>
  <div class="ja-root">
    <div class="ja-header">
      <div class="ja-header-left">
        <RouterLink to="/cv" class="ja-back">← CV</RouterLink>
        <h2 class="ja-heading">Job Applications</h2>
      </div>
      <div class="ja-header-right">
        <!-- A plain link to a static file: Violentmonkey intercepts any
             navigation to a URL ending in .user.js and offers to install it.
             The token it then asks for is generated on the admin panel. -->
        <a
          class="cv-btn"
          href="/job-capture.user.js"
          target="_blank"
          rel="noopener"
          title="Violentmonkey userscript that adds the job you are viewing. After installing, generate a capture token on the admin panel and paste it into the script's menu: Configure job capture."
        >
          Install capture script
        </a>
        <button
          class="cv-btn"
          @click="exportCsv"
          :disabled="!applications.length"
        >
          Export CSV
        </button>
      </div>
    </div>

    <div v-if="error" class="ja-error" role="alert">
      <span>{{ error }}</span>
      <button class="cv-btn cv-btn-sm" @click="error = ''">Dismiss</button>
    </div>

    <JobReferences @fail="fail" @ok="clearError" />

    <JobApplicationForm @fail="fail" @ok="clearError" />

    <!-- Pipeline bar: counts at a glance, each one a filter toggle. -->
    <div class="ja-pipeline" v-if="applications.length">
      <button
        v-for="s in STATUS_OPTIONS"
        :key="s.value"
        class="ja-pipe-tile"
        :class="{ 'ja-pipe-active': statusFilter === s.value }"
        @click="toggleStatusFilter(s.value)"
      >
        <span class="ja-pipe-count">{{ statusCounts[s.value] ?? 0 }}</span>
        <span class="ja-pipe-label">{{ s.label }}</span>
      </button>
    </div>

    <div class="ja-filterbar" v-if="applications.length">
      <input
        v-model="search"
        class="ja-input"
        type="search"
        placeholder="Search title, company, location, notes…"
      />
      <span class="ja-count"
        >{{ visibleApplications.length }} / {{ applications.length }}</span
      >
    </div>

    <table class="ja-table" v-if="visibleApplications.length">
      <thead>
        <tr>
          <th class="ja-sortable" @click="toggleSort('jobTitle')">Title</th>
          <th class="ja-sortable" @click="toggleSort('company')">Company</th>
          <th class="ja-sortable" @click="toggleSort('status')">Status</th>
          <th>Location</th>
          <th class="ja-sortable" @click="toggleSort('appliedAt')">Applied</th>
          <th>Notes</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <template v-for="app in visibleApplications" :key="app.id">
          <tr v-if="editingId !== app.id" :class="{ 'ja-stale': isStale(app) }">
            <td data-label="Title">
              <a
                v-if="app.url"
                :href="app.url"
                target="_blank"
                rel="noopener"
                class="ja-link"
                >{{ app.jobTitle }}</a
              >
              <span v-else>{{ app.jobTitle }}</span>
            </td>
            <td data-label="Company">{{ app.company }}</td>
            <td data-label="Status">
              <select
                class="ja-badge ja-badge-select"
                :class="statusClass(app.status)"
                :value="app.status?.toLowerCase()"
                @change="setStatus(app, $event.target.value)"
              >
                <option
                  v-for="s in STATUS_OPTIONS"
                  :key="s.value"
                  :value="s.value"
                >
                  {{ s.label }}
                </option>
                <!-- A status the pipeline wrote that we no longer list still
                     needs to render rather than showing a blank select. -->
                <option
                  v-if="!STATUS_OPTIONS.some((s) => s.value === app.status)"
                  :value="app.status"
                >
                  {{ statusLabel(app.status) }}
                </option>
              </select>
            </td>
            <td data-label="Location">{{ app.location || "—" }}</td>
            <td
              data-label="Applied"
              :title="app.appliedAt ? app.appliedAt.substring(0, 10) : ''"
            >
              {{ ageLabel(app) }}
            </td>
            <td class="ja-notes-cell" data-label="Notes" :title="app.notes">
              {{ app.notes ?? "" }}
            </td>
            <td class="ja-actions">
              <button
                class="cv-btn cv-btn-sm"
                title="Posting text saved when the job was captured"
                @click="toggleDescription(app)"
              >
                {{ descriptionId === app.id ? "Hide" : "Posting" }}
              </button>
              <button class="cv-btn cv-btn-sm" @click="startEdit(app)">
                Edit
              </button>
              <button
                class="cv-btn cv-btn-sm cv-btn-danger"
                @click="deleteApplication(app)"
              >
                Delete
              </button>
            </td>
          </tr>
          <tr v-else class="ja-edit-row">
            <td data-label="Title">
              <input
                v-model="editForm.jobTitle"
                class="ja-input ja-input-sm"
                placeholder="Job title"
              />
              <input
                v-model="editForm.url"
                class="ja-input ja-input-sm"
                placeholder="URL"
              />
            </td>
            <td data-label="Company">
              <input
                v-model="editForm.company"
                class="ja-input ja-input-sm"
                placeholder="Company"
              />
            </td>
            <td data-label="Status">
              <select
                v-model="editForm.status"
                class="ja-input ja-input-sm ja-select"
              >
                <option
                  v-for="s in STATUS_OPTIONS"
                  :key="s.value"
                  :value="s.value"
                >
                  {{ s.label }}
                </option>
              </select>
            </td>
            <td data-label="Location">
              <input
                v-model="editForm.location"
                class="ja-input ja-input-sm"
                placeholder="Location"
              />
            </td>
            <td data-label="Applied">
              <input
                v-model="editForm.appliedAt"
                class="ja-input ja-input-sm"
                type="date"
              />
            </td>
            <td data-label="Notes">
              <!-- Textarea, not an input: the pipeline appends newline-joined
                   findings here, so notes are routinely multi-line. -->
              <textarea
                v-model="editForm.notes"
                class="ja-input ja-input-sm ja-edit-notes"
                placeholder="Notes"
              />
            </td>
            <td class="ja-actions">
              <button
                class="cv-btn cv-btn-sm cv-btn-primary"
                @click="saveEdit(app.id)"
              >
                Save
              </button>
              <button class="cv-btn cv-btn-sm" @click="cancelEdit">
                Cancel
              </button>
            </td>
          </tr>
          <tr v-if="descriptionId === app.id" class="ja-desc-row">
            <td colspan="7">
              <p v-if="descriptions[app.id] === undefined" class="ja-empty">
                Loading…
              </p>
              <pre v-else-if="descriptions[app.id]" class="ja-desc">{{
                descriptions[app.id]
              }}</pre>
              <p v-else class="ja-empty">
                No posting text saved — only jobs added with the capture script
                have one.
              </p>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
    <p v-else-if="applications.length" class="ja-empty">
      No applications match this filter.
    </p>
    <p v-else class="ja-empty">No applications yet.</p>
  </div>
</template>

<style scoped>
/* Colours are the paper palette from CVLayout.vue; buttons are .cv-btn from
   there too. The input classes (.ja-input, .ja-input-sm, .ja-select) are in
   jobs-shared.css because the child components use them as well. */
.ja-root {
  padding: 1.5rem;
  border-top: 2px solid var(--color-ink-soft);
  background: var(--color-paper-tint);
}

.ja-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.ja-header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.ja-header-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.ja-back {
  font-size: 0.85rem;
  color: var(--color-ink-muted);
  text-decoration: none;
}

.ja-back:hover {
  color: var(--color-ink);
}

.ja-heading {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-ink-soft);
}

.ja-error {
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

.ja-edit-notes {
  resize: vertical;
  min-height: 3.5rem;
}

/* Pipeline bar */
.ja-pipeline {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}

.ja-pipe-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
  padding: 0.4rem 0.7rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 4px;
  background: var(--color-paper);
  cursor: pointer;
  min-width: 68px;
}

.ja-pipe-tile:hover {
  background: var(--color-paper-shade);
}

.ja-pipe-active {
  border-color: var(--color-ink);
  background: var(--color-paper-shade);
}

.ja-pipe-count {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-ink-soft);
}

.ja-pipe-label {
  font-size: 0.7rem;
  color: var(--color-ink-muted);
}

.ja-filterbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.ja-count {
  font-size: 0.8rem;
  color: var(--color-ink-faint);
  white-space: nowrap;
}

/* Applications table */
.ja-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.ja-table th,
.ja-table td {
  text-align: left;
  padding: 0.45rem 0.6rem;
  border-bottom: 1px solid var(--color-line-soft);
  vertical-align: middle;
}

.ja-table th {
  font-weight: 600;
  color: var(--color-ink-muted);
  background: var(--color-paper-shade);
}

.ja-sortable {
  cursor: pointer;
  user-select: none;
}

.ja-sortable:hover {
  color: var(--color-ink);
}

.ja-table tr:hover td {
  background: var(--color-paper-shade);
}

/* Open for a fortnight with nothing moving — the follow-up queue. */
.ja-stale td:first-child {
  box-shadow: inset 3px 0 0 var(--color-ink-muted);
}

.ja-edit-row td {
  padding: 0.3rem 0.4rem;
  vertical-align: top;
}

.ja-actions {
  display: flex;
  gap: 0.4rem;
  white-space: nowrap;
}

.ja-notes-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* The posting text is long, so it scrolls inside the row instead of pushing
   the rest of the table off the screen. */
.ja-desc {
  max-height: 22rem;
  overflow: auto;
  margin: 0;
  padding: 0.6rem;
  border: 1px solid var(--color-line-soft);
  border-radius: 4px;
  background: var(--color-paper);
  font-family: inherit;
  font-size: 0.8rem;
  white-space: pre-wrap;
}

.ja-desc-row:hover td {
  background: transparent;
}

.ja-link {
  color: var(--color-ink-link);
  text-decoration: none;
}

.ja-link:hover {
  text-decoration: underline;
}

/* Status pill, doubling as an inline select. Palette is CVLayout's paper
   tokens with a per-status tint, rather than off-theme Tailwind colours. */
.ja-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 10px;
  font-size: 0.78rem;
  font-weight: 500;
  border: 1px solid var(--color-line);
  background: var(--color-paper-shade);
  color: var(--color-ink-soft);
}

.ja-badge-select {
  cursor: pointer;
  appearance: none;
}

.ja-badge-saved {
  background: var(--color-paper);
  color: var(--color-ink-muted);
  border-style: dashed;
}
.ja-badge-applied {
  background: #eef2ff;
  color: #3730a3;
}
.ja-badge-screening {
  background: #fefce8;
  color: #854d0e;
}
.ja-badge-assessment {
  background: #fff7ed;
  color: #9a3412;
}
.ja-badge-interviewing {
  background: #f5f3ff;
  color: #5b21b6;
}
.ja-badge-offer {
  background: #f0fdf4;
  color: #166534;
}
.ja-badge-rejected {
  background: #fef2f2;
  color: #991b1b;
}
.ja-badge-withdrawn {
  background: var(--color-paper-shade);
  color: var(--color-ink-faint);
}

.ja-empty {
  color: var(--color-ink-faint);
  font-size: 0.9rem;
}

/* Seven columns plus an action cell will not fit a phone, so the table becomes
   stacked cards with the header text carried by each cell's data-label. */
@media (max-width: 640px) {
  .ja-table thead {
    display: none;
  }

  .ja-table,
  .ja-table tbody,
  .ja-table tr,
  .ja-table td {
    display: block;
    width: 100%;
  }

  .ja-table tr {
    margin-bottom: 0.75rem;
    border: 1px solid var(--color-line-soft);
    border-radius: 6px;
    background: var(--color-paper);
  }

  .ja-table td {
    border-bottom: none;
    padding: 0.3rem 0.6rem;
  }

  .ja-table td[data-label]::before {
    content: attr(data-label);
    display: block;
    font-size: 0.7rem;
    text-transform: uppercase;
    color: var(--color-ink-faint);
  }

  .ja-notes-cell {
    max-width: none;
    white-space: pre-wrap;
    overflow: visible;
  }

  .ja-stale td:first-child {
    box-shadow: inset 0 3px 0 var(--color-ink-muted);
  }
}
</style>
