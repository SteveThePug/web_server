"""Cheapest Travelodge (room + TfL transport) for a stay in London.

Ported from ~/cheap_hotel/cheap_travelodge.py. The pricing logic is unchanged;
the plumbing around it has been rewritten for speed:

  * Travelodge result pages are fetched in parallel batches.
  * TfL journeys are fetched by one flat worker pool (every hotel/query pair is
    submitted to it and results gathered per hotel), throttled to the anonymous API
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

Fragility
  Neither API is a documented public contract for this use. Travelodge's
  /api/v2/hotel endpoint is the JSON feed its own search page calls; the query
  parameter names ("q", "checkIn", "rooms[0][adults]", the undocumented
  "action=hotel_amend") and the response field names consumed in
  `normalise_hotel` are whatever the site happens to send today. If Travelodge
  reshapes its search page, every price here silently becomes None rather than
  raising. Same for TfL's fare JSON: `price_journey` reads `fare.fares[].cost`
  / `peak` / `offPeak` / `chargeLevel`, and falls back to "no fare found" if the
  shape changes. There is no HTML parsing anywhere in this module - both
  upstreams are JSON - so there are no CSS/XPath selectors to break, but the
  key names play the same role.

Concurrency
  Everything here is thread-based (requests is blocking), not asyncio. Module
  state - the two :class:`TTLCache` response caches, the rate limiters and the shared
  requests.Session - is process-global and guarded by plain locks, so it is
  shared across concurrent FastAPI requests. main.py additionally allows only
  one live search at a time, which is what keeps the TfL budget spendable by a
  single run.

Time budgets
  Every network path takes an optional `deadline` (a time.monotonic() value).
  Passing it down instead of using a wall-clock timeout lets a long run degrade
  gracefully: work already done is returned with summary["truncated"] = True
  instead of the whole request dying at the nginx 300s proxy timeout.
"""

import logging
import math
import os
import threading
import time
from collections import deque
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import date, datetime, timedelta

import requests

# Every upstream field is read tolerantly, so a Travelodge schema change would
# otherwise be indistinguishable from genuinely sold-out inventory. This logger
# is the only signal that the difference exists - see `hotels_for`.
log = logging.getLogger(__name__)

TL_API = "https://www.travelodge.co.uk/api/v2/hotel"
TL_SITE = "https://www.travelodge.co.uk"
TFL_API = "https://api.tfl.gov.uk/Journey/JourneyResults/{frm}/to/{to}"
# A browser User-Agent is sent because Travelodge's edge rejects or throttles
# obviously-scripted clients (python-requests/x.y). The Chrome version is
# arbitrary and is never checked for freshness; if requests start coming back
# as 403 or as a challenge page, this is the first thing to bump. The same
# headers are reused for TfL, where they are harmless.
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

# A date-range scan fetches one Travelodge search per candidate night, then
# prices TfL once per unique hotel. Stop starting new date fetches after the
# soft deadline; stop TfL pricing at the hard one and return what we have.
SCAN_FETCH_DEADLINE_S = 150
SCAN_DEADLINE_S = 240
SCAN_DATE_WORKERS = int(os.environ.get("TL_SCAN_WORKERS", "4"))
SCAN_MAX_DAYS = 91

# Normalised Travelodge results are cached per (location, dates, rooms) so a
# re-scan with different weekday/transport settings is free on the Travelodge
# side. Prices may therefore be up to TL_CACHE_TTL_S old.
TL_CACHE_TTL_S = 30 * 60
TL_CACHE_MAX = 600

# Travelodge returns 10 hotels per page whatever you ask for; fetch pages in
# parallel batches until a batch comes back with nothing new.
TL_PAGE_SIZE = 10
TL_PAGE_BATCH = 6
TL_MAX_START = 500

# TfL: anonymous callers get 50 requests/min, keyed callers 500/min. A call
# normally takes 1-5s; the odd one hangs, so time out early and retry.
TFL_WORKERS = 12
TFL_TIMEOUT_S = 15
TFL_RATE_ANON = 48
TFL_RATE_KEYED = 450
TFL_CACHE_TTL_S = 2 * 60 * 60

# Cycling estimate when TfL cycle routing isn't used: straight-line distance
# times a road-winding factor, at a relaxed London pace.
CYCLE_ROUTE_FACTOR = 1.3
CYCLE_KMH = 14.0

# One process-wide Session so TLS handshakes and any cookies the upstreams set
# are reused across requests (a fresh Session per call roughly doubles latency).
# pool_maxsize must cover the widest possible fan-out - TfL workers plus every
# scan date worker running its own batch of Travelodge page fetches - otherwise
# urllib3 starts discarding connections and logging pool-full warnings.
_session = requests.Session()
_session.mount(
    "https://",
    requests.adapters.HTTPAdapter(pool_connections=4, pool_maxsize=TFL_WORKERS + SCAN_DATE_WORKERS * TL_PAGE_BATCH),
)


class UpstreamError(Exception):
    """Travelodge or TfL could not be reached or returned garbage."""


class TTLCache:
    """A small thread-safe dict with per-entry expiry, shared by every cache here.

    There used to be three hand-rolled copies of this (Travelodge results, TfL
    journeys, and main.py's response cache) whose eviction rules genuinely
    differed, so the differences are constructor parameters rather than being
    quietly unified:

      * ``ttl_s`` - how long a stored value stays fresh.
      * ``max_entries`` - hard cap on size, or None for unbounded. Eviction is
        *insertion order*, not LRU: dicts keep insertion order, so dropping the
        first key drops the oldest *written* entry regardless of recent reads.
      * ``sweep_above`` - only walk the whole dict looking for expired entries
        once it holds more than this many. 0 (the default) means sweep on every
        write; a large value keeps the common path O(1) for a big cache whose
        entries are cheap to keep around.

    Expiry is by :func:`time.monotonic`, so a wall-clock change cannot make an
    entry immortal. Values are returned by reference and callers must treat them
    as read-only.
    """

    def __init__(self, ttl_s: float, max_entries: int | None = None, sweep_above: int = 0) -> None:
        self.ttl_s = ttl_s
        self.max_entries = max_entries
        self.sweep_above = sweep_above
        self._store: dict = {}
        self._lock = threading.Lock()

    def get(self, key):
        """Return the value stored for `key`, or None if absent or expired."""
        now = time.monotonic()
        with self._lock:
            hit = self._store.get(key)
            if hit and hit[0] > now:
                return hit[1]
        return None

    def set(self, key, value) -> None:
        """Store `value` for `key` for ttl_s seconds, evicting as configured."""
        with self._lock:
            now = time.monotonic()
            if len(self._store) > self.sweep_above:
                for k in [k for k, (exp, _) in self._store.items() if exp <= now]:
                    del self._store[k]
            if self.max_entries is not None:
                # >= because this write is about to add one more entry.
                while len(self._store) >= self.max_entries and key not in self._store:
                    del self._store[next(iter(self._store))]
            self._store[key] = (now + self.ttl_s, value)

    def __len__(self) -> int:
        with self._lock:
            return len(self._store)


