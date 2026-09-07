<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import InlineLink from "@/components/text/InlineLink.vue";

const STORAGE_KEY = "cheap-hotels-config";
const CUSTOM_ORIGIN = "__custom__";

const RANKINGS = [
    { key: "total", label: "Stay + transport" },
    { key: "room", label: "Stay only" },
    { key: "cycle", label: "Cycle time" },
    { key: "minutes", label: "Travel time" },
];

// ---------------------------------------------------------------------------
// Form state
// ---------------------------------------------------------------------------
function nextFriday() {
    const d = new Date();
    const delta = (5 - d.getDay() + 7) % 7 || 7;
    d.setDate(d.getDate() + delta);
    return d.toISOString().slice(0, 10);
}

const location = ref("Central London");
const checkin = ref(nextFriday());
const nights = ref(1);
const guests = ref("1");
const railcardHolders = ref(1);
const originKey = ref("clapham junction");
const customOrigin = ref("");
const depart = ref("19:30");
const top = ref(20);
const maxMiles = ref("");
const maxTravelMin = ref(75);
const rankBy = ref("total");

const origins = ref([]);

// ---------------------------------------------------------------------------
// Search state
// ---------------------------------------------------------------------------
const searching = ref(false);
const errorMsg = ref("");
const summary = ref(null);
const rows = ref([]);
const cached = ref(false);
const done = ref(false);
const elapsed = ref(0);
let ticker = null;
let aborter = null;

const statusText = computed(() => {
    if (errorMsg.value) return "Error";
    if (searching.value) return "Searching";
    if (done.value) return "Done";
    return "Idle";
});

const pricedCount = computed(
    () => rows.value.filter((r) => r.total !== null || r.route !== "pricing…").length,
);

const hasResults = computed(() => summary.value !== null);

function byNumber(get, tiebreak) {
    return (a, b) => {
        const x = get(a);
        const y = get(b);
        const xNull = x === null || x === undefined;
        const yNull = y === null || y === undefined;
        if (xNull !== yNull) return xNull ? 1 : -1;
        if (!xNull && x !== y) return x - y;
        return tiebreak ? tiebreak(a, b) : 0;
    };
}

const byRoom = byNumber((r) => r.room_price);
const byTotal = byNumber((r) => r.total, byRoom);
const byCycle = byNumber((r) => r.cycle_minutes, byRoom);
const byMinutes = byNumber((r) => r.travel_minutes, byTotal);

const sortedRows = computed(() => {
    const cmp = { total: byTotal, room: byRoom, cycle: byCycle, minutes: byMinutes }[
        rankBy.value
    ];
    return [...rows.value].sort(cmp);
});

function firstBy(cmp, ok) {
    const list = rows.value.filter(ok);
    if (!list.length) return null;
    return list.reduce((best, r) => (cmp(r, best) < 0 ? r : best));
}

const picks = computed(() => {
    const p = [];
    const best = firstBy(byTotal, (r) => r.total !== null);
    if (best)
        p.push({
            label: "Best value",
            row: best,
            maths: `${gbp(best.room_price)} stay + ${gbp(best.transport_total)} travel = ${gbp(best.total)}`,
            sub: `${best.travel_minutes} min by ${best.route}`,
        });
    const room = firstBy(byRoom, (r) => r.room_price !== null);
    if (room)
        p.push({
            label: "Cheapest room",
            row: room,
            maths: `${gbp(room.room_price)} for the stay`,
            sub:
                room.total !== null
                    ? `${gbp(room.total)} with transport (${room.travel_minutes} min)`
                    : "transport not priced yet",
        });
    const cycle = firstBy(byCycle, (r) => r.cycle_minutes !== null);
    if (cycle)
        p.push({
            label: "Quickest cycle",
            row: cycle,
            maths: `${cycleText(cycle)} ride, ${gbp(cycle.room_price)} stay`,
            sub: `${cycle.cycle_km} km from ${summary.value?.origin_name ?? "origin"}`,
        });
    return p;
});

