"""Logging estructurado en JSON (#21).

Uvicorn configura sus propios loggers al arrancar, con formato de texto. En un
stack con varios servicios eso hace imposible correlacionar lineas: cada uno
inventa su formato y no hay campos comunes.

Este modulo instala un formatter JSON en el logger raiz y hace que los loggers
de uvicorn propaguen a el, en vez de escribir con su handler propio. Se llama
al importar `app.main`, que es cuando uvicorn importa la app: despues de su
propia configuracion, asi que nuestra configuracion gana.

El formato se puede volver a texto con LOG_FORMAT=text, util para leer en local.
"""
import json
import logging
import os
import sys
from datetime import datetime, timezone

_RESERVED = set(logging.LogRecord("", 0, "", 0, "", (), None).__dict__) | {
    "message",
    "asctime",
}


class JsonFormatter(logging.Formatter):
    """Una linea JSON por registro, con los campos minimos para grep y agregacion."""

    def format(self, record: logging.LogRecord) -> str:
        # Sufijo Z y no +00:00: identico al `ts` que emite el Backend
        # (requestlogger), para que el agregador pueda comparar lineas de los
        # dos servicios sin normalizar antes.
        ts = datetime.fromtimestamp(record.created, tz=timezone.utc).isoformat().replace("+00:00", "Z")
        payload = {
            "level": record.levelname.lower(),
            "ts": ts,
            "service": "saas-api",
            "logger": record.name,
            "msg": record.getMessage(),
        }

        # Campos pasados por `extra=` en el logger, para contexto sin romper el
        # formato (e.g. logger.info("pedido creado", extra={"tenant_id": 7})).
        for key, value in record.__dict__.items():
            if key not in _RESERVED and key not in payload and not key.startswith("_"):
                payload[key] = value

        if record.exc_info:
            payload["exc"] = self.formatException(record.exc_info)

        return json.dumps(payload, ensure_ascii=False, default=str)


def configure_logging() -> None:
    """Instala el handler en el root y hace que uvicorn propague a el.

    Idempotente: llamarla dos veces no duplica handlers (reemplaza la lista).
    """
    use_json = os.getenv("LOG_FORMAT", "json").lower() != "text"
    level = os.getenv("LOG_LEVEL", "INFO").upper()

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(
        JsonFormatter() if use_json else logging.Formatter("%(levelname)s %(name)s %(message)s")
    )

    root = logging.getLogger()
    root.handlers[:] = [handler]
    root.setLevel(level)

    # uvicorn instala handlers propios con su formato. Se limpian y se deja
    # propagate=True para que sus lineas salgan por el handler de arriba.
    for name in ("uvicorn", "uvicorn.error", "uvicorn.access"):
        uvicorn_logger = logging.getLogger(name)
        uvicorn_logger.handlers[:] = []
        uvicorn_logger.propagate = True
