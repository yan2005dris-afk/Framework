"""Tests de integración del aislamiento multi-tenant por RLS.

Verifica contra Postgres real lo que sqlite no puede: que las politicas de
la migración `0003_rls_policies` (con `FORCE ROW LEVEL SECURITY`) impiden
leer y escribir filas de otro tenant, y que sin GUC no se ve nada.

Los tests de aislamiento conectan con el **rol de la app** (`saas_app`), no con
`postgres`. No es un detalle: un superuser ignora las politicas RLS por completo
(`FORCE` solo ata al owner de la tabla y `BYPASSRLS` es un atributo del rol), asi
que los tests corridos como superuser no probarian nada. Por eso el fixture
saltea si el rol conectado bypasea RLS, y por eso existe
`TestSuperuserBypassesRLS`, que documenta el caso contrario: como superuser se
ven los dos tenants.

Pre-requisitos:
    docker compose up -d db
    (cd fastAPI_SaaS && SAAS_DB_HOST=localhost SAAS_DB_PORT=5432 \\
        SAAS_DB_NAME=multicatalogo uv run alembic upgrade head)

Ejecutar:
    cd tests/integration
    SAAS_DB_USER=saas_app SAAS_DB_PASSWORD=saas_app \\
        uv run --project ../../fastAPI_SaaS pytest test_saas_rls_integration.py -v

Variables (todas con prefijo `SAAS_`, defaults pensados para el stack local):

| Variable | Default | Para que sirve |
|---|---|---|
| `SAAS_DB_HOST` / `SAAS_DB_PORT` / `SAAS_DB_NAME` | localhost / 5432 / multicatalogo | Dónde está la BD |
| `SAAS_DB_APP_ROLE` (o `SAAS_DB_USER`) | saas_app | Rol **no superuser** de los tests de aislamiento |
| `SAAS_DB_PASSWORD` | saas_app | Password de ese rol (la que crea la migración `0004`) |
| `SAAS_DB_ADMIN_USER` / `SAAS_DB_ADMIN_PASSWORD` | postgres / postgres | Rol admin, solo para el test del bypass |

Si la BD no responde, los tests se saltan (skip) en vez de fallar.
"""
from __future__ import annotations

import os

import psycopg2
import pytest
from psycopg2 import errors as pg_errors

DB_HOST = os.getenv("SAAS_DB_HOST", "localhost")
DB_PORT = os.getenv("SAAS_DB_PORT", "5432")
DB_NAME = os.getenv("SAAS_DB_NAME", "multicatalogo")

# Rol de la APP: es quien tiene que quedar filtrado por las politicas.
APP_ROLE = os.getenv("SAAS_DB_APP_ROLE") or os.getenv("SAAS_DB_USER", "saas_app")
APP_PASSWORD = os.getenv("SAAS_DB_PASSWORD", "saas_app")

# Rol ADMIN: el del stack. Solo lo usa el test que documenta el bypass.
ADMIN_USER = os.getenv("SAAS_DB_ADMIN_USER", "postgres")
ADMIN_PASSWORD = os.getenv("SAAS_DB_ADMIN_PASSWORD", "postgres")

#: IDs altos y fijos: no chocan con los tenants reales y se pueden borrar.
TENANT_A = 9001
TENANT_B = 9002

REQUIRED_POLICIES = {
    "users_saas_tenant_isolation",
    "subscriptions_tenant_isolation",
    "tenants_self_read",
}


def _connect(user: str, password: str):
    return psycopg2.connect(
        host=DB_HOST, port=DB_PORT, user=user, password=password, dbname=DB_NAME
    )


def _role_bypasses_rls(cur) -> bool:
    """True si el rol conectado puede saltarse las politicas RLS.

    Postgres ignora RLS para superusers y para roles con el atributo BYPASSRLS;
    en ambos casos cualquier test de aislamiento seria verde por falso positivo.
    """
    cur.execute(
        "SELECT coalesce(rolsuper, false) OR coalesce(rolbypassrls, false) "
        "FROM pg_roles WHERE rolname = current_user"
    )
    row = cur.fetchone()
    return bool(row[0]) if row else False


def _seed_fixtures(cur) -> None:
    """Un tenant con un usuario y una suscripción cada uno (rollback al final)."""
    for tenant_id, slug in ((TENANT_A, "rls-it-a"), (TENANT_B, "rls-it-b")):
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(tenant_id),))
        cur.execute(
            "INSERT INTO tenants (id, slug, name) VALUES (%s, %s, %s) "
            "ON CONFLICT (id) DO UPDATE SET slug = EXCLUDED.slug",
            (tenant_id, slug, slug),
        )
        cur.execute(
            "INSERT INTO users_saas (tenant_id, email, hashed_password, rol) "
            "VALUES (%s, %s, 'x', 'admin')",
            (tenant_id, f"user-{slug}@example.com"),
        )
        cur.execute(
            "INSERT INTO subscriptions (tenant_id, plan_id, status) "
            "VALUES (%s, (SELECT id FROM plans ORDER BY id LIMIT 1), 'active')",
            (tenant_id,),
        )


