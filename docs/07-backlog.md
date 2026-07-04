# Poli-REDI - Backlog maestro

## Objetivo

Este backlog consolida las tareas reales detectadas durante la revision inicial, la migracion a Azure SQL Database y las primeras pruebas funcionales del frontend/backend.

La idea es usar este documento como base para crear issues en GitHub Projects o para delegar tareas puntuales a Codex.

## Estado base verificado

Fecha de referencia: 2026-07-03

Estado actual:

- Backend Go/Fiber ejecutando localmente.
- Frontend Vue/Vite ejecutando localmente.
- Autenticacion con Microsoft Entra ID activa.
- Azure SQL Database configurada como base de datos objetivo.
- Backend migrado desde PostgreSQL/pgx a SQL Server/go-mssqldb.
- Datos de recursos y reservas cargan desde Azure SQL.
- Ruta `/api/health` funciona como verificacion publica.
- Rutas protegidas requieren token Bearer.
- Vista `AvailabilityView` carga datos reales.

## Convenciones sugeridas

### Columnas del tablero

```txt
Backlog -> Ready for Codex -> In Progress -> Review -> Testing -> Done
```

### Labels sugeridos

```txt
frontend
backend
database
auth
reservas
disponibilidad
admin
reportes
documentacion
testing
bug
feature
refactor
codex-ready
needs-review
blocked
```

### Prioridades

- `P0`: bloquea funcionamiento base o provoca error critico.
- `P1`: necesario para MVP funcional.
- `P2`: mejora importante, no bloquea MVP.
- `P3`: mejora deseable o deuda tecnica menor.

---

# Hito 1 - Estabilizacion tecnica

## BACK-001 - Confirmar esquema Azure SQL desde cero

Prioridad: P0
Labels: `database`, `testing`, `needs-review`
Estado sugerido: Ready for Codex

### Contexto

Los scripts `database/schema.sql`, `database/seed.sql` y `database/drop.sql` fueron migrados a Azure SQL Database. Ya cargan datos en el entorno actual, pero falta validar el flujo completo en una base limpia.

### Objetivo

Verificar que una base nueva pueda crearse desde cero usando los scripts actuales.

### Criterios de aceptacion

- [ ] `database/drop.sql` puede ejecutarse sin fallar aunque no existan objetos.
- [ ] `database/schema.sql` crea todas las tablas, indices, triggers y vistas.
- [ ] `database/seed.sql` carga datos iniciales sin errores.
- [ ] El backend conecta correctamente despues de la carga.
- [ ] El frontend muestra recursos y reservas.

### Archivos relevantes

- `database/drop.sql`
- `database/schema.sql`
- `database/seed.sql`
- `backend/internal/database/database.go`

## BACK-002 - Limpiar logs de depuracion del frontend

Prioridad: P1
Labels: `frontend`, `refactor`
Estado sugerido: Ready for Codex

### Contexto

Existen `console.log` y `console.error` usados durante la prueba inicial, especialmente en autenticacion, disponibilidad y dashboard.

### Objetivo

Eliminar logs de depuracion visibles en consola o reemplazarlos por manejo de errores amigable.

### Criterios de aceptacion

- [ ] No quedan `console.log` de prueba en vistas y componentes.
- [ ] Los errores esperados se muestran en UI cuando corresponda.
- [ ] La consola del navegador no queda saturada durante uso normal.
- [ ] `npm run build` sigue pasando.

### Archivos relevantes

- `frontend/src/auth/msalConfig.js`
- `frontend/src/views/AvailabilityView.vue`
- `frontend/src/views/DashboardView.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/services/*.js`
- `frontend/src/stores/*.js`

## BACK-003 - Revisar documentacion historica PostgreSQL

Prioridad: P1
Labels: `documentacion`, `database`
Estado sugerido: Backlog

### Contexto

La documentacion inicial aun conserva menciones historicas a PostgreSQL y migracion pendiente. Algunas son utiles como trazabilidad, otras pueden confundir.

### Objetivo

Actualizar los documentos para diferenciar claramente estado historico, estado actual y decisiones definitivas.

### Criterios de aceptacion

- [ ] `docs/00-revision-inicial.md` indica que Azure SQL ya esta implementado.
- [ ] `docs/01-instalacion-y-ejecucion.md` prioriza Azure SQL y deja PostgreSQL solo como antecedente.
- [ ] `docs/03-base-de-datos.md` queda como fuente principal del modelo actual.
- [ ] No se exponen credenciales reales.

