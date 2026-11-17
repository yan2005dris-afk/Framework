"""Tests de API del recurso `pedidos`.

Cubre tres frentes:

1. **Validacion Pydantic**: `cantidad <= 0` y `total <= 0` se rechazan con
   422 antes de tocar la BD (y la BD no guarda nada).
2. **Contrato HTTP**: 201 / 200 / 204 / 400 / 401 / 403 / 404 / 422.
3. **Aislamiento multi-tenant**: un pedido de otro tenant nunca se lee, se
   modifica ni se borra (404), un principal de otro tenant recibe 403 y un
   request sin identidad recibe 401.
"""
from __future__ import annotations

from decimal import Decimal

import pytest
from httpx import AsyncClient
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.pedidos.enums import EstadoPedido
from app.modules.pedidos.models import Pedido
from app.shared.tenancy.models import Tenant
from tests.conftest import auth_headers

pytestmark = pytest.mark.asyncio

PEDIDOS_URL = "/api/{tenant_id}/pedidos"


async def _create_tenant(db: AsyncSession, slug: str) -> int:
    tenant = Tenant(slug=slug, name=slug)
    db.add(tenant)
    await db.commit()
    await db.refresh(tenant)
    return tenant.id


async def _create_pedido(
    db: AsyncSession,
    tenant_id: int,
    *,
    cliente: str = "Cliente Test",
    cantidad: int = 2,
    total: str = "99.90",
    estado: EstadoPedido = EstadoPedido.PENDIENTE,
) -> int:
    pedido = Pedido(
        tenant_id=tenant_id,
        cliente=cliente,
        cantidad=cantidad,
        total=Decimal(total),
        estado=estado.value,
    )
    db.add(pedido)
    await db.commit()
    await db.refresh(pedido)
    return pedido.id


async def _pedido_count(db: AsyncSession, tenant_id: int) -> int:
    from app.modules.pedidos.repository import pedido_repository

    _, total = await pedido_repository.list_by_tenant(db, tenant_id)
    return total


# ---------------------------------------------------------------------------
# Validacion Pydantic (la regla central de la actividad)
# ---------------------------------------------------------------------------


