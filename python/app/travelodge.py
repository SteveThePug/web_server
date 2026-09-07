"""Cheapest Travelodge (room + TfL transport) for a stay in London.

Ported from ~/cheap_hotel/cheap_travelodge.py. The pricing logic is unchanged;
the plumbing around it has been rewritten for speed:

  * Travelodge result pages are fetched in parallel batches.
  * TfL journeys are fetched by a worker pool, throttled to the anonymous API
    limit (50 requests/min; 500/min with TFL_APP_KEY) and cached for a while
    so repeat searches don't spend the budget again.
  * `run_search` is a generator that yields the hotel list first and then one
    event per priced hotel, so callers can stream progress to a browser.

Data sources
  * Travelodge's own search JSON API -> live room prices for every hotel
    matching a location search.
  * TfL Journey Planner API -> route + pay-as-you-go fare from the origin to
    each hotel.

Method
  1. Pull every hotel for the search, drop ones with no availability, sort by
     cheapest room price and keep only the cheapest `top` hotels.
  2. For each, ask TfL for (a) the fastest route and (b) a bus-only route at
     the given departure time. Pick the cheapest that isn't slower than
     `max_travel_min`.
  3. Apply a Railcard discount (1/3 off) to OFF-PEAK rail/tube fares. Buses
     are never discounted.
  4. Assume the return journey the next morning costs the off-peak equivalent
     of the outbound.
  5. Rank by room + round-trip transport. Cycling time is reported separately
     (TfL cycle routing with an app key, otherwise a distance-based estimate).
"""

import math
import os
import threading
import time
from collections import deque
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import date, datetime

import requests

TL_API = "https://www.travelodge.co.uk/api/v2/hotel"
TL_SITE = "https://www.travelodge.co.uk"
TFL_API = "https://api.tfl.gov.uk/Journey/JourneyResults/{frm}/to/{to}"
HEADERS = {
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Safari/537.36",
    "Accept": "application/json",
}

# Well-known origins: key -> (display name, lat, lon)
ORIGINS = {
    "clapham junction": ("Clapham Junction", 51.4645, -0.1705),
    "waterloo": ("London Waterloo", 51.5031, -0.1132),
    "victoria": ("London Victoria", 51.4952, -0.1441),
    "kings cross": ("King's Cross St Pancras", 51.5308, -0.1238),
    "liverpool street": ("Liverpool Street", 51.5178, -0.0817),
    "paddington": ("London Paddington", 51.5154, -0.1755),
    "london bridge": ("London Bridge", 51.5050, -0.0864),
    "euston": ("London Euston", 51.5282, -0.1337),
}

# Rough Greater London bounding box for custom "lat,lon" origins.
LONDON_BOX = (51.25, 51.75, -0.55, 0.35)  # min lat, max lat, min lon, max lon

# Wall-clock budget for one run_search call. Nginx cuts the connection at
# 300s, so stop pricing transport well before that and return what we have.
SEARCH_DEADLINE_S = 200

# Travelodge returns 10 hotels per page whatever you ask for; fetch pages in
# parallel batches until a batch comes back with nothing new.
TL_PAGE_SIZE = 10
TL_PAGE_BATCH = 6
TL_MAX_START = 500

# TfL: anonymous callers get 50 requests/min, keyed callers 500/min. A call
# normally takes 1-5s; the odd one hangs, so time out early and retry.
TFL_WORKERS = 12
TFL_HOTEL_WORKERS = 4
TFL_TIMEOUT_S = 15
TFL_RATE_ANON = 48
TFL_RATE_KEYED = 450
TFL_CACHE_TTL_S = 2 * 60 * 60

# Cycling estimate when TfL cycle routing isn't used: straight-line distance
# times a road-winding factor, at a relaxed London pace.
CYCLE_ROUTE_FACTOR = 1.3
CYCLE_KMH = 14.0

_session = requests.Session()
_session.mount("https://", requests.adapters.HTTPAdapter(pool_connections=4, pool_maxsize=TFL_WORKERS + TL_PAGE_BATCH))


class UpstreamError(Exception):
    """Travelodge or TfL could not be reached or returned garbage."""


# --------------------------------------------------------------------------- #
# Input parsing
# --------------------------------------------------------------------------- #
def parse_guests(spec):
    """'2,2+1' -> [(2, 0), (2, 1)]  (list of (adults, children) per room)."""
    rooms = []
    for part in spec.split(","):
        part = part.strip()
        if not part:
            continue
        a, _, c = part.partition("+")
        rooms.append((int(a), int(c or 0)))
    if not rooms:
        raise ValueError("guests spec is empty")
    for adults, children in rooms:
        if adults < 1:
            raise ValueError("each room needs at least one adult")
        if children < 0:
            raise ValueError("children cannot be negative")
    return rooms


