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
- Demo online desplegada en Azure Static Web Apps y Azure App Service.
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
Estado sugerido: Done

### Contexto

Los scripts `database/schema.sql`, `database/seed.sql` y `database/drop.sql` fueron migrados a Azure SQL Database. Ya cargan datos en el entorno actual, pero falta validar el flujo completo en una base limpia.

### Objetivo

Verificar que una base nueva pueda crearse desde cero usando los scripts actuales.

### Criterios de aceptacion

- [x] `database/drop.sql` puede ejecutarse sin fallar aunque no existan objetos.
- [x] `database/schema.sql` crea todas las tablas, indices, triggers y vistas.
- [x] `database/seed.sql` carga datos iniciales sin errores.
- [x] El backend conecta correctamente despues de la carga.
- [x] El frontend muestra recursos y reservas.

### Resultado de implementacion

- Se valido el flujo desde una base limpia usando `drop.sql`, `schema.sql` y `seed.sql`.
- La base queda operativa y el frontend/backend cargan datos reales.
- Se confirmo que el sistema puede poblar usuarios autenticados por Entra ID que existen en Azure pero aun no existen en Poli-REDI.

### Archivos relevantes

- `database/drop.sql`
- `database/schema.sql`
- `database/seed.sql`
- `backend/internal/database/database.go`

## BACK-002 - Limpiar logs de depuracion del frontend

Prioridad: P1
Labels: `frontend`, `refactor`
Estado sugerido: Done

### Contexto

Existen `console.log` y `console.error` usados durante la prueba inicial, especialmente en autenticacion, disponibilidad y dashboard.

### Objetivo

Eliminar logs de depuracion visibles en consola o reemplazarlos por manejo de errores amigable.

### Criterios de aceptacion

- [x] No quedan `console.log` de prueba en vistas y componentes.
- [x] Los errores esperados se muestran en UI cuando corresponda.
- [x] La consola del navegador no queda saturada durante uso normal.
- [x] `npm run build` sigue pasando.

### Resultado de implementacion

- Se eliminaron `console.error` y `console.warn` de autenticacion, callback y cliente API.
- `api.js` deja pasar la solicitud sin token cuando no puede obtenerlo; las rutas protegidas mantienen el error visible desde stores/vistas.
- `AuthCallbackView.vue` redirige a una pantalla segura si el callback falla, sin saturar la consola.
- Se limpio el servicio legado `frontend/src/services/authService.js` y se corrigio su import de `msalConfig`.
- `npm run build` pasa.

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
Estado sugerido: Done

### Contexto

La documentacion inicial aun conserva menciones historicas a PostgreSQL y migracion pendiente. Algunas son utiles como trazabilidad, otras pueden confundir.

### Objetivo

Actualizar los documentos para diferenciar claramente estado historico, estado actual y decisiones definitivas.

### Criterios de aceptacion

- [x] `docs/00-revision-inicial.md` indica que Azure SQL ya esta implementado.
- [x] `docs/01-instalacion-y-ejecucion.md` prioriza Azure SQL y deja PostgreSQL solo como antecedente.
- [x] `docs/03-base-de-datos.md` queda como fuente principal del modelo actual.
- [x] No se exponen credenciales reales.

### Resultado de implementacion

- `docs/00-revision-inicial.md` quedo marcado como documento historico.
- `docs/01-instalacion-y-ejecucion.md` fue reemplazado por una guia vigente para Azure SQL, backend Go/Fiber y frontend Vue/Vite.
- `docs/02-arquitectura.md` documenta arquitectura, autenticacion, flujo de reserva y despliegue.
- `docs/03-base-de-datos.md` queda como referencia del modelo Azure SQL actual.

## BACK-004 - Pulir MVP 1 antes de cierre definitivo

Prioridad: P1
Labels: `documentacion`, `testing`, `security`, `deploy`, `codex-ready`
Estado sugerido: Done

### Contexto

El MVP 1 ya funciona y esta desplegado como demo online, pero se reabre para una ronda final de pulido tecnico antes de considerarlo cerrado definitivamente.

La idea no es agregar nuevas funcionalidades grandes, sino fortalecer la base: ejecucion local, despliegue, configuracion, seguridad minima, documentacion y pruebas de humo.

### Objetivo

Dejar el MVP 1 lo mas estable, demostrable y defendible posible.

### Criterios de aceptacion

- [x] README y documentos de instalacion reflejan el estado real actual.
- [x] La guia local permite levantar backend y frontend sin ambiguedades.
- [ ] La guia de despliegue indica variables obligatorias y valores seguros.
- [ ] Existe checklist de validacion MVP 1 antes de demo.
- [x] Se ejecuta o documenta resultado de `go test ./...`.
- [x] Se ejecuta o documenta resultado de `npm run build`.
- [x] Se revisa que no haya referencias activas a PostgreSQL como tecnologia vigente.
- [ ] Se revisa que no haya secretos, passwords ni tokens en documentacion versionada.

### Archivos relevantes

- `README.md`
- `docs/01-instalacion-y-ejecucion.md`
- `docs/02-arquitectura.md`
- `docs/03-base-de-datos.md`
- `docs/09-mvps-roadmap.md`
- `backend/.env.example`
- `frontend/.env.example`

## BACK-005 - Checklist de demo tecnica MVP 1

Prioridad: P1
Labels: `testing`, `documentacion`, `deploy`
Estado sugerido: Ready for Codex

### Contexto

La demo online existe, pero conviene tener una lista repetible para validar que la base tecnica sigue operativa antes de una presentacion.

