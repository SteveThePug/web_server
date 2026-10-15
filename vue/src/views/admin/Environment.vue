<script setup>
/**
 * Admin Environment tab: every variable in .env, grouped, with the ones that
 * are safe to change editable in place.
 *
 * Nothing here writes .env directly, and neither does the backend. Applying
 * sends the staged edits to PUT /api/admin/env, which queues them for a
 * script on the host (env-apply/apply.py) that rewrites the file and
 * recreates the containers that read the changed variables. This component
 * then polls GET /api/admin/env/status until the host reports back — through
 * the backend's own restart, which is why failed polls are expected and
 * ignored rather than shown.
 *
 * Secrets are write-only: the server sends "set" and at most the last four
 * characters, never the value. A generated secret is the one exception, shown
 * once from the response that created it.
 */
import { ref, reactive, computed, onMounted, onUnmounted } from "vue";
import { api } from "@/api";
import Button from "@/components/input/Button.vue";
import SecretReveal from "@/components/admin/SecretReveal.vue";
import ConfirmDialog from "@/components/util/ConfirmDialog.vue";
import "./env.css";

const groups = ref([]);
const status = ref(null);
const error = ref("");
const loading = ref(true);
const applyConfirmationOpen = ref(false);
const applyConfirmation = ref({ title: "Apply changes", message: "", confirmLabel: "Apply" });

// Staged edits, none of which exist anywhere but this component until Apply.
//   drafts:     name -> text in the row's input
//   generating: names to have the server mint a value for
//   clearing:   secrets to blank (a secret's empty input means "unchanged",
//               so clearing one needs saying explicitly)
const drafts = reactive({});
const generating = reactive(new Set());
const clearing = reactive(new Set());

// Values the server generated in the last apply, shown once.
const revealed = ref({});

const applying = ref(false);
const outcome = ref(null); // { ok, text, output }
const awaitedId = ref("");
let pollTimer = null;

const allVars = computed(() => groups.value.flatMap((g) => g.vars));

function resetDrafts() {
  for (const v of allVars.value) drafts[v.name] = v.secret ? "" : v.value || "";
  generating.clear();
  clearing.clear();
}

