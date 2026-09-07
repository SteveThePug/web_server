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
from pydantic import BaseModel, Field, field_validator, model_validator

from . import travelodge

app = FastAPI(
    title="Python API",
    version="0.2.0",
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
CACHE_TTL_S = 10 * 60

_search_slot = threading.Semaphore(1)
_cache_lock = threading.Lock()
_cache: dict = {}  # key -> (expires_at, response dict)


class Origin(BaseModel):
    key: str
    name: str
    lat: float
    lon: float


class HotelSearchRequest(BaseModel):
    location: str = Field("Central London", min_length=1, max_length=80, description="Travelodge search text")
    checkin: date = Field(..., description="Check-in date (today or later)")
    checkout: Optional[date] = Field(None, description="Check-out date; overrides nights")
    nights: int = Field(1, ge=1, le=14, description="Used when checkout is not given")
    guests: str = Field("1", max_length=40, description='Rooms/guests e.g. "1", "2+1", "2,2+2"')
    railcard_holders: int = Field(1, ge=0, le=MAX_ADULTS, description="Adults with a Railcard (1/3 off off-peak rail)")
    origin: str = Field("clapham junction", max_length=60, description="Known station key or 'lat,lon'")
    depart: str = Field("19:30", description="Outbound departure time HH:MM")
    top: int = Field(15, ge=1, le=25, description="Price transport for the N cheapest stays only")
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
    truncated: bool = False


class HotelSearchResponse(BaseModel):
    summary: HotelSearchSummary
    rows: list[HotelRow]
    cached: bool = False


@app.get("/hotels/origins", response_model=list[Origin], summary="Known origin stations")
def hotel_origins() -> list[Origin]:
    return [Origin(key=k, name=n, lat=lat, lon=lon) for k, (n, lat, lon) in travelodge.ORIGINS.items()]


@app.post(
    "/hotels/search",
    response_model=HotelSearchResponse,
    summary="Rank Travelodges by room price plus TfL transport",
    description=(
        "Blocking search: pulls live Travelodge prices, then prices a round trip on TfL for the "
        "cheapest `top` hotels. Takes roughly a minute. Only one search runs at a time; identical "
        "searches are cached for ten minutes."
    ),
    responses={429: {"description": "Another search is already running"}, 502: {"description": "Upstream API failed"}},
)
def hotel_search(req: HotelSearchRequest) -> HotelSearchResponse:
    key = json.dumps(req.model_dump(mode="json"), sort_keys=True)
    now = time.monotonic()

    with _cache_lock:
        hit = _cache.get(key)
        if hit and hit[0] > now:
            return HotelSearchResponse(**hit[1], cached=True)

    if not _search_slot.acquire(blocking=False):
        raise HTTPException(status_code=429, detail="A search is already running, try again in a minute.")
    try:
        origin_name, origin = travelodge.parse_origin(req.origin)
        result = travelodge.run_search(
            location=req.location,
            checkin=req.checkin,
            checkout=req.checkout,
            rooms=travelodge.parse_guests(req.guests),
            railcard_holders=req.railcard_holders,
            origin_name=origin_name,
            origin=origin,
            depart=req.depart,
            top=req.top,
            max_miles=req.max_miles,
            max_travel_min=req.max_travel_min,
        )
    except travelodge.UpstreamError as e:
        raise HTTPException(status_code=502, detail=str(e)) from e
    except (KeyError, TypeError) as e:
        raise HTTPException(status_code=502, detail=f"Unexpected upstream response: {e}") from e
    except ValueError as e:
        raise HTTPException(status_code=422, detail=str(e)) from e
    finally:
        _search_slot.release()

    with _cache_lock:
        expired = [k for k, (exp, _) in _cache.items() if exp <= now]
        for k in expired:
            del _cache[k]
        _cache[key] = (time.monotonic() + CACHE_TTL_S, result)

    return HotelSearchResponse(**result)