# --------------------------------------------------------------------------- #
# Input parsing
# --------------------------------------------------------------------------- #
def parse_guests(spec: str) -> list[tuple[int, int]]:
    """Parse the compact guests string into per-room occupancy.

    Comma separates rooms, "+" separates adults from children within a room,
    so '2,2+1' -> [(2, 0), (2, 1)]: two rooms, the second with one child.

    Args:
        spec: user-supplied guests string, e.g. "1", "2+1", "2,2+2".

    Returns:
        One (adults, children) tuple per room, in the order given. The index of
        each tuple becomes the ``rooms[i][...]`` index in the Travelodge query.

    Raises:
        ValueError: empty spec, non-numeric part, a room with no adult, or a
            negative child count.
    """
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


def parse_origin(spec: str) -> tuple[str, tuple[float, float]]:
    """Resolve an origin to a display name and coordinates.

    Args:
        spec: either a key of :data:`ORIGINS` (case-insensitive) or a raw
            "lat,lon" pair.

    Returns:
        (display name, (lat, lon)). For raw coordinates the display name is the
        coordinates re-formatted to 4 decimal places (~11m), which is both the
        label shown to the user and a stable cache-key component.

    Raises:
        ValueError: unknown station, unparseable coordinates, or coordinates
            outside :data:`LONDON_BOX`. The box check exists because TfL's
            journey planner will happily route from anywhere in the country and
            burn the rate-limit budget on journeys this tool cannot price
            sensibly.
    """
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
def _fetch_page(params: dict, start: int, deadline: float | None = None) -> list[dict]:
    """Fetch one page of Travelodge search results.

    Args:
        params: the shared query parameters built by :func:`fetch_hotels`.
        start: zero-based offset of the first result wanted; the API paginates
            by offset, not page number, in steps of :data:`TL_PAGE_SIZE`.
        deadline: time.monotonic() value; the socket timeout is shortened so a
            hung page cannot outlive the caller's budget.

    Returns:
        The raw ``results`` array, or [] when the key is missing - an empty page
        is the only signal the API gives that pagination is exhausted.

    Raises:
        UpstreamError: transport failure, non-2xx status, or a body that is not
            JSON (which in practice means a challenge/error HTML page).
    """
    p = dict(params, start=start)
    timeout = 30
    if deadline is not None:
        timeout = max(1, min(timeout, deadline - time.monotonic()))
    try:
        r = _session.get(TL_API, params=p, headers=HEADERS, timeout=timeout)
        r.raise_for_status()
        return r.json().get("results") or []
    except (requests.RequestException, ValueError, AttributeError) as e:
        # ValueError covers r.json() on a non-JSON body; AttributeError covers a
        # JSON body that is a list or scalar rather than an object (no .get).
        raise UpstreamError(f"Travelodge search failed: {e}") from e


