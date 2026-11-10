# Poli-REDI — Matriz de trazabilidad de requisitos, historias, casos de uso y pruebas

**Versión:** 1.2 (conciliación RF-022)
**Fecha de revisión documental:** 2026-11-05
**Rama documental consultada:** `feature/mvp3`
**Tipo de documento:** instrumento de trazabilidad y planificación de QA; **no** certifica que se hayan ejecutado pruebas.
**Estado:** revisión documental propuesta, pendiente de conciliación con las fuentes oficiales.

## 1. Propósito, fuentes y alcance

Esta matriz relaciona los 26 requisitos funcionales (RF), 12 requisitos no funcionales (RNF), 21 historias de usuario (HU, incluidas HU-006A y HU-007A) y 20 casos de uso (CU, incluidas las familias CU-SED y CU-REC) descritos en los documentos del proyecto. Registra la asignación oficial por MVP y distingue implementación, evidencia y verificación.

**Fuentes primarias (rama `feature/mvp3`):**

1. `docs/08-requisitos-historias-casos-uso.md`: definición y estado declarado de RF/RNF/HU/CU; matriz inicial **RF → HU → Backlog**.
2. `docs/09-mvps-roadmap.md`: asignación **explícita** a MVP y estados narrados del roadmap.
3. `docs/07-backlog.md`: identificadores del backlog y registro de ejecución/cierre por tarea.

**Reglas de interpretación:**

- `Fuente` significa asociación literalmente documentada en la matriz inicial RF–HU–Backlog; no significa que las pruebas hayan pasado.
- `I:` ante un CU indica **correspondencia funcional inferida por esta revisión**, no una relación formal aprobada en la matriz de origen. Si el CU cubre solo parte del requisito se anota expresamente.
- Los MVP de cada fila provienen de las **listas explícitas** de `docs/09-mvps-roadmap.md`, aunque un componente técnico haya aparecido antes. `—` significa que el documento **no asigna ese elemento individualmente**.
- **Implementación (Req.)** reproduce o condensa el estado de `docs/08-...`; **Roadmap/Backlog** puede discrepar y debe conciliase, no sustituirlo silenciosamente.
- **Verificación QA** queda `Sin evaluar en esta matriz` hasta que exista ejecución, entorno, resultado, fecha, commit y evidencia vinculada. Una implementación documentada no equivale a una prueba aprobada.
- Los escenarios `PT-*` son **propuestas de diseño de prueba**, no criterios oficiales ni evidencias de ejecución.
- No se generan RF/HU/CU nuevos para cubrir huecos sin validación previa del responsable del proyecto.

### 1.1. Vocabulario de estado

| Dimensión | Valores permitidos | Significado |
| --- | --- | --- |
| Implementación | Implementado · Parcial · Pendiente · Aprobado con alcance acotado | Estado declarado por la documentación de requisitos, no conclusión independiente de esta matriz. |
| Verificación | No ejecutada · Aprobada · Fallida · Bloqueada · No aplica | Requiere evidencia por caso/entorno. Todas las pruebas propuestas parten como **No ejecutadas en esta matriz**. |
| Relación CU | `I:CU-...` · `—` | Vinculación funcional propuesta o ausencia de un CU específico. |
| Discrepancia | `BT-##` | Brecha documental o de alcance que requiere conciliación. |

## 2. Matriz de requisitos funcionales (RF)

**Lectura:** HU y Backlog proceden literalmente de la tabla inicial del documento de requisitos; los CU son un mapeo **inferido**. Las pruebas `PT-*` son propuestas para construir evidencia, no afirmaciones de cumplimiento.

