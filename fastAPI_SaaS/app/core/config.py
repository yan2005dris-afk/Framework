"""Configuracion del servicio cargada desde variables de entorno / .env.

Convenciones:
- Variables sensibles (secrets, DSNs) usan prefijo `SAAS_`.
- Variables publicas (env, CORS origins, titulo) no llevan prefijo.
- `cors_origins` es CSV; se parsea con el property `cors_origins_list`.
- `database_url` construye el DSN asyncpg en runtime para no requerir
  un valor separado en env (single source of truth: las 4 vars).

Dos identidades, no una
-----------------------
La BD se accede con DOS roles distintos y mezclarlos rompe el aislamiento:

- `saas_db_user` / `saas_db_password` (**runtime**): rol de aplicacion
  (`saas_db_app_role`, por defecto `saas_app`). NO superuser, porque las
  politicas RLS de `alembic/versions/0003_rls_policies.py` solo filteran para
  quien no tiene `rolsuper`/`rolbypassrls`. Es lo que usa `app/core/db.py`.
- `saas_db_admin_user` / `saas_db_admin_password` (**solo Alembic**): rol con
  permisos de DDL. Las migraciones crean tablas, roles y politicas, cosas que
  el rol de la app no puede hacer (ni debe poder hacer). Lo usa
  `alembic/env.py` via `migration_database_url_sync`.

Host, puerto y base son compartidos por las dos: el aislamiento va por rol, no
por instancia.
"""
from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Settings tipados. Importar `settings` (instancia singleton) en el resto del codigo."""

    # App
    app_env: str = "dev"
    api_title: str = "MultiCatalogo SaaS API"
    cors_origins: str = "http://localhost:5173"

    # U3-T2 GA6: URL del auth-service al que el SaaS proxy-eea los
    # endpoints de auth. En docker-compose es http://auth-service:3001
    # (resoluble via DNS Docker). En dev nativo, http://localhost:3001.
    auth_service_url: str = "http://auth-service:3001"

    # DB (asyncpg). Usa puerto 5435 para no chocar con la BD del nucleo Go/Fiber (5434).
    saas_db_host: str = "localhost"
    saas_db_port: int = 5435
    saas_db_name: str = "multicatalogo_saas"

    # RUNTIME: credenciales del rol de aplicacion. El default es el rol sin
    # superuser a proposito: si falta la env var, lo que falla es un permiso
    # (ruidoso) en vez de un RLS silenciosamente bypasseado.
    saas_db_user: str = "saas_app"
    saas_db_password: str = "saas_app"

    # Nombre del rol que crea/permite la migracion `0004_saas_app_role`.
    # Lo leen la migracion y los tests de integracion de RLS, asi que hay una
    # sola fuente de verdad para el nombre del rol.
    saas_db_app_role: str = "saas_app"

    # ADMIN (solo Alembic): permisos de DDL. El default corresponde al
    # superuser del stack (`postgres`), que ademas es owner de las tablas.
    saas_db_admin_user: str = "postgres"
    saas_db_admin_password: str = "postgres"

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    @property
    def cors_origins_list(self) -> list[str]:
        """Parsea CORS_ORIGINS como CSV en una lista usable por CORSMiddleware."""
        return [origin.strip() for origin in self.cors_origins.split(",") if origin.strip()]

    @property
    def database_url(self) -> str:
        """DSN async de RUNTIME (rol de la app). Usado por `app/core/db.py`."""
        return (
            f"postgresql+asyncpg://{self.saas_db_user}:{self.saas_db_password}"
            f"@{self.saas_db_host}:{self.saas_db_port}/{self.saas_db_name}"
        )

    @property
    def database_url_sync(self) -> str:
        """DSN sync de RUNTIME. Alembic NO lo usa: las migraciones van con admin."""
        return (
            f"postgresql+psycopg2://{self.saas_db_user}:{self.saas_db_password}"
            f"@{self.saas_db_host}:{self.saas_db_port}/{self.saas_db_name}"
        )

    @property
    def migration_database_url_sync(self) -> str:
        """DSN sync de ADMIN para Alembic (las migraciones hacen DDL).

        Mismo esquema que `database_url_sync` pero con `saas_db_admin_*`: si
        Alembic corriera con el rol de la app, el `CREATE ROLE` de la migracion
        `0004_saas_app_role` fallaria por falta de permisos y, peor, el servicio
        en runtime podria escribir en el schema sin control de RLS util.
        """
        return (
            f"postgresql+psycopg2://{self.saas_db_admin_user}:{self.saas_db_admin_password}"
            f"@{self.saas_db_host}:{self.saas_db_port}/{self.saas_db_name}"
        )


settings = Settings()
