# Requisitos y trazabilidad

**Audiencia:** Analista, Arquitecto, desarrollo, QA y evaluación académica

**Propósito:** definir el catálogo vigente y enlazar alcance, implementación y evidencia

**Estado:** canónico, corte 2026-08-11

**Fuente:** requisitos consolidados, código local, decisiones vigentes y matriz académica

## Resumen

Poli-REDI conserva identificadores estables `RF-001`–`RF-025` y `RNF-001`–`RNF-012`. Este documento es propietario del catálogo y su asignación a MVP; las reglas detalladas viven en [06-reglas-y-flujos.md](06-reglas-y-flujos.md), los contratos en [02-arquitectura-y-contratos.md](02-arquitectura-y-contratos.md) y la evidencia en [07-calidad-y-evidencia.md](07-calidad-y-evidencia.md).

“Aprobado”, “implementado”, “probado localmente” y “validado en Azure” son estados distintos. Al 2026-08-11, MVP 1 es demostrable localmente, MVP 2 y MVP 3 están parciales, y MVP 4 está pendiente y acotado.

## Actores

| Actor | Responsabilidad |
|---|---|
| Usuario | Consultar disponibilidad, reservar, participar en grupos y talleres y revisar su historial. |
| Administrador | Consultar y, en MVP 3, gestionar usuarios, inventario, bloqueos y programación con auditoría. |
| Microsoft Entra ID | Autenticar y emitir tokens para la API desplegada. |
| Backend | Resolver identidad, permisos, reglas, conflictos y errores seguros. |
| Azure SQL | Persistir datos y defender integridad y concurrencia. |
| Pantalla pública | Mostrar disponibilidad sanitizada, sin datos personales ni acciones privadas, en MVP 3. |

## Requisitos funcionales

| ID | Requisito | MVP | Estado al corte |
|---|---|---:|---|
| RF-001 | Autenticación de usuarios | 1 | Local demostrable; Entra online pendiente. |
| RF-002 | Sincronización del usuario autenticado | 1 | Local demostrable; online pendiente. |
| RF-003 | Registro y validación de RUT cuando corresponde | 1 | Implementado localmente. |
| RF-004 | Impedir reserva de usuario normal sin RUT; administrador puede agendar sin RUT | 1 | Implementado localmente. |
| RF-005 | Consulta de disponibilidad sanitizada y por rango | 1–2 | Backend por rango existe; integración frontend pendiente. |
| RF-006 | Creación de reservas con fecha, hora, duración, actividad y participantes | 1–2 | Local parcial; regresión integrada pendiente. |
| RF-007 | Validación de conflictos y solapes | 1–3 | Local parcial; concurrencia Azure pendiente. |
| RF-008 | Cancelación por propietario o administrador autorizado antes de finalizar | 1–2 | Local; cutoff configurable es mejora futura. |
| RF-009 | Acceso y operaciones según rol | 1–3 | Base local; matriz administrativa pendiente. |
| RF-010 | Historial según audiencia | 1–3 | Reservas e inscripciones propias parciales; institucional pendiente. |
| RF-011 | Catálogo consultable de recursos | 1 | Implementado localmente. |
| RF-012 | Panel administrativo | 3 | Parcial. |
| RF-013 | Gestión de usuarios y bloqueos | 3 | Estructura parcial; CRUD y auditoría pendientes. |
| RF-014 | Gestión del inventario de recursos | 3 | Consulta base; administración pendiente. |
| RF-015 | Bloqueos de disponibilidad | 3 | Lectura backend parcial; CRUD e integración pendientes. |
| RF-016 | Programación de talleres, clases y otros eventos | 3 | Talleres parciales; administración institucional pendiente. |
| RF-017 | Notificaciones | 3–4 | Prioridad específica depende de MVP 3; sistema core corresponde a MVP 4. |
| RF-018 | Reportes básicos | 4 | Pendiente. |
| RF-019 | Talleres: alta, baja, cupo, ocurrencias e historial propio | 2 | Verificado localmente de forma parcial; cierre online pendiente. |
| RF-020 | Restricción semanal de reservas particulares | 2 | Implementada localmente; no aplica a `open_use`. |
| RF-021 | Confirmación por mínimo de participantes | 2 | Implementada localmente; migraciones y Azure pendientes. |
| RF-022 | Estado condicionado por política del recurso | 2 | Implementado localmente; regresión pendiente. |
| RF-023 | Conflictos y prioridad institucional deterministas | 3 | Pendiente. |
| RF-024 | Inventario oficial administrable | 3 | Pendiente. |
| RF-025 | Configuración administrativa prospectiva de políticas | 3 | Borrador técnico; interfaz y auditoría pendientes. |

## Requisitos no funcionales

