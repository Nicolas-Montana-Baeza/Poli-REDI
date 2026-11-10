# Poli-REDI - Requisitos, historias de usuario y casos de uso

## Objetivo del documento

Este documento consolida los requisitos funcionales, requisitos no funcionales, historias de usuario y casos de uso principales de Poli-REDI.

Debe mantenerse actualizado junto con el backlog, la arquitectura, el flujo de reservas y la documentacion tecnica cada vez que cambie el alcance funcional del sistema.

## Alcance del sistema

Poli-REDI es un sistema web para gestionar reservas deportivas institucionales. Permite a usuarios autenticados consultar disponibilidad, registrar reservas, cancelar reservas propias, revisar historial, inscribirse en talleres deportivos y recibir notificaciones. Los administradores pueden revisar informacion operacional, usuarios, recursos, reportes y, en iteraciones futuras, gestionar bloqueos, recursos, programacion institucional e infracciones.

## Actores

### Usuario normal

Persona autenticada que usa recursos deportivos institucionales.

Responsabilidades principales:

- Consultar disponibilidad.
- Crear reservas propias.
- Registrar o actualizar RUT cuando corresponda.
- Revisar sus reservas e historial.
- Consultar talleres deportivos e inscribirse cuando existan cupos.
- Cancelar reservas permitidas.
- Revisar notificaciones.

### Administrador

Usuario autenticado con permisos administrativos.

Responsabilidades principales:

- Acceder al panel administrativo.
- Revisar usuarios, recursos, reservas y reportes.
- Cancelar reservas cuando corresponda.
- Gestionar usuarios, recursos, bloqueos, programacion e infracciones en futuras iteraciones.

### Sistema de autenticacion Microsoft Entra ID

Proveedor externo de identidad.

Responsabilidades principales:

- Autenticar usuarios.
- Entregar tokens para consumir la API protegida.
- Proveer datos base de identidad, como correo, nombre, tenant y object id.

### Base de datos PostgreSQL 16

Sistema de persistencia.

Responsabilidades principales:

- Guardar usuarios, recursos, actividades, reservas, talleres, inscripciones, notificaciones y datos administrativos.
- Aplicar reglas de integridad y restricciones criticas.
- Proveer vistas de apoyo para reportes y calendario.

## Requisitos funcionales

### RF-001 - Autenticacion de usuarios

El sistema debe permitir ingreso mediante Microsoft Entra ID y, en ambiente de desarrollo, mediante autenticacion local controlada por variables de entorno.

**Estado vigente (2026-10-26):** Implementado y validado para el prototipo MVP3.

**Implementación verificada en código:** Middleware Entra ID con validación JWT, issuer y audience; modo de desarrollo condicionado por DEV_AUTH_ENABLED.

**Trazabilidad técnica:** `backend/internal/middleware/auth_middleware.go`; `backend/internal/routes/routes.go`.

**Seguimiento:** No constituye una certificación de seguridad para producción institucional.

### RF-002 - Sincronizacion de usuario autenticado

El sistema debe obtener o crear el usuario local asociado a la identidad autenticada.

**Estado vigente (2026-10-26):** Implementado, con deuda de seguridad pendiente.

**Implementación verificada en código:** Creación y recuperación del usuario local por correo; sincronización de entra_oid y tenant_id.

**Trazabilidad técnica:** `backend/internal/repositories/users_repoository.go`; `backend/internal/middleware/auth_middleware.go`.

**Seguimiento:** SEC-006: comprobar vinculación estable mediante tenant_id y entra_oid antes de producción.

### RF-003 - Registro y validacion de RUT

El sistema debe permitir que usuarios normales registren o actualicen su RUT, validando formato y digito verificador.

**Estado vigente (2026-10-26):** Implementado y validado funcionalmente.

**Implementación verificada en código:** Registro y actualización del RUT mediante endpoint autenticado y validador de dígito verificador.

**Trazabilidad técnica:** `backend/internal/handlers/me_handler.go`; `backend/internal/validators/rut.go`.

**Seguimiento:** Sin brechas funcionales identificadas para el alcance MVP3.

### RF-004 - Bloqueo de reserva sin RUT

El sistema debe impedir que usuarios normales sin RUT creen reservas.

**Estado vigente (2026-10-26):** Implementado y validado funcionalmente.

**Implementación verificada en código:** El backend rechaza la creación de reservas de usuarios no administradores sin RUT habilitante.

**Trazabilidad técnica:** `backend/internal/handlers/reservations_handlers.go`; `backend/internal/services/reservations_service.go`.

**Seguimiento:** La validación del servidor es obligatoria; la interfaz no constituye el control de seguridad.

### RF-005 - Consulta de disponibilidad

El sistema debe mostrar disponibilidad de recursos por fecha, considerando reservas existentes, talleres activos, actividades institucionales y bloqueos asociados al recurso.

**Estado vigente (2026-10-26):** Implementado y validado para el alcance MVP3.

**Implementación verificada en código:** Disponibilidad por rango, reservas, bloques administrativos y actividades institucionales programadas; sanitización de información personal.

**Trazabilidad técnica:** `backend/internal/handlers/reservations_handlers.go`; `backend/internal/services/reservations_service.go`; `backend/internal/handlers/reservations_handlers_test.go`; `backend/internal/services/availability_blocks_calendar_integration_test.go`.

**Seguimiento:** Se mantienen pendientes las evaluaciones de rendimiento y accesibilidad del calendario.

**Referencias históricas:** `RES-004`. Su estado debe contrastarse con el backlog vigente.

### RF-006 - Creacion de reservas

El sistema debe permitir crear solicitudes de reserva sobre recursos disponibles, asociadas al usuario autenticado y, opcionalmente, a una actividad. El servidor debe asignar el estado segun la politica del recurso, aplicar la restriccion semanal y aceptar para todos los recursos solo duraciones de 30, 60, 90, 120, 150 o 180 minutos dentro de la jornada institucional.

**Estado vigente (2026-10-26):** Implementado, con pruebas técnicas y funcionales documentadas.

**Implementación verificada en código:** Creación con usuario autenticado, políticas versionadas, duraciones permitidas, restricciones de frecuencia, reglas grupales y persistencia PostgreSQL.

**Trazabilidad técnica:** `backend/internal/services/reservations_service.go`; `backend/internal/repositories/reservations_repository.go`; `backend/internal/services/reservations_service_test.go`; `backend/internal/services/group_reservation_integration_test.go`.

**Seguimiento:** Conservar como seguimiento las comprobaciones manuales y de ambiente online aún no cerradas explícitamente en docs/15-checklist-cierre-mvp2.md.

**Referencias históricas:** `RES-008`, `RES-009`, `RES-010`, `RES-011`. Su estado debe contrastarse con el backlog vigente.

### RF-007 - Validacion de conflictos de reserva

El sistema debe rechazar reservas que se solapen con otras reservas confirmadas, talleres activos o reglas de disponibilidad. Los recursos `OPEN_USE` permiten concurrencia y no bloquean por reserva existente.

**Estado vigente (2026-10-26):** Implementado y validado técnicamente para MVP3.

**Implementación verificada en código:** Restricciones de solapamiento entre reservas, bloqueos y programación institucional, con tratamiento específico de recursos OPEN_USE.

**Trazabilidad técnica:** `database/postgres/migrations/PG16_0003_mvp1_invariants.sql`; `database/postgres/migrations/PG16_0006_mvp2_institutional_availability.sql`; `backend/internal/services/open_use_integration_test.go`; `backend/internal/services/availability_blocks_calendar_integration_test.go`.

**Seguimiento:** Mantener pruebas de concurrencia y regresión ante futuras modificaciones de políticas.

**Referencias históricas:** `RES-004`, `RES-010`, `ADMIN-004`. Su estado debe contrastarse con el backlog vigente.

### RF-008 - Cancelacion de reservas

El sistema debe permitir cancelar reservas propias y permitir que administradores cancelen reservas segun permisos, siempre que el estado admita cancelacion y la reserva no haya finalizado segun la zona horaria institucional.

**Estado vigente (2026-10-26):** Implementado, con validación técnica y funcional.

**Implementación verificada en código:** Cancelación autorizada por propietario o administrador; comprobación de estado y finalización; transacción y notificación administrativa cuando corresponda.

