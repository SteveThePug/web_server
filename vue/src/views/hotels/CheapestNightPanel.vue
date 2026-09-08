<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import SharedFields from "./SharedFields.vue";
import HotelRows from "./HotelRows.vue";
import { runStream } from "./hotelsStream.js";
import { todayIso, addDays, loadFrom, saveTo } from "./useHotelForm.js";
import {
    gbp,
    fmtDate,
    fmtTime,
    cycleText,
    byNumber,
    byRoom,
    byTotal,
    PENDING_ROUTE,
} from "./hotelsFormat.js";

const props = defineProps({
    form: { type: Object, required: true },
});

const SCAN_KEY = "cheap-hotels-scan";
const MAX_SPAN_DAYS = 91;
const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const ALL_DAYS = [0, 1, 2, 3, 4, 5, 6];
const WEEKEND = [4, 5];

const RANKINGS = [
    { key: "total", label: "Stay + transport" },
    { key: "room", label: "Stay only" },
    { key: "date", label: "Date" },
];

// ---------------------------------------------------------------------------
// Form state (this panel's own fields; shared ones live in props.form)
// ---------------------------------------------------------------------------
const start = ref(todayIso());
const end = ref(addDays(start.value, 60));
const weekdays = ref([...ALL_DAYS]);
const nights = ref(1);
const perNight = ref(5);
const rankBy = ref("total");

const endMax = computed(() => addDays(start.value, MAX_SPAN_DAYS));

function setDays(list) {
    weekdays.value = [...list];
}

// ---------------------------------------------------------------------------
// Scan state
// ---------------------------------------------------------------------------
const searching = ref(false);
const errorMsg = ref("");
const summary = ref(null);
const nightRows = ref([]);
const cached = ref(false);
const done = ref(false);
const elapsed = ref(0);
const expanded = ref(new Set());
let transportByCode = {};
let ticker = null;
let aborter = null;

const statusText = computed(() => {
    if (errorMsg.value) return "Error";
    if (searching.value) return "Scanning";
    if (done.value) return "Done";
    return "Idle";
});

const hasResults = computed(() => summary.value !== null);

const fetchedCount = computed(() => {
    const s = summary.value;
    return s ? s.nights_fetched + s.nights_failed + s.nights_skipped : 0;
});

// Best hotel per night (by total, then room) as transport fills in.
const bestByNight = computed(() => {
    const m = new Map();
    for (const n of nightRows.value) {
        if (!n.hotels.length) continue;
        m.set(n.checkin, n.hotels.reduce((b, h) => (byTotal(h, b) < 0 ? h : b)));
    }
    return m;
});

function bestOf(n) {
    return bestByNight.value.get(n.checkin) ?? null;
}

function nightTotal(n) {
    return bestOf(n)?.total ?? null;
}

function nightRoom(n) {
    if (!n.hotels.length) return null;
    return n.hotels.reduce((m, h) => Math.min(m, h.room_price), Infinity);
}

function isWeekend(n) {
    const d = new Date(n.checkin + "T00:00:00").getDay();
    return d === 5 || d === 6;
}

const byDate = (a, b) => (a.checkin < b.checkin ? -1 : a.checkin > b.checkin ? 1 : 0);
const byNightRoom = byNumber(nightRoom, byDate);
const byNightTotal = byNumber(nightTotal, byNightRoom);

const sortedNights = computed(() => {
    const cmp = { total: byNightTotal, room: byNightRoom, date: byDate }[rankBy.value];
    const ok = nightRows.value.filter((n) => n.status === "ok" && n.hotels.length);
    const rest = nightRows.value.filter((n) => !(n.status === "ok" && n.hotels.length));
    return [...ok.sort(cmp), ...rest.sort(byDate)];
});

function sortedHotels(n) {
    return [...n.hotels].sort(rankBy.value === "room" ? byRoom : byTotal);
}

function firstNight(cmp, ok) {
    const list = nightRows.value.filter((n) => n.hotels.length && ok(n));
    if (!list.length) return null;
    return list.reduce((best, n) => (cmp(n, best) < 0 ? n : best));
}