### Objetivo

Crear un checklist corto de pruebas de humo para MVP 1.

### Criterios de aceptacion

- [ ] Validar `/api/health` local.
- [ ] Validar `/api/health` online.
- [ ] Validar login institucional o modo local segun ambiente.
- [ ] Validar carga de recursos reales.
- [ ] Validar carga de actividades reales.
- [ ] Validar creacion de una reserva de prueba.
- [ ] Validar cancelacion de una reserva de prueba.
- [ ] Validar que un usuario normal no acceda a rutas admin.
- [ ] Registrar fecha, ambiente y resultado de la ultima validacion.

### Evidencia reciente

- 2026-07-06: `go test ./...` ejecutado correctamente en backend.
- 2026-07-06: `npm run build` ejecutado correctamente en frontend.

## BACK-006 - Normalizar mensajes y codificacion visible

Prioridad: P2
Labels: `backend`, `frontend`, `ux`, `refactor`, `codex-ready`
Estado sugerido: Ready for Codex

### Contexto

Durante revisiones se observaron textos con problemas de codificacion, por ejemplo acentos mostrados como caracteres rotos en algunos archivos.

### Objetivo

Revisar mensajes visibles de backend/frontend para que no aparezcan caracteres corruptos ni textos tecnicos innecesarios durante la demo.

### Criterios de aceptacion

- [ ] No hay mensajes visibles con caracteres corruptos.
- [ ] Errores frecuentes de autenticacion y reserva usan texto claro.
- [ ] Los mensajes tecnicos quedan reservados para logs o `detail` cuando corresponda.
- [ ] `npm run build` pasa.
- [ ] `go test ./...` pasa.

## BACK-007 - Crear base global de sistema visual MVP 1

Prioridad: P1
Labels: `frontend`, `ux`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

El proyecto tiene un estilo global minimo en `frontend/src/style.css`, pero `frontend/src/assets/styles/main.css` y `frontend/src/assets/styles/variables.css` estan vacios. La mayor parte de tarjetas, botones, estados, radios, colores y sombras se redefine dentro de cada componente con `<style scoped>`.

Esto provoca una sensacion de falta de unificacion visual, especialmente en pantallas del MVP 1 usadas para demo: login, layout, disponibilidad, formularios, estados y reservas.

### Objetivo

Crear una base global de estilos reutilizables sin redisenar la aplicacion completa ni romper los estilos existentes.

### Alcance sugerido para Codex

1. Definir tokens en `frontend/src/assets/styles/variables.css`.
2. Importar `variables.css` y `assets/styles/main.css` desde `frontend/src/style.css`.
3. Crear utilidades globales base en `assets/styles/main.css`.
4. Mantener los `<style scoped>` para layout especifico de cada componente.
5. No cambiar comportamiento funcional.

### Tokens minimos esperados

- Colores: fondo, superficie, texto principal, texto secundario, borde, primario, exito, advertencia, error.
- Radios: `sm`, `md`, `lg`, `xl`, `pill`.
- Sombras: tarjeta, modal, foco.
- Espaciado: escala simple para 4, 8, 12, 16, 20, 24, 32.
- Tipografia: tamanos base para titulo de pagina, titulo de seccion, cuerpo y ayuda.

### Clases globales minimas esperadas

- `.app-card`
- `.app-section-header`
- `.state-card`, `.state-card.error`, `.state-card.success`, `.state-card.warning`
- `.app-button`, `.app-button.primary`, `.app-button.secondary`, `.app-button.danger`
- `.form-field`
- `.form-error`
- `.app-badge`

### Criterios de aceptacion

- [x] `variables.css` contiene tokens reutilizables con nombres claros.
- [x] `style.css` importa la base global.
- [x] `main.css` contiene clases globales para tarjetas, botones, estados, campos y badges.
- [x] Los tokens reemplazan valores repetidos evidentes sin hacer un rediseño masivo.
- [x] No se eliminan estilos scoped necesarios para layout especifico.
- [x] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/style.css`
- `frontend/src/assets/styles/variables.css`
- `frontend/src/assets/styles/main.css`

## BACK-008 - Aplicar sistema visual global a pantallas MVP 1

Prioridad: P1
Labels: `frontend`, `ux`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

Una vez creada la base global de estilos, conviene aplicarla gradualmente solo a las pantallas y componentes visibles en la demo del MVP 1. El objetivo es mejorar consistencia sin abrir una refactorizacion visual enorme.

### Objetivo

Unificar el aspecto de tarjetas, botones, modales, estados y campos en el recorrido base del MVP 1.

### Alcance sugerido para Codex

Aplicar primero en:

- `LoginView.vue`
- `BlockedView.vue`
- `AppLayout.vue`
- `HeaderBar.vue`
- `Sidebar.vue`
- `AvailabilitySection.vue`
- `ReservationForm.vue`
- `ReservationDetailModal.vue`
- `ReservationsView.vue`
- `ResourcesView.vue`

### Reglas de implementacion

- Reemplazar duplicacion evidente por clases globales.
- Mantener estilos scoped para grillas, timeline, layout responsive y casos especiales.
- No cambiar copy funcional salvo textos tecnicos visibles.
- No introducir nuevas dependencias de UI.
- No rehacer la marca ni cambiar completamente la paleta.

### Criterios de aceptacion

- [x] Botones primarios, secundarios y de peligro se ven consistentes en el flujo base.
- [x] Tarjetas y estados de error/exito/advertencia comparten radios, bordes y colores.
- [x] Campos de formulario comparten altura, borde, foco y error.
- [x] Modales principales comparten overlay, superficie, sombra y botones.
- [x] La vista de disponibilidad conserva su layout y no pierde legibilidad.
- [x] La UI sigue siendo responsive en mobile.
- [x] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/assets/styles/main.css`
- `frontend/src/assets/styles/variables.css`
- `frontend/src/views/LoginView.vue`
- `frontend/src/views/BlockedView.vue`
- `frontend/src/components/layout/AppLayout.vue`
- `frontend/src/components/layout/HeaderBar.vue`
- `frontend/src/components/layout/Sidebar.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/availability/ReservationDetailModal.vue`
- `frontend/src/views/ReservationsView.vue`
- `frontend/src/views/ResourcesView.vue`

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
Estado sugerido: Done