**Trazabilidad técnica:** `backend/internal/handlers/reservations_handlers.go`; `backend/internal/services/reservations_service.go`; `backend/internal/repositories/reservations_repository.go`.

**Seguimiento:** Conservar seguimiento operativo de pruebas en el ambiente definitivo.

**Referencias históricas:** `RES-009`, `RES-010`. Su estado debe contrastarse con el backlog vigente.

### RF-009 - Reservas por rol

El sistema debe mostrar reservas con fecha, recurso, actividad, duracion y estado. Usuarios normales solo ven sus reservas; administradores pueden ver todas las reservas.

**Estado vigente (2026-10-26):** Implementado y validado para el prototipo.

**Implementación verificada en código:** Consulta de reservas propias y listado global restringido a administradores.

**Trazabilidad técnica:** `backend/internal/routes/routes.go`; `backend/internal/handlers/reservations_handlers.go`; `backend/internal/repositories/reservations_repository.go`.

**Seguimiento:** Mantener pruebas de autorización por rol.

### RF-010 - Historial por rol

El sistema debe mostrar reservas pasadas, canceladas o finalizadas. Usuarios normales solo ven su historial; administradores pueden ver todo el historial.

**Estado vigente (2026-10-26):** Implementado para el alcance MVP3.

**Implementación verificada en código:** Consulta de historial y detalle según propiedad de reserva o rol administrativo, con presentación de estados y fechas.

**Trazabilidad técnica:** `backend/internal/services/reservations_service.go`; `backend/internal/repositories/reservations_repository.go`; `frontend/src/views/ReservationsView.vue`.

**Seguimiento:** La paginación y el rendimiento con grandes volúmenes quedan sujetos a evaluación posterior.

### RF-011 - Catalogo de recursos

El sistema debe mostrar recursos deportivos con datos reales e imagen configurable cuando exista.

**Estado vigente (2026-10-26):** Implementado y validado funcionalmente.

**Implementación verificada en código:** Catálogo real de recursos PostgreSQL, con imagen configurable y modos operativos.

**Trazabilidad técnica:** `backend/internal/handlers/resources_handlers.go`; `backend/internal/repositories/resources_repository.go`; `frontend/src/views/ResourcesView.vue`.

**Seguimiento:** La administración completa del inventario se traza separadamente en RF-014 y RF-024.

### RF-012 - Panel administrador

El sistema debe mostrar un panel administrativo solo a usuarios administradores.

**Estado vigente (2026-10-26):** Implementado y validado para el prototipo MVP3.

**Implementación verificada en código:** Panel de administración, rutas frontend protegidas y autorización administrativa verificada en backend.

**Trazabilidad técnica:** `frontend/src/views/AdminView.vue`; `frontend/src/router/index.js`; `backend/internal/middleware/auth_middleware.go`; `backend/internal/routes/admin_authorization_http_test.go`.

**Seguimiento:** Mantener las verificaciones de permisos en toda ruta administrativa nueva.

### RF-013 - Gestion de usuarios

El sistema debe permitir a administradores listar usuarios, bloquear o desbloquear cuentas con justificacion, asignar o revocar roles administrativos y consultar la trazabilidad de esas operaciones.

Estado documentado previamente (histórico): Implementado parcialmente.

Referencias históricas de backlog: `ADMIN-002`.

**Estado vigente (2026-10-26):** Implementado y validado para MVP3.

**Implementación identificada:** Gestión de usuarios, bloqueo/desbloqueo, roles y auditoría.

**Evidencia y trazabilidad:** `docs/mvp3/cierre-admin002-usuarios.md`; `backend/internal/handlers/admin_users_handlers.go`; `backend/internal/handlers/admin_user_roles_handlers.go`

**Seguimiento:** ADMIN-002 cerrado. Revisar la vinculación de identidad Entra (SEC-006).

### RF-014 - Gestion de recursos

El sistema debe permitir que administradores creen, editen, activen o desactiven los recursos del inventario oficial. Los ocho recursos actuales constituyen la linea base aprobada y los cambios deben reflejarse en catalogo y disponibilidad.

Estado documentado previamente (histórico): IMPLEMENTADO PARCIALMENTE. Imagen, modo,
capacidad y estado activo/inactivo ya cuentan con implementacion.
Creacion, edicion completa y asignacion administrativa de sede
permanecen pendientes.

Refinamiento aprobado 2026-10-13: EV-012 amplia el alcance
de RF-014 para incorporar alta, edicion integral, reasignacion
de sede y desactivacion sin eliminacion fisica.

Referencias históricas de backlog: `ADMIN-003`.

**Estado vigente (2026-10-26):** Implementado para MVP3; parcial respecto del alcance ampliado.

**Implementación identificada:** Creación, edición, desactivación y filtros de recursos.

**Evidencia y trazabilidad:** `backend/internal/handlers/resources_handlers.go`; `backend/internal/repositories/resources_repository.go`; `frontend/src/views/ResourcesView.vue`

**Seguimiento:** ADMIN-003 cerrado para MVP3. Trasladar recursos existentes entre sedes sigue fuera del alcance implementado.

### RF-015 - Bloqueos de disponibilidad

El sistema debe permitir que administradores creen bloqueos por mantencion, cierre, evento u otro motivo.

Estado documentado previamente (histórico): Pendiente.

Referencias históricas de backlog: `ADMIN-004`.

**Estado vigente (2026-10-26):** Implementado y validado para MVP3.

**Implementación identificada:** Creación, consulta y desactivación de bloqueos, con integración al calendario.

**Evidencia y trazabilidad:** `docs/mvp3/cierre-4d-bloqueos.md`; `backend/internal/handlers/availability_blocks_admin_handlers.go`; `backend/internal/services/availability_blocks_calendar_integration_test.go`

**Seguimiento:** ADMIN-004 cerrado; mantener pruebas de concurrencia y regresión.

### RF-016 - Programacion institucional

El sistema debe permitir registrar clases, talleres, eventos, campeonatos o entrenamientos institucionales.

Estado documentado previamente (histórico): IMPLEMENTADO PARCIAL. El backend MVP 2 permite crear programacion institucional asociada a unidades y recursos, con autorizacion para administradores o gestores de unidad, tipos estructurados y horarios `SINGLE` o `WEEKLY`. Existen pruebas de integracion sobre persistencia, recurrencia, disponibilidad y bloqueos administrativos. Permanecen pendientes el cierre de interfaz y la validacion funcional completa del flujo institucional.

Referencias históricas de backlog: `ADMIN-005`.

**Estado vigente (2026-10-26):** Implementado para el prototipo MVP3.

**Implementación identificada:** Programación institucional, recurrencia, edición y autorización de gestores.

**Evidencia y trazabilidad:** `backend/internal/services/institutional_activities_service.go`; `frontend/src/views/InstitutionalActivitiesView.vue`; `backend/internal/services/institutional_activities_integration_test.go`

**Seguimiento:** ADMIN-005 cerrado para el prototipo; EV-010 continúa pendiente de ratificación institucional.

### RF-017 - Notificaciones

El sistema debe mostrar notificaciones reales del usuario.

Estado documentado previamente (histórico): Implementado parcialmente.

Referencias históricas de backlog: `NOTIF-001`.

**Estado vigente (2026-10-26):** Implementado en el prototipo; evidencia integral pendiente de consolidación.

**Implementación identificada:** Notificaciones persistidas, consulta, marcado como leídas y eventos institucionales.

**Evidencia y trazabilidad:** `backend/internal/repositories/notifications_repository.go`; `backend/internal/handlers/notifications_handlers.go`; `frontend/src/components/layout/NotificationBell.vue`

**Seguimiento:** NOTIF-001: consolidar evidencia del ciclo completo de generación y recepción.

### RF-018 - Reportes

El sistema debe mostrar indicadores de uso, horas punta y datos administrativos.

Estado documentado previamente (histórico): Implementado parcialmente.

Referencias históricas de backlog: `REP-001`, `REP-002`, `REP-003`.

**Estado vigente (2026-10-26):** Parcial: REP-003 cerrado; REP-001 y REP-002 pendientes.

**Implementación identificada:** Indicadores de reservas vigentes, duraciones y utilización de recursos.

**Evidencia y trazabilidad:** `docs/mvp3/cierre-rep003-reportes.md`; `frontend/src/utils/reportMetrics.js`; `frontend/src/views/ReportsView.vue`

