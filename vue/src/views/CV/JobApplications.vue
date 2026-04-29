<script setup>
import { ref, onMounted } from "vue";
import { RouterLink } from "vue-router";
import { gql } from "@/graphql";

const applications = ref([]);
const editingId = ref(null);
const editForm = ref({});
const form = ref({
  jobTitle: "",
  company: "",
  location: "",
  url: "",
  status: "Applied",
  notes: "",
  appliedAt: "",
});

const references = ref([]);
const refForm = ref({ category: "profile", label: "", value: "" });
const editingRefId = ref(null);
const editRefForm = ref({});
const REF_CATEGORIES = ["profile", "experience"];
const REF_FIELDS = `id category label value sortOrder createdAt`;

const STATUS_OPTIONS = [
  "Applied",
  "Screening",
  "Interview",
  "Offer",
  "Rejected",
  "Withdrawn",
];

const APP_FIELDS = `id jobTitle company location url status notes appliedAt createdAt`;

async function fetchApplications() {
  try {
    const data = await gql(`query { jobApplications { ${APP_FIELDS} } }`);
    applications.value = data.jobApplications;
  } catch (err) {
    console.error(err);
  }
}

async function createApplication() {
  if (!form.value.jobTitle || !form.value.company || !form.value.status) return;
  try {
    const input = {
      jobTitle: form.value.jobTitle,
      company: form.value.company,
      status: form.value.status,
      location: form.value.location || undefined,
      url: form.value.url || undefined,
      notes: form.value.notes || undefined,
      appliedAt: form.value.appliedAt
        ? new Date(form.value.appliedAt).toISOString()
        : undefined,
    };
    const data = await gql(
      `mutation CreateJobApplication($input: CreateJobApplicationInput!) {
                createJobApplication(input: $input) { ${APP_FIELDS} }
            }`,
      { input },
    );
    applications.value.unshift(data.createJobApplication);
    form.value = {
      jobTitle: "",
      company: "",
      location: "",
      url: "",
      status: "Applied",
      notes: "",
      appliedAt: "",
    };
  } catch (err) {
    console.error(err);
  }
}