def parse_origin(spec):
    """Return (name, (lat, lon)) for an ORIGINS key or a 'lat,lon' string."""
    key = spec.strip().lower()
    if key in ORIGINS:
        name, lat, lon = ORIGINS[key]
        return name, (lat, lon)
    if "," in spec:
        try:
            lat, lon = (float(x) for x in spec.split(",", 1))
        except ValueError:
            raise ValueError("origin must be a known station or 'lat,lon'") from None
        lo_lat, hi_lat, lo_lon, hi_lon = LONDON_BOX
        if not (lo_lat <= lat <= hi_lat and lo_lon <= lon <= hi_lon):
            raise ValueError("origin coordinates must be inside Greater London")
        return f"{lat:.4f},{lon:.4f}", (lat, lon)
    raise ValueError("unknown origin '%s' (known: %s, or 'lat,lon')" % (spec, ", ".join(ORIGINS)))


# --------------------------------------------------------------------------- #
# Travelodge
# --------------------------------------------------------------------------- #
def _fetch_page(params, start):
    p = dict(params, start=start)
    try:
        r = _session.get(TL_API, params=p, headers=HEADERS, timeout=30)
        r.raise_for_status()
        return r.json().get("results") or []
    except (requests.RequestException, ValueError, AttributeError) as e:
        raise UpstreamError(f"Travelodge search failed: {e}") from e


def fetch_hotels(location, checkin, checkout, rooms):
    """Return list of hotel dicts from the Travelodge search API (all pages).

    rooms: list of (adults, children) tuples, one per room.
    """
    params = {
        "pagination": "false",
        "checkIn": checkin.isoformat(),
        "checkOut": checkout.isoformat(),
        "q": location,
        "action": "hotel_amend",
    }
    for i, (adults, children) in enumerate(rooms):
        params[f"rooms[{i}][adults]"] = adults
        params[f"rooms[{i}][children]"] = children

    hotels, seen = [], set()

    def absorb(results):
        new = [h for h in results if h.get("code") and h.get("code") not in seen]
        for h in new:
            seen.add(h["code"])
        hotels.extend(new)
        return bool(new)

    # First page alone: tells us whether the search matched anything at all.
    first = _fetch_page(params, 0)
    if not absorb(first) or len(first) < TL_PAGE_SIZE:
        return hotels

    start = TL_PAGE_SIZE
    with ThreadPoolExecutor(max_workers=TL_PAGE_BATCH) as ex:
        while start <= TL_MAX_START:
            starts = [start + i * TL_PAGE_SIZE for i in range(TL_PAGE_BATCH)]
            pages = list(ex.map(lambda s: _fetch_page(params, s), starts))
            got_new = False
            exhausted = False
            for page in pages:  # keep API order so ranking stays stable
                if absorb(page):
                    got_new = True
                if len(page) < TL_PAGE_SIZE:
                    exhausted = True
            if not got_new or exhausted:
                break
            start += TL_PAGE_SIZE * TL_PAGE_BATCH
    return hotels


def normalise_hotel(h):
    return {
        "code": h.get("code"),
        "name": h.get("title"),
        "lat": h.get("lat"),
        "lon": h.get("lon"),
        "room_price": h.get("minPrice"),
        "room_code": h.get("minRoomCode"),
        "available": bool(h.get("hasAvailability")),
        "low_availability": bool(h.get("hasLowAvailability")),
        "distance_from_search_mi": h.get("distance"),
        "rating": (h.get("rating") or {}).get("averageRating"),
        "url": TL_SITE + (h.get("hotelUrl") or ""),
    }


# --------------------------------------------------------------------------- #
# Geometry
# --------------------------------------------------------------------------- #
def haversine_km(a, b):
    lat1, lon1 = map(math.radians, a)
    lat2, lon2 = map(math.radians, b)
    h = math.sin((lat2 - lat1) / 2) ** 2 + math.cos(lat1) * math.cos(lat2) * math.sin((lon2 - lon1) / 2) ** 2
    return 2 * 6371.0 * math.asin(math.sqrt(h))


def estimate_cycle(origin, dest):
    """(minutes, km) for cycling from origin to dest, from straight-line distance."""
    km = haversine_km(origin, dest) * CYCLE_ROUTE_FACTOR
    return max(1, round(km / CYCLE_KMH * 60)), round(km, 1)


# --------------------------------------------------------------------------- #
# TfL
# --------------------------------------------------------------------------- #
def round_5p(pence):
    return int(5 * round(pence / 5.0))


