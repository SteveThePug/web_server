<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import InlineLink from "@/components/text/InlineLink.vue";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";

// ---------------------------------------------------------------------------
// Presets
// ---------------------------------------------------------------------------
const PRESETS = [
    {
        id: "chaos",
        name: "Chaotic bouncer (runs forever)",
        tape: "1",
        head: 0,
        blank: "0",
        accepting: "",
        delta: [
            "00,22,L",
            "01,02,L",
            "02,10,R",
            "10,11,L",
            "11,02,R",
            "12,22,R",
            "20,10,L",
            "21,12,R",
            "22,01,R",
        ].join("\n"),
    },
    {
        id: "weaver",
        name: "Weaver (grows a lace pattern)",
        tape: "1",
        head: 0,
        blank: "0",
        accepting: "",
        delta: [
            "000,110,R",
            "001,010,R",
            "010,001,L",
            "011,001,R",
            "100,101,L",
            "101,111,R",
            "110,010,R",
            "111,100,R",
        ].join("\n"),
    },
    {
        // The 5-state Busy Beaver champion Turing Machine, converted to a
        // Mobile Automaton with the construction from the dissertation:
        // each cell holds (symbol, state-or-κ); lowercase = state over a 0,
        // uppercase = state over a 1. Halts (after ~47M steps!) on z/Z words.
        id: "busybeaver",
        name: "Busy Beaver — a 5-state Turing Machine, converted",
        tape: "a",
        head: 0,
        blank: "0",
        accepting: "z0,z1,Z0,Z1",
        delta: [
            "0A,1B,R",
            "0B,1C,R",
            "0C,1D,R",
            "0D,A1,L",
            "0E,1Z,R",
            "0a,1b,R",
            "0b,1c,R",
            "0c,1d,R",
            "0d,A0,L",
            "0e,1z,R",
            "1A,C1,L",
            "1B,1B,R",
            "1C,e1,L",
            "1D,D1,L",
            "1E,a1,L",
            "1a,C0,L",
            "1b,1b,R",
            "1c,e0,L",
            "1d,D0,L",
            "1e,a0,L",
            "A0,C0,L",
            "A1,C1,L",
            "B0,1b,R",
            "B1,1B,R",
            "C0,e0,L",
            "C1,e1,L",
            "D0,D0,L",
            "D1,D1,L",
            "E0,a0,L",
            "E1,a1,L",
            "a0,1b,R",
            "a1,1B,R",
            "b0,1c,R",
            "b1,1C,R",
            "c0,1d,R",
            "c1,1D,R",
            "d0,A0,L",
            "d1,A1,L",
            "e0,1z,R",
            "e1,1Z,R",
        ].join("\n"),
    },
    {
        id: "parity",
        name: "Parity checker (accepts / rejects)",
        tape: "E1011010",
        head: 0,
        blank: "_",
        accepting: "E_",
        delta: ["E0,0E,R", "E1,1O,R", "O0,0O,R", "O1,1E,R"].join("\n"),
    },
];

const STORAGE_KEY = "mobile-automata-config";

// ---------------------------------------------------------------------------
// Editable configuration
// ---------------------------------------------------------------------------
const tapeStr = ref(PRESETS[0].tape);
const headStr = ref(String(PRESETS[0].head));
const blankSym = ref(PRESETS[0].blank);
const acceptingStr = ref(PRESETS[0].accepting);
const deltaText = ref(PRESETS[0].delta);
const presetId = ref(PRESETS[0].id);

// Derived (read-only) fields
const derivedN = ref(0);
const derivedSigma = ref("");
const derivedGamma = ref("");
const derivedF = ref("");

const errorMsg = ref("");

// ---------------------------------------------------------------------------
// Machine state
// ---------------------------------------------------------------------------
let machine = null; // { n, delta: Map(word -> [outputChars, dir]), blank, gamma, accepting: Set }
let tape = []; // array of single-char symbols
let origin = 0; // cell index of tape[0]
let head = 0; // absolute cell index

