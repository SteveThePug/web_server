// Formatting and sorting helpers shared by the hotel panels.

export const PENDING_ROUTE = "pricing…";

export function gbp(v) {
    return v === null || v === undefined ? "?" : `£${v.toFixed(2)}`;
}

export function fmtDate(iso) {
    const d = new Date(iso + "T00:00:00");
    return d.toLocaleDateString("en-GB", {
        weekday: "short",
        day: "numeric",
        month: "short",
    });
}

export function fmtTime(iso) {
    return new Date(iso).toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
    });
}

export function cycleText(r) {
    if (r.cycle_minutes === null || r.cycle_minutes === undefined) return "";
    return `${r.cycle_source === "estimate" ? "~" : ""}${r.cycle_minutes} min`;
}

export function cycleTitle(r) {
    if (r.cycle_minutes === null || r.cycle_minutes === undefined) return "";
    return r.cycle_source === "estimate"
        ? `about ${r.cycle_km} km; estimated from distance at a gentle pace`
        : `${r.cycle_km} km on TfL's cycle route`;
}

export function showFaster(r) {
    return (
        r.fastest_minutes !== null &&
        r.fastest_minutes !== undefined &&
        r.travel_minutes !== null &&
        r.travel_minutes !== undefined &&
        r.fastest_minutes < r.travel_minutes - 10
    );
}

/** Comparator on a numeric getter; nulls sort last, ties fall through to `tiebreak`. */
export function byNumber(get, tiebreak) {
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

export const byRoom = byNumber((r) => r.room_price);
export const byTotal = byNumber((r) => r.total, byRoom);
export const byCycle = byNumber((r) => r.cycle_minutes, byRoom);
export const byMinutes = byNumber((r) => r.travel_minutes, byTotal);
