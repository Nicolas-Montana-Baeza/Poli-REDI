# Poli-REDI - Backlog maestro

## Objetivo
Este backlog consolida las tareas reales detectadas durante la revision inicial, la migracion a Azure SQL Database y las primeras pruebas funcionales del frontend/backend.

La idea es usar este documento como base para crear issues en GitHub Projects o para delegar tareas puntuales a Codex.

## Estado base verificado

Fecha base funcional: 2026-07-14. Evidencia automatizada actualizada: 2026-08-20.

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
- Reservas activas e Historial comparten actualmente `ReservationsView.vue`; el historial se selecciona mediante `?tab=history` y reutiliza el mismo flujo de detalle.
- `npm run build` pasa en frontend.
- `npm test` ejecuto 25 pruebas correctamente en la verificacion local del 2026-08-20.
- `go test ./...` finaliza correctamente y ya incluye cobertura inicial del reloj de negocio; `QA-001` sigue abierto para ampliar los casos de reservas.
- El MVP 1 permanece reabierto: zona horaria, estado y horario/duracion estan implementados y probados localmente, pero requieren verificacion integrada/online; tambien faltan cobertura, seguridad de errores y coherencia responsive/accesible.

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
- [x] La guia de despliegue indica variables obligatorias y valores seguros.
- [x] Existe checklist de validacion MVP 1 antes de demo.
- [x] Se ejecuta o documenta resultado de `go test ./...`.
- [x] Se ejecuta o documenta resultado de `npm run build`.
- [x] Se revisa que no haya referencias activas a PostgreSQL como tecnologia vigente.
- [ ] Se revisa que no haya secretos, passwords ni tokens en documentacion versionada.

### Resultado parcial

- `docs/10-guia-redeploy.md` fue reescrita como guia operativa de despliegue y redeploy.
- La guia separa escenarios: frontend, backend, variables `VITE_*`, base de datos y seed temporal.
- Se documentaron variables obligatorias locales y online, valores seguros para `DEV_AUTH_ENABLED`, `VITE_DEV_AUTH_ENABLED` y `CORS_ALLOWED_ORIGINS`.
- Se agrego checklist antes de publicar y problemas frecuentes.
- Queda pendiente la revision final de secretos en documentacion versionada antes del cierre definitivo.

### Archivos relevantes

- `README.md`
- `docs/01-instalacion-y-ejecucion.md`
- `docs/02-arquitectura.md`
- `docs/03-base-de-datos.md`
- `docs/09-mvps-roadmap.md`
- `docs/10-guia-redeploy.md`
- `backend/.env.example`
- `frontend/.env.example`

## BACK-005 - Checklist de demo tecnica MVP 1

Prioridad: P1
Labels: `testing`, `documentacion`, `deploy`
Estado sugerido: Validado por usuario

### Contexto

La demo online existe, pero conviene tener una lista repetible para validar que la base tecnica sigue operativa antes de una presentacion.

### Objetivo

Crear un checklist corto de pruebas de humo para MVP 1.

### Criterios de aceptacion

- [x] El checklist incluye validacion de `/api/health` local.
- [x] El checklist incluye validacion de `/api/health` online.
- [x] El checklist incluye validacion de login institucional o modo local segun ambiente.
- [x] El checklist incluye validacion de carga de recursos reales.
- [x] El checklist incluye validacion de carga de actividades reales.
- [x] El checklist incluye validacion de creacion de una reserva de prueba.
- [x] El checklist incluye validacion de cancelacion de una reserva de prueba.
- [x] El checklist incluye validacion de que un usuario normal no acceda a rutas admin.
- [x] El checklist incluye registro de fecha, ambiente y resultado de la ultima validacion.

### Resultado parcial

- Se creo `docs/12-checklist-demo-mvp1.md` con checklist manual para validacion local y online.
- El usuario confirmo que el checklist de demo MVP 1 ya fue validado.

### Evidencia reciente

- 2026-07-06: `go test ./...` ejecutado correctamente en backend.
- 2026-07-06: `npm run build` ejecutado correctamente en frontend.

### Archivos relevantes

- `docs/12-checklist-demo-mvp1.md`
- `docs/07-backlog.md`
- `docs/09-mvps-roadmap.md`

## BACK-006 - Normalizar mensajes y codificacion visible

Prioridad: P2
Labels: `backend`, `frontend`, `ux`, `refactor`, `codex-ready`
Estado sugerido: En revision de usuario

### Contexto

Durante revisiones se observaron textos con problemas de codificacion, por ejemplo acentos mostrados como caracteres rotos en algunos archivos.

### Objetivo

Revisar mensajes visibles de backend/frontend para que no aparezcan caracteres corruptos ni textos tecnicos innecesarios durante la demo.

### Criterios de aceptacion

- [x] No hay mensajes visibles con caracteres corruptos.
- [x] Errores frecuentes de autenticacion y reserva usan texto claro.
- [x] Los mensajes tecnicos quedan reservados para logs o `detail` cuando corresponda.
- [x] `npm run build` pasa.
- [x] `go test ./...` pasa.

### Resultado parcial

- Se corrigieron textos visibles sin acentos en login, header, menu de usuario, configuracion, historial, reservas y detalle de reserva.
- Los errores de red del frontend ahora muestran un mensaje claro cuando el backend no responde.
- Los errores de autenticacion y reserva del backend usan mensajes principales amigables y dejan detalles tecnicos en `detail` cuando corresponde.
- No se encontraron caracteres corruptos tipo `Ã`, `Â` o `�` en frontend/backend/documentacion.
- La tarea queda en revision de usuario hasta validar visualmente los textos en la app.

### Evidencia reciente

- 2026-07-08: `npm run build` ejecutado correctamente en frontend.
- 2026-07-08: `go test ./...` ejecutado correctamente en backend.

## BACK-007 - Crear base global de sistema visual MVP 1

Prioridad: P1
Labels: `frontend`, `ux`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

El proyecto necesitaba una base global porque tarjetas, botones, estados, radios, colores y sombras se redefinian dentro de varios componentes con `<style scoped>`. Esa base ya existe en `frontend/src/assets/styles/variables.css` y `frontend/src/assets/styles/main.css`; queda pendiente la validacion visual final.

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

### Resultado parcial

- La base global existe, compila y cumple su objetivo estructural.
- Se ajustaron radios, sombras, pesos tipograficos y contraste para reducir una apariencia demasiado pesada.
- Se reemplazo el sidebar oscuro por una version clara y se aumento la viveza del color primario.
- Los defectos responsive y de componentes compartidos se siguen en `BACK-018`, `BACK-021` y `BACK-022`; no reabren la creacion de tokens.

### Archivos relevantes

- `frontend/src/style.css`
- `frontend/src/assets/styles/variables.css`
- `frontend/src/assets/styles/main.css`

## BACK-008 - Aplicar sistema visual global a pantallas MVP 1

Prioridad: P1
Labels: `frontend`, `ux`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Partially Done

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
- [ ] La UI sigue siendo responsive en mobile sin superposiciones ni controles fuera del viewport.
- [x] `npm run build` pasa.

### Resultado parcial

- Las pantallas principales usan la base global, pero el resultado visual sigue sujeto a revision del usuario.
- Se redujo la intensidad visual en login, header, sidebar, tarjetas de reserva y detalle.
- Se ajustaron botones, mini calendario y timeline de disponibilidad para que se vean mas consistentes y menos apagados.
- La revision de 2026-07-14 encontro quiebres en header, campana, sidebar, modales y carrusel. El cierre concreto queda repartido en `BACK-018`, `BACK-021` y `BACK-022`.

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
- `frontend/src/views/ReservationsView.vue`
- `frontend/src/views/ResourcesView.vue`

## BACK-009 - Separar estado real de categoria temporal de reserva

Prioridad: P1
Labels: `frontend`, `ux`, `reservas`, `refactor`, `mvp1`
Estado sugerido: Done

### Contexto

Durante la revision UX se detecto una ambiguedad conceptual: una reserva `CONFIRMED` podia mostrarse como `Confirmada` en Mis Reservas y como `Finalizada` en Historial. Esto es correcto solo si `Finalizada` se entiende como etiqueta visual derivada por tiempo, no como estado persistido.

Para una demo y defensa tecnica, conviene separar claramente:

- Estado real persistido: `PENDING`, `CONFIRMED`, `CANCELLED`, `REJECTED`, `EXPIRED`.
- Categoria temporal derivada: futura, en curso o pasada.
- Vista: Mis Reservas o Historial.
- Etiqueta UI: Confirmada, En curso, Finalizada, Cancelada, Pendiente.

### Objetivo

Evitar que "activa", "historica" o "finalizada" se traten como estados reales cuando son clasificaciones derivadas por fecha y contexto.

### Criterios de aceptacion

- [x] Existe una utilidad compartida para calcular fecha fin de reserva.
- [x] Existe una utilidad compartida para clasificar reserva como futura, en curso o pasada.
- [x] Historial usa clasificacion derivada para mostrar reservas pasadas o canceladas.
- [x] Mis Reservas de usuario normal muestra reservas accionables, no historicas.
- [x] La etiqueta `Finalizada` se deriva de `CONFIRMED` + fecha pasada.
- [x] El detalle usa la misma etiqueta visual que las listas.
- [x] `npm run build` pasa.

### Resultado de implementacion

- Se agregaron helpers compartidos en `frontend/src/utils/reservationTime.js`.
- `ReservationsView.vue` usa `isReservationHistorical`, `isReservationActionable` y `getReservationDisplayStatus` para los tabs activo e historico.
- `ReservationsView.vue` usa `isReservationActionable`, `isReservationCancelable` y `getReservationDisplayStatus`.
- `ReservationDetailView.vue` usa la misma regla de etiqueta visual y cancelacion.
- `npm run build` paso correctamente el 2026-07-07.

### Archivos relevantes

- `frontend/src/utils/reservationTime.js`
- `frontend/src/views/ReservationsView.vue`
- `frontend/src/views/ReservationDetailView.vue`

## BACK-010 - Unificar tarjetas/listado de Mis Reservas e Historial

Prioridad: P1
Labels: `frontend`, `ux`, `reservas`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Actualizacion 2026-08-20

La unificacion se completo tambien a nivel de navegacion.

`ReservationsView.vue` concentra actualmente:

- reservas activas;
- historial mediante `?tab=history`;
- filtros historicos;
- apertura de detalle;
- cancelacion segun permisos.

La antigua `HistoryView.vue` fue eliminada. `/history` se conserva como redireccion al tab historico.


### Contexto

Mis Reservas e Historial son vistas hermanas: ambas muestran reservas con recurso, actividad, fecha, horario, duracion, estado visual y acceso al detalle. Sin embargo, hoy tienen estructuras y estilos distintos, lo que genera una sensacion de falta de coherencia visual.

La diferencia conceptual debe mantenerse:

- Mis Reservas: reservas accionables o vigentes para el usuario normal.
- Historial: reservas pasadas o canceladas.

Lo que debe unificarse es el lenguaje visual y la estructura de tarjeta.

### Objetivo

Crear un componente compartido para representar reservas en listas, reutilizable por los contextos activo e historico de `ReservationsView.vue` y por otras superficies como el dashboard.

### Alcance sugerido para Codex

1. Crear un componente reutilizable, por ejemplo `frontend/src/components/reservations/ReservationListCard.vue`.
2. Mover a ese componente la estructura visual comun de tarjeta.
3. Recibir por props la reserva, etiqueta visual, modo y flags de acciones.
4. Emitir eventos o usar slots para acciones como `Detalle`, `Ver detalle` y `Cancelar`.
5. Usar el componente desde los contextos activo e historico de `ReservationsView.vue`.

### Props/eventos sugeridos

- `reservation`
- `mode`: `active` o `history`
- `showCancel`
- `showUser`
- `detailTo`
- `cancelDisabled`
- `@cancel`
- `@open-detail`

### Criterios de aceptacion

- [x] Mis Reservas e Historial usan el mismo componente base de tarjeta.
- [x] Ambas vistas muestran los mismos datos principales con el mismo orden visual.
- [x] Las diferencias de accion se controlan por props o slots, no duplicando tarjeta completa.
- [x] El estado visual usa `getReservationDisplayStatus`.
- [x] La tarjeta conserva acceso a detalle desde ambas vistas.
- [x] La accion cancelar solo aparece donde corresponde.
- [x] La tarjeta es usable con teclado.
- [x] El responsive de ambas vistas queda consistente.
- [x] `npm run build` pasa.

### Resultado de implementacion

