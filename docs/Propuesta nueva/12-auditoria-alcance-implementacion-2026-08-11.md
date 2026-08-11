# Auditoría de alcance e implementación de Poli-REDI

**Fecha de corte:** 2026-08-11

**Commit auditado:** `939ba51` (`mvp2-participantes`)

**Tipo de revisión:** contraste documental, código, esquema, pruebas disponibles y estado de despliegue.

**Resultado:** MVP 1 demostrable localmente; MVP 2 implementado de forma parcial y todavía condicionado; MVP 3 parcial; MVP 4 pendiente o fuera de alcance según la función.

## 1. Criterio de la auditoría

Los estados de esta auditoría no mezclan alcance, implementación, prueba y despliegue:

- **Aprobado:** existe una decisión de alcance vigente.
- **Implementado observable:** existe comportamiento en código o esquema.
- **Verificado localmente:** existe una ejecución local satisfactoria y fechada.
- **Validado integradamente:** frontend, API, identidad y base desplegada fueron comprobados juntos.
- **Desplegado:** existe evidencia del artefacto o migración en el ambiente indicado.
- **Parcial:** existe una parte funcional, pero falta una condición relevante del requisito.
- **Pendiente:** no existe implementación suficiente o no hay evidencia para afirmarla.
- **Fuera de alcance:** se decidió no incorporarlo al prototipo actual.

La presencia de código no demuestra por sí sola aprobación, prueba ni despliegue. Del mismo modo, una prueba local no acredita Azure SQL, Microsoft Entra ID o la experiencia online.

## 2. Evidencia nueva de este corte

| Verificación | Resultado del 2026-08-11 | Uso permitido como evidencia |
|---|---|---|
| Estado Git | Árbol limpio en `939ba51` al cerrar la revisión | Identifica con precisión el código auditado. |
| Pruebas Node | **18 aprobadas** | Evidencia local nueva para contratos JavaScript cubiertos por ese conjunto. |
| Vitest completo | La ejecución comenzó, pero no produjo un resultado final antes del límite de la revisión | **No** constituye evidencia nueva de aprobación ni de fallo del conjunto completo. |
| Backend Go | No se obtuvo un resultado final independiente en este corte | **No** constituye evidencia nueva. La aprobación documentada del 2026-08-04 sigue siendo evidencia histórica. |
| Build frontend | No se obtuvo un resultado final independiente en este corte | **No** constituye evidencia nueva. La aprobación documentada del 2026-08-04 sigue siendo evidencia histórica. |
| Azure SQL / Entra / flujo online | No ejecutado | No validado integradamente. |

## 3. Estado por MVP

| MVP | Alcance resumido | Implementación observable | Estado de cierre | Condición pendiente principal |
|---|---|---|---|---|
| MVP 1 | Identidad, recursos, disponibilidad, reserva básica, datos de usuario, consulta, cancelación e historial básico | La base funcional existe. La disponibilidad visible todavía no consume el contrato unificado por rango | **Demostrable localmente, no validado integradamente** | Corregir integración de disponibilidad y repetir smoke test online. |
| MVP 2 | Política, frecuencia, reglas por recurso, reservas grupales, códigos, participantes, deadline y talleres | Flujo grupal y talleres son observables; integración frontend de disponibilidad y migraciones reales siguen pendientes | **Aprobable condicionado; no cerrado** | Migraciones `007`/`008`, concurrencia Azure, QA completo y coherencia frontend/backend. |
| MVP 3 | Administración institucional, prioridad, bloqueos gestionables, pantalla pública e historial institucional | Existen lecturas, estructuras y componentes parciales; no existe el flujo administrativo completo | **Parcial** | CRUD y casos de uso institucionales, prioridad, auditoría, pantalla pública e historial. La notificación específica de prioridad es una dependencia de este flujo. |
| MVP 4 | Calidad operativa, notificaciones core, reportes básicos, auditoría consultable, seguridad, soporte y despliegue | Indicadores y consulta básica de notificaciones existen; el resto está parcial o pendiente | **Pendiente y acotado** | Completar contratos, pruebas, QA, seguridad y despliegue sin incorporar campeonatos avanzados ni detección de abuso. |

