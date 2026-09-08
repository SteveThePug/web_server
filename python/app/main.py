"""Small Python API service.

Nginx proxies https://www.<DOMAIN>/py/... to this container and strips the
/py prefix, so routes here are defined relative to "/". FastAPI's root_path
(set via --root-path on the uvicorn command line) keeps the generated docs
at /py/docs working behind the proxy.
"""

import json
import re
import threading
import time
from datetime import date, datetime, timedelta, timezone
from typing import Optional

from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field, field_validator, model_validator

from . import travelodge

app = FastAPI(
    title="Python API",
    version="0.3.0",
    docs_url="/docs",
    openapi_url="/openapi.json",
)


class Health(BaseModel):
    status: str
    time: datetime


class Greeting(BaseModel):
    message: str


@app.get("/", response_model=Greeting, summary="Service index")
def index() -> Greeting:
    return Greeting(message="Python API is running. See /py/docs for the schema.")


@app.get("/health", response_model=Health, summary="Liveness check")
def health() -> Health:
    return Health(status="ok", time=datetime.now(timezone.utc))


@app.get("/hello/{name}", response_model=Greeting, summary="Greet a caller")
def hello(name: str) -> Greeting:
    return Greeting(message=f"Hello, {name}!")


# --------------------------------------------------------------------------- #
# Cheap Travelodge finder
# --------------------------------------------------------------------------- #
MAX_ROOMS = 4
MAX_ADULTS = 8
MAX_CHILDREN = 8
MAX_SCAN_DAYS = travelodge.SCAN_MAX_DAYS
CACHE_TTL_S = 10 * 60

_search_slot = threading.Semaphore(1)
_cache_lock = threading.Lock()
_cache: dict = {}  # key -> (expires_at, response dict)


class Origin(BaseModel):
    key: str
    name: str
    lat: float
    lon: float


class HotelRequestBase(BaseModel):
    """Fields shared by the single-stay search and the date-range scan."""

    location: str = Field("Central London", min_length=1, max_length=80, description="Travelodge search text")
    guests: str = Field("1", max_length=40, description='Rooms/guests e.g. "1", "2+1", "2,2+2"')
    railcard_holders: int = Field(1, ge=0, le=MAX_ADULTS, description="Adults with a Railcard (1/3 off off-peak rail)")
    origin: str = Field("clapham junction", max_length=60, description="Known station key or 'lat,lon'")
    depart: str = Field("19:30", description="Outbound departure time HH:MM")
    max_miles: Optional[float] = Field(None, gt=0, le=50, description="Drop hotels further than this from the search")
    max_travel_min: int = Field(75, ge=10, le=240, description="Ignore slower routes than this when picking cheapest")

    @field_validator("location")
    @classmethod
    def _strip_location(cls, v: str) -> str:
        v = v.strip()
        if not v:
            raise ValueError("location cannot be empty")
        return v

    @field_validator("depart")
    @classmethod
    def _check_depart(cls, v: str) -> str:
        if not re.fullmatch(r"([01]\d|2[0-3]):[0-5]\d", v.strip()):
            raise ValueError("depart must be HH:MM")
        return v.strip()

    @field_validator("guests")
    @classmethod
    def _check_guests(cls, v: str) -> str:
        try:
            rooms = travelodge.parse_guests(v)
        except ValueError as e:
            raise ValueError(f"invalid guests spec: {e}") from None
        if len(rooms) > MAX_ROOMS:
            raise ValueError(f"at most {MAX_ROOMS} rooms")
        if sum(a for a, _ in rooms) > MAX_ADULTS:
            raise ValueError(f"at most {MAX_ADULTS} adults")
        if sum(c for _, c in rooms) > MAX_CHILDREN:
            raise ValueError(f"at most {MAX_CHILDREN} children")
        return v.strip()

    @field_validator("origin")
    @classmethod
    def _check_origin(cls, v: str) -> str:
        travelodge.parse_origin(v)  # raises ValueError with a useful message
        return v.strip()

    def shared_kwargs(self) -> dict:
        origin_name, origin = travelodge.parse_origin(self.origin)
        return {
            "location": self.location,
            "rooms": travelodge.parse_guests(self.guests),
            "railcard_holders": self.railcard_holders,
            "origin_name": origin_name,
            "origin": origin,
            "depart": self.depart,
            "max_miles": self.max_miles,
            "max_travel_min": self.max_travel_min,
        }


