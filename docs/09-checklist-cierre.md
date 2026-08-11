# Checklist de cierre

**Audiencia:** responsables de cierre por rol y evaluación académica

**Propósito:** decidir si cada MVP y la entrega tienen evidencia suficiente

**Estado:** canónico; ninguna casilla se presume completada

**Fuente:** requisitos, auditoría y plan de entrega con corte 2026-08-11

## Resumen

Marcar una casilla solo con evidencia fechada, atribuida a ambiente y versión. Los hitos y fechas son propiedad de [08-plan-entrega-2026.md](08-plan-entrega-2026.md). Si existe un P0/P1, fuga de datos, permiso incorrecto o migración no recuperable, el cierre no es aprobable.

## Gobierno y alcance

- [ ] El canon está enlazado desde el repositorio y tiene propietario por tema.
- [ ] Requisitos, MVP, exclusiones, contratos y evidencia son consistentes.
- [ ] `009` consta como propuesta no aprobada y no ejecutada.
- [ ] BI avanzado, IA, integración académica, multisede, Google bidireccional, campeonatos avanzados y detección automatizada de abuso permanecen fuera.
- [ ] Cada cambio posterior al freeze tiene decisión e impacto registrado.

## MVP 1 — base

- [ ] Identidad, roles y bloqueo funcionan según contrato.
- [ ] El usuario normal completa RUT solo cuando corresponde.
- [ ] El administrador puede agendar sin RUT según la regla aprobada.
- [ ] La disponibilidad por rango está integrada en frontend.
- [ ] Carga, actualización, vacío, error y reintento son consistentes.
- [ ] La reserva valida recurso, fecha, hora, duración, actividad, participantes y solapes.
- [ ] Crear, consultar y cancelar antes de finalizar produce resultados coherentes.
- [ ] La regresión local está registrada.
- [ ] La validación online se ejecutó o el acta declara expresamente su ausencia.
- [ ] Existe acta de cierre contra el hito del plan.

## MVP 2 — usuario, grupo y talleres

### Reserva grupal

- [ ] El formulario solicita objetivo/cantidad con valor inicial aprobado.
- [ ] El propietario queda incluido y la solicitud permanece pendiente bajo el mínimo.
- [ ] El progreso y el cambio automático de estado son coherentes.
- [ ] El código solo se muestra a audiencias autorizadas.
- [ ] La consulta no distingue indebidamente código inexistente, cancelado o vencido.
- [ ] Confirmar, retirar, rotar código, expirar y editar objetivo respetan permisos y plazo.
- [ ] La cancelación se permite por rol autorizado antes de finalizar; no se exige cutoff configurable.
- [ ] `open_use` no tiene restricción semanal, permite contigüidad y rechaza solape personal.

### Talleres

- [ ] Alta permitida mientras el taller está activo, con cupo y RUT cuando corresponde.
- [ ] Se rechaza otro taller activo solapado.
- [ ] No se impone taller↔reserva personal entre recursos en MVP 2.
- [ ] La baja se permite mientras el taller está activo, sin cutoff adicional.
- [ ] La baja repetida es idempotente y libera cupo y solape.
- [ ] La reinscripción crea un episodio nuevo y conserva el anterior cancelado.
- [ ] Las inscripciones propias aparecen en historial sin inferir asistencia.

### Datos e integración

- [ ] `007` pasó backup, precheck, ejecución, postcheck, idempotencia y recuperación.
- [ ] `008` pasó backup, precheck, ejecución, postcheck, idempotencia y recuperación.
- [ ] `007` y `008` fueron verificadas en Azure SQL.
- [ ] Entra ID, CORS, frontend, API y DB pasaron E2E.
- [ ] No quedan defectos críticos o altos abiertos.
- [ ] Existe acta de cierre contra el hito del plan.

## MVP 3 — administración institucional

- [ ] La matriz de permisos por rol está implementada y probada.
- [ ] Usuarios, motivos y bloqueos se gestionan sin exposición indebida.
- [ ] Inventario, modos de recursos y disponibilidad se gestionan con auditoría.
- [ ] Talleres, clases y otros eventos se programan y diferencian.
- [ ] La prioridad institucional resuelve conflictos de forma determinista.
- [ ] La notificación específica de prioridad está integrada.
- [ ] La pantalla pública muestra datos sanitizados y no ofrece acciones privadas.
- [ ] Los cambios administrativos son prospectivos salvo excepción auditada.
- [ ] El historial institucional distingue reserva, inscripción y asistencia acreditada.
- [ ] Permisos, conflictos y concurrencia tienen E2E.
- [ ] Existe acta de cierre contra el hito del plan.

## MVP 4 — calidad, soporte y despliegue

- [ ] Notificaciones core cubren eventos, destinatarios, lectura y acción.
- [ ] Reportes básicos tienen definición y fuente controlada.
- [ ] La auditoría es consultable solo por roles autorizados.
- [ ] Los errores son claros y no exponen detalles internos.
- [ ] Autorización, secretos, dependencias y configuración pasaron revisión.
- [ ] Cada audiencia recibe solo los datos necesarios.
- [ ] QA visual cubre 377, 500, 768 y 1440 px o matriz equivalente.
- [ ] Teclado, foco, contraste, lector y movimiento reducido fueron revisados.
- [ ] Skeletons, estados vacíos y errores son consistentes.
- [ ] Despliegue y rollback son reproducibles.
- [ ] El candidato pasó smoke en Azure.
- [ ] Existe acta de cierre contra el hito del plan.

## Ambientes y evidencia

### Local

- [ ] Frontend, backend y esquema están identificados.
- [ ] Pruebas y build aplicables pasan.
- [ ] Flujos manuales críticos tienen evidencia segura.

### Copia recuperable

- [ ] Origen, fecha, responsable y recuperación están registrados.
- [ ] Migraciones requeridas son idempotentes.
- [ ] Integridad fue comprobada antes y después.

### Azure

- [ ] Frontend, API, Entra y DB corresponden a la misma versión.
- [ ] Secretos no están en repositorio, respuestas, logs ni capturas.
- [ ] Inventario de migraciones coincide con la base.
- [ ] Smoke, E2E y rollback tienen resultado fechado.
- [ ] Cuotas y ventana del ambiente están controladas.

## Documentación y entrega

- [ ] Índice, estado, requisitos, arquitectura, datos, reglas y plan están sincronizados.
- [ ] Diagramas coinciden con contratos y despliegue finales.
- [ ] La memoria distingue evidencia local, integrada y online.
- [ ] Los históricos se preservan y las correcciones usan actas o ADR.
- [ ] Enlaces, UTF-8, whitespace y formato final están validados.
- [ ] El guion de demo usa datos seguros y tiene contingencia.
- [ ] La regresión final y el ensayo de defensa están registrados.
- [ ] El paquete fue entregado y su recepción quedó registrada.

## Dictamen

- [ ] **APROBABLE:** criterios obligatorios completos y riesgos residuales aceptados.
- [ ] **APROBABLE CON OBSERVACIONES:** sin brechas críticas de objetivo, seguridad o defensa.
- [ ] **NO APROBABLE:** existe una brecha crítica de alcance, integridad, seguridad, ambiente o evidencia.

**Responsable por rol:** ____________________

**Fecha:** ____________________

**Acta/evidencia:** ____________________