- Se creo `ReservationListCard.vue` como tarjeta compartida para listas de reservas.
- `ReservationsView.vue` usa la tarjeta compartida con accion de detalle y cancelacion.
- El tab historico de `ReservationsView.vue` usa la tarjeta compartida en modo historial, con el mismo acceso al detalle.
- `ReservationsPanel.vue` del dashboard tambien usa la tarjeta compartida para proximas reservas.
- Se retiro `components/dashboard/ReservationCard.vue` para evitar dos tarjetas de reserva conviviendo.
- Se redujo duplicacion de markup, badges, metadatos de fecha/hora/duracion y estilos de tarjeta.
- `npm run build` paso correctamente el 2026-07-07.

### Archivos relevantes

- `frontend/src/components/reservations/ReservationListCard.vue`
- `frontend/src/components/dashboard/ReservationsPanel.vue`
- `frontend/src/views/ReservationsView.vue`
- `frontend/src/utils/reservationTime.js`

## BACK-011 - Unificar estados, filtros y vacios en vistas de reservas

Prioridad: P2
Labels: `frontend`, `ux`, `reservas`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Actualizacion 2026-08-20

Los estados activos e historicos utilizan la misma vista y los mismos helpers temporales.

Las reservas accionables se ordenan de forma ascendente y el historial de forma descendente. Los filtros de estado y fecha se mantienen solamente en el contexto historico.


### Contexto

Aunque exista una tarjeta compartida, Mis Reservas e Historial tambien deben coincidir en estados de carga, error, vacio, filtros y microcopy. Hoy cada vista define parte de esos patrones localmente.

### Objetivo

Alinear los estados visuales y de interaccion de las vistas relacionadas con reservas.

### Alcance sugerido para Codex

1. Revisar `state-card`, mensajes de carga, errores y vacios en ambas vistas.
2. Usar clases globales del sistema visual cuando existan.
3. Mantener filtros solo en Historial, pero con estilo coherente con formularios globales.
4. Alinear copy entre ambas vistas para que la diferencia sea conceptual, no accidental.

### Criterios de aceptacion

- [x] Mis Reservas e Historial usan el mismo estilo de carga, error, exito y vacio.
- [x] Los filtros de Historial usan estilo coherente con campos globales.
- [x] Los badges de estado son consistentes entre lista, historial y detalle.
- [x] El texto diferencia claramente reservas accionables e historicas.
- [x] Mobile mantiene filtros y tarjetas legibles a 360px de ancho.
- [x] `npm run build` pasa.

### Resultado de implementacion

- El contexto historico de `ReservationsView.vue` usa la base global para tarjetas y filtros.
- Mis Reservas, Historial y Proximas Reservas del dashboard quedan alineadas en tarjeta, estados y badges.
- El componente compartido mantiene layout responsive para tablets y mobile.
- `npm run build` paso correctamente el 2026-07-07.

### Archivos relevantes

- `frontend/src/views/ReservationsView.vue`
- `frontend/src/views/ReservationDetailView.vue`
- `frontend/src/assets/styles/main.css`
- `frontend/src/assets/styles/variables.css`

## BACK-012 - Alinear formulario de reserva con datos realmente persistidos

Prioridad: P1
Labels: `frontend`, `ux`, `reservas`, `bug`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

La revision UX detecto que `ReservationForm.vue` pide "Participantes" y valida `participantsCount`, pero `AvailabilitySection.vue` no envia ese dato al backend y `CreateReservationRequest` tampoco lo recibe. Para MVP 1 esto genera una promesa falsa: el usuario entrega un dato que el sistema no guarda.

### Objetivo

Evitar que el formulario capture datos que no forman parte del contrato real de creacion de reservas en MVP 1.

### Alcance sugerido para Codex

1. Revisar `frontend/src/components/forms/ReservationForm.vue`.
2. Retirar u ocultar el campo `Participantes` del formulario de creacion para MVP 1.
3. Eliminar validaciones, errores y estado local asociados a `participantsCount` si ya no se muestran.
4. Mantener la capacidad del recurso como dato informativo solo si ya existe en `resource.capacity`.
5. Confirmar que el payload creado en `AvailabilitySection.vue` sigue coincidiendo con `CreateReservationRequest`.
6. No modificar base de datos ni backend en esta tarea; la persistencia real de participantes queda en `RES-008`.

### Criterios de aceptacion

- [x] El usuario no ve un campo de participantes si ese dato no se persiste.
- [x] No queda validacion visible ni bloqueo por `participantsCount` en MVP 1.
- [x] El payload de creacion contiene solo campos soportados por backend.
- [x] La UI sigue mostrando recurso, fecha, hora, duracion y actividad cuando corresponda.
- [x] `npm run build` pasa.

### Resultado de implementacion

- El formulario de MVP 1 ya no solicita participantes ni conserva validaciones asociadas.
- El payload visible coincide con el contrato actual de creacion.
- La confirmacion de participantes es obligatoria para recursos grupales por decision del 2026-07-20 y se mantiene separada en `RES-008` para MVP 2.
- 2026-07-14: `npm run build` ejecutado correctamente.

### Archivos relevantes

- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `backend/internal/models/reservation.go`

## BACK-013 - Retirar controles visibles sin accion en disponibilidad

Prioridad: P1
Labels: `frontend`, `ux`, `disponibilidad`, `bug`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

En la barra de disponibilidad existen controles que parecen interactivos pero no completan ninguna accion: el boton de fecha emite `open-calendar` sin listener en `AvailabilitySection.vue`, y el boton `Filtros` no abre ni aplica filtros. En una demo esto puede sentirse como funcionalidad rota.

### Objetivo

Eliminar o neutralizar controles que prometen acciones inexistentes en MVP 1.

### Alcance sugerido para Codex

1. Revisar `CalendarToolbar.vue` y su uso en `AvailabilitySection.vue`.
2. Quitar el boton `Filtros` si no se implementara en MVP 1.
3. Convertir el boton de fecha en texto/indicador no clickeable, o conectar una accion real si se decide usar el mini calendario.
4. Mantener los botones de dia anterior, dia siguiente y hoy.
5. Revisar responsive de la barra en mobile.

### Criterios de aceptacion

- [x] No hay botones visibles que no produzcan una accion clara en la toolbar de disponibilidad.
- [x] La fecha actual sigue visible en la toolbar.
- [x] Navegar dia anterior/siguiente/hoy sigue funcionando.
- [x] El layout mobile no queda con espacios vacios o botones fantasma.
- [x] `npm run build` pasa.

### Resultado de implementacion

- Se retiraron los controles de fecha/filtros que no completaban una accion.
- La navegacion por dia y la accion `Hoy` se mantienen operativas.
- 2026-07-14: validado en navegador y con build de produccion.

### Archivos relevantes

- `frontend/src/components/availability/CalendarToolbar.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`

## BACK-014 - Rechazar cancelacion de reservas finalizadas en backend

Prioridad: P1
Labels: `backend`, `reservas`, `security`, `bug`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

El frontend oculta la accion de cancelar cuando una reserva ya esta en el pasado. El backend ahora valida tambien la ventana temporal antes de cancelar, ademas de permisos y estado `CANCELLED`.

### Objetivo

Hacer que la regla "no se cancelan reservas ya finalizadas" exista en backend, no solo en la interfaz.

### Alcance sugerido para Codex

1. Revisar `backend/internal/services/reservations_service.go`.
2. Extender la consulta del repositorio para obtener `start_time` y `duration_minutes` junto con propietario/estado, o crear una funcion especifica.
3. En `CancelReservation`, calcular fecha/hora de termino y rechazar si ya paso.
4. Retornar un mensaje amigable, por ejemplo: `No puedes cancelar una reserva ya finalizada`.
5. Mantener permisos actuales: admin puede cancelar cualquier reserva no finalizada; usuario normal solo reservas propias.
6. Agregar o actualizar pruebas backend si aplica.

### Criterios de aceptacion

- [x] Backend rechaza cancelar reservas cuyo termino ya paso.
- [x] Backend conserva rechazo por reserva inexistente, sin permisos y ya cancelada.
- [x] Frontend muestra el error del backend si la API rechaza la cancelacion.
- [x] `go test ./...` pasa.

### Resultado de implementacion

- `GetReservationCancellationSnapshot` obtiene propietario, estado, inicio y duracion de la reserva.
- `CancelReservation` calcula el termino de la reserva y retorna `no puedes cancelar una reserva finalizada` si ya paso.
- La regla aplica tanto a usuarios normales como a administradores; los permisos existentes se mantienen.
- 2026-07-14: `go test ./...` finalizo correctamente sin pruebas descubiertas. Evidencia superada el 2026-07-20: ya existen casos reales; la ampliacion de cobertura continua en `QA-001`.

### Archivos relevantes

- `backend/internal/services/reservations_service.go`
- `backend/internal/repositories/reservations_repository.go`
- `backend/internal/handlers/reservations_handlers.go`
- `frontend/src/stores/reservations.js`

## BACK-015 - Unificar estado visual en modal de disponibilidad

Prioridad: P2
Labels: `frontend`, `ux`, `reservas`, `disponibilidad`, `refactor`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

Historicamente el detalle de Disponibilidad mantenia una implementacion separada del estado visual. Actualmente el flujo fue unificado y `ReservationForm.vue` en modo `detail` reutiliza `getReservationDisplayStatus`, evitando etiquetas divergentes entre lista, detalle e inspeccion desde disponibilidad.

### Objetivo

Usar una sola regla de estado visual para reservas en todo el flujo MVP 1.

### Alcance sugerido para Codex

1. Mantener `getReservationDisplayStatus` como helper comun del detalle.
2. Evitar volver a introducir logica local de estado en modales alternativos.
3. Mantener tratamiento especial de talleres con etiqueta `Taller programado`.
4. Si se muestra badge de estado, usar clases compatibles con `app-badge` o estilos ya existentes.
5. Revisar copy de advertencia para cancelacion segun estado real.

### Criterios de aceptacion

- [x] La misma reserva muestra el mismo estado visual en lista, detalle e inspeccion desde disponibilidad.
- [x] Reservas `CONFIRMED` pasadas se ven como `Finalizada`.
- [x] Reservas en curso se ven como `En curso`.
- [x] Talleres siguen diferenciados como talleres programados.
- [x] `npm run build` pasa.

### Resultado de implementacion

- `ReservationForm.vue` en modo `detail` usa el helper compartido de estado visual.
- Lista, detalle e inspeccion desde disponibilidad conservan la misma etiqueta temporal.
- 2026-07-14: validado en navegador y con build de produccion.

### Archivos relevantes

- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/utils/reservationTime.js`
- `frontend/src/views/ReservationDetailView.vue`
- `frontend/src/components/reservations/ReservationListCard.vue`

## BACK-016 - Sincronizar mini calendario con navegacion de dias

Prioridad: P2
Labels: `frontend`, `ux`, `disponibilidad`, `bug`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

`CalendarMini.vue` mantiene `currentMonth` y `currentYear` como estado interno inicial. Si el usuario navega con dia anterior, dia siguiente u hoy desde `AvailabilitySection.vue`, el dia seleccionado cambia, pero el mini calendario puede quedar mirando otro mes.

### Objetivo

Mantener sincronizada la vista mensual del mini calendario con `selectedDate`.

### Alcance sugerido para Codex

1. Agregar un `watch` sobre `selectedDate` en `CalendarMini.vue`.
2. Parsear `selectedDate` de forma local y segura.
3. Actualizar `currentMonth` y `currentYear` cuando la fecha seleccionada cambie desde afuera.
4. Mantener navegacion manual de meses dentro del mini calendario.
5. Evitar regresiones con fechas pasadas deshabilitadas.

### Criterios de aceptacion

- [x] Al presionar `Hoy`, el mini calendario muestra el mes correspondiente a hoy.
- [x] Al navegar a un dia de otro mes, el mini calendario cambia a ese mes.
- [x] El dia seleccionado queda destacado correctamente.
- [x] Las fechas pasadas siguen deshabilitadas.
- [x] `npm run build` pasa.

### Resultado de implementacion

- `CalendarMini.vue` observa `selectedDate` y sincroniza mes y ano cuando cambia desde afuera.
- La navegacion manual y las fechas pasadas deshabilitadas se mantienen.
- 2026-07-14: build de produccion correcto.

### Archivos relevantes

- `frontend/src/components/availability/CalendarMini.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`

## BACK-017 - Hacer accesible la seleccion de instalacion en formulario

Prioridad: P2
Labels: `frontend`, `ux`, `accessibility`, `reservas`, `codex-ready`, `mvp1`
Estado sugerido: Done

### Contexto

`ResourcePicker.vue` usa `div` clickeables para seleccionar instalaciones. Funciona con mouse, pero no expone semantica de boton ni estado seleccionado para teclado/lectores de pantalla.

### Objetivo

Mejorar accesibilidad ligera del formulario de reserva sin cambiar el diseno visual.

### Alcance sugerido para Codex

1. Cambiar cada `resource-card` interactiva a `<button type="button">`.
2. Mantener clases y layout actuales.
3. Agregar `aria-pressed` o equivalente para indicar seleccion.
4. Revisar foco visible y estado disabled si aplica.
5. Confirmar que seleccionar recurso sigue emitiendo `select`.

### Criterios de aceptacion

- [x] Las instalaciones se pueden seleccionar con teclado.
- [x] El foco visible es claro.
- [x] El estado seleccionado queda anunciado semanticamente.
- [x] La apariencia visual no cambia de forma disruptiva.
- [x] `npm run build` pasa.

### Resultado de implementacion

- Las tarjetas de instalacion usan controles `button` con semantica de seleccion.
- Se conserva el layout visual y existe foco visible para teclado.
- Los estados no reservables se cubren de forma separada en `BACK-020`.

### Archivos relevantes

- `frontend/src/components/forms/ResourcePicker.vue`

## BACK-018 - Pulir dashboard y carrusel de instalaciones MVP 1

Prioridad: P1
Labels: `frontend`, `ux`, `dashboard`, `visual`, `codex-ready`, `mvp1`
Estado sugerido: Implementado; validacion responsive pendiente

### Actualizacion 2026-08-20

El carrusel fue simplificado para eliminar movimiento automatico y duplicacion por loop.

Comportamiento vigente:

- desplazamiento horizontal manual;
- sin autoplay;
- soporte para mouse, touch y trackpad;
- controles laterales;
- tarjeta completa seleccionable;
- navegacion a `/availability?resource=<id>`;
- Disponibilidad interpreta el recurso enviado y lo deja seleccionado.

La tarea restante es validar visualmente el comportamiento responsive en los anchos objetivo; no queda pendiente redisenar el mecanismo del carrusel.


### Contexto

Durante el pulido final del MVP 1 se detecto que el dashboard perdio claridad visual: los accesos rapidos resultaban redundantes, el carrusel dejo de sentirse como carrusel horizontal y algunas imagenes se veian mal encuadradas.

### Objetivo

Dejar el dashboard inicial mas limpio, con instalaciones visibles en un carrusel horizontal elegante y sin accesos rapidos redundantes.

### Alcance sugerido para Codex

1. Eliminar o mantener fuera de la vista principal los accesos rapidos redundantes.
2. Mantener las acciones principales disponibles desde navegacion existente.
3. Mantener el carrusel como una banda horizontal de desplazamiento exclusivamente manual.
4. Evitar que la primera o ultima tarjeta quede cortada en desktop y mobile.
5. Permitir control mediante mouse, touch, trackpad y controles laterales sin depender de una barra de scroll visible.
6. No usar loop ni copias duplicadas de tarjetas.
7. No utilizar autoplay ni movimiento continuo.
8. Asegurar que las imagenes de recursos mantengan proporcion y encuadre consistente.
9. Evitar que el carrusel rompa el layout en mobile.

### Criterios de aceptacion

- [x] Los accesos rapidos redundantes no aparecen en el dashboard principal.
- [x] El carrusel se percibe como carrusel horizontal.
- [x] El carrusel no utiliza autoplay ni movimiento automatico.
- [x] El usuario puede desplazarse manualmente sin depender de una barra de scroll visible.
- [ ] La primera y ultima tarjeta no quedan cortadas al cargar en desktop o mobile.
- [ ] Los controles manuales producen un desplazamiento perceptible y estable.
- [x] No existen copias de loop que dupliquen enlaces.
- [x] No existe movimiento continuo que deba desactivarse mediante `prefers-reduced-motion`.
- [ ] Las imagenes de recursos se ven proporcionadas y consistentes.
- [ ] El dashboard mantiene buen aspecto en desktop y mobile.
- [x] `npm run build` pasa.

### Resultado parcial

- Se eliminaron los accesos rapidos del dashboard principal.
- El carrusel fue reemplazado por desplazamiento horizontal manual, sin autoplay ni loop.
- Se estabilizo la proporcion visual de las tarjetas de instalaciones.
- La revision de 2026-07-14 detecto una tarjeta inicial cortada en desktop, contenido parcial en mobile, controles manuales que compiten con la animacion y ausencia de reduccion de movimiento.
- La tarea permanece abierta hasta resolver esos puntos y repetir validacion visual en 360, 611 y 1440 px.

### Evidencia reciente

- 2026-07-14: `npm run build` ejecutado correctamente en frontend.

### Archivos relevantes

- `frontend/src/views/DashboardView.vue`
- `frontend/src/components/dashboard/FacilityCarousel.vue`
- `frontend/src/components/dashboard/FacilityCard.vue`

## BACK-019 - Crear seed temporal de pruebas sin modificar seed base

Prioridad: P2
Labels: `database`, `testing`, `demo`, `mvp1`
Estado sugerido: Validado por usuario

### Contexto

Para probar disponibilidad, reservas, bloqueos y actividades del dia actual se necesita un seed temporal con fechas de hoy, pero sin alterar `database/seed.sql`, que debe mantenerse como dato base estable.

### Objetivo

Agregar un archivo SQL temporal que pueda ejecutarse despues del seed normal para reemplazar datos demo por datos del dia de prueba.

### Criterios de aceptacion

- [x] `database/seed.sql` no queda modificado por el seed temporal.
- [x] Existe un archivo SQL separado para datos temporales de hoy.
- [x] El archivo temporal limpia datos demo dependientes antes de reinsertarlos.
- [x] Las reservas, participantes, bloqueos y actividades usan fecha de hoy para pruebas.
- [x] Se valida ejecutandolo en una base local despues de `drop -> schema -> seed`.

### Resultado de validacion

- Se creo `database/seed_today_temp.sql` como overlay temporal posterior al seed normal.
- El usuario confirmo que el seed temporal fue validado.

### Archivos relevantes

- `database/seed_today_temp.sql`

## BACK-020 - Impedir seleccion de instalaciones no reservables

Prioridad: P1
Labels: `frontend`, `ux`, `reservas`, `recursos`, `bug`, `codex-ready`, `mvp1`
Estado actual: En revision

### Contexto

La disponibilidad y el selector del formulario muestran recursos inactivos, informativos o restringidos a administracion como si fueran seleccionables. El backend los rechaza al confirmar, pero el usuario ya recorrio un camino que parecia valido.

### Objetivo

Reflejar en la interfaz la misma elegibilidad que aplica el backend, manteniendolo como fuente final de verdad.

### Alcance sugerido para Codex

1. Crear o reutilizar un helper frontend que determine si un recurso es reservable para el rol actual segun `status` y `reservationMode`.
2. Aplicar el helper en `ResourcePicker.vue`, `ResourceTimeline.vue` y la apertura de `ReservationForm.vue`.
3. Mostrar recursos no reservables como deshabilitados, sin abrir el formulario ni emitir una seleccion valida.
4. Comunicar el motivo con copy breve: `No disponible`, `Solo administracion` o `Uso informativo`.
5. Mantener visibles los recursos si aportan contexto; no ocultarlos salvo que el patron actual ya lo requiera.
6. No relajar las validaciones backend existentes.

### Criterios de aceptacion

- [ ] Un usuario normal no puede abrir el formulario desde un recurso inactivo, informativo o `ADMIN_ONLY`.
- [ ] Un administrador conserva las acciones permitidas por las reglas actuales.
- [ ] El estado deshabilitado se reconoce visualmente, por teclado y con `aria-disabled` o `disabled`.
- [ ] Los recursos `OPEN_USE` mantienen su comportamiento de uso libre.
- [ ] Un payload manipulado sigue siendo rechazado por backend.
- [ ] `npm run build` pasa.

### Pruebas minimas

- Recurso activo y reservable permite abrir formulario.
- Recurso inactivo no abre formulario.
- Recurso informativo no abre formulario.
- Recurso solo administracion se bloquea para usuario normal.

### Archivos relevantes

- `frontend/src/components/forms/ResourcePicker.vue`
- `frontend/src/components/availability/ResourceTimeline.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/components/forms/ReservationForm.vue`

## BACK-021 - Corregir shell responsive y controles globales

Prioridad: P1
Labels: `frontend`, `ux`, `responsive`, `accessibility`, `layout`, `codex-ready`, `mvp1`
Estado sugerido: Ready for Codex

### Contexto

La revision en 360 px detecto que el saludo del header se superpone con la campana, el dropdown de notificaciones sale del viewport y el sidebar cerrado conserva enlaces enfocables fuera de pantalla. Ademas, `Ver todas` en la campana parece accionable pero no ejecuta ninguna accion.

### Objetivo

Hacer que el layout compartido sea estable, operable y coherente en todas las pantallas del MVP 1.

### Alcance sugerido para Codex

1. Ajustar el saludo para que pueda truncarse, ocultarse o cambiar de linea sin invadir acciones del header.
2. Limitar el dropdown de notificaciones al ancho disponible y mantenerlo dentro del viewport desde 320 px.
3. Retirar `Ver todas` mientras no exista destino, o conectarlo a una ruta funcional.
4. Impedir que enlaces del sidebar cerrado reciban foco; usar `inert`, desmontaje condicional o una estrategia equivalente.
5. Mantener cierre del sidebar al navegar y al activar el overlay en mobile.
6. Dejar un solo `h1` principal por pantalla; logo y saludo deben usar semantica secundaria.
7. Validar Dashboard, Disponibilidad, Mis Reservas, Historial, Recursos, Talleres y Reportes.

### Criterios de aceptacion

- [ ] No hay superposiciones en 320, 360, 611, 768 y 1440 px.
- [ ] La campana y el menu de usuario siempre permanecen visibles y utilizables.
- [ ] El dropdown de notificaciones no sale del viewport.
- [ ] No existen acciones visibles sin efecto dentro del header.
- [ ] El sidebar cerrado no expone enlaces en el orden de tabulacion.
- [ ] Cada pantalla tiene un unico encabezado principal anunciado como `h1`.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/components/layout/HeaderBar.vue`
- `frontend/src/components/layout/NotificationBell.vue`
- `frontend/src/components/layout/Sidebar.vue`
- `frontend/src/components/layout/AppLayout.vue`

## BACK-022 - Completar accesibilidad de modales y calendario

Prioridad: P1
Labels: `frontend`, `ux`, `accessibility`, `reservas`, `codex-ready`, `mvp1`
Estado sugerido: Ready for Codex

### Contexto

Al abrir los modales de reserva o detalle, el foco permanece en el control de fondo. Los dialogos no declaran semantica completa, no atrapan foco ni garantizan cierre con Escape. Los inputs de fecha/hora y varios botones iconicos tampoco tienen nombre accesible asociado.

### Objetivo

Permitir completar el flujo critico de reserva y detalle con teclado y lectores de pantalla sin alterar el diseno visual.

### Alcance sugerido para Codex

1. Agregar `role="dialog"`, `aria-modal="true"` y titulo asociado a cada modal critico.
2. Mover el foco al primer control util al abrir, mantenerlo dentro del dialogo y devolverlo al disparador al cerrar.
3. Cerrar con Escape cuando la accion no este procesandose.
4. Asociar cada `label` con su input mediante `for`/`id` o estructura equivalente.
5. Agregar `aria-label` a navegacion anterior/siguiente y demas botones solo icono.
6. Convertir la seleccion de hora sobre timeline en una interaccion operable por teclado o proveer una alternativa equivalente.
7. Mantener errores con `role="alert"` o `aria-live` sin borrar el contexto del formulario.

### Criterios de aceptacion