### Contexto

El modelo contempla `entra_oid` y `tenant_id`, pero el repositorio de usuarios actualmente crea/actualiza principalmente por email.

### Objetivo

Guardar `entra_oid` y `tenant_id` del token en `users` para trazabilidad y consistencia.

### Criterios de aceptacion

- [x] `GetOrCreateUserByEmail` recibe o actualiza `entra_oid` y `tenant_id`.
- [x] Usuarios existentes se actualizan sin duplicarse.
- [x] Se respeta email como identificador legible.
- [x] Se valida indice unico `(tenant_id, entra_oid)` si ambos existen.
- [x] `go test ./...` pasa.

### Resultado de implementacion

- `GetOrCreateUserByEmail` ahora retorna tambien `entra_oid` y `tenant_id`.
- Se agrego `UpdateUserEntraIdentity` para sincronizar identidad Microsoft sin duplicar usuarios.
- El middleware de Entra actualiza `entra_oid` y `tenant_id` despues de crear/leer el usuario por email.
- El modo local de desarrollo no persiste `dev-local` para evitar colisiones entre usuarios de prueba.

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

- [x] Si `/api/me` responde usuario bloqueado o 403, el frontend muestra mensaje claro.
- [x] El usuario no queda atrapado en redirecciones.
- [x] Existe opcion de cerrar sesion.
- [x] No se muestran pantallas internas a usuarios bloqueados.

### Resultado de implementacion

- `auth.js` conserva el estado HTTP del error de `/api/me`.
- El router redirige a `/blocked` cuando el usuario autenticado recibe 403.
- `BlockedView.vue` muestra un mensaje claro de cuenta bloqueada y permite cerrar sesion.
- El login local tambien envia a `/blocked` si el backend rechaza la cuenta por bloqueo.

## AUTH-004 - Solicitar y validar RUT de usuario

Prioridad: P1
Labels: `frontend`, `backend`, `database`, `auth`
Estado sugerido: Done

### Contexto

El sistema puede crear usuarios locales desde Microsoft Entra ID cuando existen en Azure pero no en Poli-REDI. Los usuarios normales deben registrar o confirmar su RUT para poder operar reservas. Los administradores no requieren RUT.

### Objetivo

Guardar RUT en `users`, validarlo en frontend/backend y bloquear reservas de usuarios normales sin RUT.

### Criterios de aceptacion

- [x] `users` incluye columna `rut`.
- [x] La base aplica validacion basica de formato si el RUT existe.
- [x] El backend valida RUT chileno con digito verificador.
- [x] `/api/me` retorna RUT del usuario autenticado.
- [x] Existe endpoint protegido para actualizar RUT del usuario autenticado.
- [x] Usuarios normales sin RUT no pueden crear reservas.
- [x] Administradores pueden operar sin RUT.
- [x] El frontend permite ingresar/actualizar RUT y muestra errores.

### Resultado de implementacion

- Se agrego `rut` a `dbo.users` y un indice unico filtrado para RUT no nulo.
- Se agrego validador backend de RUT con digito verificador.
- Se agrego `PATCH /api/me/rut`.
- `CreateReservation` rechaza usuarios normales sin RUT antes de crear la reserva.
- `SettingsView.vue` permite registrar o actualizar RUT con validacion frontend.
- La app muestra un modal obligatorio a usuarios normales sin RUT antes de permitir reservas.
- En modo desarrollo, el login local no-admin puede resetear RUT para probar el flujo de registro y confirmar cambios en base de datos.

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
- [x] El formulario limita actividades al catalogo aprobado para MVP 1.
- [x] No se usa `activityId` fijo.
- [x] La reserva creada queda asociada a la actividad correcta.
- [x] Se maneja estado sin actividades.

### Resultado de implementacion

- `GET /api/activities` entrega actividades activas desde Azure SQL.
- Para MVP 1, las actividades se seleccionan desde el catalogo existente en base de datos.
- `ReservationForm.vue` muestra un selector de actividades reales.
- `ReservationForm.vue` permite elegir una actividad del catalogo o dejar la reserva sin actividad especifica.
- `AvailabilitySection.vue` carga actividades junto con recursos/reservas y envia el `activityId` seleccionado.
- Si no existen actividades cargadas, el formulario permite dejar la reserva sin actividad especifica.

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
Estado sugerido: Done

### Contexto

El formulario valida campos obligatorios de forma basica, pero no muestra mensajes por campo ni valida rangos de duracion con claridad.

### Objetivo

Agregar validaciones visibles antes de enviar reserva.

### Criterios de aceptacion