**Seguimiento:** Completar analítica adicional y módulo de infracciones en MVP4.

### RF-019 - Talleres deportivos

El sistema debe permitir que usuarios autenticados consulten talleres activos e inscribirse cuando tengan RUT y existan cupos.

Estado documentado previamente (histórico): Implementado.

**Estado vigente (2026-10-26):** Implementado con pruebas de integración.

**Implementación identificada:** Consulta de talleres institucionales e inscripción de usuarios autorizados.

**Evidencia y trazabilidad:** `backend/internal/services/institutional_workshops_service.go`; `backend/internal/services/institutional_workshops_integration_test.go`

**Seguimiento:** Conservar pruebas funcionales sobre cupos, inscripciones y retiros.

### RF-020 - Restriccion semanal de reservas particulares

El sistema debe limitar las fechas reservables al periodo institucional configurado y evitar que un usuario cree mas de una solicitud dentro del periodo contado desde la fecha local de creacion de su solicitud anterior. `PENDING` y `CONFIRMED` consumen la oportunidad; `CANCELLED` deja de consumirla. Con siete dias, un martes permite elegir hasta el lunes siguiente y una solicitud creada ese martes permite volver a solicitar desde el martes posterior.

Estado documentado previamente (histórico): APROBADO e IMPLEMENTADO. La verificacion del 2026-07-21 corresponde a la etapa Azure SQL y se conserva como evidencia historica. La arquitectura vigente PostgreSQL cuenta con pruebas locales/de integracion; permanece pendiente el cierre integral del ambiente objetivo y la concurrencia de punta a punta donde corresponda.

Referencias históricas de backlog: `RES-012`.

**Estado vigente (2026-10-26):** Implementado con validaciones PostgreSQL.

**Implementación identificada:** Frecuencia de solicitudes, ventanas configurables y estados que consumen oportunidad.

**Evidencia y trazabilidad:** `backend/internal/repositories/reservations_repository.go`; `backend/internal/services/reservations_service.go`

**Seguimiento:** Conservar seguimiento de regresión y cierre operativo del flujo grupal.

### RF-021 - Confirmacion de participantes minimos

El sistema debe registrar participantes unicos, incluido el solicitante, y aplicar el minimo versionado del recurso. El valor inicial es 10 para Cancha 1, Cancha 2 y Sala Multiuso. Todos los participantes deben tener cuenta. Altas y retiros se aceptan antes del deadline configurable de 60 minutos y se cierran al alcanzarlo.

Estado documentado previamente (histórico): APROBADO e IMPLEMENTADO; cierre integral 14D pendiente.

Implementado actualmente:

- las reservas de recursos grupales nacen en `PENDING`;
- el solicitante queda registrado como primer participante;
- se persisten participantes por cuenta autenticada;
- se expone avance, minimo y condicion grupal;
- existe codigo de invitacion;
- alcanzar el minimo cambia la reserva a `CONFIRMED`;
- el solape personal impide incorporaciones incompatibles;
- el deadline unico rige en `PENDING` y `CONFIRMED`;
- una solicitud que vence `PENDING` bajo el minimo se cancela con `MINIMUM_NOT_MET`.

Pendiente de cierre:

- completar la evidencia manual de liberacion del horario tras `MINIMUM_NOT_MET`;
- ejecutar la cadena efimera y el smoke test online definidos para el cierre MVP2.

Referencias históricas de backlog: `RES-008`.

**Estado vigente (2026-10-26):** Implementado, con evidencia manual de cierre aún por conciliar.

**Implementación identificada:** Participantes autenticados, mínimo por recurso, capacidad y confirmación.

**Evidencia y trazabilidad:** `backend/internal/repositories/participants_repository.go`; `backend/internal/repositories/participants_repository_integration_test.go`

**Seguimiento:** Conciliar las comprobaciones pendientes de docs/15-checklist-cierre-mvp2.md.

### RF-022 - Estado y condicion grupal segun politica del recurso

El sistema debe separar el ciclo de vida de la reserva de la condicion operacional del grupo.

Para recursos sujetos a confirmacion grupal:

- la solicitud comienza `PENDING` con `groupCondition = PENDING_MINIMUM`;
- al alcanzar el minimo por primera vez cambia a `CONFIRMED` con `groupCondition = HEALTHY`;
- una reserva que ya fue confirmada no regresa a `PENDING` por una retirada posterior;
- si el conteo baja del minimo despues de confirmar, conserva `CONFIRMED` con `groupCondition = AT_RISK`;
- si recupera el minimo, vuelve a `HEALTHY`;
- si al alcanzar el plazo limite la solicitud permanece `PENDING` bajo el minimo, se cancela automaticamente con `MINIMUM_NOT_MET`, libera el horario y deja de consumir la oportunidad semanal;
- si la reserva ya alcanzo `CONFIRMED` y esta en `AT_RISK` al plazo limite, conserva `CONFIRMED` y no se cancela automaticamente por `MINIMUM_NOT_MET`;
- la ventana de alta y retiro de participantes cierra en el instante exacto del deadline (no inclusive);
- los recursos `OPEN_USE` no requieren confirmacion grupal.

Estado documentado previamente (histórico): APROBADO e IMPLEMENTADO PARCIAL.

La creacion `PENDING`, persistencia de participantes, transicion al minimo, condicion `AT_RISK` y vencimiento automatico `MINIMUM_NOT_MET` estan implementados. La comprobacion manual de liberacion y la evidencia online permanecen abiertas; los generadores generales de notificaciones estan fuera de este cierre.

Decision refinada el 2026-08-20: reemplaza la regla anterior que hacia regresar una reserva ya confirmada a `PENDING`.

Referencias históricas de backlog: `RES-008`, `RES-010`, `NOTIF-001`.

**Estado vigente (2026-10-26):** Implementado parcialmente respecto del requisito completo. El vencimiento automático MINIMUM_NOT_MET solo se aplica a reservas PENDING.

**Implementación identificada:** Estados PENDING y CONFIRMED, condiciones HEALTHY y AT_RISK, y vencimiento.

**Evidencia y trazabilidad:** `backend/internal/services/participants_service.go`; `backend/internal/repositories/reservation_expiration_integration_test.go`

**Seguimiento:** Verificar evidencia manual de liberación de horario y resultados del ambiente objetivo.

**Brecha funcional y normativa identificada:** La definición de RF-022 contempla cancelar una reserva bajo el mínimo al alcanzar el deadline. Sin embargo, `ExpirePendingGroupReservations` solo cancela solicitudes `PENDING`. Una reserva previamente confirmada que pasa a `CONFIRMED + AT_RISK` no se cancela por este proceso. El comportamiento coincide con `docs/06-flujo-reservas.md`, pero requiere reconciliación con RF-022.

**Decisión pendiente:** Formalizar si `CONFIRMED + AT_RISK` debe conservarse o cancelarse al deadline. No declarar cumplimiento completo hasta resolver esta diferencia.

### RF-023 - Resolucion de conflictos institucionales

**RF-023/v1 — Regla originalmente aprobada (antecedente historico)**

El sistema debe cancelar automaticamente una reserva particular cuando entra en conflicto con una actividad institucional y notificar al usuario afectado. Si el conflicto es entre dos actividades, debe informar al administrador y permitirle cancelar cualquiera de ellas o mantener ambas.

**RF-023/v2 — Implementacion operativa provisional EV-010 (2026-10-20)**

El sistema debe detectar y registrar conflictos de programacion institucional sin cancelar automaticamente las reservas particulares involucradas.

La resolucion corresponde exclusivamente a un administrador autorizado, que podra decidir sobre cada ocupacion mediante `KEEP`, `ALLOW`, `CANCEL` o `RESCHEDULE`, conforme a las restricciones aplicables.

Criterios de aceptacion propuestos:

- Detectar automaticamente conflictos entre reservas y actividades.
- Registrar conflictos de multiples ocupaciones como `PENDING`.
- No cancelar reservas particulares al crear actividades institucionales.
- Permitir consultar ocupaciones y conflictos pendientes.
- Restringir resoluciones administrativas a administradores.
- Exigir justificacion y confirmacion explicita de cada decision.
- Registrar responsable, decision y efectos de la resolucion.
- Rechazar planes incompatibles sin efectos parciales.
- Generar nuevos conflictos si una reprogramacion produce solapamientos.
- Notificar al propietario cuando una decision administrativa cancele su reserva.
- No exponer datos personales de terceros mediante notificaciones.
- Permitir que los MANAGER programen sus unidades sin resolver conflictos administrativos.
- Mantener los bloqueos de disponibilidad como restricciones duras.

