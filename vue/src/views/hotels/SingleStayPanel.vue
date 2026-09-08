<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import SharedFields from "./SharedFields.vue";
import HotelRows from "./HotelRows.vue";
import { runStream } from "./hotelsStream.js";
import { nextFriday, todayIso } from "./useHotelForm.js";
import {
    gbp,
    fmtDate,
    fmtTime,
    cycleText,
    byRoom,
    byTotal,
    byCycle,
    byMinutes,
    PENDING_ROUTE,
} from "./hotelsFormat.js";

const props = defineProps({
    form: { type: Object, required: true },
});

const RANKINGS = [
    { key: "total", label: "Stay + transport" },
    { key: "room", label: "Stay only" },
    { key: "cycle", label: "Cycle time" },
    { key: "minutes", label: "Travel time" },
];

// ---------------------------------------------------------------------------
// Form state (this panel's own fields; shared ones live in props.form)
// ---------------------------------------------------------------------------
const checkin = ref(nextFriday());
const nights = ref(1);
const top = ref(20);
const rankBy = ref("total");

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
    () => rows.value.filter((r) => r.total != null || r.route !== PENDING_ROUTE).length,
);

const hasResults = computed(() => summary.value !== null);

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
    const best = firstBy(byTotal, (r) => r.total != null);
    if (best)
        p.push({
            label: "Best value",
            row: best,
            maths: `${gbp(best.room_price)} stay + ${gbp(best.transport_total)} travel = ${gbp(best.total)}`,
            sub: `${best.travel_minutes} min by ${best.route}`,
        });
    const room = firstBy(byRoom, (r) => r.room_price != null);
    if (room)
        p.push({
            label: "Cheapest room",
            row: room,
            maths: `${gbp(room.room_price)} for the stay`,
            sub:
                room.total != null
                    ? `${gbp(room.total)} with transport (${room.travel_minutes} min)`
                    : "transport not priced yet",
        });
    const cycle = firstBy(byCycle, (r) => r.cycle_minutes != null);
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
// Persistence
// ---------------------------------------------------------------------------
function saveConfig() {
    props.form.save({
        checkin: checkin.value,
        nights: nights.value,
        top: top.value,
        rankBy: rankBy.value,
    });
}

function loadConfig() {
    const c = props.form.stored();
    nights.value = c.nights ?? nights.value;
    top.value = Math.min(40, c.top ?? top.value);
    if (RANKINGS.some((r) => r.key === c.rankBy)) rankBy.value = c.rankBy;
    // Only restore the date if it is still in the future.
    if (c.checkin && c.checkin >= todayIso()) checkin.value = c.checkin;
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------
function buildRequest() {
    return {
        ...props.form.sharedBody(),
        checkin: checkin.value,
        nights: Number(nights.value),
        top: Number(top.value),
    };
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
        await runStream("/py/hotels/search/stream", buildRequest(), {
            signal: aborter.signal,
            onEvent: applyEvent,
        });
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

onMounted(loadConfig);

onBeforeUnmount(() => {
    if (ticker) clearInterval(ticker);
    if (aborter) aborter.abort();
});
</script>

<template>
    <section class="panel">
        <h2 class="panelTitle">Search</h2>
        <form class="configGrid" @submit.prevent="onSearch">
            <SharedFields :form="form" />
            <label>
                Check-in
                <input type="date" v-model="checkin" required />
            </label>
            <label>
                Nights
                <input type="number" v-model.number="nights" min="1" max="14" />
            </label>
            <label>
                Hotels to price
                <input type="number" v-model.number="top" min="1" max="40" />
                <span class="hint"
                    >N cheapest by room rate; TfL allows about 24 hotels a
                    minute</span
                >
            </label>

            <div class="controls">
                <button class="btn" type="submit" :disabled="searching">
                    {{ searching ? "Searching…" : "Search" }}
                </button>
                <span class="statusChip" :class="statusText.toLowerCase()">
                    {{ statusText }}
                </span>
                <span v-if="searching" class="stepCounter">{{ elapsed }}s</span>
                <span class="note">
                    Hotels appear within a couple of seconds; transport prices
                    fill in as TfL answers.
                </span>
            </div>
        </form>

        <p v-if="errorMsg" class="errorBox">{{ errorMsg }}</p>
    </section>

    <section v-if="hasResults" class="panel">
        <h2 class="panelTitle">
            Results
            <span v-if="cached" class="cachedTag">cached</span>
            <span v-if="summary.truncated" class="cachedTag warn">time limit hit</span>
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
                        <HotelRows :rows="sortedRows" />
                    </tbody>
                </table>
            </div>
        </template>

        <details class="help">
            <summary>Notes</summary>
            <ul>
                <li>
                    Stay price is Travelodge's cheapest room(s) for the whole
                    stay and all rooms. Rank by "Stay only" to ignore transport
                    entirely.
                </li>
                <li>
                    Transport is the TfL adult pay-as-you-go single each way,
                    multiplied by adults. The return is assumed off-peak (next
                    day).
                </li>
                <li>
                    Cycle time is from the origin station on your own bike. A
                    "~" means it is estimated from the distance at about 14
                    km/h; without a tilde it is TfL's cycle route.
                </li>
                <li>
                    Children 10 and under travel free on TfL; 11–15s pay child
                    fares with a Zip card (not included).
                </li>
                <li>
                    The Railcard's 1/3 off applies to off-peak Tube and rail
                    fares only, and only when the Railcard is linked to an
                    Oyster card (not contactless).
                </li>
                <li>
                    Bus fares are flat and never discounted. Prices are rounded
                    to 5p. Room prices may be up to 30 minutes old. Everything
                    is an estimate: check the booking page and TfL before
                    paying.
                </li>
            </ul>
        </details>
    </section>
</template>
