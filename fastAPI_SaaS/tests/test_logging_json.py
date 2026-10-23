"""Tests del formatter JSON de logging (#21)."""
import json
import logging

from app.core.logging import JsonFormatter


def _record(msg: str = "hola", level: int = logging.INFO, **extra) -> logging.LogRecord:
    record = logging.LogRecord(
        name="app.test", level=level, pathname=__file__, lineno=1, msg=msg, args=(), exc_info=None
    )
    for key, value in extra.items():
        setattr(record, key, value)
    return record


def test_json_formatter_emite_una_linea_json_valida():
    line = JsonFormatter().format(_record("pedido creado"))

    assert "\n" not in line, "cada registro debe ocupar una sola linea"
    payload = json.loads(line)
    assert payload["msg"] == "pedido creado"
    assert payload["level"] == "info"
    assert payload["service"] == "saas-api"
    assert payload["logger"] == "app.test"
    # Sufijo Z y no +00:00: es el formato exacto que emite el Backend
    # (requestlogger). El agregador de logs compara el campo `ts` crudo, asi
    # que los dos servicios tienen que emitirlo identico.
    assert payload["ts"].endswith("Z"), "el timestamp debe ser UTC explicito"


def test_json_formatter_incluye_campos_extra():
    line = JsonFormatter().format(_record("consulta", tenant_id=7))
    payload = json.loads(line)

    assert payload["tenant_id"] == 7


def test_json_formatter_incluye_traceback_en_errores():
    try:
        raise ValueError("boom")
    except ValueError:
        import sys

        record = _record("fallo", level=logging.ERROR)
        record.exc_info = sys.exc_info()

    payload = json.loads(JsonFormatter().format(record))

    assert payload["level"] == "error"
    assert "ValueError: boom" in payload["exc"]