**Estado de implementacion y validacion**

La deteccion, consulta, resolucion administrativa y persistencia ya estan implementadas. Se aprobaron seis pruebas institucionales PostgreSQL 16 en entorno aislado el 2026-10-20.

La generacion transaccional de notificaciones por cancelacion institucional fue verificada en PostgreSQL 16, incluyendo rollback ante fallo inducido. El usuario confirmo la recepcion de la notificacion en Funnel tras actualizar con F5.

Las tres vistas ADMIN-005 aprobaron 18/18 pruebas Vue. La cancelacion y notificacion fueron comprobadas en Funnel. La actualizacion de notificaciones al abrir la campana tambien fue verificada en el despliegue publicado. Permanecen pendientes las validaciones integrales restantes del MVP3 y la ratificacion institucional de EV-010.

**Estado normativo:** RF-023/v2 fue adoptado provisionalmente por el responsable del proyecto para MVP3. Los responsables institucionales manifiestan una preferencia preliminar por esta version, pero no existe ratificacion definitiva. RF-023/v1 se conserva como antecedente historico. EV-010 permanece abierto exclusivamente para su formalizacion institucional y las definiciones operativas pendientes, documentadas en `docs/14-evolucion-y-trazabilidad-requisitos.md`.

Referencias históricas de backlog: `ADMIN-005`, `RES-004`, `NOTIF-001`.

**Estado vigente (2026-10-26):** Implementado bajo RF-023/v2 provisional; pendiente ratificación EV-010.

**Implementación identificada:** Detección de conflictos y resolución administrativa con persistencia y notificaciones.

**Evidencia y trazabilidad:** `backend/internal/services/scheduling_conflict_resolution_service.go`; `backend/internal/services/scheduling_conflict_resolution_integration_test.go`

**Seguimiento:** Conservar RF-023/v1 como antecedente; no declarar ratificación institucional definitiva.

### RF-024 - Inventario oficial administrable

El sistema debe permitir que el administrador mantenga el inventario oficial inicial de ocho recursos en el MVP administrativo correspondiente, conservando el historial aplicable.

Estado documentado previamente (histórico): APROBADO; IMPLEMENTADO PARCIALMENTE.
Existe consulta y actualizacion de imagen, modo, capacidad y estado.
EV-012 (2026-10-13) amplia el inventario administrable para
permitir crear recursos y asociarlos a sedes, conservando historial.

Referencias históricas de backlog: `ADMIN-003`.

**Estado vigente (2026-10-26):** Implementado para el alcance MVP3.

**Implementación identificada:** Inventario administrable con sedes, recursos, filtros y preservación de historial.

**Evidencia y trazabilidad:** `frontend/src/views/ResourcesView.vue`; `backend/internal/repositories/resources_create_repository.go`

**Seguimiento:** ADMIN-003 cerrado; traslado entre sedes fuera de MVP3.

### RF-025 - Configuracion administrativa de politicas de reserva

El sistema debe permitir exclusivamente a usuarios con rol administrador publicar nuevas versiones del periodo de reserva, el plazo previo de confirmacion y los recursos sujetos a confirmacion grupal. Los cambios normales son prospectivos y cada solicitud conserva la version vigente al crearse. Excepcionalmente, el administrador puede migrar solicitudes futuras `PENDING` o `CONFIRMED` seleccionadas a otra version mediante simulacion, motivo obligatorio, confirmacion, aplicacion atomica y auditoria; la operacion no edita versiones historicas ni cancela solicitudes implicitamente.

Estado documentado previamente (histórico): APROBADO e IMPLEMENTADO para el bloque funcional MVP2. La publicacion prospectiva, permisos, idempotencia, minimo por recurso, snapshot, participantes, transiciones, solapes personales y expiracion cuentan con evidencia local. Permanecen pendientes la compuerta 14D completa, la prueba manual y el ambiente online; la interfaz administrativa y las correcciones excepcionales no pertenecen a este cierre.

Referencias históricas de backlog: `ADMIN-006`.

Definición incremental de MVP3 (3C):

- La clasificación de recursos grupales y sus mínimos ya forma parte
  del modelo versionado y del flujo de reservas implementado en MVP2.
- En 3C el administrador podrá configurar explícitamente esos recursos
  y sus mínimos al publicar una nueva versión.
- El mínimo por recurso será la referencia operativa para solicitudes
  grupales; el parámetro global se conservará temporalmente por
  compatibilidad histórica.
- La nueva versión se aplicará prospectivamente. Reservas anteriores
  conservan su política y snapshots.
- La corrección excepcional con simulación, motivo, confirmación,
  atomicidad, auditoría y reversión sigue siendo un requisito aprobado,
  pero se implementará en un incremento posterior a 3C.

<!-- EV-012 2026-10-13 -->

**Estado vigente (2026-10-26):** Implementado para ADMIN-006; parcial frente al requisito completo.

**Implementación identificada:** Políticas prospectivas versionadas, idempotencia, edición administrativa e historial.

**Evidencia y trazabilidad:** `backend/internal/repositories/reservation_policies_repository.go`; `frontend/src/views/PolicyAdministrationView.vue`

**Seguimiento:** La corrección excepcional de reservas existentes continúa pendiente de planificación.

### RF-026 - Administracion integral de sedes

El sistema debe permitir al administrador crear, consultar,
editar, activar y desactivar las sedes institucionales.

Criterios de aceptacion:

- Nombre y pais obligatorios.
- Direccion, comuna, ciudad, region y coordenadas opcionales.
- Nombre de sede unico conforme al esquema vigente.
- Coordenadas limitadas a rangos validos.
- Una sede agrupa multiples recursos.
- Se muestran recursos asociados y estado operativo.
- No se desactiva una sede con recursos activos sin resolverlos.
- No se eliminan fisicamente sedes en MVP3.
- Las reservas e historiales existentes se conservan.
- Solo administradores modifican sedes.

Estado: APROBADO por EV-012 el 2026-10-13.
Implementación registrada anteriormente: PENDIENTE.
Dependencias: RF-014, RF-024, ADMIN-003.

**Estado vigente (2026-10-26):** Implementado y validado para el alcance MVP3.

**Implementación identificada:** Administración de sedes, ubicación, coordenadas y desactivación controlada.

**Evidencia y trazabilidad:** `backend/internal/handlers/venues_handlers.go`; `backend/internal/repositories/venues_repository.go`; `frontend/src/components/resources/VenuesAdminPanel.vue`

**Seguimiento:** ADMIN-003 cerrado; búsqueda avanzada de sedes diferida a V2.

## Requisitos no funcionales

### RNF-001 - Seguridad

Las rutas protegidas deben requerir token Bearer o mecanismo local de desarrollo explicitamente habilitado.

**Estado vigente (2026-10-26):** Implementado para el prototipo; endurecimiento pendiente.

**Implementación y evidencia:** Autenticación JWT, rutas protegidas y autorización administrativa.

**Trazabilidad técnica:** `backend/internal/middleware/auth_middleware.go`; `backend/internal/routes/admin_authorization_http_test.go`

**Seguimiento:** SEC-005, SEC-006 y comprobaciones de seguridad productiva.

### RNF-002 - No exposicion de secretos

Las credenciales reales no deben guardarse en archivos versionados ni en documentacion.

**Estado vigente (2026-10-26):** Por verificar integralmente.

**Implementación y evidencia:** Configuración privada documentada fuera del repositorio.

**Trazabilidad técnica:** `README.md`; `docs/10-guia-redeploy.md`

**Seguimiento:** Ejecutar revisión de secretos en Git, configuración y logs.

### RNF-003 - Trazabilidad

Las tareas completadas, parciales o nuevas deben reflejarse en `docs/07-backlog.md`.

**Estado vigente (2026-10-26):** Cumplimiento parcial.

**Implementación y evidencia:** Backlog y documentos de cierre registran avances y pruebas.