function startEdit(app) {
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

function cancelEdit() {
  editingId.value = null;
  editForm.value = {};
}

async function saveEdit(id) {
  try {
    const input = {
      jobTitle: editForm.value.jobTitle || undefined,
      company: editForm.value.company || undefined,
      status: editForm.value.status || undefined,
      location: editForm.value.location || undefined,
      url: editForm.value.url || undefined,
      notes: editForm.value.notes || undefined,
      appliedAt: editForm.value.appliedAt
        ? new Date(editForm.value.appliedAt).toISOString()
        : undefined,
    };
    const data = await gql(
      `mutation UpdateJobApplication($id: ID!, $input: UpdateJobApplicationInput!) {
                updateJobApplication(id: $id, input: $input) { ${APP_FIELDS} }
            }`,
      { id, input },
    );
    const idx = applications.value.findIndex((a) => a.id === id);
    if (idx !== -1) applications.value[idx] = data.updateJobApplication;
    editingId.value = null;
  } catch (err) {
    console.error(err);
  }
}

async function deleteApplication(id) {
  try {
    await gql(
      `mutation DeleteJobApplication($id: ID!) { deleteJobApplication(id: $id) }`,
      { id },
    );
    applications.value = applications.value.filter((a) => a.id !== id);
  } catch (err) {
    console.error(err);
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
  const csv = [headers, ...rows].map((r) => r.map(escape).join(",")).join("\n");
  const blob = new Blob([csv], { type: "text/csv" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "job_applications.csv";
  a.click();
  URL.revokeObjectURL(url);
}

async function fetchReferences() {
  try {
    const data = await gql(`query { jobAppReferences { ${REF_FIELDS} } }`);
    references.value = data.jobAppReferences;
  } catch (err) {
    console.error(err);
  }
}

async function createReference() {
  if (!refForm.value.label || !refForm.value.value) return;
  try {
    const input = {
      category: refForm.value.category,
      label: refForm.value.label,
      value: refForm.value.value,
    };
    const data = await gql(
      `mutation CreateJobAppReference($input: CreateJobAppReferenceInput!) {
                createJobAppReference(input: $input) { ${REF_FIELDS} }
            }`,
      { input },
    );
    references.value.push(data.createJobAppReference);
    refForm.value = { category: refForm.value.category, label: "", value: "" };
  } catch (err) {
    console.error(err);
  }
}

function startRefEdit(ref) {
  editingRefId.value = ref.id;
  editRefForm.value = {
    category: ref.category,
    label: ref.label,
    value: ref.value,
  };
}

function cancelRefEdit() {
  editingRefId.value = null;
  editRefForm.value = {};
}

async function saveRefEdit(id) {
  try {
    const input = {
      category: editRefForm.value.category || undefined,
      label: editRefForm.value.label || undefined,
      value: editRefForm.value.value || undefined,
    };
    const data = await gql(
      `mutation UpdateJobAppReference($id: ID!, $input: UpdateJobAppReferenceInput!) {
                updateJobAppReference(id: $id, input: $input) { ${REF_FIELDS} }
            }`,
      { id, input },
    );
    const idx = references.value.findIndex((r) => r.id === id);
    if (idx !== -1) references.value[idx] = data.updateJobAppReference;
    editingRefId.value = null;
  } catch (err) {
    console.error(err);
  }
}

async function deleteReference(id) {
  try {
    await gql(
      `mutation DeleteJobAppReference($id: ID!) { deleteJobAppReference(id: $id) }`,
      { id },
    );
    references.value = references.value.filter((r) => r.id !== id);
  } catch (err) {
    console.error(err);
  }
}

function refsByCategory(category) {
  return references.value.filter((r) => r.category === category);
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text);
}

function statusClass(status) {
  const map = {
    Applied: "status-applied",
    Screening: "status-screening",
    Interview: "status-interview",
    Offer: "status-offer",
    Rejected: "status-rejected",
    Withdrawn: "status-withdrawn",
  };
  return map[status] ?? "";
}

onMounted(() => {
  fetchApplications();
  fetchReferences();
});
</script>

<template>
  <div class="ja-root">
    <div class="ja-header">
      <div class="ja-header-left">
        <RouterLink to="/cv" class="ja-back">← CV</RouterLink>
        <h2 class="ja-heading">Job Applications</h2>
      </div>
      <button
        class="ja-btn"
        @click="exportCsv"
        :disabled="!applications.length"
      >
        Export CSV
      </button>
    </div>

    <div class="ja-ref-section">
      <h3 class="ja-ref-heading">Quick Reference</h3>
      <div v-for="cat in REF_CATEGORIES" :key="cat" class="ja-ref-category">
        <h4 class="ja-ref-cat-label">{{ cat }}</h4>
        <div
          v-for="ref in refsByCategory(cat)"
          :key="ref.id"
          class="ja-ref-item"
        >
          <template v-if="editingRefId !== ref.id">
            <span class="ja-ref-label">{{ ref.label }}</span>
            <span class="ja-ref-value" :title="ref.value">{{ ref.value }}</span>
            <button
              class="ja-btn ja-btn-sm"
              @click="copyToClipboard(ref.value)"
              title="Copy"
            >
              Copy
            </button>
            <button class="ja-btn ja-btn-sm" @click="startRefEdit(ref)">
              Edit
            </button>
            <button
              class="ja-btn ja-btn-sm ja-btn-danger"
              @click="deleteReference(ref.id)"
            >
              Delete
            </button>
          </template>
          <template v-else>
            <select
              v-model="editRefForm.category"
              class="ja-input ja-input-sm ja-select"
            >
              <option v-for="c in REF_CATEGORIES" :key="c" :value="c">
                {{ c }}
              </option>
            </select>
            <input
              v-model="editRefForm.label"
              class="ja-input ja-input-sm"
              placeholder="Label"
            />
            <input
              v-model="editRefForm.value"
              class="ja-input ja-input-sm"
              placeholder="Value"
            />
            <button
              class="ja-btn ja-btn-sm ja-btn-primary"
              @click="saveRefEdit(ref.id)"
            >
              Save
            </button>
            <button class="ja-btn ja-btn-sm" @click="cancelRefEdit">
              Cancel
            </button>
          </template>
        </div>
        <p v-if="!refsByCategory(cat).length" class="ja-ref-empty">
          No {{ cat }} items yet.
        </p>
      </div>
      <form class="ja-ref-form" @submit.prevent="createReference">
        <select v-model="refForm.category" class="ja-input ja-select">
          <option v-for="c in REF_CATEGORIES" :key="c" :value="c">
            {{ c }}
          </option>
        </select>
        <input
          v-model="refForm.label"
          class="ja-input"
          placeholder="Label *"
          required
        />
        <input
          v-model="refForm.value"
          class="ja-input"
          placeholder="Value *"
          required
        />
        <button type="submit" class="ja-btn ja-btn-primary">Add</button>
      </form>
    </div>

    <form class="ja-form" @submit.prevent="createApplication">
      <div class="ja-form-row">
        <input
          v-model="form.jobTitle"
          class="ja-input"
          placeholder="Job title *"
          required
        />
        <input
          v-model="form.company"
          class="ja-input"
          placeholder="Company *"
          required
        />
        <select v-model="form.status" class="ja-input ja-select">
          <option v-for="s in STATUS_OPTIONS" :key="s" :value="s">
            {{ s }}
          </option>
        </select>
      </div>
      <div class="ja-form-row">
        <input
          v-model="form.location"
          class="ja-input"
          placeholder="Location"
        />
        <input v-model="form.url" class="ja-input" placeholder="URL" />
        <input
          v-model="form.appliedAt"
          class="ja-input"
          type="date"
          title="Applied date"
        />
      </div>
      <div class="ja-form-row">
        <textarea
          v-model="form.notes"
          class="ja-input ja-textarea"
          placeholder="Notes"
        />
        <button type="submit" class="ja-btn ja-btn-primary">Add</button>
      </div>
    </form>

    <table class="ja-table" v-if="applications.length">
      <thead>
        <tr>
          <th>Title</th>
          <th>Company</th>
          <th>Status</th>
          <th>Location</th>
          <th>Applied</th>
          <th>Notes</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <template v-for="app in applications" :key="app.id">
          <tr v-if="editingId !== app.id">
            <td>
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
            <td>{{ app.company }}</td>
            <td>
              <span :class="['ja-badge', statusClass(app.status)]">{{
                app.status
              }}</span>
            </td>
            <td>{{ app.location ?? "—" }}</td>
            <td>{{ app.appliedAt ? app.appliedAt.substring(0, 10) : "—" }}</td>
            <td class="ja-notes-cell">{{ app.notes ?? "" }}</td>
            <td class="ja-actions">
              <button class="ja-btn ja-btn-sm" @click="startEdit(app)">
                Edit
              </button>
              <button
                class="ja-btn ja-btn-sm ja-btn-danger"
                @click="deleteApplication(app.id)"
              >
                Delete
              </button>
            </td>
          </tr>
          <tr v-else class="ja-edit-row">
            <td>
              <input
                v-model="editForm.jobTitle"
                class="ja-input ja-input-sm"
                placeholder="Job title"
              />
            </td>
            <td>
              <input
                v-model="editForm.company"
                class="ja-input ja-input-sm"
                placeholder="Company"
              />
            </td>
            <td>
              <select
                v-model="editForm.status"
                class="ja-input ja-input-sm ja-select"
              >
                <option v-for="s in STATUS_OPTIONS" :key="s" :value="s">
                  {{ s }}
                </option>
              </select>
            </td>
            <td>
              <input
                v-model="editForm.location"
                class="ja-input ja-input-sm"
                placeholder="Location"
              />
            </td>
            <td>
              <input
                v-model="editForm.appliedAt"
                class="ja-input ja-input-sm"
                type="date"
              />
            </td>
            <td>
              <input
                v-model="editForm.notes"
                class="ja-input ja-input-sm"
                placeholder="Notes"
              />
            </td>
            <td class="ja-actions">
              <button
                class="ja-btn ja-btn-sm ja-btn-primary"
                @click="saveEdit(app.id)"
              >
                Save
              </button>
              <button class="ja-btn ja-btn-sm" @click="cancelEdit">
                Cancel
              </button>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
    <p v-else class="ja-empty">No applications yet.</p>
  </div>
</template>

<style scoped>
.ja-root {
  padding: 1.5rem;
  border-top: 2px solid #333;
  background: #fafafa;
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

.ja-back {
  font-size: 0.85rem;
  color: #555;
  text-decoration: none;
}

.ja-back:hover {
  color: #111;
}

.ja-heading {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.ja-form {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.ja-form-row {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.ja-input {
  padding: 0.35rem 0.6rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 0.85rem;
  background: white;
  flex: 1;
  min-width: 120px;
}

.ja-input:focus {
  outline: none;
  border-color: #555;
}

.ja-select {
  cursor: pointer;
}

.ja-textarea {
  resize: vertical;
  min-height: 2.4rem;
}

.ja-input-sm {
  padding: 0.2rem 0.4rem;
  font-size: 0.8rem;
  width: 100%;
  min-width: 0;
}

.ja-btn {
  padding: 0.35rem 0.8rem;
  border: 1px solid #333;
  border-radius: 4px;
  background: white;
  color: #333;
  cursor: pointer;
  font-size: 0.85rem;
  white-space: nowrap;
  transition:
    background 0.15s,
    color 0.15s;
}

.ja-btn:hover {
  background: #eee;
}

.ja-btn-sm {
  padding: 0.2rem 0.55rem;
  font-size: 0.8rem;
}

.ja-btn-primary {
  background: #333;
  color: white;
}

.ja-btn-primary:hover {
  background: #555;
}

.ja-btn-danger {
  border-color: #c00;
  color: #c00;
}

.ja-btn-danger:hover {
  background: #fee;
}

.ja-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.ja-table th,
.ja-table td {
  text-align: left;
  padding: 0.45rem 0.6rem;
  border-bottom: 1px solid #e0e0e0;
  vertical-align: middle;
}

.ja-table th {
  font-weight: 600;
  color: #555;
  background: #f0f0f0;
}

.ja-table tr:hover td {
  background: #f5f5f5;
}

.ja-edit-row td {
  padding: 0.3rem 0.4rem;
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

.ja-link {
  color: #0066cc;
  text-decoration: none;
}

.ja-link:hover {
  text-decoration: underline;
}

.ja-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 10px;
  font-size: 0.78rem;
  font-weight: 500;
  background: #e0e0e0;
  color: #333;
}

.status-applied {
  background: #dbeafe;
  color: #1e40af;
}
.status-screening {
  background: #fef9c3;
  color: #854d0e;
}
.status-interview {
  background: #ede9fe;
  color: #5b21b6;
}
.status-offer {
  background: #dcfce7;
  color: #166534;
}
.status-rejected {
  background: #fee2e2;
  color: #991b1b;
}
.status-withdrawn {
  background: #f3f4f6;
  color: #6b7280;
}

.ja-empty {
  color: #888;
  font-size: 0.9rem;
}

.ja-ref-section {
  margin-bottom: 1.5rem;
  padding: 1rem;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: #fff;
}

.ja-ref-heading {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.75rem 0;
  color: #333;
}

.ja-ref-category {
  margin-bottom: 0.75rem;
}

.ja-ref-cat-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #555;
  text-transform: capitalize;
  margin: 0 0 0.35rem 0;
}

.ja-ref-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.3rem 0;
}

.ja-ref-label {
  font-size: 0.85rem;
  font-weight: 500;
  color: #333;
  min-width: 80px;
}

.ja-ref-value {
  font-size: 0.85rem;
  color: #555;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ja-ref-empty {
  font-size: 0.8rem;
  color: #999;
  margin: 0.2rem 0;
}

.ja-ref-form {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
  flex-wrap: wrap;
}
</style>
