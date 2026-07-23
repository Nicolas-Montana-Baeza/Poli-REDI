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
- La ruta de usuarios y la consulta completa de reservas usan `RequireAdmin`.
- La imagen de un recurso se actualiza con ruta administrativa protegida por `RequireAdmin`.
- La disponibilidad cuenta con endpoint sanitizado para ocultar datos personales de reservas ajenas a usuarios normales.
- La disponibilidad combina reservas y actividades institucionales activas; para usuarios normales sanitiza creador y titulo interno de la actividad.
- Las reservas sobre recursos `OPEN_USE` no bloquean otros usos del mismo recurso.
- La creacion de reservas rechaza cruces con talleres activos asociados al mismo recurso.
- La base de datos aplica reglas de conflicto para reservas, bloqueos y actividades programadas.
- Los errores de base de datos se traducen a mensajes mas legibles para reservas.
- CORS se configura por variable de entorno.
- La cancelacion rechaza reservas cuyo termino ya paso.

## Endpoints protegidos principales

- `GET /api/me`
- `PATCH /api/me/rut`
- `GET /api/resources`
- `GET /api/activities`
- `GET /api/notifications`
- `GET /api/workshops`
- `POST /api/workshops/:id/enroll`
- `GET /api/availability/reservations`
- `GET /api/reservations/mine`
- `POST /api/reservations`
- `PATCH /api/reservations/cancel`
- `GET /api/group-reservations/:code`: progreso agregado, sin datos nominales.
- `PUT /api/group-reservations/:code/confirmation`: confirmar o reconfirmar la participacion del usuario autenticado.
- `DELETE /api/group-reservations/:code/confirmation`: retirar la participacion propia; el propietario no puede retirarse.
- `PATCH /api/reservations/:id/target-participants`: cambiar el objetivo de una solicitud grupal propia hasta su limite de confirmacion inclusive.
- `GET /api/reservations/:id/join-code`: recuperar el codigo, solo propietario.
- `POST /api/reservations/:id/join-code/rotate`: rotar codigo/hash/secreto, solo propietario.

## Endpoints administrativos actuales

- `GET /api/users`
- `GET /api/reservations`
- `PATCH /api/resources/:id/image`
- `GET /api/admin/reservations/:id/participants`: detalle nominal, protegido por `RequireAdmin`.

## Hallazgos de seguridad leve

### Separar disponibilidad de detalle administrativo

`GET /api/availability/reservations` entrega disponibilidad a usuarios autenticados. Para usuarios normales, el handler elimina `userId`, nombre, email y RUT de reservas ajenas. `GET /api/reservations` queda reservado para administradores.

Mejora recomendada:

- Agregar filtros por fecha o rango para no devolver mas datos de los necesarios.
- Incorporar bloqueos al mismo contrato; las actividades programadas ya estan integradas.

### Middleware administrativo explicito

La ruta de usuarios, la consulta completa de reservas y la actualizacion de imagenes de recursos ya usan `RequireAdmin`. Aun asi, conviene mantener este patron para futuras rutas administrativas.

Mejora recomendada:

- Agrupar nuevas rutas administrativas bajo `RequireAdmin`.
- Mantener validacion de defensa adicional en handlers sensibles cuando corresponda.

### Logs de configuracion

La limpieza de logs de configuracion ya fue implementada en `SEC-003`: el middleware no imprime tenant, client ID, issuer, tokens ni correos en errores internos.

Pendiente relacionado:

- Mantener esta regla al agregar nuevos logs y cubrirla mediante revision de respuestas publicas (`SEC-005`).

### Modo desarrollo

El modo `DEV_AUTH_ENABLED=true` es util para pruebas locales, pero debe quedar claramente protegido para despliegue.

Estado:

- La guia de despliegue exige desactivarlo en Azure y el checklist fue validado (`SEC-004`).
- Un bloqueo automatico de arranque queda como endurecimiento posterior si el sistema pasa de demo a operacion institucional.

### Contrato temporal de reservas

Azure SQL guarda reservas en `DATETIME2` como hora local institucional. Backend y frontend usan `APP_TIMEZONE=America/Santiago`; los requests con offset se convierten a la zona de negocio y los requests sin offset se interpretan directamente en esa zona. La API serializa respuestas RFC 3339 con el offset real de Chile.

Estado para MVP 1:

- `RES-009` esta implementado y en revision por despliegue.
- Queda comparar una reserva de hora conocida en local y online antes de cerrar definitivamente la tarea.

### Estado y limites controlados por servidor

El contrato publico de creacion ya no acepta `status`; el servicio asigna `PENDING` cuando la politica versionada clasifica el recurso como grupal y `CONFIRMED` en los demas casos. La cancelacion solo permite transiciones desde estados activos. La API valida duraciones permitidas, paso de 15 minutos, apertura 08:00, cierre 22:00 y termino completo dentro de la jornada.

Estado para MVP 1:

- `RES-010` y `RES-011` estan implementados y en revision por despliegue.
- Queda ejecutar la prueba manual desplegada de request manipulado, limites de jornada y ultima reserva valida.

### Detalles internos en respuestas HTTP

Varios handlers incluyen `err.Error()` en respuestas. Esto debe considerarse informacion de diagnostico interno aunque el frontend no la muestre.

Accion requerida para MVP 1:

- Implementar respuestas publicas estables y logs internos sanitizados (`SEC-005`).

## API de politicas de reserva

