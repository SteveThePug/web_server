"""Make `python/` importable so `app.travelodge` resolves.

The Docker image runs uvicorn from /python with the app as a package, but
pytest is run from wherever the developer happens to be, so the package root is
put on sys.path explicitly rather than relying on rootdir inference.
"""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