---

# Hito 2 - Autenticacion y usuario actual

## AUTH-001 - Usar usuario autenticado al crear reservas

Prioridad: P0
Labels: `frontend`, `backend`, `auth`, `reservas`, `bug`
Estado sugerido: Done

### Contexto

La vista de disponibilidad crea reservas con `userId: 2` fijo. Esto debe cambiar para usar el usuario autenticado obtenido desde `/api/me`.

### Objetivo

Crear reservas con el ID real del usuario autenticado.

### Criterios de aceptacion

- [x] El frontend obtiene el usuario actual desde el store de auth o `/api/me`.
- [x] `ReservationForm` o `AvailabilitySection` no usa `userId` fijo.
- [x] La reserva queda asociada al usuario autenticado en Azure SQL.
- [x] Si no hay usuario autenticado, se bloquea la accion y se muestra mensaje.
- [x] `npm run build` pasa.

### Resultado de implementacion

- `POST /api/reservations` toma el usuario desde el middleware de autenticacion y sobrescribe cualquier `userId` enviado por el cliente.
- `AvailabilitySection.vue` ya no envia `userId: 2`.
- El frontend carga el usuario autenticado antes de crear reservas.
- `PATCH /api/reservations/cancel` usa el usuario autenticado desde middleware; admin puede cancelar cualquier reserva y un usuario normal solo las propias.

### Archivos relevantes

- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/stores/auth.js`
- `frontend/src/stores/reservations.js`
- `backend/internal/handlers/me_handler.go`

## AUTH-002 - Sincronizar datos Entra ID en tabla users

Prioridad: P1
Labels: `backend`, `auth`, `database`
Estado sugerido: Backlog

### Contexto

El modelo contempla `entra_oid` y `tenant_id`, pero el repositorio de usuarios actualmente crea/actualiza principalmente por email.

### Objetivo

Guardar `entra_oid` y `tenant_id` del token en `users` para trazabilidad y consistencia.

### Criterios de aceptacion

- [ ] `GetOrCreateUserByEmail` recibe o actualiza `entra_oid` y `tenant_id`.
- [ ] Usuarios existentes se actualizan sin duplicarse.
- [ ] Se respeta email como identificador legible.
- [ ] Se valida indice unico `(tenant_id, entra_oid)` si ambos existen.
- [ ] `go test ./...` pasa.

### Archivos relevantes

- `backend/internal/middleware/auth_middleware.go`
- `backend/internal/repositories/users_repoository.go`
- `backend/internal/models/user.go`
- `database/schema.sql`

## AUTH-003 - Manejo visual de usuario bloqueado

Prioridad: P1
Labels: `frontend`, `backend`, `auth`
Estado sugerido: Backlog

### Contexto

El backend puede devolver usuario bloqueado, pero falta una experiencia clara en el frontend.

### Objetivo

Mostrar una pantalla o mensaje claro cuando un usuario bloqueado intenta usar el sistema.

### Criterios de aceptacion

- [ ] Si `/api/me` responde usuario bloqueado o 403, el frontend muestra mensaje claro.
- [ ] El usuario no queda atrapado en redirecciones.
- [ ] Existe opcion de cerrar sesion.
- [ ] No se muestran pantallas internas a usuarios bloqueados.

---

# Hito 3 - Disponibilidad y reservas

## RES-001 - Conectar actividad real al crear reserva

Prioridad: P1
Labels: `frontend`, `backend`, `reservas`
Estado sugerido: Done

### Contexto

El formulario permitia escribir una actividad/deporte, pero el payload usaba un valor fijo o no asociaba la actividad de forma real.

### Objetivo

Permitir seleccionar o resolver la actividad real al crear una reserva.

### Criterios de aceptacion

- [x] El backend expone listado de actividades activas o el frontend usa una fuente real.
- [x] El formulario permite seleccionar actividad.
- [x] No se usa `activityId` fijo.
- [x] La reserva creada queda asociada a la actividad correcta.
- [x] Se maneja estado sin actividades.

### Resultado de implementacion

- `GET /api/activities` entrega actividades activas desde Azure SQL.
- `ReservationForm.vue` muestra un selector de actividades reales.
- `AvailabilitySection.vue` carga actividades junto con recursos/reservas y envia `activityId` seleccionado.
- Si no existen actividades, el formulario muestra estado deshabilitado y la reserva puede quedar sin actividad asociada.

### Archivos relevantes

- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `backend/internal/routes/routes.go`
- `backend/internal/handlers`
- `backend/internal/repositories`

## RES-002 - Implementar endpoint GET /api/activities

Prioridad: P1
Labels: `backend`, `database`, `reservas`, `feature`
Estado sugerido: Done

### Contexto

El modelo Azure SQL incluye `activities`; el backend necesitaba exponer una ruta para consultarlas desde el formulario de reserva.

### Objetivo

Crear endpoint protegido para listar actividades activas.

### Criterios de aceptacion

- [x] Existe `GET /api/activities`.
- [x] Retorna actividades activas ordenadas por nombre.
- [x] Maneja errores de base de datos.
- [x] Usa estructura handler/repository coherente con el backend actual.
- [x] `go test ./...` pasa.

### Resultado de implementacion

- Se agregaron `activity.go`, `activities_repository.go` y `activities_handlers.go`.
- La ruta quedo protegida igual que recursos y reservas.
- Las reservas creadas con actividad devuelven el nombre de actividad como titulo.

## RES-003 - Mostrar reservas solo del dia seleccionado en disponibilidad

Prioridad: P1
Labels: `frontend`, `disponibilidad`, `reservas`
Estado sugerido: Done

### Contexto

La grilla recibe todas las reservas. Se debe confirmar si `ScheduleGrid` filtra correctamente por `selectedDate`; si no, la vista puede mostrar datos incorrectos.

### Objetivo

Asegurar que la disponibilidad muestre solo reservas, bloqueos y actividades del dia seleccionado.

### Criterios de aceptacion

- [x] Al cambiar fecha se actualiza la grilla correctamente.
- [x] Reservas de otros dias no aparecen en la fecha actual.
- [x] Se consideran hora de inicio y duracion.
- [x] Se documenta criterio de zona horaria.
- [x] `npm run build` pasa.

### Resultado de implementacion

- `ScheduleGrid.vue` filtra reservas por la fecha seleccionada.
- `ReservationBlock.vue`, `ResourceTimeline.vue` y `ReservationDetailModal.vue` usan un helper comun para fecha/hora.
- Las horas de reserva se interpretan como horario local de agenda para evitar desplazamientos por UTC al leer `DATETIME2` desde Azure SQL.
- La vista muestra el bloque reservado despues de crear una reserva y mantiene la validacion de solapamiento desde la base de datos.

### Archivos relevantes

- `frontend/src/components/availability/ScheduleGrid.vue`
- `frontend/src/components/availability/ReservationBlock.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/utils/dateUtils.js`

## RES-004 - Integrar bloqueos y actividades programadas en calendario

Prioridad: P1
Labels: `frontend`, `backend`, `disponibilidad`, `database`
Estado sugerido: Backlog

### Contexto

La base incluye `availability_blocks`, `scheduled_activities` y la vista `vw_resource_calendar`, pero el backend/frontend aun se enfocan principalmente en reservas.

### Objetivo

Mostrar en disponibilidad reservas, bloqueos y actividades institucionales.

### Criterios de aceptacion

- [ ] Backend expone datos de calendario por rango de fecha.
- [ ] Frontend distingue visualmente reserva, bloqueo y actividad programada.
- [ ] No se permite seleccionar horarios bloqueados.
- [ ] Se muestra detalle al hacer clic en cada bloque.
- [ ] La informacion proviene de Azure SQL.

## RES-005 - Mejorar validaciones del formulario de reserva

Prioridad: P1
Labels: `frontend`, `reservas`, `ux`
Estado sugerido: Ready for Codex

### Contexto

El formulario valida campos obligatorios de forma basica, pero no muestra mensajes por campo ni valida rangos de duracion con claridad.

### Objetivo

Agregar validaciones visibles antes de enviar reserva.

### Criterios de aceptacion

- [ ] Muestra error si falta recurso.
- [ ] Muestra error si falta fecha.
- [ ] Muestra error si falta hora.
- [ ] Valida duracion mayor a 0.
- [ ] Valida participantes mayor a 0.
- [ ] Deshabilita boton mientras se crea la reserva.
- [ ] Muestra error devuelto por backend sin cerrar modal.

## RES-006 - Implementar vista Mis Reservas

Prioridad: P1
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Ready for Codex

### Contexto

`ReservationsView.vue` todavia muestra `Proximamente...`.

### Objetivo

Crear una vista donde el usuario vea sus reservas.

### Criterios de aceptacion

- [ ] Muestra reservas del usuario autenticado.
- [ ] Muestra recurso, actividad, fecha, hora, duracion y estado.
- [ ] Tiene estados de carga, error y vacio.
- [ ] Permite ir al detalle de una reserva.
- [ ] Usa datos reales del backend.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/views/ReservationsView.vue`
- `frontend/src/stores/reservations.js`
- `frontend/src/services/reservations.service.js`

