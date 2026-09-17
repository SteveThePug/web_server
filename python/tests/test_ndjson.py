"""Tests for the NDJSON event contract of the streaming hotel endpoints.

The Vue frontend splits the response on newlines and JSON-parses each line, so
the contract under test is: one complete JSON object per line, the documented
event names, the right payload key per event kind, `exclude_none` dropping
absent keys, and the framing headers that stop nginx buffering the stream.

No scraping and no network happens here: `_run_events` takes the upstream
generator as a callable, so a fake generator stands in for travelodge and the
whole streaming skin can be exercised on its own.
"""

import asyncio
import json
from datetime import date, datetime

import pytest
from pydantic import BaseModel

from app import main, travelodge


@pytest.fixture(autouse=True)
def fresh_cache(monkeypatch):
    """Give every test its own response cache.

    The cache is module-global and keyed by request JSON, so without this a
    test that finishes a stream would serve the next test's identical request
    from cache.
    """
    monkeypatch.setattr(main, "_cache", travelodge.TTLCache(main.CACHE_TTL_S, max_entries=main.CACHE_MAX))


class FakeRequest(BaseModel):
    """Stand-in for a validated request: `_run_events` only reads model_dump."""

    marker: str = "x"


def search_events(make_gen, req=None):
    return list(
        main._run_events(
            req=req or FakeRequest(),
            prefix="search:",
            model=main.HotelSearchEvent,
            make_gen=make_gen,
            first_event="hotels",
            list_key="rows",
            row_keys={"transport": "row"},
        )
    )


def scan_events(make_gen, req=None):
    return list(
        main._run_events(
            req=req or FakeRequest(),
            prefix="scan:",
            model=main.HotelScanEvent,
            make_gen=make_gen,
            first_event="scan",
            list_key="nights",
            row_keys={"night": "night", "transport": "transport"},
        )
    )


def hotel_row(code="LON123", **over):
    """A minimal row in the shape travelodge emits (url and route are required)."""
    return dict({"code": code, "url": "https://example.test/h", "route": "pricing…"}, **over)


def search_summary(**over):
    return dict(
        {
            "location": "London",
            "checkin": date(2030, 1, 1),
            "checkout": date(2030, 1, 2),
            "nights": 1,
            "rooms": 1,
            "adults": 1,
            "children": 0,
            "railcard_holders": 1,
            "origin_name": "London Waterloo",
            "origin_lat": 51.5031,
            "origin_lon": -0.1132,
            "depart": datetime(2030, 1, 1, 18, 30),
            "max_travel_min": 60,
            "hotels_found": 1,
            "hotels_available": 1,
            "hotels_priced": 0,
        },
        **over,
    )


def read_lines(response):
    """Drain a StreamingResponse into the lines a client would see."""

    async def drain():
        return [chunk async for chunk in response.body_iterator]

    body = "".join(
        c.decode() if isinstance(c, (bytes, bytearray)) else c for c in asyncio.run(drain())
    )
    return body


# --------------------------------------------------------------------------- #
# Framing
# --------------------------------------------------------------------------- #
def test_ndjson_frames_one_json_object_per_line():
    events = [
        main.HotelSearchEvent(event="hotels", rows=[]),
        main.HotelSearchEvent(event="done", rows=[]),
    ]
    body = read_lines(main._ndjson(iter(events)))

    assert body.endswith("\n")  # every line is terminated, including the last
    lines = body.split("\n")[:-1]
    assert [json.loads(line)["event"] for line in lines] == ["hotels", "done"]
    # A line must never contain a raw newline or the client's split breaks.
    assert all("\n" not in line for line in lines)


def test_ndjson_omits_none_fields():
    # The reader treats an absent key as null, so exclude_none is part of the
    # contract rather than only a size optimisation.
    body = read_lines(main._ndjson(iter([main.HotelSearchEvent(event="error", detail="boom")])))
    payload = json.loads(body.strip())
    assert payload == {"event": "error", "cached": False, "detail": "boom"}


