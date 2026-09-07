<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import Header from "@/components/text/Header.vue";
import Paragraph from "@/components/text/Paragraph.vue";
import InlineLink from "@/components/text/InlineLink.vue";

const STORAGE_KEY = "cheap-hotels-config";
const CUSTOM_ORIGIN = "__custom__";

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
const top = ref(15);
const maxMiles = ref("");
const maxTravelMin = ref(75);

const origins = ref([]);

// ---------------------------------------------------------------------------
// Search state
// ---------------------------------------------------------------------------
const searching = ref(false);
const errorMsg = ref("");
const result = ref(null);
const elapsed = ref(0);
let ticker = null;

const statusText = computed(() => {
    if (errorMsg.value) return "Error";
    if (searching.value) return "Searching";
    if (result.value) return "Done";
    return "Idle";
});

const bestRow = computed(
    () => result.value?.rows.find((r) => r.total !== null) ?? null,
);

const summaryText = computed(() => {
    const s = result.value?.summary;
    if (!s) return "";
    const party =
        `${s.adults} adult${s.adults === 1 ? "" : "s"}` +
        (s.children ? ` + ${s.children} child${s.children === 1 ? "" : "ren"}` : "") +
        ` in ${s.rooms} room${s.rooms === 1 ? "" : "s"}`;
    const railcard = s.railcard_holders
        ? ` (${s.railcard_holders} with railcard)`
        : "";
    return (
        `${s.hotels_priced} of ${s.hotels_available} available Travelodges near ` +
        `'${s.location}', ${fmtDate(s.checkin)} → ${fmtDate(s.checkout)} ` +
        `(${s.nights} night${s.nights === 1 ? "" : "s"}, ${party}), ` +
        `incl. round-trip transport from ${s.origin_name} departing ${fmtTime(s.depart)}${railcard}.`
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
        top.value = c.top ?? top.value;
        maxMiles.value = c.maxMiles ?? maxMiles.value;
        maxTravelMin.value = c.maxTravelMin ?? maxTravelMin.value;
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
    const detail = payload?.detail;
    if (Array.isArray(detail))
        return detail
            .map((d) => `${(d.loc || []).slice(-1)[0] ?? ""}: ${d.msg}`)
            .join("; ");
    if (typeof detail === "string") return detail;
    return `Search failed (${status}).`;
}

async function onSearch() {
    if (searching.value) return;
    errorMsg.value = "";
    result.value = null;
    searching.value = true;
    elapsed.value = 0;
    saveConfig();
    ticker = setInterval(() => elapsed.value++, 1000);
    try {
        const res = await fetch("/py/hotels/search", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(buildRequest()),
        });
        let payload = null;
        try {
            payload = await res.json();
        } catch {
            /* non-JSON body (e.g. nginx timeout page) */
        }
        if (!res.ok) throw new Error(describeError(res.status, payload));
        result.value = payload;
    } catch (e) {
        errorMsg.value = e.message || "Search failed.";
    } finally {
        clearInterval(ticker);
        ticker = null;
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
});
</script>

<template>
    <main class="flex justify-center px-4 py-10">
        <div class="max-w-4xl w-full flex flex-col gap-6">
            <section>
                <Header>Cheap Travelodge Finder</Header>
                <Paragraph>
                    Finds the cheapest Travelodge for a stay in London
                    <em>including the cost of getting there and back</em> from
                    a chosen station, so a £45 room forty minutes out can be
                    compared fairly with a £90 room in Zone 1. Room prices come
                    live from Travelodge; fares and routes come from
                    <InlineLink
                        href="https://tfl.gov.uk/plan-a-journey/"
                        target="_blank"
                        >TfL Journey Planner</InlineLink
                    >.
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
                            max="25"
                        />
                        <span class="hint"
                            >N cheapest by room rate; fewer is faster</span
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
                            Takes about a minute: every hotel is priced on TfL
                            twice.
                        </span>
                    </div>
                </form>

                <p v-if="errorMsg" class="errorBox">{{ errorMsg }}</p>
            </section>

            <section v-if="result" class="panel">
                <h2 class="panelTitle">
                    Results
                    <span v-if="result.cached" class="cachedTag">cached</span>
                </h2>
                <p class="summary">{{ summaryText }}</p>

                <div v-if="bestRow" class="best">
                    <span class="bestLabel">Best value</span>
                    <span class="bestName">{{ bestRow.name }}</span>
                    <span class="bestMaths">
                        {{ gbp(bestRow.room_price) }} stay +
                        {{ gbp(bestRow.transport_total) }} travel =
                        <strong>{{ gbp(bestRow.total) }}</strong>
                    </span>
                    <a
                        class="bookLink"
                        :href="bestRow.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        >Book →</a
                    >
                </div>

                <p v-if="!result.rows.length" class="note">
                    No hotels with availability matched that search.
                </p>

                <div v-else class="tableWrap">
                    <table class="results">
                        <thead>
                            <tr>
                                <th class="num">#</th>
                                <th class="num">Total</th>
                                <th class="num">Stay</th>
                                <th class="num">Travel</th>
                                <th class="num">Mins</th>
                                <th>Hotel</th>
                                <th>Route (cheapest acceptable)</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-for="(r, i) in result.rows" :key="r.code">
                                <tr :class="{ best: r === bestRow, unpriced: r.total === null }">
                                    <td class="num">{{ i + 1 }}</td>
                                    <td class="num total">{{ gbp(r.total) }}</td>
                                    <td class="num">{{ gbp(r.room_price) }}</td>
                                    <td class="num">{{ gbp(r.transport_total) }}</td>
                                    <td class="num">{{ r.travel_minutes ?? "" }}</td>
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
                                    <td class="route">{{ r.route }}</td>
                                </tr>
                                <tr v-if="showFaster(r)" class="fasterRow">
                                    <td colspan="6"></td>
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

                <details class="help">
                    <summary>Notes</summary>
                    <ul>
                        <li>
                            Stay price is Travelodge's cheapest room(s) for the
                            whole stay and all rooms.
                        </li>
                        <li>
                            Transport is the TfL adult pay-as-you-go single each
                            way, multiplied by adults. The return is assumed
                            off-peak (next day).
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
    margin-bottom: 0.75rem;
}

.best {
    display: flex;
    align-items: baseline;
    gap: 0.75rem;
    flex-wrap: wrap;
    border: 1px solid var(--color-primary);
    padding: 0.5rem 0.75rem;
    margin-bottom: 0.9rem;
    background: rgba(85, 255, 187, 0.06);
}

.bestLabel {
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
    color: var(--color-primary);
}

.bestName {
    color: #e5f4ee;
    font-weight: bold;
}

.bestMaths {
    font-family: monospace;
    color: var(--color-muted);
    font-size: 0.9rem;
}

.bestMaths strong {
    color: var(--color-primary);
}

.bookLink,
.hotelLink {
    color: var(--color-primary);
    text-decoration: none;
    transition: color 0.15s ease;
}

.bookLink {
    margin-left: auto;
    font-family: var(--font-heading);
    letter-spacing: 0.05em;
}

.bookLink:hover,
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

.results tr.best td {
    background: rgba(85, 255, 187, 0.08);
}

.results tr.unpriced td {
    opacity: 0.6;
}

.results .route {
    color: var(--color-muted);
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
    .configGrid {
        grid-template-columns: 1fr;
    }

    .bookLink {
        margin-left: 0;
    }
}
</style>
