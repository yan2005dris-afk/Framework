"""Entorno de Alembic.

Conecta `Base.metadata` (declarativa de SQLAlchemy 2.x) con la BD gestionada
por `app.core.config.settings`. La URL se lee de la env (via Settings)
para tener una sola fuente de verdad entre el codigo y las migraciones.

Esto es la variante clasica de alembic env.py: soporta migraciones
sincronicas (con `psycopg2`) y mantiene offline mode funcional via
`render_as_batch` para Postgres.
"""
from __future__ import annotations

import sys
from logging.config import fileConfig
from pathlib import Path

from alembic import context
from sqlalchemy import engine_from_config, pool

# Hacer que 'app' sea importable cuando se corre 'alembic upgrade head'
# desde Unidad3_SaaS/. sys.path.insert(0, ...) cubre el caso en que alembic
# se invoque desde CWD distinto.
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from app.core.config import settings  # noqa: E402
from app.models import Base  # noqa: E402,E501 (registra todos los modelos)

# Cargar logging de alembic.ini (para ver INFO/WARN de las migraciones)
config = context.config
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# Sobrescribir sqlalchemy.url con el DSN sync de nuestra config.
config.set_main_option("sqlalchemy.url", settings.database_url_sync)

target_metadata = Base.metadata


def run_migrations_offline() -> None:
    """Ejecuta migraciones en 'offline mode': genera SQL sin conectar."""
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )
    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online() -> None:
    """Ejecuta migraciones conectando a la BD."""
    connectable = engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )
    with connectable.connect() as connection:
        context.configure(connection=connection, target_metadata=target_metadata)
        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
