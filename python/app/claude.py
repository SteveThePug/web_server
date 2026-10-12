"""The Anthropic client and the reply handling every Claude caller here shares.

`rowing` (a photo in, three numbers out) and `jobs` (page text in, a few
fields out) are the same shape of call: one prompt asking for bare JSON, one
reply to parse. What they have in common lives here so the two cannot drift -
the lazily built client, the fence stripping, and the two exception types
`main` maps to 502 and 503.

Like the modules that use it, this never imports FastAPI.
"""

import functools
import os

import anthropic

MODEL = "claude-haiku-4-5"
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
    to matter. Mirrors services.StripMarkdownFence in the Go backend (still
    used by the email pipeline): "```json" is tried before a plain "```", and
    unfenced input comes back unchanged apart from surrounding whitespace.
    """
    raw = raw.strip()
    raw = raw.removeprefix("```json")
    raw = raw.removeprefix("```")
    raw = raw.removesuffix("```")
    return raw.strip()
