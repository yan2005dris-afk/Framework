"""pedidos

Crea la tabla `pedidos` y le aplica el mismo modelo de aislamiento que
`users_saas` / `subscriptions`:

1. `tenant_id` FK -> `tenants.id` con `ON DELETE CASCADE`.
2. **CHECK constraints** para las reglas de negocio que la actividad pide
   validar con Pydantic: `cantidad > 0`, `total > 0` y un estado dentro del
   conjunto valido. Pydantic es la primera barrera (422 en el borde); estas
   constraints son la ultima (la BD no acepta una fila invalida aunque se
   escriba por fuera de la API).
3. **RLS** con `FORCE ROW LEVEL SECURITY`, igual que `0003_rls_policies`:
   las politicas comparan `tenant_id` contra el GUC `app.current_tenant`
   que setea `app/shared/db/session.py::get_tenant_db` con `SET LOCAL`.
4. **GRANT** al rol de aplicacion (`saas_app` por defecto). Sin esto, el
   runtime recibe "permission denied for table pedidos" y sin `USAGE` sobre
   la secuencia el `INSERT ... RETURNING id` falla.

Revision ID: 0005_pedidos
Revises: 0004_saas_app_role
Create Date: 2026-11-17 10:00:00.000000
"""
from __future__ import annotations

from typing import Sequence, Union

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

from app.shared.config import settings

revision: str = "0005_pedidos"
down_revision: Union[str, None] = "0004_saas_app_role"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

#: Misma expresion que usa `0003_rls_policies`. `NULLIF(..., '')` evita que
#: un GUC vacio reviente el cast y lo trata como "sin tenant" (fail closed).
GUC = "NULLIF(current_setting('app.current_tenant', true), '')::int"

#: Secuencia implicita del `SERIAL` que crea la PK.
SECUENCIA = "pedidos_id_seq"

_PG = postgresql.dialect()


def _ident(name: str) -> str:
    """Nombre de rol/secuencia como identificador citado de Postgres."""
    return _PG.identifier_preparer.quote(name)


def upgrade() -> None:
    op.create_table(
        "pedidos",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("tenant_id", sa.Integer(), nullable=False),
        sa.Column("cliente", sa.String(length=255), nullable=False),
        sa.Column("cantidad", sa.Integer(), nullable=False),
        sa.Column("total", sa.Numeric(precision=12, scale=2), nullable=False),
        sa.Column(
            "estado",
            sa.String(length=20),
            nullable=False,
            server_default="pendiente",
        ),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=sa.text("now()"),
            nullable=False,
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.text("now()"),
            nullable=False,
        ),
        sa.ForeignKeyConstraint(
            ["tenant_id"],
            ["tenants.id"],
            ondelete="CASCADE",
            name="fk_pedidos_tenant_id",
        ),
        sa.PrimaryKeyConstraint("id"),
        sa.CheckConstraint("cantidad > 0", name="ck_pedidos_cantidad_positiva"),
        sa.CheckConstraint("total > 0", name="ck_pedidos_total_positivo"),
        sa.CheckConstraint(
            "estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado')",
            name="ck_pedidos_estado_valido",
        ),
    )
    op.create_index("ix_pedidos_tenant_estado", "pedidos", ["tenant_id", "estado"])

    # --- RLS: mismo patron que 0003_rls_policies -------------------------
    op.execute("ALTER TABLE pedidos ENABLE ROW LEVEL SECURITY")
    op.execute("ALTER TABLE pedidos FORCE ROW LEVEL SECURITY")
    op.execute("DROP POLICY IF EXISTS pedidos_tenant_isolation ON pedidos")
    op.execute(
        f"""
        CREATE POLICY pedidos_tenant_isolation ON pedidos
        FOR ALL
        USING (tenant_id = {GUC})
        WITH CHECK (tenant_id = {GUC})
        """
    )

    # --- Permisos del rol de aplicacion (heredados de 0004) --------------
    role = _ident(settings.saas_db_app_role)
    op.execute(sa.text(f"GRANT SELECT, INSERT, UPDATE, DELETE ON pedidos TO {role}"))
    op.execute(sa.text(f"GRANT USAGE, SELECT ON SEQUENCE {_ident(SECUENCIA)} TO {role}"))


def downgrade() -> None:
    role = _ident(settings.saas_db_app_role)

    op.execute(
        sa.text(f"REVOKE ALL ON SEQUENCE {_ident(SECUENCIA)} FROM {role}")
    )
    op.execute(sa.text(f"REVOKE ALL ON pedidos FROM {role}"))

    op.execute("DROP POLICY IF EXISTS pedidos_tenant_isolation ON pedidos")
    op.execute("ALTER TABLE pedidos NO FORCE ROW LEVEL SECURITY")
    op.execute("ALTER TABLE pedidos DISABLE ROW LEVEL SECURITY")

    op.drop_index("ix_pedidos_tenant_estado", table_name="pedidos")
    op.drop_table("pedidos")
