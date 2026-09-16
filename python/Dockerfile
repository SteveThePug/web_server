# =============================================================================
# FastAPI side-service image.
#
# requirements.txt is copied and installed before the app source so the
# dependency layer is cached across code changes. --no-cache-dir keeps pip's
# download cache out of the image (this one is a plain layer, not a BuildKit
# cache mount, because the dependency set is small).
#
# PYTHONDONTWRITEBYTECODE stops .pyc files being written into the image and the
# dev bind mount; PYTHONUNBUFFERED makes logs appear in `docker logs`
# immediately instead of sitting in a buffer — both are standard for
# containerised Python and the second one matters for debugging.
#
# ROOT_PATH=/py: nginx proxies /py/ and STRIPS the prefix, so the app's routes
# are written relative to /. --root-path tells FastAPI it is nonetheless
# mounted under a prefix, so the OpenAPI schema and Swagger UI emit correct
# /py/... URLs. Without it, /py/docs loads but cannot call anything.
#
# The CMD is a `sh -c` form on purpose: it needs the shell to expand
# ${PYTHON_PORT} and ${ROOT_PATH} at runtime, which the exec form cannot do.
# Both have defaults here so the image runs standalone; docker compose normally
# supplies them.
#
# EXPOSE 8000 is documentation only — the port is not published; nginx reaches
# it over app-network.
# =============================================================================
FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app ./app

EXPOSE 8000

# PYTHON_PORT and ROOT_PATH are supplied by docker compose (defaults below).
CMD ["sh", "-c", "uvicorn app.main:app --host 0.0.0.0 --port ${PYTHON_PORT:-8000} --root-path ${ROOT_PATH:-/py}"]