function describeNight(n, h) {
    return h.total != null
        ? `${gbp(h.room_price)} stay + ${gbp(h.transport_total)} travel = ${gbp(h.total)}`
        : `${gbp(h.room_price)} stay, transport not priced yet`;
}

const picks = computed(() => {
    const p = [];
    const cheapest = firstNight(byNightTotal, (n) => nightTotal(n) !== null);
    if (cheapest) {
        const h = bestOf(cheapest);
        p.push({
            label: "Cheapest night",
            night: cheapest,
            row: h,
            maths: describeNight(cheapest, h),
            sub: `${h.travel_minutes} min by ${h.route}`,
        });
    }
    const room = firstNight(byNightRoom, () => true);
    if (room) {
        const h = sortedHotels(room).sort(byRoom)[0];
        p.push({
            label: "Cheapest room",
            night: room,
            row: h,
            maths: `${gbp(h.room_price)} for the stay`,
            sub:
                h.total != null
                    ? `${gbp(h.total)} with transport (${h.travel_minutes} min)`
                    : "transport not priced yet",
        });
    }
    const weekend = firstNight(byNightTotal, (n) => isWeekend(n) && nightTotal(n) !== null);
    if (weekend && weekend !== cheapest) {
        const h = bestOf(weekend);
        p.push({
            label: "Cheapest Fri/Sat night",
            night: weekend,
            row: h,
            maths: describeNight(weekend, h),
            sub: `${h.travel_minutes} min by ${h.route}`,
        });
    }
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
    const days =
        s.weekdays.length === 7
            ? "any day"
            : s.weekdays.map((d) => DAYS[d]).join("/") + " check-ins";
    return (
        `${s.nights_candidate} candidate ${s.nights}-night stays near '${s.location}' ` +
        `between ${fmtDate(s.start)} and ${fmtDate(s.end)} (${days}, ${party}), ` +
        `keeping the ${s.per_night} cheapest hotels per night, with round-trip transport ` +
        `from ${s.origin_name} departing ${fmtTime(s.depart)}${railcard}.`
    );
});

function nightLabel(n) {
    return `${fmtDate(n.checkin)}${n.checkout !== addDays(n.checkin, 1) ? " → " + fmtDate(n.checkout) : ""}`;
}

function nightNote(n) {
    if (n.status === "pending") return "fetching…";
    if (n.status === "failed") return `failed: ${n.detail}`;
    if (n.status === "skipped") return n.detail || "skipped";
    if (!n.hotels.length) return "no availability";
    return "";
}