**Trazabilidad técnica:** `docs/07-backlog.md`; `docs/mvp3/cierre-mvp3-prototipo.md`

**Seguimiento:** Sincronizar estados históricos con el cierre MVP3.

### RNF-004 - Documentacion viva

La carpeta `docs/` debe actualizarse cuando cambien arquitectura, instalacion, base de datos, frontend, backend, flujo de reservas, requisitos o backlog.

**Estado vigente (2026-10-26):** Cumplimiento parcial.

**Implementación y evidencia:** Existen documentos técnicos, pero algunos conservan información antigua.

**Trazabilidad técnica:** `docs/02-arquitectura.md`; `docs/03-base-de-datos.md`; `docs/10-guia-redeploy.md`

**Seguimiento:** Actualizar arquitectura, instalación, base de datos y despliegue.

### RNF-005 - Validacion tecnica minima

Antes de cerrar cambios relevantes se debe ejecutar, cuando corresponda:

```bash
cd backend
go test ./...

cd ../frontend
npm test
npm run build
```

Las pruebas actuales cubren reglas temporales y de agenda, pero no constituyen cobertura completa de componentes, permisos ni integracion. Si no se pueden ejecutar pruebas, debe registrarse el motivo y no declararse el requisito como verificado.

**Estado vigente (2026-10-26):** Implementado como proceso de validación; cobertura integral pendiente.

**Implementación y evidencia:** Pruebas backend, frontend, PostgreSQL y compilación Vite documentadas.

**Trazabilidad técnica:** `docs/mvp3/cierre-mvp3-prototipo.md`; `backend/internal/services/group_reservation_integration_test.go`; `frontend/src/utils/reportMetrics.test.js`

**Seguimiento:** Añadir CI y ampliar pruebas de seguridad, accesibilidad e integración.

### RNF-006 - Compatibilidad local

El proyecto debe poder ejecutarse localmente con backend Go/Fiber, frontend Vue/Vite y PostgreSQL 16 configurado. El provisionamiento automatico MVP2 debe aplicar las migraciones PostgreSQL requeridas antes de habilitar `MVP_SCOPE=mvp2`.

**Estado vigente (2026-10-26):** Validado para el alcance de instalación local MVP2.

**Implementación y evidencia:** Provisionamiento PostgreSQL y pruebas sobre entornos efímeros.

**Trazabilidad técnica:** `infra/local/quadlet/install.sh`; `infra/local/quadlet/verify-mvp2-ephemeral.sh`

**Seguimiento:** Actualizar instalación reproducible hasta las migraciones MVP3.

### RNF-007 - Usabilidad

La interfaz debe mostrar estados de carga, error y vacio en flujos principales.

**Estado vigente (2026-10-26):** Implementado parcialmente; validación transversal pendiente.

**Implementación y evidencia:** Las vistas utilizan estados operativos de carga y error.

**Trazabilidad técnica:** `frontend/src/views/ReportsView.vue`; `frontend/src/stores/reservations.js`

**Seguimiento:** Verificar estados vacíos, errores y navegación en flujos críticos.

### RNF-008 - Minimizacion de datos visibles

La disponibilidad para usuarios normales debe exponer solo la informacion necesaria para identificar horarios ocupados, evitando entregar datos internos o personales de reservas ajenas.

**Estado vigente (2026-10-26):** Implementado en disponibilidad; auditoría global pendiente.

**Implementación y evidencia:** El backend oculta identificadores y datos personales ajenos.

**Trazabilidad técnica:** `backend/internal/handlers/reservations_handlers.go`; `backend/internal/services/reservations_service.go`

**Seguimiento:** Ampliar pruebas de privacidad y minimizar respuestas no necesarias.

### RNF-009 - Accesibilidad basica

Los flujos criticos deben poder operarse con controles claros, mensajes de estado comprensibles y modales con comportamiento accesible basico.

**Estado vigente (2026-10-26):** Cumplimiento parcial.

**Implementación y evidencia:** Existen controles y etiquetas accesibles, pero faltan pruebas de teclado y modales.

**Trazabilidad técnica:** `frontend/src/components/forms/ReservationForm.vue`; `frontend/src/components/availability/CalendarToolbar.vue`

**Seguimiento:** Cerrar BACK-022 y validar foco, Escape y lectores de pantalla.

### RNF-010 - Consistencia temporal

Frontend, backend y base de datos deben aplicar una unica zona horaria de negocio y un contrato explicito de serializacion. Las reglas de pasado, en curso, futuro y cancelacion no deben depender de la zona local del servidor.

Estado documentado previamente (histórico): Implementado y verificado localmente; pendiente validacion integrada/online.

**Estado vigente (2026-10-26):** Implementado y verificado localmente; validación final pendiente.

**Implementación y evidencia:** Reloj de negocio y utilidades de serialización temporal.

**Trazabilidad técnica:** `backend/internal/businessclock/clock.go`; `backend/internal/businessclock/clock_test.go`; `frontend/src/utils/reservationTime.js`

**Seguimiento:** Completar regresión integrada y pruebas en entorno objetivo.

### RNF-011 - Errores publicos seguros

Las respuestas HTTP deben entregar codigos y mensajes publicos estables sin incluir errores internos de SQL, JWT, JWKS o configuracion. El diagnostico tecnico debe permanecer en logs sanitizados.

Referencia de seguimiento: `SEC-005`.

**Estado vigente (2026-10-26):** No cumple íntegramente; brecha confirmada.

**Implementación y evidencia:** Algunos handlers incluyen err.Error() en respuestas HTTP.

**Trazabilidad técnica:** `backend/internal/handlers/reservations_handlers.go`; `backend/internal/handlers/activities_handlers.go`

**Seguimiento:** SEC-005: eliminar detalles internos de respuestas públicas y añadir pruebas.

### RNF-012 - Responsive transversal

El shell, header, sidebar, dropdowns y modales deben mantenerse operables sin superposiciones desde 320 px y no dejar controles ocultos dentro del orden de tabulacion.

Referencia de seguimiento: `BACK-021`, `BACK-022`.

**Estado vigente (2026-10-26):** Cumplimiento parcial.

**Implementación y evidencia:** Diseño responsive existente, pero falta cerrar criterios de 320 px y teclado.

**Trazabilidad técnica:** `frontend/src/components/layout/Sidebar.vue`; `frontend/src/components/layout/NotificationBell.vue`

**Seguimiento:** Cerrar BACK-021 y BACK-022 con evidencia visual y funcional.

## Historias de usuario

### HU-001 - Iniciar sesion

Como usuario institucional, quiero iniciar sesion con mi cuenta Microsoft para acceder al sistema de reservas.

Criterios de aceptacion:

- El usuario puede entrar con Entra ID.
- Las rutas internas requieren autenticacion.
- Si la sesion no existe, el usuario vuelve a `/login`.

### HU-002 - Completar RUT

Como usuario normal, quiero registrar mi RUT cuando el sistema lo solicite para poder crear reservas.

Criterios de aceptacion:

- Si no tengo RUT, el sistema me solicita registrarlo en un modal.
- El RUT se valida antes de guardar.
- Despues de guardar, puedo continuar al flujo solicitado.

### HU-003 - Consultar disponibilidad

Como usuario normal, quiero revisar la disponibilidad por fecha para elegir un horario libre.

Criterios de aceptacion:

- La grilla muestra recursos reales.
- La grilla muestra reservas del dia seleccionado.
- Cambiar de fecha actualiza la informacion.

### HU-004 - Crear reserva

Como usuario normal, quiero crear una reserva sobre un recurso disponible para asegurar mi horario.

Criterios de aceptacion:

- La reserva se asocia a mi usuario autenticado.
- El servidor asigna el estado segun la politica del recurso.
- Fecha, hora y duracion cumplen la jornada institucional.
- El sistema rechaza conflictos de horario.
- Si el recurso es grupal, la solicitud permanece pendiente hasta reunir 10 participantes confirmados.
- Si el recurso es `OPEN_USE`, no se exige confirmacion grupal.
- La restriccion semanal se aplica antes de aceptar una nueva solicitud.
- El sistema muestra exito o error.

### HU-005 - Cancelar reserva propia

Como usuario normal, quiero cancelar una reserva propia cuando ya no la usare.

Criterios de aceptacion:

