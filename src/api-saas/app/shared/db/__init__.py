"""Capa de persistencia comun.

Se mantiene deliberadamente sin imports para evitar ciclos:

- `app.shared.db.base`: declarativa + mixin de timestamps.
- `app.shared.db.session`: engine async, `get_db` y `get_tenant_db`.

Importar lo que se necesite desde el submodulo concreto.
"""