const summaryText = computed(() => {
    const s = summary.value;
    if (!s) return "";
    const party =
        `${s.adults} adult${s.adults === 1 ? "" : "s"}` +
        (s.children ? ` + ${s.children} child${s.children === 1 ? "" : "ren"}` : "") +
        ` in ${s.rooms} room${s.rooms === 1 ? "" : "s"}`;
    const railcard = s.railcard_holders
        ? ` (${s.railcard_holders} with railcard)`
        : "";
    return (
        `Cheapest ${rows.value.length} of ${s.hotels_available} available Travelodges near ` +
        `'${s.location}', ${fmtDate(s.checkin)} → ${fmtDate(s.checkout)} ` +
        `(${s.nights} night${s.nights === 1 ? "" : "s"}, ${party}), ` +
        `with round-trip transport from ${s.origin_name} departing ${fmtTime(s.depart)}${railcard}.`
    );
});

// ---------------------------------------------------------------------------
// Formatting
// ---------------------------------------------------------------------------
function gbp(v) {
    return v === null || v === undefined ? "?" : `£${v.toFixed(2)}`;
}

function fmtDate(iso) {
    const d = new Date(iso + "T00:00:00");
    return d.toLocaleDateString("en-GB", {
        weekday: "short",
        day: "numeric",
        month: "short",
    });
}

function fmtTime(iso) {
    return new Date(iso).toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
    });
}

function cycleText(r) {
    if (r.cycle_minutes === null || r.cycle_minutes === undefined) return "";
    return `${r.cycle_source === "estimate" ? "~" : ""}${r.cycle_minutes} min`;
}

function cycleTitle(r) {
    if (r.cycle_minutes === null || r.cycle_minutes === undefined) return "";
    return r.cycle_source === "estimate"
        ? `about ${r.cycle_km} km; estimated from distance at a gentle pace`
        : `${r.cycle_km} km on TfL's cycle route`;
}

