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