const stepCount = ref(0);
const running = ref(false);
const terminated = ref(false);
const accepted = ref(false);
const speed = ref(55); // slider 0..100, log scale

const stepsPerSecond = computed(() =>
    Math.max(1, Math.round(10 ** ((speed.value * 3.3) / 100))),
);

const statusText = computed(() => {
    if (errorMsg.value) return "Error";
    if (terminated.value) return accepted.value ? "Accepted" : "Rejected";
    return running.value ? "Running" : "Paused";
});

// ---------------------------------------------------------------------------
// Parsing (same formats as the original simulator)
// ---------------------------------------------------------------------------
function parseDelta(str) {
    if (str.trim() === "")
        throw new Error("Transition function cannot be empty");
    const delta = new Map();
    let n = null;
    for (const line of str.split("\n")) {
        if (line.trim() === "") continue;
        const parts = line.split(",").map((s) => s.trim());
        if (parts.length !== 3)
            throw new Error(
                `Invalid transition "${line}" (should be "input,output,direction")`,
            );
        const [input, output, direction] = parts;
        if (direction !== "L" && direction !== "R")
            throw new Error(`Direction in "${line}" must be L or R`);
        if (n === null) n = input.length;
        if (input.length !== n)
            throw new Error(`"${line}": expected input of length ${n}`);
        if (output.length !== n)
            throw new Error(
                `"${line}": input and output must have the same length`,
            );
        if (delta.has(input))
            throw new Error(`Duplicate transition for input "${input}"`);
        delta.set(input, [output.split(""), direction === "R" ? 1 : -1]);
    }
    if (n === null) throw new Error("Transition function cannot be empty");
    return { delta, n };
}

function allWords(gamma, n) {
    let words = [""];
    for (let i = 0; i < n; i++) {
        const next = [];
        for (const w of words) for (const s of gamma) next.push(w + s);
        words = next;
    }
    return words;
}

function loadMachine() {
    errorMsg.value = "";
    terminated.value = false;
    accepted.value = false;
    stepCount.value = 0;
    try {
        const blank = blankSym.value;
        if (!blank || blank.length !== 1)
            throw new Error("Blank symbol must be a single character");

        const headPos = parseInt(headStr.value, 10);
        if (Number.isNaN(headPos))
            throw new Error("Head position must be a number");

        const { delta, n } = parseDelta(deltaText.value);

        // Derive sigma from the tape, gamma from sigma + delta symbols
        const sigma = new Set(tapeStr.value.split(""));
        sigma.add(blank);
        const gamma = new Set(sigma);
        for (const [input, [output]] of delta) {
            for (const c of input) gamma.add(c);
            for (const c of output) gamma.add(c);
        }

        const accepting = new Set(
            acceptingStr.value
                .split(",")
                .map((s) => s.trim())
                .filter((s) => s.length > 0),
        );
        for (const word of accepting) {
            if (word.length !== n)
                throw new Error(
                    `Accepting word "${word}" must have length ${n}`,
                );
            for (const c of word)
                if (!gamma.has(c))
                    throw new Error(
                        `Symbol "${c}" in accepting word "${word}" is not in the alphabet`,
                    );
        }

        machine = { n, delta, blank, gamma: [...gamma], accepting };

        derivedN.value = n;
        derivedSigma.value = [...sigma].join(",");
        derivedGamma.value = machine.gamma.join(",");
        if (machine.gamma.length ** n <= 4096) {
            derivedF.value = allWords(machine.gamma, n)
                .filter((w) => !delta.has(w))
                .join(",");
        } else {
            derivedF.value = "(too many to list)";
        }

        tape = tapeStr.value.split("");
        if (tape.length === 0) tape = [blank];
        origin = 0;
        head = headPos;

        assignColors();
        resetCanvases();

        localStorage.setItem(
            STORAGE_KEY,
            JSON.stringify({
                preset: presetId.value,
                tape: tapeStr.value,
                head: headStr.value,
                blank: blankSym.value,
                accepting: acceptingStr.value,
                delta: deltaText.value,
            }),
        );
    } catch (e) {
        machine = null;
        running.value = false;
        errorMsg.value = e.message;
    }
}

