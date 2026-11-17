"""billing

Tablas de monetizacion del SaaS + catalogo de planes semilla:
- plans: catalogo de planes (free, pro, enterprise).
- subscriptions: suscripcion 1:1 por tenant, con estado y periodo de Stripe.

Las politicas RLS de estas tablas llegan en la migracion siguiente
(`0003_rls_policies`): una migracion no debe referenciar tablas que todavia
no existen.

Revision ID: 0002_billing
Revises: 0001_initial
Create Date: 2026-10-22 16:10:00.000000
"""
from __future__ import annotations

from decimal import Decimal
from typing import Sequence, Union

import sqlalchemy as sa
from alembic import op

revision: str = "0002_billing"
down_revision: Union[str, None] = "0001_initial"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        "plans",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("slug", sa.String(length=64), nullable=False),
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("description", sa.String(length=1000), nullable=True),
        sa.Column("price_monthly", sa.Numeric(precision=10, scale=2), nullable=False),
        sa.Column("price_yearly", sa.Numeric(precision=10, scale=2), nullable=True),
        sa.Column("features", sa.JSON(), nullable=True),
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
        sa.PrimaryKeyConstraint("id"),
    )
    # El modelo declara `unique=True, index=True`: eso es un indice unico,
    # no una constraint. Evita el indice duplicado que dejaba un UniqueConstraint.
    op.create_index("ix_plans_slug", "plans", ["slug"], unique=True)

    op.create_table(
        "subscriptions",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        # UNIQUE: un tenant tiene como maximo una suscripcion (S13).
        sa.Column("tenant_id", sa.Integer(), nullable=False),
        sa.Column("plan_id", sa.Integer(), nullable=False),
        sa.Column("status", sa.String(length=50), nullable=False, server_default="active"),
        sa.Column("stripe_subscription_id", sa.String(length=255), nullable=True),
        sa.Column("current_period_start", sa.DateTime(), nullable=True),
        sa.Column("current_period_end", sa.DateTime(), nullable=True),
        sa.Column(
            "cancel_at_period_end", sa.Boolean(), nullable=False, server_default=sa.false()
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
            ["tenant_id"], ["tenants.id"], ondelete="CASCADE", name="fk_subscriptions_tenant_id"
        ),
        sa.ForeignKeyConstraint(
            ["plan_id"], ["plans.id"], ondelete="CASCADE", name="fk_subscriptions_plan_id"
        ),
        sa.PrimaryKeyConstraint("id"),
        # UNIQUE: un tenant tiene como maximo una suscripcion (S13). Es una
        # constraint (no `index=True` en el modelo), asi que no lleva indice extra.
        sa.UniqueConstraint("tenant_id"),
    )

    # Catalogo semilla. `features` es JSON (array de strings) segun el modelo.
    plans_table = sa.table(
        "plans",
        sa.column("slug", sa.String),
        sa.column("name", sa.String),
        sa.column("description", sa.String),
        sa.column("price_monthly", sa.Numeric(10, 2)),
        sa.column("price_yearly", sa.Numeric(10, 2)),
        sa.column("features", sa.JSON),
    )
    op.bulk_insert(
        plans_table,
        [
            {
                "slug": "free",
                "name": "Free",
                "description": "Plan gratuito para probar el SaaS",
                "price_monthly": Decimal("0.00"),
                "price_yearly": Decimal("0.00"),
                "features": ["1 tenant", "5 usuarios", "Catalogo ilimitado"],
            },
            {
                "slug": "pro",
                "name": "Pro",
                "description": "Plan para equipos que ya venden en multicanal",
                "price_monthly": Decimal("29.90"),
                "price_yearly": Decimal("299.00"),
                "features": [
                    "10 tenants",
                    "Usuarios ilimitados",
                    "Reportes de ventas",
                    "Integraciones",
                ],
            },
            {
                "slug": "enterprise",
                "name": "Enterprise",
                "description": "Plan corporativo con soporte dedicado",
                "price_monthly": Decimal("99.90"),
                "price_yearly": Decimal("999.00"),
                "features": ["Tenants ilimitados", "SSO", "SLA 99.9%", "Soporte 24/7"],
            },
        ],
    )


def downgrade() -> None:
    op.drop_table("subscriptions")
    op.drop_index("ix_plans_slug", table_name="plans")
    op.drop_table("plans")