def price_journey(journey, adults=1, railcard_holders=1):
    """
    Return (outbound_pence, return_pence, detail) for one TfL journey for the
    whole party, or None if the fare can't be determined.

    adults: number of fare-paying adults.  railcard_holders: how many of them
    get 1/3 off off-peak rail fares (0 = nobody has a railcard).
    """
    railcard_holders = max(0, min(adults, railcard_holders))
    full_fare = adults - railcard_holders
    legs = journey.get("legs") or []
    modes = [(l.get("mode") or {}).get("id") for l in legs]
    modes = [m for m in modes if m]  # skip legs without a mode id
    non_walk = [m for m in modes if m != "walking"]
    fare = journey.get("fare")

    if not non_walk:  # walking only
        return 0, 0, "walk"

    if not fare or not fare.get("fares"):
        return None

    out = ret = 0
    for f in fare["fares"]:
        cost = f.get("cost") or 0
        peak, offpeak = f.get("peak") or 0, f.get("offPeak") or 0
        is_bus = (peak == 0 and offpeak == 0) or any(
            (t.get("tapDetails") or {}).get("modeType") == "Bus" for t in (f.get("taps") or [])
        )
        if is_bus:
            # Bus: flat fare, hopper legs already come through as cost 0. No railcard.
            out += cost * adults
            ret += cost * adults
        else:
            level = (f.get("chargeLevel") or "").lower()
            out_full = cost
            ret_full = offpeak or cost  # next-day return assumed off-peak
            out_disc = round_5p(out_full * 2 / 3) if ("off" in level or cost == offpeak) else out_full
            ret_disc = round_5p(ret_full * 2 / 3)
            out += out_full * full_fare + out_disc * railcard_holders
            ret += ret_full * full_fare + ret_disc * railcard_holders

    # short human summary of the route
    parts = []
    for l in legs:
        m = (l.get("mode") or {}).get("id")
        if not m or m == "walking":
            continue
        ri = l.get("routeOptions") or [{}]
        name = ri[0].get("name") or ""
        parts.append(f"{m.replace('national-rail','rail')}{(' ' + name) if name else ''}")
    return out, ret, " > ".join(parts) if parts else "walk"


class RateLimiter:
    """Sliding-window limiter: at most `per_minute` acquisitions in any 60s."""

    def __init__(self, per_minute):
        self.per_minute = per_minute
        self._times = deque()
        self._lock = threading.Lock()

    def acquire(self, deadline=None):
        """Block until a slot is free. Returns False if `deadline` would pass first."""
        while True:
            with self._lock:
                now = time.monotonic()
                while self._times and self._times[0] <= now - 60:
                    self._times.popleft()
                if len(self._times) < self.per_minute:
                    self._times.append(now)
                    return True
                wait = self._times[0] + 60 - now
            if deadline is not None and now + wait > deadline:
                return False
            time.sleep(min(wait, 1.0))


_limiters = {}
_limiters_lock = threading.Lock()


def _limiter(app_key):
    rate = TFL_RATE_KEYED if app_key else TFL_RATE_ANON
    with _limiters_lock:
        lim = _limiters.get(rate)
        if lim is None:
            lim = _limiters[rate] = RateLimiter(rate)
        return lim


_tfl_cache = {}  # key -> (expires_at, journeys)
_tfl_cache_lock = threading.Lock()


def sleep_within(seconds, deadline):
    """Sleep for `seconds` unless that would overrun `deadline` (a time.monotonic()
    value, or None for no limit). Returns False if the sleep was skipped."""
    if deadline is not None and time.monotonic() + seconds > deadline:
        return False
    time.sleep(seconds)
    return True