- [ ] El foco entra al modal, no escapa con Tab y vuelve al control de origen al cerrar.
- [ ] Escape cierra los modales cuando no hay una operacion en curso.
- [ ] Fecha, hora, duracion y actividad tienen nombre accesible.
- [ ] Los botones iconicos del calendario anuncian su accion.
- [ ] Se puede iniciar una reserva sin depender exclusivamente del mouse.
- [ ] Los errores son anunciados y permanecen visibles.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/forms/DateTimePicker.vue`
- `frontend/src/components/availability/CalendarToolbar.vue`
- `frontend/src/components/availability/ResourceTimeline.vue`

## BACK-023 - Estabilizar navegacion y carga de sesion

Prioridad: P1
Labels: `frontend`, `router`, `auth`, `performance`, `bug`, `codex-ready`, `mvp1`
Estado sugerido: Partially Done

### Actualizacion 2026-08-20

Se implemento un bootstrap global de autenticacion mediante el store.

El estado `initialized` permite distinguir entre:

- sesion aun no resuelta;
- usuario autenticado;
- usuario anonimo.

`AuthLoadingScreen.vue` se reutiliza durante inicializacion, callback y cierre de sesion. El router evita mostrar prematuramente `/login` mientras la sesion se esta resolviendo.

Tambien se corrigio la transicion de logout para evitar el flash `Preparando tu cuenta` entre cierre y login.

Pendiente de esta tarea:

- revisar llamadas residuales a `loadAuthUser()` desde vistas individuales;
- completar/verificar el comportamiento 404;
- ampliar pruebas automatizadas especificas del router y bootstrap.


### Contexto

Una ruta inexistente deja el contenido principal vacio aunque existe `NotFoundView.vue`. Ademas, el router, el header y varias vistas pueden ejecutar `loadAuthUser()` durante la misma navegacion, generando consultas duplicadas y esperas innecesarias.

### Objetivo

Hacer que la navegacion protegida tenga una salida clara y una sola carga coordinada del usuario actual.

### Alcance sugerido para Codex

1. Registrar una ruta catch-all que muestre `NotFoundView.vue` sin romper las reglas de autenticacion.
2. Agregar cache de sesion e in-flight promise a `authStore.loadAuthUser()`, con opcion explicita de refresco.
3. Centralizar la carga inicial en el guard o store y retirar llamadas redundantes de componentes cuando no sean necesarias.
4. Mantener limpieza de sesion y cache al cerrar sesion o recibir un fallo de autenticacion.
5. Verificar que usuarios normales sigan sin acceder a rutas administrativas.

### Criterios de aceptacion

- [ ] Una URL desconocida muestra una vista 404 con accion para volver.
- [ ] Navegar entre vistas protegidas no dispara solicitudes duplicadas a `/api/me`.
- [ ] Dos consumidores simultaneos comparten la misma promesa de carga.
- [ ] Cerrar sesion invalida usuario y cache.
- [ ] No se introducen bucles de redireccion.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/router/index.js`
- `frontend/src/stores/auth.js`
- `frontend/src/components/layout/HeaderBar.vue`
- `frontend/src/views/NotFoundView.vue`

## BACK-024 - Hacer reproducible la instalacion frontend y retirar duplicados muertos

Prioridad: P2
Labels: `frontend`, `dependencies`, `cleanup`, `build`, `codex-ready`, `mvp1`
Estado sugerido: Ready for Codex

### Contexto

`lucide-vue-next` esta declarado en el `package.json` de la raiz, pero el despliegue compila `./frontend`. El build local pasa porque existen dependencias instaladas en la raiz, lo que puede ocultar un fallo en una instalacion limpia. Tambien existen stores/composables vacios y un servicio de autenticacion duplicado que no forman parte del flujo real.

### Objetivo

Conseguir que `frontend/` sea instalable y compilable de forma independiente, reduciendo archivos que confunden a futuros agentes.

### Alcance sugerido para Codex

1. Declarar `lucide-vue-next` en `frontend/package.json` y actualizar `frontend/package-lock.json`.
2. Verificar desde una instalacion limpia que `npm ci` y `npm run build` funcionan dentro de `frontend/`.
3. Confirmar con busqueda de imports que los archivos vacios o duplicados no se usan antes de eliminarlos.
4. Retirar `frontend/src/stores/index.js` si conserva un router obsoleto no importado.
5. Mantener una sola implementacion de `authService` y corregir imports si fuera necesario.
6. No cambiar comportamiento funcional durante la limpieza.

### Criterios de aceptacion

- [ ] Todas las dependencias usadas por `frontend/src` estan declaradas en `frontend/package.json`.
- [ ] `npm ci` seguido de `npm run build` funciona dentro de `frontend/` sin depender de `node_modules` raiz.
- [ ] No quedan stores, composables, routers o servicios duplicados sin uso.
- [ ] El workflow de Azure sigue compilando `./frontend`.
- [ ] El diff no contiene refactors funcionales ajenos a la tarea.

### Archivos relevantes

- `package.json`
- `frontend/package.json`
- `frontend/package-lock.json`
- `frontend/src/stores/index.js`
- `frontend/src/stores/reservationStore.js`
- `frontend/src/composables/useReservations.js`
- `frontend/src/composables/useAvailability.js`
- `frontend/src/composables/useAuth.js`
- `frontend/src/services/authService.js`
- `frontend/src/auth/authService.js`

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

## AUTH-005 - Corregir retorno posterior al login Microsoft

Prioridad: P0
Labels: `frontend`, `auth`, `entra`, `mvp1`
Estado sugerido: En revision

### Contexto

Despues de autenticarse con Microsoft, MSAL podia devolver al usuario a `/login?...` porque conservaba la pagina desde la que comenzo el flujo. La sesion quedaba activa, pero era necesario quitar manualmente `/login` de la URL para entrar a la aplicacion.

### Criterios de aceptacion

- [ ] El callback de Microsoft termina en la ruta interna solicitada o en `/`.
- [ ] Nunca se conserva `/login`, `/auth/callback` ni `/blocked` como destino posterior al login.
- [ ] Un usuario autenticado que abre `/login` es redirigido automaticamente.
- [ ] El flujo funciona en Azure Static Web Apps con un usuario real de Entra ID.

### Resultado parcial

- Se configuro MSAL con `navigateToLoginRequestUrl: false` para mantener el control en `/auth/callback`.
- Se agrego sanitizacion del destino guardado en `redirectAfterLogin`.
- `LoginView` detecta una sesion valida y abandona automaticamente la pantalla de acceso.
- 2026-07-14: `npm run build` pasa.
- Pendiente: redeploy del frontend y validacion del flujo Microsoft por el usuario.

### Archivos relevantes

- `frontend/src/auth/msalConfig.js`
- `frontend/src/auth/authService.js`
- `frontend/src/views/LoginView.vue`
- `frontend/src/views/AuthCallbackView.vue`

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

Asegurar que la disponibilidad muestre solo los items recibidos para el dia seleccionado.

### Criterios de aceptacion

- [x] Al cambiar fecha se actualiza la grilla correctamente.
- [x] Reservas de otros dias no aparecen en la fecha actual.
- [x] Actividades institucionales de otros dias no aparecen en la fecha actual.
- [x] Se consideran hora de inicio y duracion.
- [x] Se documenta criterio de zona horaria.
- [x] `npm run build` pasa.

### Resultado de implementacion

- `ScheduleGrid.vue` filtra reservas por la fecha seleccionada.
- `ReservationBlock.vue`, `ResourceTimeline.vue` y `ReservationForm.vue` usan helpers comunes para representar fecha/hora.
- Las horas de reserva se interpretan actualmente como horario local de agenda; `RES-009` debe reemplazar esta convencion implicita por un contrato temporal explicito.
- La vista muestra el bloque reservado despues de crear una reserva y mantiene la validacion de solapamiento desde la base de datos.

### Archivos relevantes

