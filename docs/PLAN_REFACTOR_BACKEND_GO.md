# Plan de refactorización — Backend Go hacia bounded contexts

| Campo | Valor |
|---|---|
| Estado | **Propuesta de diseño. No implementado.** |
| Rama | `yandris-rivera/AA1-5` |
| Fecha | 2026-11-17 |
| Alcance | `Backend/` (Go + Fiber + pgx + Redis) |
| Fuera de alcance | `auth-service/`, `gateway/`, `Frontend/`, `fastAPI_SaaS/` |
| Relación con la actividad AA1-5 | Documento de referencia. La implementación de la actividad se hace en `fastAPI_SaaS/`, no acá. |

---

## 1. Diagnóstico (con evidencia del código actual)

El Backend Go funciona, pero está organizado por **capas técnicas globales**: `controllers/`, `repository/`, `models/`, `routes/`, `middleware/`. Eso genera cinco problemas concretos y medibles.

| Archivo | Tamaño | Problema |
|---|---|---|
| `repository/interfaces.go` | 47 líneas | `DatabaseRepository` tiene **24 métodos** de 6 dominios distintos (auth, catálogo, pedidos, red, comisiones, reportes). Es una *god interface*. |
| `repository/repository.go` | ≈1.280 líneas | Mezcla SQL, transacciones, reglas de negocio, restauración de stock y reversión de comisiones. Además expone un singleton oculto (`defaultRepo`) y funciones a nivel de paquete. |
| `controllers/checkoutController.go` | 129 líneas | Orquesta la transacción, dispara una goroutine para comisiones e invalida caché. La capa HTTP conoce reglas de negocio y efectos asíncronos. |
| `controllers/pedidoController.go` | 166 líneas | Parseo HTTP + autorización + validación de estados + mapeo de errores, todo junto y sin tests unitarios de la regla. |
| `models/models.go` | 243 líneas | Entidades de dominio, DTOs de request y `APIError`, que **importa Fiber** (`Write(c *fiber.Ctx)`). El paquete "modelo" depende del framework web. |
| `routes/routes.go` | 103 líneas | `SetupRoutes` recibe **11 parámetros posicionales**. |
| `main.go` | 112 líneas | Composition root con 8 controladores y 2 repositorios instanciados a mano. |

Síntomas derivados:

1. **Acoplamiento total entre dominios.** Cambiar un método de pedidos obliga a tocar la interfaz y **todos** los mocks: `MockDatabaseRepository`, `stubDBRepo`, etc.
2. **Reglas de negocio no testeables sin Postgres.** La transición de estados y la reversión de stock viven en SQL.
3. **Un paquete de modelos que no se puede reutilizar** porque depende de Fiber.
4. **Tests frágiles.** Los mocks deben implementar 24 métodos para probar uno.
5. **Crecimiento cuadrático del costo.** Cada dominio nuevo agranda la interfaz común y el repositorio común.

---

## 2. Principios del plan

1. **Strangler fig, no big bang.** La estructura vieja y la nueva conviven. El código nuevo no importa el viejo.
2. **Un contexto por vez**, empezando por `pedidos` (es el que tiene casos de uso más claros).
3. **Contrato HTTP intacto.** Mismos paths, mismos códigos, mismo JSON. El `Frontend` no se entera.
4. **Cada fase deja la rama verde**: `go build ./...`, `go vet ./...`, `go test ./...`.
5. **La atomicidad no se rompe.** Ver §6.3: la transacción sigue siendo una sola.
6. **Aplicación selectiva.** Donde no hay lógica (p. ej. health), no se layeriza.

---

## 3. Estructura objetivo

```text
Backend/
├── cmd/
│   └── api/
│       └── main.go                 # composition root: único lugar que conoce todo
├── internal/
│   ├── pedidos/
│   │   ├── pedido.go               # entidad + Validar()
│   │   ├── estado.go               # estados y transiciones
│   │   ├── errors.go               # errores de dominio
│   │   ├── application/
│   │   │   ├── ports.go            # Repository, Reloj, etc.
│   │   │   ├── listar_mios.go
│   │   │   ├── listar_admin.go
│   │   │   ├── obtener_detalle.go
│   │   │   └── cambiar_estado.go
│   │   └── adapters/
│   │       ├── httpapi/
│   │       │   ├── handler.go
│   │       │   └── dto.go
│   │       └── postgres/
│   │           └── repository.go
│   ├── catalogo/
│   ├── comisiones/
│   ├── identidad/
│   ├── reportes/
│   └── platform/
│       ├── postgres/               # pool, helpers de transacción
│       ├── redis/                  # caché
│       ├── httperr/                # APIError sin Fiber
│       ├── httpserver/             # Fiber: middlewares comunes, CORS, logger
│       └── logging/
├── migrations/                     # mover database/migrations/*
├── database/init.sql               # se mantiene mientras dure la migración
└── go.mod
```

