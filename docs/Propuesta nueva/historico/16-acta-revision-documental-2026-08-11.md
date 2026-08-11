# Acta 16 — Revisión documental y adopción de canon

**Fecha:** 2026-08-11

**Estado:** DECISIÓN REGISTRADA
**Ámbito:** gobierno documental, alcance y planificación de cierre.

## 1. Antecedente

La documentación del proyecto coexistía entre una serie raíz anterior, documentos académicos, entregables previos y la carpeta denominada `docs/Propuesta nueva/`. Esa coexistencia permitía interpretar como vigente material con fechas, alcance o estados distintos.

## 2. Decisión

El Orquestador adopta formalmente `docs/Propuesta nueva/` como canon vigente desde el 2026-08-11.

Se conserva temporalmente el nombre de la carpeta para no romper enlaces. Los documentos raíz anteriores `docs/00` a `docs/04` quedan supersedidos como fuente operativa, pero se preservan. Los históricos y `Entregables_Previos` mantienen su contexto original y no deben reescribirse.

## 3. Estado reconocido al corte

- MVP 1: demostrable en local; no cerrado online.
- MVP 2: parcial y no cerrado; faltan integración frontend de disponibilidad, `007`/`008` y evidencia Azure.
- MVP 3: parcial.
- MVP 4: pendiente y acotado.
- Disponibilidad por rango: backend existente; integración frontend pendiente.
- Migración `009`: propuesta no aprobada y fuera de ejecución.

## 4. Asignación final de alcance

- **MVP 1:** base técnica, identidad, persistencia, disponibilidad y reserva básica.
- **MVP 2:** usuario, reserva grupal y talleres.
- **MVP 3:** administración, inventario, bloqueos, programación, prioridad, notificación específica asociada, pantalla pública sanitizada e historial institucional.
- **MVP 4:** calidad, soporte, sistema core de notificaciones, reportes básicos, auditoría y despliegue.

Quedan fuera del prototipo BI avanzado, IA, integración académica, multisede, sincronización bidireccional con Google, gestión avanzada de campeonatos y detección automatizada de abuso.

## 5. Fechas de gobierno

| Entrega | Fecha objetivo |
|---|---:|
| MVP 1 | 2026-08-28 |
| MVP 2 | 2026-09-25 |
| MVP 3 | 2026-10-30 |
| MVP 4 / feature freeze | 2026-11-27 |
| Entrega total | 2026-12-10 |

## 6. Documentos afectados

- índice y gobierno documental;
- resumen ejecutivo y estado;
- requisitos, casos de uso y trazabilidad;
- backlog priorizado;
- cronograma de cierre;
- checklist de cierre total;
- índice maestro raíz y README del repositorio;
- alcance definitivo y backlog maestro de tesis.

Los documentos [`../12-auditoria-alcance-implementacion-2026-08-11.md`](../12-auditoria-alcance-implementacion-2026-08-11.md) y [`../13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md`](../13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md) se integran al canon por trabajos coordinados.

## 7. Regla de continuidad

Toda corrección posterior debe crear una nueva revisión o acta. No se alterarán actas previas para cambiar retrospectivamente su dictamen. Un estado solo se considerará cerrado cuando tenga evidencia fechada del ambiente correspondiente.

## 8. Resultado

Se elimina la ambigüedad de autoridad documental y se fija una ruta de cierre con fecha límite absoluta 2026-12-10. Esta acta no declara ejecutadas pruebas, migraciones ni despliegues pendientes.