## RES-007 - Implementar cancelacion de reservas desde frontend

Prioridad: P1
Labels: `frontend`, `backend`, `reservas`
Estado sugerido: Done

### Contexto

Existe `PATCH /api/reservations/cancel`; ya usa usuario autenticado en backend, pero falta conectar un flujo visual completo desde la interfaz.

### Objetivo

Permitir cancelar reservas desde la interfaz usando el usuario autenticado.

### Criterios de aceptacion

- [x] No se usa `requestedByUserId` fijo.
- [x] El backend determina usuario desde token o valida contra usuario autenticado.
- [x] La UI pide confirmacion antes de cancelar.
- [x] La reserva cambia a estado `CANCELLED`.
- [x] Se muestra mensaje de exito o error.
- [x] La lista se actualiza sin recargar toda la app.

### Resultado de implementacion

- Los bloques de reserva en disponibilidad son seleccionables.
- Al seleccionar una reserva se abre `ReservationDetailModal`.
- Admin puede cancelar cualquier reserva; usuario normal solo ve accion de cancelacion para reservas propias.
- El modal muestra errores del backend si la cancelacion falla.
- Al cancelar, la grilla se refresca y el bloque desaparece de la disponibilidad activa.

---

# Hito 4 - Pantallas pendientes

## UI-001 - Implementar vista Recursos

Prioridad: P1
Labels: `frontend`, `recursos`, `feature`
Estado sugerido: Ready for Codex

### Contexto

`ResourcesView.vue` ya lista recursos reales desde `/api/resources`; falta agregar filtros.

### Objetivo

Mostrar catalogo de recursos deportivos con datos reales.

### Criterios de aceptacion

- [x] Lista recursos desde `/api/resources`.
- [x] Muestra nombre, tipo, modo de reserva, capacidad y estado.
- [ ] Permite filtrar por tipo o sede si los datos estan disponibles.
- [x] Tiene estados de carga, error y vacio.
- [x] Mantiene estilo visual actual.

### Resultado parcial

- `ResourcesView.vue` ya no muestra `Proximamente...`.
- `GET /api/resources` ahora incluye `capacity` desde Azure SQL.
- Queda pendiente agregar filtros.

## UI-002 - Implementar detalle de reserva

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Backlog

### Contexto

`ReservationDetailView.vue` muestra `Proximamente...`.

### Objetivo

Mostrar informacion detallada de una reserva.

### Criterios de aceptacion

- [ ] Muestra recurso, actividad, usuario, fecha, hora, duracion y estado.
- [ ] Muestra participantes si existen.
- [ ] Permite volver a Mis Reservas o Disponibilidad.
- [ ] Maneja reserva no encontrada.

## UI-003 - Implementar Historial

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Backlog

### Contexto

`HistoryView.vue` muestra `Proximamente...`.

### Objetivo

Mostrar historial de reservas pasadas, canceladas o rechazadas.

### Criterios de aceptacion

- [ ] Lista reservas historicas del usuario.
- [ ] Permite filtrar por estado.
- [ ] Permite filtrar por fecha.
- [ ] Tiene estados de carga, error y vacio.

## UI-004 - Conectar Dashboard a datos reales

Prioridad: P1
Labels: `frontend`, `dashboard`, `feature`
Estado sugerido: Done

### Contexto

`DashboardView.vue` usa arreglos locales de instalaciones y reservas.

### Objetivo

Reemplazar datos mock del dashboard por datos reales.

### Criterios de aceptacion

- [x] Instalaciones provienen de `/api/resources`.
- [x] Proximas reservas provienen del backend.
- [x] Se eliminan logs de seleccion.
- [x] Se muestran estados de carga/error.
- [x] El dashboard no depende de imagenes externas si no son necesarias.

### Resultado de implementacion

- `DashboardView.vue` ya no tiene arreglos locales de instalaciones ni reservas.
- `FacilityCarousel.vue` consume recursos reales.
- `ReservationsPanel.vue` consume reservas recibidas por props.
- `FacilityCard.vue` usa una imagen generada por CSS cuando el recurso no tiene imagen en base de datos.

