"""Tests for the job lead scorer and its internal-only route.

As in test_jobs, no Claude call happens: `job_score._client` is replaced with
a fake, so what is under test is the request that is built, how the reply is
parsed, how failures map to status codes, and the guard that keeps the route
off the public internet.
"""

from types import SimpleNamespace

import anthropic
import httpx2 as httpx
import pytest
from fastapi.testclient import TestClient

from app import job_score, main

BODY = {
    "profile": "Five years of Go. Want a backend role in London.",
    "jobTitle": "Backend Engineer",
    "company": "Acme",
    "location": "London",
    "description": "We build payments in Go.",
}
REPLY = '{"score": 82, "reasons": "Your Go experience matches; no payments background."}'


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
    monkeypatch.setattr(job_score, "_client", lambda: client)
    return state


@pytest.fixture
def http():
    return TestClient(main.app)


def score(**overrides):
    args = {**BODY, **overrides}
    return job_score.score_posting(args["profile"], args["jobTitle"], args["company"], args["location"], args["description"])


# --- score_posting --------------------------------------------------------


def test_score_sends_profile_and_posting_as_delimited_data(fake_claude):
    result = score()

    assert result.model_dump() == {"score": 82, "reasons": "Your Go experience matches; no payments background."}
    (call,) = fake_claude.calls
    assert call["model"] == job_score.MODEL
    # The instructions travel in the system prompt; the profile and the
    # third-party posting travel in the user turn, inside the tags the system
    # prompt tells the model to treat as data.
    assert call["system"] == job_score.SYSTEM
    (message,) = call["messages"]
    assert message["role"] == "user"
    assert message["content"] == (
        "<profile>\nFive years of Go. Want a backend role in London.\n</profile>\n"
        "<posting>\n<title>Backend Engineer</title>\n<company>Acme</company>\n"
        "<location>London</location>\n<description>\nWe build payments in Go.\n</description>\n</posting>"
    )


def test_score_accepts_fenced_json(fake_claude):
    fake_claude.reply = text_reply(f"```json\n{REPLY}\n```")
    assert score().score == 82


@pytest.mark.parametrize(
    "reply, want",
    [
        ('{"score": 82.0, "reasons": " ok "}', (82, "ok")),
        ('{"score": "40"}', (40, "")),
        ('{"score": 0, "reasons": null}', (0, "")),
        ('{"score": 100, "reasons": "Perfect"}', (100, "Perfect")),
    ],
)
def test_score_normalises_what_the_model_returns(fake_claude, reply, want):
    fake_claude.reply = text_reply(reply)
    result = score()
    assert (result.score, result.reasons) == want


@pytest.mark.parametrize(
    "reply, message",
    [
        (SimpleNamespace(content=[]), "empty response from the posting scorer"),
        (text_reply("This looks like a great role!"), "failed to parse the score"),
        # No usable number is a failure, never a score of zero.
        (text_reply('{"reasons": "Good fit"}'), "failed to parse the score"),
        (text_reply('{"score": null, "reasons": "Good fit"}'), "failed to parse the score"),
        (text_reply('{"score": "high"}'), "failed to parse the score"),
        (text_reply('{"score": 140}'), "failed to parse the score"),
        (text_reply('{"score": -5}'), "failed to parse the score"),
        (
            anthropic.APIConnectionError(request=httpx.Request("POST", "https://api.anthropic.com/v1/messages")),
            "failed to score the posting",
        ),
    ],
)
def test_score_failures(fake_claude, reply, message):
    fake_claude.reply = reply
    with pytest.raises(job_score.ReadError, match=message):
        score()


def test_missing_api_key_is_not_configured(monkeypatch):
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    job_score._client.cache_clear()
    with pytest.raises(job_score.NotConfigured):
        score()


# --- the route ------------------------------------------------------------


def test_route_returns_score(fake_claude, http):
    res = http.post("/internal/jobs/score", json=BODY)
    assert res.status_code == 200
    assert res.json() == {"score": 82, "reasons": "Your Go experience matches; no payments background."}


def test_route_needs_only_profile_and_title(fake_claude, http):
    assert http.post("/internal/jobs/score", json={"profile": "p", "jobTitle": "Engineer"}).status_code == 200


@pytest.mark.parametrize("header", ["X-Real-IP", "X-Forwarded-For"])
def test_route_refuses_requests_proxied_by_nginx(fake_claude, http, header):
    """nginx sets these on everything it forwards; the Go backend sets neither."""
    res = http.post("/internal/jobs/score", json=BODY, headers={header: "203.0.113.9"})
    assert res.status_code == 404
    assert fake_claude.calls == []


@pytest.mark.parametrize(
    "body",
    [
        {**BODY, "profile": ""},
        {**BODY, "jobTitle": ""},
        {**BODY, "description": "x" * (job_score.MAX_DESCRIPTION_CHARS + 1)},
        {**BODY, "profile": "x" * (job_score.MAX_PROFILE_CHARS + 1)},
    ],
)
def test_route_validates_the_body(fake_claude, http, body):
    assert http.post("/internal/jobs/score", json=body).status_code == 422
    assert fake_claude.calls == []


def test_route_maps_failures_to_status_codes(fake_claude, http, monkeypatch):
    fake_claude.reply = text_reply("nope")
    res = http.post("/internal/jobs/score", json=BODY)
    assert res.status_code == 502
    assert res.json() == {"detail": "failed to parse the score"}

    def not_configured():
        raise job_score.NotConfigured("CLAUDE_API_KEY is not set")

    monkeypatch.setattr(job_score, "_client", not_configured)
    assert http.post("/internal/jobs/score", json=BODY).status_code == 503
