"""Read a rowing machine display from a photo with Claude.

This is the image half of the Go backend's admin-only POST /rowing. The Go
handler keeps everything that needs the database or the session cookie - the
admin gate, the EXIF date, the duplicate check, the sanity bounds, the insert -
and hands the image here only once it has decided the Claude call is worth
paying for. This module just turns pixels into three numbers.

Like `travelodge`, it never imports FastAPI; `main` is the HTTP skin.
"""

import functools
import logging
import os

import anthropic
from pydantic import BaseModel, Field, ValidationError

log = logging.getLogger(__name__)

MODEL = "claude-haiku-4-5"
# The answer is a three-key JSON object, so this is a ceiling, not a target.
MAX_TOKENS = 256
# Seconds per attempt. The SDK retries twice on top (429/5xx/connection), so
# the worst case is roughly three of these plus backoff; the Go caller's own
# timeout is set above that so it sees this module's error, not its own.
TIMEOUT_S = 30.0

# The keys named here are the JSON contract: RowingReading's field names below
# and ExtractedRowingData in backend/handlers/handle_rowing.go must match them.
PROMPT = """Look at this rowing machine display. Extract the total elapsed time and total distance.

Return ONLY a JSON object with these exact keys and numeric values:
- "timeMinutes": total minutes (e.g. 2:30 = 2)
- "timeSeconds": total seconds (e.g. 2:30 = 30)
- "distance": distance in meters as a number (e.g. 5000)

No text, no markdown, no explanation. Just the JSON object."""


class RowingReading(BaseModel):
    """What Claude read off the display.

    camelCase on purpose: these are the keys the prompt asks for and the Go
    handler unmarshals. A key the model left out defaults to 0, which the Go
    side's sanity bounds then reject - the same outcome as when Go parsed the
    model's output itself.
    """

    timeMinutes: int = Field(0, ge=0)
    timeSeconds: int = Field(0, ge=0)
    distance: int = Field(0, ge=0)


class ReadError(Exception):
    """The display could not be read; the message is safe to show the admin."""


class NotConfigured(ReadError):
    """CLAUDE_API_KEY is not set in this container."""


@functools.lru_cache(maxsize=1)
def _client() -> anthropic.Anthropic:
    """Build the Anthropic client on first use.

    Lazy so that a missing key only breaks this feature: the hotel finder must
    still boot without it. The key is passed explicitly because the stack's
    variable is CLAUDE_API_KEY, not the ANTHROPIC_API_KEY the SDK looks for.
    """
    key = os.environ.get("CLAUDE_API_KEY", "")
    if not key:
        raise NotConfigured("CLAUDE_API_KEY is not set")
    return anthropic.Anthropic(api_key=key, timeout=TIMEOUT_S)


def strip_markdown_fence(raw: str) -> str:
    """Remove a surrounding markdown code fence from a model's response.

    The prompt asks for bare JSON and the model fences it anyway often enough
    to matter. Mirrors services.StripMarkdownFence in the Go backend (still
    used by the email pipeline): "```json" is tried before a plain "```", and
    unfenced input comes back unchanged apart from surrounding whitespace.
    """
    raw = raw.strip()
    raw = raw.removeprefix("```json")
    raw = raw.removeprefix("```")
    raw = raw.removesuffix("```")
    return raw.strip()


def read_display(media_type: str, data: str) -> RowingReading:
    """Ask Claude for the elapsed time and distance shown in an image.

    Args:
        media_type: one of the image types the Anthropic API accepts; the
            caller is responsible for the allow-list.
        data: the image, base64-encoded (no data: prefix).

    Raises:
        NotConfigured: no API key.
        ReadError: the API call failed, or the reply was empty or not the JSON
            object that was asked for.
    """
    try:
        message = _client().messages.create(
            model=MODEL,
            max_tokens=MAX_TOKENS,
            messages=[
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "image",
                            "source": {"type": "base64", "media_type": media_type, "data": data},
                        },
                        {"type": "text", "text": PROMPT},
                    ],
                }
            ],
        )
    except anthropic.APIStatusError as e:
        log.warning("Claude API error %s: %s", e.status_code, e.message)
        raise ReadError("failed to process image") from e
    except anthropic.APIConnectionError as e:
        log.warning("Claude API unreachable: %s", e)
        raise ReadError("failed to process image") from e

    text = next((block.text for block in message.content if block.type == "text"), None)
    if text is None:
        raise ReadError("empty response from image processor")

    try:
        return RowingReading.model_validate_json(strip_markdown_fence(text))
    except ValidationError as e:
        log.warning("unparseable model output %r: %s", text, e)
        raise ReadError("failed to parse image data") from e