def fetch_hotels(location, checkin, checkout, rooms, deadline=None) -> list[dict]:
    """Return the raw hotel dicts for a search, following pagination.

    Args:
        location: free-text search string, passed straight through as ``q``.
        checkin, checkout: :class:`datetime.date`; the API wants ISO dates.
        rooms: list of (adults, children) tuples from :func:`parse_guests`.
        deadline: time.monotonic() value limiting each page request.

    Returns:
        Hotel dicts in API order, de-duplicated by ``code``. Order matters: it
        is the site's own relevance ranking and later sorts are stable, so ties
        on price keep the site's ordering.

    Raises:
        UpstreamError: propagated from :func:`_fetch_page`.

    Note:
        ``action=hotel_amend`` is an undocumented parameter copied from the
        site's own XHR; without it the endpoint answers differently. Likewise
        ``pagination=false`` does *not* disable paging - it only suppresses the
        pagination metadata block - which is why the offset loop below exists.
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

    hotels: list[dict] = []
    seen: set[str] = set()

    def absorb(results: list[dict]) -> bool:
        """Append hotels not seen before; return True if any were new."""
        new = [h for h in results if h.get("code") and h.get("code") not in seen]
        for h in new:
            seen.add(h["code"])
        hotels.extend(new)
        return bool(new)

    # First page alone: tells us whether the search matched anything at all.
    first = _fetch_page(params, 0, deadline)
    if not absorb(first) or len(first) < TL_PAGE_SIZE:
        return hotels

    # Fetch TL_PAGE_BATCH pages at a time and stop on the first batch that is
    # either short (a page with fewer than TL_PAGE_SIZE results is the last one)
    # or wholly duplicate. Requesting speculatively overshoots by up to a batch,
    # but a strictly sequential walk of ~50 pages would blow the time budget.
    # TL_MAX_START is a hard stop in case the API ever pages forever.
    start = TL_PAGE_SIZE
    with ThreadPoolExecutor(max_workers=TL_PAGE_BATCH) as ex:
        while start <= TL_MAX_START:
            starts = [start + i * TL_PAGE_SIZE for i in range(TL_PAGE_BATCH)]
            pages = list(ex.map(lambda s: _fetch_page(params, s, deadline), starts))
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


def normalise_hotel(h: dict) -> dict:
    """Map one raw Travelodge hotel dict onto this service's stable field names.

    This is the single place where upstream key names are read, so it is the
    only file location that needs changing when Travelodge renames a field.
    Every lookup is tolerant: a renamed or missing key yields None (or False)
    rather than an exception, so a shape change degrades into "no price"
    instead of a 502. ``minPrice`` is the cheapest bookable room for the whole
    stay in pounds (already a number, no currency string to parse), and
    ``hotelUrl`` is site-relative so :data:`TL_SITE` is prefixed here.
    """
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


# (location, checkin, checkout, rooms) -> [normalised hotel dicts]. Bounded by
# count, and expired entries are swept on every write (the cache is small).
_tl_cache = TTLCache(TL_CACHE_TTL_S, max_entries=TL_CACHE_MAX)


def hotels_for(location, checkin, checkout, rooms, deadline=None):
    """Normalised hotel list for a search, cached for TL_CACHE_TTL_S.

    Failures are not cached. Callers must treat the returned dicts as read-only.
    """
    # Case/whitespace-insensitive location so "Central London" and
    # "central london " share one cache entry; rooms must be hashable, hence
    # the tuple.
    key = (location.strip().lower(), checkin, checkout, tuple(rooms))
    hit = _tl_cache.get(key)
    if hit is not None:
        return hit

    # The fetch deliberately happens outside any cache lock: a slow upstream
    # must not block other threads reading the cache. Two threads racing on the
    # same key both fetch and the second write wins, which is harmless.
    raw = fetch_hotels(location, checkin, checkout, rooms, deadline=deadline)
    hotels = [normalise_hotel(h) for h in raw]

    # A response that parsed but yielded nothing usable is the fingerprint of a
    # renamed upstream field: the tolerant .get()s in `normalise_hotel` turn a
    # schema change into "no hotels available", which looks exactly like a
    # sold-out search. Genuine sold-out inventory still carries names and codes,
    # so warn only when the shape itself came back empty.
    if raw and not any(h["code"] and h["name"] for h in hotels):
        log.warning(
            "Travelodge returned %d hotels for %r (%s..%s) but none had a usable "
            "code/name - upstream field names may have changed",
            len(raw), location, checkin, checkout,
        )
    elif raw and not any(h["available"] and h["room_price"] is not None for h in hotels):
        log.warning(
            "Travelodge returned %d hotels for %r (%s..%s) but none were priced "
            "and available - sold out, or minPrice/hasAvailability renamed",
            len(raw), location, checkin, checkout,
        )

    _tl_cache.set(key, hotels)
    return hotels


# --------------------------------------------------------------------------- #
# Geometry
# --------------------------------------------------------------------------- #
def haversine_km(a: tuple[float, float], b: tuple[float, float]) -> float:
    """Great-circle distance in km between two (lat, lon) points in degrees.

    6371.0 km is the mean Earth radius; at London scale the error from assuming
    a sphere is far smaller than the road-winding fudge applied on top.
    """
    lat1, lon1 = map(math.radians, a)
    lat2, lon2 = map(math.radians, b)
    h = math.sin((lat2 - lat1) / 2) ** 2 + math.cos(lat1) * math.cos(lat2) * math.sin((lon2 - lon1) / 2) ** 2
    return 2 * 6371.0 * math.asin(math.sqrt(h))


def estimate_cycle(origin: tuple[float, float], dest: tuple[float, float]) -> tuple[int, float]:
    """Estimate (minutes, km) for a cycle journey without calling TfL.

    Straight-line distance is inflated by :data:`CYCLE_ROUTE_FACTOR` (roads do
    not go in straight lines) and divided by :data:`CYCLE_KMH`, a relaxed urban
    pace that absorbs traffic lights. Both constants are hand-tuned guesses, not
    measurements; they are only used when no TFL_APP_KEY is configured. Minutes
    are clamped to at least 1 so a hotel next to the origin never reads "0 min".
    """
    km = haversine_km(origin, dest) * CYCLE_ROUTE_FACTOR
    return max(1, round(km / CYCLE_KMH * 60)), round(km, 1)


# --------------------------------------------------------------------------- #
# TfL
# --------------------------------------------------------------------------- #
def round_5p(pence: float) -> int:
    """Round a pence amount to the nearest 5p.

    TfL quotes every discounted fare in whole 5p steps, so a raw 1/3-off
    multiplication has to be snapped back onto that grid to match what the
    passenger is actually charged. Python's banker's rounding applies at exact
    half-steps (2.5p), which is a sub-penny discrepancy and ignored.
    """
    return int(5 * round(pence / 5.0))


def price_journey(journey, adults=1, railcard_holders=1):
    """
    Return (outbound_pence, return_pence, detail) for one TfL journey for the
    whole party, or None if the fare can't be determined.

    adults: number of fare-paying adults.  railcard_holders: how many of them
    get 1/3 off off-peak rail fares (0 = nobody has a railcard).

    All money is integer pence, straight from TfL - no currency strings are
    parsed and no floats are introduced until the final division by 100 in
    :func:`_transport_fields`, which keeps the arithmetic exact.

    Children are not counted: the caller passes adults only, on the assumption
    that under-11s travel free on TfL.

    The return leg is modelled, not fetched. It is assumed to be the same route
    made the next morning at an off-peak time, so it is priced at the outbound
    fare's ``offPeak`` value. That is optimistic for an early-morning return
    into zone 1, which is a real peak fare.
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

    # TfL omits the fare block entirely for some journeys (e.g. ones involving
    # a mode it cannot price). Returning None makes the caller drop this option
    # rather than treat it as free.
    if not fare or not fare.get("fares"):
        return None

    out = ret = 0
    for f in fare["fares"]:
        cost = f.get("cost") or 0
        peak, offpeak = f.get("peak") or 0, f.get("offPeak") or 0
        # Two independent bus tests, because TfL is inconsistent about which
        # it provides: a fare with neither a peak nor an off-peak price is a
        # flat bus fare by construction, and otherwise the tap records name the
        # mode outright. Getting this wrong would apply a Railcard discount to
        # a bus fare, which does not exist.
        is_bus = (peak == 0 and offpeak == 0) or any(
            (t.get("tapDetails") or {}).get("modeType") == "Bus" for t in (f.get("taps") or [])
        )
        if is_bus:
            # Bus: flat fare, hopper legs already come through as cost 0. No railcard.
            out += cost * adults
            ret += cost * adults
        else:
            # Railcards give 1/3 off *off-peak* rail/tube fares only, so the
            # outbound is discounted only when this fare is off-peak. That is
            # detected two ways because chargeLevel is free text that has
            # historically read "Off Peak", "Off-peak" or been absent: a
            # substring match on "off", or the cost simply equalling the
            # off-peak price. The return is always treated as off-peak (see the
            # docstring), so it is always discounted for railcard holders.
            level = (f.get("chargeLevel") or "").lower()
            out_full = cost
            ret_full = offpeak or cost  # next-day return assumed off-peak
            out_disc = round_5p(out_full * 2 / 3) if ("off" in level or cost == offpeak) else out_full
            ret_disc = round_5p(ret_full * 2 / 3)
            out += out_full * full_fare + out_disc * railcard_holders
            ret += ret_full * full_fare + ret_disc * railcard_holders

    # Short human summary of the route, e.g. "rail > tube Victoria". Walking
    # legs are dropped as noise, and "national-rail" is shortened purely to fit
    # the column in the UI.
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

    def __init__(self, per_minute: int) -> None:
        """Args: per_minute: maximum acquisitions allowed in any 60s window."""
        self.per_minute = per_minute
        self._times = deque()
        self._lock = threading.Lock()

    def acquire(self, deadline: float | None = None) -> bool:
        """Block until a slot is free.

        Args:
            deadline: time.monotonic() value; if the wait would run past it,
                give up immediately instead of sleeping.

        Returns:
            True when a slot was taken, False when the deadline would pass
            first (the caller should then abandon the request).

        The sleep is capped at 1s per iteration rather than sleeping the full
        computed wait, so that a slot freed by another thread finishing early is
        picked up promptly, and so the deadline is re-checked often. The lock is
        released before sleeping - holding it would serialise every waiter.
        """
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


