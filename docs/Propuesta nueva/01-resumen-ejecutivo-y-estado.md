# Resumen ejecutivo y estado actual de Poli-REDI

**Estado documental:** CANÓNICO

**Corte:** 2026-08-11
**Fecha límite del prototipo:** 2026-12-10 (`America/Santiago`)

## 1. Problema y objetivo

El proceso levantado depende de gestión manual y Google Calendar, concentra decisiones en el encargado y no ofrece una vista institucional única de disponibilidad, reglas ni trazabilidad.

Poli-REDI busca validar un prototipo web desacoplado que centralice recursos deportivos, disponibilidad, reservas particulares, solicitudes grupales, talleres e información administrativa básica, aplicando reglas críticas en servidor y Azure SQL.

## 2. Lectura correcta del estado

Poli-REDI es una **demo funcional avanzada verificada principalmente en local**, no una plataforma institucional cerrada. La documentación no acredita una validación online vigente de frontend, Microsoft Entra ID, CORS, API y Azure SQL trabajando juntos después de los cambios recientes.

Los estados de este documento separan implementación, verificación local y validación online. Ningún resultado local permite declarar por sí solo cerrado un MVP que exige Azure, migraciones o checklist integrado.

## 3. Estado por área

| Área | Estado vigente | Límite de la evidencia |
|---|---|---|
| Autenticación e identidad | IMPLEMENTADO LOCALMENTE | Entra ID y modo local existen; falta una validación online vigente del corte. |
| Roles y autorización | IMPLEMENTADO Y REVISADO LOCALMENTE | El backend determina identidad, rol y bloqueo; falta repetir la matriz en el ambiente desplegado. |
| Perfil y RUT | IMPLEMENTADO LOCALMENTE | Usuario normal requiere RUT para altas que lo exigen; la desinscripción propia de taller no requiere RUT. |
| Recursos | PARCIAL | Catálogo y actualización de imagen existen; falta administración completa del inventario oficial. |
| Disponibilidad backend | IMPLEMENTADO LOCALMENTE | Existe consulta por rango de fechas y contrato unificado; falta acreditar Azure. |
| Disponibilidad frontend | PARCIAL | La integración completa del rango, política, elegibilidad, bloqueos y protección frente a respuestas obsoletas está pendiente de cierre y prueba. |
| Reserva particular | IMPLEMENTADO LOCALMENTE | Servidor controla usuario, horario, duración, ventana, frecuencia y conflictos; falta E2E online vigente. |
| Flujo grupal | PARCIAL | Código, progreso, confirmación, retiro, rotación, deadline y expiración tienen evidencia local; `007`/`008`, Azure SQL y flujo integrado siguen pendientes. |
| Talleres | IMPLEMENTADO Y VERIFICADO LOCALMENTE | Alta y baja mientras el taller está activo, sin cutoff; el solape MVP 2 se evalúa solo taller↔taller, no taller↔reserva personal entre recursos. Historial y reinscripción local; integración online pendiente. |
| Cancelación | PARCIAL | Propietario o administrador pueden cancelar antes de que la reserva finalice; un cutoff configurable es mejora futura, no criterio de cierre de MVP 2. |
| Notificaciones | PARCIAL | La notificación específica asociada a prioridad depende de MVP 3; consulta, lectura, destinos y cobertura core corresponden a MVP 4. |
| Administración | PARCIAL | Panel y lecturas básicas; faltan usuarios, inventario, bloqueos, programación, conflictos e infracciones completos. |
| Pantalla pública | PENDIENTE MVP 3 | Debe mostrar disponibilidad institucional sanitizada, sin datos personales ni acciones privadas. |
| Reportes y auditoría | PARCIAL | Existen indicadores y registros, pero no la experiencia básica completa definida para MVP 4. |
| Accesibilidad y UX | PARCIAL VERIFICADO LOCALMENTE | Existen mejoras de foco, teclado, skeletons y privacidad; falta matriz manual 377/500/768/1440 y lector de pantalla. |
| Despliegue | CONFIGURADO, NO VALIDADO EN ESTE CORTE | No existe evidencia online vigente posterior a los cambios recientes. |

## 4. Estado de los MVP

| MVP | Estado recomendado | Condición de cierre | Fecha objetivo |
|---|---|---|---:|
| MVP 1 — Base operativa | DEMOSTRABLE LOCALMENTE; NO CERRADO ONLINE | Validar identidad, permisos, disponibilidad, reserva, cancelación, historial, CORS y recuperación en ambiente integrado. | 2026-08-28 |
| MVP 2 — Usuario, grupo y talleres | PARCIAL; NO CERRADO | Cerrar integración frontend de disponibilidad, ejecutar `007`/`008` y aprobar E2E local/Azure del flujo grupal y talleres. | 2026-09-25 |
| MVP 3 — Administración institucional | PARCIAL | Completar inventario, usuarios, bloqueos, programación, prioridad, notificación específica de prioridad, conflictos, pantalla pública sanitizada e historial institucional. | 2026-10-30 |
| MVP 4 — Calidad y soporte | PENDIENTE Y ACOTADO | Completar sistema core de notificaciones, reportes básicos, auditoría consultable, pruebas, documentación, despliegue y recuperación. | 2026-11-27 |