function showFaster(r) {
    return (
        r.fastest_minutes !== null &&
        r.travel_minutes !== null &&
        r.fastest_minutes < r.travel_minutes - 10
    );
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------
function saveConfig() {
    try {
        localStorage.setItem(
            STORAGE_KEY,
            JSON.stringify({
                location: location.value,
                checkin: checkin.value,
                nights: nights.value,
                guests: guests.value,
                railcardHolders: railcardHolders.value,
                originKey: originKey.value,
                customOrigin: customOrigin.value,
                depart: depart.value,
                top: top.value,
                maxMiles: maxMiles.value,
                maxTravelMin: maxTravelMin.value,
                rankBy: rankBy.value,
            }),
        );
    } catch {
        /* storage unavailable: nothing to do */
    }
}

function loadConfig() {
    try {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (!saved) return;
        const c = JSON.parse(saved);
        location.value = c.location ?? location.value;
        nights.value = c.nights ?? nights.value;
        guests.value = c.guests ?? guests.value;
        railcardHolders.value = c.railcardHolders ?? railcardHolders.value;
        originKey.value = c.originKey ?? originKey.value;
        customOrigin.value = c.customOrigin ?? customOrigin.value;
        depart.value = c.depart ?? depart.value;
        top.value = Math.min(40, c.top ?? top.value);
        maxMiles.value = c.maxMiles ?? maxMiles.value;
        maxTravelMin.value = c.maxTravelMin ?? maxTravelMin.value;
        if (RANKINGS.some((r) => r.key === c.rankBy)) rankBy.value = c.rankBy;
        // Only restore the date if it is still in the future.
        if (c.checkin && c.checkin >= new Date().toISOString().slice(0, 10))
            checkin.value = c.checkin;
    } catch {
        /* corrupted saved config: keep the defaults */
    }
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------
function buildRequest() {
    const body = {
        location: location.value,
        checkin: checkin.value,
        nights: Number(nights.value),
        guests: guests.value,
        railcard_holders: Number(railcardHolders.value),
        origin:
            originKey.value === CUSTOM_ORIGIN
                ? customOrigin.value
                : originKey.value,
        depart: depart.value,
        top: Number(top.value),
        max_travel_min: Number(maxTravelMin.value),
    };
    if (maxMiles.value !== "" && maxMiles.value !== null)
        body.max_miles = Number(maxMiles.value);
    return body;
}

function describeError(status, payload) {
    if (status === 429)
        return "Someone else is searching right now, try again shortly.";
    if (status === 503)
        return "Too many searches from your address, wait a minute and retry.";
    const detail = payload?.detail;
    if (Array.isArray(detail))
        return detail
            .map((d) => `${(d.loc || []).slice(-1)[0] ?? ""}: ${d.msg}`)
            .join("; ");
    if (typeof detail === "string") return detail;
    return `Search failed (${status}).`;
}

function applyEvent(ev) {
    switch (ev.event) {
        case "hotels":
            summary.value = ev.summary;
            rows.value = ev.rows;
            cached.value = !!ev.cached;
            break;
        case "transport": {
            const existing = rows.value.find((r) => r.code === ev.row.code);
            if (existing) Object.assign(existing, ev.row);
            else rows.value.push(ev.row);
            break;
        }
        case "done":
            summary.value = ev.summary;
            rows.value = ev.rows;
            cached.value = !!ev.cached;
            done.value = true;
            break;
        case "error":
            errorMsg.value = ev.detail || "Search failed.";
            break;
        default:
            break;
    }
}

async function readStream(res) {
    const decoder = new TextDecoder();
    let buffer = "";
    const handleChunk = (text, final) => {
        buffer += text;
        const lines = buffer.split("\n");
        buffer = final ? "" : lines.pop();
        for (const line of lines) {
            if (!line.trim()) continue;
            try {
                applyEvent(JSON.parse(line));
            } catch {
                /* skip a malformed line rather than abort the whole search */
            }
        }
    };
    if (!res.body?.getReader) {
        handleChunk(await res.text(), true);
        return;
    }
    const reader = res.body.getReader();
    for (;;) {
        const { value, done: finished } = await reader.read();
        if (finished) break;
        handleChunk(decoder.decode(value, { stream: true }), false);
    }
    handleChunk(decoder.decode(), true);
}

async function onSearch() {
    if (searching.value) return;
    errorMsg.value = "";
    summary.value = null;
    rows.value = [];
    cached.value = false;
    done.value = false;
    searching.value = true;
    elapsed.value = 0;
    saveConfig();
    ticker = setInterval(() => elapsed.value++, 1000);
    aborter = new AbortController();
    try {
        const res = await fetch("/py/hotels/search/stream", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(buildRequest()),
            signal: aborter.signal,
        });
        if (!res.ok) {
            let payload = null;
            try {
                payload = await res.json();
            } catch {
                /* non-JSON body (e.g. nginx error page) */
            }
            throw new Error(describeError(res.status, payload));
        }
        await readStream(res);
        if (!done.value && !errorMsg.value)
            errorMsg.value = "The search ended early; showing what came back.";
    } catch (e) {
        if (e.name !== "AbortError") errorMsg.value = e.message || "Search failed.";
    } finally {
        clearInterval(ticker);
        ticker = null;
        aborter = null;
        searching.value = false;
    }
}

async function loadOrigins() {
    try {
        const res = await fetch("/py/hotels/origins");
        if (res.ok) origins.value = await res.json();
    } catch {
        /* dropdown falls back to the default key only */
    }
    if (
        originKey.value !== CUSTOM_ORIGIN &&
        origins.value.length &&
        !origins.value.some((o) => o.key === originKey.value)
    )
        originKey.value = origins.value[0].key;
}

onMounted(() => {
    loadConfig();
    loadOrigins();
});

onBeforeUnmount(() => {
    if (ticker) clearInterval(ticker);
    if (aborter) aborter.abort();
});
</script>

<template>
    <main class="flex justify-center px-4 py-10">
        <div class="max-w-5xl w-full flex flex-col gap-6">
            <section>
                <Header>Cheap Travelodge Finder</Header>
                <Paragraph>
                    Ranks the cheapest Travelodges for a stay in London
                    <em>including the cost of getting there and back</em> from
                    a chosen station, so a £45 room forty minutes out can be
                    compared fairly with a £90 room in Zone 1. Room prices come
                    live from Travelodge; fares and routes come from
                    <InlineLink
                        href="https://tfl.gov.uk/plan-a-journey/"
                        target="_blank"
                        >TfL Journey Planner</InlineLink
                    >. Every hotel also gets a cycling time from the origin, for
                    when the bike is the cheapest transport of all.
                </Paragraph>
            </section>

            <section class="panel">
                <h2 class="panelTitle">Search</h2>
                <form class="configGrid" @submit.prevent="onSearch">
                    <label>
                        Search area
                        <input
                            type="text"
                            v-model="location"
                            maxlength="80"
                            required
                        />
                    </label>
                    <label>
                        Check-in
                        <input type="date" v-model="checkin" required />
                    </label>
                    <label>
                        Nights
                        <input
                            type="number"
                            v-model.number="nights"
                            min="1"
                            max="14"
                        />
                    </label>
                    <label>
                        Rooms &amp; guests
                        <input
                            type="text"
                            v-model="guests"
                            placeholder='"1", "2+1", "2,2+2"'
                            maxlength="40"
                        />
                        <span class="hint"
                            >rooms separated by commas, each ADULTS or
                            ADULTS+CHILDREN</span
                        >
                    </label>
                    <label>
                        Railcard holders
                        <input
                            type="number"
                            v-model.number="railcardHolders"
                            min="0"
                            max="8"
                        />
                        <span class="hint"
                            >adults getting 1/3 off off-peak rail</span
                        >
                    </label>
                    <label>
                        Origin station
                        <select v-model="originKey">
                            <option v-for="o in origins" :key="o.key" :value="o.key">
                                {{ o.name }}
                            </option>
                            <option v-if="!origins.length" value="clapham junction">
                                Clapham Junction
                            </option>
                            <option :value="CUSTOM_ORIGIN">Custom lat,lon…</option>
                        </select>
                        <input
                            v-if="originKey === CUSTOM_ORIGIN"
                            type="text"
                            v-model="customOrigin"
                            placeholder="51.5031,-0.1132"
                            class="mt-1"
                            required
                        />
                    </label>
                    <label>
                        Leave at
                        <input type="time" v-model="depart" required />
                        <span class="hint"
                            >16:00–19:00 Mon–Fri is peak (dearer, no
                            railcard)</span
                        >
                    </label>
                    <label>
                        Hotels to price
                        <input
                            type="number"
                            v-model.number="top"
                            min="1"
                            max="40"
                        />
                        <span class="hint"
                            >N cheapest by room rate; TfL allows about 24
                            hotels a minute</span
                        >
                    </label>
                    <label>
                        Max miles from search area
                        <input
                            type="number"
                            v-model="maxMiles"
                            min="0.5"
                            max="50"
                            step="0.5"
                            placeholder="any"
                        />
                    </label>
                    <label>
                        Max travel minutes
                        <input
                            type="number"
                            v-model.number="maxTravelMin"
                            min="10"
                            max="240"
                        />
                        <span class="hint"
                            >slower routes are ignored when picking the
                            cheapest</span
                        >
                    </label>

                    <div class="controls">
                        <button class="btn" type="submit" :disabled="searching">
                            {{ searching ? "Searching…" : "Search" }}
                        </button>
                        <span class="statusChip" :class="statusText.toLowerCase()">
                            {{ statusText }}
                        </span>
                        <span v-if="searching" class="stepCounter"
                            >{{ elapsed }}s</span
                        >
                        <span class="note">
                            Hotels appear within a couple of seconds; transport
                            prices fill in as TfL answers.
                        </span>
                    </div>
                </form>

                <p v-if="errorMsg" class="errorBox">{{ errorMsg }}</p>
            </section>

            <section v-if="hasResults" class="panel">
                <h2 class="panelTitle">
                    Results
                    <span v-if="cached" class="cachedTag">cached</span>
                    <span v-if="summary.truncated" class="cachedTag warn"
                        >time limit hit</span
                    >
                </h2>
                <p class="summary">{{ summaryText }}</p>

                <div class="progress" v-if="rows.length">
                    <div class="progressBar">
                        <div
                            class="progressFill"
                            :style="{ width: `${(100 * pricedCount) / rows.length}%` }"
                        ></div>
                    </div>
                    <span class="progressText">
                        {{ pricedCount }} / {{ rows.length }} priced on TfL
                    </span>
                </div>

                <div v-if="picks.length" class="picks">
                    <div v-for="p in picks" :key="p.label" class="pick">
                        <span class="pickLabel">{{ p.label }}</span>
                        <a
                            class="pickName"
                            :href="p.row.url"
                            target="_blank"
                            rel="noopener noreferrer"
                            >{{ p.row.name }}</a
                        >
                        <span class="pickMaths">{{ p.maths }}</span>
                        <span class="pickSub">{{ p.sub }}</span>
                    </div>
                </div>

                <p v-if="!rows.length" class="note">
                    No hotels with availability matched that search.
                </p>

                <template v-else>
                    <div class="rankBar">
                        <span class="rankLabel">Rank by</span>
                        <button
                            v-for="r in RANKINGS"
                            :key="r.key"
                            type="button"
                            class="rankBtn"
                            :class="{ active: rankBy === r.key }"
                            @click="rankBy = r.key; saveConfig()"
                        >
                            {{ r.label }}
                        </button>
                    </div>

                    <div class="tableWrap">
                        <table class="results">
                            <thead>
                                <tr>
                                    <th class="num">#</th>
                                    <th class="num" :class="{ sorted: rankBy === 'total' }">Total</th>
                                    <th class="num" :class="{ sorted: rankBy === 'room' }">Stay</th>
                                    <th class="num">Travel</th>
                                    <th class="num" :class="{ sorted: rankBy === 'minutes' }">Mins</th>
                                    <th class="num" :class="{ sorted: rankBy === 'cycle' }">Cycle</th>
                                    <th>Hotel</th>
                                    <th>Route (cheapest acceptable)</th>
                                </tr>
                            </thead>
                            <tbody>
                                <template v-for="(r, i) in sortedRows" :key="r.code">
                                    <tr
                                        :class="{
                                            best: i === 0,
                                            unpriced: r.total === null,
                                        }"
                                    >
                                        <td class="num">{{ i + 1 }}</td>
                                        <td class="num total">{{ gbp(r.total) }}</td>
                                        <td class="num stay">{{ gbp(r.room_price) }}</td>
                                        <td class="num">{{ gbp(r.transport_total) }}</td>
                                        <td class="num">{{ r.travel_minutes ?? "" }}</td>
                                        <td class="num cycle" :title="cycleTitle(r)">
                                            {{ cycleText(r) }}
                                        </td>
                                        <td>
                                            <a
                                                :href="r.url"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                class="hotelLink"
                                                >{{ r.name }}</a
                                            >
                                            <span v-if="r.low_availability" class="lowTag"
                                                >low availability</span
                                            >
                                        </td>
                                        <td class="route" :class="{ pending: r.route === 'pricing…' }">
                                            {{ r.route }}
                                        </td>
                                    </tr>
                                    <tr v-if="showFaster(r)" class="fasterRow">
                                        <td colspan="7"></td>
                                        <td class="route">
                                            faster: {{ r.fastest_minutes }} min for
                                            {{ gbp(r.fastest_total) }} return via
                                            {{ r.fastest_route }}
                                        </td>
                                    </tr>
                                </template>
                            </tbody>
                        </table>
                    </div>
                </template>

                <details class="help">
                    <summary>Notes</summary>
                    <ul>
                        <li>
                            Stay price is Travelodge's cheapest room(s) for the
                            whole stay and all rooms. Rank by "Stay only" to
                            ignore transport entirely.
                        </li>
                        <li>
                            Transport is the TfL adult pay-as-you-go single each
                            way, multiplied by adults. The return is assumed
                            off-peak (next day).
                        </li>
                        <li>
                            Cycle time is from the origin station on your own
                            bike. A "~" means it is estimated from the distance
                            at about 14 km/h; without a tilde it is TfL's cycle
                            route.
                        </li>
                        <li>
                            Children 10 and under travel free on TfL; 11–15s pay
                            child fares with a Zip card (not included).
                        </li>
                        <li>
                            The Railcard's 1/3 off applies to off-peak Tube and
                            rail fares only, and only when the Railcard is linked
                            to an Oyster card (not contactless).
                        </li>
                        <li>
                            Bus fares are flat and never discounted. Prices are
                            rounded to 5p. Everything is an estimate: check the
                            booking page and TfL before paying.
                        </li>
                    </ul>
                </details>
            </section>
        </div>
    </main>
