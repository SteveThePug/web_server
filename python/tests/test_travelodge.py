"""Tests for travelodge.py's pure functions.

Nothing here touches the network: only the fare rounding, guest parsing,
geometry and row ordering, which are the parts whose behaviour is a decision
rather than an upstream's answer.
"""

import math

import pytest

from app import travelodge as tl


# --------------------------------------------------------------------------- #
# round_5p
# --------------------------------------------------------------------------- #
@pytest.mark.parametrize(
    "pence, want",
    [
        (0, 0),
        (100, 100),
        (101, 100),
        (102, 100),
        (103, 105),
        # Banker's rounding at the exact half-step: 21.5 and 22.5 both round to
        # 22, i.e. 110p. Documented as an ignored sub-penny discrepancy.
        (107.5, 110),
        (112.5, 110),
        # The real use: 1/3 off a rail fare must land on TfL's 5p grid.
        (290 * 2 / 3, 195),
        (490 * 2 / 3, 325),
    ],
)
def test_round_5p_snaps_to_the_5p_grid(pence, want):
    assert tl.round_5p(pence) == want


def test_round_5p_returns_an_int():
    # Fares are integer pence everywhere upstream of the final /100, so this
    # must not leak a float back into the arithmetic.
    assert isinstance(tl.round_5p(103.0), int)


@pytest.mark.parametrize("pence", [0, 5, 100, 103, 1234.7])
def test_round_5p_result_is_always_a_multiple_of_five(pence):
    assert tl.round_5p(pence) % 5 == 0


# --------------------------------------------------------------------------- #
# parse_guests
# --------------------------------------------------------------------------- #
@pytest.mark.parametrize(
    "spec, want",
    [
        ("1", [(1, 0)]),
        ("2", [(2, 0)]),
        ("2+1", [(2, 1)]),
        ("2,2+2", [(2, 0), (2, 2)]),
        ("1,1,1", [(1, 0)] * 3),
        (" 2 + 1 ", [(2, 1)]),  # int() tolerates the inner spaces
        ("2,,2", [(2, 0), (2, 0)]),  # an empty room is skipped, not an error
        ("2+0", [(2, 0)]),
    ],
)
def test_parse_guests(spec, want):
    assert tl.parse_guests(spec) == want


def test_parse_guests_preserves_room_order():
    # Room order is the rooms[i][...] index in the Travelodge query, so it is
    # part of the request rather than cosmetic.
    assert tl.parse_guests("3+1,1,2+2") == [(3, 1), (1, 0), (2, 2)]


@pytest.mark.parametrize("spec", ["", "   ", ",", ",,"])
def test_parse_guests_rejects_an_empty_spec(spec):
    with pytest.raises(ValueError, match="empty"):
        tl.parse_guests(spec)


@pytest.mark.parametrize("spec", ["0", "2,0", "0+2"])
def test_parse_guests_rejects_a_room_with_no_adult(spec):
    with pytest.raises(ValueError, match="at least one adult"):
        tl.parse_guests(spec)


def test_parse_guests_rejects_negative_children():
    with pytest.raises(ValueError, match="children cannot be negative"):
        tl.parse_guests("2+-1")


@pytest.mark.parametrize("spec", ["two", "2+x", "2.5"])
def test_parse_guests_rejects_non_numeric_parts(spec):
    with pytest.raises(ValueError):
        tl.parse_guests(spec)


# --------------------------------------------------------------------------- #
# haversine_km
# --------------------------------------------------------------------------- #
def test_haversine_km_is_zero_for_the_same_point():
    assert tl.haversine_km((51.5031, -0.1132), (51.5031, -0.1132)) == 0.0


def test_haversine_km_is_symmetric():
    a, b = (51.5031, -0.1132), (51.5308, -0.1238)
    assert tl.haversine_km(a, b) == pytest.approx(tl.haversine_km(b, a))


def test_haversine_km_matches_a_known_london_distance():
    # Waterloo to King's Cross is a little over 3km as the crow flies.
    d = tl.haversine_km(tl.ORIGINS["waterloo"][1:], tl.ORIGINS["kings cross"][1:])
    assert d == pytest.approx(3.14, abs=0.15)


def test_haversine_km_one_degree_of_latitude_is_about_111km():
    # A degree of latitude is the same everywhere, so this pins the radius
    # constant independently of any London coordinate.
    assert tl.haversine_km((0.0, 0.0), (1.0, 0.0)) == pytest.approx(111.19, abs=0.05)


def test_haversine_km_a_degree_of_longitude_shrinks_with_latitude():
    # cos(lat) scaling is the part a sign error in the formula would break.
    equator = tl.haversine_km((0.0, 0.0), (0.0, 1.0))
    london = tl.haversine_km((51.5, 0.0), (51.5, 1.0))
    assert london == pytest.approx(equator * math.cos(math.radians(51.5)), rel=1e-3)


def test_haversine_km_handles_antipodes():
    half_circumference = math.pi * 6371.0
    assert tl.haversine_km((0.0, 0.0), (0.0, 180.0)) == pytest.approx(half_circumference, rel=1e-9)


def test_estimate_cycle_never_reports_zero_minutes():
    # A hotel on top of the origin must still read at least 1 min.
    minutes, km = tl.estimate_cycle((51.5031, -0.1132), (51.5031, -0.1132))
    assert minutes == 1
    assert km == 0.0


# --------------------------------------------------------------------------- #
# sort_rows
# --------------------------------------------------------------------------- #
def row(code, total, room_price):
    return {"code": code, "total": total, "room_price": room_price}


def test_sort_rows_orders_by_total_and_sorts_in_place():
    rows = [row("c", 90.0, 60.0), row("a", 50.0, 40.0), row("b", 70.0, 65.0)]
    assert tl.sort_rows(rows) is None  # mutates, returns nothing
    assert [r["code"] for r in rows] == ["a", "b", "c"]


def test_sort_rows_puts_unpriced_rows_last():
    # An unpriced row must not sort as if it were free, which is what the
    # `total is None` leading key prevents.
    rows = [row("unpriced", None, 10.0), row("priced", 99.0, 99.0)]
    tl.sort_rows(rows)
    assert [r["code"] for r in rows] == ["priced", "unpriced"]


def test_sort_rows_breaks_ties_on_room_price():
    rows = [row("dear-room", 100.0, 95.0), row("cheap-room", 100.0, 80.0)]
    tl.sort_rows(rows)
    assert [r["code"] for r in rows] == ["cheap-room", "dear-room"]


def test_sort_rows_orders_unpriced_rows_among_themselves_by_room_price():
    rows = [row("b", None, 70.0), row("a", None, 30.0)]
    tl.sort_rows(rows)
    assert [r["code"] for r in rows] == ["a", "b"]


def test_sort_rows_handles_a_zero_total():
    # `or 0` collapses a genuine 0.0 total onto the None fallback value, so a
    # free row must still sort ahead of an unpriced one.
    rows = [row("unpriced", None, 5.0), row("free", 0.0, 0.0)]
    tl.sort_rows(rows)
    assert [r["code"] for r in rows] == ["free", "unpriced"]


def test_sort_rows_is_stable_for_fully_equal_rows():
    rows = [row("first", 10.0, 10.0), row("second", 10.0, 10.0)]
    tl.sort_rows(rows)
    assert [r["code"] for r in rows] == ["first", "second"]


def test_sort_rows_on_an_empty_list():
    rows = []
    tl.sort_rows(rows)
    assert rows == []