async function fetchEnv() {
  try {
    const res = await api.get("/admin/env", {
      fallback: "Failed to load the environment",
    });
    groups.value = res.groups;
    status.value = res.status;
    error.value = "";
    resetDrafts();
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

// What a row would send: null when it is untouched.
function changeFor(v) {
  if (!v.editable) return null;
  if (generating.has(v.name)) return { name: v.name, type: "generate" };
  if (clearing.has(v.name)) return { name: v.name, type: "clear" };
  const draft = (drafts[v.name] ?? "").trim();
  if (v.secret) return draft ? { name: v.name, type: "set", value: draft } : null;
  if (draft === (v.value || "")) return null;
  return { name: v.name, type: draft ? "set" : "clear", value: draft };
}

const changes = computed(() =>
  allVars.value.map((v) => ({ v, change: changeFor(v) })).filter((c) => c.change),
);

// Mirrors EnvManager.ServicesFor: the backend is always recreated.
const affectedServices = computed(() => {
  const services = new Set(["backend"]);
  for (const { v } of changes.value)
    for (const s of v.services || []) services.add(s);
  return [...services].sort();
});

const pendingNames = computed(() => new Set(status.value?.pending || []));

function toggleGenerate(v) {
  if (generating.has(v.name)) return generating.delete(v.name);
  generating.add(v.name);
  clearing.delete(v.name);
  drafts[v.name] = "";
}

function toggleClear(v) {
  if (clearing.has(v.name)) return clearing.delete(v.name);
  clearing.add(v.name);
  generating.delete(v.name);
  drafts[v.name] = "";
}

// Typing a value withdraws a staged generate or clear for the same row.
function onInput(v) {
  generating.delete(v.name);
  clearing.delete(v.name);
}

function pillFor(v) {
  if (changeFor(v)) return { text: "changed", cls: "is-changed" };
  if (pendingNames.value.has(v.name)) return { text: "pending", cls: "is-changed" };
  if (!v.set) return { text: "unset", cls: "is-unset" };
  if (v.secret) return { text: v.hint ? `set ••••${v.hint}` : "set", cls: "is-set" };
  return { text: "set", cls: "is-set" };
}

async function apply() {
  const warnings = changes.value.filter((c) => c.v.warning).map((c) => `• ${c.v.warning}`);
  const lines = [
    `Apply ${changes.value.length} change(s)?`,
    `This recreates: ${affectedServices.value.join(", ")}. The site's API is unavailable for a few seconds.`,
    ...warnings,
  ];
  applyConfirmation.value = { title: "Apply environment changes", message: lines.join("\n"), confirmLabel: "Apply" };
  applyConfirmationOpen.value = true;
}

async function confirmApply() {
  applyConfirmationOpen.value = false;

  const body = { set: {}, generate: [] };
  for (const { change } of changes.value) {
    if (change.type === "generate") body.generate.push(change.name);
    else body.set[change.name] = change.type === "clear" ? "" : change.value;
  }

  applying.value = true;
  outcome.value = null;
  try {
    const res = await api.put("/admin/env", body, {
      fallback: "Failed to queue the change",
    });
    revealed.value = res.generated || {};
    status.value = res.status;
    awaitedId.value = res.status.id;
    // Staged secrets have been sent; do not keep them in the inputs.
    resetDrafts();
    startPolling();
  } catch (err) {
    applying.value = false;
    outcome.value = { ok: false, text: err.message };
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(poll, 2000);
}
function stopPolling() {
  if (pollTimer) clearInterval(pollTimer);
  pollTimer = null;
}

async function poll() {
  let next;
  try {
    next = await api.get("/admin/env/status");
  } catch (err) {
    // 401/403 after an apply means the change invalidated this session
    // (rotating BACKEND_SECRET does). Anything else is the backend
    // restarting, which is the apply working.
    if ([401, 403].includes(err.status)) {
      stopPolling();
      applying.value = false;
      outcome.value = {
        ok: true,
        text: "Applied. Your session was signed with the old secret — log in again.",
      };
    }
    return;
  }
  status.value = next;
  if (next.id !== awaitedId.value) return;
  if (next.state === "applied") {
    stopPolling();
    applying.value = false;
    outcome.value = {
      ok: true,
      text: `Applied. Recreated ${(next.services || []).join(", ")}.`,
    };
    await fetchEnv();
  } else if (next.state === "failed") {
    stopPolling();
    applying.value = false;
    outcome.value = { ok: false, text: "The host could not apply the change.", output: next.output };
    await fetchEnv();
  }
}

async function cancelPending() {
  try {
    status.value = await api.delete("/admin/env/pending", {
      fallback: "Failed to withdraw the change",
    });
    stopPolling();
    applying.value = false;
    outcome.value = { ok: true, text: "Pending change withdrawn. Nothing was applied." };
  } catch (err) {
    outcome.value = { ok: false, text: err.message };
  }
}

const progressText = computed(() => {
  const s = status.value;
  if (!s) return "";
  if (s.state === "pending") return "Queued — waiting for the host to pick it up…";
  if (s.state === "applying")
    return `Recreating ${(s.services || []).join(", ")}…`;
  return "";
});

onMounted(async () => {
  await fetchEnv();
  // A change queued before this tab was opened is still worth following.
  if (["pending", "applying"].includes(status.value?.state)) {
    awaitedId.value = status.value.id;
    applying.value = true;
    startPolling();
  }
});
onUnmounted(stopPolling);
</script>

<template>
  <div class="env">
    <h1>Environment</h1>
    <p class="env-hint">
      The site's <code>.env</code>. Edits are staged here, then applied on the
      server, which restarts only the services that read what changed. Secrets
      can be replaced but never read back.
    </p>
    <p v-if="error" class="env-error">{{ error }}</p>
    <p v-if="loading" class="env-hint">Loading…</p>

    <div v-if="Object.keys(revealed).length" class="env-revealed">
      <SecretReveal
        v-for="(value, name) in revealed"
        :key="name"
        :label="name"
        :value="value"
      />
    </div>

    <div v-if="applying" class="env-banner">
      <p>{{ progressText || "Applying…" }}</p>
      <template v-if="status?.state === 'pending' && !status.watcher">
        <p class="env-hint">
          Nothing on the host has picked this up. On the Pi, run
          <code>./env-apply/install.sh</code> once to install the watcher. To
          apply this change by hand, from the repo root:
        </p>
        <code class="env-command">python3 env-apply/apply.py</code>
        <Button class="px-2" @click="cancelPending">Withdraw change</Button>
      </template>
    </div>
    <div v-else-if="outcome" class="env-banner" :class="outcome.ok ? 'is-ok' : 'is-failed'">
      <p>{{ outcome.text }}</p>
      <pre v-if="outcome.output" class="env-output">{{ outcome.output }}</pre>
    </div>

    <section v-for="group in groups" :key="group.name" class="env-card">
      <header class="env-card-head">
        <h2>{{ group.name }}</h2>
        <span class="env-hint">
          {{ group.vars.filter((v) => v.set).length }}/{{ group.vars.length }} set
        </span>
      </header>

      <div
        v-for="v in group.vars"
        :key="v.name"
        class="env-row"
        :class="{ 'is-locked': !v.editable, 'is-dirty': !!changeFor(v) }"
      >
        <div class="env-meta">
          <div class="env-name">
            <code>{{ v.name }}</code>
            <span class="env-pill" :class="pillFor(v).cls">{{ pillFor(v).text }}</span>
          </div>
          <p class="env-hint">
            {{ v.purpose }}
            <a
              v-if="v.docsUrl && v.editable"
              :href="v.docsUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="env-docs"
              >{{ v.secret ? "Get a key" : "Open" }} ↗</a
            >
          </p>
        </div>

        <div v-if="v.editable" class="env-control">
          <select
            v-if="v.kind === 'enum' || v.kind === 'bool'"
            v-model="drafts[v.name]"
            :disabled="applying"
          >
            <option value="">(unset)</option>
            <option
              v-for="option in v.kind === 'bool' ? ['true', 'false'] : v.options"
              :key="option"
              :value="option"
            >
              {{ option }}
            </option>
          </select>
          <template v-else>
            <input
              v-model="drafts[v.name]"
              :type="v.secret ? 'password' : 'text'"
              :placeholder="
                generating.has(v.name)
                  ? 'A new value will be generated'
                  : clearing.has(v.name)
                    ? 'Will be cleared'
                    : v.secret
                      ? v.set
                        ? 'Paste a replacement'
                        : 'Paste a value'
                      : ''
              "
              :disabled="applying"
              autocomplete="off"
              spellcheck="false"
              @input="onInput(v)"
            />
            <Button
              v-if="v.kind === 'generated'"
              class="px-2"
              :class="{ 'is-on': generating.has(v.name) }"
              :disabled="applying"
              @click="toggleGenerate(v)"
            >
              {{ v.set ? "Regenerate" : "Generate" }}
            </Button>
            <Button
              v-if="v.secret && v.set"
              class="px-2"
              :class="{ 'is-on': clearing.has(v.name) }"
              :disabled="applying"
              @click="toggleClear(v)"
            >
              Clear
            </Button>
          </template>
          <p v-if="v.warning && changeFor(v)" class="env-warning">
            {{ v.warning }}
          </p>
        </div>
        <p v-else class="env-control env-locked" :title="v.lockedReason">
          <span v-if="!v.secret && v.set" class="env-value">{{ v.value }}</span>
          <span class="env-hint">🔒 {{ v.lockedReason }}</span>
        </p>
      </div>
    </section>

    <footer v-if="changes.length && !applying" class="env-apply-bar">
      <p>
        <strong>{{ changes.length }}</strong>
        {{ changes.length === 1 ? "change" : "changes" }} — recreates
        {{ affectedServices.join(", ") }}
      </p>
      <div class="env-apply-actions">
        <Button class="px-2" @click="resetDrafts">Discard</Button>
        <Button class="px-2 env-apply" @click="apply">Apply</Button>
      </div>
    </footer>
    <ConfirmDialog
      v-if="applyConfirmationOpen"
      v-model="applyConfirmationOpen"
      :title="applyConfirmation.title"
      :message="applyConfirmation.message"
      :confirm-label="applyConfirmation.confirmLabel"
      @confirm="confirmApply"
    />
  </div>
</template>

<style scoped>
.env {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.env-revealed {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.env-banner {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  padding: 8px 10px;
  border: 1px solid var(--color-primary);
  background: var(--color-surface-tint);
}
.env-banner p {
  margin: 0;
}
.env-banner.is-failed {
  border-color: var(--color-error, #f87171);
}
.env-banner.is-failed > p {
  color: var(--color-error, #f87171);
}
.env-command {
  user-select: all;
  font-size: 0.8rem;
}
.env-output {
  width: 100%;
  max-height: 200px;
  overflow: auto;
  margin: 0;
  padding: 6px;
  font-size: 0.75rem;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--color-primary);
  background: var(--color-surface-deep);
}

.env-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
  gap: 4px 12px;
  align-items: center;
  padding: 6px 10px;
  border-bottom: 1px solid var(--color-quaternary);
  border-left: 2px solid transparent;
}
.env-row:last-child {
  border-bottom: none;
}
.env-row.is-dirty {
  border-left-color: var(--color-tertiary);
  background: var(--color-surface-tint);
}
.env-row.is-locked .env-meta {
  opacity: 0.6;
}

.env-name {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}
.env-name code {
  font-size: 0.8rem;
  word-break: break-all;
}
.env-docs {
  margin-left: 4px;
  padding: 0 4px;
  font-size: 0.75rem;
  white-space: nowrap;
}

.env-control {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin: 0;
}
.env-control input,
.env-control select {
  flex: 1;
  min-width: 120px;
  padding: 3px 6px;
  font-size: 0.85rem;
  color: var(--color-primary);
  background: var(--color-surface);
  border: 1px solid var(--color-quaternary);
}
.env-control input:focus,
.env-control select:focus {
  border-color: var(--color-primary);
  outline: none;
}
.env-control :deep(button.is-on) {
  color: var(--color-tertiary);
  border-color: var(--color-tertiary);
}
.env-control :deep(button:disabled),
.env-control input:disabled,
.env-control select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.env-warning {
  flex-basis: 100%;
  margin: 0;
  font-size: 0.8rem;
  color: var(--color-tertiary);
}
.env-locked {
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
}
.env-value {
  font-size: 0.85rem;
  color: var(--color-primary);
  word-break: break-all;
}

/* Stays in view while scrolling a long list of rows, so the count of staged
   changes and the Apply button are never off-screen. */
.env-apply-bar {
  position: sticky;
  bottom: 8px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid var(--color-tertiary);
  background: var(--color-surface);
  box-shadow: 0 0 12px var(--color-surface-deep);
}
.env-apply-bar p {
  margin: 0;
}
.env-apply-actions {
  display: flex;
  gap: 8px;
}
.env-apply-actions :deep(.env-apply) {
  color: var(--color-surface-deep);
  background: var(--color-primary);
}

@media (max-width: 700px) {
  .env-row {
    grid-template-columns: 1fr;
  }
}
</style>
