"""The Anthropic client and the one call every Claude caller here makes.

`rowing` (a photo in, three numbers out) is a call of a general shape: one
prompt asking for bare JSON, one reply to parse. That whole sequence is
`ask_json` below, so a caller is a prompt, a pydantic model and three error
strings, and another caller would get the same handling around the call. Also
here: the lazily built client and the two exception types `main` maps to 502
and 503.

Like the modules that use it, this never imports FastAPI.
"""

import functools
import logging
import os
from typing import Callable, Optional, TypeVar

import anthropic
from pydantic import BaseModel, ValidationError

log = logging.getLogger(__name__)

T = TypeVar("T", bound=BaseModel)

# Which model every call uses. CLAUDE_MODEL is set from the admin panel's
# Environment tab (it is in backend/services/env_registry.json, whose `options`
# list is what the panel offers); changing it there recreates this container,
# which is why reading it once at import is enough. Blank means the default.
#
# The registry's options are the ids this module has been written for. Another
# id is passed through as it is, treated as a model that thinks (see below);
# if the API does not know it, every call fails with its `failed` message and
# the reason is in this service's log.
DEFAULT_MODEL = "claude-sonnet-5-5"
MODEL = os.environ.get("CLAUDE_MODEL", "").strip() or DEFAULT_MODEL

# Models that take no `output_config.effort` and do not think unless asked to.
# Everything newer thinks by default, and that changes two things in the
# request - see _model_options.
_NO_EFFORT_PREFIXES = ("claude-haiku-4-5",)

# Thinking is billed against `max_tokens` like the reply is, and the callers'
# ceilings are sized for the reply alone (a few hundred tokens of JSON). On a
# model that thinks, this is added on top so that a posting worth a moment's
# thought cannot use the whole allowance and leave no room for the answer.
THINKING_HEADROOM = 4000

# Every call here is extraction or classification: read something, return a
# small JSON object. "low" is the level meant for that - the model skips
# thinking on most simple requests - and it is set explicitly because the
# default differs between models.
EFFORT = "low"

# Seconds per attempt. The SDK retries twice on top (429/5xx/connection), so
# the worst case is roughly three of these plus backoff; the Go callers' own
# timeout is set above that so they see this side's error, not their own.
TIMEOUT_S = 30.0


class ReadError(Exception):
    """Claude could not produce the answer; the message is safe to show the admin."""


class NotConfigured(ReadError):
    """CLAUDE_API_KEY is not set in this container."""


@functools.lru_cache(maxsize=1)
def client() -> anthropic.Anthropic:
    """Build the Anthropic client on first use.

    Lazy so that a missing key only breaks the Claude features: the hotel
    finder must still boot without it. The key is passed explicitly because the
    stack's variable is CLAUDE_API_KEY, not the ANTHROPIC_API_KEY the SDK looks
    for.
    """
    key = os.environ.get("CLAUDE_API_KEY", "")
    if not key:
        raise NotConfigured("CLAUDE_API_KEY is not set")
    return anthropic.Anthropic(api_key=key, timeout=TIMEOUT_S)


def strip_markdown_fence(raw: str) -> str:
    """Remove a surrounding markdown code fence from a model's response.

    The prompts ask for bare JSON and the model fences it anyway often enough
    to matter. "```json" is tried before a plain "```", and unfenced input
    comes back unchanged apart from surrounding whitespace. This is the only
    copy: the Go backend had its own while it still called Claude itself, and
    every Claude call now goes through this service.
    """
    raw = raw.strip()
    raw = raw.removeprefix("```json")
    raw = raw.removeprefix("```")
    raw = raw.removesuffix("```")
    return raw.strip()


def _model_options(model: str, max_tokens: int) -> dict:
    """The parts of the request that depend on which model is configured.

    `max_tokens` is the caller's ceiling on the reply. A model that thinks
    gets THINKING_HEADROOM on top and an explicit effort level; Haiku 4.5
    gets neither, because it rejects `output_config.effort` and does not think
    unless a budget is requested, which nothing here does.
    """
    if model.startswith(_NO_EFFORT_PREFIXES):
        return {"max_tokens": max_tokens}
    return {"max_tokens": max_tokens + THINKING_HEADROOM, "output_config": {"effort": EFFORT}}


def ask_json(
    model_cls: type[T],
    *,
    max_tokens: int,
    messages: list,
    system: Optional[str] = None,
    failed: str,
    empty: str,
    unparseable: str,
    get_client: Optional[Callable[[], anthropic.Anthropic]] = None,
) -> T:
    """Send one prompt that asks for a bare JSON object and validate the reply.

    The three message arguments are what the caller's user ends up reading:
    the Go backend relays a ReadError's text verbatim, so each caller words
    its own and nothing from the API or the model's reply leaks into it. The
    detail goes to the log instead.

    Args:
        model_cls: the pydantic model the reply must validate as; its field
            names are the keys the caller's prompt asks for.
        max_tokens: ceiling on the reply; see _model_options for what is
            actually sent.
        messages: the `messages` list, passed to the API as is.
        system: the system prompt, left out of the request when None.
        failed: ReadError text when the API call itself failed.
        empty: ReadError text when the reply has no text block.
        unparseable: ReadError text when the reply is not JSON, or is JSON
            that `model_cls` rejects.
        get_client: returns the Anthropic client. Callers pass their own
            module-level `_client` so that a test can replace the client for
            one module; resolved here, per call, when not given.

    Raises:
        NotConfigured: no API key.
        ReadError: with one of the three messages above.
    """
    request = {"model": MODEL, "messages": messages, **_model_options(MODEL, max_tokens)}
    if system is not None:
        request["system"] = system

    # Outside the try: NotConfigured is not an API failure and must reach
    # `main` as itself, to become the 503.
    api = (get_client or client)()
    try:
        message = api.messages.create(**request)
    except anthropic.APIStatusError as e:
        log.warning("Claude API error %s: %s", e.status_code, e.message)
        raise ReadError(failed) from e
    except anthropic.APIConnectionError as e:
        log.warning("Claude API unreachable: %s", e)
        raise ReadError(failed) from e

    # A refusal is an ordinary 200 whose content is not the answer. Say so in
    # the log rather than letting it surface as an "empty" or "unparseable"
    # reply, which would point at the prompt instead of at the classifier.
    if getattr(message, "stop_reason", None) == "refusal":
        log.warning("Claude declined the request: %s", getattr(message, "stop_details", None))
        raise ReadError(failed)

    # By type, not position: on a model that thinks, the first block can be a
    # (possibly empty) thinking block.
    text = next((block.text for block in message.content if block.type == "text"), None)
    if text is None:
        if getattr(message, "stop_reason", None) == "max_tokens":
            log.warning("Claude ran out of tokens before answering (max_tokens=%s)", request["max_tokens"])
        raise ReadError(empty)

    try:
        return model_cls.model_validate_json(strip_markdown_fence(text))
    except ValidationError as e:
        log.warning("unparseable model output %r: %s", text, e)
        raise ReadError(unparseable) from e