class HotelSearchRequest(HotelRequestBase):
    checkin: date = Field(..., description="Check-in date (today or later)")
    checkout: Optional[date] = Field(None, description="Check-out date; overrides nights")
    nights: int = Field(1, ge=1, le=14, description="Used when checkout is not given")
    top: int = Field(20, ge=1, le=40, description="Price transport for the N cheapest stays only")

    @model_validator(mode="after")
    def _check_dates(self):
        if self.checkin < date.today():
            raise ValueError("checkin must be today or later")
        checkout = self.checkout or self.checkin + timedelta(days=self.nights)
        if checkout <= self.checkin:
            raise ValueError("checkout must be after checkin")
        if (checkout - self.checkin).days > 14:
            raise ValueError("stays are limited to 14 nights")
        self.checkout = checkout
        self.nights = (checkout - self.checkin).days
        return self


class HotelScanRequest(HotelRequestBase):
    start: date = Field(..., description="First candidate check-in date (today or later)")
    end: date = Field(..., description=f"Last candidate check-in date, at most {MAX_SCAN_DAYS} days after start")
    weekdays: list[int] = Field(
        default_factory=lambda: list(range(7)),
        description="Check-in weekdays to consider, 0=Mon .. 6=Sun (default all)",
    )
    nights: int = Field(1, ge=1, le=7, description="Length of each candidate stay")
    per_night: int = Field(5, ge=1, le=10, description="Keep the N cheapest hotels for each night")

    @field_validator("weekdays")
    @classmethod
    def _check_weekdays(cls, v: list[int]) -> list[int]:
        days = []
        for d in v:
            if d < 0 or d > 6:
                raise ValueError("weekdays must be 0 (Mon) to 6 (Sun)")
            if d not in days:
                days.append(d)
        if not days:
            raise ValueError("pick at least one weekday")
        return days

    @model_validator(mode="after")
    def _check_dates(self):
        if self.start < date.today():
            raise ValueError("start must be today or later")
        if self.end < self.start:
            raise ValueError("end must be on or after start")
        if (self.end - self.start).days > MAX_SCAN_DAYS:
            raise ValueError(f"scans are limited to {MAX_SCAN_DAYS} days")
        if not travelodge.candidate_checkins(self.start, self.end, set(self.weekdays), self.nights):
            raise ValueError("no dates in the range match those weekdays")
        return self


class HotelRow(BaseModel):
    code: Optional[str] = None
    name: Optional[str] = None
    lat: Optional[float] = None
    lon: Optional[float] = None
    room_price: Optional[float] = None
    room_code: Optional[str] = None
    available: bool = True
    low_availability: bool = False
    distance_from_search_mi: Optional[float] = None
    rating: Optional[float] = None
    url: str
    transport_out: Optional[float] = None
    transport_return: Optional[float] = None
    transport_total: Optional[float] = None
    travel_minutes: Optional[int] = None
    route: str
    fastest_minutes: Optional[int] = None
    fastest_total: Optional[float] = None
    fastest_route: Optional[str] = None
    total: Optional[float] = None
    cycle_minutes: Optional[int] = None
    cycle_km: Optional[float] = None
    cycle_source: str = "estimate"


class HotelSearchSummary(BaseModel):
    location: str
    checkin: date
    checkout: date
    nights: int
    rooms: int
    adults: int
    children: int
    railcard_holders: int
    origin_name: str
    origin_lat: float
    origin_lon: float
    depart: datetime
    max_travel_min: int
    hotels_found: int
    hotels_available: int
    hotels_priced: int
    hotels_shortlisted: int = 0
    tfl_rate_per_min: int = 0
    truncated: bool = False


class HotelSearchResponse(BaseModel):
    summary: HotelSearchSummary
    rows: list[HotelRow]
    cached: bool = False


@app.get("/hotels/origins", response_model=list[Origin], summary="Known origin stations")
def hotel_origins() -> list[Origin]:
    return [Origin(key=k, name=n, lat=lat, lon=lon) for k, (n, lat, lon) in travelodge.ORIGINS.items()]


class HotelSearchEvent(BaseModel):
    """One line of the NDJSON stream from /hotels/search/stream."""

    event: str = Field(..., description='"hotels", "transport", "done" or "error"')
    summary: Optional[HotelSearchSummary] = None
    rows: Optional[list[HotelRow]] = None
    row: Optional[HotelRow] = None
    cached: bool = False
    detail: Optional[str] = None


