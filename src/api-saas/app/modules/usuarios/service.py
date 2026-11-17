"""Casos de uso del modulo `usuarios`.

Passwords: `bcrypt` directo (no passlib). Passlib 1.7.4 esta sin mantener desde
2020 y su backend bcrypt rompe con bcrypt 5.x al sondear el bug de wrap
(`detect_wrap_bug` -> ValueError). Antes de hashear se aplica SHA-256 + base64
para eliminar el limite de 72 bytes de bcrypt (mismo criterio que el esquema
`bcrypt_sha256` de passlib).

El hashing vive aqui (regla de negocio), no en el repositorio ni en el router.
"""
from __future__ import annotations

import base64
import hashlib
from typing import Optional

import bcrypt
from sqlalchemy.ext.asyncio import AsyncSession

from app.modules.usuarios.exceptions import EmailDuplicadoError, UsuarioNoEncontradoError
from app.modules.usuarios.models import UserSaaS
from app.modules.usuarios.repository import UserRepositoryPort, user_repository

#: Coste de bcrypt. El nucleo Go usa 10 para auth.users (datos demo); los
#: usuarios SaaS son credenciales propias, asi que sube a 12.
BCRYPT_ROUNDS = 12


def _prepare_password(password: str) -> bytes:
    """SHA-256 + base64: deja el secret en 44 bytes, dentro del limite de bcrypt."""
    return base64.b64encode(hashlib.sha256(password.encode("utf-8")).digest())


class UserService:
    """Operaciones de negocio sobre usuarios SaaS de un tenant."""

    def __init__(self, repository: UserRepositoryPort = user_repository) -> None:
        self._repository = repository

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
        *,
        check_unique: bool = False,
    ) -> UserSaaS:
        """Crea un usuario SaaS y hashea su password.

        `check_unique=True` (lo usa el router) traduce la violacion del UNIQUE
        `(tenant_id, email)` a un `EmailDuplicadoError` (400) antes de tocar la
        BD. Con `False` se deja que la BD aplique la constraint (tests de
        defensa en profundidad).
        """
        if check_unique and await self._repository.get_by_email(db, email, tenant_id):
            raise EmailDuplicadoError(email)

        return await self._repository.create(
            db,
            tenant_id=tenant_id,
            email=email,
            hashed_password=self.hash_password(password),
            rol=rol,
        )

    async def get_by_id(
        self, db: AsyncSession, user_id: int, tenant_id: int
    ) -> Optional[UserSaaS]:
        """Devuelve el usuario del tenant o `None`."""
        return await self._repository.get_by_id(db, user_id, tenant_id)

    async def get_by_email(
        self, db: AsyncSession, email: str, tenant_id: int
    ) -> Optional[UserSaaS]:
        """Devuelve el usuario por email dentro del tenant o `None`."""
        return await self._repository.get_by_email(db, email, tenant_id)

    async def obtener(self, db: AsyncSession, tenant_id: int, user_id: int) -> UserSaaS:
        """Devuelve el usuario del tenant o lanza `UsuarioNoEncontradoError`."""
        user = await self._repository.get_by_id(db, user_id, tenant_id)
        if user is None:
            raise UsuarioNoEncontradoError(user_id)
        return user

    async def list_by_tenant(
        self, db: AsyncSession, tenant_id: int, skip: int = 0, limit: int = 100
    ) -> tuple[list[UserSaaS], int]:
        """Lista usuarios del tenant y el total para paginar."""
        return await self._repository.list_by_tenant(db, tenant_id, skip, limit)

    async def update(
        self,
        db: AsyncSession,
        user_id: int,
        tenant_id: int,
        email: Optional[str] = None,
        rol: Optional[str] = None,
        password: Optional[str] = None,
        *,
        check_unique: bool = False,
    ) -> UserSaaS:
        """Actualiza email, rol o password de un usuario del tenant.

        Lanza `UsuarioNoEncontradoError` (404) si no existe en el tenant y
        `EmailDuplicadoError` (400) si `check_unique` y el email ya lo usa
        otro usuario del mismo tenant.
        """
        if check_unique and email is not None:
            clash = await self._repository.get_by_email(db, email, tenant_id)
            if clash and clash.id != user_id:
                raise EmailDuplicadoError(email)

        user = await self._repository.update(
            db,
            user_id=user_id,
            tenant_id=tenant_id,
            email=email,
            rol=rol,
            hashed_password=self.hash_password(password) if password is not None else None,
        )
        if user is None:
            raise UsuarioNoEncontradoError(user_id)
        return user

    async def delete(self, db: AsyncSession, user_id: int, tenant_id: int) -> bool:
        """Elimina un usuario del tenant o lanza `UsuarioNoEncontradoError`."""
        deleted = await self._repository.delete(db, user_id, tenant_id)
        if not deleted:
            raise UsuarioNoEncontradoError(user_id)
        return True


#: Instancia sin estado: se inyecta via `UserServiceDep`.
user_service = UserService()
