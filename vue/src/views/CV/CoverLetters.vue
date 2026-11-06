<script setup>
/**
 * Editable cover letter (contenteditable fields, print to PDF), split into
 * named sections whose paragraphs can be dragged within and between them.
 * STATUS: live — CV.vue imports it as the "Cover Letter" tab on /cv.
 *
 * Everyone gets the editor; /cv is public and nothing here needs the server
 * until something is saved. An admin also gets the toolbar (saved letters,
 * stores/coverLetters.js) and the paragraph bank (CoverSnippetBank.vue,
 * stores/coverSnippets.js). The letter's data shape and the moves live in
 * ./coverLetter.js.
 *
 * Section titles are editor furniture: they are `no-print`, so the printed
 * letter is one run of paragraphs, as it was before sections existed.
 *
 * Editing: contenteditable elements are filled with v-text and read back on
 * blur. v-text, not {{ }}, because it replaces the element's whole content —
 * an interpolated text node would sit alongside the <div>s a browser adds on
 * Enter and duplicate the text when Vue patched it.
 *
 * Dragging is native HTML5 DnD. `drag` holds what is in flight (dataTransfer
 * can't be read until the drop) and `dropAt` where it would land, which is
 * what draws the indicator line. Only the ⋮⋮ handle is draggable, so text in
 * a paragraph stays selectable.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useCoverLettersStore } from "@/stores/coverLetters";
import { useCoverSnippetsStore } from "@/stores/coverSnippets";
import CoverSnippetBank from "./CoverSnippetBank.vue";
import {
    insertParagraph,
    moveParagraph,
    newLetter,
    newParagraph,
    newSection,
    parseBody,
    serialize,
} from "./coverLetter";

const auth = useAuthStore();
const letters = useCoverLettersStore();
const snippets = useCoverSnippetsStore();
const isAdmin = computed(() => auth.loggedIn && !!auth.user.admin);

const letter = ref(newLetter());
// The open saved letter's id, or null for one never saved (or deleted).
const currentId = ref(null);
// The body as last saved or loaded; the letter is unsaved when it differs.
// Starts as the blank template so an untouched page doesn't nag.
const savedBody = ref(serialize(letter.value));
const dirty = computed(() => serialize(letter.value) !== savedBody.value);
const current = computed(() =>
    letters.coverLetters.find((l) => l.id === currentId.value),
);

const busy = ref(false);
const error = ref("");

onMounted(() => auth.ensureLoaded());

// The stores are admin only, so they are fetched once an admin is known,
// which can be after mount on a cold load.
watch(
    isAdmin,
    (admin) => {
        if (!admin) return;
        if (!letters.loaded) letters.fetch();
        if (!snippets.loaded) snippets.fetch();
    },
    { immediate: true },
);

// Closing the tab with unsaved changes asks first. Only for an admin, who
// can save; a visitor's letter is throwaway by design.
function onBeforeUnload(event) {
    if (isAdmin.value && dirty.value) event.preventDefault();
}
onMounted(() => window.addEventListener("beforeunload", onBeforeUnload));
onBeforeUnmount(() =>
    window.removeEventListener("beforeunload", onBeforeUnload),
);

// ── Saving ──────────────────────────────────────────────────────────────

async function run(fn) {
    error.value = "";
    busy.value = true;
    try {
        await fn();
    } catch (err) {
        error.value = err.message;
    } finally {
        busy.value = false;
    }
}

function discardOk() {
    return !dirty.value || confirm("Discard the unsaved changes?");
}

function open(letterRow) {
    letter.value = parseBody(letterRow.body);
    currentId.value = letterRow.id;
    // The normalised form, not the raw body, so an older row doesn't open
    // already marked unsaved.
    savedBody.value = serialize(letter.value);
}

function onPick(event) {
    const id = event.target.value;
    const row = letters.coverLetters.find((l) => l.id === id);
    if (row && discardOk()) open(row);
    // Put the select back if the switch was refused or "New" was picked.
    event.target.value = currentId.value ?? "";
}

function startNew() {
    if (!discardOk()) return;
    letter.value = newLetter();
    currentId.value = null;
    savedBody.value = serialize(letter.value);
}

function save() {
    if (!currentId.value) return saveAs();
    const body = serialize(letter.value);
    run(async () => {
        await letters.update(currentId.value, { body });
        savedBody.value = body;
    });
}

function saveAs() {
    const suggestion = current.value
        ? `${current.value.name} (copy)`
        : letter.value.role.trim();
    const name = prompt("Name this cover letter", suggestion);
    if (name === null) return;
    const body = serialize(letter.value);
    run(async () => {
        const created = await letters.create({ name, body });
        currentId.value = created.id;
        savedBody.value = body;
    });
}

function rename() {
    if (!current.value) return;
    const name = prompt("Rename cover letter", current.value.name);
    if (name === null || name === current.value.name) return;
    run(() => letters.update(currentId.value, { name }));
}

function removeLetter() {
    if (!current.value) return;
    if (!confirm(`Delete "${current.value.name}"? This can't be undone.`))
        return;
    run(async () => {
        await letters.remove(currentId.value);
        // The text stays on the page, now unsaved, in case it was a misclick.
        currentId.value = null;
        savedBody.value = "";
    });
}

// ── Editing ─────────────────────────────────────────────────────────────

function addParagraph(section) {
    section.paragraphs.push(newParagraph());
}

function removeParagraph(section, index) {
    section.paragraphs.splice(index, 1);
}

function addSection() {
    letter.value.sections.push(newSection("New section", ["New paragraph..."]));
}

function removeSection(index) {
    const section = letter.value.sections[index];
    if (
        section.paragraphs.some((p) => p.text.trim()) &&
        !confirm(`Remove "${section.title || "this section"}" and its text?`)
    )
        return;
    letter.value.sections.splice(index, 1);
}

const sectionTitles = computed(() =>
    [...new Set(letter.value.sections.map((s) => s.title.trim()))].filter(
        Boolean,
    ),
);

function saveToBank(section, para) {
    run(() =>
        snippets.create({ text: para.text, section: section.title.trim() }),
    );
}

// ── Dragging ────────────────────────────────────────────────────────────

// { kind: "para", from: { section, index } } | { kind: "snippet", text } | null
const drag = ref(null);
// { section, index } — where a drop would land, or null.
const dropAt = ref(null);

function onParaDragStart(si, pi, event) {
    drag.value = { kind: "para", from: { section: si, index: pi } };
    event.dataTransfer.effectAllowed = "copyMove";
    // Firefox won't start a drag without some data set.
    event.dataTransfer.setData(
        "text/plain",
        letter.value.sections[si].paragraphs[pi].text,
    );
    event.dataTransfer.setDragImage(
        event.target.closest(".paragraph-wrapper"),
        16,
        16,
    );
}

function onSnippetDrag(snippet) {
    drag.value = { kind: "snippet", text: snippet.text };
}

function endDrag() {
    drag.value = null;
    dropAt.value = null;
}

function accept(event) {
    if (!drag.value) return false;
    event.preventDefault();
    event.dataTransfer.dropEffect = drag.value.kind === "para" ? "move" : "copy";
    return true;
}

/** Over a paragraph: before it if the pointer is in its top half, else after. */
function onParaDragOver(si, pi, event) {
    if (!accept(event)) return;
    const rect = event.currentTarget.getBoundingClientRect();
    const after = event.clientY > rect.top + rect.height / 2;
    dropAt.value = { section: si, index: after ? pi + 1 : pi };
}