class TransportFields(BaseModel):
    """Transport and cycling columns for one hotel, independent of the night."""

    code: str
    transport_out: Optional[float] = None
    transport_return: Optional[float] = None
    transport_total: Optional[float] = None
    travel_minutes: Optional[int] = None
    route: str
    fastest_minutes: Optional[int] = None
    fastest_total: Optional[float] = None
    fastest_route: Optional[str] = None
    cycle_minutes: Optional[int] = None
    cycle_km: Optional[float] = None
    cycle_source: str = "estimate"


class NightRow(BaseModel):
    checkin: date
    checkout: date
    status: str = Field("pending", description='"pending", "ok", "failed" or "skipped"')
    detail: Optional[str] = Field(None, description="Why a night failed or was skipped")
    hotels_available: int = 0
    hotels: list[HotelRow] = Field(default_factory=list, description="Cheapest per_night hotels for this night")
    best_total: Optional[float] = None
    best_room: Optional[float] = None
    best_code: Optional[str] = None


class HotelScanSummary(BaseModel):
    location: str
    start: date
    end: date
    weekdays: list[int]
    nights: int
    per_night: int
    rooms: int
    adults: int
    children: int
    railcard_holders: int
    origin_name: str
    origin_lat: float
    origin_lon: float
    depart: datetime
    max_travel_min: int
    nights_candidate: int
    nights_fetched: int = 0
    nights_failed: int = 0
    nights_skipped: int = 0
    hotels_unique: int = 0
    hotels_priced: int = 0
    tfl_rate_per_min: int = 0
    truncated: bool = False


class HotelScanEvent(BaseModel):
    """One line of the NDJSON stream from /hotels/scan/stream."""

    event: str = Field(..., description='"scan", "night", "transport", "done" or "error"')
    summary: Optional[HotelScanSummary] = None
    nights: Optional[list[NightRow]] = None
    night: Optional[NightRow] = None
    transport: Optional[TransportFields] = None
    cached: bool = False
    detail: Optional[str] = None


def _cache_get(key):
    now = time.monotonic()
    with _cache_lock:
        hit = _cache.get(key)
        if hit and hit[0] > now:
            return hit[1]
    return None


def _cache_put(key, result):
    now = time.monotonic()
    with _cache_lock:
        expired = [k for k, (exp, _) in _cache.items() if exp <= now]
        for k in expired:
            del _cache[k]
        _cache[key] = (now + CACHE_TTL_S, result)


ALREADY_RUNNING = "A search is already running, try again in a minute."


def _guarded(run, make_error):
    """Run a live search generator while holding the single search slot.

    Upstream failures are reported as a final "error" event rather than raised,
    so a stream that has already started can still tell the client what went
    wrong.
    """
    if not _search_slot.acquire(blocking=False):
        yield make_error(ALREADY_RUNNING)
        return
    try:
        yield from run()
    except travelodge.UpstreamError as e:
        yield make_error(str(e))
    except (KeyError, TypeError) as e:
        yield make_error(f"Unexpected upstream response: {e}")
    except ValueError as e:
        yield make_error(str(e))
    finally:
        _search_slot.release()


def _search_events(req: HotelSearchRequest):
    """Yield HotelSearchEvent objects for a request, serving from cache when possible."""
    key = "search:" + json.dumps(req.model_dump(mode="json"), sort_keys=True)
    hit = _cache_get(key)
    if hit:
        yield HotelSearchEvent(event="hotels", cached=True, **hit)
        yield HotelSearchEvent(event="done", cached=True, **hit)
        return

    def run():
        gen = travelodge.run_search(
            checkin=req.checkin,
            checkout=req.checkout,
            top=req.top,
            **req.shared_kwargs(),
        )
        for ev in gen:
            if ev[0] == "transport":
                yield HotelSearchEvent(event="transport", row=ev[1])
            else:
                kind, summary, rows = ev
                if kind == "done":
                    _cache_put(key, {"summary": summary, "rows": rows})
                yield HotelSearchEvent(event=kind, summary=summary, rows=rows)

    yield from _guarded(run, lambda detail: HotelSearchEvent(event="error", detail=detail))


