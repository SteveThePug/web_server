"""Cheapest Travelodge (room + TfL transport) for a stay in London.

Ported from ~/cheap_hotel/cheap_travelodge.py. The pricing logic is unchanged;
only the CLI/report/CSV plumbing has been removed.

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
  5. Rank by room + round-trip transport.
"""

import os
import time
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
SEARCH_DEADLINE_S = 240


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
def fetch_hotels(location, checkin, checkout, rooms, page_size=10):
    """Return list of hotel dicts from the Travelodge search API (all pages).

    rooms: list of (adults, children) tuples, one per room.
    """
    hotels, start, seen = [], 0, set()
    while True:
        params = {
            "pagination": "false",
            "checkIn": checkin.isoformat(),
            "checkOut": checkout.isoformat(),
            "q": location,
            "action": "hotel_amend",
            "start": start,
        }
        for i, (adults, children) in enumerate(rooms):
            params[f"rooms[{i}][adults]"] = adults
            params[f"rooms[{i}][children]"] = children
        try:
            r = requests.get(TL_API, params=params, headers=HEADERS, timeout=30)
            r.raise_for_status()
            results = r.json().get("results") or []
        except (requests.RequestException, ValueError, AttributeError) as e:
            raise UpstreamError(f"Travelodge search failed: {e}") from e
        if not results:
            break
        new = [h for h in results if h.get("code") and h.get("code") not in seen]
        if not new:
            break
        for h in new:
            seen.add(h["code"])
        hotels.extend(new)
        start += page_size
        if start > 500:  # safety
            break
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


def sleep_within(seconds, deadline):
    """Sleep for `seconds` unless that would overrun `deadline` (a time.monotonic()
    value, or None for no limit). Returns False if the sleep was skipped."""
    if deadline is not None and time.monotonic() + seconds > deadline:
        return False
    time.sleep(seconds)
    return True


def tfl_journeys(frm, to, when, extra=None, app_key=None, retries=3, deadline=None):
    params = {
        "date": when.strftime("%Y%m%d"),
        "time": when.strftime("%H%M"),
        "timeIs": "Departing",
        "journeyPreference": "LeastTime",
    }
    if extra:
        params.update(extra)
    if app_key:
        params["app_key"] = app_key
    url = TFL_API.format(frm=f"{frm[0]},{frm[1]}", to=f"{to[0]},{to[1]}")
    for attempt in range(retries):
        remaining = 40 if deadline is None else deadline - time.monotonic()
        if remaining < 1:
            return []
        try:
            r = requests.get(url, params=params, headers=HEADERS, timeout=min(40, remaining))
        except requests.RequestException:
            if not sleep_within(2, deadline):
                return []
            continue
        if r.status_code == 429:
            if not sleep_within(10 * (attempt + 1), deadline):
                return []
            continue
        if r.status_code >= 500:
            if not sleep_within(2, deadline):
                return []
            continue
        if r.status_code != 200:
            return []
        try:
            return r.json().get("journeys") or []
        except ValueError:
            return []
    return []


def best_transport(origin, hotel, when, adults, railcard_holders, max_minutes, app_key=None, pause=1.2, deadline=None):
    """Cheapest acceptable round-trip transport to a hotel. Returns dict or None."""
    dest = (hotel["lat"], hotel["lon"])
    options = []
    queries = (("fast", None), ("bus", {"mode": "bus,walking"}))
    for i, (label, extra) in enumerate(queries):
        for j in tfl_journeys(origin, dest, when, extra, app_key, deadline=deadline):
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
        if i < len(queries) - 1:  # pause between TfL calls, not after the last
            sleep_within(pause, deadline)
    if not options:
        return None
    fastest = min(options, key=lambda o: o["minutes"])
    ok = [o for o in options if o["minutes"] <= max_minutes] or [fastest]
    cheapest = min(ok, key=lambda o: (o["total_p"], o["minutes"]))
    cheapest["fastest_minutes"] = fastest["minutes"]
    cheapest["fastest_total_p"] = fastest["total_p"]
    cheapest["fastest_route"] = fastest["route"]
    return cheapest


# --------------------------------------------------------------------------- #
# Search
# --------------------------------------------------------------------------- #
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
    """Everything main() in the original script did between parsing and reporting."""
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
    avail = [h for h in hotels if h["available"] and h["room_price"]]
    if max_miles is not None:
        avail = [h for h in avail if (h["distance_from_search_mi"] or 0) <= max_miles]

    avail.sort(key=lambda h: h["room_price"])
    shortlist = avail[:top]

    rows = []
    truncated = False
    for h in shortlist:
        skipped = time.monotonic() >= deadline
        if skipped:
            truncated = True
            t = None
        else:
            t = best_transport(origin, h, depart_dt, adults, railcard_holders, max_travel_min, app_key, deadline=deadline)
        row = dict(h)
        if t:
            row.update({
                "transport_out": t["outbound_p"] / 100,
                "transport_return": t["return_p"] / 100,
                "transport_total": t["total_p"] / 100,
                "travel_minutes": t["minutes"],
                "route": t["route"],
                "fastest_minutes": t["fastest_minutes"],
                "fastest_total": t["fastest_total_p"] / 100,
                "fastest_route": t["fastest_route"],
                "total": round(h["room_price"] + t["total_p"] / 100, 2),
            })
        else:
            row.update({
                "transport_out": None,
                "transport_return": None,
                "transport_total": None,
                "travel_minutes": None,
                "route": "not priced (time limit)" if skipped else "no TfL fare found",
                "fastest_minutes": None,
                "fastest_total": None,
                "fastest_route": None,
                "total": None,
            })
        rows.append(row)

    rows.sort(key=lambda r: (r["total"] is None, r["total"] or 0, r["room_price"]))

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
        "hotels_priced": len(rows),
        "truncated": truncated,
    }
    return {"summary": summary, "rows": rows}