def _limiter(app_key: str | None) -> RateLimiter:
    """Return the shared limiter for the rate that `app_key` unlocks.

    Limiters are keyed by rate, not by key, so every caller at a given tier
    shares one 60s window - which is the point, since TfL counts per source IP.
    """
    rate = TFL_RATE_KEYED if app_key else TFL_RATE_ANON
    with _limiters_lock:
        lim = _limiters.get(rate)
        if lim is None:
            lim = _limiters[rate] = RateLimiter(rate)
        return lim


# journey key -> journeys. Unbounded on purpose (entries are small and a long
# scan wants them all), so expired entries are only swept once it is large,
# keeping the common write O(1).
_tfl_cache = TTLCache(TFL_CACHE_TTL_S, sweep_above=5000)


def sleep_within(seconds, deadline):
    """Sleep for `seconds` unless that would overrun `deadline` (a time.monotonic()
    value, or None for no limit). Returns False if the sleep was skipped."""
    if deadline is not None and time.monotonic() + seconds > deadline:
        return False
    time.sleep(seconds)
    return True


def _tfl_journeys(frm, to, when, extra=None, app_key=None, retries=3, deadline=None) -> tuple[list[dict], bool]:
    """Journeys from TfL, plus whether the call was abandoned on the deadline.

    Same contract as :func:`tfl_journeys` but returns (journeys, abandoned).
    `abandoned` is True only when every attempt was given up because the
    deadline was too close (no slot from the rate limiter, no time left for the
    request, or a backoff sleep that would overrun it). It is False for a
    genuine "TfL has no route/fare here" answer, which is what lets the caller
    tell those two cases apart instead of guessing from the clock.

    Args:
        frm, to: (lat, lon) tuples; TfL accepts raw coordinates in the path.
        when: naive :class:`datetime` of the intended departure, London local
            time (TfL interprets it as such; no timezone is sent).
        extra: additional query parameters, e.g. ``{"mode": "bus,walking"}`` to
            force a bus-only route or ``{"mode": "cycle"}`` for cycle routing.
        app_key: TfL API key from the TFL_APP_KEY env var, or None for the
            anonymous tier.
        retries: attempts before giving up.
        deadline: time.monotonic() value bounding the whole call.

    Returns:
        The ``journeys`` array, or [] on any failure. This function never
        raises: a missing route and a dead API look the same to the caller, who
        simply ends up with no priced option for that hotel. That is deliberate
        - one flaky lookup should not abort a 40-hotel search - but it does mean
        a total TfL outage shows up in the UI as "no TfL fare found" rather than
        as an error.
    """
    params = {
        "date": when.strftime("%Y%m%d"),
        "time": when.strftime("%H%M"),
        "timeIs": "Departing",
        "journeyPreference": "LeastTime",
    }
    if extra:
        params.update(extra)
    # Coordinates are rounded to 5dp (~1m) so that floating-point noise in
    # otherwise identical requests still hits the cache. app_key is excluded
    # from the key on purpose: the route and fare do not depend on it.
    key = (round(frm[0], 5), round(frm[1], 5), round(to[0], 5), round(to[1], 5), params["date"], params["time"],
           tuple(sorted((extra or {}).items())))
    hit = _tfl_cache.get(key)
    if hit is not None:
        return hit, False

    if app_key:
        params["app_key"] = app_key
    url = TFL_API.format(frm=f"{frm[0]},{frm[1]}", to=f"{to[0]},{to[1]}")
    limiter = _limiter(app_key)
    journeys = None
    # Retry policy, all deadline-aware (a skipped sleep means the deadline is
    # too close, so stop rather than hammer):
    #   network error -> wait 2s
    #   429 rate limited -> linear backoff, 5s then 10s then 15s
    #   5xx -> wait 2s
    #   any other non-200 (notably 300 "disambiguation", meaning TfL wants the
    #     caller to pick between candidate locations) -> treat as "no route",
    #     do not retry, and cache the empty result
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

    # journeys is still None only if every attempt was abandoned; return empty
    # without caching so a later, less rushed call can try again. An empty list
    # from a *successful* response is cached - "no route" is a real answer.
    if journeys is None:
        return [], True
    # TFL_CACHE_TTL_S is long because journey times and fares for a fixed
    # date/time do not change.
    _tfl_cache.set(key, journeys)
    return journeys, False


def tfl_journeys(frm, to, when, extra=None, app_key=None, retries=3, deadline=None) -> list[dict]:
    """Journeys from TfL for one origin/destination/time/mode, cached and rate limited.

    Args:
        frm, to: (lat, lon) tuples; TfL accepts raw coordinates in the path.
        when: naive :class:`datetime` of the intended departure, London local
            time (TfL interprets it as such; no timezone is sent).
        extra: additional query parameters, e.g. ``{"mode": "bus,walking"}`` to
            force a bus-only route or ``{"mode": "cycle"}`` for cycle routing.
        app_key: TfL API key from the TFL_APP_KEY env var, or None for the
            anonymous tier.
        retries: attempts before giving up.
        deadline: time.monotonic() value bounding the whole call.

    Returns:
        The ``journeys`` array, or [] on any failure. This function never
        raises: a missing route and a dead API look the same to the caller, who
        simply ends up with no priced option for that hotel. That is deliberate
        - one flaky lookup should not abort a 40-hotel search - but it does mean
        a total TfL outage shows up in the UI as "no TfL fare found" rather than
        as an error. Callers that need to tell "no route" from "abandoned on the
        deadline" apart should call :func:`_tfl_journeys` instead.
    """
    return _tfl_journeys(frm, to, when, extra, app_key, retries, deadline)[0]