function cellAt(i) {
    const j = i - origin;
    return j >= 0 && j < tape.length ? tape[j] : machine.blank;
}

function setCell(i, v) {
    let j = i - origin;
    while (j < 0) {
        tape.unshift(machine.blank);
        origin--;
        j++;
    }
    while (j >= tape.length) tape.push(machine.blank);
    tape[j] = v;
}

function stepMachine() {
    if (!machine || terminated.value) return false;
    let word = "";
    for (let i = 0; i < machine.n; i++) word += cellAt(head + i);
    const tr = machine.delta.get(word);
    if (!tr) {
        terminated.value = true;
        accepted.value = machine.accepting.has(word);
        running.value = false;
        return false;
    }
    const [output, dir] = tr;
    for (let i = 0; i < machine.n; i++) setCell(head + i, output[i]);
    head += dir;
    stepCount.value++;
    return true;
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------
const historyCanvas = ref(null);
const tapeCanvas = ref(null);
const canvasWrap = ref(null);

const PALETTE = [
    "#55ffbb",
    "#ff579a",
    "#ffd166",
    "#00a2ff",
    "#c084fc",
    "#ff9a00",
    "#62ff57",
    "#f0f0f0",
];
const BG = "#04080f";
const HEAD_COLOR = "#ffffff";

let colorOf = {};

function assignColors() {
    colorOf = {};
    let k = 0;
    for (const s of machine.gamma) {
        if (s === machine.blank) continue;
        colorOf[s] = PALETTE[k % PALETTE.length];
        k++;
    }
}

let dpr = 1;
let cell = 3; // history cell size, device px
let cols = 0;
let rowY = 0;
let viewLeft = 0;
let hctx = null;
let tctx = null;

function resetCanvases() {
    const hc = historyCanvas.value;
    const tc = tapeCanvas.value;
    if (!hc || !tc || !canvasWrap.value) return;
    dpr = window.devicePixelRatio || 1;
    const cssW = canvasWrap.value.clientWidth;

    cell = Math.max(2, Math.round(5 * dpr));
    hc.width = Math.floor(cssW * dpr);
    hc.height = Math.round(420 * dpr);
    hctx = hc.getContext("2d");
    hctx.fillStyle = BG;
    hctx.fillRect(0, 0, hc.width, hc.height);

    tc.width = Math.floor(cssW * dpr);
    tc.height = Math.round(56 * dpr);
    tctx = tc.getContext("2d");

    cols = Math.floor(hc.width / cell);
    viewLeft = head - (cols >> 1);
    rowY = 0;
    drawHistoryRow();
    drawTapeStrip();
}

function drawHistoryRow() {
    if (!hctx || !machine) return;
    const hc = historyCanvas.value;

    // Keep the head inside the viewport, panning existing history sideways
    const margin = 10;
    let dx = 0;
    if (head < viewLeft + margin) dx = head - (viewLeft + margin);
    else if (head > viewLeft + cols - margin)
        dx = head - (viewLeft + cols - margin);
    if (dx !== 0) {
        viewLeft += dx;
        hctx.drawImage(hc, -dx * cell, 0);
        hctx.fillStyle = BG;
        if (dx > 0)
            hctx.fillRect(hc.width - dx * cell, 0, dx * cell, hc.height);
        else hctx.fillRect(0, 0, -dx * cell, hc.height);
    }

    // Scroll up when the bottom is reached
    if (rowY + cell > hc.height) {
        hctx.drawImage(hc, 0, -cell);
        hctx.fillStyle = BG;
        hctx.fillRect(0, hc.height - cell, hc.width, cell);
        rowY = hc.height - cell;
    }

    hctx.fillStyle = BG;
    hctx.fillRect(0, rowY, hc.width, cell);
    for (let i = 0; i < cols; i++) {
        const s = cellAt(viewLeft + i);
        if (s === machine.blank) continue;
        hctx.fillStyle = colorOf[s] || "#888888";
        hctx.fillRect(i * cell, rowY, cell, cell);
    }
    hctx.fillStyle = HEAD_COLOR;
    hctx.fillRect((head - viewLeft) * cell, rowY, cell, cell);
    rowY += cell;
}

function drawTapeStrip() {
    if (!tctx || !machine) return;
    const tc = tapeCanvas.value;
    const w = tc.width;
    const h = tc.height;
    const cw = Math.round(30 * dpr);
    const count = Math.floor(w / cw);
    const first = head - (count >> 1);

    tctx.fillStyle = "#070d18";
    tctx.fillRect(0, 0, w, h);
    tctx.font = `${Math.round(16 * dpr)}px monospace`;
    tctx.textAlign = "center";
    tctx.textBaseline = "middle";

    for (let i = 0; i < count; i++) {
        const idx = first + i;
        const x = i * cw;
        const inWindow = idx >= head && idx < head + machine.n;
        if (inWindow) {
            tctx.fillStyle = "rgba(85, 255, 187, 0.15)";
            tctx.fillRect(x, 0, cw, h);
        }
        tctx.strokeStyle = inWindow ? "#55ffbb" : "#1f2b3d";
        tctx.lineWidth = inWindow ? 2 * dpr : dpr;
        tctx.strokeRect(x + 1, 1, cw - 2, h - 2);
        const s = cellAt(idx);
        tctx.fillStyle =
            s === machine.blank ? "#3b4a61" : colorOf[s] || "#cccccc";
        tctx.fillText(s, x + cw / 2, h / 2);
    }
}

// ---------------------------------------------------------------------------
// Animation loop
// ---------------------------------------------------------------------------
let rafId = 0;
let lastTime = 0;
let stepDebt = 0;

function frame(now) {
    rafId = requestAnimationFrame(frame);
    if (!running.value || !machine) {
        lastTime = now;
        return;
    }
    const dt = Math.min(0.25, (now - lastTime) / 1000);
    lastTime = now;
    stepDebt += dt * stepsPerSecond.value;
    let budget = Math.min(64, Math.floor(stepDebt));
    stepDebt -= Math.floor(stepDebt);
    let moved = false;
    while (budget-- > 0) {
        if (!stepMachine()) break;
        drawHistoryRow();
        moved = true;
    }
    if (moved) drawTapeStrip();
}

// ---------------------------------------------------------------------------
// UI actions
// ---------------------------------------------------------------------------
function onToggleRun() {
    if (!machine || terminated.value) return;
    running.value = !running.value;
}

function onStep() {
    if (!machine) return;
    running.value = false;
    if (stepMachine()) {
        drawHistoryRow();
        drawTapeStrip();
    }
}

function onReset() {
    loadMachine();
    if (machine) running.value = true;
}

function applyPreset(id) {
    const p = PRESETS.find((x) => x.id === id);
    if (!p) return;
    presetId.value = p.id;
    tapeStr.value = p.tape;
    headStr.value = String(p.head);
    blankSym.value = p.blank;
    acceptingStr.value = p.accepting;
    deltaText.value = p.delta;
    onReset();
}

function onConfigChange() {
    presetId.value = "";
    loadMachine();
    if (machine) running.value = true;
}

let resizeObserver = null;

onMounted(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
        try {
            const c = JSON.parse(saved);
            tapeStr.value = c.tape ?? tapeStr.value;
            headStr.value = c.head ?? headStr.value;
            blankSym.value = c.blank ?? blankSym.value;
            acceptingStr.value = c.accepting ?? acceptingStr.value;
            deltaText.value = c.delta ?? deltaText.value;
            presetId.value = c.preset ?? "";
        } catch {
            /* corrupted saved config: fall back to the default preset */
        }
    }
    loadMachine();
    if (machine) running.value = true;
    rafId = requestAnimationFrame((t) => {
        lastTime = t;
        frame(t);
    });
    resizeObserver = new ResizeObserver(() => {
        if (machine) resetCanvases();
    });
    if (canvasWrap.value) resizeObserver.observe(canvasWrap.value);
});

