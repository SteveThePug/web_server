"""Tests for the job posting reader and its internal-only route.

As in test_rowing, no Claude call happens: `jobs._client` is replaced with a
fake, so what is under test is the request that is built, how the reply is
parsed, how failures map to status codes, and the guard that keeps the route
off the public internet.
"""

from types import SimpleNamespace

import anthropic
import httpx2 as httpx
import pytest
from fastapi.testclient import TestClient

from app import jobs, main

BODY = {"url": "https://jobs.example/1", "title": "Engineer | Acme", "text": "Acme is hiring an Engineer in London."}
REPLY = '{"jobTitle": "Engineer", "company": "Acme", "location": "London", "notes": "£70k, hybrid"}'


def text_reply(text):
    return SimpleNamespace(content=[SimpleNamespace(type="text", text=text)])


@pytest.fixture
def fake_claude(monkeypatch):
    """Stand in for the Anthropic client; see the fixture of the same name in test_rowing."""
    state = SimpleNamespace(calls=[], reply=text_reply(REPLY))

    def create(**kwargs):
        state.calls.append(kwargs)
        if isinstance(state.reply, Exception):
            raise state.reply
        return state.reply

    client = SimpleNamespace(messages=SimpleNamespace(create=create))
    monkeypatch.setattr(jobs, "_client", lambda: client)
    return state


@pytest.fixture
def http():
    return TestClient(main.app)


# --- extract_posting ------------------------------------------------------


def test_extract_sends_page_as_delimited_data(fake_claude):
    posting = jobs.extract_posting(BODY["url"], BODY["title"], BODY["text"])

    assert posting.model_dump() == {
        "jobTitle": "Engineer",
        "company": "Acme",
        "location": "London",
        "notes": "£70k, hybrid",
    }
    (call,) = fake_claude.calls
    assert call["model"] == jobs.MODEL
    # The instructions travel in the system prompt and the page in the user
    # turn, inside the <page> tags the system prompt tells the model to treat
    # as data.
    assert call["system"] == jobs.SYSTEM
    (message,) = call["messages"]
    assert message["role"] == "user"
    assert message["content"] == (
        "<url>https://jobs.example/1</url>\n<title>Engineer | Acme</title>\n"
        "<page>\nAcme is hiring an Engineer in London.\n</page>"
    )


def test_extract_accepts_fenced_json(fake_claude):
    fake_claude.reply = text_reply(f"```json\n{REPLY}\n```")
    assert jobs.extract_posting("", "", "x").company == "Acme"


@pytest.mark.parametrize(
    "reply, want",
    [
        # Not a posting: the popup is filled in by hand rather than erroring.
        ('{"jobTitle": "", "company": "", "location": null, "notes": null}', ("", "", None, None)),
        # Nulls where strings were asked for, and keys left out altogether.
        ('{"jobTitle": null, "company": null}', ("", "", None, None)),
        ("{}", ("", "", None, None)),
        # Whitespace is trimmed, and a blank optional field becomes null.
        ('{"jobTitle": " Engineer ", "company": "Acme\\n", "location": "  ", "notes": ""}', ("Engineer", "Acme", None, None)),
    ],
)
def test_extract_normalises_what_the_model_returns(fake_claude, reply, want):
    fake_claude.reply = text_reply(reply)
    posting = jobs.extract_posting("", "", "x")
    assert (posting.jobTitle, posting.company, posting.location, posting.notes) == want


@pytest.mark.parametrize(
    "reply, message",
    [
        (SimpleNamespace(content=[]), "empty response from the posting reader"),
        (text_reply("This looks like a great role!"), "failed to parse the posting"),
        (text_reply('{"jobTitle": ["Engineer"], "company": "Acme"}'), "failed to parse the posting"),
        (
            anthropic.APIConnectionError(request=httpx.Request("POST", "https://api.anthropic.com/v1/messages")),
            "failed to read the posting",
        ),
    ],
)
def test_extract_failures(fake_claude, reply, message):
    fake_claude.reply = reply
    with pytest.raises(jobs.ReadError, match=message):
        jobs.extract_posting("", "", "x")


def test_missing_api_key_is_not_configured(monkeypatch):
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    jobs._client.cache_clear()
    with pytest.raises(jobs.NotConfigured):
        jobs.extract_posting("", "", "x")


# --- the route ------------------------------------------------------------


def test_route_returns_posting(fake_claude, http):
    res = http.post("/internal/jobs/extract", json=BODY)
    assert res.status_code == 200
    assert res.json() == {"jobTitle": "Engineer", "company": "Acme", "location": "London", "notes": "£70k, hybrid"}


def test_route_needs_only_the_text(fake_claude, http):
    assert http.post("/internal/jobs/extract", json={"text": "x"}).status_code == 200


@pytest.mark.parametrize("header", ["X-Real-IP", "X-Forwarded-For"])
def test_route_refuses_requests_proxied_by_nginx(fake_claude, http, header):
    """nginx sets these on everything it forwards; the Go backend sets neither."""
    res = http.post("/internal/jobs/extract", json=BODY, headers={header: "203.0.113.9"})
    assert res.status_code == 404
    assert fake_claude.calls == []


@pytest.mark.parametrize("text", ["", "x" * (jobs.MAX_TEXT_CHARS + 1)])
def test_route_rejects_empty_or_oversized_text(fake_claude, http, text):
    res = http.post("/internal/jobs/extract", json={**BODY, "text": text})
    assert res.status_code == 422
    assert fake_claude.calls == []


def test_route_maps_read_error_to_502(fake_claude, http):
    fake_claude.reply = text_reply("no JSON here")
    res = http.post("/internal/jobs/extract", json=BODY)
    assert res.status_code == 502
    assert res.json() == {"detail": "failed to parse the posting"}


def test_route_maps_missing_key_to_503(monkeypatch, http):
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    jobs._client.cache_clear()
    res = http.post("/internal/jobs/extract", json=BODY)
    assert res.status_code == 503