# Two journey queries per hotel, because TfL's default "LeastTime" answer is
# usually a tube fare and the cheapest option is very often a slower bus. The
# bus query has to name "walking" alongside "bus" or TfL refuses routes that
# need a walk to the stop.
TRANSPORT_QUERIES = (("fast", None), ("bus", {"mode": "bus,walking"}))
CYCLE_QUERY = {"mode": "cycle"}


def _score_transport(results, adults, railcard_holders, max_minutes):
    """Pick the cheapest acceptable option out of already-fetched journeys.

    Args:
        results: iterable of (label, journeys) - the "fast" and "bus" answers
            from TfL, in any order.
        adults, railcard_holders: party makeup, passed to :func:`price_journey`.
        max_minutes: routes slower than this are rejected - unless *every* route
            is slower, in which case the fastest is kept so the hotel still gets
            a price rather than vanishing from the results.

    Returns:
        The chosen option as a dict with keys ``kind`` ("fast"/"bus"),
        ``outbound_p``, ``return_p``, ``total_p`` (integer pence), ``minutes``,
        ``route``, plus ``fastest_*`` fields describing the quickest option so
        the UI can show the speed/price trade-off. None if nothing was priceable.

    Ties are broken by duration, so two equally cheap routes resolve to the
    quicker one. This is pure - no network, no clock - which is what lets the
    fetching be reorganised without touching the pricing rules.
    """
    options = []
    for label, journeys in results:
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


def best_transport(origin, hotel, when, adults, railcard_holders, max_minutes, app_key=None, deadline=None, pool=None):
    """Cheapest acceptable round-trip transport to a single hotel.

    Convenience wrapper: fetch both TfL queries for one hotel, then score them
    with :func:`_score_transport` (see there for the return shape and the
    selection rules). :func:`_price_hotels` does not use this - it submits every
    (hotel, query) pair to one flat pool instead - so this is the one-off path.

    Args:
        origin: (lat, lon) of the journey start.
        hotel: normalised hotel dict; only lat/lon are read.
        when: departure datetime for the outbound leg.
        adults, railcard_holders: party makeup.
        max_minutes: slowest acceptable route, softened as described above.
        app_key: TfL key or None.
        deadline: time.monotonic() budget.
        pool: optional executor used to run the two independent queries
            concurrently. Only pass a pool this call is not itself running
            inside, or the map below can block behind its own parent task.
    """
    dest = (hotel["lat"], hotel["lon"])

    def query(q):
        return tfl_journeys(origin, dest, when, q[1], app_key, deadline=deadline)

    if pool is not None:
        results = list(pool.map(query, TRANSPORT_QUERIES))
    else:
        results = [query(q) for q in TRANSPORT_QUERIES]
    return _score_transport(
        [(label, journeys) for (label, _), journeys in zip(TRANSPORT_QUERIES, results)],
        adults, railcard_holders, max_minutes,
    )


def cycle_time(origin, hotel, when, app_key, deadline=None, journeys=None):
    """(minutes, km, source) for cycling to the hotel.

    Real TfL cycle routing is used only when an app key is configured, because
    it costs a third rate-limited request per hotel and the anonymous 50/min
    budget is already spent on fares; without a key this falls straight back to
    :func:`estimate_cycle`. ``source`` is "tfl" or "estimate" so the UI can say
    which it is showing.

    Args:
        journeys: an already-fetched cycle response, so the caller can start the
            request early and overlap it with the fare lookups.

    Returns:
        (minutes, km, source). The km comes from summing per-leg ``distance``
        values, which TfL reports in metres; if that sum is zero (the field is
        sometimes absent) the straight-line estimate is used for distance while
        keeping TfL's duration.
    """
    dest = (hotel["lat"], hotel["lon"])
    est_min, est_km = estimate_cycle(origin, dest)
    if not app_key:
        return est_min, est_km, "estimate"
    if journeys is None:
        journeys = tfl_journeys(origin, dest, when, CYCLE_QUERY, app_key, retries=1, deadline=deadline)
    for j in journeys:
        if isinstance(j.get("duration"), int):
            km = sum((l.get("distance") or 0) for l in (j.get("legs") or [])) / 1000
            return j["duration"], round(km, 1) if km else est_km, "tfl"
    return est_min, est_km, "estimate"


# --------------------------------------------------------------------------- #
# Search
# --------------------------------------------------------------------------- #
# Every transport column blanked out. Splatted into a row to reset it, so
# adding a transport field means adding it here too or it will leak a stale
# value from a previous night when rows are reused in a scan.
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

TIME_LIMIT_ROUTE = "not priced (time limit)"


def _transport_fields(hotel: dict, t: dict | None) -> dict:
    """Convert a :func:`best_transport` result into the row's output columns.

    Pence become pounds here (the only place), and ``total`` is room price plus
    round-trip transport, rounded to 2dp for display. When `t` is None every
    transport column is None and ``route`` carries the reason, which is what the
    UI renders in place of a price.
    """
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


def sort_rows(rows: list[dict]) -> None:
    """Sort rows in place by total cost, unpriced rows last.

    ``r["total"] is None`` sorts False(0) before True(1), so priced rows come
    first; ``or 0`` then keeps the comparison from hitting None. Room price is
    the tiebreak.
    """
    rows.sort(key=lambda r: (r["total"] is None, r["total"] or 0, r["room_price"]))


