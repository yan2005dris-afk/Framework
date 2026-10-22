"""CRUD operations para UserSaaS.

Operaciones de base de datos para usuarios SaaS con aislamiento por tenant.

Passwords: `bcrypt` directo (no passlib). Passlib 1.7.4 esta sin mantener desde
2020 y su backend bcrypt rompe con bcrypt 5.x al sondear el bug de wrap
(`detect_wrap_bug` -> ValueError). Antes de hashear se aplica SHA-256 + base64
para eliminar el limite de 72 bytes de bcrypt (mismo criterio que el esquema
`bcrypt_sha256` de passlib).
"""
from __future__ import annotations

import base64
import hashlib
from typing import Optional

import bcrypt
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.user import UserSaaS

#: Coste de bcrypt. El nucleo Go usa 10 para auth.users (datos demo); los
#: usuarios SaaS son credenciales propias, asi que sube a 12.
BCRYPT_ROUNDS = 12


def _prepare_password(password: str) -> bytes:
    """SHA-256 + base64: deja el secret en 44 bytes, dentro del limite de bcrypt."""
    return base64.b64encode(hashlib.sha256(password.encode("utf-8")).digest())


class UserSaaSCRUD:
    """CRUD operations para UserSaaS."""

    @staticmethod
    def hash_password(password: str) -> str:
        """Hashea un password con bcrypt."""
        return bcrypt.hashpw(
            _prepare_password(password),
            bcrypt.gensalt(rounds=BCRYPT_ROUNDS),
        ).decode("utf-8")

    @staticmethod
    def verify_password(plain_password: str, hashed_password: str) -> bool:
        """Verifica un password contra un hash. False si el hash esta corrupto."""
        try:
            return bcrypt.checkpw(
                _prepare_password(plain_password), hashed_password.encode("utf-8")
            )
        except ValueError:
            return False

    async def create(
        self,
        db: AsyncSession,
        tenant_id: int,
        email: str,
        password: str,
        rol: str = "member",
    ) -> UserSaaS:
        """Crea un nuevo usuario SaaS en un tenant."""
        hashed_password = self.hash_password(password)
        user = UserSaaS(
            tenant_id=tenant_id,
            email=email,
            hashed_password=hashed_password,
            rol=rol,
        )
        db.add(user)
        # flush(), no commit(): la transaccion la cierra `get_tenant_db` al
        # final del request, para que el GUC de RLS siga activo durante toda
        # la operacion (un commit intermedio borraria el SET LOCAL).
        await db.flush()
        await db.refresh(user)
        return user

    async def get_by_id(self, db: AsyncSession, user_id: int, tenant_id: int) -> Optional[UserSaaS]:
        """Obtiene un usuario por ID y tenant (aislamiento)."""
        result = await db.execute(
            select(UserSaaS).where(
                UserSaaS.id == user_id,
                UserSaaS.tenant_id == tenant_id,
            )
        )
        return result.scalar_one_or_none()

    async def get_by_email(
        self, db: AsyncSession, email: str, tenant_id: int
    ) -> Optional[UserSaaS]:
        """Obtiene un usuario por email y tenant (aislamiento)."""
        result = await db.execute(
            select(UserSaaS).where(
                UserSaaS.email == email,
                UserSaaS.tenant_id == tenant_id,
            )
        )
        return result.scalar_one_or_none()

    async def list_by_tenant(
        self, db: AsyncSession, tenant_id: int, skip: int = 0, limit: int = 100
    ) -> tuple[list[UserSaaS], int]:
        """Lista todos los usuarios de un tenant."""
        result = await db.execute(
            select(UserSaaS)
            .where(UserSaaS.tenant_id == tenant_id)
            .offset(skip)
            .limit(limit)
        )
        users = list(result.scalars().all())

        total = await db.scalar(
            select(func.count()).select_from(UserSaaS).where(UserSaaS.tenant_id == tenant_id)
        )

        return users, int(total or 0)

    async def update(
        self,
        db: AsyncSession,
        user_id: int,
        tenant_id: int,
        email: Optional[str] = None,
        rol: Optional[str] = None,
        password: Optional[str] = None,
    ) -> Optional[UserSaaS]:
        """Actualiza un usuario SaaS."""
        user = await self.get_by_id(db, user_id, tenant_id)
        if not user:
            return None

        if email is not None:
            user.email = email
        if rol is not None:
            user.rol = rol
        if password is not None:
            user.hashed_password = self.hash_password(password)

        await db.flush()
        await db.refresh(user)
        return user

    async def delete(self, db: AsyncSession, user_id: int, tenant_id: int) -> bool:
        """Elimina un usuario SaaS."""
        user = await self.get_by_id(db, user_id, tenant_id)
        if not user:
            return False

        await db.delete(user)
        await db.flush()
        return True


user_crud = UserSaaSCRUD()