**Regla de dependencias:**

```text
adapters → application → domain (raíz del contexto)
platform: lo usan todos, no conoce ningún dominio
```

- `pedidos` (raíz) no importa `application` ni `adapters`.
- `application` no importa Fiber, pgx ni Redis.
- `adapters` implementa los puertos de `application`.
- `cmd/api/main.go` es el único que conoce todas las piezas.

---

## 4. Contextos y contenido

| Contexto | Qué contiene | Qué NO debe conocer |
|---|---|---|
| `pedidos` | Pedido, ítems, estados, checkout, cancelación, restauración de stock | SQL de catálogo, Fiber, Redis |
| `catalogo` | Producto, stock, precios, categorías | Pedidos, comisiones |
| `comisiones` | Ledger, niveles, liquidaciones, reversión | HTTP de pedidos |
| `identidad` | Login, registro, bcrypt, JWT, sesión | Reglas de pedidos |
| `reportes` | Read models / KPIs (solo lectura) | Transacciones de escritura |
| `platform` | Pool PostgreSQL, Redis, logs, errores HTTP, servidor | Reglas de negocio |

`pedido_item` **no** es un contexto aparte: es parte de `pedidos`.

---

## 5. Fases

### Fase 0 — Línea base y red de seguridad (≈0.5 día)

- [ ] `go build ./...`, `go vet ./...`, `go test ./...` y guardar la salida como línea base.
- [ ] Capturar **golden files** de las respuestas actuales:
  - `GET /api/pedidos/mios`
  - `GET /api/pedidos?estado=pendiente`
  - `GET /api/pedidos/:id`
  - `PATCH /api/pedidos/:id/estado`
- [ ] Documentar el contrato vigente (paths, status codes, shape JSON) en este mismo archivo o en `docs/`.
- [ ] Agregar `golangci-lint` y `go-cleanarch` al CI (aún sin reglas estrictas).

**DoD:** existe una forma objetiva de comparar "antes vs después" sin depender de la vista humana.

---

### Fase 1 — `internal/pedidos` (≈2–3 días) ← **piloto**

Es el contexto que sirve de plantilla para el resto.

**Mapeo origen → destino**

| Hoy | Destino |
|---|---|
| `models.Pedido`, `PedidoItem`, `PedidoDetalleItem`, `PedidoConDetalle` | `internal/pedidos/pedido.go` |
| Mapa de estados en `pedidoController.ActualizarEstado` | `internal/pedidos/estado.go` |
| `repository.ListarPedidosMios / ListarPedidosAdmin / ObtenerPedidoDetalle / ActualizarEstadoPedido` | `internal/pedidos/adapters/postgres/repository.go` |
| `controllers/pedidoController.go` | `internal/pedidos/adapters/httpapi/handler.go` |
| `models.ActualizarEstadoPedidoRequest` | `internal/pedidos/adapters/httpapi/dto.go` |
| `models.APIError` | `internal/platform/httperr` (sin Fiber en el tipo) |
| `routes.SetupRoutes` (parte pedidos) | `internal/pedidos/adapters/httpapi/routes.go` |

**Tareas**

- [ ] Crear dominio: `Pedido`, `Estado`, `Validar()`.
- [ ] Crear `estado.go` con los 5 estados y la regla **actual**: `cancelado` es terminal (ver §6.1).
- [ ] Crear `application/ports.go` con una interfaz **pequeña y de pedidos**.
- [ ] Crear los 4 casos de uso.
- [ ] Crear el adaptador Postgres moviendo el SQL tal cual (sin reescribir queries).
- [ ] Crear el adaptador HTTP moviendo el parseo y el mapeo de errores.
- [ ] Registrar las mismas rutas desde `main.go`.
- [ ] Borrar los métodos de pedidos del `DatabaseRepository` viejo y de `pedidoController` (una vez verificado).