def _shortlist(hotels: list[dict], max_miles: float | None, top: int) -> tuple[list[dict], list[dict]]:
    """Filter and rank hotels, then take the cheapest `top`.

    Hotels are dropped unless they are bookable, have a price, and have
    coordinates - the last because a hotel without lat/lon cannot be routed and
    would crash the TfL step. ``distance_from_search_mi`` is Travelodge's own
    distance from the searched location (miles, not from the travel origin),
    and a missing value is treated as 0 so it survives the filter rather than
    being silently excluded.

    Returns:
        (all qualifying hotels sorted by room price, the cheapest `top` of them).
        Only the shortlist gets transport priced, which is what bounds the
        number of TfL calls.
    """
    avail = [h for h in hotels if h["available"] and h["room_price"] and h["lat"] is not None and h["lon"] is not None]
    if max_miles is not None:
        avail = [h for h in avail if (h["distance_from_search_mi"] or 0) <= max_miles]
    avail.sort(key=lambda h: h["room_price"])
    return avail, avail[:top]


def _make_row(h: dict, origin: tuple[float, float]) -> dict:
    """Build the initial result row for a hotel.

    Transport columns start blank with route "pricing…" - the stream emits these
    rows immediately so the UI can render the hotel list while TfL is still
    being queried, then patches each row in place when its "transport" event
    arrives. The cycle estimate is cheap and local, so it is filled in up front
    and only overwritten later if TfL cycle routing is available.
    """
    row = dict(h, **_UNPRICED, route="pricing…")
    est_min, est_km = estimate_cycle(origin, (h["lat"], h["lon"]))
    row.update({"cycle_minutes": est_min, "cycle_km": est_km, "cycle_source": "estimate"})
    return row


def _price_hotels(hotels, origin, depart_dt, adults, railcard_holders, max_travel_min, app_key, deadline):
    """Price TfL transport for each hotel, cheapest rooms first.

    Generator of (code, transport_or_None, cycle_or_None, timed_out) in the
    order TfL answers. Owns the worker pool; leftover work is cancelled if the
    consumer stops early.

    One flat pool, not nested ones: every (hotel, query) pair - the two fare
    queries plus, with an app key, the cycle query - is submitted to a single
    executor and the results are gathered per hotel as they land. An earlier
    version ran hotels in one pool and their sub-queries in another, which is
    only deadlock-free while both pools are bigger than one worker and the
    nesting is exactly two deep; nothing here waits on the pool from inside the
    pool any more, so that hazard is gone by construction. Submission order is
    still cheapest room first, so the most useful rows are asked for first, and
    the shared :class:`RateLimiter` - not the pool size - is what keeps the run
    inside TfL's per-minute budget.

    `timed_out` means "we gave up on the deadline before getting an answer", as
    reported by :func:`_tfl_journeys`, not inferred from the clock: a hotel TfL
    genuinely cannot price is reported with timed_out False even if its answer
    happens to arrive after the deadline. It is True when the hotel ended up
    unpriced and at least one of its queries was abandoned, since in that case
    we never really finished asking. The caller surfaces it as the
    "not priced (time limit)" route and summary["truncated"].

    The cancel loop in `finally` is best-effort - it only stops tasks that have
    not started - and the `with` block still joins running workers, so an
    abandoned stream takes up to one in-flight TfL timeout to unwind.
    """
    with ThreadPoolExecutor(max_workers=TFL_WORKERS) as pool:
        pending = {}   # code -> {hotel, fares: {label: (journeys, abandoned)}, left, cycle future}
        owner = {}     # future -> (code, label)
        for h in hotels:
            code = h["code"]
            if code in pending:  # a hotel is priced once even if it is listed twice
                continue
            dest = (h["lat"], h["lon"])
            entry = {"hotel": h, "fares": {}, "left": len(TRANSPORT_QUERIES), "cycle": None}
            pending[code] = entry
            for label, extra in TRANSPORT_QUERIES:
                fut = pool.submit(_tfl_journeys, origin, dest, depart_dt, extra, app_key, deadline=deadline)
                owner[fut] = (code, label)
            if app_key:
                # Cycle routing costs a third rate-limited request, so it is
                # only asked for when a key lifts the anonymous budget.
                entry["cycle"] = pool.submit(
                    _tfl_journeys, origin, dest, depart_dt, CYCLE_QUERY, app_key, 1, deadline
                )

        try:
            for fut in as_completed(list(owner)):
                code, label = owner[fut]
                entry = pending[code]
                entry["fares"][label] = fut.result()
                entry["left"] -= 1
                if entry["left"]:
                    continue  # this hotel's other query is still outstanding
                h = entry["hotel"]
                t = _score_transport(
                    [(lbl, entry["fares"][lbl][0]) for lbl, _ in TRANSPORT_QUERIES],
                    adults, railcard_holders, max_travel_min,
                )
                abandoned = any(entry["fares"][lbl][1] for lbl, _ in TRANSPORT_QUERIES)
                cyc = None
                if entry["cycle"] is not None:
                    # Already submitted above, so this wait is usually over by
                    # now; waiting here (in the consumer, never in a worker) is
                    # safe with a single flat pool.
                    cyc = cycle_time(origin, h, depart_dt, app_key, deadline=deadline,
                                     journeys=entry["cycle"].result()[0])
                yield code, t, cyc, (t is None and abandoned)
        finally:
            for f in owner:
                f.cancel()
            for entry in pending.values():
                if entry["cycle"] is not None:
                    entry["cycle"].cancel()


def _apply_transport(row: dict, t: dict | None, cyc: tuple | None, timed_out: bool) -> dict:
    """Merge one hotel's transport and cycle results into its row, in place."""
    row.update(_transport_fields(row, t))
    if timed_out:
        row["route"] = TIME_LIMIT_ROUTE
    if cyc:
        row["cycle_minutes"], row["cycle_km"], row["cycle_source"] = cyc
    return row


def _party(rooms: list[tuple[int, int]], railcard_holders: int) -> tuple[int, int, int]:
    """Totals for the booking: (adults, children, railcard holders).

    Railcard holders are clamped to the number of adults so an over-stated
    value cannot discount more fares than there are payers.
    """
    adults = sum(a for a, _ in rooms)
    children = sum(c for _, c in rooms)
    return adults, children, max(0, min(adults, railcard_holders))