## 4. Matriz funcional exhaustiva

| Área / RF / MVP | Alcance aprobado | Implementación observable en `939ba51` | Prueba local | Validación integrada / despliegue | Estado | Pendiente y evidencia principal |
|---|---|---|---|---|---|---|
| **Identidad, roles y RUT — RF-04 — MVP 1** | Autenticación institucional o de prueba; usuario y administrador; usuario normal debe registrar RUT para reservar o inscribirse; administrador exento en esas operaciones | Middleware resuelve identidad, rol y bloqueo. `PATCH /api/me/rut` actualiza RUT. Reservas y talleres aplican la excepción administrativa. Confirmar participación exige cuenta activa y RUT también al administrador | Existen pruebas de RUT, autenticación local y contratos de handlers. Las 18 Node se revalidaron, pero no cubren por sí solas toda la integración | Entra ID y ambiente online no se revalidaron | **Implementado localmente** | Ejecutar matriz usuario normal/admin/bloqueado/sin RUT online. Evidencia: `backend/internal/middleware/auth_middleware.go`, `handlers/me_handler.go`, `handlers/reservations_handlers.go`, `handlers/workshops_handlers.go` y `routes/routes.go`. |
| **Recursos — RF-01 — MVP 1** | Registrar y consultar el inventario deportivo aprobado | `GET /api/resources` y edición administrativa de imagen. El modelo contempla modo, capacidad y estado; no existe CRUD administrativo completo de recursos | Lectura utilizada por vistas y pruebas indirectas | Sin validación integrada reciente | **Parcial** | Registrar no está implementado como operación administrativa. Evidencia: `resources_handlers.go`, `resources_repository.go`, `ResourcesView.vue`. |
| **Disponibilidad unificada — RF-02 — MVP 1/2** | Mostrar por fecha reservas, talleres, actividades y bloqueos, sin revelar datos de terceros | Backend por rango agrega reservas, ocurrencias normalizadas de talleres, actividades y bloqueos; calcula `blocksResource` e `isCurrentUserConflict` y sanitiza por audiencia. El frontend principal llama el contrato legado sin `from/to`, reconstruye talleres y no recibe bloqueos | Pruebas backend de rango, repositorio y privacidad presentes. No hay nueva ejecución completa de Vitest en este corte | No validado online | **Parcial; discrepancia funcional P1** | `AvailabilitySection.vue` debe pedir rango y consumir la respuesta única. Evidencia: `availability_repository.go`, `reservations_service.go`, `reservations_handlers.go`, `AvailabilitySection.vue`, `reservations.js`. |
| **Crear reserva particular — RF-03 — MVP 1** | Usuario autenticado elegible crea una reserva futura dentro de política y sin conflicto | Contrato estricto, actor controlado por servidor, validación de fecha/hora/duración/ventana, recurso, frecuencia y conflictos. Estados no grupales se crean `CONFIRMED` | Pruebas unitarias, contratos y validadores existentes; sin nueva ejecución completa Go | SQL/Azure y flujo online no revalidados | **Implementado localmente, condicionado** | Completar prueba integrada y revisar que la UI solo ofrezca recursos/slots admitidos por política. Evidencia: `reservations_handlers.go`, `reservations_service.go`, `reservations_repository.go`, `reservationrules/`. |
| **Participantes y quórum — RF-05/RF-06 — MVP 2** | Owner incluido; mínimo configurable; participantes con cuenta; transición reversible antes del límite | Participantes transaccionales, owner confirmado, capacidad congelada, conteo, `PENDING↔CONFIRMED`, retiro y reconfirmación implementados | Pruebas de reglas, contratos y SQL mock existentes | Concurrencia real y migraciones Azure pendientes | **Implementado localmente; no cerrado** | Ejecutar prueba multisesión y carrera real. Evidencia: `participants_repository.go`, `participants_rules.go`, `participants_handlers.go`. |
| **Duración, jornada y ventana — RF-07 — MVP 2** | Política configurable; configuración inicial 08:00–22:00, slots de 15 minutos, duraciones 30–180 y ventana de 7 días | Backend aplica snapshot de política. Formulario consume parte de la política; timeline principal sigue recibiendo constantes. Frontend alinea slot respecto de apertura y backend respecto de medianoche | Pruebas locales de reglas existentes; 18 Node revalidadas incluyen reglas horarias | No validado online con políticas no predeterminadas | **Parcial por integración** | Unificar fórmula de alineación y pasar apertura, cierre e intervalo al calendario. Evidencia: `reservationrules/schedule.go`, `reservationRules.js`, `AvailabilitySection.vue`, `ResourceTimeline.vue`. |
| **Clasificación de reserva — RF-08 — MVP 1/3** | Diferenciar reserva, uso libre y programación institucional | Modelos y chips reconocen reserva, reserva grupal, taller, actividad institucional y bloqueo. La programación institucional carece de gestión completa | Pruebas del sistema de tipos/chips existentes; Vitest completo no revalidado | No validado online | **Parcial** | No confundir clasificación visual con administración institucional implementada. Evidencia: `availability_item.go`, `AvailabilityTypeChip.vue`, `availabilityType.js`. |
| **Prioridad institucional — RF-09 — MVP 3** | Actividad institucional desplaza reserva particular con aviso; entre actividades decide el administrador | No existe endpoint o flujo de alta que aplique la prioridad. Los controles de conflicto actuales previenen el solape, pero no ejecutan cancelación y notificación aprobadas | Sin prueba de caso de uso completo | No desplegado | **Pendiente** | Diseñar transacción, vista previa, auditoría, mensaje y rollback. Evidencia: ausencia en `routes.go`; lectura en `scheduled_activities_repository.go`; decisión en `06-flujos-y-reglas-de-negocio.md`. |
| **Cancelación propia básica — MVP 1** | El propietario puede cancelar una reserva activa permitida mientras no haya finalizado | El servicio acepta cancelación del propietario o administrador para estados activos y rechaza reservas finalizadas. No aplica todavía un cutoff configurable anterior al inicio | Pruebas de cancelación, permiso y estados existentes | No validado online | **Implementado localmente con regla temporal futura pendiente** | Si se mantiene el cutoff configurable aprobado para una evolución futura, debe incorporarse como contrato, política y prueba antes de declararlo implementado. Evidencia: `reservations_handlers.go`, `reservations_service.go`, `ReservationsView.vue`. |
| **Administración de reservas — RF-10 — MVP 3** | Aprobar, rechazar, modificar o cancelar administrativamente según permisos, motivo y reglas | El administrador reutiliza la cancelación básica; no existen aprobar, rechazar ni modificar administrativamente una reserva, ni cutoff configurable o motivo manual completo | Cobertura local solo de cancelación y permisos | No validado online | **Parcial** | No usar `PATCH /api/reservations/cancel` como evidencia del requisito completo. Diseñar estados, motivo, cutoff, auditoría y permisos. |
| **Frecuencia semanal — RF-11 — MVP 2** | Recursos que consumen oportunidad impiden otra solicitud hasta la próxima fecha; `OPEN_USE` no consume | Backend consulta la última reserva consumidora y aplica la política vigente de la solicitud anterior | Pruebas locales de cálculo y servicio existentes | Azure pendiente | **Implementado localmente** | Verificar límite de fecha y concurrencia en Azure. Evidencia: `GetLatestConsumingReservation`, `validateRequestFrequency`. |
| **Bloqueo de usuario y motivo — RF-12/RF-14 — MVP 3** | Administrador bloquea/desbloquea con motivo auditable | Usuario posee campos de bloqueo y motivo; middleware impide acceso. No existe ruta/UI de mutación administrativa ni historial completo de decisiones | Cobertura local del rechazo indirecto | No validado online | **Parcial** | Implementar CU administrativo, permisos, motivo obligatorio y auditoría. Evidencia: `models/user.go`, `auth_middleware.go`, `UsersView.vue`, ausencia en `routes.go`. |
| **Impedimento por bloqueo — RF-13 — MVP 3** | Cuenta bloqueada no puede operar | Middleware rechaza al usuario bloqueado antes de rutas protegidas | Pruebas locales existentes | Entra/online pendiente | **Implementado localmente** | Validar expiración/desbloqueo cuando exista gestión administrativa. |
| **Motivo de cancelación — RF-15 — MVP 3** | Cancelación administrativa exige motivo y lo conserva | Existe `cancellation_reason` y la expiración grupal registra `CONFIRMATION_DEADLINE`. La cancelación manual actual no solicita ni persiste motivo administrativo | Pruebas de expiración, no del flujo administrativo solicitado | No desplegado | **Parcial estructural; flujo pendiente** | Ampliar contrato de cancelación admin, validación, auditoría y UI. Evidencia: `participants_repository.go`, `reservations_repository.go`, `CancelReservationRequest`. |
| **Historial personal y trazabilidad — RF-16 — MVP 1/2/3** | Reservas propias/participadas en MVP 1; episodios de talleres en MVP 2; clases/eventos institucionales en MVP 3; no inferir asistencia | Historial personal reúne reservas y altas/bajas de talleres. Admin consulta reservas, pero no un historial institucional completo. Auditoría especializada existe para participantes, objetivo y talleres | Pruebas de vistas y repositorios existentes; Vitest completo no revalidado | No validado online | **Parcial** | Añadir relación explícita usuario–actividad antes de mostrar clases/eventos como actividad personal. Evidencia: `HistoryView.vue`, `GetReservationsByUserID`, `GetWorkshopEnrollmentsForUser`. |
| **Pantalla informativa pública — RF-17 — MVP 3** | Disponibilidad pública sanitizada | No hay ruta pública de disponibilidad o recursos. `/availability` y `/resources` requieren autenticación | Sin prueba de modo público | No desplegado | **Pendiente** | Diseñar contrato público con minimización y caché. Evidencia: `frontend/src/router/index.js`, `backend/internal/routes/routes.go`. |
| **Bloqueos de disponibilidad — RF-18 — MVP 3** | Administrador deshabilita intervalos por mantención, limpieza u otra causa | Esquema y repositorio leen `availability_blocks`; el backend por rango los expone sanitizados. No hay CRUD administrativo y la UI principal no pide el rango, por lo que no los muestra | Pruebas backend de composición presentes | No validado online | **Parcial** | Implementar gestión admin y conectar el rango en frontend. Evidencia: `availability_repository.go`, `routes.go`, `AvailabilitySection.vue`. |
| **Talleres — RF-19 — MVP 2** | Listar, inscribir, desinscribir, reinscribir, controlar cupo y solape taller↔taller, conservar historial | Alta idempotente, baja propia idempotente, nueva fila al reinscribir, cupo, ocurrencias normalizadas, auditoría e historial implementados. MVP 2 no agrega bloqueo taller↔reserva ni cutoff mientras el taller permanezca activo | **Verificado localmente el 2026-08-04** según evidencia canónica. Sin nueva suite completa en este corte | Integración online pendiente | **Parcial para cierre: alcance funcional verificado localmente; integración online pendiente** | Ejecutar exactamente el mismo alcance en E2E y online. No quedan como brechas de RF-19 las reglas taller↔reserva ni el cutoff retirados del alcance. Evidencia: `workshops_repository.go`, `workshops_service.go`, `workshops_handlers.go`, pruebas asociadas. |
| **Campeonatos — RF-20 — fuera de alcance** | La gestión avanzada está delegada a DAVE; una actividad institucional puede clasificarse visualmente como campeonato sin implementar un torneo | Solo es posible una clasificación genérica dentro de programación institucional | No aplica como módulo completo | No desplegado | **Fuera de alcance del prototipo actual** | Evitar presentar un chip o fila de agenda como gestión de campeonato. |
| **Notificaciones core — RF-21 — MVP 4** | Alertas internas, consulta, lectura, destinatarios y acciones para los eventos esenciales del prototipo | `GET /api/notifications`, contador y creación de notificación por expiración grupal. No hay marcar leído, vista completa funcional ni cobertura general de eventos | Cobertura parcial | No validado online | **Parcial** | Diseñar eventos, deduplicación, lectura y destinos. La notificación puntual requerida por prioridad institucional se implementa como dependencia de RF-09/MVP 3 sin adelantar el cierre completo de RF-21. Evidencia: `notifications_repository.go`, `notifications_handlers.go`, `NotificationBell.vue`. |
| **Detección de abuso — RF-22 — fuera de alcance** | Posible trabajo futuro sujeto a definición ética y aprobación específica | No existe motor, regla ni interfaz específica | Sin pruebas | No desplegado | **Fuera de alcance del prototipo actual** | No incorporarlo a MVP 4 ni presentarlo como capacidad del prototipo. |
| **`OPEN_USE` — regla transversal MVP 2** | No usa flujo grupal ni frecuencia; admite concurrencia del recurso; sí respeta bloqueo activo y conflicto de agenda personal; extremos contiguos permitidos | Backend diferencia bloqueo de recurso y conflicto del usuario. El frontend legado retorna “sin conflicto” para cualquier candidato `OPEN_USE`, por lo que puede ofrecer un horario incompatible que el servidor rechazará | Pruebas nuevas del helper/backend existen; el flujo principal no usa el helper | No validado online | **Backend implementado; frontend incompleto P1** | Integrar `hasAvailabilityIntervalConflict()` y `currentUserId`. Evidencia: `AvailabilitySection.vue`, `availabilityRules.js`, `availability_range_test.go`. |
| **Política prospectiva — MVP 2** | Nuevas políticas aplican prospectivamente; cada reserva conserva snapshot; corrección excepcional separada | GET actual, historial/publicación admin backend, idempotencia y referencias desde reserva. No existe UI administrativa completa. No se valida expresamente que `groupResourceIds` excluya modos incompatibles como `OPEN_USE` | Pruebas de servicio/repositorio existentes | No validado en Azure | **Parcial** | Agregar invariantes por modo, vista previa e interfaz. Evidencia: `reservation_policies_service.go`, `reservation_policies_repository.go`, `reservation_policy_*_test.go`. |
| **Código grupal, deadline y privacidad — MVP 2** | Código solo owner, hash para búsqueda, secreto cifrado recuperable, rotación, respuesta indistinguible para código inválido/cancelado, límite configurable | Implementado con `join_code_hash`, secreto cifrado/versionado, endpoints owner y `/group-reservations/:code`; expiración perezosa y worker | Pruebas de contrato, cifrado, handlers y SQL mock existentes | Claves y worker online no revalidados | **Implementado localmente; condicionado** | Verificar variables `JOIN_CODE_*`, rotación y expiración concurrente en despliegue. Evidencia: `joinsecret/`, `participants_repository.go`, `participants_handlers.go`. |
| **Clases, eventos y actividades institucionales — MVP 3** | Programación, disponibilidad e historial institucional; relación explícita para atribución personal | `scheduled_activities` se consulta y aparece en disponibilidad; no existe gestión, prioridad ni historial institucional completo | Cobertura parcial de disponibilidad | No desplegado como flujo | **Parcial estructural** | Implementar CRUD, permisos, estados, relación personal y auditoría. |
| **Administración general — MVP 3** | Usuarios, recursos, políticas, bloqueos, programación, conflictos e infracciones | Panel, métricas, lecturas de usuarios/reservas/recursos y edición de imagen; publicación de política solo API. Reportes son indicadores derivados | Pruebas UI parciales | No validado online | **Parcial** | Separar indicador de reporte oficial e implementar mutaciones autorizadas. Evidencia: `AdminView.vue`, `UsersView.vue`, `ResourcesView.vue`, `ReportsView.vue`. |