**DoD:** `GET/PATCH` de pedidos responden exactamente igual; los tests nuevos cubren transición inválida sin tocar Postgres; la interfaz vieja ya no tiene métodos de pedidos.

---

### Fase 2 — `internal/catalogo` (≈2 días)

- [ ] Producto, stock, validaciones de precio/stock.
- [ ] Casos de uso: listar, detalle, listar admin, crear, actualizar, eliminar lógico.
- [ ] Mover la invalidación de caché (`productos:all`) a un `CachePort` del contexto.
- [ ] `prodController` desaparece.

**Riesgo principal:** la caché. Hoy el controller llama a `cacheRepo.Del(...)` mezclado con la lógica. El caso de uso debe depender de un puerto `Cache` y el adaptador Redis vive en `platform/redis`.

---

### Fase 3 — `internal/comisiones` (≈2–3 días)

Es la fase más delicada por el ledger y la goroutine del checkout.

- [ ] Modelar `Comision`, `Liquidacion`, `ResumenComisiones`.
- [ ] Casos de uso: registrar comisiones de un pedido, resumen por referido, liquidar, listar.
- [ ] Reemplazar la goroutine de `checkoutController` por **una llamada explícita a un puerto** `ComisionesPort` (síncrona, idempotente) o por un job con reintento.
- [ ] Eliminar `ActualizarComisiones` o justificar por qué sobrevive (hoy hay dos caminos que tocan comisiones: `ActualizarComisiones` y `RegistrarComisionesPedido`; conviene unificarlos o documentar la diferencia).

**Decisión pendiente de producto:** si la comisión se calcula en la misma transacción del pedido (consistencia fuerte) o después (eventual). Hoy es "después, en goroutine", lo que significa que un pedido puede existir sin comisiones si el proceso muere.

---

### Fase 4 — `internal/identidad` (≈1–2 días)

- [ ] `authController` → `internal/identidad/adapters/httpapi`.
- [ ] `bcrypt.CompareHashAndPassword` sale del controller y pasa a un caso de uso.
- [ ] `middleware.RequireAuth` se mueve a `internal/identidad/adapters/http` o a `platform/httpserver`, según si se considera regla de identidad o infraestructura.
- [ ] El proxy hacia `auth-service` se modela como un puerto (`AuthServicePort`).

**Nota:** este contexto se puede dejar para el final sin bloquear nada.

---

### Fase 5 — `internal/reportes` (≈1 día)

- [ ] Solo read models: `ReporteResumen`, `TopProducto`, `TopReferido`, `DeudaDistribuidor`.
- [ ] **Sin dominio**: son queries de lectura. Un `application/service.go` con 4 métodos alcanza.
- [ ] Considerar vistas SQL o queries dedicadas en lugar de reutilizar el repositorio transaccional.

---

### Fase 6 — Limpieza y enforcement (≈1 día)

- [ ] Borrar `models/`, `controllers/`, `repository/` si quedaron vacíos.
- [ ] `routes.SetupRoutes` con 11 parámetros → un registro por contexto:

```go
pedidoshttp.Register(api, pedidosHandler)
catalogohttp.Register(api, catalogoHandler)
```

- [ ] Activar `go-cleanarch` en CI con estas reglas:
  - `internal/<ctx>` no importa `internal/<ctx>/adapters`
  - `internal/<ctx>` no importa `internal/<ctx>/application`
  - `application` no importa Fiber, pgx, go-redis
  - `internal/<ctx>` no importa `internal/<otro-ctx>`
- [ ] Escribir un ADR corto: `docs/adr/0001-bounded-contexts.md`.

---

## 6. Detalle técnico de la Fase 1 (`pedidos`)

### 6.1 Dominio

```go
// internal/pedidos/pedido.go
package pedidos

type Pedido struct {
    ID            int
    UsuarioID     int
    NombreCliente string
    EmailCliente  string
    MetodoPago    string
    Total         float64
    Estado        Estado
    CreadoAt      string
}
```

```go
// internal/pedidos/estado.go
package pedidos

type Estado string

const (
    EstadoPendiente  Estado = "pendiente"
    EstadoPreparando Estado = "preparando"
    EstadoEnviado    Estado = "enviado"
    EstadoEntregado  Estado = "entregado"
    EstadoCancelado  Estado = "cancelado"
)

func ParseEstado(s string) (Estado, error)
func (e Estado) Valido() bool
func (p Pedido) EsCancelable() bool  // !p.Estado.EsTerminal()
```

