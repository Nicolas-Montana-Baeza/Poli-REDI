# Poli-REDI - Backend, API y seguridad ligera

## Objetivo del documento

Este documento registra el estado actual del backend y las mejoras recomendadas para reforzar seguridad ligera, permisos y pruebas del flujo de reservas.

## Estado actual observado

El backend usa Go, Fiber, Azure SQL Database y autenticacion con Microsoft Entra ID.

La estructura principal esta organizada por capas:

- `cmd/`: punto de entrada de la API.
- `internal/routes/`: registro de rutas.
- `internal/middleware/`: autenticacion y usuario local.
- `internal/handlers/`: entrada HTTP.
- `internal/services/`: reglas de negocio.
- `internal/repositories/`: acceso a Azure SQL.
- `internal/models/`: modelos de dominio.
- `internal/validators/`: validaciones reutilizables.

## Fortalezas actuales

- Las rutas internas requieren autenticacion.
- La creacion de reservas usa el usuario autenticado y no confia en `userId` enviado por el cliente.
- La cancelacion valida que el usuario sea propietario de la reserva o administrador.
- Los usuarios normales sin RUT no pueden crear reservas.
- La inscripcion a talleres usa el usuario autenticado, exige RUT a usuarios normales y valida cupos.
- La ruta de usuarios administrativos ya esta agrupada bajo `RequireAdmin`.
- La base de datos aplica reglas de conflicto para reservas, bloqueos y actividades programadas.
- Los errores de base de datos se traducen a mensajes mas legibles para reservas.
- CORS se configura por variable de entorno.

## Endpoints protegidos principales

- `GET /api/me`
- `PATCH /api/me/rut`
- `GET /api/resources`
- `GET /api/activities`
- `GET /api/notifications`
- `GET /api/workshops`
- `POST /api/workshops/:id/enroll`
- `GET /api/reservations/mine`
- `GET /api/reservations`
- `POST /api/reservations`
- `PATCH /api/reservations/cancel`

## Endpoints administrativos actuales

- `GET /api/users`

## Hallazgos de seguridad leve

### Separar disponibilidad de detalle administrativo

`GET /api/reservations` entrega reservas generales a usuarios autenticados. Para disponibilidad, un usuario normal no necesita recibir datos internos de reservas ajenas.

Mejora recomendada:

- Crear un endpoint de disponibilidad por fecha o rango.
- Para usuarios normales, devolver solo datos necesarios para pintar ocupacion: recurso, inicio, duracion y tipo de bloqueo.
- Reservar datos completos de reservas para administradores o para `GET /api/reservations/mine`.

### Middleware administrativo explicito

La ruta de usuarios ya esta agrupada bajo `RequireAdmin`. Aun asi, conviene mantener este patron para futuras rutas administrativas.

Mejora recomendada:

- Agrupar nuevas rutas administrativas bajo `RequireAdmin`.
- Mantener validacion de defensa adicional en handlers sensibles cuando corresponda.

### Logs de configuracion

El middleware de autenticacion imprime valores de configuracion de Entra ID al iniciar.

Mejora recomendada:

- Evitar imprimir tenant, client ID o issuer completos.
- Registrar solo si la configuracion existe o falta.
- No registrar tokens, cabeceras ni datos sensibles.

### Modo desarrollo

El modo `DEV_AUTH_ENABLED=true` es util para pruebas locales, pero debe quedar claramente protegido para despliegue.

Mejora recomendada:

- Validar en arranque que `DEV_AUTH_ENABLED=true` no se use con origenes productivos.
- Documentar que solo puede usarse localmente.
- Agregar una prueba o checklist de despliegue que confirme que esta desactivado en Azure.

## Pruebas backend recomendadas

La prioridad debe estar en reglas de negocio y permisos.

### Casos criticos de reservas

- Crear reserva valida.
- Rechazar reserva sin usuario autenticado.
- Rechazar usuario normal sin RUT.
- Rechazar recurso inexistente.
- Rechazar recurso inactivo.
- Rechazar recurso informativo.
- Rechazar recurso solo admin para usuario normal.
- Rechazar conflicto por recurso.
- Rechazar conflicto por usuario.
- Rechazar cruce con bloqueo.
- Rechazar cruce con actividad programada.

### Casos criticos de cancelacion

- Usuario cancela reserva propia.
- Admin cancela reserva ajena.
- Usuario normal no cancela reserva ajena.
- No se cancela una reserva inexistente.
- No se cancela dos veces una reserva ya cancelada.

### Casos de seguridad

- Usuario sin token recibe 401.
- Usuario normal no accede a rutas admin.
- Usuario bloqueado recibe 403.
- Modo dev sin cabeceras requeridas recibe 401.

### Casos criticos de talleres

- Listar talleres activos para usuario autenticado.
- Rechazar listado sin autenticacion.
- Inscribir usuario con RUT en taller con cupos.
- Rechazar usuario normal sin RUT.
- Rechazar taller inexistente o inactivo.
- Rechazar taller sin cupos.
- Rechazar inscripcion duplicada.

## Prioridades sugeridas

1. Endpoint de disponibilidad sanitizado.
2. Limpieza de logs de configuracion.
3. Pruebas backend para reservas, cancelacion y talleres.
4. Checklist productivo para `DEV_AUTH_ENABLED=false`.