- [x] Muestra error si falta recurso.
- [x] Muestra error si falta fecha.
- [x] Muestra error si falta hora.
- [x] Valida duracion mayor a 0.
- [x] Valida participantes mayor a 0.
- [x] Deshabilita boton mientras se crea la reserva.
- [x] Muestra error devuelto por backend sin cerrar modal.

### Resultado de implementacion

- `ReservationForm.vue` valida recurso, fecha, hora y participantes antes de enviar.
- Los campos invalidos muestran mensajes visibles y estado visual de error.
- El boton de confirmacion queda bloqueado mientras se crea la reserva o no hay recursos cargados.
- Los errores de backend se mantienen visibles dentro del modal sin cerrarlo.

## RES-006 - Implementar vista Mis Reservas

Prioridad: P1
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Done

### Contexto

`ReservationsView.vue` lista reservas reales. Para usuarios normales muestra solo reservas propias; para administradores muestra todas las reservas del sistema.

### Objetivo

Crear una vista donde el usuario vea sus reservas.

### Criterios de aceptacion

- [x] Muestra reservas del usuario autenticado.
- [x] Si el usuario es administrador, muestra todas las reservas.
- [x] Muestra recurso, actividad, fecha, hora, duracion y estado.
- [x] Tiene estados de carga, error y vacio.
- [x] Permite ir al detalle de una reserva.
- [x] Usa datos reales del backend.
- [x] `npm run build` pasa.

### Resultado parcial

- `GET /api/reservations/mine` lista reservas del usuario autenticado.
- `GET /api/reservations` permite que la vista admin cargue todas las reservas.
- `ReservationsView.vue` ya no muestra `Proximamente...`.
- `ReservationsView.vue` cambia titulo, mensaje vacio y fuente de datos segun rol.
- La vista permite cancelar reservas propias disponibles.
- `ReservationDetailView.vue` permite revisar una reserva desde `/reservations/:id`; admin puede abrir reservas del listado global.

### Archivos relevantes

- `frontend/src/views/ReservationsView.vue`
- `frontend/src/stores/reservations.js`
- `frontend/src/services/reservations.service.js`

## RES-007 - Implementar cancelacion de reservas desde frontend

Prioridad: P1
Labels: `frontend`, `backend`, `reservas`
Estado sugerido: Partial

### Contexto

Existe `PATCH /api/reservations/cancel`; ya usa usuario autenticado en backend, pero falta conectar un flujo visual completo desde la interfaz.

### Objetivo

Permitir cancelar reservas desde la interfaz usando el usuario autenticado.

### Criterios de aceptacion

- [x] No se usa `requestedByUserId` fijo.
- [x] El backend determina usuario desde token o valida contra usuario autenticado.
- [ ] La UI pide confirmacion fuerte antes de cancelar.
- [x] La reserva cambia a estado `CANCELLED`.
- [x] Se muestra mensaje de exito o error.
- [x] La lista se actualiza sin recargar toda la app.

### Resultado parcial

- Los bloques de reserva en disponibilidad son seleccionables.
- Al seleccionar una reserva se abre `ReservationDetailModal`.
- Admin puede cancelar cualquier reserva; usuario normal solo ve accion de cancelacion para reservas propias.
- El modal muestra errores del backend si la cancelacion falla.
- Al cancelar, la grilla se refresca y el bloque desaparece de la disponibilidad activa.
- Segunda revision UX: el modal muestra una advertencia, pero aun falta una confirmacion fuerte adicional antes de ejecutar la cancelacion.

---

# Hito 4 - Pantallas pendientes

## UI-001 - Implementar vista Recursos

Prioridad: P1
Labels: `frontend`, `recursos`, `feature`
Estado sugerido: Done

### Contexto

`ResourcesView.vue` ya lista recursos reales desde `/api/resources`; falta agregar filtros.

### Objetivo

Mostrar catalogo de recursos deportivos con datos reales.

### Criterios de aceptacion

- [x] Lista recursos desde `/api/resources`.
- [x] Muestra nombre, tipo, modo de reserva, capacidad y estado.
- [x] Permite filtrar por tipo o sede si los datos estan disponibles.
- [x] Tiene estados de carga, error y vacio.
- [x] Mantiene estilo visual actual.

### Resultado parcial

- `ResourcesView.vue` ya no muestra `Proximamente...`.
- `GET /api/resources` ahora incluye `capacity` desde Azure SQL.
- `ResourcesView.vue` permite buscar por nombre/tipo y filtrar por tipo, modo de reserva y estado.

## UI-002 - Implementar detalle de reserva

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Ready for Codex

### Contexto

`ReservationDetailView.vue` ya muestra datos reales de una reserva. Usuarios normales consultan solo sus reservas; administradores pueden abrir reservas del listado global.

### Objetivo

Mostrar informacion detallada de una reserva.

### Criterios de aceptacion

- [x] Muestra recurso, actividad, usuario, fecha, hora, duracion y estado.
- [ ] Muestra participantes si existen.
- [x] Permite volver a Mis Reservas o Disponibilidad.
- [x] Maneja reserva no encontrada.

### Resultado parcial

- Existe ruta `/reservations/:id`.
- La vista carga reservas reales desde `/api/reservations/mine` para usuario normal y desde `/api/reservations` para administrador.
- Permite cancelar desde el detalle si la reserva corresponde.
- Participantes queda pendiente porque el campo no esta persistido en el modelo actual.

## UI-003 - Implementar Historial

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Done

### Contexto

`HistoryView.vue` ya lista reservas historicas reales con comportamiento por rol.

### Objetivo

Mostrar historial de reservas pasadas, canceladas o rechazadas.