def _depart_dt(day: date, depart: str) -> datetime:
    """Combine a check-in date with an "HH:MM" string into a naive datetime.

    Naive on purpose: TfL reads the date/time as London local time, so
    attaching a timezone here would only invite a double conversion. The format
    is already validated by the Pydantic model in main.py.
    """
    hh, mm = map(int, depart.split(":"))
    return datetime(day.year, day.month, day.day, hh, mm)


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
    adults, children, railcard_holders = _party(rooms, railcard_holders)
    depart_dt = _depart_dt(checkin, depart)
    app_key = os.environ.get("TFL_APP_KEY") or None
    deadline = time.monotonic() + SEARCH_DEADLINE_S

    hotels = hotels_for(location, checkin, checkout, rooms, deadline=deadline)
    avail, shortlist = _shortlist(hotels, max_miles, top)

    rows = [_make_row(h, origin) for h in shortlist]
    by_code = {row["code"]: row for row in rows}

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

    # Submit cheapest rooms first so the most useful rows arrive earliest.
    for code, t, cyc, timed_out in _price_hotels(
        shortlist, origin, depart_dt, adults, railcard_holders, max_travel_min, app_key, deadline
    ):
        row = _apply_transport(by_code[code], t, cyc, timed_out)
        if timed_out:
            summary["truncated"] = True
        summary["hotels_priced"] += 1
        yield "transport", dict(row)

    sort_rows(rows)
    yield "done", dict(summary), [dict(r) for r in rows]


# --------------------------------------------------------------------------- #
# Date-range scan
# --------------------------------------------------------------------------- #
def candidate_checkins(start: date, end: date, weekdays, nights: int) -> list[tuple[date, date]]:
    """Enumerate candidate (checkin, checkout) pairs for a date-range scan.

    Every day from `start` to `end` inclusive whose weekday (0=Mon, matching
    :meth:`date.weekday`) is in `weekdays` becomes a check-in, with check-out
    `nights` days later. Note the checkout may fall past `end` - `end` bounds
    the check-in date, not the stay. main.py also calls this during validation
    purely to reject a range that matches no weekday.
    """
    out = []
    day = start
    while day <= end:
        if day.weekday() in weekdays:
            out.append((day, day + timedelta(days=nights)))
        day += timedelta(days=1)
    return out


def _copy_night(n: dict) -> dict:
    """Two-level copy of a night before yielding it.

    The generator keeps mutating `night_list` after each event, so both the
    night dict and its hotel rows must be copied or the caller's already-emitted
    JSON would change underneath it.
    """
    return dict(n, hotels=[dict(r) for r in n["hotels"]])


def _scan_rooms(*, candidates, by_checkin, summary, location, rooms, max_miles, per_night, origin, fetch_deadline):
    """Phase 1 of a scan: Travelodge room prices, a few candidate nights at a time.

    Mutates each night dict in `by_checkin` in place and bumps the matching
    counters on the shared `summary`, yielding ("night", summary snapshot,
    night copy) as each date answers - in completion order, not date order.
    One bad night must not sink the scan, so an upstream failure is recorded on
    that night only, and a night not started before `fetch_deadline` is marked
    "skipped" and sets summary["truncated"].
    """

    def fetch_night(ci, co):
        if time.monotonic() > fetch_deadline:
            return ci, "skipped", [], 0, "not fetched (time limit)"
        try:
            hotels = hotels_for(location, ci, co, rooms, deadline=fetch_deadline + 30)
        except UpstreamError as e:
            return ci, "failed", [], 0, str(e)
        avail, short = _shortlist(hotels, max_miles, per_night)
        return ci, "ok", short, len(avail), None

    with ThreadPoolExecutor(max_workers=SCAN_DATE_WORKERS) as ex:
        futures = [ex.submit(fetch_night, ci, co) for ci, co in candidates]
        try:
            for fut in as_completed(futures):
                ci, status, short, n_avail, detail = fut.result()
                night = by_checkin[ci]
                night["status"] = status
                night["detail"] = detail
                night["hotels_available"] = n_avail
                if status == "ok":
                    night["hotels"] = [_make_row(h, origin) for h in short]
                    if night["hotels"]:
                        night["best_room"] = night["hotels"][0]["room_price"]
                    summary["nights_fetched"] += 1
                elif status == "failed":
                    summary["nights_failed"] += 1
                else:
                    summary["nights_skipped"] += 1
                    summary["truncated"] = True
                yield "night", dict(summary), _copy_night(night)
        finally:
            for f in futures:
                f.cancel()


def _scan_order(night_list):
    """Decide the order hotels get their (night-independent) fares priced in.

    A hotel can appear on many nights but is priced once, so this collapses the
    nights into one entry per hotel code. Its ordering key is its best (lowest)
    position in any night's list, tiebroken by room price, so hotels that top
    some night get their fares first and a truncated scan still has the likely
    winners priced.

    Returns:
        (ordered, cycle_base) where `ordered` is a list of minimal hotel dicts
        (code/lat/lon - all :func:`_price_hotels` reads) in pricing order, and
        `cycle_base` maps code -> the local cycle estimate already computed for
        that hotel, used when TfL gives no cycle answer. Deliberately *not* the
        live row objects: the ordering record and the output rows used to be the
        same mutable dict, which made it far too easy to write a phase-2 field
        onto one night's row by accident.
    """
    best = {}        # code -> (best rank, room price at that rank)
    coords = {}      # code -> minimal hotel dict for _price_hotels
    cycle_base = {}  # code -> the local cycle estimate already on the rows
    for n in night_list:
        for rank, row in enumerate(n["hotels"]):
            code = row["code"]
            key = (rank, row["room_price"])
            if code not in best or key < best[code]:
                best[code] = key
            if code not in coords:
                coords[code] = {"code": code, "lat": row["lat"], "lon": row["lon"]}
                # The estimate is a pure function of the coordinates, so any
                # night's row carries the same values.
                cycle_base[code] = {k: row[k] for k in ("cycle_minutes", "cycle_km", "cycle_source")}
    # sorted() is stable and `best` is in first-seen order, matching the old
    # index-based ordering exactly.
    ordered = [coords[code] for code in sorted(best, key=lambda c: best[c])]
    return ordered, cycle_base


