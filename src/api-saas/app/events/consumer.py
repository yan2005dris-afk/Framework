"""Worker asíncrono para el procesamiento de eventos con Valkey Streams.

Escucha el stream `stream:pedidos` usando Consumer Groups (`XREADGROUP` + `XACK`),
lo que garantiza semántica at-least-once y previene pérdida de mensajes si el worker
falla antes de completar el procesamiento.
"""
from __future__ import annotations

import asyncio
import json
import logging
from decimal import Decimal
from typing import Any

from app.modules.pedidos.service import PedidoService
from app.shared.db.session import async_session_factory
from app.shared.valkey import get_valkey_client

logger = logging.getLogger("uvicorn.error")

STREAM_NAME = "stream:pedidos"
CONSUMER_GROUP = "saas-pedidos-group"
CONSUMER_NAME = "saas-worker-1"


class PedidosEventConsumer:
    """Consumidor de eventos de pedidos respaldado por Valkey Streams."""

    def __init__(self) -> None:
        self.running = False
        self.service = PedidoService()

    async def init_consumer_group(self, client: Any) -> None:
        """Crea el grupo de consumo si no existe."""
        try:
            await client.xgroup_create(STREAM_NAME, CONSUMER_GROUP, id="0", mkstream=True)
            logger.info("📦 Consumer Group '%s' creado para Stream '%s'", CONSUMER_GROUP, STREAM_NAME)
        except Exception as e:
            # Si el grupo ya existe (BUSYGROUP), se ignora el error
            if "BUSYGROUP" in str(e):
                pass
            else:
                logger.warning("Error inicializando Consumer Group: %s", e)

    async def start(self) -> None:
        """Bucle principal de consumo de eventos."""
        self.running = True
        client = await get_valkey_client()
        await self.init_consumer_group(client)

        logger.info("🚀 PedidosEventConsumer iniciado en stream: %s", STREAM_NAME)

        while self.running:
            try:
                # Lee mensajes pendientes/nuevos usando el consumer group
                # '>' significa solo mensajes que no han sido entregados a otros consumidores
                response = await client.xreadgroup(
                    groupname=CONSUMER_GROUP,
                    consumername=CONSUMER_NAME,
                    streams={STREAM_NAME: ">"},
                    count=10,
                    block=2000,  # 2 segundos de timeout
                )

                if not response:
                    continue

                for stream, messages in response:
                    for message_id, data in messages:
                        await self.process_message(client, message_id, data)

            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error("⚠️ Error en ciclo de PedidosEventConsumer: %s", e)
                await asyncio.sleep(2)

        logger.info("🛑 PedidosEventConsumer detenido")

    async def process_message(self, client: Any, message_id: str, data: dict[str, Any]) -> None:
        """Procesa un mensaje y emite XACK al confirmar."""
        event_type = data.get("event_type")
        raw_payload = data.get("payload", "{}")

        logger.info("📥 Procesando evento [%s] ID: %s", event_type, message_id)

        try:
            payload = json.loads(raw_payload)

            if event_type == "pedido.creado":
                tenant_id = int(payload["tenant_id"])
                cliente = payload["cliente"]
                cantidad = int(payload["cantidad"])
                total = Decimal(str(payload["total"]))

                # Inserción en base de datos multi-tenant
                async with async_session_factory() as session:
                    # Configurar contexto de tenant para RLS
                    await session.execute(f"SET LOCAL app.current_tenant = '{tenant_id}'")
                    await self.service.crear(
                        session,
                        tenant_id=tenant_id,
                        cliente=cliente,
                        cantidad=cantidad,
                        total=total,
                    )
                    await session.commit()

                logger.info("✅ Pedido sincronizado en tenant %s para cliente %s", tenant_id, cliente)

            # Confirmar procesamiento con ACK
            await client.xack(STREAM_NAME, CONSUMER_GROUP, message_id)

        except Exception as e:
            logger.error("❌ Fallo procesando evento %s (ID %s): %s", event_type, message_id, e)
            # En caso de error, no enviamos XACK para permitir reintento posterior.

    def stop(self) -> None:
        """Detiene la ejecución del worker."""
        self.running = False


# Instancia global del worker
pedidos_consumer = PedidosEventConsumer()