**Importante:** hoy el controller acepta cualquiera de los 5 estados y el repositorio solo rechaza `cancelado → cancelado`. **La Fase 1 debe preservar ese comportamiento exacto.** La máquina de estados estricta (`pendiente → preparando → enviado → entregado`) es una mejora funcional, no un refactor, y requiere decisión de producto + tests propios.

### 6.2 Casos de uso

```go
// internal/pedidos/application/listar_mios.go
package application

type ListarMios struct{ repo Repository }

func (uc *ListarMios) Execute(ctx context.Context, in ListarMiosInput) ([]pedidos.Pedido, error)
```

Cuatro archivos: `listar_mios.go`, `listar_admin.go`, `obtener_detalle.go`, `cambiar_estado.go`. **Un archivo por caso de uso, un solo paquete.** No se crea una carpeta por caso.

### 6.3 Puertos y atomicidad (decisión clave)

`ActualizarEstadoPedido` hoy hace, **dentro de una transacción**:

1. `SELECT ... FOR UPDATE` sobre el pedido.
2. Si el nuevo estado es `cancelado`: devuelve stock de cada ítem.
3. Marca comisiones `pendiente` como `revertida`.
4. `UPDATE pedidos SET estado`.

Si se parte en "caso de uso + varios repositorios", se pierde la atomicidad. Dos opciones:

**Opción A (recomendada para la Fase 1):** el caso de uso decide *si* se puede cancelar; el adaptador Postgres ejecuta el cambio completo en una transacción.

```go
type Repository interface {
    ListarMios(ctx context.Context, usuarioID, limit, offset int) ([]pedidos.Pedido, error)
    ListarAdmin(ctx context.Context, estado string, limit, offset int) ([]pedidos.Pedido, error)
    ObtenerDetalle(ctx context.Context, pedidoID int) (*pedidos.PedidoConDetalle, error)
    CambiarEstado(ctx context.Context, pedidoID int, nuevo pedidos.Estado) error
}
```

Ventaja: se mueve la **regla de negocio** (qué estados son válidos, cuándo se puede cancelar) al dominio sin romper la transacción. Costo: el adaptador sigue teniendo efectos secundarios.

**Opción B (evolución posterior):** puerto `UnitOfWork` con `WithinTx(ctx, fn func(tx Repos) error)` y repositorios separados por agregado. Más correcto a largo plazo, mucho más trabajo y más riesgo. **No hacerla en la Fase 1.**

### 6.4 Adaptador Postgres

Mover las queries **sin reescribirlas**. Preservar:

- `pgx.ErrNoRows` → `pedidos.ErrNoEncontrado`.
- El fallback de `getPool()` a `config.DB` **no** se traslada: el adaptador recibe el pool por constructor. `SetupDefaultRoutes` (que hoy instancia `&repository.PostgresRepository{}` con pool nil) debe reemplazarse por un constructor de test con `nil` explícito o por fakes.

### 6.5 Adaptador HTTP y errores

El handler no valida reglas ni arma SQL. Solo:

- lee `Locals` (usuario, rol) y params;
- llama al caso de uso;
- traduce error de dominio → status.

| Error de dominio | HTTP |
|---|---|
| `ErrNoEncontrado` | 404 |
| `ErrEstadoInvalido` | 400 |
| `ErrNoAutorizado` (pedido ajeno) | 403 |
| error de infraestructura | 500 (+ `helpers.SanitizeErrorDetails`) |

`APIError` se mueve a `internal/platform/httperr` y **deja de importar Fiber**: el tipo no debe tener `Write(c *fiber.Ctx)`. La escritura se hace con una función del adaptador HTTP.

### 6.6 Wiring

```go
// cmd/api/main.go
pedidosRepo := pedidospg.New(dbPool)

listarMios := application.NewListarMios(pedidosRepo)
listarAdmin := application.NewListarAdmin(pedidosRepo)
obtenerDetalle := application.NewObtenerDetalle(pedidosRepo)
cambiarEstado := application.NewCambiarEstado(pedidosRepo)

pedidosHandler := pedidoshttp.New(listarMios, listarAdmin, obtenerDetalle, cambiarEstado)
pedidoshttp.Register(api, pedidosHandler, requireAuth)
```

### 6.7 Tests

- Tabla de casos para `estado.go` (sin mocks, sin Fiber, sin DB).
- Fakes escritos a mano para `Repository` (4 métodos, no 24).
- Los tests HTTP existentes (`pedidoController_test.go`) se reescriben contra el handler nuevo y se mantienen como regresión.