onBeforeUnmount(() => {
    cancelAnimationFrame(rafId);
    if (resizeObserver) resizeObserver.disconnect();
});
</script>

<template>
    <main class="flex justify-center px-4 py-10">
        <div class="max-w-4xl w-full flex flex-col gap-6">
            <section>
                <Header>Mobile Automata Simulator</Header>
                <Paragraph>
                    A Mobile Automaton repeatedly rewrites a small window of an
                    infinite tape and moves it left or right. My dissertation,
                    <InlineLink href="/pdf/dissertation.pdf">
                        Equivalences and Properties of Mobile
                        Automata</InlineLink
                    >, proves these machines are computationally equivalent to
                    Turing Machines. Below, one is running live.
                </Paragraph>
            </section>

            <section class="panel" ref="canvasWrap">
                <canvas ref="tapeCanvas" class="w-full block"></canvas>
                <canvas ref="historyCanvas" class="w-full block mt-1"></canvas>

                <div class="statusRow">
                    <span class="statusChip" :class="statusText.toLowerCase()">
                        {{ statusText }}
                    </span>
                    <span class="stepCounter">step {{ stepCount }}</span>
                    <span v-if="terminated" class="terminalNote">
                        halted on a word with no transition —
                        {{
                            accepted
                                ? "it is in the accepting set"
                                : "not in the accepting set"
                        }}
                    </span>
                </div>

                <div class="controls">
                    <button
                        class="btn"
                        @click="onToggleRun"
                        :disabled="terminated || !!errorMsg"
                    >
                        {{ running ? "Pause" : "Run" }}
                    </button>
                    <button
                        class="btn"
                        @click="onStep"
                        :disabled="terminated || !!errorMsg"
                    >
                        Step
                    </button>
                    <button class="btn" @click="onReset">Reset</button>

                    <label class="speedLabel">
                        speed
                        <input
                            type="range"
                            min="0"
                            max="100"
                            v-model.number="speed"
                        />
                        <span class="speedValue">{{ stepsPerSecond }}/s</span>
                    </label>

                    <select
                        class="presetSelect"
                        :value="presetId"
                        @change="applyPreset($event.target.value)"
                    >
                        <option value="" disabled>Presets…</option>
                        <option v-for="p in PRESETS" :key="p.id" :value="p.id">
                            {{ p.name }}
                        </option>
                    </select>
                </div>

                <p v-if="errorMsg" class="errorBox">{{ errorMsg }}</p>
            </section>

            <section class="panel">
                <h2 class="panelTitle">Machine definition</h2>
                <div class="configGrid">
                    <label>
                        Tape contents
                        <input
                            type="text"
                            v-model="tapeStr"
                            @change="onConfigChange"
                        />
                    </label>
                    <label>
                        Head position
                        <input
                            type="number"
                            v-model="headStr"
                            @change="onConfigChange"
                        />
                    </label>
                    <label>
                        Blank symbol
                        <input
                            type="text"
                            v-model="blankSym"
                            maxlength="1"
                            @change="onConfigChange"
                        />
                    </label>
                    <label>
                        Accepting words
                        <input
                            type="text"
                            v-model="acceptingStr"
                            placeholder="comma-separated"
                            @change="onConfigChange"
                        />
                    </label>
                    <label class="deltaField">
                        Transition function δ — one
                        <code>input,output,direction</code> per line
                        <textarea
                            v-model="deltaText"
                            rows="9"
                            spellcheck="false"
                            @change="onConfigChange"
                        ></textarea>
                    </label>
                </div>

                <div class="derivedGrid">
                    <label>
                        Window size (n)
                        <input type="text" :value="derivedN" readonly />
                    </label>
                    <label>
                        Input alphabet (Σ)
                        <input type="text" :value="derivedSigma" readonly />
                    </label>
                    <label>
                        Working alphabet (Γ)
                        <input type="text" :value="derivedGamma" readonly />
                    </label>
                    <label>
                        Halting words
                        <input type="text" :value="derivedF" readonly />
                    </label>
                </div>

                <details class="help">
                    <summary>How it works</summary>
                    <ul>
                        <li>
                            Each step, the machine reads the n symbols under the
                            highlighted window, replaces them using δ, and moves
                            one cell left (L) or right (R).
                        </li>
                        <li>
                            It halts when the current word has no transition,
                            and <em>accepts</em> if that word is in the
                            accepting set.
                        </li>
                        <li>
                            n, Σ and Γ are derived automatically from the tape
                            and δ. Your machine is saved in this browser.
                        </li>
                        <li>
                            The lower canvas is a space–time diagram: each row
                            is the tape at one step (white marks the window
                            position), most recent at the bottom.
                        </li>
                    </ul>
                </details>
            </section>
        </div>
    </main>
