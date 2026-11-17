"""Repositorio de `UserSaaS` con aislamiento por tenant.

- **Toda** query filtra por `tenant_id`: un usuario de otro tenant nunca se
  lee, se modifica ni se borra, aunque RLS no exista (sqlite en tests).
- Se usa `flush()` y no `commit()`: la transaccion del request la cierra
  `get_tenant_db`, para que el GUC de RLS siga activo.
- El hashing de passwords NO vive aca: es una regla de negocio del servicio
  (`app/modules/usuarios/service.py`). Este repositorio solo persiste el
  `hashed_password` que recibe.
"""
from __future__ import annotations

from typing import Optional, Protocol

from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.usuarios.models import UserSaaS


class UserRepositoryPort(Protocol):
    """Contrato de persistencia que consume `UserService`."""

    async def create(
        self,
        db: AsyncSession,
        tenant_id: int,
        email: str,
        hashed_password: str,
        rol: str = "member",
    ) -> UserSaaS: ...

    async def get_by_id(
        self, db: AsyncSession, user_id: int, tenant_id: int
    ) -> Optional[UserSaaS]: ...

    async def get_by_email(
        self, db: AsyncSession, email: str, tenant_id: int
    ) -> Optional[UserSaaS]: ...

    async def list_by_tenant(
        self, db: AsyncSession, tenant_id: int, skip: int = 0, limit: int = 100
    ) -> tuple[list[UserSaaS], int]: ...

    async def update(
        self,
        db: AsyncSession,
        user_id: int,
        tenant_id: int,
        email: Optional[str] = None,
        rol: Optional[str] = None,
        hashed_password: Optional[str] = None,
    ) -> Optional[UserSaaS]: ...

    async def delete(self, db: AsyncSession, user_id: int, tenant_id: int) -> bool: ...


class UserRepository:
    """Operaciones de persistencia para `UserSaaS`."""

    async def create(
        self,
        db: AsyncSession,
        tenant_id: int,
        email: str,
        hashed_password: str,
        rol: str = "member",
    ) -> UserSaaS:
        """Crea un usuario SaaS en un tenant."""
        user = UserSaaS(
            tenant_id=tenant_id,
            email=email,
            hashed_password=hashed_password,
            rol=rol,
        )
        db.add(user)
        await db.flush()
        await db.refresh(user)
        return user

    async def get_by_id(
        self, db: AsyncSession, user_id: int, tenant_id: int
    ) -> Optional[UserSaaS]:
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
        hashed_password: Optional[str] = None,
    ) -> Optional[UserSaaS]:
        """Actualiza un usuario SaaS del tenant. `None` si no existe."""
        user = await self.get_by_id(db, user_id, tenant_id)
        if not user:
            return None

        if email is not None:
            user.email = email
        if rol is not None:
            user.rol = rol
        if hashed_password is not None:
            user.hashed_password = hashed_password

        await db.flush()
        await db.refresh(user)
        return user

    async def delete(self, db: AsyncSession, user_id: int, tenant_id: int) -> bool:
        """Elimina un usuario SaaS del tenant. `False` si no existe."""
        user = await self.get_by_id(db, user_id, tenant_id)
        if not user:
            return False

        await db.delete(user)
        await db.flush()
        return True


#: Instancia sin estado: se inyecta via `UserServiceDep`.
user_repository = UserRepository()