- `frontend/src/components/availability/ScheduleGrid.vue`
- `frontend/src/components/availability/ReservationBlock.vue`
- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/utils/businessTime.js`
- `frontend/src/utils/reservationRules.js`

## RES-004 - Integrar bloqueos y actividades programadas en calendario

Prioridad: P1
Labels: `frontend`, `backend`, `disponibilidad`, `database`, `codex-ready`, `mvp3`
Estado sugerido: Partially Done

### Contexto

La base incluye `availability_blocks`, `scheduled_activities` y la vista `vw_resource_calendar`. La disponibilidad ya combina reservas con actividades institucionales activas; todavia falta incorporar bloqueos y filtrar el contrato por rango.

### Objetivo

Mostrar en disponibilidad reservas, bloqueos y actividades institucionales.

### Criterios de aceptacion

- [ ] Backend expone datos de calendario por rango de fecha.
- [ ] Frontend distingue visualmente reserva, bloqueo y actividad programada; reserva y actividad ya estan resueltas.
- [ ] No se permite seleccionar horarios bloqueados; las actividades institucionales ya ocupan su rango.
- [x] Se muestra detalle al hacer clic en reservas y actividades institucionales.
- [x] Las reservas y actividades programadas provienen de Azure SQL.

### Resultado parcial

- `GetAvailabilityItems` combina reservas con `scheduled_activities` activas y genera una clave estable por tipo.
- Para usuarios normales, las actividades ocultan creador y titulo interno bajo `Actividad institucional`.
- Frontend diferencia `Programacion institucional`, muestra detalle y evita cancelarla como reserva.
- Pendiente: `availability_blocks`, filtros backend por fecha/rango y pruebas automatizadas.
- 2026-07-14: `go test ./...` y `npm run build` pasaron sin pruebas backend descubiertas. Evidencia superada el 2026-07-20: backend ya ejecuta casos reales.

### Archivos relevantes

- `backend/internal/models/availability_item.go`
- `backend/internal/repositories/scheduled_activities_repository.go`
- `backend/internal/services/reservations_service.go`
- `backend/internal/handlers/reservations_handlers.go`
- `frontend/src/components/availability/ReservationBlock.vue`
- `frontend/src/components/forms/ReservationForm.vue`

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
- [x] Bloquea fechas y horarios pasados en frontend y backend.
- [x] El formulario de creacion no solicita un conteo manual de participantes; las reservas grupales gestionan participantes mediante cuentas autenticadas y codigo de invitacion (`RES-008`).
- [x] Deshabilita boton mientras se crea la reserva.
- [x] Muestra error devuelto por backend sin cerrar modal.

### Resultado de implementacion

- `ReservationForm.vue` valida recurso, fecha, hora y duracion antes de enviar.
- Los campos invalidos muestran mensajes visibles y estado visual de error.
- El boton de confirmacion queda bloqueado mientras se crea la reserva o no hay recursos cargados.
- Los errores de backend se mantienen visibles dentro del modal sin cerrarlo.
- `CalendarMini.vue` resalta el dia seleccionado y diferencia el dia actual.
- `AvailabilitySection.vue` bloquea seleccion de horarios pasados antes de abrir el formulario.
- `ReservationForm.vue` evita confirmar reservas con fecha u hora pasada aunque el usuario edite manualmente.
- `reservations_service.go` rechaza reservas en el pasado desde backend.
- 2026-07-08: `npm run build` pasa.
- 2026-07-08: `go test ./...` paso; al 2026-07-14 aun no existian archivos de prueba. Evidencia superada el 2026-07-20: backend ya ejecuta casos reales.
- La politica completa de zona horaria, duraciones permitidas y jornada se separa en `RES-009` y `RES-011`.

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
Estado sugerido: Done

### Actualizacion 2026-08-20

El flujo frontend de cancelacion fue unificado.

`ReservationForm.vue` controla la presentacion y la confirmacion destructiva inline, mientras que la vista padre/store conserva la responsabilidad de ejecutar la mutacion.

Comportamiento vigente:

- no se usa `window.confirm`;
- no se abre un modal sobre otro modal;
- el usuario puede volver sin cancelar;
- la cancelacion requiere confirmacion explicita;
- la vista padre no vuelve a solicitar confirmacion;
- el detalle compartido se reutiliza en los principales accesos.

Las restricciones de estados y autorizacion del backend se documentan y prueban en los items de integridad correspondientes, por lo que no mantienen `RES-007` abierto.


### Contexto

Existe `PATCH /api/reservations/cancel`; ya usa usuario autenticado en backend, pero falta conectar un flujo visual completo desde la interfaz.

### Objetivo

Permitir cancelar reservas desde la interfaz usando el usuario autenticado.

### Criterios de aceptacion

- [x] No se usa `requestedByUserId` fijo.
- [x] El backend determina usuario desde token o valida contra usuario autenticado.
- [x] La UI pide confirmacion fuerte antes de cancelar.
- [x] La reserva cambia a estado `CANCELLED`.
- [x] El resultado queda reflejado en el estado visible y los errores permanecen dentro del contexto.
- [x] La lista se actualiza sin recargar toda la app.

### Resultado parcial

- Los bloques de reserva en disponibilidad son seleccionables.
- Al seleccionar una reserva se reutiliza `ReservationForm.vue` en modo `detail`.
- Admin puede cancelar cualquier reserva; usuario normal solo ve accion de cancelacion para reservas propias.
- El modal muestra errores del backend si la cancelacion falla.
- Al cancelar, la grilla se refresca y el bloque desaparece de la disponibilidad activa.
- La cancelacion usa confirmacion destructiva inline dentro del mismo modal, sin `window.confirm` ni modal sobre modal.

## RES-008 - Persistir participantes de reserva y validar capacidad

Prioridad: P0
Labels: `frontend`, `backend`, `database`, `reservas`, `ux`, `needs-architecture`, `mvp2`
Estado sugerido: Partially Done

### Contexto

El usuario confirmo el 2026-07-20 que Cancha 1, 2 y 3 corresponden formalmente a multicancha 1, 2 y 3. Cada solicitud requiere al menos 10 usuarios con cuenta, incluido el solicitante. La solicitud `PENDING` bloquea el horario y consume la oportunidad semanal. Las confirmaciones pueden registrarse o retirarse hasta exactamente una hora antes inclusive, valor configurable. Si vence bajo el minimo, se cancela, libera el horario y deja de consumir la oportunidad.

El flujo grupal ya registra participantes persistidos y no confirma todas las reservas inmediatamente. Una reserva de recurso grupal nace en `PENDING`, registra al solicitante como primer participante y pasa a `CONFIRMED` cuando alcanza el minimo.

Durante MVP 2 la regla fue refinada deliberadamente: una reserva que ya alcanzo el minimo conserva `reservation.status = CONFIRMED` si posteriormente pierde participantes y pasa a `groupCondition = AT_RISK`. Esta separacion entre ciclo de vida y condicion grupal reemplaza la regla anterior que hacia regresar a `PENDING`.

La tarea permanece parcial por vencimiento automatico, liberacion asociada de la oportunidad semanal, integracion completa de recuperacion y notificaciones.

### Objetivo

Registrar confirmaciones de participantes unicos, mantener la solicitud grupal en `PENDING` bajo el minimo y confirmarla automaticamente al alcanzar 10 confirmaciones validas, sin permitir que el cliente controle el estado o el conteo.

### Alcance funcional aprobado

1. Aplicar la regla a multicancha 1, 2 y 3, identificadas como Cancha 1, 2 y 3 en el inventario.
2. Registrar cada participante mediante su cuenta, sin duplicados, incluyendo al solicitante dentro del conteo.
3. Mostrar al solicitante el estado y avance hacia el minimo.
4. Mantener `PENDING` con menos de 10 confirmaciones vigentes.
5. Al registrar la decima confirmacion, volver a validar las demas reglas y cambiar una sola vez a `CONFIRMED`.
6. Mantener el servidor como autoridad del estado y del conteo.
7. No exigir este flujo a `OPEN_USE`.
8. Respetar la capacidad maxima cuando el recurso la defina.
9. Permitir retirar una confirmacion hasta el limite configurable. Si una reserva nunca alcanzo el minimo permanece `PENDING_MINIMUM`; si ya estaba confirmada y baja del minimo conserva `CONFIRMED` con condicion `AT_RISK`.
10. Con la configuracion inicial, aceptar cambios hasta exactamente una hora antes inclusive y rechazarlos despues.
11. Mientras este `PENDING`, bloquear el horario para solicitudes incompatibles.
12. Al llegar al limite bajo el minimo, cambiar a `CANCELLED`, liberar el horario y liberar la oportunidad semanal.
13. Restringir a administradores los cambios de recursos sujetos a la regla y del plazo previo.

### Criterios de aceptacion

- [x] Una solicitud grupal nueva comienza en `PENDING`.
- [ ] Un mismo participante no puede aumentar dos veces el conteo vigente.
- [x] Con 9 confirmaciones la solicitud permanece `PENDING`.
- [x] La decima confirmacion valida cambia la solicitud a `CONFIRMED` si las demas reglas siguen cumpliendose.
- [ ] `OPEN_USE` no solicita confirmaciones de participantes.
- [ ] El cliente no puede forzar `CONFIRMED` ni declarar un conteo arbitrario.
- [ ] Se rechaza una confirmacion que supere la capacidad del recurso cuando corresponda.
- [ ] El solicitante puede ver el avance y los errores recuperables.
- [x] Si una retirada valida reduce el conteo de 10 a 9 despues de haber confirmado, la reserva conserva `CONFIRMED` y cambia a `AT_RISK`.
- [ ] Confirmar o retirar despues del limite configurado se rechaza sin cambiar el conteo.
- [ ] El valor inicial del limite es una hora antes del inicio y un cambio autorizado se refleja en las solicitudes posteriores aplicables.
- [x] El solicitante cuenta una vez entre los 10 y las operaciones de participacion utilizan al usuario autenticado.
- [ ] Mientras esta `PENDING`, el horario aparece ocupado para solicitudes incompatibles.
- [ ] Exactamente una hora antes se admite el ultimo cambio; despues se rechaza.
- [ ] Al vencer bajo el minimo cambia a `CANCELLED`, libera horario y oportunidad semanal.
- [ ] Un usuario normal no puede cambiar el plazo ni los recursos sujetos a confirmacion.
- [ ] `go test ./...` pasa.
- [ ] `npm test` pasa.
- [ ] `npm run build` pasa.

### Decision de producto refinada 2026-08-20

Los MVP se utilizan tambien para refinar requisitos a medida que el dominio se valida.

Para reservas grupales queda vigente la separacion:

- `status`: ciclo de vida de la reserva;
- `groupCondition`: salud operacional del grupo.

Una reserva confirmada que cae bajo el minimo no vuelve a `PENDING`; pasa a `CONFIRMED + AT_RISK`.

Esto permite que MVP posteriores implementen notificaciones por:

- ingreso a riesgo;
- recuperacion del minimo;
- proximidad del vencimiento;
- cancelacion final bajo el minimo.

Las reglas administrativas de politica siguen aplicandose prospectivamente a solicitudes nuevas salvo mecanismos excepcionales explicitamente definidos.

## RES-009 - Definir zona horaria de negocio para reservas

Prioridad: P0
Labels: `backend`, `frontend`, `database`, `reservas`, `timezone`, `bug`, `codex-ready`, `mvp1`
Estado actual: En revision

### Contexto

Azure SQL usa columnas `DATETIME2` sin zona. El frontend interpreta los valores serializados con sufijo `Z` como hora de muro local, mientras el backend parsea fechas sin offset con `time.Local` y compara contra `time.Now()`. En un contenedor Azure configurado en UTC, las reglas de pasado, cancelacion y disponibilidad pueden adelantarse o atrasarse respecto de Chile.

### Objetivo

Aplicar una unica zona horaria de negocio, explicita y comprobable, en creacion, serializacion, clasificacion temporal y cancelacion.

### Alcance sugerido para Codex

1. Definir `APP_TIMEZONE` con valor esperado `America/Santiago` y documentar fallback seguro.
2. Crear un helper backend para cargar la ubicacion y obtener `now` de negocio; evitar usos directos de `time.Local` en reservas.
3. Parsear `startTime` segun un contrato unico. Preferencia: aceptar ISO 8601 con offset y normalizar antes de persistir; si se mantiene fecha sin offset, interpretarla explicitamente en `APP_TIMEZONE`.
4. Definir si `DATETIME2` guarda UTC normalizado o hora local institucional y documentarlo en `docs/03-base-de-datos.md`.
5. Alinear `parseReservationDateTime` con ese contrato, sin ignorar silenciosamente un offset real.
6. Usar el mismo reloj para pasado/en curso/futuro y para cancelacion.
7. Agregar pruebas alrededor de medianoche y del cambio de horario de Chile.

### Criterios de aceptacion

- [ ] Local y Azure clasifican la misma reserva en la misma categoria temporal.
- [ ] Una reserva de Chile no cambia tres o cuatro horas al viajar API -> frontend.
- [x] Cancelacion y rechazo de fechas pasadas usan la zona configurada.
- [x] El contrato de fecha/hora queda documentado con un ejemplo de request y response.
- [x] Una `APP_TIMEZONE` invalida impide iniciar o produce un error de configuracion claro.
- [x] Existen pruebas con un reloj inyectable; no dependen de la hora real del equipo.
- [x] `go test ./...`, `npm test` y `npm run build` pasan localmente.

### Resultado de implementacion

- `DATETIME2` conserva la hora institucional de muro; no se requieren columnas nuevas ni transformar reservas existentes.
- Backend y frontend usan `America/Santiago` de forma explicita, incluidos cambios de horario de invierno y verano.
- Requests con offset se convierten a Santiago; requests sin offset se interpretan directamente en la zona institucional.
- Se agregaron pruebas de cambio estacional, cruce de medianoche, configuracion invalida y rechazo de reservas pasadas con reloj controlado.
- Falta configurar las variables en Azure, desplegar ambos componentes y comparar una reserva de hora conocida en local y online antes de cerrar la tarea.

### Archivos relevantes

- `backend/internal/handlers/reservations_handlers.go`
- `backend/internal/services/reservations_service.go`
- `backend/internal/models/reservation.go`
- `frontend/src/utils/reservationTime.js`
- `backend/.env.example`
- `docs/03-base-de-datos.md`
- `docs/10-guia-redeploy.md`

## RES-010 - Hacer que el servidor controle estados de reserva

Prioridad: P0
Labels: `backend`, `reservas`, `security`, `integrity`, `codex-ready`, `mvp1`
Estado actual: En revision

### Contexto

`CreateReservationRequest` acepta `status` y el handler lo copia al modelo. Un cliente modificado puede intentar crear reservas `PENDING`, `CANCELLED`, `REJECTED` o `EXPIRED`; algunos de esos estados no participan en las mismas reglas de conflicto. La cancelacion, por su parte, solo rechaza `CANCELLED` y no limita la transicion a estados activos.

### Objetivo

Tratar el estado y sus transiciones como reglas exclusivas del servidor.

### Alcance sugerido para Codex

1. Retirar `status` del contrato publico `CreateReservationRequest`.
2. Forzar en el servicio el estado inicial definido para MVP 1, actualmente `CONFIRMED`.
3. Rechazar campos desconocidos o documentar que `status` recibido se ignora; preferir rechazo para detectar clientes desactualizados.
4. Definir una funcion de transicion que permita cancelar solo estados activos (`CONFIRMED` y `PENDING` si este ultimo se usa).
5. Mantener idempotencia o error estable para una reserva ya cancelada.
6. Confirmar que triggers/consultas de conflicto no puedan eludirse mediante estados enviados por cliente.
7. Agregar pruebas de request manipulado y transiciones invalidas.

### Criterios de aceptacion

- [x] El cliente no puede decidir el estado inicial de una reserva.
- [x] Toda reserva creada por el endpoint publico fuerza el estado inicial esperado en el servicio.
- [x] No se puede cancelar una reserva `REJECTED`, `EXPIRED` o ya `CANCELLED`.
- [x] Propietario/admin y ventana temporal siguen validandose.
- [x] Un intento de enviar `status` se rechaza antes de acceder a la base de datos.
- [x] `go test ./...` pasa localmente con casos positivos y negativos.

### Resultado de implementacion

- `CreateReservationRequest` ya no expone `status` y el endpoint rechaza campos JSON desconocidos.
- El servicio fuerza `CONFIRMED` aunque otro llamador interno entregue un estado distinto.
- Solo `CONFIRMED` y `PENDING` pueden transicionar a `CANCELLED`.
- La condicion se repite en el `UPDATE` SQL para proteger el cambio frente a concurrencia.
- Se agregaron pruebas del endpoint manipulado, estado inicial y transiciones validas e invalidas.
- Falta desplegar y ejecutar la prueba manual del checklist antes de cerrar la tarea.

### Archivos relevantes

- `backend/internal/models/reservation.go`
- `backend/internal/handlers/reservations_handlers.go`
- `backend/internal/services/reservations_service.go`
- `backend/internal/repositories/reservations_repository.go`

## RES-011 - Validar duracion y horario operativo en backend

Prioridad: P0
Labels: `backend`, `frontend`, `reservas`, `validation`, `codex-ready`, `mvp1`
Estado actual: En revision

### Contexto

El formulario ofrece duraciones acotadas y la disponibilidad muestra una jornada aproximada de 08:00 a 22:00, pero la API solo valida que `durationMinutes` sea mayor que cero. Un cliente modificado puede crear reservas excesivamente largas, fuera de horario o cuyo termino exceda el cierre.

El usuario aprobo el 2026-07-20 el catalogo de 30, 60, 90, 120, 150 y 180 minutos para todos los recursos. La implementacion local coincide; falta su verificacion integrada/online.

### Objetivo

Definir y aplicar en backend las reglas institucionales minimas de duracion e intervalo horario.

### Alcance sugerido para Codex

1. Aplicar la configuracion MVP 1: apertura `08:00`, cierre `22:00`, paso de 15 minutos y duraciones aprobadas `30, 60, 90, 120, 150, 180`.
2. Validar que inicio y termino completo queden dentro de la jornada de la fecha seleccionada.
3. Rechazar duraciones fuera del catalogo, valores cero/negativos y overflow de fecha.
4. Validar que la hora de inicio respete el paso institucional.
5. Compartir las mismas constantes con el frontend mediante configuracion o contrato estable, evitando reglas divergentes.
6. Devolver mensajes accionables sin exponer detalles internos.
7. Agregar pruebas de limites: apertura, ultimo bloque valido, cierre excedido y duracion manipulada.

### Criterios de aceptacion

- [x] Backend acepta el primer y ultimo rango validos de la jornada.
- [x] Backend rechaza inicio antes de apertura y termino despues del cierre.
- [x] Backend rechaza duracion no permitida aunque el frontend sea omitido.
- [x] Backend rechaza inicios fuera del paso institucional.
- [x] Frontend solo ofrece combinaciones que el backend acepta.
- [x] Los mensajes indican como corregir fecha, hora o duracion.
- [x] `go test ./...`, `npm test` y `npm run build` pasan localmente.

### Resultado de implementacion

- El backend centraliza apertura 08:00, cierre 22:00, paso de 15 minutos y duraciones `30, 60, 90, 120, 150, 180`.
- La API valida inicio, segundos, paso, duracion y termino completo antes de consultar la base de datos.
- El formulario usa el mismo catalogo, elimina 45 minutos y limita la ultima hora segun la duracion seleccionada.
- La grilla y el mapa de uso comparten las constantes de jornada y segmento del formulario.
- Se agregaron pruebas unitarias y de servicio para apertura, ultimo bloque, cierre excedido, paso y duracion manipulada.
- No se requieren columnas nuevas ni migracion; falta probar los limites contra el despliegue antes de cerrar la tarea.

### Archivos relevantes

- `backend/internal/services/reservations_service.go`
- `backend/internal/handlers/reservations_handlers.go`
- `frontend/src/components/forms/DateTimePicker.vue`
- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/availability/ResourceTimeline.vue`

