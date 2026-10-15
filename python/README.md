# Python API service

A small FastAPI app (`app/main.py`) served by uvicorn on port 8000 inside the
`python` container. It also hosts the site's one Claude call, as an internal
route: the **rowing display reader** (see below). (The job tracker at `/jobs/`
is a separate app, `adamf/job-tracker`, with its own Claude calls.) Most of it
is one feature: **cheap Travelodge finder** —
given a London location and dates it pulls live room prices from Travelodge and
adds the cost of getting there from a chosen station on TfL, so you can rank
hotels by what the night *actually* costs, not just the room.

## Layout

| File | Role |
| --- | --- |
| `app/main.py` | HTTP skin: request models/validation, response cache, one-at-a-time guard, NDJSON framing, route handlers |
| `app/travelodge.py` | All the real work: upstream calls, caching, rate limiting, fare maths, the two search generators |
| `app/rowing.py` | The Claude call that reads time and distance off a photo of a rowing machine display |
| `app/claude.py` | What a Claude caller needs: the lazily built Anthropic client, the error types, and `ask_json` — the call itself |
| `app/__init__.py` | Empty; makes `app` a package so `app.main:app` resolves |
| `tests/` | pytest suite; needs `requirements-dev.txt` (see *Tests*) |

`travelodge.py`, `rowing.py` and `claude.py` never import FastAPI and can be
driven from a plain script.

## Endpoints

Paths below are as the service sees them; from a browser prefix everything with
`/py` (see *Behind nginx*).

| Method | Path | What it does |
| --- | --- | --- |
| GET | `/` | Index message |
| GET | `/health` | Liveness: `{"status": "ok", "time": ...}` |
| GET | `/hello/{name}` | Demo route |
| GET | `/hotels/origins` | The built-in origin stations the UI offers |
| POST | `/hotels/search` | Blocking search; returns once every shortlisted hotel is priced |
| POST | `/hotels/search/stream` | Same search as NDJSON progress lines |
| POST | `/hotels/scan/stream` | Cheapest night across a date range, NDJSON |
| POST | `/internal/rowing/read` | **Internal.** Read a rowing display photo with Claude |

Interactive schema: `/py/docs` (Swagger UI), `/py/openapi.json`.

### Streaming format

The stream endpoints return `application/x-ndjson`: one JSON object per line,
no SSE framing. Read a line, parse it, apply it. `exclude_none=True` is set, so
**absent keys mean null**.

`/hotels/search/stream` emits:

1. `hotels` — every shortlisted hotel with its room price and an estimated
   cycling time; transport columns are still null and `route` reads `pricing…`.
2. `transport` — one line per hotel as TfL answers; patch that row by `code`.
3. `done` — final summary plus all rows sorted by total.

`/hotels/scan/stream` emits `scan` (all candidate nights, pending), then a
`night` per date, then a `transport` per **unique hotel code** (apply it to
every night containing that code — transport is priced once and reused), then
`done`.

An `error` line ends a failed stream. Streams are always HTTP 200: the response
starts before the work does, so failures cannot become a status code. The
blocking `/hotels/search` is the exception — it returns 429 when a search is
already running and 502 on upstream failure.

