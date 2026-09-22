<script setup>
/**
 * Admin panel: configuration checklist and log tail.
 *
 * Both come from the backend as JSON (GET /api/admin/config, GET
 * /api/admin/logs). Nothing here is GraphQL — the data is process
 * environment and filesystem, not the site's data model.
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

const blank = computed(() =>
  vars.value.filter((v) => !v.set),
);

onMounted(() => {
  fetchConfig();
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

    <h1>Logs</h1>
    <p class="hint">
      Tail of the backend log (up to 500 lines): HTTP traffic and application
      messages, newest last.
    </p>
    <div class="log-controls">
      <Button :disabled="loadingLogs" @click="fetchLogs">
        {{ loadingLogs ? "Loading…" : "Refresh" }}
      </Button>
    </div>
    <p v-if="logsError" class="error">{{ logsError }}</p>
    <pre v-else class="log-output">{{ lines.join("\n") }}</pre>
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

.log-controls {
  display: flex;
  justify-content: flex-end;
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