- Solo puedo cancelar reservas propias.
- Solo puedo cancelar una reserva en estado cancelable y no finalizada.
- El sistema pide confirmacion fuerte antes de ejecutar la cancelacion.
- La lista o grilla se actualiza luego de cancelar.

### HU-006 - Revisar mis reservas

Como usuario normal, quiero ver mis reservas actuales y su detalle para hacer seguimiento.

Criterios de aceptacion:

- Se muestran reservas reales del usuario autenticado.
- Se muestra recurso, fecha, hora, duracion, actividad y estado.
- Existe acceso al detalle.

### HU-006A - Revisar reservas como administrador

Como administrador, quiero ver todas las reservas del sistema para apoyar la supervision operacional.

Criterios de aceptacion:

- Se muestran reservas reales de todos los usuarios.
- Usuarios normales siguen viendo solo sus propias reservas.
- Existe acceso al detalle de cada reserva visible.

### HU-007 - Revisar historial

Como usuario normal, quiero revisar reservas pasadas o canceladas para consultar mi actividad anterior.

Criterios de aceptacion:

- Se muestran reservas historicas reales.
- Existen estados de carga, error y vacio.
- Se puede filtrar por fecha y estado dentro del historial visible para el rol.

### HU-007A - Revisar historial global como administrador

Como administrador, quiero revisar todo el historial de reservas para auditar actividad del sistema.

Criterios de aceptacion:

- Se muestran reservas historicas reales de todos los usuarios.
- Usuarios normales siguen viendo solo su historial.
- Los filtros de estado y fecha funcionan sobre el conjunto visible segun rol.

### HU-008 - Revisar notificaciones

Como usuario normal, quiero ver notificaciones asociadas a mis reservas o eventos relevantes.

Criterios de aceptacion:

- La campana muestra contador real.
- No se consulta la API si no hay sesion.
- En una iteracion futura, podre marcar notificaciones como leidas.

### HU-009 - Acceder al panel administrador

Como administrador, quiero acceder a un panel con resumen de reservas, recursos y actividad para monitorear el sistema.

Criterios de aceptacion:

- Solo administradores ven accesos administrativos.
- Usuarios normales no ven la seccion de administracion.
- El panel carga datos reales.

### HU-010 - Gestionar usuarios

Como administrador, quiero listar usuarios y bloquear o desbloquear cuentas para controlar el acceso al sistema.

Criterios de aceptacion:

- Puedo ver usuarios reales.
- En una iteracion futura, podre bloquear o desbloquear usuarios.
- El sistema no debe permitir autobloqueo accidental.

### HU-011 - Gestionar recursos

Como administrador, quiero mantener el catalogo de recursos deportivos para
que las reservas usen informacion vigente.

Criterios de aceptacion:

- Puedo actualizar la imagen de un recurso existente.
- Puedo modificar modo de reserva, capacidad y estado activo/inactivo.
- Solo un administrador puede ejecutar estas modificaciones.
- Los modos de reserva se limitan a los valores soportados por el dominio.
- Un recurso con confirmacion grupal vigente debe permanecer `RESERVABLE`.
- Su capacidad no puede quedar bajo el minimo grupal vigente.
- Las reservas ya creadas conservan sus snapshots historicos.
- Puedo crear recursos y editar nombre, tipo y sede.
- Puedo seleccionar una sede registrada.
- Un recurso nuevo no entra automaticamente en la politica vigente.
- Los cambios no destruyen historiales ni snapshots.
- La eliminacion fisica queda fuera de MVP3.
- La implementacion completa de estos criterios sigue pendiente.
- No se eliminan recursos con historial sin criterio definido.

### HU-012 - Crear bloqueos

Como administrador, quiero bloquear horarios por mantencion, cierre o eventos para evitar reservas no disponibles.

Criterios de aceptacion:

- Puedo seleccionar recurso, fecha y rango horario.
- El sistema valida solapamientos.
- Los bloqueos aparecen en el calendario.

### HU-013 - Consultar reportes

Como administrador, quiero revisar reportes de uso para apoyar decisiones operativas.

Criterios de aceptacion:

- Se muestran indicadores de uso.
- Se muestran horas punta.
- La definicion de reserva activa excluye finalizadas, canceladas, rechazadas y expiradas.
- Las horas se presentan sin redondeos que alteren el total real.
- En una iteracion futura, se integraran infracciones y vistas SQL dedicadas.

### HU-014 - Inscribirse en taller deportivo

Como usuario normal, quiero revisar talleres deportivos disponibles e inscribirme en uno con cupos para participar en actividades institucionales.

Criterios de aceptacion:

- Se muestran talleres activos con dia, horario, lugar, capacidad e inscritos.
- Puedo buscar talleres por nombre, dia o lugar.
- Si ya estoy inscrito, el sistema lo indica y no duplica la inscripcion.
- Si no tengo RUT, el backend rechaza la inscripcion.
- Si no hay cupos, el sistema impide la inscripcion.

### HU-015 - Reunir participantes para una reserva grupal

Como solicitante de un recurso grupal, quiero que los integrantes confirmen su participacion para que la solicitud se confirme al alcanzar el minimo institucional.

Criterios de aceptacion:

- La solicitud comienza en estado `PENDING`.
- El solicitante cuenta una vez y todos los participantes se identifican mediante una cuenta.
- La solicitud `PENDING` bloquea el horario para usos incompatibles.
- Antes de alcanzar el minimo snapshot permanece `PENDING + PENDING_MINIMUM`.
- La confirmacion que alcanza el minimo cambia automaticamente la solicitud a `CONFIRMED + HEALTHY` si las demas reglas siguen vigentes.
- Si una reserva ya confirmada pierde participantes y baja del minimo, conserva `CONFIRMED` y cambia a `AT_RISK`.
- Si recupera el minimo antes del vencimiento, conserva `CONFIRMED` y vuelve a `HEALTHY`.
- Las altas y retiradas se aceptan antes del deadline de una hora y se cierran al alcanzarlo.
- Si al deadline sigue `PENDING` bajo el minimo, cambia a `CANCELLED`, libera el horario y la oportunidad semanal.
- Si ya fue confirmada y luego baja del minimo, conserva `CONFIRMED + AT_RISK`; no se cancela por `MINIMUM_NOT_MET`.
- El cliente no puede forzar el estado ni el conteo.

### HU-016 - Resolver conflictos institucionales

Como administrador, quiero conocer y resolver los conflictos de una actividad institucional para aplicar la prioridad institucional sin impedir usos compartidos validos.

Criterios de aceptacion vigentes no controvertidos:

- El sistema detecta todas las ocupaciones involucradas en un conflicto.
- El administrador puede consultar el conflicto y sus elementos.
- Los conflictos entre actividades permiten decisiones administrativas explicitas.
- La resolucion puede conservar, autorizar coexistencia, cancelar o reprogramar cuando corresponda.
- La decision y sus efectos quedan trazables.
- El usuario afectado por una cancelacion debe recibir una notificacion sin exponer informacion de terceros.

Evolucion propuesta `EV-010` (pendiente de ratificacion institucional):

- RF-023/v1 establecia cancelacion automatica ante prioridad institucional.
- RF-023/v2 propone detectar conflictos sin cancelar reservas automaticamente.
- El administrador decide expresamente sobre las ocupaciones afectadas.
- La decision requiere justificacion y confirmacion.
- La resolucion administrativa ya existe en el sistema.
- La sustitucion normativa de v1 por v2 no se considera formalizada hasta su ratificacion institucional.

### HU-017 - Mantener inventario oficial

Como administrador, quiero modificar el inventario oficial de recursos para mantener el sistema alineado con los espacios disponibles.

Criterios de aceptacion:

- Los ocho recursos actuales forman la linea base.
- El administrador puede crear recursos adicionales al inventario inicial.
- Puede editar datos, modo, capacidad, imagen, sede y estado operativo.
- Los cambios se reflejan en catalogo y disponibilidad.
- El historial no se pierde por desactivar o modificar un recurso.

### HU-018 - Configurar politicas de reserva

Como administrador, quiero modificar el periodo de reserva, el plazo previo y los recursos sujetos a confirmacion para ajustar las reglas institucionales sin otorgar esta facultad a usuarios normales.

Criterios de aceptacion:

