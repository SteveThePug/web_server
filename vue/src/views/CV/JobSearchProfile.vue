<script setup>
/**
 * The "Search profile" box on `/cv/leads`: what a lead is judged against.
 * Rendered only by JobLeads.vue; data comes from stores/jobSearchProfile.js.
 *
 * Two different things are edited here. The keyword lists drive a cheap
 * filter on the server that decides which postings are shown and scored at
 * all; the profile text is what Claude scores those against. Both take effect
 * on the next run, not on save — the parent says so beside "Check now".
 *
 * Like JobReferences.vue it has no error display of its own: failures are
 * emitted as `fail` and successes as `ok`, for the page's single banner.
 */
import { ref, computed, watch, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useJobSearchProfileStore } from "@/stores/jobSearchProfile";
import "./jobs-shared.css";

const emit = defineEmits(["fail", "ok"]);

const store = useJobSearchProfileStore();
const { profile } = storeToRefs(store);

const blank = () => ({ profile: "", includeKeywords: "", excludeKeywords: "", locations: "", minScore: 60 });
const form = ref(blank());
const saving = ref(false);

function reset() {
  form.value = profile.value
    ? {
        profile: profile.value.profile,
        includeKeywords: profile.value.includeKeywords,
        excludeKeywords: profile.value.excludeKeywords,
        locations: profile.value.locations,
        minScore: profile.value.minScore,
      }
    : blank();
}
watch(profile, reset);

const dirty = computed(
  () => !!profile.value && Object.keys(form.value).some((key) => form.value[key] !== profile.value[key]),
);

// Open by default until there is something in it: an empty profile means no
// lead is ever scored, which is worth making obvious.
const isEmpty = computed(() => !!profile.value && !profile.value.profile.trim());

async function save() {
  saving.value = true;
  try {
    await store.save({ ...form.value, minScore: Number(form.value.minScore) || 0 });
    emit("ok");
  } catch (err) {
    emit("fail", err);
  } finally {
    saving.value = false;
  }
}

onMounted(async () => {
  await store.fetch();
  if (store.error) emit("fail", store.error);
  reset();
});
</script>

<template>
  <details class="jl-box" :open="isEmpty">
    <summary class="jl-box-heading">
      Search profile
      <span v-if="isEmpty" class="jl-warn">empty — leads will not be scored</span>
    </summary>
    <form class="jl-profile" @submit.prevent="save">
      <label class="jl-field jl-field-wide">
        <span>About you and what you want</span>
        <textarea
          v-model="form.profile"
          class="ja-input jl-textarea"
          rows="8"
          placeholder="Experience, skills, the kind of role, level, location and working arrangement you are after. Each lead is scored against this text and nothing else."
        ></textarea>
      </label>
      <label class="jl-field">
        <span>Title must include one of</span>
        <textarea v-model="form.includeKeywords" class="ja-input jl-textarea" rows="4" placeholder="engineer&#10;developer"></textarea>
      </label>
      <label class="jl-field">
        <span>Title must not include</span>
        <textarea v-model="form.excludeKeywords" class="ja-input jl-textarea" rows="4" placeholder="intern&#10;principal"></textarea>
      </label>
      <label class="jl-field">
        <span>Location must include one of</span>
        <textarea v-model="form.locations" class="ja-input jl-textarea" rows="4" placeholder="london&#10;remote"></textarea>
      </label>
      <p class="jl-hint jl-field-wide">
        One entry per line; an empty list is no restriction. Title keywords match whole words, so
        "intern" does not exclude "Internal". Only leads that pass are shown and scored.
      </p>
      <div class="jl-profile-actions jl-field-wide">
        <label class="jl-inline">
          Hide leads scoring below
          <input v-model="form.minScore" class="ja-input jl-num" type="number" min="0" max="100" />
        </label>
        <button type="button" class="cv-btn cv-btn-sm" :disabled="!dirty || saving" @click="reset">Reset</button>
        <button type="submit" class="cv-btn cv-btn-sm cv-btn-primary" :disabled="!dirty || saving">
          {{ saving ? "Saving…" : "Save" }}
        </button>
      </div>
    </form>
  </details>
</template>

<style scoped>
/* The box and its heading (.jl-box, .jl-box-heading, .jl-hint, .jl-warn) are
   in JobLeads.vue's unscoped block, shared with JobSources.vue. */
.jl-profile {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.jl-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.8rem;
  color: var(--color-ink-muted);
}

.jl-field-wide {
  grid-column: 1 / -1;
}

.jl-textarea {
  resize: vertical;
  font-family: inherit;
  line-height: 1.4;
}

.jl-profile-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.jl-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-right: auto;
  font-size: 0.8rem;
  color: var(--color-ink-muted);
}

.jl-num {
  flex: 0 0 4.5rem;
  min-width: 0;
}

@media (max-width: 640px) {
  .jl-profile {
    grid-template-columns: 1fr;
  }
}
</style>
