<script setup>
/**
 * Admin panel: tail of the backend log.
 *
 * GET /api/admin/logs returns JSON read from the filesystem, not GraphQL —
 * it is not part of the site's data model. Configuration and the job capture
 * token, which used to share this tab, are on the Environment tab.
 *
 * An admin who remounts the component via tab switches re-triggers
 * onMounted, which is the desired "refresh on tab change" behaviour.
 */
import { ref, onMounted, computed } from "vue";
import { api } from "@/api";
import Button from "@/components/input/Button.vue";

const lines = ref([]);
const logsError = ref("");
const loadingLogs = ref(false);

async function fetchLogs() {
  loadingLogs.value = true;
  try {
    const res = await api.get("/admin/logs", {
      fallback: "Failed to load logs",
    });
    lines.value = res.lines;
    logsError.value = "";
  } catch (err) {
    logsError.value = err.message;
  } finally {
    loadingLogs.value = false;
  }
}

// Traffic view: a filter over the log tail for "who is using my site" —
// lines with a client IP (the formatter appends it only for real visitors
// and error statuses). Skips the standard line numbers of JSON-array output.
const trafficOnly = ref(true);
const shownLines = computed(() => {
  if (!trafficOnly.value) return lines.value;
  return lines.value.filter((l) => /\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/.test(l.trim()) || l.includes("!!! SERVER ERROR") || l.includes(" !!"));
});

onMounted(fetchLogs);
</script>

<template>
  <div class="diagnostics">
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
