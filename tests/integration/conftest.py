"""Configuracion compartida de los tests de integracion.

Los tests corren con `uv run pytest` desde `fastAPI_SaaS/`, pero el secreto
JWT y las credenciales de la BD viven en el `.env` de la raiz del proyecto.
Sin cargarlo, `os.getenv("AUTH_JWT_SECRET")` devuelve "" y los tests que
verifican la firma del token fallan con "Signature verification failed"
porque firman la verificacion con la clave vacia.
"""
from __future__ import annotations

import os
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[2]
ENV_FILE = PROJECT_ROOT / ".env"


def _load_env_file() -> None:
    """Carga KEY=value del .env sin pisar lo que ya venga del entorno."""
    if not ENV_FILE.is_file():
        return
    for raw in ENV_FILE.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip().strip("\"'")
        # `setdefault`: el entorno real (CI, docker compose) tiene prioridad.
        os.environ.setdefault(key, value)


_load_env_file()