- Solo usuarios con rol administrador pueden modificar las politicas.
- Los valores iniciales son siete dias y una hora antes del inicio.
- Cancha 1, Cancha 2 y Sala Multiuso comienzan sujetas a
  confirmación grupal con mínimo inicial 10 por recurso.
- Puedo seleccionar recursos permitidos para reservar.
- Puedo definir por recurso si requiere confirmación grupal
  y su mínimo de participantes.
- El sistema rechaza mínimos superiores a la capacidad del recurso
  o incompatibles con su modo de reserva.
- Puedo configurar jornada, intervalo, frecuencia y duraciones
  permitidas.
- Puedo revisar el historial y comparar la política vigente
  con los nuevos valores antes de confirmar.
- El cambio informa desde cuando rige.
- Los cambios normales se aplican prospectivamente y las solicitudes conservan su version.
- Una correccion excepcional exige seleccionar solicitudes futuras activas, previsualizar el resultado, informar un motivo y confirmar la aplicacion.
- Si una solicitud es incompatible, el lote completo se rechaza sin cancelaciones implicitas.
- La correccion y cualquier reversion quedan auditadas.

### HU-019 - Administrar sedes institucionales

Como administrador, quiero crear y mantener las sedes deportivas
para organizar los recursos por recinto y mantener actualizadas
sus ubicaciones.

Criterios de aceptacion:

- Puedo crear y editar sedes con sus datos institucionales.
- Puedo activar y desactivar sedes.
- Puedo consultar y filtrar sedes y sus recursos asociados.
- No puedo desactivar una sede con recursos activos sin resolverlos.
- Ninguna operacion elimina reservas historicas.
- Los usuarios no administradores no pueden modificar sedes.

Estado: APROBADA por EV-012; funcionalidad principal implementada y validada para el prototipo MVP3 en `ADMIN-003`. Nota de alcance: la busqueda y filtrado de sedes se difirio a V2.0; el criterio de aceptacion correspondiente sigue pendiente fuera de MVP3.

## Casos de uso

### CU-001 - Autenticarse en el sistema

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario tiene cuenta valida.
- La configuracion de Entra ID o modo local esta disponible.

Flujo principal:

1. El usuario abre la aplicacion.
2. El sistema muestra login.
3. El usuario inicia sesion.
4. El sistema obtiene datos del usuario local.
5. El sistema muestra la pantalla inicial segun permisos y estado del perfil.

Postcondiciones:

- El usuario queda autenticado.
- El sistema puede consumir rutas protegidas.

### CU-002 - Registrar RUT obligatorio

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario no tiene RUT registrado.

Flujo principal:

1. El usuario intenta entrar a una ruta interna.
2. El sistema detecta que falta RUT.
3. El sistema muestra un modal obligatorio para registrar RUT.
4. El usuario ingresa RUT en el modal.
5. El sistema valida y guarda el RUT.
6. El sistema permite continuar en la aplicacion.

Postcondiciones:

- El usuario queda habilitado para crear reservas.

### CU-003 - Crear una reserva

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene RUT valido.
- Existen recursos disponibles.

Flujo principal:

1. El usuario abre disponibilidad.
2. El usuario selecciona fecha, recurso, hora, duracion y actividad.
3. El usuario envia la solicitud.
4. El backend valida usuario, recurso, restriccion semanal, zona horaria, jornada, duracion y conflicto horario.
5. El sistema asigna el estado segun la politica del recurso.
6. Si el recurso requiere grupo, la solicitud queda `PENDING` hasta alcanzar 10 participantes confirmados.
7. Si el recurso no requiere confirmacion grupal, se aplica directamente su comportamiento aprobado.
8. La disponibilidad se actualiza.

Postcondiciones:

- La solicitud queda asociada al usuario autenticado con el estado controlado por servidor.

Flujos alternativos:

- Si hay conflicto horario, el sistema rechaza la reserva.
- Si el horario o duracion estan fuera de regla, el sistema rechaza la reserva con mensaje accionable.
- Si un cliente intenta enviar un estado, no puede modificar el estado inicial ni evitar validaciones.
- Si falta RUT, el sistema bloquea la creacion.
- Si no se cumple la restriccion semanal, el sistema rechaza la solicitud y comunica la proxima fecha permitida.
- Si falta un dato obligatorio, el frontend muestra error.

### CU-004 - Cancelar una reserva

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario esta autenticado.
- Existe una reserva cancelable.

Flujo principal:

1. El usuario abre el detalle de una reserva.
2. El usuario solicita cancelar.
3. El sistema pide confirmacion.
4. El backend valida permisos, estado cancelable y termino segun la zona institucional.
5. El sistema cambia el estado a cancelada.
6. La interfaz actualiza la lista o grilla.

Postcondiciones:

- La reserva queda cancelada.

Flujos alternativos:

- Si el usuario normal intenta cancelar reserva ajena, el sistema rechaza la accion.
- Si la reserva esta cancelada, rechazada, expirada o finalizada, el sistema rechaza la accion.

### CU-005 - Consultar mis reservas

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.

Flujo principal:

1. El usuario abre Mis Reservas.
2. El frontend solicita reservas del usuario autenticado.
3. El sistema muestra listado y estados.
4. El usuario abre un detalle si lo necesita.

Postcondiciones:

- El usuario visualiza sus reservas vigentes.

### CU-006 - Revisar panel administrador

Actor principal: Administrador.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene rol administrador.

Flujo principal:

1. El administrador abre el panel.
2. El sistema carga recursos y reservas.
3. El sistema muestra indicadores y accesos administrativos.

Postcondiciones:

- El administrador visualiza informacion operacional.

Flujos alternativos:

- Si un usuario normal intenta acceder, el sistema bloquea la ruta y no muestra el menu administrativo.

### CU-007 - Consultar notificaciones

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario esta autenticado.

Flujo principal:

1. El usuario abre la aplicacion.
2. La campana consulta notificaciones del usuario.
3. El sistema muestra contador y listado.

Postcondiciones:

- El usuario visualiza sus notificaciones.

Flujos alternativos:

- Si no hay sesion, la campana no consulta la API.

### CU-008 - Inscribirse en taller

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene RUT registrado.
- Existe al menos un taller activo con cupos.

Flujo principal:

1. El usuario abre Talleres.
2. El frontend solicita talleres activos.
3. El sistema muestra cupos, horario, lugar y estado de inscripcion.
4. El usuario solicita inscribirse.
5. El backend valida usuario, RUT, taller activo, cupo disponible e inscripcion duplicada.
6. El sistema registra la inscripcion y actualiza el taller.

Postcondiciones:

- El usuario queda inscrito en el taller seleccionado.

Flujos alternativos:

- Si falta RUT, el backend rechaza la inscripcion.
- Si el taller esta lleno, el sistema muestra error.
- Si el usuario ya estaba inscrito, el sistema evita duplicar la inscripcion.

### CU-009 - Confirmar participantes de una solicitud grupal

Actor principal: Participante.

Precondiciones:

- Existe una solicitud `PENDING` sobre un recurso grupal.
- El participante tiene cuenta, puede ser identificado y aun no ha confirmado; el solicitante ya cuenta una vez.
- La solicitud pendiente bloquea el horario.

Flujo principal:

1. El participante revisa la solicitud correspondiente.
2. El participante confirma su participacion.
3. El sistema valida identidad y duplicado.
4. El sistema actualiza el conteo de confirmaciones vigentes.
5. Al alcanzar el minimo snapshot, vuelve a validar las demas reglas y cambia la solicitud a `CONFIRMED + HEALTHY`.
6. Antes del deadline configurable una persona no owner puede retirar su confirmacion; al alcanzar el limite los cambios quedan cerrados.
7. Si la reserva ya habia alcanzado el minimo y el conteo baja de dicho valor, conserva `CONFIRMED` y cambia su condicion grupal a `AT_RISK`.
8. Si recupera el minimo antes del vencimiento, conserva `CONFIRMED` y vuelve a condicion `HEALTHY`.
9. Si la solicitud llega al deadline todavia `PENDING` bajo el minimo, el housekeeping la cambia a `CANCELLED` con `MINIMUM_NOT_MET`, libera el horario y la oportunidad semanal.

Postcondiciones:

- La confirmacion queda registrada una sola vez.
- La primera confirmacion de la reserva ocurre al alcanzar el minimo; una perdida posterior se representa mediante `AT_RISK` sin borrar dicha transicion.

