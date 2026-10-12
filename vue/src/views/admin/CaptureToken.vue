<script setup>
/**
 * The job capture token card on the admin Environment tab.
 *
 * The token is the credential the job capture userscript
 * (public/job-capture.user.js) sends. It is not a .env variable: the server
 * keeps only a hash in backend/token/, so it can be shown exactly once, in
 * the response to generating it, and replacing it needs no restart.
 */
import { ref, onMounted } from "vue";
import axios from "axios";
import Button from "@/components/input/Button.vue";
import SecretReveal from "@/components/admin/SecretReveal.vue";
import "./env.css";

// `newToken` holds a just-generated token and lives only in this component's
// memory: leaving the tab is the end of it.
const capture = ref(null);
const error = ref("");
const busy = ref(false);
const newToken = ref("");

async function fetchCaptureToken() {
  try {
    capture.value = (await axios.get("/api/admin/capture-token")).data;
    error.value = "";
  } catch (err) {
    error.value =
      err.response?.data?.error || "Failed to load the capture token state";
  }
}

async function generate() {
  // Generating replaces: an installed script stops working until it is
  // given the new token, which is worth a pause when one already exists.
  if (
    capture.value?.set &&
    !confirm(
      "Replace the existing capture token? The installed userscript will stop working until you paste the new one into it.",
    )
  )
    return;
  busy.value = true;
  try {
    const res = await axios.post("/api/admin/capture-token");
    newToken.value = res.data.token;
    capture.value = { set: true, createdAt: res.data.createdAt };
    error.value = "";
  } catch (err) {
    error.value = err.response?.data?.error || "Failed to generate a token";
  } finally {
    busy.value = false;
  }
}

async function revoke() {
  if (!confirm("Revoke the capture token? The userscript will stop working."))
    return;
  busy.value = true;
  try {
    capture.value = (await axios.delete("/api/admin/capture-token")).data;
    newToken.value = "";
    error.value = "";
  } catch (err) {
    error.value = err.response?.data?.error || "Failed to revoke the token";
  } finally {
    busy.value = false;
  }
}

onMounted(fetchCaptureToken);
</script>

<template>
  <section class="env-card">
    <header class="env-card-head">
      <h2>Job capture</h2>
      <span v-if="capture" class="env-pill" :class="capture.set ? 'is-set' : 'is-unset'">
        {{ capture.set ? "active" : "no token" }}
      </span>
    </header>
    <div class="capture-body">
      <p class="env-hint">
        The credential for the
        <a href="/job-capture.user.js" target="_blank" rel="noopener"
          >job capture userscript</a
        >. Generate one, then paste it into the script's
        <em>Configure job capture</em> menu command. Only a hash is stored, so
        it takes effect at once and cannot be recovered — generate again to
        replace it.
      </p>
      <p v-if="error" class="env-error">{{ error }}</p>
      <template v-if="capture">
        <p v-if="capture.set" class="env-hint">
          Generated {{ new Date(capture.createdAt).toLocaleString() }}.
        </p>
        <SecretReveal v-if="newToken" :value="newToken" />
        <div class="capture-controls">
          <Button class="px-2" :disabled="busy" @click="generate">
            {{ capture.set ? "Regenerate" : "Generate token" }}
          </Button>
          <Button v-if="capture.set" class="px-2" :disabled="busy" @click="revoke">
            Revoke
          </Button>
        </div>
      </template>
    </div>
  </section>
</template>

<style scoped>
.capture-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 10px;
}
.capture-controls {
  display: flex;
  gap: 12px;
}
</style>