---

## RES-012 - Aplicar restriccion semanal de reservas particulares

Prioridad: P0
Labels: `reservas`, `reglas-negocio`, `backend`, `frontend`, `mvp2`, `needs-architecture`
Estado sugerido: Implemented; VERIFICADO LOCALMENTE, integracion SQL/Azure pendiente

### Contexto

El usuario confirmo el 2026-07-20 que la regla vigente usa un periodo configurable de siete dias. En `America/Santiago`, un martes se pueden elegir fechas desde ese martes hasta el lunes siguiente. Si la solicitud se crea ese martes, el usuario puede volver a crear otra desde el martes siguiente, aunque la primera reserva sea para el miercoles. El sistema actual no aplica ninguna de estas restricciones.

### Objetivo

Limitar las fechas reservables al periodo institucional configurado y rechazar nuevas solicitudes antes del siguiente periodo contado desde la fecha local de creacion de la solicitud anterior, comunicando la proxima fecha permitida.

### Criterios de aceptacion

- [ ] La regla usa exclusivamente la identidad autenticada.
- [ ] Con el valor vigente de siete dias, un martes admite fechas desde ese martes hasta el lunes siguiente inclusive y rechaza el martes posterior.
- [ ] Una solicitud relevante creada un martes impide crear otra hasta el lunes siguiente inclusive, con independencia de la fecha reservada.
- [ ] Una solicitud `PENDING` consume la oportunidad desde su creacion.
- [x] Al pasar a `CANCELLED`, deja de consumirla y la proxima fecha permitida se recalcula con las demas solicitudes vigentes.
- [ ] Desde el martes siguiente, la nueva solicitud puede continuar si cumple las demas reglas.
- [ ] El rechazo comunica la proxima fecha permitida en `America/Santiago`.
- [ ] Solicitudes de otro usuario no afectan el resultado.
- [ ] El servidor mantiene la regla aunque el cliente sea manipulado.
- [ ] Un cambio autorizado del numero de dias modifica tanto la ventana reservable como la siguiente fecha permitida sin reinterpretar reservas historicas.
- [ ] Existen pruebas del limite inclusivo por fecha, cambio de fecha y horario de verano.

### Decision de producto cerrada

`PENDING` y `CONFIRMED` consumen la oportunidad; `CANCELLED` no la consume. Una solicitud rechazada que nunca fue creada tampoco la consume.

### Evidencia QA del 2026-07-21

- [x] La reserva obtiene y persiste la politica vigente dentro de la transaccion que adquiere los bloqueos; ventana y frecuencia se evaluan con el snapshot referenciado.
- [x] Pruebas locales de servicio y contrato SQL cubren version historica, cancelacion y seleccion de politica.
- [ ] Falta ejecutar el esquema actualizado contra SQL Server/Azure SQL real y validar concurrencia con transacciones simultaneas.

- [x] Dos solicitudes concurrentes del mismo usuario fueron ejecutadas contra Azure SQL: una se creo y la otra fue rechazada por la proteccion de frecuencia, conservando una sola solicitud activa.
- [x] La reserva activa fue cancelada desde el frontend y una nueva solicitud del mismo usuario pudo crearse, verificando de punta a punta que `CANCELLED` libera la oportunidad.
- [ ] Falta verificar que una solicitud historica conserve `request_frequency_days` de su `policy_id` despues de publicar una politica nueva con una frecuencia diferente.

---

# Hito 4 - Pantallas pendientes

## UI-001 - Implementar vista Recursos

Prioridad: P1
Labels: `frontend`, `recursos`, `feature`
Estado sugerido: Done

### Contexto

`ResourcesView.vue` ya lista recursos reales desde `/api/resources`, permite filtros y muestra imagenes cuando estan configuradas.

### Objetivo

Mostrar catalogo de recursos deportivos con datos reales.

### Criterios de aceptacion

- [x] Lista recursos desde `/api/resources`.
- [x] Muestra nombre, tipo, modo de reserva, capacidad y estado.
- [x] Permite filtrar por tipo o sede si los datos estan disponibles.
- [x] Muestra imagen del recurso cuando existe y fallback cuando no.
- [x] Tiene estados de carga, error y vacio.
- [x] Mantiene estilo visual actual.

### Resultado parcial

- `ResourcesView.vue` ya no muestra `Proximamente...`.
- `GET /api/resources` ahora incluye `capacity` desde Azure SQL.
- `ResourcesView.vue` permite buscar por nombre/tipo y filtrar por tipo, modo de reserva y estado.
- `ResourcesView.vue` muestra `imageUrl` cuando existe y una inicial como fallback.

## UI-002 - Implementar detalle de reserva

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Done para MVP 1

### Contexto

`ReservationDetailView.vue` ya muestra datos reales de una reserva. Usuarios normales consultan solo sus reservas; administradores pueden abrir reservas del listado global.

### Objetivo

Mostrar informacion detallada de una reserva.

### Criterios de aceptacion

- [x] Muestra recurso, actividad, usuario, fecha, hora, duracion y estado.
- [x] No promete participantes mientras el dato no exista; su incorporacion queda en `RES-008`.
- [x] Permite volver a Mis Reservas o Disponibilidad.
- [x] Maneja reserva no encontrada.

### Resultado parcial

- Existe ruta `/reservations/:id`.
- La vista carga reservas reales desde `/api/reservations/mine` para usuario normal y desde `/api/reservations` para administrador.
- Permite cancelar desde el detalle si la reserva corresponde.
- Participantes queda pendiente porque el campo no esta persistido en el modelo actual.
- 2026-07-14: se verifico en navegador el flujo `Historial -> Detalle -> Volver a Historial`.
- La optimizacion para cargar una sola reserva por ID queda separada en `API-006` para MVP 2.

## UI-003 - Implementar Historial

Prioridad: P2
Labels: `frontend`, `reservas`, `feature`
Estado sugerido: Done

### Contexto

El contexto `?tab=history` de `ReservationsView.vue` lista reservas historicas reales con comportamiento por rol.

### Objetivo

Mostrar historial de reservas pasadas, canceladas o rechazadas.

### Criterios de aceptacion

- [x] Lista reservas historicas del usuario.
- [x] Si el usuario es administrador, lista todo el historial del sistema.
- [x] Permite filtrar por estado.
- [x] Permite filtrar por fecha.
- [x] Tiene estados de carga, error y vacio.

### Resultado parcial

- El tab historico de `ReservationsView.vue` muestra reservas pasadas o canceladas desde la fuente correspondiente al rol.
- El tab historico permite filtrar por estado y rango de fecha.

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

## UI-006 - Implementar Talleres deportivos

Prioridad: P2
Labels: `frontend`, `backend`, `database`, `talleres`, `feature`
Estado sugerido: Done

### Contexto

El sistema ahora incluye una vista de talleres deportivos, endpoints protegidos y tablas dedicadas para inscripciones.

### Objetivo

Permitir que usuarios autenticados consulten talleres activos y se inscriban cuando existan cupos disponibles.

### Criterios de aceptacion

- [x] Existe vista `/workshops` accesible desde la navegacion.
- [x] Lista talleres activos desde `GET /api/workshops`.
- [x] Permite buscar por taller, dia o lugar.
- [x] Muestra cupos usados y capacidad.
- [x] Muestra si el usuario ya esta inscrito.
- [x] Permite inscripcion mediante `POST /api/workshops/:id/enroll`.
- [x] Rechaza usuarios normales sin RUT.
- [x] Rechaza cupos completos e inscripcion duplicada.

### Resultado de implementacion

- Se agregaron `workshops` y `workshop_enrollments` en Azure SQL.
- Se agregaron modelo, repositorio, servicio y handler de talleres en backend.
- Se agrego store Pinia y servicio frontend para talleres.
- Se agrego `WorkshopsView.vue` con busqueda, estados de carga/error/exito y accion de inscripcion.
- La inscripcion usa transaccion serializable para validar cupos antes de insertar.

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
Estado sugerido: Partially Done

### Objetivo

Permitir crear, editar, activar o desactivar recursos deportivos, usando los ocho recursos actuales como inventario oficial inicial.

### Criterios de aceptacion

- [ ] CRUD basico de recursos.
- [x] Permite actualizar la imagen del recurso desde una ruta administrativa.
- [ ] Recurso pertenece a una sede.
- [ ] Valida `reservation_mode`.
- [ ] Solo un administrador puede definir o modificar la politica que determina si el recurso requiere confirmacion grupal.
- [ ] No permite eliminar recursos con reservas historicas sin criterio definido.
- [ ] Refresca disponibilidad luego de cambios.

### Resultado parcial de implementacion

- Se agrego `resources.image_url` en `database/schema.sql` y datos iniciales en `database/seed.sql`.
- `GET /api/resources` y la consulta por ID devuelven `imageUrl`.
- Se agrego `PATCH /api/resources/:id/image`, protegido con `RequireAdmin`.
- El backend valida ID, largo maximo de URL y acepta solo `http://`, `https://` o rutas locales iniciadas en `/`.
- `ResourcesView.vue` permite a administradores editar o limpiar la imagen del recurso y actualiza el store sin recargar toda la pagina.
- Pendiente: CRUD completo del inventario oficial, sede, modo de reserva, politica de confirmacion grupal, activacion/desactivacion y criterio de eliminacion.

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

Prioridad: P0
Labels: `frontend`, `backend`, `admin`, `disponibilidad`
Estado sugerido: Backlog

### Objetivo

Permitir registrar clases, talleres, eventos, campeonatos o entrenamientos institucionales y resolver sus conflictos segun la prioridad aprobada.

### Criterios de aceptacion

- [ ] Admin crea actividad programada sobre un recurso.
- [ ] El sistema detecta y muestra todas las ocupaciones solapadas antes de aplicar efectos.
- [ ] Si existe una reserva particular, se cancela automaticamente al confirmar la actividad institucional.
- [ ] Si existen dos actividades en conflicto, el administrador puede cancelar cualquiera de ellas o mantener ambas.
- [ ] Mantener ambas requiere una decision administrativa explicita y no se rechaza solo por compartir recurso y horario.
- [ ] Actividades aparecen en calendario.
- [ ] Soporta descripcion y tipo.
- [ ] La decision y las cancelaciones quedan trazables.
- [ ] El administrador ve un resumen de reservas y actividades afectadas.
- [ ] El usuario cuya reserva particular se cancela automaticamente recibe una notificacion del cambio sin exponer datos de otras personas.
- [ ] Define comportamiento futuro de recurrencia.

## ADMIN-006 - Configurar politicas institucionales de reserva

Prioridad: P1
Labels: `admin`, `reservas`, `reglas-negocio`, `architecture-approved`, `mvp3`
Estado sugerido: Implemented parcial; VERIFICADO LOCALMENTE

### Objetivo

Permitir que solo usuarios con rol administrador modifiquen el periodo de reserva, el plazo previo de confirmacion y los recursos sujetos a confirmacion grupal.

### Criterios de aceptacion