def test_ndjson_headers_defeat_proxy_buffering():
    response = main._ndjson(iter([]))
    assert response.media_type == "application/x-ndjson"
    # Without X-Accel-Buffering nginx holds the whole response and the stream
    # is pointless.
    assert response.headers["x-accel-buffering"] == "no"
    assert response.headers["cache-control"] == "no-cache"


def test_ndjson_of_no_events_is_an_empty_body():
    assert read_lines(main._ndjson(iter([]))) == ""


# --------------------------------------------------------------------------- #
# Search event shapes
# --------------------------------------------------------------------------- #
def test_search_events_carry_the_documented_shapes():
    summary = search_summary()
    rows = [hotel_row()]

    def gen():
        yield "hotels", dict(summary), rows
        yield "transport", hotel_row(route="tube")  # run_search emits a 2-tuple here
        yield "done", dict(summary, hotels_priced=1), rows

    events = search_events(gen)
    assert [e.event for e in events] == ["hotels", "transport", "done"]

    opening, transport, done = events
    assert opening.rows is not None and opening.row is None
    assert opening.summary.hotels_priced == 0
    # The 2-tuple is normalised to a None summary, which exclude_none then drops.
    assert transport.summary is None
    assert transport.row.route == "tube"
    assert transport.rows is None
    assert "summary" not in json.loads(transport.model_dump_json(exclude_none=True))
    assert done.rows is not None and done.summary.hotels_priced == 1
    assert all(e.cached is False for e in events)


def test_search_stream_serialises_dates_as_iso_strings():
    def gen():
        yield "hotels", search_summary(), []

    line = json.loads(search_events(gen)[0].model_dump_json(exclude_none=True))
    assert line["summary"]["checkin"] == "2030-01-01"
    assert line["summary"]["depart"].startswith("2030-01-01T18:30")


# --------------------------------------------------------------------------- #
# Scan event shapes
# --------------------------------------------------------------------------- #
def test_scan_events_carry_the_documented_shapes():
    summary = {
        "location": "London",
        "start": date(2030, 1, 1),
        "end": date(2030, 1, 8),
        "weekdays": [4],
        "nights": 1,
        "per_night": 3,
        "rooms": 1,
        "adults": 1,
        "children": 0,
        "railcard_holders": 1,
        "origin_name": "London Waterloo",
        "origin_lat": 51.5031,
        "origin_lon": -0.1132,
        "depart": datetime(2030, 1, 4, 18, 30),
        "max_travel_min": 60,
        "nights_candidate": 1,
    }
    night = {"checkin": date(2030, 1, 4), "checkout": date(2030, 1, 5), "status": "ok", "hotels": [hotel_row()]}

    def gen():
        yield "scan", dict(summary), [night]
        yield "night", dict(summary), night
        yield "transport", dict(summary), {"code": "LON123", "route": "bus"}
        yield "done", dict(summary), [night]

    events = scan_events(gen)
    assert [e.event for e in events] == ["scan", "night", "transport", "done"]

    opening, one_night, transport, done = events
    assert opening.nights is not None and opening.night is None
    assert one_night.night.status == "ok" and one_night.nights is None
    assert transport.transport.code == "LON123" and transport.nights is None
    assert done.nights is not None


# --------------------------------------------------------------------------- #
# Caching
# --------------------------------------------------------------------------- #
def test_a_cached_result_replays_as_the_opening_event_plus_done():
    summary = search_summary()
    rows = [hotel_row()]
    calls = []

    def gen():
        calls.append(1)
        yield "hotels", dict(summary), rows
        yield "done", dict(summary), rows

    req = FakeRequest(marker="cache-me")
    assert [e.event for e in search_events(gen, req)] == ["hotels", "done"]

    replay = search_events(gen, req)
    assert len(calls) == 1  # the upstream generator was not run again
    assert [e.event for e in replay] == ["hotels", "done"]
    assert all(e.cached for e in replay)
    # The replay carries the finished rows on both lines, so a streaming reader
    # needs no special case for a cache hit.
    assert replay[0].rows == replay[1].rows


