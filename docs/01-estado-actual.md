# Estado actual de Poli-REDI

**Audiencia:** dirección, tesis, producto y responsables de cierre.

**Propósito:** presentar un único dictamen vigente sin duplicar requisitos ni planificación.

**Estado:** corte 2026-08-11.

**Fuente propietaria:** este archivo gobierna el estado; la evidencia detallada está en [`anexos/evidencia/auditoria-2026-08-11.md`](anexos/evidencia/auditoria-2026-08-11.md).

## Resumen ejecutivo

Poli-REDI es una demo funcional avanzada verificada principalmente en local. No existe evidencia vigente de frontend, Microsoft Entra ID, CORS, API y Azure SQL validados juntos después de los cambios recientes.

El backend de disponibilidad por rango existe. Su integración completa en frontend, las migraciones `007`/`008` y el flujo Azure siguen condicionando el cierre de MVP 1–2.

## Estado por MVP

| MVP | Alcance | Estado | Bloqueador principal |
|---|---|---|---|
| MVP 1 | Base, identidad, disponibilidad y reserva básica | Demostrable localmente; no cerrado online | Integración frontend del rango y smoke integrado. |
| MVP 2 | Usuario, reserva grupal y talleres | Parcial; no cerrado | `007`/`008`, concurrencia y E2E Azure. |
| MVP 3 | Administración, programación, prioridad, pantalla pública e historial institucional | Parcial | Operaciones administrativas y flujos institucionales incompletos. |
| MVP 4 | Calidad, soporte, notificaciones core, reportes básicos, auditoría y despliegue | Pendiente y acotado | Contratos, QA, seguridad y candidato desplegable. |

Las fechas y dependencias pertenecen al [`plan de entrega`](08-plan-entrega-2026.md).

## Estado por capacidad crítica

| Capacidad | Dictamen al corte |
|---|---|
| Autenticación, rol y RUT | Implementados localmente; Entra online pendiente. |
| Reserva particular | Implementada localmente; E2E online pendiente. |
| Disponibilidad | Backend por rango implementado; frontend parcial. |
| Flujo grupal | Implementado parcialmente; migraciones y Azure pendientes. |
| Talleres | Alcance funcional local observable; integración online pendiente. |
| Cancelación | Permitida antes de finalizar; cutoff configurable es futuro. |
| Administración | Lecturas básicas; gestión completa pendiente. |
| Calidad y accesibilidad | Mejoras locales; matriz manual y online pendientes. |
| Despliegue | Configurado parcialmente; no revalidado en este corte. |

## Evidencia disponible

La auditoría revalidó **18 pruebas Node** sobre el commit `939ba51`.

No obtuvo nueva evidencia concluyente de:

- suite completa de Vitest;
- pruebas Go;
- build frontend;
- migraciones Azure SQL;
- integración Entra ID/CORS/API/DB;
- flujo online.

Los resultados del 2026-08-04 se conservan como evidencia histórica y no se presentan como nueva ejecución.

## Decisiones aplicables

- [`ADR-001`](decisiones/ADR-001-gobierno-documental.md): gobierno documental.
- [`ADR-002`](decisiones/ADR-002-evolucion-base-unica.md): una base y migraciones aprobadas.
- [`ADR-003`](decisiones/ADR-003-alcance-mvp-y-exclusiones.md): alcance MVP y exclusiones.
- [`ADR-004`](decisiones/ADR-004-reglas-temporales-y-solapes.md): reglas temporales y solapes.

## Dictamen

> MVP 1 es demostrable localmente, pero no está cerrado online. MVP 2 permanece parcial. MVP 3 es parcial y MVP 4 está pendiente. Ningún MVP debe declararse cerrado sin satisfacer los gates del checklist y conservar evidencia del ambiente correspondiente.

Las acciones pendientes se mantienen exclusivamente en [`08-plan-entrega-2026.md`](08-plan-entrega-2026.md).
