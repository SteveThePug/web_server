<script setup>
/**
 * The "add an application" form on `/cv/jobs`. Rendered only by
 * JobApplications.vue; it creates through stores/jobApplications.js, so the new
 * row appears in the parent's table without the two having to talk.
 *
 * Like JobReferences.vue it reports through the parent's single error banner:
 * `fail` carries the error, `ok` says a create succeeded.
 *
 * Its root sits inside the parent's `.ja-root`, which jobs-shared.css relies on.
 */
import { ref } from "vue";
import { useJobApplicationsStore } from "@/stores/jobApplications";
import { STATUS_OPTIONS } from "./jobStatus";
import "./jobs-shared.css";

const emit = defineEmits(["fail", "ok"]);

const store = useJobApplicationsStore();

function today() {
  return new Date().toISOString().substring(0, 10);
}

function blankForm() {
  return {
    jobTitle: "",
    company: "",
    location: "",
    url: "",
    status: "applied",
    notes: "",
    // Logging an application on the day you send it is the common case, so the
    // date defaults to today rather than landing null and ageing as unknown.
    appliedAt: today(),
  };
}

const form = ref(blankForm());

async function createApplication() {
  if (!form.value.jobTitle.trim() || !form.value.company.trim()) return;
  try {
    await store.create({
      jobTitle: form.value.jobTitle.trim(),
      company: form.value.company.trim(),
      status: form.value.status,
      location: form.value.location || undefined,
      url: form.value.url || undefined,
      notes: form.value.notes || undefined,
      appliedAt: form.value.appliedAt
        ? new Date(form.value.appliedAt).toISOString()
        : undefined,
    });
    form.value = blankForm();
    emit("ok");
  } catch (err) {
    // The form is left as typed so the create can be retried.
    emit("fail", err);
  }
}
</script>

<template>
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
        <option v-for="s in STATUS_OPTIONS" :key="s.value" :value="s.value">
          {{ s.label }}
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
      <button type="submit" class="cv-btn cv-btn-primary">Add</button>
    </div>
  </form>
</template>

<style scoped>
/* Inputs are styled by jobs-shared.css; buttons are .cv-btn from CVLayout.vue. */
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

.ja-textarea {
  resize: vertical;
  min-height: 2.4rem;
}
</style>