## UI-005 - Implementar Configuracion

Prioridad: P3
Labels: `frontend`, `settings`, `feature`
Estado sugerido: Done

### Contexto

`SettingsView.vue` ya muestra una configuracion basica de cuenta con datos reales del usuario autenticado.

### Objetivo

Crear una vista inicial de configuracion de cuenta/sistema.

### Criterios de aceptacion

- [x] Muestra datos del usuario autenticado.
- [x] Permite cerrar sesion.
- [x] Indica rol actual.
- [x] No expone datos sensibles.

### Resultado de implementacion

- `SettingsView.vue` muestra nombre, correo, rol y estado desde el store de autenticacion.
- El menu de usuario navega a `/settings`.
- `/settings` queda disponible para cualquier usuario autenticado.

---

# Hito 5 - Administracion

## ADMIN-001 - Implementar panel administrador base

Prioridad: P1
Labels: `frontend`, `backend`, `admin`, `feature`
Estado sugerido: Backlog

### Contexto

`AdminView.vue` muestra `Proximamente...`. El modelo contempla usuarios admin, recursos, bloqueos, actividades e infracciones.

### Objetivo

Crear un panel administrador inicial con accesos a gestion de recursos, usuarios, bloqueos y reportes.

### Criterios de aceptacion

- [ ] Solo usuarios admin pueden acceder.
- [ ] Muestra tarjetas/resumen de recursos, reservas, usuarios e infracciones.
- [ ] Tiene enlaces a secciones administrativas.
- [ ] Maneja usuario sin permisos.

## ADMIN-002 - Implementar gestion de usuarios

Prioridad: P2
Labels: `frontend`, `backend`, `admin`, `auth`
Estado sugerido: Backlog

### Objetivo

Permitir que un admin vea usuarios y pueda bloquear/desbloquear cuentas.

### Criterios de aceptacion

- [ ] Existe endpoint para listar usuarios.
- [ ] Existe endpoint para bloquear/desbloquear usuario.
- [ ] La UI muestra email, nombre, rol y estado.
- [ ] No permite bloquearse a si mismo accidentalmente.
- [ ] Registra auditoria.

## ADMIN-003 - Implementar gestion de recursos

Prioridad: P2
Labels: `frontend`, `backend`, `admin`, `recursos`
Estado sugerido: Backlog

### Objetivo

Permitir crear, editar, activar o desactivar recursos deportivos.

### Criterios de aceptacion

- [ ] CRUD basico de recursos.
- [ ] Recurso pertenece a una sede.
- [ ] Valida `reservation_mode`.
- [ ] No permite eliminar recursos con reservas historicas sin criterio definido.
- [ ] Refresca disponibilidad luego de cambios.

## ADMIN-004 - Implementar bloqueos de disponibilidad

Prioridad: P1
Labels: `frontend`, `backend`, `admin`, `disponibilidad`
Estado sugerido: Backlog

### Objetivo

Permitir que admin cree bloqueos por mantencion, cierre, evento u otro motivo.

### Criterios de aceptacion

- [ ] Admin selecciona recurso, fecha, hora inicio y hora termino.
- [ ] Backend valida solapamientos.
- [ ] Bloqueos aparecen en calendario/disponibilidad.
- [ ] Se puede desactivar un bloqueo.
- [ ] Se muestra error si cruza reserva confirmada.

## ADMIN-005 - Implementar programacion institucional

Prioridad: P2
Labels: `frontend`, `backend`, `admin`, `disponibilidad`
Estado sugerido: Backlog

### Objetivo

Permitir registrar clases, talleres, eventos, campeonatos o entrenamientos institucionales.

### Criterios de aceptacion

- [ ] Admin crea actividad programada sobre un recurso.
- [ ] Backend valida solapamientos.
- [ ] Actividades aparecen en calendario.
- [ ] Soporta descripcion y tipo.
- [ ] Define comportamiento futuro de recurrencia.

---

# Hito 6 - Reportes, infracciones y notificaciones

## REP-001 - Implementar vista Reportes

Prioridad: P2
Labels: `frontend`, `backend`, `reportes`
Estado sugerido: Backlog

### Contexto

`ReportsView.vue` muestra `Proximamente...`. La base incluye vistas de uso de recursos, horas punta e infracciones.

### Objetivo

Mostrar reportes iniciales desde vistas SQL.

### Criterios de aceptacion