def tfl_journeys(frm, to, when, extra=None, app_key=None, retries=3, deadline=None):
    """Journeys from TfL for one origin/destination/time/mode, cached and rate limited."""
    params = {
        "date": when.strftime("%Y%m%d"),
        "time": when.strftime("%H%M"),
        "timeIs": "Departing",
        "journeyPreference": "LeastTime",
    }
    if extra:
        params.update(extra)
    key = (round(frm[0], 5), round(frm[1], 5), round(to[0], 5), round(to[1], 5), params["date"], params["time"],
           tuple(sorted((extra or {}).items())))
    now = time.monotonic()
    with _tfl_cache_lock:
        hit = _tfl_cache.get(key)
        if hit and hit[0] > now:
            return hit[1]

    if app_key:
        params["app_key"] = app_key
    url = TFL_API.format(frm=f"{frm[0]},{frm[1]}", to=f"{to[0]},{to[1]}")
    limiter = _limiter(app_key)
    journeys = None
    for attempt in range(retries):
        remaining = TFL_TIMEOUT_S if deadline is None else deadline - time.monotonic()
        if remaining < 1 or not limiter.acquire(deadline):
            break
        try:
            r = _session.get(url, params=params, headers=HEADERS, timeout=min(TFL_TIMEOUT_S, max(1, remaining)))
        except requests.RequestException:
            if not sleep_within(2, deadline):
                break
            continue
        if r.status_code == 429:
            if not sleep_within(5 * (attempt + 1), deadline):
                break
            continue
        if r.status_code >= 500:
            if not sleep_within(2, deadline):
                break
            continue
        if r.status_code != 200:
            journeys = []  # e.g. 300 "disambiguation": no usable route
            break
        try:
            journeys = r.json().get("journeys") or []
        except ValueError:
            journeys = []
        break

    if journeys is None:
        return []  # gave up: don't cache a failure
    with _tfl_cache_lock:
        if len(_tfl_cache) > 5000:
            expired = [k for k, (exp, _) in _tfl_cache.items() if exp <= now]
            for k in expired:
                del _tfl_cache[k]
        _tfl_cache[key] = (time.monotonic() + TFL_CACHE_TTL_S, journeys)
    return journeys


TRANSPORT_QUERIES = (("fast", None), ("bus", {"mode": "bus,walking"}))
CYCLE_QUERY = {"mode": "cycle"}


def best_transport(origin, hotel, when, adults, railcard_holders, max_minutes, app_key=None, deadline=None, pool=None):
    """Cheapest acceptable round-trip transport to a hotel. Returns dict or None.

    Runs the fast and bus-only TfL queries concurrently on `pool` when given.
    """
    dest = (hotel["lat"], hotel["lon"])

    def query(q):
        return tfl_journeys(origin, dest, when, q[1], app_key, deadline=deadline)

    if pool is not None:
        results = list(pool.map(query, TRANSPORT_QUERIES))
    else:
        results = [query(q) for q in TRANSPORT_QUERIES]

    options = []
    for (label, _), journeys in zip(TRANSPORT_QUERIES, results):
        for j in journeys:
            if not isinstance(j.get("duration"), int):
                continue
            priced = price_journey(j, adults, railcard_holders)
            if priced is None:
                continue
            out, ret, desc = priced
            options.append({
                "kind": label,
                "outbound_p": out,
                "return_p": ret,
                "total_p": out + ret,
                "minutes": j["duration"],
                "route": desc,
            })
    if not options:
        return None
    fastest = min(options, key=lambda o: o["minutes"])
    ok = [o for o in options if o["minutes"] <= max_minutes] or [fastest]
    cheapest = min(ok, key=lambda o: (o["total_p"], o["minutes"]))
    cheapest["fastest_minutes"] = fastest["minutes"]
    cheapest["fastest_total_p"] = fastest["total_p"]
    cheapest["fastest_route"] = fastest["route"]
    return cheapest


def cycle_time(origin, hotel, when, app_key, deadline=None, journeys=None):
    """(minutes, km, source) for cycling to the hotel. TfL routing only with an app key.

    `journeys` may carry an already-fetched TfL cycle response to avoid a second call.
    """
    dest = (hotel["lat"], hotel["lon"])
    est_min, est_km = estimate_cycle(origin, dest)
    if not app_key:
        return est_min, est_km, "estimate"
    if journeys is None:
        journeys = tfl_journeys(origin, dest, when, CYCLE_QUERY, app_key, deadline=deadline)
    for j in journeys:
        if isinstance(j.get("duration"), int):
            km = sum((l.get("distance") or 0) for l in (j.get("legs") or [])) / 1000
            return j["duration"], round(km, 1) if km else est_km, "tfl"
    return est_min, est_km, "estimate"


# --------------------------------------------------------------------------- #
# Search
# --------------------------------------------------------------------------- #
_UNPRICED = {
    "transport_out": None,
    "transport_return": None,
    "transport_total": None,
    "travel_minutes": None,
    "fastest_minutes": None,
    "fastest_total": None,
    "fastest_route": None,
    "total": None,
}


def _transport_fields(hotel, t):
    if not t:
        return dict(_UNPRICED, route="no TfL fare found")
    return {
        "transport_out": t["outbound_p"] / 100,
        "transport_return": t["return_p"] / 100,
        "transport_total": t["total_p"] / 100,
        "travel_minutes": t["minutes"],
        "route": t["route"],
        "fastest_minutes": t["fastest_minutes"],
        "fastest_total": t["fastest_total_p"] / 100,
        "fastest_route": t["fastest_route"],
        "total": round(hotel["room_price"] + t["total_p"] / 100, 2),
    }