La fecha final absoluta de entrega es **2026-12-10**. Los cierres y buffers se detallan en [`11-cronograma-cierre-2026.md`](11-cronograma-cierre-2026.md).

## 5. Evidencia disponible y límites

La adopción del canon no ejecutó la suite completa. La auditoría 12 revalidó **18 pruebas Node** sobre el commit `939ba51`. No existe nueva evidencia atribuible a este corte para Vitest, pruebas Go, build frontend ni flujo online. La desinscripción de talleres conserva evidencia local histórica del 2026-08-04; no debe extrapolarse a Azure.

| Verificación obligatoria de cierre | Estado al 2026-08-11 |
|---|---|
| Node del commit `939ba51` | 18 pruebas revalidadas por auditoría 12. |
| Vitest, Go y build frontend del último incremento | Sin nueva evidencia en este corte. |
| Integración frontend del endpoint de disponibilidad por rango | Pendiente. |
| Azure SQL migraciones `007`/`008` | Pendiente. |
| Flujo online Entra ID/CORS/API/DB | Pendiente. |
| QA visual, accesibilidad y privacidad por audiencia | Pendiente. |
| Checklist total de MVP | Pendiente. |

## 6. Decisiones vigentes

1. El backend determina identidad, rol, bloqueo y capacidades.
2. La creación falla cerrada si no existe una política publicada válida.
3. La agenda usa hora institucional de muro `America/Santiago`; los timestamps técnicos usan UTC.
4. `OPEN_USE` permite concurrencia del recurso, pero no solape de agenda personal.
5. Las tres canchas grupales exigen mínimo de 10 participantes; el propietario cuenta y no retira su propia participación.
6. La solicitud grupal bloquea el horario desde `PENDING` y confirma al alcanzar el mínimo.
7. El código se obtiene bajo demanda, solo por el propietario y en estados permitidos.
8. Talleres e inscripciones propias pertenecen a MVP 2. El solape se evalúa solo taller↔taller; no taller↔reserva personal entre recursos. Alta y baja se permiten mientras el taller esté activo, sin cutoff.
9. La cancelación actual se permite antes de finalizar la reserva; un cutoff configurable queda como mejora futura.
10. Administración, inventario, bloqueos, programación, prioridad, notificación específica de prioridad, pantalla pública sanitizada e historial institucional pertenecen a MVP 3.
11. Calidad, soporte, sistema core de notificaciones, reportes básicos, auditoría y despliegue pertenecen a MVP 4.
12. `007` y `008` son prospectivas y no reescriben reservas históricas.
13. La migración `009` es una propuesta no aprobada: no forma parte de la secuencia productiva vigente ni puede declararse desplegada.

## 7. Fuera del alcance del prototipo

- Business Intelligence y reportes avanzados;
- analítica predictiva o inteligencia artificial;
- gestión avanzada de campeonatos;
- detección automatizada de abuso;
- integración productiva con sistemas académicos;
- operación multisede;
- sincronización bidireccional con Google Calendar;
- control físico de acceso, pagos o aplicación móvil nativa.

## 8. Pendientes que impiden el cierre

### Cierre MVP 1–2

- completar la integración frontend del contrato de disponibilidad por rango;
- ejecutar `007` y `008` con backup, precheck, postcheck, idempotencia, recuperación y concurrencia;
- validar online Microsoft Entra ID, CORS, API y Azure SQL;
- aprobar los checklists MVP 1 y MVP 2;
- ejecutar QA visual, accesibilidad y privacidad por audiencia;
- registrar una nueva evidencia con fecha, versión y ambiente.

### Cierre MVP 3–4

- completar operación administrativa sin escritura directa en base de datos;
- resolver prioridad y conflictos institucionales con auditoría y su notificación específica;
- completar pantalla pública sanitizada e historial institucional sin inferir asistencia personal;
- cerrar sistema core de notificaciones, reportes básicos, infracciones y consulta de auditoría;
- consolidar seguridad, E2E, despliegue, recuperación, documentación y defensa.

## 9. Dictamen recomendado

> MVP 1 es demostrable localmente, pero no está cerrado online. MVP 2 permanece parcial y no cerrado por la integración frontend de disponibilidad, las migraciones `007`/`008` y la validación Azure. MVP 3 es parcial. MVP 4 está pendiente y acotado al soporte necesario del prototipo. El cierre total debe ocurrir antes del 2026-12-10 y demostrarse mediante el checklist canónico.