## 5. Requisitos no funcionales, seguridad y operación

| RNF / área | Alcance aprobado | Implementación observable | Prueba local | Validación integrada / despliegue | Estado | Pendiente y evidencia |
|---|---|---|---|---|---|---|
| **RNF-01 Usabilidad y responsive** | Uso móvil y escritorio | Componentes responsive, skeletons, modales accesibles y tokens | Pruebas de componentes documentadas; QA visual completo no revalidado | No validado multidispositivo online | **Parcial** | Ejecutar 377, 500, 768 y 1440 px, teclado y lector. |
| **RNF-02 Claridad visual** | Distinguir libre, ocupado, pendiente, taller y bloqueo con texto además de color | Sistema de chips y detalles existe; los bloqueos no llegan al flujo principal por falta del rango | Pruebas de tipo/chip existentes | No validado online | **Parcial** | Resolver primero la integración de disponibilidad. |
| **RNF-03 Seguridad y privacidad** | Autorizar en backend, minimizar RUT/correo, proteger secretos y transporte | Sanitización por audiencia, actor server-side, código cifrado y variables de entorno. Configuración exige cifrado DB; la confianza de certificado queda limitada a servidor local | Pruebas de audiencia, contrato y cifrado existentes | Secretos, Entra y TLS productivo no revalidados | **Implementado localmente; condicionado** | Revisar configuración real, rotación, logs y respuestas de error. |
| **RNF-04 Trazabilidad** | Acción crítica con actor, instante y motivo | Auditorías de participantes, objetivo y talleres; esquema general de auditoría. Cancelación manual y administración completa no cubren aún todos los motivos | Pruebas parciales | No validado online | **Parcial** | Crear matriz evento→auditoría y verificar retención. |
| **RNF-05 Mantenibilidad** | Capas frontend/API/SQL separadas y documentación viva | La autoridad documental quedó resuelta el 2026-08-11: `docs/Propuesta nueva/` es el canon y la serie raíz quedó supersedida. Aún existen contenido histórico preservado y dos árboles completos de base | Compilación histórica; sin evidencia nueva completa | No aplica | **Gobierno documental resuelto; deuda técnica parcial** | Mantener un solo canon y resolver la duplicación de base sin reescribir históricos. |
| **RNF-06 Disponibilidad operativa** | Compatibilidad con jornada institucional | Reglas horarias implementadas; no existe evidencia de continuidad o SLA | Reglas locales cubiertas | Sin monitoreo/uptime validado | **Parcial** | No interpretar validación horaria como alta disponibilidad. |
| **RNF-07 Concurrencia** | Evitar dobles reservas/participaciones | Transacciones serializables, locks, triggers y consultas de solape | Unitarias y SQL mock existentes | Concurrencia real Azure pendiente | **Implementado localmente; no validado integrado** | Ejecutar carreras reales después de `007`/`008`. |
| **RNF-08 Escalabilidad** | Agregar recursos/sedes sin rediseño total | Esquema relacional admite múltiples recursos; políticas referencian recursos | Revisión estructural | Sin prueba de carga o multisede | **Soporte estructural parcial** | Definir sede, capacidad y rendimiento antes de afirmar escalabilidad operativa. |
| **Base de datos y migraciones** | Una sola base compatible y migraciones acumulativas, idempotentes y recuperables | Cadena canónica `001`–`008`. La propuesta `009` está excluida de producción y permanece únicamente dentro del árbol duplicado `database/poliredi_database_improved/poliredi_database_improved/` | Validación estática documentada; `007`/`008` no tienen nueva ejecución real en este corte | Pendiente en Azure SQL | **P0 por `007`/`008`; `009` fuera de ruta** | No crear otra base ni ejecutar `009`. Su aprobación o descarte es una decisión P3 que no bloquea MVP 2 mientras permanezca excluida. |
| **Despliegue** | Frontend, API, Entra y Azure SQL operativos juntos | Existe workflow de Azure Static Web Apps para frontend. Backend y migraciones no poseen en el repositorio una canalización equivalente que pruebe el despliegue conjunto | Build histórico del 2026-08-04; no revalidado en este corte | Flujo online no ejecutado | **Pendiente de validación integrada** | Registrar URL/versiones/commit, health, CORS, identidad, API y DB en una misma acta. Evidencia: `.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml`. |
| **Deuda técnica y documental** | Mantener una fuente clara y código legible | Persisten parser textual legado de talleres, mojibake en textos, ZIP versionado, árbol DB duplicado y documentos supersedidos preservados. La autoridad canónica ya está resuelta | Revisión estática | No aplica | **P3** | Resolver duplicaciones técnicas sin reescribir históricos ni reabrir la decisión de gobierno. |