Only **one** live search or scan runs process-wide at a time (the TfL budget
can't be shared), and identical requests are cached for 10 minutes.

### Internal routes

Routes under `/internal/` (today only the rowing reader) are the only ones in
this service that are **not public**. Each spends a paid Claude call and has no
auth of its own — the admin check happens in the Go backend, which calls the container
directly (`http://python:8000/...`). Two things keep them internal: nginx
answers `/py/internal/` with a 404, and the route itself 404s any request
carrying the `X-Real-IP` / `X-Forwarded-For` headers nginx adds to everything
it proxies (the Go backend sends neither).

Their plumbing is in two places, shared by any route added later:

- **`claude.ask_json`** (`app/claude.py`) is the whole Claude call: send one
  prompt that asks for a bare JSON object, take the first text block, strip a
  markdown fence, validate into a pydantic model. The caller supplies the model
  class, the messages, an optional system prompt and three fixed error strings
  (`failed`, `empty`, `unparseable`). It raises `NotConfigured` when
  `CLAUDE_API_KEY` is unset and `ReadError` with one of those three strings
  otherwise. The API's own error text and the model's reply go to the log,
  never into the message, because the Go backend relays the message verbatim.
- **`_internal_claude_route`** (`app/main.py`) is `app.post` plus the
  internal-only guard and the 502/503 entries in the docs. An exception handler
  beside it turns `ReadError` into a 502 and `NotConfigured` into a 503, both
  with a string `detail`. The guard runs as a dependency, so a request from
  outside is a 404 before its body is validated.

Every route uses one model, `MODEL` in `claude.py`: the `CLAUDE_MODEL` variable,
or `claude-sonnet-5-5` when it is blank. It is set in `.env` to one of
`claude-haiku-4-5`, `claude-sonnet-5-5` and `claude-opus-5-5`; recreate this
service after changing it. The request is
shaped per model by `_model_options`: a model that thinks gets
`output_config.effort: "low"` and 4,000 tokens of headroom on top of the
caller's `max_tokens`, because thinking is billed against it; Haiku 4.5 gets
neither, since it rejects the effort setting. A refusal (`stop_reason:
"refusal"`) is reported as the route's `failed` message. The Go callers set
their own timeout above this side's (30s per attempt, two SDK retries), so
they see this service's error rather than their own.

#### Adding another Claude-backed internal route

1. **A module** `app/<thing>.py`, shaped like `rowing.py`: the prompt, a
   pydantic model whose field names are the keys the prompt asks for (give
   every field a default — the model leaves keys out), and one function that
   returns `ask_json(Model, max_tokens=..., system=..., messages=[...],
   failed="...", empty="...", unparseable="...", get_client=_client)`. Import
   the client as `from .claude import client as _client`; that name is the
   seam tests replace. No FastAPI imports.
2. **A route** in `main.py`: a request model with length bounds, then

   ```python
   @_internal_claude_route("/internal/<thing>/<verb>", response_model=thing.Model, summary="...", description="...")
   def thing_verb(req: ThingRequest) -> thing.Model:
       return thing.do_it(req.field)
   ```

   A plain `def`, so the blocking call runs in the threadpool. No `try`, no
   guard call — the wrapper supplies both. Add the route to the endpoint list
   in `main.py`'s docstring and bump the FastAPI `version`.
3. **Tests** `tests/test_<thing>.py`, copied from `test_rowing.py`: fake the
   client with `monkeypatch.setattr(thing, "_client", ...)`.
4. **Docs**: the endpoint table and a section here. Nothing to do in nginx —
   it already 404s everything under `/py/internal/`.

If third-party text goes into the prompt (a web page, an email), say in the
system prompt that it is data and not instructions (the job tracker's
prompts, in its own repo, are examples) — and keep the output to fields that are stored or
shown, never acted on.

### Rowing display reader (internal)

`POST /internal/rowing/read` takes `{"media_type": "image/jpeg", "data":
"<base64>"}` and returns `{"timeMinutes", "timeSeconds", "distance"}` as read
by Claude. It is the image half of the Go backend's
admin-only `POST /api/rowing`: the Go handler does the admin check, the EXIF
date, the duplicate check, the sanity bounds and the insert, and calls this
only once it has decided the paid call is worth making. Nothing here judges
whether the numbers are plausible.

Errors: 502 with a fixed `detail` (`failed to process image`, `empty response
from image processor`, `failed to parse image data`) that the Go handler
relays to the admin; 503 if `CLAUDE_API_KEY` is unset.

## Behind nginx

`nginx/nginx.conf.template` proxies `/py/` to this container and **strips the
prefix**, so routes here are written relative to `/`. `--root-path /py` (set on
the uvicorn command line via `ROOT_PATH`) is what keeps the generated docs and
OpenAPI URLs correct behind that rewrite.

The hotel routes get their own nginx location with `proxy_buffering off` and a
300s read timeout, because otherwise nginx would hold the whole NDJSON stream
until it finished. The app also sends `X-Accel-Buffering: no` as a belt-and-
braces version of the same thing. The internal deadlines in `travelodge.py`
(200s for a search, 240s for a scan) exist to finish *before* that 300s cut, and
return partial results with `summary.truncated = true` instead of dying.

Env vars (from `docker-compose.yml`): `PYTHON_PORT`, `ROOT_PATH`, the optional
`TFL_APP_KEY` (see below), and `CLAUDE_API_KEY` and `CLAUDE_MODEL` for the internal Claude route. Dev mode mounts `python/app` and runs
uvicorn with `--reload`.

## Running locally

Inside the whole stack:

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
# http://localhost/py/docs
```

Or just this service:

```
cd python
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
# http://localhost:8000/docs   (no /py prefix when run directly)
```

Runtime dependencies are `fastapi`, `uvicorn[standard]`, `requests` and
`anthropic` — nothing else. `requirements.txt` is all the Docker image
installs.

## Tests

`requirements-dev.txt` is `requirements.txt` plus what the suite needs:
`pytest`, and `httpx`, which FastAPI's `TestClient` will not import without.

```
cd python
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements-dev.txt
python -m pytest tests -q
```

Nothing in the suite touches the network. The Claude tests replace the
Anthropic client with a fake (`monkeypatch.setattr(<module>, "_client", ...)`),
so they cover the request that is built, how the reply is parsed, the status
code each failure maps to and the internal-only guard — not what the model
actually answers. `test_claude.py` covers `ask_json` on its own.

## How the pricing works

1. Search Travelodge for the location and dates, following pagination, and drop
   hotels with no availability, no price, or no coordinates.
2. Keep the cheapest `top` (search) or `per_night` (scan) of them.
3. For each, ask TfL for the fastest route *and* a bus-only route at the given
   departure time. Pick the cheapest that isn't slower than `max_travel_min`.
4. Apply a Railcard discount (1/3 off, rounded to 5p) to off-peak rail/tube
   fares only — never buses. The return leg is assumed to be the same journey
   the next morning at off-peak rates.
5. Rank by room + round-trip transport. Cycling time is shown alongside.

### Fragility — read this before debugging "no prices"

Both upstreams are JSON APIs, not HTML, so there are no CSS selectors to break —
but the **key names play exactly that role** and neither API is a public,
documented contract for this use:

- Travelodge's `/api/v2/hotel` is the feed its own search page calls. The query
  parameters (`q`, `checkIn`, `rooms[0][adults]`, the undocumented
  `action=hotel_amend`, and `pagination=false`, which only hides the pagination
  metadata and does *not* return everything in one page) and the response keys
  read in `normalise_hotel` are whatever the site sends today.
- A browser `User-Agent` is sent because the edge throttles obviously scripted
  clients. If everything starts coming back empty or 403, bump it first.
- Every field read is tolerant — a renamed key yields `None`, not an exception.
  So a shape change looks like *"every hotel is unavailable"*, not a 502.
- TfL fares come from `fare.fares[].cost/peak/offPeak/chargeLevel`. Off-peak is
  detected by substring-matching `chargeLevel` for "off" *or* the cost equalling
  the off-peak price, because that field's spelling has moved around.
- Pagination stops when a batch of pages returns nothing new or comes back
  short, with `TL_MAX_START` as a hard backstop.

There is no scraping of pages behind a login and no attempt to book anything,
but this is still an undocumented internal API being polled from a public web
page — check Travelodge's terms before pointing it at anything heavier.

### TfL rate limit

Anonymous callers get 50 requests/min from a single IP; with a `TFL_APP_KEY` it
is 500/min. A search spends 2–3 requests per hotel, so the anonymous tier is the
binding constraint and is what the single-search-slot guard protects. Setting
`TFL_APP_KEY` also switches cycling times from a distance-based estimate to real
TfL cycle routing (`cycle_source` tells you which you're looking at).

## Frontend

Consumed by `vue/src/views/hotels/`:

- `SingleStayPanel.vue` → `POST /py/hotels/search/stream`
- `CheapestNightPanel.vue` → `POST /py/hotels/scan/stream`
- `useHotelForm.js` → `GET /py/hotels/origins`
- `hotelsStream.js` — the NDJSON reader (splits on newlines, skips a malformed
  line rather than aborting the search)

The Pydantic field names and JSON keys in `main.py` are that contract; renaming
one breaks the UI silently.
