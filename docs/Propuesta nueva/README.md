# Poli-REDI — Documentación canónica mejorada

**Corte documental:** 2026-08-04  
**Propósito:** entregar una única estructura navegable, trazable y segura para comprender, desarrollar, probar y desplegar Poli-REDI.

## Cómo usar este paquete

1. Comenzar por [`01-resumen-ejecutivo-y-estado.md`](01-resumen-ejecutivo-y-estado.md).
2. Consultar [`00-indice-y-gobierno-documental.md`](00-indice-y-gobierno-documental.md) para saber cuál documento manda cuando existen diferencias.
3. Usar los documentos `02` a `10` según el trabajo técnico o académico.
4. Tratar `historico/` como evidencia temporal, no como estado vigente.
5. No declarar MVP 2 cerrado hasta completar la validación integrada y las migraciones pendientes sobre Azure SQL.

## Documentos canónicos

| Documento | Uso principal |
|---|---|
| [`00-indice-y-gobierno-documental.md`](00-indice-y-gobierno-documental.md) | Índice, jerarquía, trazabilidad y mantenimiento. |
| [`01-resumen-ejecutivo-y-estado.md`](01-resumen-ejecutivo-y-estado.md) | Estado real del producto y pendientes. |
| [`02-arquitectura-y-contratos.md`](02-arquitectura-y-contratos.md) | Arquitectura, límites y contratos entre capas. |
| [`03-requisitos-casos-uso-y-trazabilidad.md`](03-requisitos-casos-uso-y-trazabilidad.md) | Requisitos, actores, HU, CU y relación con MVP. |
| [`04-base-de-datos-y-migraciones.md`](04-base-de-datos-y-migraciones.md) | Modelo de datos, integridad y operación de migraciones. |
| [`05-instalacion-despliegue-y-recuperacion.md`](05-instalacion-despliegue-y-recuperacion.md) | Ejecución local, Azure, redeploy y recuperación. |
| [`06-flujos-y-reglas-de-negocio.md`](06-flujos-y-reglas-de-negocio.md) | Reservas, grupos, talleres, cancelación y conflictos. |
| [`07-calidad-pruebas-y-checklists.md`](07-calidad-pruebas-y-checklists.md) | Evidencia automatizada y pruebas manuales. |
| [`08-backlog-priorizado.md`](08-backlog-priorizado.md) | Pendientes reales, prioridades y criterio de cierre. |
| [`09-plan-corte-google-calendar.md`](09-plan-corte-google-calendar.md) | Transición desde la operación legada. |
| [`10-guia-documentacion-y-mantenimiento.md`](10-guia-documentacion-y-mantenimiento.md) | Reglas para mantener documentación y código legibles. |

## Estado de este paquete

Esta reorganización se basa exclusivamente en los archivos suministrados. La actualización más reciente documentada es la desinscripción de talleres verificada localmente el 2026-08-04. No se realizó una nueva ejecución del código, Azure SQL, Microsoft Entra ID ni la demo online durante la generación de este paquete.

El detalle de cambios está en [`docs/CHANGELOG.md`](docs/CHANGELOG.md) y la validación estructural en [`docs/VALIDATION.md`](docs/VALIDATION.md).