## 6. Hallazgos priorizados

### P0 — impiden declarar cierre o desplegar con confianza

1. Ejecutar `007` y `008` en una copia/backup recuperable de la única Azure SQL y verificar concurrencia real.
2. Garantizar que `009` y el árbol `poliredi_database_improved` no se ejecuten ni creen una segunda base; su aprobación o descarte no bloquea MVP 2 mientras permanezca fuera de la ruta.
3. Validar conjuntamente frontend, backend, Entra y base desplegada antes de declarar MVP 2 cerrado.

### P1 — estabilización funcional inmediata

1. Hacer que la vista de disponibilidad use obligatoriamente el contrato por rango.
2. Eliminar la reconstrucción paralela de talleres y consumir la respuesta unificada del backend.
3. Pasar política, apertura, cierre, intervalo y usuario actual al timeline.
4. Corregir la prevención de solape personal para `OPEN_USE` en frontend.
5. Filtrar recursos por `policy.resourceIds` también en el selector y validar invariantes de modos al publicar políticas.
6. Resolver la diferencia de alineación de slots entre frontend y backend.
7. Corregir cualquier documento que declare implementadas prioridad institucional o administración completa.

### P2 — alcance funcional pendiente

- CRUD administrativo de usuarios, recursos, bloqueos y actividades.
- Prioridad institucional con vista previa, cancelación y auditoría.
- Notificación específica de prioridad institucional como dependencia de MVP 3.
- Pantalla pública e historial institucional según el alcance de MVP 3.
- Notificaciones core con lectura y eventos completos en MVP 4.