### Criterios de aceptacion

- [x] Lista reservas historicas del usuario.
- [x] Si el usuario es administrador, lista todo el historial del sistema.
- [x] Permite filtrar por estado.
- [x] Permite filtrar por fecha.
- [x] Tiene estados de carga, error y vacio.

### Resultado parcial

- `HistoryView.vue` muestra reservas pasadas o canceladas desde `/api/reservations/mine` para usuario normal y desde `/api/reservations` para administrador.
- `HistoryView.vue` permite filtrar por estado y rango de fecha.

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
- Al cerrar sesion desde el menu o desde configuracion, la app limpia estado local y redirige a `/login`.

---

# Hito 5 - Administracion

## ADMIN-001 - Implementar panel administrador base

Prioridad: P1
Labels: `frontend`, `backend`, `admin`, `feature`
Estado sugerido: Done

### Contexto

`AdminView.vue` ya muestra un resumen base. El modelo contempla usuarios admin, recursos, bloqueos, actividades e infracciones para siguientes iteraciones.

### Objetivo

Crear un panel administrador inicial con accesos a gestion de recursos, usuarios, bloqueos y reportes.

### Criterios de aceptacion

- [x] Solo usuarios admin pueden acceder.
- [x] Muestra tarjetas/resumen de recursos y reservas.
- [x] Tiene enlaces a secciones administrativas.
- [x] Maneja usuario sin permisos.

### Resultado de implementacion

- `AdminView.vue` ya no muestra `Proximamente...`.
- El panel carga recursos y reservas desde la API.
- Muestra recursos activos, reservas confirmadas, reservas del dia y proximas reservas.
- El menu lateral oculta la seccion de administracion para usuarios normales y deja configuracion fuera de esa seccion.
- Usuarios, reportes e infracciones quedan como tareas especificas posteriores.

## ADMIN-002 - Implementar gestion de usuarios

Prioridad: P2
Labels: `frontend`, `backend`, `admin`, `auth`
Estado sugerido: Backlog

### Objetivo

Permitir que un admin vea usuarios y pueda bloquear/desbloquear cuentas.

### Criterios de aceptacion

- [x] Existe endpoint para listar usuarios.
- [ ] Existe endpoint para bloquear/desbloquear usuario.
- [x] La UI muestra email, nombre, rol y estado.
- [ ] No permite bloquearse a si mismo accidentalmente.
- [ ] Registra auditoria.

### Resultado parcial

- `GET /api/users` lista usuarios desde Azure SQL solo para administradores.
- `UsersView.vue` ya no muestra `Proximamente...`.
- La vista muestra totales, administradores, usuarios activos y bloqueados.
- Queda pendiente bloquear/desbloquear con auditoria.

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

`ReportsView.vue` ya muestra indicadores iniciales calculados desde reservas y recursos. La base incluye vistas adicionales de uso de recursos, horas punta e infracciones que pueden integrarse despues.

### Objetivo

Mostrar reportes iniciales desde vistas SQL.

### Criterios de aceptacion

- [x] Reporte uso de recursos.
- [x] Reporte horas punta.
- [ ] Reporte infracciones por usuario.
- [x] Estados de carga/error.
- [x] Acceso restringido a admin.

### Resultado parcial

- `ReportsView.vue` ya no muestra `Proximamente...`.
- Muestra reservas activas, horas reservadas, recursos con uso, uso por recurso, estados y horas punta.
- Queda pendiente conectar infracciones y, si corresponde, vistas SQL analiticas dedicadas.

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
- La campana no consulta `/api/notifications` si no hay sesion activa y limpia su estado al cerrar sesion.
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
- [x] Permite filtrar por usuario autenticado.
- [ ] Permite filtrar por estado.
- [ ] Disponibilidad usa rango de fecha.
- [x] Mis Reservas usa usuario autenticado.

### Resultado parcial

- Se agrego `GET /api/reservations/mine` para reservas del usuario autenticado.
- Falta agregar filtros por rango de fecha y estado.

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

## API-004 - Crear endpoint de disponibilidad sanitizado

Prioridad: P1
Labels: `backend`, `frontend`, `reservas`, `disponibilidad`, `security`, `ux`
Estado sugerido: Ready for Codex

### Contexto

La vista de disponibilidad necesita saber que horarios estan ocupados, pero un usuario normal no necesita recibir datos internos de reservas ajenas.

### Objetivo

Exponer un endpoint de disponibilidad por fecha o rango que entregue solo la informacion necesaria para pintar ocupacion.

### Criterios de aceptacion

- [ ] El endpoint acepta fecha o rango de fechas.
- [ ] Para usuario normal no expone `userId` de reservas ajenas.
- [ ] Devuelve recurso, inicio, duracion y tipo de ocupacion.
- [ ] Incluye reservas confirmadas, bloqueos y actividades programadas cuando esten disponibles.
- [ ] La vista de disponibilidad consume este endpoint en lugar de depender de todas las reservas.
- [ ] Admin mantiene acceso a detalle administrativo cuando corresponda.

## API-005 - Centralizar validacion de administrador

Prioridad: P1
Labels: `backend`, `auth`, `admin`, `security`
Estado sugerido: Done

### Contexto

Algunas rutas administrativas validan rol dentro del handler. Para reducir errores futuros, conviene centralizar la regla de autorizacion.

### Objetivo

Crear middleware `RequireAdmin` y aplicarlo a rutas administrativas.

### Criterios de aceptacion