/** Over a section's end zone, or its title row: the end of that section. */
function onSectionEndDragOver(si, event) {
    if (!accept(event)) return;
    dropAt.value = {
        section: si,
        index: letter.value.sections[si].paragraphs.length,
    };
}

function onDrop(event) {
    if (!drag.value || !dropAt.value) return endDrag();
    event.preventDefault();
    if (drag.value.kind === "para")
        moveParagraph(letter.value, drag.value.from, dropAt.value);
    else insertParagraph(letter.value, dropAt.value, drag.value.text);
    endDrag();
}

/** A letter paragraph dropped on the bank: save a copy there. */
function onBankDrop() {
    const from = drag.value?.from;
    endDrag();
    if (!from) return;
    const section = letter.value.sections[from.section];
    saveToBank(section, section.paragraphs[from.index]);
}

const indicatorAt = (si, index) =>
    dropAt.value?.section === si && dropAt.value?.index === index;
</script>

<template>
    <main :class="['cl-wrap', { 'with-bank': isAdmin }]">
        <div class="cl-letter">
            <div v-if="isAdmin" class="no-print toolbar">
                <select
                    class="letter-pick"
                    :value="currentId ?? ''"
                    :disabled="busy"
                    @change="onPick"
                >
                    <option value="" disabled>
                        {{
                            letters.coverLetters.length
                                ? "Open a saved letter…"
                                : "No saved letters"
                        }}
                    </option>
                    <option
                        v-for="l in letters.coverLetters"
                        :key="l.id"
                        :value="l.id"
                    >
                        {{ l.name }}
                    </option>
                </select>
                <button
                    class="cv-btn cv-btn-sm"
                    :disabled="busy"
                    @click="startNew"
                >
                    New
                </button>
                <button
                    class="cv-btn cv-btn-sm cv-btn-primary"
                    :disabled="busy || (currentId && !dirty)"
                    @click="save"
                >
                    Save
                </button>
                <button
                    class="cv-btn cv-btn-sm"
                    :disabled="busy"
                    @click="saveAs"
                >
                    Save as…
                </button>
                <button
                    class="cv-btn cv-btn-sm"
                    :disabled="busy || !current"
                    @click="rename"
                >
                    Rename
                </button>
                <button
                    class="cv-btn cv-btn-sm cv-btn-danger"
                    :disabled="busy || !current"
                    @click="removeLetter"
                >
                    Delete
                </button>
                <span v-if="dirty" class="unsaved">● unsaved</span>
                <span v-else-if="current" class="saved">saved</span>
                <p v-if="error" class="toolbar-error">{{ error }}</p>
                <p v-else-if="letters.error" class="toolbar-error">
                    Couldn't load saved letters: {{ letters.error.message }}
                </p>
            </div>

            <div class="a5page" @dragend="endDrag">
                <div class="contact">
                    <h1
                        contenteditable="true"
                        v-text="letter.name"
                        @blur="letter.name = $event.target.innerText"
                    ></h1>
                    <div class="contact-details">
                        <p
                            contenteditable="true"
                            v-text="letter.phone"
                            @blur="letter.phone = $event.target.innerText"
                        ></p>
                        <p
                            contenteditable="true"
                            v-text="letter.email"
                            @blur="letter.email = $event.target.innerText"
                        ></p>
                    </div>
                </div>
                <h2
                    contenteditable="true"
                    v-text="letter.role"
                    @blur="letter.role = $event.target.innerText"
                ></h2>

                <section
                    v-for="(section, si) in letter.sections"
                    :key="section.id"
                    class="letter-section"
                >
                    <div
                        class="no-print section-head"
                        @dragover="onSectionEndDragOver(si, $event)"
                        @drop="onDrop"
                    >
                        <input
                            v-model="section.title"
                            class="section-title"
                            placeholder="Section title"
                            aria-label="Section title"
                        />
                        <button
                            class="section-btn"
                            title="Add a paragraph to this section"
                            @click="addParagraph(section)"
                        >
                            + paragraph
                        </button>
                        <button
                            class="section-btn danger"
                            title="Remove this section"
                            @click="removeSection(si)"
                        >
                            remove
                        </button>
                    </div>

                    <div
                        v-for="(p, pi) in section.paragraphs"
                        :key="p.id"
                        :class="[
                            'paragraph-wrapper',
                            {
                                'drop-before': indicatorAt(si, pi),
                                dragging:
                                    drag?.kind === 'para' &&
                                    drag.from.section === si &&
                                    drag.from.index === pi,
                            },
                        ]"
                        @dragover="onParaDragOver(si, pi, $event)"
                        @drop="onDrop"
                    >
                        <span
                            class="no-print drag-handle"
                            draggable="true"
                            title="Drag to move"
                            @dragstart="onParaDragStart(si, pi, $event)"
                            >⋮⋮</span
                        >
                        <p
                            contenteditable="true"
                            v-text="p.text"
                            @blur="p.text = $event.target.innerText"
                        ></p>
                        <div class="no-print para-actions">
                            <button
                                v-if="isAdmin"
                                class="bank-btn"
                                title="Save to the paragraph bank"
                                @click="saveToBank(section, p)"
                            >
                                ☆
                            </button>
                            <button
                                class="remove-btn"
                                title="Remove paragraph"
                                @click="removeParagraph(section, pi)"
                            >
                                ×
                            </button>
                        </div>
                    </div>

                    <div
                        :class="[
                            'no-print',
                            'section-end',
                            {
                                'drop-before': indicatorAt(
                                    si,
                                    section.paragraphs.length,
                                ),
                                empty: !section.paragraphs.length,
                            },
                        ]"
                        @dragover="onSectionEndDragOver(si, $event)"
                        @drop="onDrop"
                    >
                        <span v-if="!section.paragraphs.length">
                            Empty — drop a paragraph here
                        </span>
                    </div>
                </section>

                <p
                    contenteditable="true"
                    v-text="letter.signoff"
                    @blur="letter.signoff = $event.target.innerText"
                ></p>
                <button class="no-print add-btn" @click="addSection()">
                    + Add section
                </button>
            </div>
        </div>

        <CoverSnippetBank
            v-if="isAdmin"
            :accepting="drag?.kind === 'para'"
            :section-titles="sectionTitles"
            @drag-snippet="onSnippetDrag"
            @drag-end="endDrag"
            @drop-paragraph="onBankDrop"
        />
    </main>
