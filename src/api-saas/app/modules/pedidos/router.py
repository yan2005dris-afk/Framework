"""API de pedidos multi-tenant.

Endpoints (los que pide la actividad):

- `PUT    /api/{tenant_id}/pedidos/{pedido_id}/estado`  -> cambia el estado
- `DELETE /api/{tenant_id}/pedidos/{pedido_id}`         -> elimina el pedido

Endpoints de apoyo para poder demostrar la validacion de Pydantic de punta
a punta:

- `POST /api/{tenant_id}/pedidos`  -> crea (aqui actuan `cantidad > 0` y `total > 0`)
- `GET  /api/{tenant_id}/pedidos`  -> lista (paginado, filtro por estado)
- `GET  /api/{tenant_id}/pedidos/{pedido_id}` -> detalle

Aislamiento
-----------
Todo pasa por `get_tenant_db`, que:

1. exige identidad (401 si el gateway no la inyecto),
2. valida que el `{tenant_id}` del path sea el del token (403 si no),
3. setea el GUC `app.current_tenant` para RLS,
4. cierra la transaccion del request.

Ademas, cada query del repositorio filtra por `tenant_id`: un pedido de otro
tenant devuelve 404, nunca 403 con datos ajenos.

Errores de dominio (`PedidoNoEncontradoError`, `TransicionInvalidaError`) los
traduce `register()` en `app/modules/pedidos/__init__.py`, asi que los
handlers quedan sin `try/except` de traduccion HTTP.
"""
from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Query, status
from fastapi.responses import Response

from app.modules.pedidos.dependencies import DbSession, PedidoServiceDep
from app.modules.pedidos.enums import EstadoPedido
from app.modules.pedidos.schemas import (
    PedidoCreate,
    PedidoEstadoUpdate,
    PedidoListResponse,
    PedidoResponse,
)
from app.shared.tenancy.dependencies import TenancyServiceDep, tenant_or_404
from app.shared.tenancy.principal import PrincipalDep, require_tenant_admin

router = APIRouter(prefix="/api/{tenant_id}/pedidos", tags=["pedidos-saas"])


@router.post("", response_model=PedidoResponse, status_code=status.HTTP_201_CREATED)
async def create_pedido(
    tenant_id: int,
    pedido_data: PedidoCreate,
    db: DbSession,
    service: PedidoServiceDep,
    tenancy: TenancyServiceDep,
    principal: PrincipalDep,
) -> object:
    """Crea un pedido en el tenant.

    `cantidad` y `total` llegan ya validados por Pydantic (`gt=0`): si el
    cliente manda 0 o un valor negativo, FastAPI responde **422** y este
    handler ni se ejecuta.
    """
    await tenant_or_404(db, tenant_id, tenancy)

    return await service.crear(
        db,
        tenant_id,
        cliente=pedido_data.cliente,
        cantidad=pedido_data.cantidad,
        total=pedido_data.total,
    )


@router.get("", response_model=PedidoListResponse)
async def list_pedidos(
    tenant_id: int,
    db: DbSession,
    service: PedidoServiceDep,
    principal: PrincipalDep,
    estado: Annotated[EstadoPedido | None, Query(description="Filtro por estado")] = None,
    skip: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 100,
) -> PedidoListResponse:
    """Lista los pedidos del tenant (paginado)."""
    pedidos, total = await service.listar(
        db, tenant_id, skip=skip, limit=limit, estado=estado
    )
    return PedidoListResponse(pedidos=pedidos, total=total)


@router.get("/{pedido_id}", response_model=PedidoResponse)
async def get_pedido(
    tenant_id: int,
    pedido_id: int,
    db: DbSession,
    service: PedidoServiceDep,
    principal: PrincipalDep,
) -> object:
    """Devuelve un pedido del tenant."""
    return await service.obtener(db, tenant_id, pedido_id)


@router.put(
    "/{pedido_id}/estado",
    response_model=PedidoResponse,
    responses={
        400: {"description": "Transicion de estado invalida"},
        403: {"description": "Se requiere admin/superadmin del tenant"},
        404: {"description": "Pedido no encontrado en el tenant"},
    },
)
async def update_pedido_estado(
    tenant_id: int,
    pedido_id: int,
    estado_data: PedidoEstadoUpdate,
    db: DbSession,
    service: PedidoServiceDep,
    principal: PrincipalDep,
) -> object:
    """Cambia el estado de un pedido aplicando la maquina de estados.

    Reglas:

    - solo `admin`/`superadmin` del tenant (403 en caso contrario);
    - el pedido tiene que pertenecer al tenant (404);
    - la transicion tiene que ser valida (400). `entregado` y `cancelado`
      son terminales.
    """
    require_tenant_admin(principal)
    return await service.cambiar_estado(
        db, tenant_id, pedido_id, estado_data.estado
    )


@router.delete(
    "/{pedido_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    responses={
        403: {"description": "Se requiere admin/superadmin del tenant"},
        404: {"description": "Pedido no encontrado en el tenant"},
    },
)
async def delete_pedido(
    tenant_id: int,
    pedido_id: int,
    db: DbSession,
    service: PedidoServiceDep,
    principal: PrincipalDep,
) -> Response:
    """Elimina un pedido del tenant (admin/superadmin)."""
    require_tenant_admin(principal)
    await service.eliminar(db, tenant_id, pedido_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)
