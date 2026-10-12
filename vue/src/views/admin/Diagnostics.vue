<script setup>
/**
 * Admin panel: configuration checklist, the job capture token, and log tail.
 *
 * All of it comes from the backend as JSON (GET /api/admin/config,
 * /api/admin/capture-token, /api/admin/logs). Nothing here is GraphQL — the
 * data is process environment and filesystem, not the site's data model.
 *
 * The capture token is the credential the job capture userscript
 * (public/job-capture.user.js) sends. The server keeps only a hash, so the
 * token can be shown exactly once, in the response to generating it.
 *
 * The log tail loads lazily (only when the Logs tab is selected first)
 * because reading the file on every /admin visit would be waste; an admin
 * who remounts the component via tab switches re-triggers onMounted, which
 * is the desired "refresh on tab change" behaviour.
 */
import { ref, onMounted, computed } from "vue";
import axios from "axios";
import Button from "@/components/input/Button.vue";

const vars = ref([]);
const configError = ref("");

const lines = ref([]);
const logsError = ref("");
const loadingLogs = ref(false);

async function fetchConfig() {
  try {
    const res = await axios.get("/api/admin/config");
    vars.value = res.data.vars;
    configError.value = "";
  } catch (err) {
    configError.value = err.response?.data?.error || "Failed to load config";
  }
}

async function fetchLogs() {
  loadingLogs.value = true;
  try {
    const res = await axios.get("/api/admin/logs");
    lines.value = res.data.lines;
    logsError.value = "";
  } catch (err) {
    logsError.value = err.response?.data?.error || "Failed to load logs";
  } finally {
    loadingLogs.value = false;
  }
}

// Job capture token. `newToken` holds a just-generated token and lives only in
// this component's memory: leaving the tab is the end of it.
const capture = ref(null);
const captureError = ref("");
const captureBusy = ref(false);
const newToken = ref("");
const copied = ref(false);

async function fetchCaptureToken() {
  try {
    capture.value = (await axios.get("/api/admin/capture-token")).data;
    captureError.value = "";
  } catch (err) {
    captureError.value =
      err.response?.data?.error || "Failed to load the capture token state";
  }
}

async function generateCaptureToken() {
  // Generating replaces: an installed script stops working until it is
  // given the new token, which is worth a pause when one already exists.
  if (
    capture.value?.set &&
    !confirm(
      "Replace the existing capture token? The installed userscript will stop working until you paste the new one into it.",
    )
  )
    return;
  captureBusy.value = true;
  try {
    const res = await axios.post("/api/admin/capture-token");
    newToken.value = res.data.token;
    capture.value = { set: true, createdAt: res.data.createdAt };
    copied.value = false;
    captureError.value = "";
  } catch (err) {
    captureError.value =
      err.response?.data?.error || "Failed to generate a token";
  } finally {
    captureBusy.value = false;
  }
}

async function revokeCaptureToken() {
  if (!confirm("Revoke the capture token? The userscript will stop working."))
    return;
  captureBusy.value = true;
  try {
    capture.value = (await axios.delete("/api/admin/capture-token")).data;
    newToken.value = "";
    captureError.value = "";
  } catch (err) {
    captureError.value =
      err.response?.data?.error || "Failed to revoke the token";
  } finally {
    captureBusy.value = false;
  }
}

async function copyToken() {
  try {
    await navigator.clipboard.writeText(newToken.value);
    copied.value = true;
  } catch {
    // Clipboard access can be refused; the token is on screen to select.
    captureError.value = "Could not copy — select the token and copy it by hand.";
  }
}

const blank = computed(() =>
  vars.value.filter((v) => !v.set),
);

// Traffic view: a filter over the log tail for "who is using my site" —
// lines with a client IP (the formatter appends it only for real visitors
// and error statuses). Skips the standard line numbers of JSON-array output.
const trafficOnly = ref(true);
const shownLines = computed(() => {
  if (!trafficOnly.value) return lines.value;
  return lines.value.filter((l) => /\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/.test(l.trim()) || l.includes("!!! SERVER ERROR") || l.includes(" !!"));
});

onMounted(() => {
  fetchConfig();
  fetchCaptureToken();
  fetchLogs();
});
</script>