</template>

<style scoped>
.panel {
    border: 1px solid var(--quaternary);
    background: rgba(4, 8, 15, 0.55);
    padding: 0.9rem;
}

.panelTitle {
    color: var(--primary);
    font-family: var(--font_heading);
    font-size: 1.4rem;
    margin-bottom: 0.5rem;
}

.statusRow {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.6rem;
    flex-wrap: wrap;
}

.statusChip {
    font-family: var(--font_heading);
    letter-spacing: 0.05em;
    padding: 0.1rem 0.6rem;
    border: 1px solid currentColor;
    color: var(--primary);
}

.statusChip.paused {
    color: #ffd166;
}

.statusChip.accepted {
    color: var(--secondary);
}

.statusChip.rejected,
.statusChip.error {
    color: var(--tertiary);
}

.stepCounter {
    font-family: monospace;
    color: var(--portal_grey);
}

.terminalNote {
    color: var(--portal_grey);
    font-size: 0.85rem;
}

.controls {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-top: 0.75rem;
    flex-wrap: wrap;
}

.btn {
    padding: 0.35rem 1.1rem;
    border: 1px solid var(--primary);
    color: var(--primary);
    font-family: var(--font_heading);
    letter-spacing: 0.05em;
    cursor: pointer;
    transition:
        background 0.15s ease,
        color 0.15s ease;
}