| RF | Requisito | HU de fuente | Backlog de fuente | CU funcional propuesto | MVP oficial | Implementación (Req.) | Diseño QA |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RF-001 | Autenticación | HU-001 | AUTH-001, API-003 | I:CU-001 | 1 | Implementado | PT-001 |
| RF-002 | Sincronizar usuario autenticado | HU-001 | AUTH-002 | I:CU-001 (solo flujo de acceso) | 1 | Implementado; integración online pendiente | PT-001 |
| RF-003 | Registrar/validar RUT | HU-002 | AUTH-004 | I:CU-002 | 2 | Implementado | PT-002 |
| RF-004 | Bloquear reserva sin RUT | HU-002, HU-004 | AUTH-004 | I:CU-002, CU-003 | 2 | Implementado | PT-002, PT-003 |
| RF-005 | Consultar disponibilidad | HU-003 | RES-003, RES-004 | I:CU-003 (**solo verificación de disponibilidad previa a reservar**; falta CU autónomo) | 2, 3 | Implementado y validado para alcance MVP3 (2026-10-26) | PT-004 |
| RF-006 | Crear reservas | HU-004, HU-015 | RES-001, RES-002, RES-005, RES-008, RES-009, RES-010, RES-011, RES-012 | I:CU-003, CU-009 (grupales) | 1, 2 | Implementado con evidencias locales; cierre integral pendiente | PT-003, PT-005, PT-006 |
| RF-007 | Validar conflictos | HU-004, HU-012, HU-016 | RES-004, RES-010, ADMIN-004, ADMIN-005, QA-001 | I:CU-003, CU-010 (subflujo institucional) | 2, 3 | Implementado y validado técnicamente para MVP3 (2026-10-26) | PT-004, PT-007, PT-015 |
| RF-008 | Cancelar reservas | HU-005 | RES-007, RES-009, RES-010, API-003 | I:CU-004 | 1, 2 | Implementado; validación integral/online separada | PT-008 |
| RF-009 | Reservas visibles según rol | HU-006, HU-006A | RES-006 | I:CU-005 (**solo usuario normal**; admin sin CU dedicado) | 2 | Implementado | PT-009 |
| RF-010 | Historial según rol | HU-007, HU-007A | UI-003 | — (CU-005 no describe integralmente historial) | 2 | Implementado | PT-009 |
| RF-011 | Catálogo de recursos | HU-003, HU-011 | UI-001, API-001 | I:CU-REC-05 (**consulta admin**; catálogo público no cubierto) | 2 | Implementado | PT-004, PT-010 |
| RF-012 | Panel administrador | HU-009 | ADMIN-001 | I:CU-006 | 3 | Implementado | PT-011 |
| RF-013 | Gestión de usuarios | HU-010 | ADMIN-002 | — (CU-006 es panel, no CRUD de usuarios) | 3 | Implementado y validado para MVP3 (2026-10-26) | PT-011, PT-012 |
| RF-014 | Gestión de recursos | HU-011, HU-017 | ADMIN-003 | I:CU-REC-01 a CU-REC-05 | 3 | Implementado para MVP3; parcial respecto del alcance ampliado (2026-10-26) | PT-010, PT-013 |
| RF-015 | Bloqueos administrativos | HU-012 | ADMIN-004 | — (CU-010 es resolución de conflictos, no CRUD de bloqueos) | 3 | Implementado y validado para MVP3 (2026-10-26) | PT-014 |
| RF-016 | Programación institucional | HU-012 | ADMIN-005 | I:CU-010 (**solo resolución de conflicto**; no programación completa) | 3 | Implementado para prototipo MVP3 (2026-10-26) | PT-015 |
| RF-017 | Notificaciones | HU-008 | NOTIF-001 | I:CU-007 | 3 | Parcial según Req.; HU-008 figura en MVP2 | PT-016 |
| RF-018 | Reportes | HU-013 | REP-001, REP-002, REP-003 | — (CU-006 no sustituye CU de reportes) | 3, 4 | Parcial: REP-003 cerrado; REP-001 y REP-002 pendientes (2026-10-26) | PT-017 |
| RF-019 | Talleres deportivos | HU-014 | UI-006 | I:CU-008 | 2 | Implementado | PT-018 |
| RF-020 | Restricción semanal | HU-004 | RES-012 | I:CU-003 (validación de negocio) | 2 | Aprobado e implementado; cierre de ambiente objetivo pendiente | PT-006 |
| RF-021 | Mínimo de participantes | HU-015 | RES-008 | I:CU-009 | 2 | Aprobado e implementado; cierre integral 14D pendiente | PT-019 |
| RF-022 | Estado/condición grupal | HU-004, HU-015 | RES-008, RES-010 | I:CU-003, CU-009 | 2 | Backend: transiciones `PENDING/CONFIRMED/AT_RISK` implementadas; `MINIMUM_NOT_MET` solo en `PENDING`; validación integral pendiente | PT-019, PT-020 |
| RF-023 | Conflictos institucionales | HU-016 | ADMIN-005, RES-004, NOTIF-001 | I:CU-010 | 3 | v2 técnica implementada y probada; **ratificación institucional pendiente** | PT-015, PT-021 |
| RF-024 | Inventario oficial administrable | HU-011, HU-017 | ADMIN-003 | I:CU-REC-01 a CU-REC-05 | 3 | Implementado para el alcance MVP3 (2026-10-26) | PT-013 |
| RF-025 | Versionado de políticas | HU-018 | ADMIN-006 | I:CU-011 | 3 | ADMIN-006 implementado; requisito completo parcial por migración excepcional pendiente (2026-10-26) | PT-022 |
| RF-026 | Administrar sedes | HU-019 | ADMIN-003, API-001 | I:CU-SED-01 a CU-SED-04 | 3 | Implementado y validado para alcance MVP3 (2026-10-26) | PT-023 |

