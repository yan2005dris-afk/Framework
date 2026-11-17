"""Infraestructura transversal compartida por todos los modulos de feature.

Aqui vive lo que no pertenece a un bounded context concreto:

- `config`: settings tipados (pydantic-settings).
- `logging`: formatter JSON de una linea.
- `db`: `Base`/`TimestampMixin` + engine async y dependencias de sesion.
- `tenancy`: identidad del request, guards de scope y modelo `Tenant`.
- `auth_proxy`: proxy HTTP al auth-service.
- `health`: endpoints publicos de salud.

Regla de dependencias (verificada con import-linter):
`app.shared` NO importa `app.modules`; los modulos de feature si pueden
importar `app.shared`, nunca al reves.
"""