### P3 — deuda controlable

- Código legado de talleres, mojibake, ZIP y árboles duplicados.
- Reportes todavía no auditables institucionalmente.
- Diagramas y casos de uso incompletos.
- Decidir si la propuesta `009` se integra en un incremento futuro o se descarta; no es dependencia de MVP 2.

## 7. Decisiones resueltas y pendientes

1. **Gobierno documental — resuelto:** desde el 2026-08-11, `docs/Propuesta nueva/` es el canon adoptado y el índice raíz redirige a este paquete.
2. **Migración `009` — decisión P3:** aprobarla para un incremento futuro o descartarla; mientras permanezca excluida no bloquea MVP 2 y nunca debe tratarse como una nueva base productiva.
3. **Agenda personal de talleres — resuelta para MVP 2:** solo se controla solape taller↔taller. No se agrega taller↔reserva ni cutoff mientras el taller esté activo; cualquier ampliación futura exige una nueva decisión.
4. **MVP 3 institucional:** aprobar el contrato de prioridad, modificación administrativa, motivos y la notificación específica al afectado antes de programarlo.
5. **Pantalla pública — MVP 3:** definir audiencia, campos y política de actualización sin reutilizar respuestas administrativas.

## 8. Casos de uso que requieren diagrama

- Máquina de estados grupal: creación, confirmación, retiro, deadline, expiración, cancelación y rotación de código.
- Secuencia de prioridad institucional: actividad↔reserva y actividad↔actividad.
- Composición de disponibilidad y sanitización por audiencia.
- Publicación prospectiva de política y corrección excepcional.
- Bloqueo/desbloqueo de usuario con motivo y auditoría.
- Migración, rollback y recuperación sobre la única base.

## 9. Fuente documental adoptada

1. `docs/Propuesta nueva/` es la única fuente documental canónica adoptada desde el 2026-08-11; su índice define precedencia y estado vigente.
2. `Documentos/01-alcance-definitivo-prototipo.md` se conserva como antecedente del alcance aprobado, no como un segundo canon operativo.
3. `docs/00-indice-maestro-y-trazabilidad.md` actúa como aviso de supersesión y redirige al canon vigente.
4. `Documentos/04-matriz-requisitos-integrada.md` resume estados y enlaza esta auditoría; no constituye una interpretación independiente.
5. Cada cambio debe registrar por separado alcance, implementación, prueba local, validación integrada, despliegue, fecha, ambiente y commit.