def test_different_requests_do_not_share_a_cache_entry():
    def gen():
        yield "done", search_summary(), []

    search_events(gen, FakeRequest(marker="a"))
    assert search_events(gen, FakeRequest(marker="b"))[0].cached is False


def test_search_and_scan_caches_cannot_collide():
    # Both key on the same request JSON, so only the prefix separates them.
    def search_gen():
        yield "done", search_summary(), []

    def scan_gen():
        yield "done", None, []

    req = FakeRequest(marker="same")
    search_events(search_gen, req)
    assert scan_events(scan_gen, req)[0].cached is False


def test_an_unfinished_stream_is_not_cached():
    # Only "done" writes the cache, so a stream that errored part-way must be
    # re-run rather than replayed as if it had completed.
    def gen():
        yield "hotels", search_summary(), []
        raise travelodge.UpstreamError("Travelodge search failed")

    req = FakeRequest(marker="partial")
    assert [e.event for e in search_events(gen, req)] == ["hotels", "error"]
    assert [e.event for e in search_events(gen, req)] == ["hotels", "error"]


# --------------------------------------------------------------------------- #
# Errors and the single-search guard
# --------------------------------------------------------------------------- #
@pytest.mark.parametrize(
    "exc, want_detail",
    [
        (travelodge.UpstreamError("Travelodge search failed: 503"), "Travelodge search failed: 503"),
        (ValueError("no dates in the range match those weekdays"), "no dates in the range match those weekdays"),
        (KeyError("minPrice"), "Unexpected upstream response: 'minPrice'"),
        (TypeError("NoneType is not subscriptable"), "Unexpected upstream response: NoneType is not subscriptable"),
    ],
)
def test_upstream_failures_become_a_final_error_event(exc, want_detail):
    def gen():
        yield "hotels", search_summary(), []
        raise exc

    events = search_events(gen)
    assert [e.event for e in events] == ["hotels", "error"]
    assert events[-1].detail == want_detail


def test_a_failure_while_binding_the_generator_is_also_an_error_event():
    # make_gen is called inside the guard precisely so this is a stream event
    # rather than a 500 from the endpoint.
    def gen():
        raise ValueError("bad request")

    events = search_events(gen)
    assert [e.event for e in events] == ["error"]
    assert events[0].detail == "bad request"


def test_only_one_search_runs_at_a_time():
    main._search_slot.acquire()
    try:
        events = search_events(lambda: iter([]))
    finally:
        main._search_slot.release()
    assert [e.event for e in events] == ["error"]
    assert events[0].detail == main.ALREADY_RUNNING


def test_the_search_slot_is_released_after_a_failure():
    # A leaked semaphore would wedge the endpoint for the life of the process.
    def gen():
        raise travelodge.UpstreamError("boom")
        yield  # pragma: no cover - makes gen a generator

    search_events(gen, FakeRequest(marker="fail-1"))
    assert main._search_slot.acquire(blocking=False)
    main._search_slot.release()


def test_an_abandoned_stream_releases_the_search_slot():
    # The frontend closes the connection when the user navigates away, so the
    # generator is closed part-way; the finally in _guarded must still run.
    def gen():
        yield "hotels", search_summary(), []
        yield "done", search_summary(), []

    stream = main._run_events(
        req=FakeRequest(marker="abandon"),
        prefix="search:",
        model=main.HotelSearchEvent,
        make_gen=gen,
        first_event="hotels",
        list_key="rows",
        row_keys={"transport": "row"},
    )
    next(stream)
    stream.close()
    assert main._search_slot.acquire(blocking=False)
    main._search_slot.release()