- [x] Existe middleware reutilizable para exigir usuario administrador.
- [x] Rutas como usuarios, reportes administrativos y futuras operaciones admin usan el middleware.
- [x] Usuario normal recibe 403.
- [x] Usuario no autenticado recibe 401 desde el middleware de autenticacion.
- [x] Se mantiene compatibilidad con modo local de desarrollo.

### Resultado de implementacion

- Se agrego `RequireAdmin` en `backend/internal/middleware/auth_middleware.go`.
- `GET /api/users` ahora se registra bajo un grupo admin protegido por `RequireAdmin`.
- `users_handlers.go` queda enfocado en obtener usuarios, sin duplicar validacion de rol.
- Las futuras rutas administrativas pueden colgar del grupo `admin` en `routes.go`.
- `go test ./...` pasa.

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
- [ ] Rechazar usuario normal sin RUT.
- [ ] Rechazar recurso inactivo.
- [ ] Rechazar recurso informativo.
- [ ] Rechazar recurso solo admin para usuario normal.
- [ ] Rechazar conflicto horario.
- [ ] Rechazar usuario bloqueado.
- [ ] Rechazar cruce con bloqueo.
- [ ] Rechazar cruce con actividad programada.
- [ ] Cancelar reserva propia.
- [ ] Cancelar reserva como admin.
- [ ] Rechazar cancelacion sin permisos.
- [ ] Rechazar cancelacion de reserva inexistente.
- [ ] Rechazar cancelacion duplicada.

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
- [ ] Usuario normal no entra a rutas administrativas.
- [ ] Usuario sin RUT no puede abrir creacion de reserva.
- [ ] Conflicto de horario mantiene error visible en el modal.
- [ ] Cancelacion exige confirmacion antes de llamar a la API.

## UX-001 - Profesionalizar experiencia de disponibilidad y reserva

Prioridad: P1
Labels: `frontend`, `ux`, `reservas`, `disponibilidad`
Estado sugerido: Ready for Codex

### Contexto

La vista de disponibilidad ya permite reservar, pero la segunda revision UX detecto mejoras para reducir errores y hacer la experiencia mas institucional.

### Objetivo

Mejorar claridad, control y confianza durante seleccion de horario, creacion y cancelacion de reservas.

### Criterios de aceptacion

- [ ] La seleccion de horario usa intervalos consistentes de 15 o 30 minutos.
- [ ] El usuario ve el rango final completo antes de confirmar.
- [ ] La UI muestra capacidad del recurso cuando exista.
- [ ] Participantes se validan contra capacidad cuando corresponda.
- [ ] Modos de reserva se muestran con etiquetas humanas, no codigos tecnicos.
- [ ] Existe leyenda visual para disponibilidad, reservado, bloqueo, mantencion y actividad institucional.
- [ ] La experiencia movil permite enfocarse en un recurso sin desplazamiento horizontal excesivo.

## UX-002 - Mejorar accesibilidad de modales y estados

Prioridad: P2
Labels: `frontend`, `ux`, `accessibility`
Estado sugerido: Backlog

### Objetivo

Reforzar accesibilidad ligera en flujos criticos.

### Criterios de aceptacion

- [ ] Botones iconicos tienen `aria-label`.
- [ ] Modales cierran con Escape.
- [ ] El foco vuelve al elemento que abrio el modal.
- [ ] Mensajes de exito usan `aria-live`.
- [ ] Errores usan `role="alert"` cuando correspondan.
- [ ] Los estados no dependen solo del color.

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
Estado sugerido: Done

### Contexto

El backend usaba `AllowOrigins: "*"`. Durante el despliegue online se necesitaba limitar y configurar origenes por ambiente.

### Objetivo

Configurar origenes permitidos por ambiente.

### Criterios de aceptacion

- [x] Desarrollo permite localhost.
- [x] Produccion limita origenes autorizados.
- [x] Configuracion viene por variable de entorno.

### Resultado de implementacion

- El backend lee `CORS_ALLOWED_ORIGINS`.
- El entorno local permite `http://localhost:5173`.
- El entorno online permite la URL de Azure Static Web Apps.
- La configuracion queda fuera del codigo y se define por variables de entorno del App Service.

### Archivos relevantes

- `backend/cmd/main.go`
- `backend/.env.example`

## SEC-003 - Limpiar logs de configuracion de autenticacion

Prioridad: P1
Labels: `backend`, `auth`, `security`
Estado sugerido: Done

### Contexto

El middleware de autenticacion imprime valores de configuracion de Entra ID al iniciar. Aunque no son passwords, conviene reducir exposicion en logs.

### Objetivo

Reemplazar logs de configuracion sensible por mensajes de estado seguros.

### Criterios de aceptacion

- [x] No se imprimen tenant, client ID ni issuer completos.
- [x] Si falta configuracion critica, el error es claro para diagnostico.
- [x] No se registran tokens, cabeceras ni datos personales innecesarios.
- [x] `go test ./...` pasa.

### Resultado de implementacion

- `RequireAuth` ya no imprime `ENTRA_TENANT_ID`, `ENTRA_API_CLIENT_ID` ni `ENTRA_ISSUER`.
- El backend solo informa si usa autenticacion Microsoft Entra ID o modo local de desarrollo.
- Se elimino el JWT crudo del usuario guardado en contexto porque no se usaba.
- Se quitaron correos de respuestas internas de error 500 en el middleware de autenticacion.
- Se corrigieron mensajes visibles con caracteres rotos en autenticacion.
- `go test ./...` pasa.