def _open(user: str, password: str):
    """Conecta y devuelve `(connection, cursor)` con autocommit off.

    Saltar (no fallar) si la BD no esta o si faltan las politicas: estos tests
    son opcionales en un entorno sin Postgres.
    """
    try:
        connection = _connect(user, password)
    except psycopg2.OperationalError as exc:
        pytest.skip(f"Postgres no disponible en {DB_HOST}:{DB_PORT} ({exc})")

    connection.autocommit = False
    cur = connection.cursor()

    cur.execute("SELECT policyname FROM pg_policies WHERE schemaname = current_schema()")
    missing = REQUIRED_POLICIES - {row[0] for row in cur.fetchall()}
    if missing:
        connection.close()
        pytest.skip(
            "Faltan politicas RLS "
            f"({', '.join(sorted(missing))}). Corré: "
            "cd fastAPI_SaaS && SAAS_DB_HOST=localhost SAAS_DB_PORT=5432 "
            "SAAS_DB_NAME=multicatalogo uv run alembic upgrade head"
        )
    return connection, cur


@pytest.fixture
def conn():
    """Conexion del ROL DE LA APP con TODAS las pruebas dentro de una transaccion
    que se revierte.

    Importante: se aplica `SET LOCAL app.current_tenant` sobre la MISMA
    transaccion, que es exactamente como funciona `get_tenant_db` en runtime.
    """
    connection, cur = _open(APP_ROLE, APP_PASSWORD)

    if _role_bypasses_rls(cur):
        connection.close()
        pytest.skip(
            f"El rol {APP_ROLE} bypasea RLS (superuser o BYPASSRLS): las politicas "
            "no filtrarian nada y estos tests serian verde por falso positivo. "
            "Hace falta un rol sin superuser (lo crea la migracion "
            "0004_saas_app_role) y conectar el servicio con ese rol, no con postgres."
        )

    _seed_fixtures(cur)
    yield connection, cur

    connection.rollback()  # no deja rastro
    connection.close()


@pytest.fixture
def admin_conn():
    """Conexion del rol ADMIN (superuser del stack), misma transaccion reversible."""
    connection, cur = _open(ADMIN_USER, ADMIN_PASSWORD)

    if not _role_bypasses_rls(cur):
        connection.close()
        pytest.skip(
            f"El rol admin {ADMIN_USER} no bypasea RLS, asi que este archivo no puede "
            "demostrar el bypass. Apuntá SAAS_DB_ADMIN_USER al superuser de la BD."
        )

    _seed_fixtures(cur)
    yield connection, cur

    connection.rollback()
    connection.close()


class TestRowLevelSecurity:
    def test_read_only_returns_own_rows(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute("SELECT tenant_id FROM users_saas")
        assert {row[0] for row in cur.fetchall()} == {TENANT_A}

    def test_subscriptions_are_isolated(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute("SELECT tenant_id FROM subscriptions")
        assert {row[0] for row in cur.fetchall()} == {TENANT_A}

    def test_tenants_row_is_isolated(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute("SELECT id FROM tenants WHERE id IN (%s, %s)", (TENANT_A, TENANT_B))
        assert {row[0] for row in cur.fetchall()} == {TENANT_A}

    def test_insert_into_other_tenant_is_denied(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        with pytest.raises(pg_errors.InsufficientPrivilege):
            cur.execute(
                "INSERT INTO users_saas (tenant_id, email, hashed_password, rol) "
                "VALUES (%s, 'cross-tenant@example.com', 'x', 'member')",
                (TENANT_B,),
            )
        connection.rollback()  # la sentencia aborta la transaccion

    def test_update_other_tenant_row_is_denied(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute(
            "UPDATE users_saas SET rol = 'superadmin' WHERE email = %s",
            ("user-rls-it-b@example.com",),
        )
        assert cur.rowcount == 0
        connection.rollback()

    def test_delete_other_tenant_row_is_denied(self, conn):
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute("DELETE FROM users_saas WHERE email = %s", ("user-rls-it-b@example.com",))
        assert cur.rowcount == 0
        connection.rollback()

    def test_without_guc_nothing_is_visible(self, conn):
        """Fail closed: sin `app.current_tenant` no se ve ninguna fila."""
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', '', true)")
        cur.execute("SELECT count(*) FROM users_saas")
        assert cur.fetchone()[0] == 0
        cur.execute("SELECT count(*) FROM subscriptions")
        assert cur.fetchone()[0] == 0

    def test_superadmin_path_switches_tenant_and_sees_it(self, conn):
        """Asi es como un superadmin cruza tenants: el GUC pasa a ser su destino."""
        connection, cur = conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_B),))
        cur.execute("SELECT tenant_id FROM users_saas")
        assert {row[0] for row in cur.fetchall()} == {TENANT_B}


class TestSuperuserBypassesRLS:
    """El motivo por el que el servicio no puede conectar como `postgres`.

    Si este test falla porque el rol admin no bypasea RLS, el problema no es el
    test: es que el stack ya no tiene un superuser con ese nombre y las politicas
    se estan aplicando de verdad. El fixture lo saltea en ese caso.
    """

    def test_superuser_sees_rows_of_every_tenant(self, admin_conn):
        """Con GUC en TENANT_A, el superuser igual ve los datos de TENANT_B."""
        connection, cur = admin_conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute(
            "SELECT tenant_id FROM users_saas WHERE tenant_id IN (%s, %s)",
            (TENANT_A, TENANT_B),
        )
        assert {row[0] for row in cur.fetchall()} == {TENANT_A, TENANT_B}

    def test_superuser_sees_every_tenant_row(self, admin_conn):
        """Mismo chequeo sobre `tenants`: sin RLS, la fila del otro tenant existe."""
        connection, cur = admin_conn
        cur.execute("SELECT set_config('app.current_tenant', %s, true)", (str(TENANT_A),))
        cur.execute("SELECT id FROM tenants WHERE id IN (%s, %s)", (TENANT_A, TENANT_B))
        assert {row[0] for row in cur.fetchall()} == {TENANT_A, TENANT_B}