| ID | Requisito | Evidencia necesaria |
|---|---|---|
| RNF-001 | Seguridad y autorización en backend | Matriz de permisos, pruebas negativas y revisión de configuración. |
| RNF-002 | No exposición de secretos | Revisión de repositorio, respuestas y logs. |
| RNF-003 | Trazabilidad de operaciones críticas | Actor, fecha, motivo y cambio consultables según rol. |
| RNF-004 | Documentación viva | Alcance, código, pruebas y estado se actualizan juntos. |
| RNF-005 | Validación técnica mínima | Pruebas, build, smoke y evidencia fechada. |
| RNF-006 | Compatibilidad local reproducible | Instalación desde instrucciones en ambiente limpio. |
| RNF-007 | Usabilidad | Estados comprensibles, acciones coherentes y errores útiles. |
| RNF-008 | Minimización de datos visibles | Respuestas y pantallas ajustadas a cada audiencia. |
| RNF-009 | Accesibilidad básica | Teclado, foco, contraste, texto además de color y lector. |
| RNF-010 | Consistencia temporal | Reglas en `America/Santiago` y timestamps técnicos normalizados. |
| RNF-011 | Errores públicos seguros | Mensaje de dominio sin detalles internos ni enumeración indebida. |
| RNF-012 | Responsive transversal | QA representativo en 377, 500, 768 y 1440 px. |

## Historias y casos de uso clave

| Caso | Actor | Resultado esperado | Requisitos |
|---|---|---|---|
| HU/CU-01 Consultar disponibilidad | Usuario | Ve recursos y ocupación sin datos privados. | RF-005, RF-011, RNF-008 |
| HU/CU-02 Crear reserva | Usuario/admin | Obtiene estado según política, cupo y solapes. | RF-003–009, RF-020–022 |
| HU/CU-03 Gestionar grupo | Propietario/participante | Comparte código, confirma o retira y observa progreso. | RF-006–008, RF-021–022 |
| HU/CU-04 Gestionar taller | Usuario | Se inscribe, desinscribe o reinscribe mientras está activo. | RF-019, RNF-003 |
| HU/CU-05 Administrar operación | Administrador | Gestiona inventario, usuarios, bloques y programación auditada. | RF-012–016, RF-023–025 |
| HU/CU-06 Consultar pantalla pública | Público | Ve disponibilidad sanitizada sin iniciar acciones privadas. | RF-005, RF-023, RNF-008 |

## Trazabilidad de cierre

| Capacidad | Contrato/regla | Implementación | Evidencia de salida |
|---|---|---|---|
| Identidad y RUT | [02](02-arquitectura-y-contratos.md), [06](06-reglas-y-flujos.md) | Middleware, usuarios y rutas protegidas | Login local y Entra online; permisos negativos. |
| Disponibilidad | [02](02-arquitectura-y-contratos.md) | Backend por rango y frontend parcial | Vista integrada, estados UI, privacidad y Azure. |
| Reserva básica | [06](06-reglas-y-flujos.md) | Servicio, repositorio y defensas DB | Flujos feliz, error, contigüidad y concurrencia. |
| Reserva grupal | [04](04-base-de-datos-y-migraciones.md), [06](06-reglas-y-flujos.md) | Endpoints y UI parciales | Migraciones `007`/`008`, E2E y Azure. |
| Talleres | [06](06-reglas-y-flujos.md) | Alta, baja e historial local | Taller activo, cupo, solape taller↔taller y online. |
| Administración | [02](02-arquitectura-y-contratos.md) | Componentes parciales | CRUD, permisos, prioridad, pantalla pública y auditoría. |
| Calidad operativa | [07](07-calidad-y-evidencia.md) | Parcial o pendiente | Seguridad, accesibilidad, smoke y rollback. |

La matriz académica conserva el mapeo detallado en [Documentos/04-matriz-requisitos-integrada.md](../../Documentos/04-matriz-requisitos-integrada.md). Si una síntesis académica difiere de este catálogo, prevalecen este documento y los ADR vigentes.

## Decisiones y exclusiones

- [ADR-001](decisiones/ADR-001-gobierno-documental.md): gobierno y precedencia documental.
- [ADR-002](decisiones/ADR-002-evolucion-base-unica.md): evolución incremental sobre una base única.
- [ADR-003](decisiones/ADR-003-alcance-mvp-y-exclusiones.md): asignación final de MVP y exclusiones.
- [ADR-004](decisiones/ADR-004-reglas-temporales-y-solapes.md): reglas temporales, talleres y solapes.
- La migración `009` es una propuesta no aprobada y queda fuera del cierre.
- BI avanzado, IA, integración académica, multisede, Google bidireccional, campeonatos avanzados y detección automatizada de abuso quedan fuera del prototipo.

## Regla de mantenimiento

Un cambio funcional debe actualizar el requisito, contrato, regla, prueba, plan y evidencia correspondientes. Solo se usa “verificado” cuando constan ambiente, fecha, versión, procedimiento y resultado.
