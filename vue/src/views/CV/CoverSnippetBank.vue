<script setup>
/**
 * The paragraph bank beside the cover letter (admin only, never printed):
 * reusable paragraphs from stores/coverSnippets.js, grouped by the section
 * they were written for.
 *
 * Dragging works through the parent, which owns the one drag in progress:
 * dragging a snippet out emits `drag-snippet` so CoverLetters.vue can drop a
 * copy into the letter, and a letter paragraph dropped here (`accepting` is
 * true while one is being dragged) emits `drop-paragraph` for the parent to
 * save. Snippets are edited in place and saved on blur.
 */
import { computed, ref } from "vue";
import { useCoverSnippetsStore } from "@/stores/coverSnippets";

const props = defineProps({
    // True while a letter paragraph is being dragged, so the bank can light up.
    accepting: { type: Boolean, default: false },
    // The letter's section titles, offered when adding a snippet.
    sectionTitles: { type: Array, default: () => [] },
});
const emit = defineEmits(["drag-snippet", "drag-end", "drop-paragraph"]);

const store = useCoverSnippetsStore();

const filter = ref("");
const error = ref("");
const over = ref(false);

const newText = ref("");
const newSection = ref("");

/** [{ section, snippets }] in the store's order (section, then newest),
 *  narrowed by the filter, which matches text or section. */
const groups = computed(() => {
    const q = filter.value.trim().toLowerCase();
    const out = [];
    for (const s of store.coverSnippets) {
        if (
            q &&
            !s.text.toLowerCase().includes(q) &&
            !s.section.toLowerCase().includes(q)
        )
            continue;
        let group = out.find((g) => g.section === s.section);
        if (!group) out.push((group = { section: s.section, snippets: [] }));
        group.snippets.push(s);
    }
    return out;
});

async function run(fn) {
    error.value = "";
    try {
        await fn();
    } catch (err) {
        error.value = err.message;
    }
}

function add() {
    run(async () => {
        await store.create({ text: newText.value, section: newSection.value });
        newText.value = "";
    });
}

function save(snippet, field, value) {
    if (snippet[field] === value) return;
    run(() => store.update(snippet.id, { [field]: value }));
}

function remove(snippet) {
    if (!confirm("Delete this paragraph from the bank?")) return;
    run(() => store.remove(snippet.id));
}

function onDragStart(snippet, event) {
    event.dataTransfer.effectAllowed = "copy";
    // Firefox won't start a drag without some data set.
    event.dataTransfer.setData("text/plain", snippet.text);
    // Only the grip is draggable, so text in the textareas stays selectable
    // (Firefox can't place a caret inside a draggable element); show the
    // whole card moving rather than the grip.
    event.dataTransfer.setDragImage(event.target.closest(".snippet"), 12, 12);
    emit("drag-snippet", snippet);
}

function onDragOver(event) {
    if (!props.accepting) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "copy";
    over.value = true;
}

function onDrop(event) {
    over.value = false;
    if (!props.accepting) return;
    event.preventDefault();
    emit("drop-paragraph");
}
</script>