def sort_rows(rows):
    rows.sort(key=lambda r: (r["total"] is None, r["total"] or 0, r["room_price"]))


def run_search(
    *,
    location: str,
    checkin: date,
    checkout: date,
    rooms: list,
    railcard_holders: int,
    origin_name: str,
    origin: tuple,
    depart: str,
    top: int,
    max_miles,
    max_travel_min: int,
):
    """Generator of progress events for one search.

    Yields, in order:
      ("hotels", summary, rows)   every shortlisted hotel with room price and
                                   cycling time but transport still unpriced
      ("transport", row)          one per hotel as TfL prices come back
      ("done", summary, rows)     final summary and rows sorted by total
    """
    nights = (checkout - checkin).days
    adults = sum(a for a, _ in rooms)
    children = sum(c for _, c in rooms)
    railcard_holders = max(0, min(adults, railcard_holders))
    hh, mm = map(int, depart.split(":"))
    depart_dt = datetime(checkin.year, checkin.month, checkin.day, hh, mm)
    app_key = os.environ.get("TFL_APP_KEY") or None
    deadline = time.monotonic() + SEARCH_DEADLINE_S

    raw = fetch_hotels(location, checkin, checkout, rooms)
    hotels = [normalise_hotel(h) for h in raw]
    avail = [h for h in hotels if h["available"] and h["room_price"] and h["lat"] is not None and h["lon"] is not None]
    if max_miles is not None:
        avail = [h for h in avail if (h["distance_from_search_mi"] or 0) <= max_miles]

    avail.sort(key=lambda h: h["room_price"])
    shortlist = avail[:top]

    rows = []
    by_code = {}
    for h in shortlist:
        row = dict(h, **_UNPRICED, route="pricing…")
        est_min, est_km = estimate_cycle(origin, (h["lat"], h["lon"]))
        row.update({"cycle_minutes": est_min, "cycle_km": est_km, "cycle_source": "estimate"})
        rows.append(row)
        by_code[row["code"]] = row

    summary = {
        "location": location,
        "checkin": checkin,
        "checkout": checkout,
        "nights": nights,
        "rooms": len(rooms),
        "adults": adults,
        "children": children,
        "railcard_holders": railcard_holders,
        "origin_name": origin_name,
        "origin_lat": origin[0],
        "origin_lon": origin[1],
        "depart": depart_dt,
        "max_travel_min": max_travel_min,
        "hotels_found": len(hotels),
        "hotels_available": len(avail),
        "hotels_priced": 0,
        "hotels_shortlisted": len(rows),
        "tfl_rate_per_min": TFL_RATE_KEYED if app_key else TFL_RATE_ANON,
        "truncated": False,
    }
    yield "hotels", dict(summary), [dict(r) for r in rows]

    priced = 0
    with ThreadPoolExecutor(max_workers=TFL_WORKERS) as tfl_pool, \
            ThreadPoolExecutor(max_workers=TFL_HOTEL_WORKERS) as hotel_pool:

        def price(h):
            # Kick off the cycle lookup first so it overlaps the fare lookups.
            cycle_fut = None
            if app_key:
                cycle_fut = tfl_pool.submit(
                    tfl_journeys, origin, (h["lat"], h["lon"]), depart_dt, CYCLE_QUERY, app_key, deadline=deadline
                )
            t = best_transport(origin, h, depart_dt, adults, railcard_holders, max_travel_min, app_key,
                               deadline=deadline, pool=tfl_pool)
            cyc = None
            if cycle_fut is not None:
                cyc = cycle_time(origin, h, depart_dt, app_key, deadline=deadline, journeys=cycle_fut.result())
            return h["code"], t, cyc

        # Submit cheapest rooms first so the most useful rows arrive earliest.
        futures = [hotel_pool.submit(price, h) for h in shortlist]
        try:
            for fut in as_completed(futures):
                code, t, cyc = fut.result()
                row = by_code[code]
                timed_out = t is None and time.monotonic() >= deadline
                row.update(_transport_fields(row, t))
                if timed_out:
                    row["route"] = "not priced (time limit)"
                    summary["truncated"] = True
                if cyc:
                    row["cycle_minutes"], row["cycle_km"], row["cycle_source"] = cyc
                priced += 1
                summary["hotels_priced"] = priced
                yield "transport", dict(row)
        finally:
            for f in futures:
                f.cancel()

    sort_rows(rows)
    yield "done", dict(summary), [dict(r) for r in rows]