function toggle(n) {
    if (!n.hotels.length) return;
    if (expanded.value.has(n.checkin)) expanded.value.delete(n.checkin);
    else expanded.value.add(n.checkin);
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------
function saveConfig() {
    props.form.save();
    saveTo(SCAN_KEY, {
        start: start.value,
        end: end.value,
        weekdays: weekdays.value,
        nights: nights.value,
        perNight: perNight.value,
        rankBy: rankBy.value,
    });
}

function loadConfig() {
    const c = loadFrom(SCAN_KEY);
    if (!c) return;
    const today = todayIso();
    if (c.start && c.start >= today) start.value = c.start;
    if (c.end && c.end >= start.value && c.end <= endMax.value) end.value = c.end;
    if (Array.isArray(c.weekdays) && c.weekdays.length)
        weekdays.value = c.weekdays.filter((d) => ALL_DAYS.includes(d));
    nights.value = Math.min(7, Math.max(1, c.nights ?? nights.value));
    perNight.value = Math.min(10, Math.max(1, c.perNight ?? perNight.value));
    if (RANKINGS.some((r) => r.key === c.rankBy)) rankBy.value = c.rankBy;
}

// ---------------------------------------------------------------------------
// Scan
// ---------------------------------------------------------------------------
function buildRequest() {
    return {
        ...props.form.sharedBody(),
        start: start.value,
        end: end.value,
        weekdays: [...weekdays.value].sort(),
        nights: Number(nights.value),
        per_night: Number(perNight.value),
    };
}

function applyTransport(h, fields) {
    if (!fields) return;
    Object.assign(h, fields);
    h.total =
        h.transport_total !== null && h.transport_total !== undefined
            ? Math.round((h.room_price + h.transport_total) * 100) / 100
            : null;
}

function applyEvent(ev) {
    switch (ev.event) {
        case "scan":
            summary.value = ev.summary;
            nightRows.value = ev.nights;
            cached.value = !!ev.cached;
            break;
        case "night": {
            summary.value = ev.summary;
            const night = ev.night;
            for (const h of night.hotels) applyTransport(h, transportByCode[h.code]);
            const i = nightRows.value.findIndex((n) => n.checkin === night.checkin);
            if (i >= 0) nightRows.value[i] = night;
            else nightRows.value.push(night);
            break;
        }
        case "transport": {
            summary.value = ev.summary;
            const { code, ...fields } = ev.transport;
            transportByCode[code] = fields;
            for (const n of nightRows.value)
                for (const h of n.hotels) if (h.code === code) applyTransport(h, fields);
            break;
        }
        case "done":
            summary.value = ev.summary;
            nightRows.value = ev.nights;
            cached.value = !!ev.cached;
            done.value = true;
            break;
        case "error":
            errorMsg.value = ev.detail || "Scan failed.";
            break;
        default:
            break;
    }
}

async function onScan() {
    if (searching.value) return;
    errorMsg.value = "";
    summary.value = null;
    nightRows.value = [];
    transportByCode = {};
    expanded.value = new Set();
    cached.value = false;
    done.value = false;
    searching.value = true;
    elapsed.value = 0;
    saveConfig();
    ticker = setInterval(() => elapsed.value++, 1000);
    aborter = new AbortController();
    try {
        await runStream("/py/hotels/scan/stream", buildRequest(), {
            signal: aborter.signal,
            onEvent: applyEvent,
        });
        if (!done.value && !errorMsg.value)
            errorMsg.value = "The scan ended early; showing what came back.";
    } catch (e) {
        if (e.name !== "AbortError") errorMsg.value = e.message || "Scan failed.";
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
        <h2 class="panelTitle">Scan a date range</h2>
        <form class="configGrid" @submit.prevent="onScan">
            <SharedFields :form="form" />
            <label>
                First check-in
                <input type="date" v-model="start" :min="todayIso()" required />
            </label>
            <label>
                Last check-in
                <input type="date" v-model="end" :min="start" :max="endMax" required />
                <span class="hint">up to about 3 months after the first</span>
            </label>
            <label>
                Nights per stay
                <input type="number" v-model.number="nights" min="1" max="7" />
            </label>
            <label>
                Hotels per night
                <input type="number" v-model.number="perNight" min="1" max="10" />
                <span class="hint"
                    >N cheapest by room rate each night; transport is priced
                    once per hotel</span
                >
            </label>
            <div class="dayRow">
                <span class="rankLabel">Check-in days</span>
                <label v-for="(d, i) in DAYS" :key="d" class="chk">
                    <input type="checkbox" :value="i" v-model="weekdays" />
                    {{ d }}
                </label>
                <button type="button" class="rankBtn" @click="setDays(ALL_DAYS)">all</button>
                <button type="button" class="rankBtn" @click="setDays(WEEKEND)">Fri + Sat</button>
            </div>

            <div class="controls">
                <button
                    class="btn"
                    type="submit"
                    :disabled="searching || !weekdays.length"
                >
                    {{ searching ? "Scanning…" : "Scan" }}
                </button>
                <span class="statusChip" :class="statusText.toLowerCase()">
                    {{ statusText }}
                </span>
                <span v-if="searching" class="stepCounter">{{ elapsed }}s</span>
                <span class="note">
                    Nights fill in over a minute or so; TfL prices for each
                    hotel follow. Shares the single search slot, so a scan can
                    take a few minutes.
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

        <div class="progress" v-if="nightRows.length">
            <div class="progressBar">
                <div
                    class="progressFill"
                    :style="{ width: `${(100 * fetchedCount) / summary.nights_candidate}%` }"
                ></div>
            </div>
            <span class="progressText">
                {{ fetchedCount }} / {{ summary.nights_candidate }} nights fetched
            </span>
        </div>
        <div class="progress" v-if="summary.hotels_unique">
            <div class="progressBar">
                <div
                    class="progressFill"
                    :style="{ width: `${(100 * summary.hotels_priced) / summary.hotels_unique}%` }"
                ></div>
            </div>
            <span class="progressText">
                {{ summary.hotels_priced }} / {{ summary.hotels_unique }} hotels priced on TfL
            </span>
        </div>

        <div v-if="picks.length" class="picks">
            <div v-for="p in picks" :key="p.label" class="pick">
                <span class="pickLabel">{{ p.label }}</span>
                <span class="date">{{ nightLabel(p.night) }}</span>
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

        <p v-if="!nightRows.length" class="note">No candidate nights in that range.</p>

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
                <span class="note">click a night to see its hotels</span>
            </div>

            <div class="tableWrap">
                <table class="results">
                    <thead>
                        <tr>
                            <th class="num">#</th>
                            <th :class="{ sorted: rankBy === 'date' }">Night</th>
                            <th class="num" :class="{ sorted: rankBy === 'total' }">Total</th>
                            <th class="num" :class="{ sorted: rankBy === 'room' }">Stay</th>
                            <th class="num">Travel</th>
                            <th class="num">Mins</th>
                            <th class="num">Cycle</th>
                            <th>Cheapest hotel</th>
                            <th>Route (cheapest acceptable)</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <template v-for="(n, i) in sortedNights" :key="n.checkin">
                            <tr
                                class="nightRow"
                                :class="{
                                    best: i === 0 && nightTotal(n) !== null,
                                    muted: !n.hotels.length,
                                    open: expanded.has(n.checkin),
                                    unpriced: nightTotal(n) === null,
                                }"
                                @click="toggle(n)"
                            >
                                <td class="num">{{ i + 1 }}</td>
                                <td class="date">{{ nightLabel(n) }}</td>
                                <template v-if="bestOf(n)">
                                    <td class="num total">{{ gbp(bestOf(n).total) }}</td>
                                    <td class="num stay">{{ gbp(bestOf(n).room_price) }}</td>
                                    <td class="num">{{ gbp(bestOf(n).transport_total) }}</td>
                                    <td class="num">{{ bestOf(n).travel_minutes ?? "" }}</td>
                                    <td class="num cycle">{{ cycleText(bestOf(n)) }}</td>
                                    <td>
                                        <a
                                            :href="bestOf(n).url"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="hotelLink"
                                            @click.stop
                                            >{{ bestOf(n).name }}</a
                                        >
                                        <span v-if="bestOf(n).low_availability" class="lowTag"
                                            >low availability</span
                                        >
                                    </td>
                                    <td
                                        class="route"
                                        :class="{ pending: bestOf(n).route === PENDING_ROUTE }"
                                    >
                                        {{ bestOf(n).route }}
                                    </td>
                                    <td class="num"><span class="chev">▶</span></td>
                                </template>
                                <template v-else>
                                    <td colspan="8" class="route">{{ nightNote(n) }}</td>
                                </template>
                            </tr>
                            <tr v-if="expanded.has(n.checkin)" class="expandRow">
                                <td colspan="10">
                                    <table class="results sub">
                                        <thead>
                                            <tr>
                                                <th class="num">#</th>
                                                <th class="num">Total</th>
                                                <th class="num">Stay</th>
                                                <th class="num">Travel</th>
                                                <th class="num">Mins</th>
                                                <th class="num">Cycle</th>
                                                <th>Hotel</th>
                                                <th>Route</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <HotelRows :rows="sortedHotels(n)" :highlight-first="false" />
                                        </tbody>
                                    </table>
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
                    Each night keeps only its N cheapest hotels by room rate, so
                    a dearer room with cheap transport can be missed; raise
                    "Hotels per night" to widen the net.
                </li>
                <li>
                    Transport is priced once per hotel at the first candidate
                    date's departure time and reused for every night. Fares
                    are the same every day off-peak; routes can differ at
                    weekends.
                </li>
                <li>
                    Room prices may be up to 30 minutes old. A night marked
                    "failed" can be retried by scanning again; "skipped" means
                    the time limit was reached first, so try a shorter range.
                </li>
                <li>
                    Everything else (railcard, buses, children, cycling) works
                    as in the single-stay search: see its notes.
                </li>
            </ul>
        </details>
    </section>
</template>