<template>
    <aside
        :class="['bank', 'no-print', { accepting, over }]"
        @dragover="onDragOver"
        @dragleave="over = false"
        @drop="onDrop"
    >
        <div class="bank-head">
            <h3>Paragraph bank</h3>
            <span class="hint">drag into the letter</span>
        </div>
        <input
            v-model="filter"
            class="bank-filter"
            type="search"
            placeholder="Filter…"
        />

        <p v-if="accepting" class="drop-hint">Drop here to save to the bank</p>
        <p v-if="error" class="bank-error">{{ error }}</p>
        <p v-if="store.error" class="bank-error">
            Couldn't load the bank: {{ store.error.message }}
        </p>
        <p
            v-else-if="store.loaded && !store.coverSnippets.length"
            class="empty"
        >
            Nothing saved yet. Drag a paragraph here, or use ☆ on one.
        </p>

        <section v-for="g in groups" :key="g.section" class="group">
            <h4>{{ g.section || "Unsorted" }}</h4>
            <div v-for="s in g.snippets" :key="s.id" class="snippet">
                <span
                    class="grip"
                    draggable="true"
                    title="Drag into the letter"
                    @dragstart="onDragStart(s, $event)"
                    @dragend="emit('drag-end')"
                    >⋮⋮</span
                >
                <textarea
                    :value="s.text"
                    rows="3"
                    @blur="save(s, 'text', $event.target.value.trim())"
                ></textarea>
                <div class="snippet-foot">
                    <input
                        :value="s.section"
                        class="snippet-section"
                        placeholder="Section"
                        @change="save(s, 'section', $event.target.value.trim())"
                    />
                    <button
                        class="cv-btn cv-btn-sm cv-btn-danger"
                        title="Delete from the bank"
                        @click="remove(s)"
                    >
                        ×
                    </button>
                </div>
            </div>
        </section>

        <form class="add" @submit.prevent="add">
            <textarea
                v-model="newText"
                rows="3"
                placeholder="New reusable paragraph…"
            ></textarea>
            <div class="snippet-foot">
                <input
                    v-model="newSection"
                    class="snippet-section"
                    list="cover-section-titles"
                    placeholder="Section"
                />
                <datalist id="cover-section-titles">
                    <option v-for="t in sectionTitles" :key="t" :value="t" />
                </datalist>
                <button
                    class="cv-btn cv-btn-sm"
                    type="submit"
                    :disabled="!newText.trim()"
                >
                    Add
                </button>
            </div>
        </form>
    </aside>
</template>

<style scoped>
.bank {
    align-self: start;
    position: sticky;
    top: 3.5rem;
    max-height: calc(100vh - 4.5rem);
    overflow-y: auto;
    margin: 1rem 0;
    padding: 0.75rem;
    border: 1px solid var(--color-line);
    border-radius: 6px;
    background: var(--color-paper-tint);
    font-size: 0.85rem;
    transition:
        border-color 0.15s,
        background 0.15s;
}

.bank.accepting {
    border-style: dashed;
    border-color: var(--color-ink-muted);
}

.bank.over {
    background: var(--color-paper-shade);
}

.bank-head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 0.5rem;
}

.bank-head h3 {
    font-size: 0.95rem;
    font-weight: 600;
}

.hint,
.empty,
.drop-hint {
    color: var(--color-ink-muted);
    font-size: 0.8rem;
}

.drop-hint {
    margin: 0.5rem 0;
    text-align: center;
}

.bank-error {
    color: var(--color-ink-danger);
    margin: 0.5rem 0;
}

.bank-filter,
.bank textarea,
.snippet-section {
    width: 100%;
    padding: 0.25rem 0.4rem;
    border: 1px solid var(--color-line);
    border-radius: 4px;
    font: inherit;
}

.bank textarea {
    resize: vertical;
    line-height: 1.4;
}

.group {
    margin-top: 0.75rem;
}

.group h4 {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-ink-muted);
    margin-bottom: 0.3rem;
}

.snippet {
    position: relative;
    margin-bottom: 0.5rem;
    padding: 0.4rem 0.4rem 0.4rem 1.3rem;
    border: 1px solid var(--color-line-soft);
    border-radius: 4px;
    background: var(--color-paper);
}

/* A full-height strip down the left edge, so it is easy to grab. */
.grip {
    position: absolute;
    inset: 0 auto 0 0;
    width: 1.1rem;
    padding-top: 0.45rem;
    text-align: center;
    color: var(--color-ink-faint);
    font-size: 0.75rem;
    letter-spacing: -0.15em;
    cursor: grab;
    user-select: none;
}

.grip:hover {
    color: var(--color-ink-soft);
    background: var(--color-paper-tint);
}

.snippet-foot {
    display: flex;
    gap: 0.4rem;
    margin-top: 0.3rem;
}

.snippet-section {
    flex: 1;
    font-size: 0.8rem;
}

.add {
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--color-line);
}
</style>