- [x] Un administrador puede publicar una nueva version del periodo cuyo valor inicial es siete dias.
- [x] Un administrador puede publicar una nueva version del plazo previo cuyo valor inicial es una hora.
- [x] Un administrador puede versionar recursos permitidos, jornada, intervalo y duraciones permitidas.
- [ ] La politica clasifica que recursos requieren confirmacion grupal; participantes, minimo y transiciones asociadas siguen pendientes.
- [x] Un usuario normal no puede consultar el historial ni modificar estas politicas; solo recibe el DTO operativo minimo vigente.
- [x] El historial administrativo conserva valores, autoria y vigencias; la interfaz para mostrarlos sigue pendiente.
- [x] El cambio normal crea una version inmutable con vigencia inmediata y se aplica a solicitudes creadas posteriormente.
- [x] Cada solicitud conserva la version vigente al crearse.
- [ ] Un administrador puede previsualizar una correccion excepcional sobre solicitudes futuras `PENDING` o `CONFIRMED` seleccionadas.
- [ ] La correccion exige motivo, confirmacion explicita, auditoria e idempotencia por lote.
- [ ] La aplicacion revalida todas las solicitudes y es atomica: si una es incompatible, no cambia ninguna.
- [ ] La correccion no edita versiones historicas ni cancela solicitudes implicitamente.
- [ ] Una reversion se registra como una nueva correccion hacia la version anterior.

### Arquitectura aprobada

- Persistir versiones en `reservation_policies` y sus recursos permitidos en `reservation_policy_resources`; la clasificacion de confirmacion grupal requiere implementacion posterior.
- Asociar cada reserva mediante `reservations.policy_id`.
- Registrar excepciones en `reservation_policy_corrections` con version anterior/nueva, administrador, motivo, fecha UTC y lote.
- Separar `preview` y `apply` en rutas administrativas protegidas.
- Mantener `ADMIN-005` fuera de esta entrega y disenarlo posteriormente.

### Resultado implementado y evidencia

- Contratos: `GET /api/reservation-policy/current`, `GET /api/admin/reservation-policies` y `POST /api/admin/reservation-policies`.
- La publicacion exige `Idempotency-Key`: `201` nuevo, `200` replay identico y `409` replay divergente.
- El esquema protege version vigente unica, idempotencia, snapshot completo e inmutabilidad; el seed completa y marca una sola vez el bootstrap tecnico de recursos permitidos despues de cargar el catalogo.
- QA acepto/verifico localmente tras cuatro rondas. No se ejecuto el esquema actualizado en SQL Server/Azure SQL real; concurrencia fue validada estaticamente y `go vet` no se ejecuto en la ronda final por cuota.
- La vista previa/aplicacion de correcciones excepcionales y su auditoria permanecen pendientes y fuera de este incremento. La retencion institucional definitiva tambien sigue pendiente; las politicas referenciadas por reservas no pueden eliminarse.

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

## REP-003 - Corregir semantica y precision de indicadores

Prioridad: P1
Labels: `frontend`, `backend`, `reportes`, `bug`, `codex-ready`, `mvp3`
Estado sugerido: Ready for Codex

### Contexto

La vista actual considera activa toda reserva cuyo estado no sea `CANCELLED`. Esto incluye reservas finalizadas, rechazadas o expiradas. Tambien redondea minutos con `Math.round`, por lo que 90 minutos se presentan como 2 horas y alteran los totales por recurso.

### Objetivo

Hacer que cada indicador tenga una definicion funcional explicita y represente los datos sin redondeos engañosos.

### Alcance sugerido para Codex

1. Definir `reservas activas` como estados activos y rango temporal vigente/futuro usando el helper temporal compartido.
2. Separar, si aporta valor, reservas confirmadas historicas de reservas actualmente accionables.
3. Calcular minutos como unidad base y formatear horas con una precision estable, por ejemplo `1 h 30 min`.
4. Aplicar la misma poblacion a KPI, uso por recurso y horas punta, o nombrar cada metrica si usa otra poblacion.
5. Agregar subtitulo o tooltip corto que explique periodo y criterio sin texto tecnico.
6. Cubrir estados `CONFIRMED`, `CANCELLED`, `REJECTED`, `EXPIRED`, pasada, en curso y futura.

### Criterios de aceptacion

