# Imagen oficial de uv con Python 3.11 preinstalado y uv en PATH.
# Documentacion: https://docs.astral.sh/uv/guides/integration/docker/
FROM ghcr.io/astral-sh/uv:python3.11-bookworm-slim AS base

# Para que uv instale en /app/.venv y pytest/uvicorn vivan ahi
ENV UV_PROJECT_ENVIRONMENT=/app/.venv \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    UV_COMPILE_BYTECODE=1

WORKDIR /app

# Capa 1: resolver e instalar deps (cache-friendly; cambia solo si cambia el lock)
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-install-project --no-dev

# Capa 2: instalar el proyecto en si mismo (importable como 'app')
# El README se copia porque pyproject.toml declara `readme = "README.md"` y
# hatchling falla el build del paquete si el archivo no esta en el contexto.
COPY README.md ./README.md
COPY app ./app
COPY alembic ./alembic
COPY alembic.ini ./alembic.ini
RUN uv sync --frozen --no-dev

ENV PATH="/app/.venv/bin:$PATH"

EXPOSE 8000

# Healthcheck simple: hit /health cada 30s, timeout 5s, 3 reintentos.
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health').read()" || exit 1

# Por defecto, solo arrancamos uvicorn. Para correr migraciones antes del
# arranque en produccion, usar el override de command en docker-compose.yml.
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