</template>

<style scoped>
.panel {
    border: 1px solid var(--color-quaternary);
    background: rgba(4, 8, 15, 0.55);
    padding: 0.9rem;
}

.panelTitle {
    color: var(--color-primary);
    font-family: var(--font-heading);
    font-size: 1.4rem;
    margin-bottom: 0.5rem;
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
}

.cachedTag {
    font-size: 0.8rem;
    color: var(--color-muted);
    border: 1px solid var(--color-quaternary);
    padding: 0 0.4rem;
    letter-spacing: 0.05em;
}

.cachedTag.warn {
    color: #ffd166;
    border-color: currentColor;
}

.configGrid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.6rem 1rem;
}

.configGrid label {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    color: var(--color-muted);
    font-size: 0.85rem;
}

.configGrid input,
.configGrid select {
    background: var(--color-surface-deep);
    border: 1px solid var(--color-quaternary);
    color: #e5f4ee;
    padding: 0.35rem 0.5rem;
    font-family: monospace;
    color-scheme: dark;
}

.configGrid input:focus,
.configGrid select:focus {
    outline: 1px solid var(--color-primary);
}

.hint {
    font-size: 0.72rem;
    opacity: 0.7;
}

.controls {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.4rem;
    flex-wrap: wrap;
}

.btn {
    padding: 0.35rem 1.1rem;
    border: 1px solid var(--color-primary);
    color: var(--color-primary);
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    cursor: pointer;
    transition:
        background 0.15s ease,
        color 0.15s ease;
}

