// Search lifecycle shared by the "Single stay" and "Cheapest night" panels.
//
// Both panels run the same routine: reset the result state, start a 1s elapsed
// ticker and an AbortController, stream NDJSON events from a Python endpoint,
// then tear everything down in a `finally` (and again on unmount, so leaving
// the page cancels an in-flight request). Only the events themselves differ,
// so each panel supplies its own applyEvent().

import { ref, computed, onBeforeUnmount } from "vue";
import { runStream } from "./hotelsStream.js";

/**
 * @param {object} options
 * @param {string} options.url            NDJSON streaming endpoint to POST to
 * @param {() => object} options.buildRequest   request body for one run
 * @param {(ev: object) => void} options.applyEvent  apply one streamed event.
 *        Called for every event; it is the panel's job to write into
 *        `summary` / `done` / `errorMsg` from the returned object. Declare it
 *        as a hoisted `function` so it can reference that object.
 * @param {() => void} [options.onStart]  reset panel-local state before a run
 * @param {string} [options.busyLabel]    status chip text while running
 * @param {string} [options.failMessage]  fallback message for a thrown error
 * @param {string} [options.earlyMessage] shown when the stream ends without `done`
 */
export function useHotelSearch({
    url,
    buildRequest,
    applyEvent,
    onStart,
    busyLabel = "Searching",
    failMessage = "Search failed.",
    earlyMessage = "The search ended early; showing what came back.",
}) {
    const searching = ref(false);
    const errorMsg = ref("");
    const summary = ref(null);
    const done = ref(false);
    const elapsed = ref(0);
    let ticker = null;
    let aborter = null;

    const statusText = computed(() => {
        if (errorMsg.value) return "Error";
        if (searching.value) return busyLabel;
        if (done.value) return "Done";
        return "Idle";
    });

    const hasResults = computed(() => summary.value !== null);

    async function run() {
        if (searching.value) return;
        errorMsg.value = "";
        summary.value = null;
        done.value = false;
        searching.value = true;
        elapsed.value = 0;
        if (onStart) onStart();
        ticker = setInterval(() => elapsed.value++, 1000);
        aborter = new AbortController();
        try {
            await runStream(url, buildRequest(), {
                signal: aborter.signal,
                onEvent: applyEvent,
            });
            if (!done.value && !errorMsg.value) errorMsg.value = earlyMessage;
        } catch (e) {
            // An abort is the user leaving or restarting, not a failure.
            if (e.name !== "AbortError") errorMsg.value = e.message || failMessage;
        } finally {
            clearInterval(ticker);
            ticker = null;
            aborter = null;
            searching.value = false;
        }
    }

    onBeforeUnmount(() => {
        if (ticker) clearInterval(ticker);
        if (aborter) aborter.abort();
    });

    return { searching, errorMsg, summary, done, elapsed, statusText, hasResults, run };
}