Flujos alternativos:

- Si nunca ha alcanzado el minimo snapshot, permanece `PENDING + PENDING_MINIMUM` antes del deadline.
- Una alta o retirada en el instante exacto del deadline se rechaza sin cambiar el conteo.
- Si la solicitud deja de ser valida antes de alcanzar el minimo, no se confirma automaticamente.

### CU-010 - Resolver conflicto de actividad institucional

Actor principal: Administrador.

Precondiciones:

- Existe una actividad institucional nueva o modificada.
- El mismo recurso presenta al menos una ocupacion solapada.

Flujo base vigente:

1. El sistema detecta todas las ocupaciones conectadas por el conflicto.
2. Registra un conflicto con sus elementos afectados.
3. El administrador consulta el conflicto y su estado.
4. Para los elementos que admiten resolucion administrativa puede utilizar `KEEP`, `ALLOW`, `CANCEL` o `RESCHEDULE` segun las reglas aplicables.
5. Cada decision exige trazabilidad y actualiza los efectos correspondientes.
6. El conflicto se cierra cuando no quedan elementos pendientes.

Evolucion propuesta `EV-010` (pendiente de ratificacion institucional):

- RF-023/v1 contempla cancelacion automatica por prioridad institucional.
- RF-023/v2 propone registrar el conflicto sin modificar automaticamente las reservas.
- El administrador revisa y resuelve expresamente las ocupaciones afectadas.
- Cada decision requiere justificacion y confirmacion.
- Solo una decision administrativa de cancelacion puede cancelar la reserva dentro de este flujo.
- Las restricciones duras de disponibilidad mantienen sus validaciones.
- La adopcion definitiva de RF-023/v2 requiere ratificacion institucional.

Postcondiciones:

- Las decisiones aplicadas quedan persistidas y trazables.
- Una cancelacion actualiza el estado de la ocupacion correspondiente.
- Una coexistencia autorizada permanece explicitamente registrada.
- Una reprogramacion debe volver a participar en deteccion de conflictos.

Flujos alternativos:

- Un plan incompatible debe rechazarse sin efectos parciales.
- Puede autorizarse coexistencia cuando institucionalmente corresponda.
- Una reprogramacion que genere un nuevo solapamiento origina un nuevo conflicto.

### CU-011 - Configurar politicas de reserva

Actor principal: Administrador.

Precondiciones:

- El usuario tiene una cuenta vigente con rol administrador.

Flujo principal:

1. El administrador consulta las politicas institucionales vigentes.
2. Modifica ventana, frecuencia, plazo de confirmación, jornada,
   duraciones y recursos permitidos.
3. Configura qué recursos requieren confirmación grupal y el mínimo
   de participantes de cada uno.
4. El sistema valida su rol, valores, recursos y compatibilidad
   entre mínimo grupal y capacidad.
5. El sistema muestra los valores anteriores, los nuevos y desde
   cuándo regirá la versión.
6. El administrador confirma la publicación de la nueva política.

Postcondiciones:

- La politica queda disponible para las solicitudes a las que corresponda.
- El cambio queda atribuible al administrador.

Flujos alternativos:

- Un usuario normal no puede modificar las politicas.
- Un valor invalido se rechaza sin alterar la politica vigente.

### CU-SED-01 - Crear sede

Actor: Administrador.
Flujo: ingresar nombre, pais y ubicacion opcional; validar;
registrar la sede; confirmar su creacion.
Alternativas: rechazar nombre duplicado o coordenadas invalidas.

### CU-SED-02 - Editar sede

Actor: Administrador.
Flujo: seleccionar sede; modificar datos; validar; guardar.
Alternativas: rechazar datos invalidos o duplicados.

### CU-SED-03 - Desactivar o activar sede

Actor: Administrador.
Flujo: seleccionar sede; solicitar cambio; comprobar
recursos asociados; confirmar y guardar.
Alternativas: rechazar desactivacion con recursos activos;
no ejecutar cambios implicitos en cascada.

### CU-SED-04 - Consultar sedes

Actor: Administrador.
Flujo: listar sedes, buscar o filtrar y consultar
recursos asociados sin modificar datos.

### CU-REC-01 - Crear recurso

Actor: Administrador.
Flujo: ingresar nombre, tipo, sede, modo, capacidad
e imagen opcional; validar; crear.
Alternativas: rechazar sede no valida, duplicados y datos
incompatibles. No modificar la politica vigente.

### CU-REC-02 - Editar recurso

Actor: Administrador.
Flujo: seleccionar recurso; editar datos y configuracion;
validar reglas grupales y guardar.
Alternativas: rechazar modo o capacidad incompatible.

### CU-REC-03 - Activar o desactivar recurso

Actor: Administrador.
Flujo: seleccionar recurso, cambiar estado y guardar.
Postcondicion: el historial permanece y la disponibilidad
refleja el estado vigente.

### CU-REC-04 - Cambiar sede de recurso

Actor: Administrador.
Flujo: elegir una nueva sede activa; validar las
reservas futuras relacionadas; confirmar y guardar.
Alternativas: rechazar traslados incompatibles.

### CU-REC-05 - Consultar y filtrar recursos

Actor: Administrador.
Flujo: buscar por nombre y filtrar por sede, tipo,
modo y estado; consultar la configuracion vigente.

## Matriz de trazabilidad inicial

| Requisito | Historias relacionadas | Backlog relacionado |
| --- | --- | --- |
| RF-001 | HU-001 | AUTH-001, API-003 |
| RF-002 | HU-001 | AUTH-002 |
| RF-003 | HU-002 | AUTH-004 |
| RF-004 | HU-002, HU-004 | AUTH-004 |
| RF-005 | HU-003 | RES-003, RES-004 |
| RF-006 | HU-004, HU-015 | RES-001, RES-002, RES-005, RES-008, RES-009, RES-010, RES-011, RES-012 |
| RF-007 | HU-004, HU-012, HU-016 | RES-004, RES-010, ADMIN-004, ADMIN-005, QA-001 |
| RF-008 | HU-005 | RES-007, RES-009, RES-010, API-003 |
| RF-009 | HU-006, HU-006A | RES-006 |
| RF-010 | HU-007, HU-007A | UI-003 |
| RF-011 | HU-003, HU-011 | UI-001, API-001 |
| RF-012 | HU-009 | ADMIN-001 |
| RF-013 | HU-010 | ADMIN-002 |
| RF-014 | HU-011, HU-017 | ADMIN-003 |
| RF-015 | HU-012 | ADMIN-004 |
| RF-016 | HU-012 | ADMIN-005 |
| RF-017 | HU-008 | NOTIF-001 |
| RF-018 | HU-013 | REP-001, REP-002, REP-003 |
| RF-019 | HU-014 | UI-006 |
| RF-020 | HU-004 | RES-012 |
| RF-021 | HU-015 | RES-008 |
| RF-022 | HU-004, HU-015 | RES-008, RES-010 |
| RF-023 | HU-016 | ADMIN-005, RES-004, NOTIF-001 |
| RF-024 | HU-011, HU-017 | ADMIN-003 |
| RF-025 | HU-018 | ADMIN-006 |
| RF-026 | HU-019 | ADMIN-003, API-001 |
| RNF-008 | HU-003, HU-004 | API-004 |
| RNF-009 | HU-003, HU-004, HU-005 | UX-001, UX-002 |

## Relacion con MVPs

La agrupacion oficial de estos requisitos por incremento se mantiene en `docs/09-mvps-roadmap.md`.

Resumen:

- MVP 1: base tecnica funcional.
- MVP 2: flujo usuario completo.
- MVP 3: administracion institucional.
- MVP 4: entrega, calidad y soporte.

## Protocolo de mantenimiento

Este documento debe actualizarse cuando:

- Se agregue, quite o cambie una funcionalidad.
- Cambie el comportamiento de un actor.
- Se complete una tarea del backlog que afecte requisitos.
- Se agregue una pantalla, endpoint o regla de negocio.
- Se detecte un nuevo caso de uso o una excepcion relevante.

Cada cambio funcional debe mantener coherencia entre:

- `docs/07-backlog.md`
- `docs/08-requisitos-historias-casos-uso.md`
- `docs/09-mvps-roadmap.md`
- Documentos tecnicos afectados dentro de `docs/`
