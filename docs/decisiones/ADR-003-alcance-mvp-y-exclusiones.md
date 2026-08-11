# ADR-003 — Alcance de MVP y exclusiones

**Audiencia:** Analista, Arquitecto, equipo de entrega y evaluación académica

**Propósito:** fijar la asignación final sin ampliar implícitamente el prototipo

**Estado:** aceptada el 2026-08-11

**Fuente:** alcance académico definitivo y revisión integral

## Resumen

El prototipo se divide en cuatro MVP acumulativos. MVP 1 entrega la base; MVP 2 completa flujos de usuario, grupo y talleres; MVP 3 incorpora operación institucional; MVP 4 cierra calidad, soporte y despliegue. La fecha absoluta y los hitos viven en [08-plan-entrega-2026.md](../08-plan-entrega-2026.md).

## Decisión

| MVP | Alcance obligatorio |
|---|---|
| MVP 1 | Base técnica, identidad, RUT condicionado, catálogo, disponibilidad, reserva básica e historial propio de reservas. |
| MVP 2 | Flujo de usuario, reservas grupales, código, participantes, políticas, talleres e historial propio de inscripciones. |
| MVP 3 | Administración, usuarios, inventario, bloqueos, programación, prioridad, notificación específica de prioridad, pantalla pública e historial institucional. |
| MVP 4 | Calidad, sistema core de notificaciones, reportes básicos, auditoría, seguridad, accesibilidad y despliegue recuperable. |

El historial de clases y otros eventos institucionales corresponde a MVP 3. La notificación específica causada por prioridad depende de MVP 3; la infraestructura general de notificaciones corresponde a MVP 4.

## Exclusiones

Quedan fuera del prototipo:

- BI y analítica avanzada;
- IA y detección automatizada de abuso;
- integración con sistemas académicos;
- operación multisede completa;
- sincronización bidireccional con Google Calendar;
- gestión avanzada de campeonatos;
- funcionalidades no trazadas que desplacen seguridad, evidencia o entrega.

## Control de cambios

Un cambio posterior debe indicar requisito afectado, MVP, dependencia, riesgo, prueba y trabajo equivalente retirado. No se incorpora una función solo porque exista un borrador técnico o una migración propuesta.

## Consecuencias

- MVP 2 no se declara cerrado por código parcial: requiere `007`/`008` y E2E Azure.
- MVP 3 y MVP 4 quedan deliberadamente acotados.
- Las exclusiones pueden convertirse en trabajo futuro mediante una nueva decisión, no durante el cierre actual.
