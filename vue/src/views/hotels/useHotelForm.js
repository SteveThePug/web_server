// Form state shared by the "Single stay" and "Cheapest night" panels: where
// you start from, who is travelling and how far you are willing to go. Kept
// in one reactive object owned by CheapHotels.vue so values survive tab
// switches, and persisted under the same localStorage key (and field names)
// the page has always used.

import { reactive, ref } from "vue";

export const STORAGE_KEY = "cheap-hotels-config";
export const CUSTOM_ORIGIN = "__custom__";

export function todayIso() {
    return new Date().toISOString().slice(0, 10);
}

export function addDays(iso, n) {
    const d = new Date(iso + "T00:00:00");
    d.setDate(d.getDate() + n);
    return d.toISOString().slice(0, 10);
}

export function nextFriday() {
    const d = new Date();
    const delta = (5 - d.getDay() + 7) % 7 || 7;
    d.setDate(d.getDate() + delta);
    return d.toISOString().slice(0, 10);
}

export function loadFrom(key) {
    try {
        const saved = localStorage.getItem(key);
        return saved ? JSON.parse(saved) : null;
    } catch {
        return null; // storage unavailable or corrupted: caller keeps defaults
    }
}

export function saveTo(key, value) {
    try {
        localStorage.setItem(key, JSON.stringify(value));
    } catch {
        /* storage unavailable: nothing to do */
    }
}

export function useHotelForm() {
    const form = reactive({
        location: ref("Central London"),
        guests: ref("1"),
        railcardHolders: ref(1),
        originKey: ref("clapham junction"),
        customOrigin: ref(""),
        depart: ref("19:30"),
        maxMiles: ref(""),
        maxTravelMin: ref(75),
        origins: ref([]),

        originValue() {
            return form.originKey === CUSTOM_ORIGIN
                ? form.customOrigin
                : form.originKey;
        },

        /** The shared part of a search or scan request body. */
        sharedBody() {
            const body = {
                location: form.location,
                guests: form.guests,
                railcard_holders: Number(form.railcardHolders),
                origin: form.originValue(),
                depart: form.depart,
                max_travel_min: Number(form.maxTravelMin),
            };
            if (form.maxMiles !== "" && form.maxMiles !== null)
                body.max_miles = Number(form.maxMiles);
            return body;
        },

        /** Stored config object (shared + per-panel fields), or {}. */
        stored() {
            return loadFrom(STORAGE_KEY) ?? {};
        },

        /** Persist the shared fields, merged with any per-panel `extra`. */
        save(extra = {}) {
            saveTo(STORAGE_KEY, {
                ...form.stored(),
                location: form.location,
                guests: form.guests,
                railcardHolders: form.railcardHolders,
                originKey: form.originKey,
                customOrigin: form.customOrigin,
                depart: form.depart,
                maxMiles: form.maxMiles,
                maxTravelMin: form.maxTravelMin,
                ...extra,
            });
        },

        /** Restore the shared fields from storage. */
        load() {
            const c = form.stored();
            form.location = c.location ?? form.location;
            form.guests = c.guests ?? form.guests;
            form.railcardHolders = c.railcardHolders ?? form.railcardHolders;
            form.originKey = c.originKey ?? form.originKey;
            form.customOrigin = c.customOrigin ?? form.customOrigin;
            form.depart = c.depart ?? form.depart;
            form.maxMiles = c.maxMiles ?? form.maxMiles;
            form.maxTravelMin = c.maxTravelMin ?? form.maxTravelMin;
        },

        async loadOrigins() {
            try {
                const res = await fetch("/py/hotels/origins");
                if (res.ok) form.origins = await res.json();
            } catch {
                /* dropdown falls back to the default key only */
            }
            if (
                form.originKey !== CUSTOM_ORIGIN &&
                form.origins.length &&
                !form.origins.some((o) => o.key === form.originKey)
            )
                form.originKey = form.origins[0].key;
        },
    });
    return form;
}