.btn:hover:enabled {
    background: var(--color-primary);
    color: var(--color-surface-deep);
}

.btn:disabled {
    opacity: 0.35;
    cursor: default;
}

.statusChip {
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    padding: 0.1rem 0.6rem;
    border: 1px solid currentColor;
    color: var(--color-muted);
}

.statusChip.searching {
    color: #ffd166;
}

.statusChip.done {
    color: var(--color-secondary);
}

.statusChip.error {
    color: var(--color-tertiary);
}

.stepCounter {
    font-family: monospace;
    color: var(--color-muted);
}

.note {
    color: var(--color-muted);
    font-size: 0.85rem;
    opacity: 0.8;
}

.errorBox {
    margin-top: 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-tertiary);
    color: var(--color-tertiary);
    font-family: monospace;
    font-size: 0.9rem;
}

.summary {
    color: var(--color-muted);
    font-size: 0.9rem;
    margin-bottom: 0.6rem;
}

.progress {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.9rem;
}

.progressBar {
    flex: 1;
    height: 4px;
    background: rgba(2, 73, 66, 0.5);
}

.progressFill {
    height: 100%;
    background: var(--color-primary);
    transition: width 0.3s ease;
}

.progressText {
    font-family: monospace;
    font-size: 0.8rem;
    color: var(--color-muted);
    white-space: nowrap;
}