</template>

<style scoped>
.cl-wrap.with-bank {
    display: grid;
    grid-template-columns: minmax(0, 210mm) 18rem;
    justify-content: center;
    gap: 1.5rem;
    padding: 0 1rem;
}

/* Narrow screens: the bank goes under the letter. */
@media (max-width: 1100px) {
    .cl-wrap.with-bank {
        grid-template-columns: minmax(0, 210mm);
    }
}

.toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
    max-width: 210mm;
    margin: 1rem auto 0;
}

.letter-pick {
    min-width: 12rem;
    padding: 0.2rem 0.4rem;
    border: 1px solid var(--color-ink-soft);
    border-radius: 4px;
    font-size: 0.85rem;
}

.unsaved {
    color: var(--color-ink-danger);
    font-size: 0.8rem;
}

.saved {
    color: var(--color-ink-muted);
    font-size: 0.8rem;
}

.toolbar-error {
    flex-basis: 100%;
    color: var(--color-ink-danger);
    font-size: 0.85rem;
}

.a5page {
    max-width: 210mm;
    min-height: 148mm;
    margin: 1rem auto;
    padding: 2rem 3rem;
    background: white;
    color: #222;
    line-height: 1.5;
}

.a5page h1 {
    font-size: 1.6rem;
    font-weight: 700;
}