## SEC-004 - Checklist productivo de modo desarrollo

Prioridad: P1
Labels: `backend`, `deploy`, `security`
Estado sugerido: Backlog

### Objetivo

Evitar que `DEV_AUTH_ENABLED=true` quede activo en un ambiente publico.

### Criterios de aceptacion

- [ ] La documentacion de despliegue exige `DEV_AUTH_ENABLED=false`.
- [ ] Existe verificacion manual o automatizada antes de entrega.
- [ ] Se evalua bloquear arranque si modo dev esta activo con origenes productivos.
- [ ] El README explica que las cabeceras `X-Dev-Auth-*` son solo locales.

---

# Hito 9 - Documentacion y entrega FIP/tesis

## DOC-001 - Actualizar README principal

Prioridad: P1
Labels: `documentacion`
Estado sugerido: Ready for Codex

### Objetivo

Actualizar `README.md` con stack real, instalacion, ejecucion, Azure SQL y autenticacion.

### Criterios de aceptacion

- [x] Describe frontend, backend y base de datos actual.
- [x] Explica variables de entorno.
- [x] Explica como ejecutar backend y frontend.
- [x] Explica como validar `/api/health`.
- [x] No contiene secretos.

### Resultado de implementacion

- `README.md` describe stack vigente, Azure SQL, variables de entorno, ejecucion local, rutas principales, demo online y checklist MVP 1.

## DOC-002 - Documentar arquitectura

Prioridad: P1
Labels: `documentacion`, `arquitectura`
Estado sugerido: Done

### Objetivo

Completar `docs/02-arquitectura.md` con diagrama y descripcion de flujo.

### Criterios de aceptacion

- [x] Describe frontend Vue.
- [x] Describe backend Go/Fiber.
- [x] Describe autenticacion Entra ID.
- [x] Describe Azure SQL Database.
- [x] Describe flujo de reserva.
- [x] Incluye diagrama Mermaid.

### Resultado de implementacion

- `docs/02-arquitectura.md` documenta componentes, autenticacion, base de datos, flujo de reserva y despliegue inicial.

## DOC-003 - Documentar flujo de reservas

Prioridad: P1
Labels: `documentacion`, `reservas`
Estado sugerido: Done

### Objetivo

Completar `docs/06-flujo-reservas.md`.

### Criterios de aceptacion

- [x] Describe flujo usuario normal.
- [x] Describe flujo admin.
- [x] Describe validaciones de conflicto.
- [x] Describe estados de reserva.
- [x] Incluye diagrama de secuencia o actividad.

### Resultado de implementacion

- `docs/06-flujo-reservas.md` describe creacion, cancelacion, validaciones frontend/backend/base de datos, reglas UX y pruebas recomendadas.

## DOC-004 - Documentar requisitos, historias y casos de uso

Prioridad: P1
Labels: `documentacion`, `requisitos`
Estado sugerido: Done

### Objetivo

Consolidar requisitos funcionales, requisitos no funcionales, historias de usuario y casos de uso principales de Poli-REDI.

### Criterios de aceptacion

- [x] Existe documento dedicado en `docs/`.
- [x] Incluye actores principales.
- [x] Incluye requisitos funcionales y no funcionales.
- [x] Incluye historias de usuario con criterios de aceptacion.
- [x] Incluye casos de uso principales.
- [x] Incluye trazabilidad inicial contra backlog.

### Resultado de implementacion

- Se creo `docs/08-requisitos-historias-casos-uso.md`.
- El documento queda como fuente viva para alcance funcional, historias, casos de uso y trazabilidad inicial.

## DOC-005 - Documentar roadmap de MVPs

Prioridad: P1
Labels: `documentacion`, `mvp`, `roadmap`
Estado sugerido: Done

### Objetivo

Formalizar los MVPs incrementales de Poli-REDI, indicando que implementa cada uno, su estado y sus criterios de cierre.

### Criterios de aceptacion

- [x] Existe documento dedicado en `docs/`.
- [x] Define cantidad y nombre de MVPs.
- [x] Describe proposito, alcance y estado de cada MVP.
- [x] Relaciona MVPs con tareas del backlog.
- [x] Relaciona MVPs con requisitos, historias o casos de uso.
- [x] Incluye criterios de cierre y dependencias entre MVPs.

### Resultado de implementacion

- Se creo `docs/09-mvps-roadmap.md`.
- El proyecto queda organizado en cuatro MVPs: base tecnica funcional, flujo usuario completo, administracion institucional, y entrega/calidad/soporte.

---

# Hito 10 - Despliegue

## DEPLOY-001 - Definir estrategia de despliegue

Prioridad: P2
Labels: `deploy`, `documentacion`
Estado sugerido: Done

### Objetivo

Definir donde se desplegaran frontend y backend.

### Opciones a evaluar

- Azure App Service para backend.
- Azure Static Web Apps para frontend.
- Variables de entorno en Azure.
- Dominio o URL institucional.

### Criterios de aceptacion

- [x] Documento corto con alternativa elegida.
- [x] Variables necesarias definidas.
- [x] Pasos de despliegue inicial documentados.

### Resultado de implementacion

- Se eligio Azure Static Web Apps para frontend.
- Se eligio Azure App Service para backend desplegado con Docker.
- Se uso Azure SQL Database existente como base de datos.
- Se conecto Microsoft Entra ID para login real en nube.
- Se configuraron variables `VITE_*` en GitHub Actions para compilar el frontend.
- Se agrego `staticwebapp.config.json` para fallback de rutas Vue.
- La demo online quedo operativa en `https://purple-ground-0205c9f10.7.azurestaticapps.net/`.