**Nota de alcance:** `RF-025` contiene una capacidad posterior de migración excepcional de solicitudes. El cierre del bloque prospectivo no prueba ni exige que esa migración excepcional ya esté implementada. `RF-023/v2` conserva carácter provisional hasta su ratificación (BT-004).

## 3. Matriz de requisitos no funcionales (RNF)

Los controles son **propuestas de verificación QA** basadas en el texto de cada RNF. No se deben presentar como criterios de aceptación citados literalmente.

| RNF | Necesidad | MVP oficial | Relación HU/Backlog de fuente, si consta | Verificación QA propuesta | Prueba |
| --- | --- | --- | --- | --- | --- |
| RNF-001 | Seguridad | 1 | — | Acceso protegido sin credencial, con credencial inválida y con roles no autorizados | PT-N01 |
| RNF-002 | No exposición de secretos | 1 | — | Revisar archivos versionados, documentación, respuestas públicas y logs | PT-N02 |
| RNF-003 | Trazabilidad | 4 | — | Comprobar correspondencia entre cambios, backlog y evidencia de cierre | PT-N03 |
| RNF-004 | Documentación viva | 4 | — | Contrastar documentación de arquitectura, instalación, BD y flujo con código actual | PT-N04 |
| RNF-005 | Validación técnica mínima | 4 | — | Ejecutar `go test ./...`, `npm test` y `npm run build`, o documentar bloqueo | PT-N05 |
| RNF-006 | Compatibilidad local | 1, 4 | — | Instalar en entorno limpio con Go/Fiber, Vue/Vite y PostgreSQL 16; migrar antes de habilitar MVP2 | PT-N06 |
| RNF-007 | Usabilidad | 4 | — | Estados de carga, vacío, error, foco y retroalimentación en flujos principales | PT-N07 |
| RNF-008 | Minimización de datos | — | HU-003, HU-004; API-004 | Confirmar que disponibilidad ajena no expone PII ni datos internos | PT-N08 |
| RNF-009 | Accesibilidad básica | — | HU-003, HU-004, HU-005; UX-001, UX-002 | Teclado, labels, enfoque, Escape y modales en flujos críticos | PT-N09 |
| RNF-010 | Consistencia temporal | — | — | Zona `America/Santiago`, serialización, cambios de día y límites de cancelación | PT-N10 |
| RNF-011 | Errores públicos seguros | — | SEC-005 (mencionado en la sección RNF) | Errores sin SQL/JWT/JWKS/configuración interna y logs seguros | PT-N11 |
| RNF-012 | Responsive transversal | — | BACK-021, BACK-022 (mencionados en RNF) | Uso desde 320 px sin superposiciones ni controles inaccesibles | PT-N12 |

**Pendiente documental:** `RNF-008` a `RNF-012` carecen de asignación individual explícita a MVP en el roadmap. No se les asigna uno por analogía (BT-003).

## 4. Historias de usuario y trazabilidad con casos de uso

La columna CU es una propuesta funcional `I:`. **La asignación HU→MVP no se deriva del MVP de los RF**, sino de la enumeración literal del roadmap.

