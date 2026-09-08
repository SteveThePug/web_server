// POST a JSON body to a Python API endpoint and feed each NDJSON line to onEvent.

export function describeError(status, payload) {
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

async function readNdjson(res, onEvent) {
    const decoder = new TextDecoder();
    let buffer = "";
    const handleChunk = (text, final) => {
        buffer += text;
        const lines = buffer.split("\n");
        buffer = final ? "" : lines.pop();
        for (const line of lines) {
            if (!line.trim()) continue;
            let ev;
            try {
                ev = JSON.parse(line);
            } catch {
                continue; // skip a malformed line rather than abort the whole search
            }
            onEvent(ev);
        }
    };
    if (!res.body?.getReader) {
        handleChunk(await res.text(), true);
        return;
    }
    const reader = res.body.getReader();
    for (;;) {
        const { value, done } = await reader.read();
        if (done) break;
        handleChunk(decoder.decode(value, { stream: true }), false);
    }
    handleChunk(decoder.decode(), true);
}

/** Throws an Error with a user-facing message on a non-2xx response. */
export async function runStream(url, body, { signal, onEvent }) {
    const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        signal,
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
    await readNdjson(res, onEvent);
}