async def test_create_pedido_valido_returns_201(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        PEDIDOS_URL.format(tenant_id=tenant_id),
        json={"cliente": "Ana Perez", "cantidad": 3, "total": "149.90"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 201, resp.text
    body = resp.json()
    assert body["tenant_id"] == tenant_id
    assert body["cliente"] == "Ana Perez"
    assert body["cantidad"] == 3
    assert body["total"] == 149.9
    assert body["estado"] == "pendiente"


@pytest.mark.parametrize("cantidad", [0, -1, -50])
async def test_create_pedido_con_cantidad_no_positiva_returns_422(
    client: AsyncClient, db: AsyncSession, cantidad: int
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        PEDIDOS_URL.format(tenant_id=tenant_id),
        json={"cliente": "Ana", "cantidad": cantidad, "total": "10.00"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 422, resp.text
    assert "cantidad" in resp.text
    # La validacion ocurre en el borde: no se persistio nada.
    assert await _pedido_count(db, tenant_id) == 0


@pytest.mark.parametrize("total", ["0", "0.00", "-0.01", "-100"])
async def test_create_pedido_con_total_no_positivo_returns_422(
    client: AsyncClient, db: AsyncSession, total: str
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        PEDIDOS_URL.format(tenant_id=tenant_id),
        json={"cliente": "Ana", "cantidad": 1, "total": total},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 422, resp.text
    assert "total" in resp.text
    assert await _pedido_count(db, tenant_id) == 0


async def test_db_check_constraint_rechaza_cantidad_invalida(
    db: AsyncSession,
) -> None:
    """Defensa en profundidad: la BD tampoco acepta una fila invalida."""
    tenant_id = await _create_tenant(db, "acme")
    db.add(
        Pedido(
            tenant_id=tenant_id,
            cliente="Sin Pydantic",
            cantidad=0,
            total=Decimal("10.00"),
        )
    )

    with pytest.raises(IntegrityError):
        await db.flush()

    await db.rollback()


# ---------------------------------------------------------------------------
# Identidad y scope
# ---------------------------------------------------------------------------


async def test_create_pedido_sin_identidad_returns_401(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.post(
        PEDIDOS_URL.format(tenant_id=tenant_id),
        json={"cliente": "Ana", "cantidad": 1, "total": "10.00"},
    )

    assert resp.status_code == 401, resp.text


async def test_admin_de_otro_tenant_returns_403(
    client: AsyncClient, db: AsyncSession
) -> None:
    own = await _create_tenant(db, "own")
    other = await _create_tenant(db, "other")

    resp = await client.get(
        PEDIDOS_URL.format(tenant_id=other),
        headers=auth_headers(tenant_id=own),
    )

    assert resp.status_code == 403, resp.text


async def test_superadmin_puede_operar_sobre_cualquier_tenant(
    client: AsyncClient, db: AsyncSession
) -> None:
    other = await _create_tenant(db, "other")

    resp = await client.get(
        PEDIDOS_URL.format(tenant_id=other),
        headers=auth_headers(tenant_id=1, rol_saas="superadmin"),
    )

    assert resp.status_code == 200, resp.text


async def test_create_pedido_en_tenant_inexistente_returns_404(
    client: AsyncClient,
) -> None:
    resp = await client.post(
        PEDIDOS_URL.format(tenant_id=999),
        json={"cliente": "Ana", "cantidad": 1, "total": "10.00"},
        headers=auth_headers(tenant_id=999),
    )

    assert resp.status_code == 404, resp.text


# ---------------------------------------------------------------------------
# Listado y detalle
# ---------------------------------------------------------------------------


async def test_list_aisla_pedidos_por_tenant(
    client: AsyncClient, db: AsyncSession
) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    await _create_pedido(db, first, cliente="A")
    await _create_pedido(db, second, cliente="B")

    resp = await client.get(
        PEDIDOS_URL.format(tenant_id=first),
        headers=auth_headers(tenant_id=first),
    )

    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["total"] == 1
    assert [p["cliente"] for p in body["pedidos"]] == ["A"]


async def test_list_filtra_por_estado(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    await _create_pedido(db, tenant_id, cliente="Pendiente")
    await _create_pedido(db, tenant_id, cliente="Enviado", estado=EstadoPedido.ENVIADO)

    resp = await client.get(
        PEDIDOS_URL.format(tenant_id=tenant_id),
        params={"estado": "enviado"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["total"] == 1
    assert body["pedidos"][0]["cliente"] == "Enviado"


async def test_get_pedido_de_otro_tenant_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    pedido_id = await _create_pedido(db, second)

    resp = await client.get(
        f"{PEDIDOS_URL.format(tenant_id=first)}/{pedido_id}",
        headers=auth_headers(tenant_id=first),
    )

    assert resp.status_code == 404, resp.text


# ---------------------------------------------------------------------------
# PUT /api/{tenant_id}/pedidos/{pedido_id}/estado
# ---------------------------------------------------------------------------


async def test_put_estado_transicion_valida_returns_200(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}/estado",
        json={"estado": "preparando"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 200, resp.text
    assert resp.json()["estado"] == "preparando"


async def test_put_estado_transicion_invalida_returns_400(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)  # pendiente

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}/estado",
        json={"estado": "entregado"},  # pendiente -> entregado no esta permitido
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 400, resp.text
    assert "Transicion invalida" in resp.json()["detail"]


async def test_put_estado_inexistente_returns_422(
    client: AsyncClient, db: AsyncSession
) -> None:
    """El enum del schema rechaza estados que no existen."""
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}/estado",
        json={"estado": "volando"},
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 422, resp.text
    assert "estado" in resp.text


async def test_put_estado_de_pedido_ajeno_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    pedido_id = await _create_pedido(db, second)

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=first)}/{pedido_id}/estado",
        json={"estado": "preparando"},
        headers=auth_headers(tenant_id=first),
    )

    assert resp.status_code == 404, resp.text


async def test_put_estado_requiere_admin(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}/estado",
        json={"estado": "preparando"},
        headers=auth_headers(tenant_id=tenant_id, rol_saas="member"),
    )

    assert resp.status_code == 403, resp.text


async def test_put_estado_sin_identidad_returns_401(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.put(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}/estado",
        json={"estado": "preparando"},
    )

    assert resp.status_code == 401, resp.text


# ---------------------------------------------------------------------------
# DELETE /api/{tenant_id}/pedidos/{pedido_id}
# ---------------------------------------------------------------------------


async def test_delete_returns_204_y_borra(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.delete(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}",
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 204, resp.text
    assert resp.content == b""

    from app.modules.pedidos.repository import pedido_repository

    assert await pedido_repository.get_by_id(db, pedido_id, tenant_id) is None


async def test_delete_de_pedido_ajeno_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    first = await _create_tenant(db, "acme")
    second = await _create_tenant(db, "globex")
    pedido_id = await _create_pedido(db, second)

    resp = await client.delete(
        f"{PEDIDOS_URL.format(tenant_id=first)}/{pedido_id}",
        headers=auth_headers(tenant_id=first),
    )

    assert resp.status_code == 404, resp.text
    # El pedido del otro tenant sigue existiendo.
    from app.modules.pedidos.repository import pedido_repository

    assert await pedido_repository.get_by_id(db, pedido_id, second) is not None


async def test_delete_requiere_admin(client: AsyncClient, db: AsyncSession) -> None:
    tenant_id = await _create_tenant(db, "acme")
    pedido_id = await _create_pedido(db, tenant_id)

    resp = await client.delete(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/{pedido_id}",
        headers=auth_headers(tenant_id=tenant_id, rol_saas="member"),
    )

    assert resp.status_code == 403, resp.text


async def test_delete_inexistente_returns_404(
    client: AsyncClient, db: AsyncSession
) -> None:
    tenant_id = await _create_tenant(db, "acme")

    resp = await client.delete(
        f"{PEDIDOS_URL.format(tenant_id=tenant_id)}/424242",
        headers=auth_headers(tenant_id=tenant_id),
    )

    assert resp.status_code == 404, resp.text