- [ ] Reporte uso de recursos.
- [ ] Reporte horas punta.
- [ ] Reporte infracciones por usuario.
- [ ] Estados de carga/error.
- [ ] Acceso restringido a admin.

## REP-002 - Implementar infracciones

Prioridad: P2
Labels: `backend`, `frontend`, `admin`, `reportes`
Estado sugerido: Backlog

### Objetivo

Permitir registrar infracciones de usuarios, asociadas opcionalmente a una reserva.

### Criterios de aceptacion

- [ ] Endpoint para crear infraccion.
- [ ] Endpoint para listar infracciones por usuario.
- [ ] UI admin para registrar infraccion.
- [ ] Trigger crea notificacion automaticamente.
- [ ] Se registra auditoria.

## NOTIF-001 - Conectar campana de notificaciones

Prioridad: P2
Labels: `frontend`, `backend`, `notificaciones`
Estado sugerido: Backlog

### Objetivo

Mostrar notificaciones reales al usuario desde Azure SQL.

### Criterios de aceptacion

- [x] Endpoint para listar notificaciones del usuario.
- [ ] Endpoint para marcar como leida.
- [x] Campana muestra contador real.
- [ ] UI diferencia leidas/no leidas.
- [x] Maneja estado vacio.

### Resultado parcial

- `GET /api/notifications` lista notificaciones del usuario autenticado desde Azure SQL.
- `NotificationBell.vue` ya no usa notificaciones locales.
- Queda pendiente marcar como leida y diferenciar visualmente leidas/no leidas.

---

# Hito 7 - Backend/API

## API-001 - Filtrar recursos por sede, tipo y estado

Prioridad: P2
Labels: `backend`, `recursos`, `feature`
Estado sugerido: Backlog

### Objetivo

Mejorar `GET /api/resources` con filtros opcionales.

### Criterios de aceptacion

- [ ] Filtro por `venueId`.
- [ ] Filtro por `type`.
- [ ] Filtro por `isActive`.
- [ ] Mantiene compatibilidad con llamada actual sin filtros.

## API-002 - Filtrar reservas por usuario, fecha y estado

Prioridad: P1
Labels: `backend`, `reservas`, `feature`
Estado sugerido: Ready for Codex

### Objetivo

Evitar que el frontend reciba todas las reservas cuando solo necesita un subconjunto.

### Criterios de aceptacion

- [ ] `GET /api/reservations` acepta filtro por fecha/rango.
- [ ] Permite filtrar por usuario autenticado.
- [ ] Permite filtrar por estado.
- [ ] Disponibilidad usa rango de fecha.
- [ ] Mis Reservas usa usuario autenticado.

## API-003 - Usar usuario autenticado en operaciones protegidas

Prioridad: P0
Labels: `backend`, `auth`, `reservas`, `security`
Estado sugerido: Done

### Contexto

Algunas operaciones todavia reciben IDs de usuario desde el frontend, por ejemplo cancelacion.

### Objetivo

Evitar confiar en IDs enviados por cliente cuando el usuario ya viene en el token.

### Criterios de aceptacion

- [x] Crear reserva usa usuario autenticado si corresponde.
- [x] Cancelar reserva usa usuario autenticado/admin desde middleware.
- [x] El frontend no envia `requestedByUserId` fijo.
- [x] Usuarios no pueden operar reservas ajenas salvo admin.

### Resultado de implementacion

- `CreateReservation` toma el usuario desde `middleware.GetLocalUser`.
- `CancelReservation` toma el usuario desde `middleware.GetLocalUser`.
- Admin puede cancelar cualquier reserva.
- Usuario normal solo puede cancelar reservas donde `reservations.user_id` coincide con su usuario local.
- El frontend envia solo `reservationId` al endpoint de cancelacion.

---

# Hito 8 - Calidad, pruebas y seguridad

## QA-001 - Agregar pruebas backend para reglas de reservas

Prioridad: P1
Labels: `backend`, `testing`, `reservas`
Estado sugerido: Backlog

### Objetivo

Crear pruebas de la capa servicio/repositorio para reglas criticas.

### Casos minimos

- [ ] Crear reserva valida.
- [ ] Rechazar recurso inexistente.
- [ ] Rechazar usuario inexistente.
- [ ] Rechazar conflicto horario.
- [ ] Rechazar usuario bloqueado.
- [ ] Cancelar reserva como admin.
- [ ] Rechazar cancelacion sin permisos.

## QA-002 - Agregar pruebas frontend basicas

