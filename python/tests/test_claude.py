"""Tests for `ask_json`, the one Claude call the three callers share.

test_rowing, test_jobs and test_email_classify each cover their own prompt,
model and messages through this function; what is pinned here is the function's
own contract - what it sends, which failure produces which of the caller's
three messages, and where it gets its client from.
"""

from types import SimpleNamespace

import anthropic
import httpx2 as httpx
import pytest
from pydantic import BaseModel

from app import claude

REQUEST = httpx.Request("POST", "https://api.anthropic.com/v1/messages")
MESSAGES = [{"role": "user", "content": "how many?"}]
ERRORS = {"failed": "call failed", "empty": "nothing came back", "unparseable": "could not parse"}


class Count(BaseModel):
    n: int = 0


def text_reply(text):
    return SimpleNamespace(content=[SimpleNamespace(type="text", text=text)])


def fake_client(reply):
    """A client whose `messages.create` records its kwargs and returns or raises `reply`."""
    calls = []

    def create(**kwargs):
        calls.append(kwargs)
        if isinstance(reply, Exception):
            raise reply
        return reply

    return SimpleNamespace(messages=SimpleNamespace(create=create), calls=calls)


def ask(reply, **kwargs):
    client = fake_client(reply)
    result = claude.ask_json(Count, max_tokens=64, messages=MESSAGES, get_client=lambda: client, **ERRORS, **kwargs)
    return result, client.calls


def test_sends_the_request_and_validates_the_reply():
    result, (call,) = ask(text_reply('{"n": 3}'))

    assert result == Count(n=3)
    # No system prompt was given, so the key is absent rather than None: the
    # SDK would send a null the API rejects.
    assert call == {"model": claude.MODEL, "max_tokens": 64, "messages": MESSAGES}


def test_passes_a_system_prompt_through():
    _, (call,) = ask(text_reply("{}"), system="count things")
    assert call["system"] == "count things"


def test_accepts_fenced_json():
    result, _ = ask(text_reply('```json\n{"n": 3}\n```'))
    assert result.n == 3


def test_reads_the_first_text_block():
    """Anything that is not text (a thinking block, say) is stepped over."""
    reply = SimpleNamespace(
        content=[SimpleNamespace(type="thinking"), SimpleNamespace(type="text", text='{"n": 3}')]
    )
    result, _ = ask(reply)
    assert result.n == 3


@pytest.mark.parametrize(
    "reply, message",
    [
        (anthropic.APIConnectionError(request=REQUEST), "call failed"),
        (anthropic.APITimeoutError(request=REQUEST), "call failed"),
        (
            anthropic.APIStatusError("overloaded", response=httpx.Response(529, request=REQUEST), body=None),
            "call failed",
        ),
        (SimpleNamespace(content=[]), "nothing came back"),
        (SimpleNamespace(content=[SimpleNamespace(type="thinking")]), "nothing came back"),
        (text_reply("three, I think"), "could not parse"),
        # Valid JSON that the model class rejects is the same failure.
        (text_reply('{"n": "lots"}'), "could not parse"),
        (text_reply("[3]"), "could not parse"),
    ],
)
def test_each_failure_raises_the_callers_message(reply, message):
    with pytest.raises(claude.ReadError) as raised:
        ask(reply)
    # Exactly the caller's text: the Go backend relays it verbatim, so nothing
    # from the API error or the model's reply may be appended to it.
    assert str(raised.value) == message
    assert not isinstance(raised.value, claude.NotConfigured)


def test_missing_api_key_is_not_configured_not_a_failed_call(monkeypatch):
    """`main` tells the two apart by type to choose between 503 and 502."""
    monkeypatch.delenv("CLAUDE_API_KEY", raising=False)
    claude.client.cache_clear()
    with pytest.raises(claude.NotConfigured):
        claude.ask_json(Count, max_tokens=64, messages=MESSAGES, **ERRORS)


def test_uses_the_shared_client_when_none_is_given(monkeypatch):
    client = fake_client(text_reply('{"n": 3}'))
    monkeypatch.setattr(claude, "client", lambda: client)

    assert claude.ask_json(Count, max_tokens=64, messages=MESSAGES, **ERRORS).n == 3
    assert len(client.calls) == 1