| HU | Propósito resumido | MVP explícito | CU funcional propuesto | Observación |
| --- | --- | --- | --- | --- |
| HU-001 | Iniciar sesión | 2 | I:CU-001 | RF-001/002 están en MVP1: niveles de asignación distintos. |
| HU-002 | Completar RUT | 2 | I:CU-002 | Incluye restricción previa a la reserva. |
| HU-003 | Consultar disponibilidad | 2 | I:CU-003 (parcial) | No hay CU autónomo de consulta de disponibilidad. |
| HU-004 | Crear reserva | 2 | I:CU-003 | No asignar HU a MVP1 por el RF-006. |
| HU-005 | Cancelar reserva propia | 2 | I:CU-004 | No asignar HU a MVP1 por el RF-008. |
| HU-006 | Revisar mis reservas | 2 | I:CU-005 | Flujo de usuario normal. |
| HU-006A | Revisar reservas como administrador | — | — | No tiene asignación individual explícita; CU-005 no cubre admin. |
| HU-007 | Revisar historial | 2 | — | Falta CU específico de historial. |
| HU-007A | Historial global administrador | — | — | No tiene asignación individual explícita ni CU propio. |
| HU-008 | Revisar notificaciones | 2 | I:CU-007 | RF-017 está en MVP3; inconsistencia de granularidad (BT-002). |
| HU-009 | Acceder a panel administrador | 3 | I:CU-006 | El panel no cubre por sí solo otras gestiones administrativas. |
| HU-010 | Gestionar usuarios | 3 | — | Requiere caso(s) de uso específicos si se formalizan. |
| HU-011 | Gestionar recursos | 3 | I:CU-REC-01 a CU-REC-05 | No consta en MVP2 como HU individual. |
| HU-012 | Crear bloqueos | 3 | — | CU-010 no equivale a CRUD de bloqueos. |
| HU-013 | Consultar reportes | 3 | — | RF-018 también aparece en MVP4; HU no. |
| HU-014 | Inscribirse en taller | 2 | I:CU-008 | — |
| HU-015 | Completar grupo | 2 | I:CU-009 | — |
| HU-016 | Resolver conflictos | 3 | I:CU-010 | Regla v2 provisional. |
| HU-017 | Mantener inventario | 3 | I:CU-REC-01 a CU-REC-05 | Abarca recursos administrables. |
| HU-018 | Configurar políticas | 3 | I:CU-011 | Distinguir prospectivo de migración excepcional. |
| HU-019 | Administrar sedes | 3 | I:CU-SED-01 a CU-SED-04 | La HU está **incluida** en el roadmap MVP3. |

## 5. Inventario y cobertura de casos de uso

| CU existente | Función efectiva según documento de requisitos | MVP listado explícitamente | Relación funcional propuesta |
| --- | --- | --- | --- |
| CU-001 | Autenticarse | 2 | I:HU-001 / RF-001, RF-002 |
| CU-002 | Registrar RUT obligatorio | 2 | I:HU-002 / RF-003, RF-004 |
| CU-003 | Crear reserva | 2 | I:HU-004 / RF-006, RF-007, RF-020, RF-022 (verificar subflujos) |
| CU-004 | Cancelar reserva | 2 | I:HU-005 / RF-008 |
| CU-005 | Consultar mis reservas | 2 | I:HU-006 / RF-009 (usuario normal) |
| CU-006 | Revisar panel administrador | 3 | I:HU-009 / RF-012; **no** CRUD de usuarios/recursos |
| CU-007 | Consultar notificaciones | 2 | I:HU-008 / RF-017 |
| CU-008 | Inscribirse en taller | 2 | I:HU-014 / RF-019 |
| CU-009 | Confirmar participantes grupales | 2 | I:HU-015 / RF-021, RF-022 |
| CU-010 | Resolver conflicto institucional | 3 | I:HU-016 / RF-023, parte de RF-007/RF-016 |
| CU-011 | Configurar políticas | 3 | I:HU-018 / RF-025 |
| CU-SED-01 | Crear sede | — | I:HU-019 / RF-026 |
| CU-SED-02 | Editar sede | — | I:HU-019 / RF-026 |
| CU-SED-03 | Activar/desactivar sede | — | I:HU-019 / RF-026 |
| CU-SED-04 | Consultar sedes | — | I:HU-019 / RF-026 |
| CU-REC-01 | Crear recurso | — | I:HU-011, HU-017 / RF-014, RF-024 |
| CU-REC-02 | Editar recurso | — | I:HU-011, HU-017 / RF-014, RF-024 |
| CU-REC-03 | Activar/desactivar recurso | — | I:HU-011, HU-017 / RF-014, RF-024 |
| CU-REC-04 | Cambiar sede de recurso | — | I:HU-011, HU-017 / RF-014, RF-024 |
| CU-REC-05 | Consultar y filtrar recursos | — | I:HU-011, HU-017 / RF-014, RF-024; consulta admin de RF-011 |