### Archivos relevantes

- `.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml`
- `frontend/public/staticwebapp.config.json`
- `backend/cmd/main.go`

## DEPLOY-002 - Preparar backend para produccion

Prioridad: P2
Labels: `backend`, `deploy`, `security`
Estado sugerido: Done

### Criterios de aceptacion

- [x] Puerto configurable.
- [x] CORS configurable.
- [x] Logs no exponen secretos.
- [x] Variables de Azure SQL definidas en entorno.
- [x] Health check disponible.

### Resultado de implementacion

- El backend usa `PORT` desde variables de entorno.
- El backend usa `CORS_ALLOWED_ORIGINS` para separar local y nube.
- Las credenciales de Azure SQL se mantienen fuera del repositorio.
- `/api/health` permite validar que el App Service este activo.
- La imagen Docker del backend puede desplegarse en Azure App Service.

## OPS-001 - Plan de corte desde Google Calendar legado

Prioridad: P0
Labels: `operacion`, `documentacion`, `reservas`, `deploy`
Estado sugerido: Ready for Codex

### Contexto

La operacion real puede seguir usando calendarios Google como sistema legado. Si Poli-REDI entra en operacion sin plan de corte, existe riesgo de doble reserva, informacion desactualizada y usuarios operando en dos fuentes distintas.

### Objetivo

Definir y ejecutar un plan minimo para mover la operacion desde Google Calendar hacia Poli-REDI.

### Criterios de aceptacion

- [ ] Calendarios Google vigentes inventariados.
- [ ] Cada calendario legado tiene recurso equivalente en Poli-REDI.
- [ ] Se define fecha y hora de congelamiento del legado.
- [ ] Reservas futuras criticas se migran o registran en Poli-REDI.
- [ ] Se comunica que nuevas reservas se crean solo en Poli-REDI.
- [ ] Google Calendar queda como consulta historica temporal o respaldo.
- [ ] Se valida que no existan dobles reservas en el periodo de transicion.

### Resultado parcial

- Se creo `docs/11-plan-corte-google-calendar.md` con plan de corte operativo, riesgos y criterios de aceptacion.
- La integracion automatica con Google Calendar queda fuera de MVP 1.

### Archivos relevantes

- `docs/11-plan-corte-google-calendar.md`
- `docs/07-backlog.md`
- `docs/09-mvps-roadmap.md`

---

# Inventario actual de datos duros detectados

Revision realizada durante la conexion de actividades reales.

## Datos que ya pueden reemplazarse por Azure SQL

- `frontend/src/components/layout/NotificationBell.vue`: falta marcar notificaciones como leidas; relacionado con `NOTIF-001`.
- No quedan pantallas principales con `Proximamente...` en `frontend/src/views`.

## Datos duros resueltos

- `frontend/src/components/layout/HeaderBar.vue`: el saludo usa el nombre del usuario autenticado.
- `frontend/src/components/layout/UserMenu.vue`: nombre, correo, rol y avatar iniciales vienen del usuario autenticado o de la cuenta Microsoft.
- `frontend/src/views/SettingsView.vue`: muestra datos reales de la cuenta autenticada.
- `frontend/src/views/ReservationsView.vue`: lista reservas propias para usuarios normales y todas las reservas para administradores.
- `frontend/src/views/ReservationDetailView.vue`: muestra detalle real de reservas propias o globales segun rol.
- `frontend/src/views/HistoryView.vue`: lista historial propio para usuarios normales y todo el historial para administradores.
- `frontend/src/views/AdminView.vue`: muestra resumen administrativo con recursos y reservas reales.
- `frontend/src/views/ReportsView.vue`: muestra indicadores calculados desde reservas y recursos reales.
- `frontend/src/views/UsersView.vue`: lista usuarios reales desde Azure SQL para administradores.
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
8. `DOC-004` Requisitos, historias y casos de uso.
9. `DOC-005` Roadmap de MVPs.

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

# Protocolo de mantenimiento de documentacion

Fecha de inicio: 2026-07-05

Desde esta fecha, cualquier intervencion sobre `Poli-REDI` debe cerrar con una revision de la documentacion viva del proyecto.

## Regla operativa

- Si una tarea se completa, actualizar su estado sugerido, criterios de aceptacion y resultado de implementacion.
- Si una tarea queda parcialmente resuelta, mantenerla abierta y agregar resultado parcial con lo pendiente.
- Si durante el trabajo aparece una necesidad nueva, agregarla como tarea nueva en el hito correspondiente.
- Si cambia el alcance de una tarea existente, actualizar contexto, objetivo y archivos relevantes.
- Si cambia una funcionalidad, revisar tambien requisitos, historias de usuario y casos de uso.
- Si cambia arquitectura, instalacion, base de datos, backend, frontend o flujo de reservas, actualizar el documento tecnico correspondiente en `docs/`.
- Si se toca autenticacion, reservas, base de datos o seguridad, dejar evidencia de pruebas o de la razon por la que no pudieron ejecutarse.
- Si no hay impacto documental, registrar explicitamente en el cierre del trabajo que no se requirio cambio en `docs/`.

## Responsable

Codex queda encargado de revisar y mantener `docs/` cada vez que trabaje sobre `Poli-REDI`, con foco minimo en `docs/07-backlog.md`, `docs/08-requisitos-historias-casos-uso.md` y `docs/09-mvps-roadmap.md`.