def _scan_transport(*, ordered, cycle_base, transport, summary, origin, depart_dt, adults, railcard_holders,
                    max_travel_min, app_key, deadline):
    """Phase 2 of a scan: one TfL price per unique hotel, filling `transport`.

    `transport` is an out-parameter (code -> transport + cycle fields) that
    phase 3 then applies to every night containing that code. Yields
    ("transport", summary snapshot, fields copy) per hotel as TfL answers.
    """
    for code, t, cyc, timed_out in _price_hotels(
        ordered, origin, depart_dt, adults, railcard_holders, max_travel_min, app_key, deadline
    ):
        # Transport here is night-independent, so there is no room price to add:
        # a dummy 0 is passed and the resulting "total" discarded. Each night
        # computes its own total in phase 3.
        fields = _transport_fields({"room_price": 0}, t)
        fields.pop("total", None)
        if timed_out:
            fields["route"] = TIME_LIMIT_ROUTE
            summary["truncated"] = True
        if cyc:
            fields["cycle_minutes"], fields["cycle_km"], fields["cycle_source"] = cyc
        else:
            # No TfL cycle result: keep the local estimate computed when the
            # rows were built, since these fields are copied wholesale onto
            # every night's copy of this hotel.
            fields.update(cycle_base.get(code, {}))
        fields["code"] = code
        transport[code] = fields
        summary["hotels_priced"] += 1
        yield "transport", dict(summary), dict(fields)


def _scan_totals(night_list, transport, summary) -> None:
    """Phase 3 of a scan: apply transport to every night and order the results.

    Mutates `night_list` in place: fills each row's transport columns and total,
    sorts each night's hotels, records the night's best total/room/code, and
    finally sorts the nights themselves. A hotel with no entry in `transport`
    was never reached before the deadline, which is marked on the row *and* on
    summary["truncated"] - phase 2 only sets the flag for hotels it actually
    started, so without this a scan could show unpriced rows while claiming it
    was complete.
    """
    for n in night_list:
        if n["status"] != "ok":
            continue
        for row in n["hotels"]:
            fields = transport.get(row["code"])
            # Missing means phase 2 ran out of time before reaching this hotel.
            if fields is None:
                row["route"] = TIME_LIMIT_ROUTE
                summary["truncated"] = True
                continue
            row.update({k: v for k, v in fields.items() if k != "code"})
            if row["transport_total"] is not None:
                row["total"] = round(row["room_price"] + row["transport_total"], 2)
        sort_rows(n["hotels"])
        if n["hotels"]:
            best = n["hotels"][0]
            n["best_total"] = best["total"]
            n["best_code"] = best["code"]
            n["best_room"] = min(r["room_price"] for r in n["hotels"])

    # Cheapest night first; nights with no total (failed, skipped, or unpriced)
    # fall to the bottom, ordered by best room price and then by date so the
    # list is still deterministic.
    night_list.sort(
        key=lambda n: (
            n["best_total"] is None,
            n["best_total"] or 0,
            n["best_room"] is None,
            n["best_room"] or 0,
            n["checkin"],
        )
    )


def run_scan(
    *,
    location: str,
    start: date,
    end: date,
    weekdays,
    nights: int,
    per_night: int,
    rooms: list,
    railcard_holders: int,
    origin_name: str,
    origin: tuple,
    depart: str,
    max_miles,
    max_travel_min: int,
):
    """Generator of progress events for a cheapest-night scan.

    Every candidate check-in in start..end (filtered by weekday) is searched on
    Travelodge and its `per_night` cheapest rooms kept. TfL transport is then
    priced ONCE per unique hotel, at the first candidate night's departure
    time, and reused for every night that hotel appears in.

    Yields, in order:
      ("scan", summary, nights)        every candidate night, status "pending"
      ("night", summary, night)        one per night as Travelodge answers:
                                        status ok/failed/skipped, hotels sorted
                                        by room price, transport unpriced
      ("transport", summary, fields)   one per unique hotel as TfL answers;
                                        `fields` has the hotel code plus the
                                        transport and cycle columns
      ("done", summary, nights)        totals filled in, hotels sorted by
                                        total, nights sorted by best total
    """
    candidates = candidate_checkins(start, end, weekdays, nights)
    if not candidates:
        raise ValueError("no dates in the range match those weekdays")
    adults, children, railcard_holders = _party(rooms, railcard_holders)
    depart_dt = _depart_dt(candidates[0][0], depart)
    app_key = os.environ.get("TFL_APP_KEY") or None
    t0 = time.monotonic()
    fetch_deadline = t0 + SCAN_FETCH_DEADLINE_S
    deadline = t0 + SCAN_DEADLINE_S

    night_list = [
        {
            "checkin": ci,
            "checkout": co,
            "status": "pending",
            "detail": None,
            "hotels_available": 0,
            "hotels": [],
            "best_total": None,
            "best_room": None,
            "best_code": None,
        }
        for ci, co in candidates
    ]
    by_checkin = {n["checkin"]: n for n in night_list}

    summary = {
        "location": location,
        "start": start,
        "end": end,
        "weekdays": sorted(weekdays),
        "nights": nights,
        "per_night": per_night,
        "rooms": len(rooms),
        "adults": adults,
        "children": children,
        "railcard_holders": railcard_holders,
        "origin_name": origin_name,
        "origin_lat": origin[0],
        "origin_lon": origin[1],
        "depart": depart_dt,
        "max_travel_min": max_travel_min,
        "nights_candidate": len(night_list),
        "nights_fetched": 0,
        "nights_failed": 0,
        "nights_skipped": 0,
        "hotels_unique": 0,
        "hotels_priced": 0,
        "tfl_rate_per_min": TFL_RATE_KEYED if app_key else TFL_RATE_ANON,
        "truncated": False,
    }
    yield "scan", dict(summary), [_copy_night(n) for n in night_list]

    # Phase 1: room prices per candidate night.
    yield from _scan_rooms(
        candidates=candidates,
        by_checkin=by_checkin,
        summary=summary,
        location=location,
        rooms=rooms,
        max_miles=max_miles,
        per_night=per_night,
        origin=origin,
        fetch_deadline=fetch_deadline,
    )

    # Phase 2: one TfL price per unique hotel, cheapest-on-some-night first.
    ordered, cycle_base = _scan_order(night_list)
    summary["hotels_unique"] = len(ordered)
    transport = {}  # code -> transport + cycle fields
    yield from _scan_transport(
        ordered=ordered,
        cycle_base=cycle_base,
        transport=transport,
        summary=summary,
        origin=origin,
        depart_dt=depart_dt,
        adults=adults,
        railcard_holders=railcard_holders,
        max_travel_min=max_travel_min,
        app_key=app_key,
        deadline=deadline,
    )

    # Phase 3: totals per night, then ordering.
    _scan_totals(night_list, transport, summary)
    yield "done", dict(summary), [_copy_night(n) for n in night_list]