.picks {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.6rem;
    margin-bottom: 0.9rem;
}

.pick {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    border: 1px solid var(--color-primary);
    padding: 0.5rem 0.75rem;
    background: rgba(85, 255, 187, 0.06);
    min-width: 0;
}

.pickLabel {
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    color: var(--color-primary);
    font-size: 0.85rem;
}

.pickName {
    color: #e5f4ee;
    font-weight: bold;
    text-decoration: none;
    transition: color 0.15s ease;
}

.pickName:hover {
    color: var(--color-tertiary);
}

.pickMaths {
    font-family: monospace;
    color: var(--color-secondary);
    font-size: 0.85rem;
}

.pickSub {
    font-size: 0.78rem;
    color: var(--color-muted);
}

.rankBar {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-wrap: wrap;
    margin-bottom: 0.6rem;
}

.rankLabel {
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    color: var(--color-muted);
    font-size: 0.85rem;
    margin-right: 0.2rem;
}

.rankBtn {
    padding: 0.15rem 0.6rem;
    border: 1px solid var(--color-quaternary);
    color: var(--color-muted);
    font-size: 0.8rem;
    cursor: pointer;
    transition:
        color 0.15s ease,
        border-color 0.15s ease;
}

.rankBtn:hover {
    color: var(--color-primary);
}