**Importante:** CU-SED y CU-REC **ya están definidos** en `docs/08-requisitos-historias-casos-uso.md`, aunque sus IDs no aparecen individualmente en las listas MVP del roadmap. No se propone duplicarlos.

## 6. Diseño de pruebas funcionales pendiente de ejecución

Estas pruebas definen qué **comprobar**, sin afirmar que ya se ejecutaron. Cada ejecución debe conservar su resultado y evidencia.

| ID | RF de referencia | Escenario / comprobación propuesta | Tipo sugerido | Resultado en esta matriz |
| --- | --- | --- | --- | --- |
| PT-001 | RF-001, 002 | Acceso autorizado/denegado, identidad Entra, sincronización de `oid`/tenant y persistencia | Integración + seguridad | No ejecutada |
| PT-002 | RF-003, 004 | RUT válido, inválido, duplicado y ausencia de RUT antes de reservar | API + UI | No ejecutada |
| PT-003 | RF-004, 006 | Crear reserva autenticada, validar recurso, fecha, duración permitida (30–180 min) y estado de servidor | API + BD | No ejecutada |
| PT-004 | RF-005, 007, 011 | Disponibilidad por recurso/fecha, ocupaciones institucionales, talleres, bloqueos y filtros fecha/rango; datos ajenos sanitizados | Integración + UI | No ejecutada |
| PT-005 | RF-006 | Duración, horario operativo, día institucional y capacidad; fallas atómicas | API + BD | No ejecutada |
| PT-006 | RF-006, 020 | Frecuencia semanal versionada: `PENDING` consume, cancelación libera y concurrencia respeta restricción | Integración + concurrencia | No ejecutada |
| PT-007 | RF-007 | Colisiones concurrentes reserva–reserva, reservas–bloqueos y agenda institucional | Integración PostgreSQL | No ejecutada |
| PT-008 | RF-008 | Propietario/admin, estado cancelable, confirmación y rechazo por reserva finalizada | API + UI | No ejecutada |
| PT-009 | RF-009, 010 | Reservas e historial por rol, filtros y aislamiento de terceros | API + UI + permisos | No ejecutada |
| PT-010 | RF-011, 014 | Catálogo, consulta y filtros; activación/desactivación reflejada en disponibilidad | API + UI | No ejecutada |
| PT-011 | RF-012, 013 | Acceso al panel, autorización por rol y controles administrativos visibles | Seguridad + UI | No ejecutada |
| PT-012 | RF-013 | Gestión de cuentas/roles/bloqueos y auditoría con privilegios mínimos | Integración + seguridad | No ejecutada |
| PT-013 | RF-014, 024 | Alta/edición de recursos, datos obligatorios, sede, historial, invariantes, rechazo de traslado incompatible | API + BD + UI | No ejecutada |
| PT-014 | RF-015 | Alta/consulta/desactivación de bloqueo, motivo, solapamientos y efecto en calendario | API + BD + UI | No ejecutada |
| PT-015 | RF-007, 016, 023 | Programación SINGLE/WEEKLY, conflictos múltiples, permisos de gestor y resolución administrativa | Integración + UI | No ejecutada |
| PT-016 | RF-017 | Notificaciones por eventos, recepción y visibilidad sin PII de terceros | Integración + UI | No ejecutada |
| PT-017 | RF-018 | Reportes de reservas, duraciones, métricas/infracciones, filtros y permisos; validar REP-001/002/003 por separado | SQL + API + UI | No ejecutada |
| PT-018 | RF-019 | Listar talleres, cupos, inscribir, bloquear duplicación, validar RUT y reflejar ocupación | Integración + UI | No ejecutada |
| PT-019 | RF-021, 022 | Solicitante como primer participante, cuentas únicas, código de invitación, mínimo por recurso y plazo | Integración + UI | No ejecutada |
| PT-020 | RF-022 | `PENDING/PENDING_MINIMUM` → `CONFIRMED/HEALTHY` → `CONFIRMED/AT_RISK` → recuperación `HEALTHY`; en el instante exacto del deadline se rechazan altas/retiros; al vencer solo `PENDING` bajo mínimo se cancela con `MINIMUM_NOT_MET`; `CONFIRMED + AT_RISK` permanece confirmada | Integración PostgreSQL + tiempo simulado | No ejecutada en esta matriz |
| PT-021 | RF-023 | EV-010 v2: detección, decisión KEEP/ALLOW/CANCEL/RESCHEDULE, justificación, auditoría, notificación, atomicidad/rollback | PostgreSQL + permisos + UI | No ejecutada |
| PT-022 | RF-025 | Versión prospectiva, permisos, idempotencia, snapshot histórico; **separar** migración excepcional posterior | PostgreSQL + API + UI | No ejecutada |
| PT-023 | RF-026 | Sedes CRUD lógico, unicidad, coordenadas, relación recursos/sedes, impedir desactivación inválida, preservar historial | API + BD + UI | No ejecutada |

