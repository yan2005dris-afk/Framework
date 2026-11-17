"""Cliente Singleton asíncrono para Valkey.

Maneja una única instancia del pool de conexiones para FastAPI,
reutilizable en tareas en segundo plano (Workers) o endpoints.
"""
from __future__ import annotations

import logging
from typing import Optional

from redis.asyncio import Redis

from app.shared.config import settings

logger = logging.getLogger("uvicorn.error")

_valkey_client: Optional[Redis] = None


async def get_valkey_client() -> Redis:
    """Devuelve la instancia Singleton del cliente asíncrono de Valkey."""
    global _valkey_client
    if _valkey_client is None:
        _valkey_client = Redis(
            host=settings.valkey_host,
            port=settings.valkey_port,
            password=settings.valkey_password or None,
            decode_responses=True,
            socket_timeout=5.0,
            socket_connect_timeout=5.0,
        )
        try:
            await _valkey_client.ping()
            logger.info("[valkey] Conexion Singleton exitosa a Valkey (%s:%s)", settings.valkey_host, settings.valkey_port)
        except Exception as e:
            logger.warning("[valkey] No se pudo conectar a Valkey (%s:%s): %s", settings.valkey_host, settings.valkey_port, e)

    return _valkey_client


async def close_valkey_client() -> None:
    """Cierra ordenadamente el cliente Singleton de Valkey."""
    global _valkey_client
    if _valkey_client is not None:
        await _valkey_client.aclose()
        _valkey_client = None
        logger.info("[valkey] Conexion Singleton a Valkey cerrada")