- [ ] Una reserva finalizada no aumenta `Reservas activas`.
- [ ] Reservas canceladas, rechazadas y expiradas no cuentan como uso activo.
- [ ] Noventa minutos no se muestran como dos horas completas.
- [ ] Todos los widgets que dicen usar reservas activas comparten el mismo conjunto.
- [ ] El periodo y criterio de cada KPI son comprensibles para administracion.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/views/ReportsView.vue`
- `frontend/src/utils/reservationTime.js`

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
Labels: `frontend`, `backend`, `notificaciones`, `codex-ready`, `mvp2`
Estado sugerido: Ready for Codex

### Objetivo

Mostrar notificaciones reales al usuario desde Azure SQL.

### Alcance sugerido para Codex

1. Agregar endpoint autenticado para marcar una notificacion propia como leida.
2. Rechazar IDs inexistentes o de otro usuario sin revelar informacion ajena.
3. Diferenciar leidas/no leidas en store y campana sin depender solo del color.
4. Marcar como leida al activar una notificacion o mediante una accion explicita consistente.
5. Definir un destino real para `Ver todas`; si no se implementa una vista, mantener ese control fuera de la UI.
6. Actualizar el contador sin recargar toda la aplicacion.

### Criterios de aceptacion

- [x] Endpoint para listar notificaciones del usuario.
- [ ] Endpoint para marcar como leida.
- [x] Campana muestra contador real.
- [ ] UI diferencia leidas/no leidas.
- [x] Maneja estado vacio.
- [ ] `Ver todas` navega a una vista funcional o no se muestra.
- [ ] Usuario no puede marcar notificaciones ajenas.
- [ ] `go test ./...` y `npm run build` pasan.

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
Estado sugerido: Partially Done

### Contexto

La vista de disponibilidad necesita saber que horarios estan ocupados, pero un usuario normal no necesita recibir datos internos de reservas ajenas.

### Objetivo

Exponer un endpoint de disponibilidad por fecha o rango que entregue solo la informacion necesaria para pintar ocupacion.

### Criterios de aceptacion

- [ ] El endpoint acepta fecha o rango de fechas.
- [x] Para usuario normal no expone `userId` de reservas ajenas.
- [x] Devuelve recurso, inicio y duracion para pintar ocupacion.
- [x] Incluye reservas y actividades programadas activas.
- [ ] Incluye bloqueos de disponibilidad cuando esten disponibles.
- [x] La vista de disponibilidad consume este endpoint en lugar de depender de todas las reservas.
- [x] Admin mantiene acceso a detalle administrativo cuando corresponda.

### Resultado parcial de implementacion

- Se agrego `GET /api/availability/reservations`.
- Para usuarios normales, el handler limpia `userId`, nombre, email y RUT antes de responder.
- `GET /api/reservations` quedo protegido con `RequireAdmin`.
- El store frontend separa `availabilityReservations` de las reservas administrativas.
- `GetAvailabilityItems` unifica reservas y `scheduled_activities` con clave estable y sanitizacion para usuario normal.
- Pendiente: agregar filtros por fecha/rango e incorporar `availability_blocks` al contrato.

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

## API-006 - Crear endpoint de detalle de reserva por ID

Prioridad: P2
Labels: `backend`, `frontend`, `reservas`, `performance`, `security`, `codex-ready`, `mvp2`
Estado sugerido: Ready for Codex

### Contexto

`ReservationDetailView.vue` carga la coleccion completa de reservas del usuario o del administrador y luego busca el ID en frontend. El costo y la exposicion de datos crecen con el historial aunque la pantalla solo necesita una reserva.

### Objetivo

Entregar el detalle minimo autorizado de una reserva mediante un endpoint dedicado.

### Alcance sugerido para Codex

1. Agregar `GET /api/reservations/:id` bajo autenticacion.
2. Permitir acceso al propietario y a administradores; responder de forma estable para inexistente o no autorizado.
3. Reutilizar una consulta de repositorio que incluya recurso, actividad y datos necesarios del usuario solo para admin.
4. Evitar exponer email, RUT u otros datos personales al propietario cuando no aporten al detalle.
5. Crear metodo de servicio/store frontend y migrar `ReservationDetailView.vue`.
6. Mantener el parametro `from` solo para decidir el destino de regreso, no para autorizar.

### Criterios de aceptacion

- [ ] Propietario obtiene su reserva por ID.
- [ ] Administrador obtiene cualquier reserva por ID.
- [ ] Usuario normal no obtiene una reserva ajena.
- [ ] La respuesta no incluye datos personales innecesarios.
- [ ] La vista de detalle no descarga la coleccion completa.
- [ ] Acceso desde Mis Reservas e Historial conserva el regreso correcto.
- [ ] `go test ./...` y `npm run build` pasan.

### Archivos relevantes

- `backend/internal/routes/routes.go`
- `backend/internal/handlers/reservations_handlers.go`
- `backend/internal/services/reservations_service.go`
- `backend/internal/repositories/reservations_repository.go`
- `frontend/src/services/reservations.service.js`
- `frontend/src/stores/reservations.js`
- `frontend/src/views/ReservationDetailView.vue`

---

# Hito 8 - Calidad, pruebas y seguridad

## QA-001 - Agregar pruebas backend para reglas de reservas

Prioridad: P0
Labels: `backend`, `testing`, `reservas`, `security`, `codex-ready`, `mvp1`
Estado sugerido: Partially Done

### Objetivo

Crear pruebas deterministas de la capa servicio para cerrar las reglas criticas del MVP 1. `go test ./...` ya ejecuta pruebas reales de reloj de negocio, parsing JSON, reglas de horario y casos iniciales del servicio de reservas; falta ampliar cobertura sobre permisos, conflictos y rutas admin.

### Alcance sugerido para Codex

1. Extraer interfaces pequenas o dependencias inyectables para repositorio y reloj solo donde las pruebas lo necesiten.
2. Usar pruebas table-driven para validaciones puras y transiciones de estado.
3. Mantener una cantidad acotada de pruebas de integracion para SQL si existe ambiente reproducible; no exigir Azure SQL para toda la suite unitaria.
4. Cubrir primero `RES-009`, `RES-010` y `RES-011`.
5. Evitar sleeps y dependencia de `time.Now()` real.

### Casos minimos

- [ ] Crear reserva valida.
- [ ] Rechazar recurso inexistente.
- [ ] Rechazar usuario inexistente.
- [ ] Rechazar usuario normal sin RUT.
- [ ] Rechazar recurso inactivo.
- [ ] Rechazar recurso informativo.
- [ ] Rechazar recurso solo admin para usuario normal.
- [ ] Rechazar conflicto horario.
- [x] Rechazar `status` controlado por cliente o ignorarlo de forma segura segun contrato.
- [x] Crear siempre con estado inicial del servidor.
- [x] Rechazar duracion y horario operativo invalidos.
- [x] Clasificar y cancelar correctamente usando `APP_TIMEZONE` y reloj inyectable.
- [ ] Rechazar usuario bloqueado.
- [ ] Rechazar cruce con bloqueo.
- [ ] Rechazar cruce con actividad programada.
- [ ] Cancelar reserva propia.
- [ ] Cancelar reserva como admin.
- [ ] Rechazar cancelacion sin permisos.
- [ ] Rechazar cancelacion de reserva inexistente.
- [ ] Rechazar cancelacion duplicada.
- [ ] Rechazar cancelacion de estado no activo.
- [ ] Usuario normal no accede a endpoints admin.
- [ ] `go test ./...` ejecuta pruebas reales y reporta paquetes con casos de test.

## QA-002 - Agregar pruebas frontend basicas

Prioridad: P1
Labels: `frontend`, `testing`, `ux`, `accessibility`, `codex-ready`, `mvp1`
Estado sugerido: Partially Done

### Actualizacion 2026-08-20

La base automatizada frontend crecio y fue verificada nuevamente.

Evidencia local 2026-08-20:

- `npm test`: 25 pruebas correctas;
- `npm run build`: correcto;
- `git diff --check`: correcto.

La suite actual cubre utilidades de tiempo de negocio, reglas de reserva, disponibilidad, clasificacion/foco de reservas y configuracion de alcance.

La tarea permanece parcial porque aun falta cobertura directa de componentes Vue, router, autenticacion y flujos end-to-end.


### Objetivo

Agregar pruebas o checklist automatizado para pantallas criticas.

### Alcance sugerido para Codex

1. Agregar Vitest, Vue Test Utils y entorno DOM al `frontend/package.json`.
2. Mantener un script `test` reproducible y definir su ejecucion no interactiva para CI cuando corresponda.
3. Priorizar helpers temporales, store de autenticacion, formulario de reserva y guards del router.
4. Simular servicios HTTP; las pruebas de componentes no deben depender del backend real.
5. Mantener el checklist manual de navegador para responsive y foco como complemento, no sustituto permanente.

### Casos minimos

- [ ] Render de disponibilidad.
- [ ] Estados de carga/error.
- [ ] Formulario de reserva valida campos.
- [ ] Router no entra en bucle.
- [ ] Ruta desconocida muestra Not Found.
- [ ] Usuario normal no entra a rutas administrativas.
- [ ] Cargas concurrentes del usuario comparten una sola solicitud.
- [ ] Usuario sin RUT no puede abrir creacion de reserva.
- [ ] Recurso no reservable no abre el formulario.
- [ ] Conflicto de horario mantiene error visible en el modal.
- [ ] Cancelacion exige confirmacion antes de llamar a la API.
- [ ] Helpers temporales cubren futura, en curso, finalizada y zona horaria definida.
- [x] `npm test` y `npm run build` pasan para la base actual; falta ampliar los casos anteriores.

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

## UX-003 - Mostrar feedback preventivo de conflicto antes de confirmar reserva

Prioridad: P2
Labels: `frontend`, `ux`, `reservas`, `disponibilidad`, `codex-ready`, `mvp2`
Estado sugerido: Ready for Codex

### Contexto

El backend y `AvailabilitySection.vue` ya rechazan cruces de horario al confirmar. Sin embargo, el usuario puede cambiar duracion, fecha, hora o recurso dentro del modal y recien enterarse del conflicto al presionar confirmar.

Para MVP 2, el flujo debe prevenir mejor el error antes del submit, especialmente cuando el usuario ajusta duracion.

### Objetivo

Dar feedback inmediato o casi inmediato cuando la seleccion actual cruza con una reserva, taller, bloqueo o actividad institucional conocida por el frontend.

### Alcance sugerido para Codex

1. Exponer desde `AvailabilitySection.vue` una validacion reutilizable para la seleccion actual.
2. Pasar al formulario un mensaje preventivo o un flag de conflicto.
3. Recalcular conflicto al cambiar recurso, fecha, hora o duracion.
4. Mostrar el rango final completo antes de confirmar.
5. Deshabilitar `Confirmar Reserva` solo cuando el conflicto sea seguro.
6. Mantener la validacion final de backend como fuente de verdad.

### Criterios de aceptacion

- [ ] Al cambiar duracion, la UI avisa si el rango cruza con ocupacion existente.
- [ ] El formulario muestra inicio, termino y duracion final.
- [ ] Si el conflicto es seguro, el boton de confirmar queda deshabilitado.
- [ ] Si el frontend no puede validar por datos incompletos, el backend sigue mostrando error claro al confirmar.
- [ ] No se bloquean recursos `OPEN_USE` por concurrencia si su regla permite uso compartido.
- [ ] `npm run build` pasa.

### Archivos relevantes

- `frontend/src/components/availability/AvailabilitySection.vue`
- `frontend/src/components/forms/ReservationForm.vue`
- `frontend/src/components/forms/DateTimePicker.vue`
- `frontend/src/utils/reservationTime.js`

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
Estado sugerido: Validado por usuario

### Objetivo

Evitar que `DEV_AUTH_ENABLED=true` quede activo en un ambiente publico.

### Criterios de aceptacion

- [x] La documentacion de despliegue exige `DEV_AUTH_ENABLED=false`.
- [x] Existe verificacion manual antes de entrega.
- [x] Se evalua bloquear arranque si modo dev esta activo con origenes productivos.
- [x] El README explica que las cabeceras `X-Dev-Auth-*` son solo locales.

### Resultado de validacion

- El usuario confirmo que el checklist productivo de modo desarrollo esta validado para MVP 1.
- Para MVP 1 se mantiene como control manual/documental; un bloqueo automatico de arranque puede evaluarse despues si la operacion pasa de demo a produccion institucional.

## SEC-005 - No exponer errores internos en respuestas HTTP

Prioridad: P1
Labels: `backend`, `security`, `errors`, `observability`, `codex-ready`, `mvp1`
Estado sugerido: Ready for Codex

### Contexto

Varios handlers y el middleware de autenticacion incluyen `err.Error()` en campos `detail` o `error` de la respuesta. Aunque el frontend no siempre lo muestre, cualquier cliente puede leer detalles de SQL, JWT o configuracion interna.

### Objetivo

Separar diagnostico interno de mensajes publicos, conservando respuestas utiles para el usuario y trazabilidad para desarrollo.

### Alcance sugerido para Codex

1. Crear un helper de respuesta de error con codigo HTTP, codigo publico estable y mensaje seguro.
2. Registrar el error tecnico en backend con contexto minimo y sin tokens, passwords, RUT ni datos personales innecesarios.
3. Reemplazar `err.Error()` en respuestas 500 de handlers y middleware.
4. Mantener mensajes de negocio controlados para conflictos, RUT faltante, recurso no reservable y permisos.
5. No diferenciar publicamente entre recurso privado inexistente y no autorizado cuando eso permita enumeracion.
6. Permitir detalle tecnico solo bajo una bandera local explicita, desactivada por defecto y prohibida en Azure.
7. Agregar pruebas que aseguren que respuestas 500 no contienen texto del error interno simulado.

### Criterios de aceptacion

- [ ] Ninguna respuesta 500 incluye `err.Error()` sin sanitizar.
- [ ] Errores de negocio conservan mensajes accionables y codigos estables.
- [ ] Logs internos permiten correlacionar el fallo sin exponer secretos.
- [ ] Autenticacion no devuelve detalle de validacion JWT o JWKS al cliente.
- [ ] Usuario normal no puede inferir existencia de datos ajenos por diferencias de error.
- [ ] `go test ./...` pasa.

### Archivos relevantes

- `backend/internal/middleware/auth_middleware.go`
- `backend/internal/handlers/reservations_handlers.go`
- `backend/internal/handlers/resources_handlers.go`
- `backend/internal/handlers/workshops_handlers.go`
- `backend/internal/handlers/users_handlers.go`

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

## DOC-006 - Documentar guia de documentacion y legibilidad

Prioridad: P1
Labels: `documentacion`, `calidad`, `codigo`
Estado sugerido: Done

### Objetivo

Formalizar criterios para documentar codigo, SQL y Markdown sin agregar comentarios redundantes ni alterar comportamiento.

### Criterios de aceptacion

- [x] Existe documento dedicado en `docs/`.
- [x] Define que comentar y que evitar.
- [x] Cubre Go, Vue/JavaScript, SQL Server y Markdown.
- [x] Registra contratos criticos de Poli-REDI.
- [x] Indica validaciones esperadas segun tipo de cambio.

### Resultado de implementacion

- Se creo `docs/12-guia-documentacion-legibilidad.md`.
- Se agregaron comentarios de alto valor en autenticacion, reservas, manejo temporal y triggers SQL.

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
- Se recibieron y analizaron cinco calendarios iCal unicos: Cancha 1, Cancha 2, Cancha 3, Sala Multiuso y Sala de Musculacion/Gimnasio.
- Se detecto y excluyo una copia duplicada de Sala Multiuso.
- Se propuso el mapeo de cada calendario recibido al recurso equivalente de Poli-REDI; falta validacion operativa del usuario.
- Se adapto la disponibilidad para incluir `scheduled_activities`: el administrador ve el titulo original y el usuario normal ve un bloque institucional sin datos personales.
- La expansion preliminar produjo 217 ocurrencias futuras y detecto 14 solapamientos heredados, ademas de talleres duplicados respecto de la planilla oficial.
- La carga de eventos y el congelamiento del legado siguen pendientes; no se ha ejecutado importacion sobre Azure SQL.

### Archivos relevantes

- `docs/11-plan-corte-google-calendar.md`
- `docs/07-backlog.md`
- `docs/09-mvps-roadmap.md`

---

## QA-003 - Hacer coherente el seed temporal con su fecha de uso

**Estado:** Cerrado y verificado en Azure SQL el 2026-07-21

**Severidad QA:** P2 - Medio

**Responsable sugerido:** Desarrollador

### Contexto

QA detecto que `database/seed_today_temp.sql` se presentaba como un seed temporal para pruebas "de hoy", pero usaba fechas fijas del 2026-07-14. La correccion calcula una fecha base institucional mediante `Pacific SA Standard Time` y deriva desde ella reservas, participantes, bloqueos y actividades. El encadenamiento con `schema.sql` y `seed.sql` fue ejecutado correctamente en Azure SQL.

### Correccion requerida

- Verificar que la fecha base corresponde al dia institucional de `America/Santiago`.
- Ejecutar el script corregido despues de `database/schema.sql` y `database/seed.sql` en una base descartable.

No modificar requisitos, reglas de reserva ni el alcance de los incrementos posteriores para resolver esta tarea.

### Criterios de aceptacion

- [x] El nombre, encabezado y comportamiento del script describen el mismo proposito por inspeccion estatica.
- [x] El script deriva sus fechas desde el dia institucional de ejecucion.
- [x] Reservas, participantes, bloqueos y actividades usan una fecha base comun por inspeccion estatica.
- [x] El script se ejecuta despues de `database/schema.sql` y `database/seed.sql` sin errores en una base descartable.
- [x] `go test ./...`, `go vet ./...`, `npm test` y `npm run build` continúan aprobando.

### Evidencia de QA

- Detectado el 2026-07-21 durante la validacion del incremento tecnico 1 de politicas.
- `database/schema.sql` y `database/seed.sql` fueron reportados como ejecutados correctamente sobre una base limpia.
- El 2026-07-21 Azure SQL ejecuto todos los lotes de `seed_today_temp.sql` sin errores en 1,936 segundos: 8 reservas, 6 participantes, 2 bloqueos, 3 actividades programadas y 3 notificaciones insertadas; el lote finalizo con `Commands completed successfully`.

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
- `frontend/src/views/ReservationsView.vue`: concentra reservas activas e historial, con fuente de datos segun rol.
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
- Accesos rapidos del dashboard: eliminados de la vista actual para simplificar el panel principal.
- `frontend/src/components/availability/CalendarMini.vue`: nombres de meses y dias.
- `frontend/src/components/forms/DateTimePicker.vue`: opciones de duracion centralizadas en `frontend/src/utils/reservationRules.js` y alineadas con la validacion backend de `RES-011`.

---

# Orden recomendado de ejecucion

## Bloque 1 - Verificacion final de reglas MVP 1

1. Verificar online `RES-009` con una reserva de hora conocida en local y Azure.
2. Verificar online `RES-010` con request manipulado y transiciones invalidas.
3. Verificar online `RES-011` con apertura, ultimo bloque valido, cierre excedido y duracion invalida.
4. `QA-001` Ampliar pruebas backend para permisos, conflictos y casos aun no cubiertos.
5. `QA-003` Corregir o renombrar el seed temporal para que su fecha coincida con su proposito declarado.

Estas tareas ya tienen implementacion inicial. Antes de cerrar el MVP 1 falta evidencia desplegada y ampliar cobertura de permisos/conflictos que no queda cubierta por las pruebas unitarias actuales.

## Bloque 2 - Flujo visible y coherencia transversal MVP 1

1. `BACK-020` Impedir seleccion de instalaciones no reservables.
2. `BACK-021` Corregir shell responsive y controles globales.
3. `BACK-022` Completar accesibilidad de modales y calendario.
4. `BACK-023` Estabilizar navegacion y carga de sesion.
5. `SEC-005` No exponer errores internos en respuestas HTTP.
6. `BACK-018` Terminar carrusel y dashboard.
7. `BACK-024` Hacer reproducible la instalacion frontend y retirar duplicados muertos.
8. `QA-002` Agregar regresion frontend automatizada para los flujos corregidos.

## Bloque 3 - Cierre de flujo usuario MVP 2

1. `RES-012` Aplicar la ventana, frecuencia y liberacion de oportunidad aprobadas.
2. `RES-008` Registrar participantes con cuenta, bloquear `PENDING` y cancelar al vencer bajo el minimo.
3. `API-002` Filtrar reservas por fecha/rango/estado.
4. `API-004` Filtrar disponibilidad e integrar ocupaciones institucionales.
5. `API-006` Consultar detalle de reserva por ID.
6. `UX-003` Prevenir conflictos antes de confirmar.
7. `NOTIF-001` Completar leidas/no leidas y destino real de notificaciones.

## Bloque 4 - Administracion y analitica

1. `RES-004` Completar integracion de bloqueos; actividades programadas ya estan incorporadas.
2. `ADMIN-004` Crear bloqueos de disponibilidad.
3. `ADMIN-003` Completar gestion del inventario oficial de ocho recursos.
4. `ADMIN-002` Completar gestion de usuarios.
5. `ADMIN-005` Programacion institucional y resolucion de conflictos aprobada.
6. `ADMIN-006` Configurar politicas de reserva exclusivamente como administrador.
7. `REP-003` Corregir semantica y precision de indicadores.
8. `REP-002` Completar infracciones.

# Tareas Codex-ready recomendadas

La siguiente tarea para el agente de desarrollo debe ser la verificacion desplegada de `RES-009`, `RES-010` y `RES-011`. Si esa evidencia pasa, continuar con `QA-001` para ampliar cobertura de permisos/conflictos y luego `SEC-005`.

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
npm test
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