### 6.1. Casuística mínima crítica de grupos (PT-019 y PT-020)

1. Solicitud grupal creada: `PENDING` + `PENDING_MINIMUM`; solicitante contabilizado una vez.
2. Llegada al mínimo configurado: `CONFIRMED` + `HEALTHY`.
3. Retirada posterior al primer confirmado: **conservar `CONFIRMED`** y cambiar condición a `AT_RISK`.
4. Reingreso/confirmación suficiente: `CONFIRMED` + `HEALTHY`.
5. Vencimiento bajo mínimo **todavía en `PENDING`**: `CANCELLED` con `MINIMUM_NOT_MET`; liberar horario y oportunidad semanal.
6. Vencimiento en `CONFIRMED + AT_RISK`: no cancelar automáticamente ni liberar su horario por `MINIMUM_NOT_MET` (regresión pendiente de vincular a evidencia).
7. Instante exacto del deadline: rechazar alta y retirada de participantes.
8. Una cuenta no puede confirmarse dos veces ni incorporarse cuando existe solape personal incompatible.
9. Los recursos `OPEN_USE` no aplican al flujo de confirmación grupal.

**Conciliación RF-022:** el roadmap local ya distingue `CONFIRMED + AT_RISK` de `PENDING`; la discrepancia residual se encontraba en la frase genérica de RF-022 sobre cancelación al deadline. El contrato revisado especifica que `MINIMUM_NOT_MET` solo cancela reservas que todavía estén `PENDING`. Las verificaciones `PT-019/PT-020` siguen sin marcarse como ejecutadas en esta matriz.

### 6.2. Casuística institucional crítica (PT-015 y PT-021)

- `RF-023/v1` es antecedente histórico (cancelación automática); **no utilizarlo como oráculo de prueba vigente**.
- `RF-023/v2`, adoptado provisionalmente en EV-010, registra conflictos y exige decisión administrativa explícita; crear la actividad no cancela automáticamente reservas particulares.
- Probar permisos (MANAGER puede programar dentro de su unidad, no resolver como ADMIN), conflictos múltiples, deshacer cambios por error transaccional, justificación/auditoría y aviso al afectado.
- El conjunto no se considera normativamente cerrado hasta ratificación institucional de EV-010.

### 6.3. Control de ejecución y evidencia (plantilla)

Usar una fila **por ejecución**, sin reemplazar la definición de la prueba:

| Prueba | Fecha | Commit / rama | Entorno y versión de BD | Ejecutó | Resultado | Evidencia (ruta/URL) | Incidencia |
| --- | --- | --- | --- | --- | --- | --- | --- |
| PT-___ | AAAA-MM-DD | SHA / rama | local / aislado / Funnel / online; PG16 | — | No ejecutada | — | — |

Las ejecuciones existentes citadas en documentación (por ejemplo, suite institucional aislada y pruebas Vue de octubre de 2026) son **evidencias históricas referenciables**, pero deben vincularse explícitamente a los `PT-*` pertinentes antes de marcar una prueba de esta matriz como `Aprobada`.

## 7. Diseño de verificaciones no funcionales