Prioridad: P2
Labels: `frontend`, `testing`
Estado sugerido: Backlog

### Objetivo

Agregar pruebas o checklist automatizado para pantallas criticas.

### Casos minimos

- [ ] Render de disponibilidad.
- [ ] Estados de carga/error.
- [ ] Formulario de reserva valida campos.
- [ ] Router no entra en bucle.

## SEC-001 - Revisar exposicion de secretos

Prioridad: P0
Labels: `security`, `documentacion`
Estado sugerido: Done

### Objetivo

Asegurar que credenciales reales no queden versionadas ni documentadas.

### Criterios de aceptacion

- [x] `.env` no aparece en git.
- [x] `.env.example` no contiene password.
- [x] Documentacion no contiene password real.
- [x] Se recomienda rotar claves si fueron expuestas en chat o capturas.

### Resultado de revision

- No se encontraron coincidencias de la clave real en archivos versionados.
- No se encontraron variables secretas activas con valor en archivos versionados.
- `.gitignore` protege archivos `.env` de raiz, backend y frontend.
- Los ejemplos de conexion quedaron comentados o con placeholders.
- Recomendacion: rotar la clave de Azure SQL porque fue compartida durante la configuracion local.

## SEC-002 - Revisar CORS para ambiente productivo

Prioridad: P2
Labels: `backend`, `security`
Estado sugerido: Backlog

### Contexto

El backend usa `AllowOrigins: "*"`.

### Objetivo

Configurar origenes permitidos por ambiente.

### Criterios de aceptacion

- [ ] Desarrollo permite localhost.
- [ ] Produccion limita origenes autorizados.
- [ ] Configuracion viene por variable de entorno.

---

# Hito 9 - Documentacion y entrega FIP/tesis

## DOC-001 - Actualizar README principal

Prioridad: P1
Labels: `documentacion`
Estado sugerido: Ready for Codex

### Objetivo

Actualizar `README.md` con stack real, instalacion, ejecucion, Azure SQL y autenticacion.

### Criterios de aceptacion

- [ ] Describe frontend, backend y base de datos actual.
- [ ] Explica variables de entorno.
- [ ] Explica como ejecutar backend y frontend.
- [ ] Explica como validar `/api/health`.
- [ ] No contiene secretos.

## DOC-002 - Documentar arquitectura

Prioridad: P1
Labels: `documentacion`, `arquitectura`
Estado sugerido: Backlog

### Objetivo

Completar `docs/02-arquitectura.md` con diagrama y descripcion de flujo.

### Criterios de aceptacion

- [ ] Describe frontend Vue.
- [ ] Describe backend Go/Fiber.
- [ ] Describe autenticacion Entra ID.
- [ ] Describe Azure SQL Database.
- [ ] Describe flujo de reserva.
- [ ] Incluye diagrama Mermaid.

## DOC-003 - Documentar flujo de reservas

Prioridad: P1
Labels: `documentacion`, `reservas`
Estado sugerido: Backlog

### Objetivo

Completar `docs/06-flujo-reservas.md`.

### Criterios de aceptacion

- [ ] Describe flujo usuario normal.
- [ ] Describe flujo admin.
- [ ] Describe validaciones de conflicto.
- [ ] Describe estados de reserva.
- [ ] Incluye diagrama de secuencia o actividad.

---

# Hito 10 - Despliegue

## DEPLOY-001 - Definir estrategia de despliegue

Prioridad: P2
Labels: `deploy`, `documentacion`
Estado sugerido: Backlog

### Objetivo

Definir donde se desplegaran frontend y backend.

### Opciones a evaluar

- Azure App Service para backend.
- Azure Static Web Apps para frontend.
- Variables de entorno en Azure.
- Dominio o URL institucional.

### Criterios de aceptacion

- [ ] Documento corto con alternativa elegida.
- [ ] Variables necesarias definidas.
- [ ] Pasos de despliegue inicial documentados.

## DEPLOY-002 - Preparar backend para produccion

Prioridad: P2
Labels: `backend`, `deploy`, `security`
Estado sugerido: Backlog

### Criterios de aceptacion

- [ ] Puerto configurable.
- [ ] CORS configurable.
- [ ] Logs no exponen secretos.
- [ ] Variables de Azure SQL definidas en entorno.
- [ ] Health check disponible.

---

# Inventario actual de datos duros detectados

Revision realizada durante la conexion de actividades reales.

## Datos que ya pueden reemplazarse por Azure SQL

