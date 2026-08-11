# Poli-REDI — documentación canónica vigente

**Estado:** CANÓNICO ADOPTADO

**Corte documental:** 2026-08-11

**Autoridad:** decisión del Orquestador Principal de Poli-REDI del 2026-08-11
**Fecha límite del prototipo:** 2026-12-10 (`America/Santiago`)

## Cómo usar este paquete

1. Comenzar por [`01-resumen-ejecutivo-y-estado.md`](01-resumen-ejecutivo-y-estado.md).
2. Consultar [`00-indice-y-gobierno-documental.md`](00-indice-y-gobierno-documental.md) para resolver precedencia, estados y trazabilidad.
3. Usar los documentos `02` a `14` según el trabajo técnico, de calidad, planificación o entrega.
4. Tratar `historico/`, `referencia/`, `../historico_y_checklists/` y `../../Documentos/Entregables_Previos/` como evidencia temporal, no como estado vigente.
5. Los documentos anteriores `../00-indice-maestro-y-trazabilidad.md` a `../04-guias-y-despliegue.md` quedan supersedidos desde este corte. Se preservan para no perder trazabilidad ni romper enlaces existentes.
6. No declarar un MVP cerrado sin cumplir su criterio de salida y registrar ambiente, versión, pasos, resultado y limitaciones.

## Documentos canónicos

| Documento | Uso principal |
|---|---|
| [`00-indice-y-gobierno-documental.md`](00-indice-y-gobierno-documental.md) | Autoridad, precedencia, índice, trazabilidad y mantenimiento. |
| [`01-resumen-ejecutivo-y-estado.md`](01-resumen-ejecutivo-y-estado.md) | Estado real del producto y límites de la evidencia. |
| [`02-arquitectura-y-contratos.md`](02-arquitectura-y-contratos.md) | Arquitectura y contratos entre frontend, backend y base de datos. |
| [`03-requisitos-casos-uso-y-trazabilidad.md`](03-requisitos-casos-uso-y-trazabilidad.md) | Requisitos, actores, historias, casos de uso y asignación final de MVP. |
| [`04-base-de-datos-y-migraciones.md`](04-base-de-datos-y-migraciones.md) | Modelo de datos, integridad y operación de migraciones. |
| [`05-instalacion-despliegue-y-recuperacion.md`](05-instalacion-despliegue-y-recuperacion.md) | Ejecución local, Azure, redespliegue y recuperación. |
| [`06-flujos-y-reglas-de-negocio.md`](06-flujos-y-reglas-de-negocio.md) | Reservas, grupos, talleres, cancelación y conflictos. |
| [`07-calidad-pruebas-y-checklists.md`](07-calidad-pruebas-y-checklists.md) | Evidencia automatizada y pruebas manuales. |
| [`08-backlog-priorizado.md`](08-backlog-priorizado.md) | Pendientes reales, prioridades y definición de terminado. |
| [`09-plan-corte-google-calendar.md`](09-plan-corte-google-calendar.md) | Transición desde la operación legada. |
| [`10-guia-documentacion-y-mantenimiento.md`](10-guia-documentacion-y-mantenimiento.md) | Reglas para mantener documentación y código legibles. |
| [`11-cronograma-cierre-2026.md`](11-cronograma-cierre-2026.md) | Cronograma obligatorio, ruta crítica, responsables y buffers hasta el 10-12-2026. |
| [`12-auditoria-alcance-implementacion-2026-08-11.md`](12-auditoria-alcance-implementacion-2026-08-11.md) | Auditoría integral de alcance e implementación incorporada por el rol responsable. |
| [`13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md`](13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md) | Diagramas canónicos de arquitectura, datos y flujos de MVP 1–2 incorporados por el rol responsable. |
| [`14-checklist-cierre-total-2026-12-10.md`](14-checklist-cierre-total-2026-12-10.md) | Checklist único de cierre por MVP, ambiente y evidencia. |

## Estado del producto en este corte

- **MVP 1:** demostrable localmente; la validación online vigente no está acreditada.
- **MVP 2:** parcial y no cerrado. El backend de disponibilidad por rango existe, pero su integración frontend está pendiente; también faltan `007`/`008` y validación Azure.
- **MVP 3:** parcial; incluye administración, programación, pantalla pública sanitizada y la notificación específica asociada a prioridad.
- **MVP 4:** pendiente y acotado al prototipo; contiene el sistema core de notificaciones.
- **Migración `009`:** propuesta no aprobada y fuera de la secuencia productiva vigente.

El acta que adopta este canon es [`historico/16-acta-revision-documental-2026-08-11.md`](historico/16-acta-revision-documental-2026-08-11.md). La adopción documental no ejecutó la suite completa, migraciones, Azure, Microsoft Entra ID ni pruebas online. La auditoría [`12-auditoria-alcance-implementacion-2026-08-11.md`](12-auditoria-alcance-implementacion-2026-08-11.md) revalidó **18 pruebas Node** sobre el commit `939ba51`; no aportó nueva evidencia de Vitest, Go, build frontend ni flujo online.
