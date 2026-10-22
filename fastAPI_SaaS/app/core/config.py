"""Configuracion del servicio cargada desde variables de entorno / .env.

Convenciones:
- Variables sensibles (secrets, DSNs) usan prefijo `SAAS_`.
- Variables publicas (env, CORS origins, titulo) no llevan prefijo.
- `cors_origins` es CSV; se parsea con el property `cors_origins_list`.
- `database_url` construye el DSN asyncpg en runtime para no requerir
  un valor separado en env (single source of truth: las 4 vars).
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
    saas_db_user: str = "postgres"
    saas_db_password: str = "postgres"
    saas_db_name: str = "multicatalogo_saas"

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
        """DSN async para SQLAlchemy 2.x. Usado por `app/core/db.py` y Alembic."""
        return (
            f"postgresql+asyncpg://{self.saas_db_user}:{self.saas_db_password}"
            f"@{self.saas_db_host}:{self.saas_db_port}/{self.saas_db_name}"
        )

    @property
    def database_url_sync(self) -> str:
        """DSN sync para Alembic (alembic usa psycopg2/psycopg sync por defecto)."""
        return (
            f"postgresql+psycopg2://{self.saas_db_user}:{self.saas_db_password}"
            f"@{self.saas_db_host}:{self.saas_db_port}/{self.saas_db_name}"
        )


settings = Settings()