- `frontend/src/components/layout/NotificationBell.vue`: falta marcar notificaciones como leidas; relacionado con `NOTIF-001`.
- Vistas `ReservationsView.vue`, `ReservationDetailView.vue`, `HistoryView.vue`, `ReportsView.vue`, `AdminView.vue` y `UsersView.vue`: pantallas `Proximamente...` que deben conectarse a endpoints reales o nuevas tareas.

## Datos duros resueltos

- `frontend/src/components/layout/HeaderBar.vue`: el saludo usa el nombre del usuario autenticado.
- `frontend/src/components/layout/UserMenu.vue`: nombre, correo, rol y avatar iniciales vienen del usuario autenticado o de la cuenta Microsoft.
- `frontend/src/views/SettingsView.vue`: muestra datos reales de la cuenta autenticada.
- `frontend/src/views/DashboardView.vue`: instalaciones y proximas reservas ahora vienen de stores/API.
- `frontend/src/components/dashboard/ReservationsPanel.vue`: ya no contiene reservas locales.
- `frontend/src/components/layout/NotificationBell.vue`: ya no contiene notificaciones locales.
- `frontend/src/components/forms/ReservationForm.vue`: se quito `participants: 10` porque aun no se persiste en base de datos.
- `frontend/src/views/ResourcesView.vue`: ya carga recursos reales desde Azure SQL.

## Datos estaticos que pueden mantenerse por ahora

- `frontend/src/components/layout/Sidebar.vue`: menu de navegacion.
- `frontend/src/components/dashboard/QuickActions.vue`: accesos rapidos de navegacion.
- `frontend/src/components/availability/CalendarMini.vue`: nombres de meses y dias.
- `frontend/src/components/forms/DateTimePicker.vue`: opciones de duracion; podrian pasar a configuracion en una tarea futura, pero no bloquean el MVP.

---

# Orden recomendado de ejecucion

## Sprint 1 - Estabilizar MVP tecnico

1. `SEC-001` Revisar exposicion de secretos.
2. `AUTH-001` Usar usuario autenticado al crear reservas.
3. `API-003` Usar usuario autenticado en operaciones protegidas.
4. `RES-002` Implementar `GET /api/activities`.
5. `RES-001` Conectar actividad real al crear reserva.
6. `BACK-002` Limpiar logs de depuracion del frontend.

## Sprint 2 - Completar flujo usuario

1. `RES-003` Mostrar reservas solo del dia seleccionado.
2. `RES-005` Mejorar validaciones del formulario.
3. `RES-006` Implementar Mis Reservas.
4. `RES-007` Implementar cancelacion desde frontend.
5. `UI-001` Implementar vista Recursos.
6. `UI-004` Conectar Dashboard a datos reales.

## Sprint 3 - Administracion y calendario completo

1. `RES-004` Integrar bloqueos y actividades programadas.
2. `ADMIN-001` Panel administrador base.
3. `ADMIN-004` Bloqueos de disponibilidad.
4. `ADMIN-003` Gestion de recursos.
5. `ADMIN-002` Gestion de usuarios.
6. `ADMIN-005` Programacion institucional.

## Sprint 4 - Reportes, calidad y documentacion

1. `REP-001` Reportes.
2. `REP-002` Infracciones.
3. `NOTIF-001` Notificaciones.
4. `QA-001` Pruebas backend.
5. `DOC-001` README.
6. `DOC-002` Arquitectura.
7. `DOC-003` Flujo de reservas.

# Tareas Codex-ready iniciales

Estas son las mejores primeras tareas para delegar a Codex porque tienen bajo riesgo y alto impacto:

1. `SEC-001` Revisar exposicion de secretos.
2. `BACK-002` Limpiar logs de depuracion del frontend.
3. `AUTH-001` Usar usuario autenticado al crear reservas.
4. `RES-002` Implementar `GET /api/activities`.
5. `RES-001` Conectar actividad real al crear reserva.
6. `RES-006` Implementar vista Mis Reservas.

# Notas importantes

- No volver a usar scripts PostgreSQL en Azure SQL.
- No guardar passwords en documentos ni en archivos versionados.
- Cualquier tarea que toque autenticacion debe probarse con usuario real de Entra ID.
- Cualquier tarea que toque reservas debe probar conflicto horario.
- Antes de cerrar una tarea, ejecutar como minimo:

```bash
cd backend
go test ./...

cd ../frontend
npm run build
```