.rankBtn.active {
    color: var(--color-primary);
    border-color: var(--color-primary);
}

.hotelLink {
    color: var(--color-primary);
    text-decoration: none;
    transition: color 0.15s ease;
}

.hotelLink:hover {
    color: var(--color-tertiary);
}

.tableWrap {
    overflow-x: auto;
}

.results {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
    color: #e5f4ee;
}

.results th {
    text-align: left;
    color: var(--color-primary);
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    font-weight: normal;
    border-bottom: 1px solid var(--color-quaternary);
    padding: 0.3rem 0.5rem;
    white-space: nowrap;
}

.results th.sorted {
    text-decoration: underline;
    text-underline-offset: 3px;
}

.results td {
    padding: 0.3rem 0.5rem;
    border-bottom: 1px solid rgba(2, 73, 66, 0.4);
    vertical-align: top;
}

.results .num {
    text-align: right;
    font-family: monospace;
    white-space: nowrap;
}

.results .total {
    color: var(--color-secondary);
}

.results .cycle {
    color: var(--color-muted);
}

.results tr.best td {
    background: rgba(85, 255, 187, 0.08);
}

.results tr.unpriced td.total,
.results tr.unpriced td.route {
    opacity: 0.6;
}

.results .route {
    color: var(--color-muted);
}

.results .route.pending {
    font-style: italic;
    opacity: 0.5;
}

.fasterRow td {
    border-bottom: 1px solid rgba(2, 73, 66, 0.4);
    font-size: 0.8rem;
    opacity: 0.75;
    padding-top: 0;
}

.fasterRow td:first-child {
    border-bottom: none;
}

.lowTag {
    margin-left: 0.4rem;
    font-size: 0.72rem;
    color: var(--color-tertiary);
    border: 1px solid currentColor;
    padding: 0 0.3rem;
    white-space: nowrap;
}

.help {
    margin-top: 0.9rem;
    color: var(--color-muted);
    font-size: 0.9rem;
}

.help summary {
    cursor: pointer;
    color: var(--color-primary);
    font-family: var(--font-heading);
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
    .picks {
        grid-template-columns: 1fr;
    }
}
</style>