Estado: IMPLEMENTADA y VERIFICADA LOCALMENTE. Todas las rutas exigen autenticacion y las dos rutas bajo `/api/admin` tambien pasan por `RequireAdmin`.

- `GET /api/reservation-policy/current`: DTO publico minimo de condiciones operativas, sin identificador, autoria, vigencias ni fechas de auditoria.
- `GET /api/admin/reservation-policies`: historial completo para administradores.
- `POST /api/admin/reservation-policies`: snapshot completo con vigencia inmediata. Exige `Idempotency-Key`; devuelve `201` para publicacion nueva, `200` para replay identico y `409` para replay divergente.

El repositorio publica con aislamiento serializable y bloqueos `UPDLOCK, HOLDLOCK`. Esta garantia fue cubierta localmente y por inspeccion estatica, pero no se sometio a carga concurrente contra SQL Server/Azure SQL real. La correccion excepcional permanece fuera de alcance y no tiene rutas implementadas.

## API de participantes grupales

Estado: ACCEPTED LOCALLY. API y frontend cubren detalle/progreso, codigo recuperable owner-only, rotacion, `/join` manual o por URL, confirmar, retirar y reconfirmar.

`POST /api/reservations` crea la solicitud grupal sin exponer el codigo. El propietario lo recupera bajo demanda con `GET /api/reservations/:id/join-code`. La base guarda hash para lookup y secreto cifrado. Las consultas por codigo exponen progreso agregado; confirmar, reconfirmar y retirar se serializan, exigen cuenta activa con RUT, respetan capacidad y auditan cada cambio.

La creacion acepta `targetParticipants` opcional. En solicitudes grupales, el valor por defecto es `minimumParticipants` y debe cumplir `minimo <= objetivo <= capacidad snapshot`; en recursos no grupales, enviarlo produce `400`. El umbral de confirmacion sigue siendo el minimo, mientras el objetivo limita nuevas altas. El propietario puede modificarlo mediante `PATCH /api/reservations/:id/target-participants` hasta `confirmationDeadline` inclusive; no puede bajarlo del minimo ni del conteo vigente, ni subirlo sobre la capacidad. Cada cambio se serializa y se registra en auditoria append-only. El progreso agrega `targetParticipants`, `confirmationDeadline`, `canEditTarget` e `isOwner`.

El deadline inclusivo en `America/Santiago` se aplica a objetivo, confirmacion y retiro. Bajo el minimo, una solicitud `PENDING` expira a `CANCELLED` de forma perezosa y mediante worker cada 30 segundos. `GET /api/reservations/:id/join-code` y `POST /api/reservations/:id/join-code/rotate` son owner-only y usan respuesta uniforme `404`; la rotacion permite migrar reservas legacy. Configuracion obligatoria: `JOIN_CODE_ENCRYPTION_KEYS` y `JOIN_CODE_KEY_VERSION`.

Estados HTTP relevantes: `400` payload/regla invalida, `403` operacion no permitida, `404` codigo o reserva no accesible, `409` capacidad/estado incompatible y `410` deadline vencido. Azure SQL real, idempotencia y concurrencia integrada siguen pendientes.

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
- Rechazar estado inicial enviado por cliente.
- Asignar estado inicial desde servidor.
- Rechazar duracion fuera del catalogo permitido.
- Rechazar inicio o termino fuera de jornada.
- Mantener hora y clasificacion temporal con `America/Santiago`.
- Permitir concurrencia en recursos `OPEN_USE`. Estas solicitudes no consumen
  ni quedan limitadas por la frecuencia configurada. El mismo usuario puede
  encadenarlas o combinarlas con reservas normales, pero no puede mantener dos
  reservas activas con horas realmente solapadas.
- Rechazar cruce con talleres activos del recurso.
- Rechazar cruce con bloqueo.
- Rechazar cruce con actividad programada.

### Casos criticos de cancelacion

- Usuario cancela reserva propia.
- Admin cancela reserva ajena.
- Usuario normal no cancela reserva ajena.
- No se cancela una reserva inexistente.
- No se cancela dos veces una reserva ya cancelada.
- No se cancela una reserva rechazada o expirada.
- No se cancela una reserva finalizada segun reloj y zona configurados.

### Casos de seguridad

- Usuario sin token recibe 401.
- Usuario normal no accede a rutas admin.
- Usuario bloqueado recibe 403.
- Modo dev sin cabeceras requeridas recibe 401.
- Usuario normal no puede actualizar imagenes de recursos.
- Respuestas 500 no contienen el texto del error interno simulado.

### Casos criticos de recursos

- Listar recursos con `imageUrl` cuando exista.
- Actualizar imagen de recurso como administrador.
- Rechazar ID de recurso invalido.
- Rechazar URL de imagen con formato no permitido.
- Permitir limpiar la imagen enviando valor vacio.

### Casos criticos de talleres

- Listar talleres activos para usuario autenticado.
- Rechazar listado sin autenticacion.
- Inscribir usuario con RUT en taller con cupos.
- Rechazar usuario normal sin RUT.
- Rechazar taller inexistente o inactivo.
- Rechazar taller sin cupos.
- Rechazar inscripcion duplicada.

## Prioridades sugeridas

1. Verificar online `RES-009`, `RES-010` y `RES-011` con una reserva de hora conocida, request manipulado y limites de jornada.
2. `QA-001`: ampliar pruebas reales de reglas de reservas y permisos mas alla de los casos ya cubiertos.
3. `SEC-005`: respuestas publicas sin detalles internos.
4. `API-004`: filtros de fecha/rango para disponibilidad.