.btn:hover:enabled {
    background: var(--primary);
    color: var(--bg_secondary);
}

.btn:disabled {
    opacity: 0.35;
    cursor: default;
}

.speedLabel {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--portal_grey);
    font-size: 0.9rem;
}

.speedLabel input[type="range"] {
    accent-color: var(--primary);
    width: 8rem;
}

.speedValue {
    font-family: monospace;
    min-width: 4.5rem;
}

.presetSelect {
    margin-left: auto;
    background: var(--bg_secondary);
    color: var(--primary);
    border: 1px solid var(--quaternary);
    padding: 0.3rem 0.5rem;
}

.errorBox {
    margin-top: 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--tertiary);
    color: var(--tertiary);
    font-family: monospace;
    font-size: 0.9rem;
}

.configGrid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.6rem 1rem;
}

.derivedGrid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.6rem 1rem;
    margin-top: 0.9rem;
    opacity: 0.75;
}

.configGrid label,
.derivedGrid label {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    color: var(--portal_grey);
    font-size: 0.85rem;
}

.deltaField {
    grid-column: 1 / -1;
}

.configGrid input,
.configGrid textarea,
.derivedGrid input {
    background: var(--bg_secondary);
    border: 1px solid var(--quaternary);
    color: #e5f4ee;
    padding: 0.35rem 0.5rem;
    font-family: monospace;
}

.configGrid textarea {
    resize: vertical;
}

.derivedGrid input {
    color: var(--portal_grey);
}

.configGrid input:focus,
.configGrid textarea:focus {
    outline: 1px solid var(--primary);
}

.help {
    margin-top: 0.9rem;
    color: var(--portal_grey);
    font-size: 0.9rem;
}

.help summary {
    cursor: pointer;
    color: var(--primary);
    font-family: var(--font_heading);
    letter-spacing: 0.05em;
}

.help ul {
    list-style: disc;
    padding-left: 1.4rem;
    margin-top: 0.4rem;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

@media (max-width: 640px) {
    .configGrid,
    .derivedGrid {
        grid-template-columns: 1fr;
    }

    .presetSelect {
        margin-left: 0;
    }
}
</style>
