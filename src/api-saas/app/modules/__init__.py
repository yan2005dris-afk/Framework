"""Bounded contexts de la aplicacion.

Cada subpaquete es un modulo de feature autocontenido:

- `pedidos`: pedidos tenant-scoped con maquina de estados.
- `usuarios`: CRUD de users_saas (bcrypt + aislamiento por tenant).
- `billing`: planes, suscripciones y webhook de Stripe (stub).

Contrato de dependencias (verificado con import-linter):

- los modulos no se importan entre si;
- todos pueden importar `app.shared`;
- dentro de cada modulo el flujo es `router -> service -> repository`.
"""