---

## 7. Riesgos y mitigaciones

| Riesgo | Impacto | Mitigación |
|---|---|---|
| Cambiar el shape JSON que consume el Frontend | Alto | Golden files + comparación byte a byte en la Fase 0. |
| Perder atomicidad al cancelar (stock + comisiones) | Alto | Opción A de §6.3: una sola transacción en el adaptador. |
| Doble registro de comisiones | Alto | La goroutine ya tiene el flag `replay`; al moverla, mantener la idempotencia y agregar test. |
| `getPool()` con fallback global usado por tests | Medio | Eliminar el fallback al pasar el pool por constructor; ajustar tests. |
| Mocks existentes (24 métodos) | Medio | Se reemplazan por fakes pequeños a medida que cada contexto migra. |
| Forzar `tenant_id` en el Go "por simetría" con FastAPI | Medio | **No hacerlo.** El Backend Go es single-tenant; el multi-tenant es del SaaS. |
| Refactor eterno sin cierre | Alto | Cada fase tiene DoD y se mergea por separado. |

---

## 8. Enforcement (para que no se degrade)

`go-cleanarch` + `depguard` en `.golangci.yml` por módulo, como hace Three Dots Labs. Sin esto, la estructura es una convención de caballeros y en seis meses `domain` importa `pgx`.

---

## 9. Anti-objetivos (qué NO hacer en este refactor)

- No microservicios. Es un **modular monolith**.
- No `internal/domain`, `internal/models`, `internal/services`, `internal/adapters` globales.
- No una carpeta por caso de uso si el caso tiene 20 líneas.
- No una interfaz por struct. Interfaces pequeñas, definidas por el consumidor.
- No tocar `auth-service`, `gateway` ni `Frontend`.
- No cambiar paths ni contratos JSON.
- No introducir `tenant_id` en el Backend Go.
- No reescribir queries SQL "de paso".
- No implementar esto en la rama de la actividad AA1-5: va en una rama aparte y con PRs por fase.

---

## 10. Relación con la actividad AA1-5 (FastAPI)

| | Backend Go actual | FastAPI SaaS (AA1-5) |
|---|---|---|
| Organización | Capas técnicas globales (`controllers`, `repository`, `models`) | API-First: router + schemas Pydantic + dependencias |
| Validación | Manual en controller + SQL | Pydantic v2 en el borde (`Field(gt=0)`) |
| Multi-tenant | No (single-tenant) | RLS + `app.current_tenant` |
| Contrato | Fiber + JSON manual | OpenAPI generado automáticamente |
| Evolución propuesta | Bounded contexts (este documento) | Módulo vertical `app/modules/pedidos` |
| Ejecución | **Documentado, no implementado** | **Se implementa en esta actividad** |

La comparación del informe se apoya justamente en esto: el Backend Go es un monolito por capas con reglas en el SQL; FastAPI permite poner la validación en el borde y derivar el contrato. Y el plan de evolución del Go es el mismo principio (bounded contexts), aplicado con las herramientas idiomáticas de Go.

---

## 11. Estimación

| Fase | Esfuerzo |
|---|---|
| 0 — Base y golden files | 0.5 día |
| 1 — `pedidos` (piloto) | 2–3 días |
| 2 — `catalogo` | 2 días |
| 3 — `comisiones` | 2–3 días |
| 4 — `identidad` | 1–2 días |
| 5 — `reportes` | 1 día |
| 6 — Limpieza + CI + ADR | 1 día |
| **Total** | **≈10–13 días de desarrollo** |

Se puede cortar en la Fase 1 y dejar el resto como deuda planificada: `pedidos` sola ya demuestra el patrón y no rompe nada.

---

## 12. Referencias

- Three Dots Labs — *How to implement Clean Architecture in Go*: https://threedots.tech/post/introducing-clean-architecture/
- Three Dots Labs — *Wild Workouts*: https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example
- `go-cleanarch` (Robert Laszczak): https://github.com/roblaszczak/go-cleanarch
- Redowan Delowar — *Let the domain guide your application structure*: https://rednafi.com/go/app-structure/
- LdeJ — *A Collection of Structuring Go*: https://ldej.nl/post/structuring-go/
- Dave Cheney — *SOLID Go Design*: https://dave.cheney.net/2016/08/20/solid-go-design
