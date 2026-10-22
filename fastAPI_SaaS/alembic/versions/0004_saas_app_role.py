"""saas_app_role

Crea el rol de aplicacion y le da solo los permisos que el runtime necesita.

Por que hace falta
-----------------
`0003_rls_policies` activa RLS sobre `tenants`, `users_saas` y `subscriptions`,
pero un **SUPERUSER ignora las politicas por completo**: `FORCE ROW LEVEL
SECURITY` ata unicamente al owner de la tabla y `BYPASSRLS` es un atributo del
rol, no del objeto. En este deploy el rol de la BD es `postgres`, o sea
superuser, asi que las politicas quedan **inertes**: cada request ve las filas de
todos los tenants y el aislamiento solo existe en el filtro del CRUD.

Este rol (`saas_app` por defecto, `SAAS_DB_APP_ROLE`) cierra el agujero:

- `NOSUPERUSER` + `NOBYPASSRLS` + nada de `CREATEDB`/`CREATEROLE`/`REPLICATION`.
- `LOGIN` con el password de `SAAS_DB_PASSWORD`: son las mismas credenciales que
  usa el servicio en runtime, asi que no hay una tercera fuente de verdad.
- No es owner de ninguna tabla y no tiene DDL. Las migraciones siguen corriendo
  con el rol admin (`SAAS_DB_ADMIN_USER`), que si puede crearlas.

Por eso el `docker-compose` arranca `alembic upgrade head` con las credenciales
de admin y `uvicorn` con las de la app: si uvicorn conectara como superuser, las
politicas de `0003` no filtrarian nada.

Idempotente: si el rol ya existe se re-sincroniza (mismos atributos + password),
asi que `alembic upgrade head` se puede correr en cada arranque del contenedor
sin fallar. Los `GRANT` son idempotentes por naturaleza.

Nota de permisos: las tablas las crea el rol admin (owner), asi que no hace
falta `ALTER DEFAULT PRIVILEGES`. Si alguna vez las creara el rol de la app,
habria que agregar DEFAULT PRIVILEGES para que las tablas nuevas hereden los
mismos permisos.

Revision ID: 0004_saas_app_role
Revises: 0003_rls_policies
Create Date: 2026-10-22 18:00:00.000000
"""
from __future__ import annotations

from typing import Sequence, Union

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

# Misma fuente de verdad que `alembic/env.py` y que el runtime: el rol de la
# app y su password salen de la config, no de literales en la migracion.
from app.core.config import settings

revision: str = "0004_saas_app_role"
down_revision: Union[str, None] = "0003_rls_policies"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

#: Tablas que toca el runtime del SaaS (lectura y escritura, nada de DDL).
APP_TABLES = ("tenants", "users_saas", "plans", "subscriptions")
#: Secuencias de esas tablas: sin USAGE el INSERT ... RETURNING id falla.
APP_SEQUENCES = (
    "tenants_id_seq",
    "users_saas_id_seq",
    "plans_id_seq",
    "subscriptions_id_seq",
)

_PG = postgresql.dialect()


def _ident(name: str) -> str:
    """Nombre de rol o de base como identificador citado de Postgres."""
    return _PG.identifier_preparer.quote(name)


def _lit(value: str) -> str:
    """Valor como literal citado de Postgres (usado para el password)."""
    return str(sa.literal(value).compile(dialect=_PG, compile_kwargs={"literal_binds": True}))


def _database_name() -> str:
    """Base de datos destino.

    Online sale del DSN de la conexion (la fuente de verdad); en offline
    (`alembic upgrade head --sql`) no hay engine, asi que cae al setting.
    """
    try:
        return str(op.get_bind().engine.url.database)
    except Exception:  # noqa: BLE001 - offline mode: MockConnection sin engine
        return settings.saas_db_name


def _create_or_sync_role() -> None:
    """Crea el rol de la app, o lo re-sincroniza si ya existe.

    Va en un `DO` porque `CREATE ROLE` no se puede escribir estaticamente dentro
    de PL/pgSQL (solo un subconjunto de comandos); `EXECUTE format(...)` con
    `%I`/`%L` hace el quoting por nosotros y evita inyeccion en el password.
    """
    op.execute(
        sa.text(
            f"""
            DO $do$
            DECLARE
                rol text := {_lit(settings.saas_db_app_role)};
                pwd text := {_lit(settings.saas_db_password)};
            BEGIN
                IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = rol) THEN
                    EXECUTE format(
                        'CREATE ROLE %I LOGIN NOSUPERUSER NOCREATEDB '
                        'NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD %L',
                        rol, pwd
                    );
                ELSE
                    EXECUTE format(
                        'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB '
                        'NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD %L',
                        rol, pwd
                    );
                END IF;
            END
            $do$;
            """
        )
    )


def upgrade() -> None:
    role = _ident(settings.saas_db_app_role)

    _create_or_sync_role()

    op.execute(sa.text(f'GRANT CONNECT ON DATABASE {_ident(_database_name())} TO {role}'))
    op.execute(sa.text(f"GRANT USAGE ON SCHEMA public TO {role}"))
    op.execute(
        sa.text(
            "GRANT SELECT, INSERT, UPDATE, DELETE ON "
            f"{', '.join(_ident(t) for t in APP_TABLES)} TO {role}"
        )
    )
    op.execute(
        sa.text(
            f"GRANT USAGE, SELECT ON SEQUENCE {', '.join(_ident(s) for s in APP_SEQUENCES)} "
            f"TO {role}"
        )
    )


def downgrade() -> None:
    role = _ident(settings.saas_db_app_role)

    op.execute(sa.text(f"REVOKE ALL ON {', '.join(_ident(s) for s in APP_SEQUENCES)} FROM {role}"))
    op.execute(sa.text(f"REVOKE ALL ON {', '.join(_ident(t) for t in APP_TABLES)} FROM {role}"))
    op.execute(sa.text(f"REVOKE USAGE ON SCHEMA public FROM {role}"))
    op.execute(sa.text(f"REVOKE CONNECT ON DATABASE {_ident(_database_name())} FROM {role}"))

    # `IF EXISTS` porque el rol pudo haberse creado a mano antes de la
    # migracion: en ese caso el DROP tiene que ser inocuo igual.
    op.execute(sa.text(f"DROP ROLE IF EXISTS {role}"))