<template>
  <div class="diagnostics">
    <h1>Configuration</h1>
    <p v-if="blank.length === 0" class="ok">All known configuration set.</p>
    <p v-else class="warn">
      {{ blank.length }} unset. Blank vars with <em>required</em> break their
      feature; optional ones degrade it.
    </p>
    <p v-if="configError" class="error">{{ configError }}</p>

    <table v-if="vars.length">
      <thead>
        <tr>
          <th>Variable</th>
          <th>For</th>
          <th>State</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="v in vars" :key="v.name">
          <td><code>{{ v.name }}</code></td>
          <td>{{ v.purpose }}</td>
          <td>
            <span v-if="!v.set" class="warn">unset</span>
            <span v-else-if="v.secret">set (secret)</span>
            <span v-else>{{ v.value }}</span>
          </td>
        </tr>
      </tbody>
    </table>

    <h1>Job capture token</h1>
    <p class="hint">
      The credential for the
      <a href="/job-capture.user.js" target="_blank" rel="noopener"
        >job capture userscript</a
      >. Generate one, then paste it into the script's
      <em>Configure job capture</em> menu command. Only a hash is stored, so a
      token is shown once and cannot be recovered — generate again to replace
      it.
    </p>
    <p v-if="captureError" class="error">{{ captureError }}</p>
    <template v-if="capture">
      <p v-if="capture.set">
        A token is active, generated
        {{ new Date(capture.createdAt).toLocaleString() }}.
      </p>
      <p v-else class="warn">
        No token — the capture script cannot reach the tracker.
      </p>
      <div v-if="newToken" class="token-box">
        <p class="hint">Copy this now. It will not be shown again.</p>
        <code class="token">{{ newToken }}</code>
        <Button @click="copyToken">{{ copied ? "Copied" : "Copy" }}</Button>
      </div>
      <div class="token-controls">
        <Button :disabled="captureBusy" @click="generateCaptureToken">
          {{ capture.set ? "Regenerate" : "Generate token" }}
        </Button>
        <Button
          v-if="capture.set"
          :disabled="captureBusy"
          @click="revokeCaptureToken"
        >
          Revoke
        </Button>
      </div>
    </template>

    <h1>Logs</h1>
    <p class="hint">
      Tail of the backend log (up to 500 lines): HTTP traffic and application
      messages, newest last.
    </p>
    <div class="log-controls">
      <label class="hint toggle">
        <input type="checkbox" v-model="trafficOnly" /> visitors only
      </label>
      <Button :disabled="loadingLogs" @click="fetchLogs">
        {{ loadingLogs ? "Loading…" : "Refresh" }}
      </Button>
    </div>
    <p v-if="logsError" class="error">{{ logsError }}</p>
    <p v-else-if="shownLines.length === 0" class="hint">
      No {{ trafficOnly ? "visitor" : "" }} lines in the current tail.
    </p>
    <pre v-else class="log-output">{{ shownLines.join("\n") }}</pre>
  </div>
</template>

<style scoped>
.diagnostics {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}
th,
td {
  text-align: left;
  padding: 4px 8px;
  border-bottom: 1px solid var(--color-quaternary);
}

.ok {
  color: var(--color-tertiary);
}
.warn {
  color: var(--color-error, #b91c1c);
}
.error {
  color: var(--color-error, #b91c1c);
}
.hint {
  font-size: 0.85rem;
  opacity: 0.75;
}

.token-box {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--color-quaternary);
}
.token-box .hint {
  flex-basis: 100%;
  margin: 0;
}
.token {
  flex: 1;
  min-width: 0;
  font-size: 0.8rem;
  word-break: break-all;
  user-select: all;
}
.token-controls {
  display: flex;
  gap: 12px;
}

.log-controls {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  align-items: center;
}

.toggle {
  display: flex;
  gap: 4px;
  align-items: center;
  cursor: pointer;
}

.log-output {
  max-height: 400px;
  overflow-y: auto;
  padding: 8px;
  background: var(--color-surface-tint, #111);
  color: var(--color-primary);
  font-size: 0.75rem;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  border: 1px solid var(--color-quaternary);
}
</style>