.a5page h2 {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0.75rem 0;
}

.a5page p {
    margin-bottom: 0.75rem;
}

.contact {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    border-bottom: 1px solid #333;
    padding-bottom: 0.5rem;
    margin-bottom: 0.75rem;
}

.contact-details {
    text-align: right;
    font-size: 0.85rem;
}

.contact-details p {
    margin: 0;
}

/* innerText turns a line break typed into a paragraph into "\n"; this keeps
   it showing as one once v-text writes it back. */
[contenteditable="true"] {
    white-space: pre-wrap;
}

[contenteditable="true"]:hover,
[contenteditable="true"]:focus {
    outline: 1px dashed #999;
    outline-offset: 2px;
}

.section-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.25rem 0 0.35rem;
    font-size: 0.75rem;
}

.section-title {
    flex: 1;
    padding: 0;
    border: none;
    border-bottom: 1px dashed transparent;
    background: transparent;
    color: #888;
    font: inherit;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.section-title:hover,
.section-title:focus {
    border-bottom-color: #bbb;
    outline: none;
    color: #555;
}

.section-btn {
    border: none;
    background: none;
    color: #888;
    cursor: pointer;
    font-size: 0.75rem;
    opacity: 0;
    transition: opacity 0.15s;
}

.section-btn:hover {
    color: #333;
}

.section-btn.danger:hover {
    color: #c00;
}

.letter-section:hover .section-btn,
.section-btn:focus-visible {
    opacity: 1;
}

.paragraph-wrapper {
    position: relative;
}

.paragraph-wrapper.dragging {
    opacity: 0.4;
}

/* The drop indicator: a line just above the paragraph (or end zone) the
   dragged one would land before. */
.drop-before::before {
    content: "";
    position: absolute;
    left: 0;
    right: 0;
    top: -0.45rem;
    height: 2px;
    border-radius: 1px;
    background: #2563eb;
}

.drag-handle {
    position: absolute;
    top: 0.1rem;
    left: -1.6rem;
    padding: 0 0.25rem;
    color: #aaa;
    cursor: grab;
    font-size: 0.85rem;
    letter-spacing: -0.15em;
    user-select: none;
    opacity: 0;
    transition: opacity 0.15s;
}

.para-actions {
    position: absolute;
    top: 0;
    right: -2.6rem;
    display: flex;
    gap: 0.1rem;
    opacity: 0;
    transition: opacity 0.15s;
}

.paragraph-wrapper:hover .drag-handle,
.paragraph-wrapper:hover .para-actions,
.para-actions:focus-within {
    opacity: 1;
}

.para-actions button {
    border: none;
    background: none;
    cursor: pointer;
    font-size: 1rem;
    line-height: 1.5;
}

.remove-btn {
    color: #c00;
}

.bank-btn {
    color: #b8860b;
}

.section-end {
    position: relative;
    min-height: 0.5rem;
}

.section-end.empty {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 2.5rem;
    margin-bottom: 0.75rem;
    border: 1px dashed #ccc;
    border-radius: 4px;
    color: #999;
    font-size: 0.8rem;
}

.add-btn {
    margin-top: 0.5rem;
    padding: 0.3rem 0.8rem;
    border: 1px solid #333;
    border-radius: 4px;
    background: white;
    color: #333;
    cursor: pointer;
    font-size: 0.85rem;
}

.add-btn:hover {
    background: #eee;
}

@media print {
    @page {
        size: A5 landscape;
        margin: 0;
    }
    .no-print {
        display: none !important;
    }
    .cl-wrap.with-bank {
        display: block;
        padding: 0;
    }
}
</style>
