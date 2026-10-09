"""Tests for the rowing display reader and its internal-only route.

No Claude call happens here: `rowing._client` is replaced with a fake whose
`messages.create` returns a canned reply, so what is under test is everything
around the call - the request that is built, how the reply is parsed, how
failures map to status codes, and the guard that keeps the route off the public
internet.
"""

from types import SimpleNamespace

import anthropic
import httpx2 as httpx
import pytest
from fastapi.testclient import TestClient

from app import main, rowing

BODY = {"media_type": "image/jpeg", "data": "aGVsbG8="}


def text_reply(text):
    return SimpleNamespace(content=[SimpleNamespace(type="text", text=text)])


@pytest.fixture
def fake_claude(monkeypatch):
    """Stand in for the Anthropic client; returns the list of calls it received.

    Set `fake_claude.reply` to a message-like object, or to an exception
    instance to have the call raise it.
    """
    state = SimpleNamespace(calls=[], reply=text_reply('{"timeMinutes": 20, "timeSeconds": 5, "distance": 5000}'))

    def create(**kwargs):
        state.calls.append(kwargs)
        if isinstance(state.reply, Exception):
            raise state.reply
        return state.reply

    client = SimpleNamespace(messages=SimpleNamespace(create=create))
    monkeypatch.setattr(rowing, "_client", lambda: client)
    return state


@pytest.fixture
def http():
    return TestClient(main.app)


# --- fence stripping (must agree with services.StripMarkdownFence in Go) ---


@pytest.mark.parametrize(
    "raw, want",
    [
        ('{"a": 1}', '{"a": 1}'),
        ('  {"a": 1}\n', '{"a": 1}'),
        ('```json\n{"a": 1}\n```', '{"a": 1}'),
        ('```\n{"a": 1}\n```', '{"a": 1}'),
        ('```json\n{"a": 1}', '{"a": 1}'),
    ],
)
def test_strip_markdown_fence(raw, want):
    assert rowing.strip_markdown_fence(raw) == want


# --- read_display ---------------------------------------------------------


def test_read_display_sends_image_then_prompt(fake_claude):
    reading = rowing.read_display("image/png", "QUJD")

    assert (reading.timeMinutes, reading.timeSeconds, reading.distance) == (20, 5, 5000)
    (call,) = fake_claude.calls
    assert call["model"] == rowing.MODEL
    image, prompt = call["messages"][0]["content"]
    assert image == {"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "QUJD"}}
    assert prompt == {"type": "text", "text": rowing.PROMPT}


def test_read_display_accepts_fenced_json(fake_claude):
    fake_claude.reply = text_reply('```json\n{"timeMinutes": 1, "timeSeconds": 2, "distance": 300}\n```')
    assert rowing.read_display("image/jpeg", "QUJD").distance == 300


def test_read_display_missing_key_defaults_to_zero(fake_claude):
    """The Go side turns a zero distance into its own 400; it must not fail here."""
    fake_claude.reply = text_reply('{"timeMinutes": 1, "timeSeconds": 2}')
    assert rowing.read_display("image/jpeg", "QUJD").distance == 0


@pytest.mark.parametrize(
    "reply, message",
    [
        (SimpleNamespace(content=[]), "empty response from image processor"),
        (text_reply("I can't read that display."), "failed to parse image data"),
        (text_reply('{"timeMinutes": -1, "timeSeconds": 2, "distance": 300}'), "failed to parse image data"),
        (
            anthropic.APIConnectionError(request=httpx.Request("POST", "https://api.anthropic.com/v1/messages")),
            "failed to process image",
        ),
    ],
)
def test_read_display_failures(fake_claude, reply, message):
    fake_claude.reply = reply
    with pytest.raises(rowing.ReadError, match=message):
        rowing.read_display("image/jpeg", "QUJD")


def test_missing_api_key_is_not_configured(monkeypatch):
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    rowing._client.cache_clear()
    with pytest.raises(rowing.NotConfigured):
        rowing.read_display("image/jpeg", "QUJD")


# --- the route ------------------------------------------------------------


def test_route_returns_reading(fake_claude, http):
    res = http.post("/internal/rowing/read", json=BODY)
    assert res.status_code == 200
    assert res.json() == {"timeMinutes": 20, "timeSeconds": 5, "distance": 5000}


@pytest.mark.parametrize("header", ["X-Real-IP", "X-Forwarded-For"])
def test_route_refuses_requests_proxied_by_nginx(fake_claude, http, header):
    """nginx sets these on everything it forwards; the Go backend sets neither."""
    res = http.post("/internal/rowing/read", json=BODY, headers={header: "203.0.113.9"})
    assert res.status_code == 404
    assert fake_claude.calls == []


def test_route_rejects_unsupported_media_type(fake_claude, http):
    res = http.post("/internal/rowing/read", json={**BODY, "media_type": "image/heic"})
    assert res.status_code == 422
    assert fake_claude.calls == []


def test_route_maps_read_error_to_502(fake_claude, http):
    fake_claude.reply = text_reply("no numbers here")
    res = http.post("/internal/rowing/read", json=BODY)
    assert res.status_code == 502
    assert res.json() == {"detail": "failed to parse image data"}


def test_route_maps_missing_key_to_503(monkeypatch, http):
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    rowing._client.cache_clear()
    res = http.post("/internal/rowing/read", json=BODY)
    assert res.status_code == 503
