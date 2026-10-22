"""rls_policies

Row-Level Security: el filtro por `tenant_id` deja de depender solo del codigo
y pasa a estar en la propia Postgres.

Idea
----
Cada request tenant-scoped setea el GUC `app.current_tenant`
(`app/core/db.py::get_tenant_db`, `set_config(..., true)` = SET LOCAL) con el
tenant efectivo. Las politicas comparan `tenant_id` contra ese GUC.

Dos detalles que hacen que esto funcione de verdad:

1. `FORCE ROW LEVEL SECURITY`. Sin el, el **owner** de la tabla ignora RLS, y
   en este deploy la app conecta con el mismo usuario que corre las
   migraciones (`POSTGRES_USER=postgres`). Con FORCE, las politicas aplican
   tambien al owner.
2. `SET LOCAL` dura lo que la transaccion. Por eso el request entero es una
   transaccion (un solo commit, en `get_tenant_db`). Un commit intermedio
   borraria el GUC y el resto del request veria cero filas (fail closed, pero
   roto).

Fail closed: sin GUC, `current_setting(..., true)` devuelve NULL, la
comparacion da NULL y no pasa ninguna fila. Un bug de contexto se traduce en
un endpoint vacio, nunca en fuga de datos entre tenants.

Tables:
- `users_saas`, `subscriptions`: filtradas por `tenant_id` (aislamiento real).
- `tenants`: cada tenant solo se ve a si mismo. Un `superadmin` entra con el
  GUC seteado al tenant de la ruta, asi que tambien puede leerlo.

Revision ID: 0003_rls_policies
Revises: 0002_billing
Create Date: 2026-10-22 17:00:00.000000
"""
from __future__ import annotations

from typing import Sequence, Union

from alembic import op

revision: str = "0003_rls_policies"
down_revision: Union[str, None] = "0002_billing"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

# NULLIF(... , ''): un GUC presente pero vacio no debe reventar la query con
# un error de cast; debe caer en el caso "sin tenant" (fail closed).
GUC = "NULLIF(current_setting('app.current_tenant', true), '')::int"

#: Tablas con datos de un tenant concreto.
TENANT_TABLES = ("users_saas", "subscriptions")


def _enable(table: str) -> None:
    op.execute(f"ALTER TABLE {table} ENABLE ROW LEVEL SECURITY")
    op.execute(f"ALTER TABLE {table} FORCE ROW LEVEL SECURITY")


def _disable(table: str) -> None:
    op.execute(f"ALTER TABLE {table} NO FORCE ROW LEVEL SECURITY")
    op.execute(f"ALTER TABLE {table} DISABLE ROW LEVEL SECURITY")


def upgrade() -> None:
    for table in TENANT_TABLES:
        _enable(table)
        op.execute(f"DROP POLICY IF EXISTS {table}_tenant_isolation ON {table}")
        op.execute(
            f"""
            CREATE POLICY {table}_tenant_isolation ON {table}
            FOR ALL
            USING (tenant_id = {GUC})
            WITH CHECK (tenant_id = {GUC})
            """
        )

    # Un tenant solo se ve a si mismo. La FK `ON DELETE CASCADE` de
    # users_saas/subscriptions sigue funcionando: Postgres valida la politica
    # referencial con los RI triggers del owner.
    _enable("tenants")
    op.execute("DROP POLICY IF EXISTS tenants_self_read ON tenants")
    op.execute(
        f"""
        CREATE POLICY tenants_self_read ON tenants
        FOR ALL
        USING (id = {GUC})
        WITH CHECK (id = {GUC})
        """
    )


def downgrade() -> None:
    for table in TENANT_TABLES:
        op.execute(f"DROP POLICY IF EXISTS {table}_tenant_isolation ON {table}")
        _disable(table)

    op.execute("DROP POLICY IF EXISTS tenants_self_read ON tenants")
    _disable("tenants")