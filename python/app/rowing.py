"""Read a rowing machine display from a photo with Claude.

This is the image half of the Go backend's admin-only POST /rowing. The Go
handler keeps everything that needs the database or the session cookie - the
admin gate, the EXIF date, the duplicate check, the sanity bounds, the insert -
and hands the image here only once it has decided the Claude call is worth
paying for. This module just turns pixels into three numbers.

Like `travelodge`, it never imports FastAPI; `main` is the HTTP skin.
"""

import logging

import anthropic
from pydantic import BaseModel, Field, ValidationError

# Re-exported under the names this module has always had: `main` and the tests
# reach the client and the exceptions through `rowing`.
from .claude import MODEL, NotConfigured, ReadError, strip_markdown_fence
from .claude import client as _client

log = logging.getLogger(__name__)

# The answer is a three-key JSON object, so this is a ceiling, not a target.
MAX_TOKENS = 256

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