| ID | Qué registrar como evidencia | Estado de ejecución en esta matriz |
| --- | --- | --- |
| PT-N01 | Requests no autenticadas, inválidas y de rol insuficiente, con códigos HTTP | No ejecutada |
| PT-N02 | Resultado de revisión de secretos en Git, `.env`, logs y respuestas públicas | No ejecutada |
| PT-N03 | Cruce de commits/requisitos/tareas y backlog sincronizado | No ejecutada |
| PT-N04 | Diff de documentación respecto a rutas, base de datos y despliegue actuales | No ejecutada |
| PT-N05 | Salida fechada de pruebas Go, frontend y compilación, por SHA | No ejecutada |
| PT-N06 | Bitácora de instalación limpia y migración PG16, sin dependencias de Azure SQL vigentes | No ejecutada |
| PT-N07 | Capturas/registro de carga, error y vacío en recorridos principales | No ejecutada |
| PT-N08 | Respuestas sanitizadas del endpoint de disponibilidad bajo usuarios ajenos | No ejecutada |
| PT-N09 | Checklist de teclado y manejo de foco/Escape en modales | No ejecutada |
| PT-N10 | Casos frontera de zona horaria Chile y serialización (API/BD/UI) | No ejecutada |
| PT-N11 | Errores seguros con correlación a logs técnicos sanitizados | No ejecutada |
| PT-N12 | Capturas/pruebas de flujo desde ancho 320 px, menú, modales y tabulación | No ejecutada |

## 8. Registro de brechas y contradicciones documentales

Se registran **diferencias entre documentos**; no se declara automáticamente una falla del sistema. Prioridades propuestas: `Alta` cuando puede alterar alcance, política o casos de prueba; `Media` cuando distorsiona la trazabilidad; `Baja` cuando exige solo nomenclatura.

| ID | Prioridad | Observación contrastable | Acción documental propuesta |
| --- | --- | --- | --- |
| BT-001 | Media | **Resuelto en esta matriz:** RF-026 y HU-019 sí figuran en MVP3; CU-SED-01..04 ya existen. La versión previa de la matriz los trataba como ausentes. | No abrir HU/CU duplicados; mantener enlaces correctos. |
| BT-002 | Media | El roadmap enumera HU-008/CU-007 en MVP2, pero RF-017 en MVP3. | Documentar si hubo entrega incremental de notificaciones en MVP2 y alcance administrativo en MVP3, sin cambiar listas históricas sin acuerdo. |
| BT-003 | Media | RNF-008..RNF-012 no tienen asignación individual de MVP en roadmap. | Asignar formalmente si se requiere, o registrar que son controles transversales. |
| BT-004 | Alta | RF-023/v2 (EV-010) es la regla operativa provisional; ratificación institucional definitiva pendiente. | Obtener aprobación institucional y actualizar RF, CU, pruebas y decisión EV-010. |
| BT-005 | Baja | Inventario/sedes y nombres de recursos de referencia requieren cotejo: documentación refiere ocho recursos; otras secciones describen «tres multicanchas» y mínimos para Cancha 1/Cancha 2/Sala Multiuso. | Validar inventario real, clasificación `GROUP` y mínimos por recurso antes de fijar datos de QA. |
| BT-006 | Alta | **Conciliada documentalmente:** el backend mantiene `CONFIRMED + AT_RISK` sin cancelarla por `MINIMUM_NOT_MET`; solo expiran `PENDING` bajo mínimo. La frase general de RF-022 y la narrativa histórica sobre regreso a `PENDING` requerían precisar ese alcance. | Regla aclarada en RF-022; pendiente prueba explícita de no cancelación de `CONFIRMED + AT_RISK` en el deadline y validación de aceptación. |
| BT-007 | Media | RF-005/007 mantienen brechas documentadas de filtros/visualización de bloqueos; roadmap registra avances de ADMIN-004. | Separar disponibilidad API, calendario UI y pruebas por ambiente para determinar qué continúa pendiente. |
| BT-008 | Alta | RF-013 figura `Parcial` en Req.; el roadmap indica `ADMIN-002` cerrado. | Conciliar alcance de usuarios, permisos, auditoría, UI y evidencia de cierre. |
| BT-009 | Alta | RF-014/024 figuran `Parcial` en Req.; el roadmap indica `ADMIN-003` cerrado y gestión de sedes/recursos validada. | Actualizar estados por capacidad y alcance; documentar qué quedó explícitamente fuera (p. ej. traslados históricos). |
| BT-010 | Alta | RF-015 figura `Pendiente` en Req.; el roadmap indica `ADMIN-004` cerrado (bloqueos). | Sincronizar estado en requisitos con cierre técnico y evidencias, manteniendo limitaciones si aplica. |
| BT-011 | Alta | RF-016 figura `Parcial` en Req.; el roadmap indica `ADMIN-005` cerrado para prototipo. Además, CU-010 cubre resolución, no la programación entera. | Registrar distinción entre programación, conflictos, UI, pruebas y ratificación EV-010. |
| BT-012 | Alta | RF-026 figura implementación `Pendiente` en Req.; roadmap `ADMIN-003` cerrado incluyendo sedes, y EV-012 aprobó el alcance. | Determinar si `RF-026` está implementado en rama actual y actualizar la declaración de fuente. |
| BT-013 | Media | `ADMIN-006` cerrado en roadmap para versión prospectiva; RF-025 describe además interfaz administrativa y migración excepcional posterior. | Separar entregables por alcance y no dar por cubierta la migración excepcional. |
| BT-014 | Media | Roadmap MVP3 afirma «cerrado técnicamente» al 2026-10-24, pero otra sección dice que permanece abierto hasta completar módulos y criterios; reportes y validaciones se describen pendientes. | Revisar fechas, criterios y alcance de cierre; distinguir cierre del prototipo de validación institucional y salida a producción. |
| BT-015 | Media | HU-006A/HU-007A, CU-SED y CU-REC no están individualmente asignados a MVP en el listado formal. | Formalizar asignación si es necesaria; no inferirla del RF relacionado. |
| BT-016 | Media | No hay CU dedicado identificado para historial por rol, gestión de usuarios, bloqueos, reportes ni consulta autónoma de disponibilidad. | Revisar cobertura y decidir si ampliar CU existentes o aprobar nuevos CU; no utilizar CU-006 como comodín. |
| BT-017 | Media | Cierre QA integral/14D aún pendiente; falta vincular prueba E2E/PG16 específica de persistencia de `CONFIRMED + AT_RISK` al vencer, además de liberar horario para canceladas `PENDING`. | Incorporar pruebas frontera con reloj controlado, resultados y evidencia por SHA/entorno; no declarar validación integral por inspección documental. |
| BT-018 | Baja | Referencias a Azure SQL corresponden a evidencia histórica; roadmap y requisitos declaran PostgreSQL 16 como arquitectura vigente. | Mantener Azure SQL etiquetado «histórico» y probar instalación/operación sobre PostgreSQL 16. |