def _scan_events(req: HotelScanRequest):
    """Yield HotelScanEvent objects for a scan, serving from cache when possible."""
    key = "scan:" + json.dumps(req.model_dump(mode="json"), sort_keys=True)
    hit = _cache_get(key)
    if hit:
        yield HotelScanEvent(event="scan", cached=True, **hit)
        yield HotelScanEvent(event="done", cached=True, **hit)
        return

    def run():
        gen = travelodge.run_scan(
            start=req.start,
            end=req.end,
            weekdays=set(req.weekdays),
            nights=req.nights,
            per_night=req.per_night,
            **req.shared_kwargs(),
        )
        for kind, summary, payload in gen:
            if kind == "night":
                yield HotelScanEvent(event=kind, summary=summary, night=payload)
            elif kind == "transport":
                yield HotelScanEvent(event=kind, summary=summary, transport=payload)
            else:
                if kind == "done":
                    _cache_put(key, {"summary": summary, "nights": payload})
                yield HotelScanEvent(event=kind, summary=summary, nights=payload)

    yield from _guarded(run, lambda detail: HotelScanEvent(event="error", detail=detail))


_SEARCH_DESCRIPTION = (
    "Pulls live Travelodge prices, then prices a round trip on TfL for the cheapest `top` hotels. "
    "TfL allows 50 lookups a minute without an app key (two per hotel), so a full run takes up to "
    "about a minute. Only one live search or scan runs at a time; identical requests are cached for ten minutes."
)

_SCAN_DESCRIPTION = (
    "Searches Travelodge once per candidate check-in date in `start`..`end` (filtered by `weekdays`), keeps the "
    "`per_night` cheapest hotels for each, then prices TfL transport once per unique hotel at the first "
    "night's departure time and reuses it for every night. Up to about three months; a scan can take a few "
    "minutes and shares the single live-search slot with /hotels/search. Identical scans are cached for ten minutes."
)


def _ndjson(events):
    return StreamingResponse(
        (ev.model_dump_json(exclude_none=True) + "\n" for ev in events),
        media_type="application/x-ndjson",
        headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
    )


@app.post(
    "/hotels/search",
    response_model=HotelSearchResponse,
    summary="Rank Travelodges by room price plus TfL transport",
    description="Blocking form of the search: returns once every hotel is priced. " + _SEARCH_DESCRIPTION,
    responses={429: {"description": "Another search is already running"}, 502: {"description": "Upstream API failed"}},
)
def hotel_search(req: HotelSearchRequest) -> HotelSearchResponse:
    final = None
    for ev in _search_events(req):
        if ev.event == "error":
            status = 429 if ev.detail == ALREADY_RUNNING else 502
            raise HTTPException(status_code=status, detail=ev.detail)
        if ev.event == "done":
            final = ev
    if final is None:
        raise HTTPException(status_code=502, detail="Search ended without a result")
    return HotelSearchResponse(summary=final.summary, rows=final.rows, cached=final.cached)


@app.post(
    "/hotels/search/stream",
    summary="Streaming form of /hotels/search",
    description=(
        "Newline-delimited JSON (`application/x-ndjson`). Each line is a `HotelSearchEvent`: first "
        '`hotels` (every shortlisted hotel with room price and cycling time), then one `transport` '
        "line per hotel as TfL fares arrive, then `done` (rows sorted by total). An `error` line "
        "ends a failed stream. " + _SEARCH_DESCRIPTION
    ),
    response_class=StreamingResponse,
    responses={200: {"content": {"application/x-ndjson": {"schema": HotelSearchEvent.model_json_schema()}}}},
)
def hotel_search_stream(req: HotelSearchRequest):
    return _ndjson(_search_events(req))


@app.post(
    "/hotels/scan/stream",
    summary="Find the cheapest night to stay across a date range",
    description=(
        "Newline-delimited JSON (`application/x-ndjson`). Each line is a `HotelScanEvent`: first `scan` "
        "(every candidate night, pending), then one `night` line per date as Travelodge answers (its cheapest "
        "hotels, transport unpriced), then one `transport` line per unique hotel as TfL fares arrive (apply it "
        "to every night containing that hotel code), then `done` (totals filled in, nights sorted by best "
        "total). An `error` line ends a failed stream. " + _SCAN_DESCRIPTION
    ),
    response_class=StreamingResponse,
    responses={200: {"content": {"application/x-ndjson": {"schema": HotelScanEvent.model_json_schema()}}}},
)
def hotel_scan_stream(req: HotelScanRequest):
    return _ndjson(_scan_events(req))
