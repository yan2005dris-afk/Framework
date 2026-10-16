# U2-T4 (frontend) — Cookie httpOnly + servicios con credenciales

Migrar el JWT de localStorage a cookie `httpOnly` para cerrar el riesgo XSS sobre el token. Hacer una sola rama (`feature/u2-t4-frontend-cookie-auth`) con backend + frontend en dos work-unit commits. Cierra los pendientes de U1-T4 (services/http.ts + apiFetch + ApiError tipado) y de U1-T3 (lint limpio en AuthContext sin `eslint-disable` por fake-jwt-token).

## Specs

- **S1-FE** — `services/http.ts` único: `apiFetch` con `credentials: "include"`, `Authorization: Bearer` opcional (override), `ApiError` tipado con `{status, message, details?}`. Si la respuesta es 401 dispara `auth.onUnauthorized()` para limpiar sesión.
- **S2-FE** — `AuthContext` ya no persiste token en localStorage. Persiste solo `{email, rol}` en una clave `mc_session`. El token vive únicamente en la cookie `auth_token` httpOnly del backend.
- **S3-FE** — Al montar la app, `AuthContext` llama a `GET /api/me` (que lee la cookie) para hidratar sesión; mientras tanto `loading: true`. Esto permite recargar la página sin perder sesión.
- **S4-FE** — `Login.tsx` ya no guarda token en localStorage; solo actualiza `user` y deja que el backend setee la cookie. Logout llama a `POST /api/logout` (que limpia la cookie en el backend) y limpia estado.
- **S5-FE** — `services/api.ts` refactorizado: usa `apiFetch` en vez de `fetch` directo. Tipos `LoginResponse` ajustado: ahora es `{email, rol, expires}` sin `token` (el token viaja solo en cookie).
- **S6-BE** — `Login` setea cookie `auth_token` con `HTTPOnly=true`, `SameSite=Lax`, `Secure=(APP_ENV != "dev")`, `Path=/`, `Expires=now+ttl`. El body JSON ya no contiene `token`.
- **S7-BE** — `RequireAuth` lee el token desde la cookie `auth_token` si no hay header `Authorization` (mantiene el header como override para integraciones).
- **S8-BE** — Endpoint `POST /api/logout` que limpia la cookie (devuelve `{}`).
- **S9-BE** — Endpoint `GET /api/me` protegido por `RequireAuth` que retorna `{id, email, rol}` desde `c.Locals`.
- **S10-BE** — CORS reconfigurado: `AllowOrigins` deja de ser `"*"` y pasa a leer la env `CORS_ORIGINS` (default `"http://localhost:5173"`). `AllowCredentials: true`. Si `APP_ENV=prod`, exigir `AllowOrigins` no wildcard.
- **S11-BE** — Tests: middleware/auth_test.go agrega caso `RequireAuth_DesdeCookie_Valida` (cookie en lugar de header); authController_test.go agrega casos para `Login` (cookie setteada), `Logout` (cookie max-age=0), `Me` (200 con locals, 401 sin locals).

## Tasks (work-unit commits)

- **F1 · `feat(backend): cookie httpOnly auth + /api/me + CORS credentials`** — autocontenido. Login setea cookie, RequireAuth la lee, endpoint `Me` y `Logout`. CORS con credenciales y origen específico. No toca frontend todavía. Cubre S6-S10 y los tests de S11.

  Archivos:
  - `Backend/controllers/authController.go`: agrega `appEnv` al constructor; `Login` setea cookie + body sin token; nuevos métodos `Me` y `Logout`.
  - `Backend/middleware/auth.go`: lee cookie si no hay header.
  - `Backend/main.go`: CORS con `AllowCredentials` y `CORS_ORIGINS` de env; pasa `appEnv` a `NewAuthController`.
  - `Backend/routes/routes.go`: registra `POST /api/logout` (público, sin middleware porque settea cookie MaxAge=-1) y `GET /api/me` (protegido).
  - `Backend/middleware/auth_test.go`: caso `RequireAuth_DesdeCookie_Valida` y `RequireAuth_Header_Gana_SobreCookie`.
  - `Backend/controllers/auth_test.go` (nuevo): tests de `Login` (cookie presente), `Logout` (clear), `Me` (200 con locals; 401 sin locals).

- **F2 · `feat(frontend): apiFetch with credentials, AuthContext cookie-based, /api/me hydration`** — autocontenido. Cierra los pendientes de U1-T3/U1-T4 del checklist.

  Archivos:
  - `Frontend/src/services/http.ts` (nuevo): `apiFetch`, `ApiError`, `apiBaseURL` (lee `VITE_API_URL` con fallback).
  - `Frontend/src/services/api.ts` (refactor): reescrito para usar `http.ts`. `login` retorna `{email, rol, expires}` sin token. Nuevo `me()`, `logout()`.
  - `Frontend/src/Context/AuthContext.tsx` (rewrite): estado `user`+`loading`; al mount llama `api.me()`; storage único `mc_session` con `{email, rol}`; sin `eslint-disable`.
  - `Frontend/src/Components/Auth/Login/Login.tsx`: ya no pasa token a `login(user)`.

- **F3 · `docs(odd): feature doc for u2-t4-frontend-cookie-auth task unit`** — registro del trabajo.

## Log

- **L1** (usuario, 2026): "ok el plan el frontend es con localStorage no?" — consulta sobre el storage.
- **L2** (usuario): "1. una sola rama | 2. hacerlo directamente con cookie httpOnly" — decisión de no usar localStorage y migrar a cookie httpOnly. Esto requiere cambios en backend para setear/limpiar/parsear la cookie y reconfigurar CORS, además del frontend.
- **L3** (implementación, pendiente): commits F1, F2, F3 en `feature/u2-t4-frontend-cookie-auth` desde `main`.