## 9. Orden recomendado de conciliación

1. **Política y corrección funcional:** BT-004, BT-006, BT-017 (RF-023/v2 y estados grupales) antes de cerrar pruebas de aceptación.
2. **Desfase de estados entre documentos:** BT-008, BT-009, BT-010, BT-011, BT-012 y BT-014; cruzar evidencia `ADMIN-*` con la descripción RF actualizada.
3. **Cobertura de QA y trazabilidad:** BT-007, BT-013, BT-016; incorporar pruebas con SHA, entorno y resultado.
4. **Organización y consistencia:** BT-002, BT-003, BT-005, BT-015 y BT-018; fijar alcance transversal y denominaciones.

## 10. Criterios para mantener la matriz

- Todo RF nuevo debe relacionarse con HU/Backlog de fuente; asociar CU y MVP solo si la fuente los respalda o dejar claramente marcada la inferencia.
- Al cambiar un estado RF, actualizar tanto requisitos como roadmap/backlog, o dejar registrada la discrepancia en `BT-*` hasta validación.
- No marcar `Aprobada` una prueba sin evidencia enlazada a SHA, ambiente, base de datos y fecha.
- Distinguir validaciones **locales/aisladas**, **Funnel**, **online** y **ratificación institucional**: no son equivalentes.
- Mantener trazabilidad histórica de decisiones `EV-010` y `EV-012`; no sobrescribir reglas originales sin registrar su sustitución.
- Antes de comprometer cambios, validar que existen las 26 filas RF, 12 RNF, 21 HU y 20 CU y que ningún identificador fue renombrado accidentalmente.

---

**Control de esta revisión:** elaboración de la matriz v1.1 a partir de documentos fuente, sin edición del código, sin ejecución de pruebas y sin declaración de cierre de MVP. Los conflictos BT permanecen pendientes hasta validación y sincronización formal.
