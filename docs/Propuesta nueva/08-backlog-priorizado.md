# Backlog priorizado de Poli-REDI

**Estado:** CANÓNICO ADOPTADO

**Corte de revisión:** 2026-08-11

**Fecha límite absoluta:** 2026-12-10
**Fuente detallada preservada:** [`referencia/07-backlog-completo.md`](referencia/07-backlog-completo.md)

## 1. Regla de lectura

El backlog representa trabajo pendiente o evidencia por completar. Un ítem solo se cierra cuando alcance, implementación, pruebas, documentación y ambiente objetivo quedan registrados. “Funciona en local” no equivale a “cerrado online”.

## 2. Bloque de cierre MVP 1 — objetivo 2026-08-28

| ID | Trabajo | Responsable principal | Criterio de salida |
|---|---|---|---|
| M1-BASE | Congelar baseline funcional y de contratos | Arquitecto | Contratos, dependencias y riesgos registrados. |
| M1-DISP-FE | Integrar en frontend la disponibilidad por rango ya existente en backend | Frontend | Carga, filtros, selección, errores y estados vacíos verificados. |
| M1-REG | Ejecutar regresión de reserva básica e identidad | QA | Casos críticos locales reproducibles sin regresiones abiertas. |
| M1-ONLINE | Preparar integración Entra ID, API, CORS y Azure SQL | DevOps | Configuración y plan de despliegue/reversión listos para validación. |
| M1-EVID | Consolidar evidencia y dictamen | Documentador | Checklist con ambiente, fecha y resultado; sin afirmar online si no se ejecutó. |

## 3. Bloque de cierre MVP 2 — objetivo 2026-09-25

| ID | Trabajo | Responsable principal | Criterio de salida |
|---|---|---|---|
| DB-007 | Ensayar y ejecutar migración `007` | Backend / DevOps | Backup, pre/postcheck, idempotencia, no retroactividad y reversión. |
| DB-008 | Ensayar y ejecutar migración `008` | Backend / DevOps | Solape bidireccional, contigüidad e idempotencia verificados. |
| M2-GRUPO | Cerrar flujo grupal y privacidad del código | Backend / Frontend | Crear, consultar, confirmar, retirar, editar objetivo, cancelar antes de finalizar y expirar con mensajes consistentes. El cutoff de cancelación configurable queda como mejora futura. |
| M2-TALLER | Cerrar talleres y conflictos entre talleres | Backend / Frontend | Alta, baja y reinscripción mientras el taller esté activo, sin cutoff; cupo, solape taller↔taller e historial propio E2E. No imponer taller↔reserva personal entre recursos. |
| M2-AZURE | Validar MVP 1–2 en ambiente Azure | QA / DevOps | Entra ID, CORS, API, DB y frontend integrados con evidencia. |
| M2-BUFFER | Resolver defectos de cierre | QA | Sin defectos críticos o altos abiertos; riesgos residuales aceptados. |

## 4. Bloque de cierre MVP 3 — objetivo 2026-10-30

| ID | Trabajo | Responsable principal | Criterio de salida |
|---|---|---|---|
| M3-CONTR | Congelar contratos administrativos | Analista / Arquitecto | Permisos, estados, prioridad, filtros y auditoría definidos. |
| M3-INV | Completar inventario y modos de recursos | Backend / Frontend | Alta/edición/estado con validación y trazabilidad. |
| M3-USR | Completar usuarios y bloqueos | Backend / Frontend | Operaciones autorizadas, minimizadas y auditadas. |
| M3-PROG | Completar bloqueos, programación y prioridad institucional | Backend / Frontend | Conflictos deterministas, aplicación prospectiva y notificación específica de prioridad verificables. |
| M3-PUBLIC | Completar pantalla pública de disponibilidad | Frontend / Backend | Información institucional sanitizada, sin datos personales ni acciones privadas. |
| M3-HIST | Completar historial institucional | Backend / Frontend | Reservas, talleres, clases y otros eventos distinguibles sin inferir asistencia. |
| M3-E2E | Ejecutar E2E administrativo | QA | Matriz rol/acción, concurrencia y auditoría aprobadas. |

## 5. Bloque de cierre MVP 4 — objetivo 2026-11-27

| ID | Trabajo | Responsable principal | Criterio de salida |
|---|---|---|---|
| M4-NOTIF | Completar sistema core de notificaciones | Backend / Frontend | Eventos generales, destinatarios, lectura y acciones observables; la notificación específica de prioridad se cierra en MVP 3. |
| M4-REPORT | Completar reportes básicos | Backend / Frontend | Indicadores definidos y derivados desde fuentes controladas. |
| M4-AUDIT | Habilitar consulta protegida de auditoría | Backend / Frontend | Permisos, filtros, minimización y paginación verificados. |
| M4-SEC | Revisar seguridad y privacidad | Arquitecto / QA | Autorización, secretos, exposición de datos y dependencias revisados. |
| M4-UX | Ejecutar QA visual y accesibilidad | Diseñador UX / QA | Responsive, teclado, foco, contraste, lector y estados de carga aprobados. |
| M4-DEPLOY | Preparar candidato y despliegue final | DevOps | Despliegue reproducible, monitoreo básico, rollback y smoke test. |

## 6. Entrega y defensa — fecha absoluta 2026-12-10

| ID | Trabajo | Responsable principal | Criterio de salida |
|---|---|---|---|
| DOC-INTEG | Integrar memoria, diagramas, anexos y evidencia | Documentador | Trazabilidad completa y referencias vigentes. |
| QA-FINAL | Ejecutar smoke y regresión final | QA | Resultado fechado del candidato de entrega. |
| DEFENSA | Preparar demostración y defensa | Analista / Documentador | Guion, evidencia, contingencia y ensayo completados. |
| ENTREGA | Entregar prototipo y documentación | Documentador | Paquete recibido antes o durante el 2026-12-10. |

## 7. Deuda y decisiones controladas

- eliminar referencias vigentes a motores o contratos obsoletos fuera del histórico;
- mantener sincronizados endpoints, requisitos, casos de uso, diagramas y pruebas;
- registrar ambiente, fecha y resultado de cada verificación;
- evitar estados absolutos como “cerrado” sin evidencia integrada;
- revisar enlaces Markdown y codificación UTF-8;
- mantener `009` como **propuesta no aprobada** hasta una decisión formal; no ejecutarla por inferencia;
- conservar los documentos históricos sin reescribir su contexto original.

## 8. Fuera de alcance del prototipo

- BI avanzado;
- inteligencia artificial;
- gestión avanzada de campeonatos;
- detección automatizada de abuso;
- integración académica;
- multisede;
- sincronización bidireccional con Google.

## 9. Definición de terminado

Una tarea se considera terminada cuando:

- el alcance y responsable por rol están explícitos;
- la implementación existe en la rama o versión evaluada;
- las pruebas aplicables pasan y conservan evidencia;
- la documentación y diagramas relacionados están actualizados;
- no se exponen secretos ni datos indebidos;
- el ambiente objetivo fue verificado cuando corresponde;
- los riesgos residuales y la decisión de aceptación están registrados.

El orden semanal, buffers y ruta crítica se detallan en [`11-cronograma-cierre-2026.md`](11-cronograma-cierre-2026.md). El cierre se controla con [`14-checklist-cierre-total-2026-12-10.md`](14-checklist-cierre-total-2026-12-10.md).
