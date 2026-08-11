# Checklist de cierre total — 2026-12-10

**Estado:** CANÓNICO ADOPTADO

**Corte inicial:** 2026-08-11
**Regla:** las casillas permanecen sin marcar hasta que exista evidencia fechada y atribuida a un ambiente.

## 1. Gobierno y alcance

- [ ] El canon documental vigente está enlazado desde el repositorio.
- [ ] La asignación final de MVP 1–4 coincide en alcance, backlog y cronograma.
- [ ] Las exclusiones del prototipo están registradas.
- [ ] Gestión avanzada de campeonatos y detección automatizada de abuso permanecen fuera del prototipo.
- [ ] La migración `009` consta como propuesta no aprobada y no ejecutada.
- [ ] Cada decisión nueva tiene acta y trazabilidad.

## 2. MVP 1 — base

- [ ] Identidad, roles y bloqueo funcionan según contrato.
- [ ] RUT se solicita solo cuando corresponde y el administrador puede agendar sin RUT conforme a la regla aprobada.
- [ ] Disponibilidad por rango del backend está integrada en frontend.
- [ ] Los estados de carga, vacío, error y reintento son consistentes.
- [ ] La reserva básica valida fecha, hora, duración, recurso, participantes y solapes.
- [ ] Crear, consultar y cancelar una reserva antes de su finalización produce resultados coherentes; no se exige cutoff configurable.
- [ ] La regresión local de MVP 1 está documentada.
- [ ] La integración online de MVP 1 está documentada; si no fue ejecutada, el MVP no se declara online.
- [ ] Existe acta de cierre de MVP 1 con fecha máxima 2026-08-28.

## 3. MVP 2 — usuario, grupo y talleres

### Reserva grupal

- [ ] El formulario solicita objetivo/cantidad de participantes con valor inicial aprobado.
- [ ] La reserva comienza pendiente cuando no alcanza el mínimo.
- [ ] El propietario queda incluido y puede consultar el progreso.
- [ ] El código se muestra solo a audiencias autorizadas.
- [ ] La consulta por código no filtra información de reservas inexistentes, canceladas o vencidas.
- [ ] Confirmar, retirar y editar objetivo respetan sus permisos y plazo; cancelar se permite antes de finalizar según permisos. El cutoff de cancelación configurable queda como mejora futura.
- [ ] Las reservas contiguas de uso libre son válidas y los solapes personales se rechazan.

### Talleres

- [ ] El usuario puede inscribirse mientras el taller está activo si hay cupo y no tiene otro taller activo solapado.
- [ ] El usuario puede desinscribirse mientras el taller está activo, sin cutoff adicional.
- [ ] La reinscripción conserva el episodio histórico cancelado.
- [ ] MVP 2 no compara taller↔reserva personal entre recursos; el control de solape se limita a taller↔taller.
- [ ] Las inscripciones a talleres aparecen correctamente en el historial propio.

### Datos e integración

- [ ] `007` pasó preflight, ejecución, postcheck e idempotencia en copia recuperable.
- [ ] `008` pasó preflight, ejecución, postcheck e idempotencia en copia recuperable.
- [ ] `007` y `008` fueron ejecutadas y verificadas en Azure SQL.
- [ ] Entra ID, CORS, frontend, API y DB fueron validados de extremo a extremo.
- [ ] No quedan defectos críticos o altos abiertos.
- [ ] Existe acta de cierre de MVP 2 con fecha máxima 2026-09-25.

## 4. MVP 3 — administración institucional

- [ ] La matriz de permisos por rol está implementada y probada.
- [ ] Inventario y modos de recursos pueden gestionarse con auditoría.
- [ ] Usuarios y bloqueos pueden gestionarse sin exponer datos indebidos.
- [ ] Bloqueos de disponibilidad pueden crearse, consultarse y modificarse según permisos.
- [ ] Talleres, clases y otros eventos pueden programarse y diferenciarse.
- [ ] La prioridad institucional resuelve conflictos de forma determinista.
- [ ] La notificación específica asociada a prioridad está integrada y probada.
- [ ] La pantalla pública presenta disponibilidad sanitizada, sin datos personales ni acciones privadas.
- [ ] Los cambios administrativos son prospectivos salvo excepción explícita y auditada.
- [ ] El historial institucional distingue reserva, inscripción y asistencia; no infiere asistencia sin evidencia.
- [ ] Conflictos, concurrencia y permisos tienen pruebas E2E.
- [ ] Existe acta de cierre de MVP 3 con fecha máxima 2026-10-30.

## 5. MVP 4 — calidad, soporte y despliegue

- [ ] El sistema core de notificaciones tiene eventos generales, destinatarios, estado de lectura y acción; no duplica el criterio específico de prioridad de MVP 3.
- [ ] Los reportes básicos tienen definición y fuente de datos controlada.
- [ ] La auditoría es consultable solo por roles autorizados y con filtros.
- [ ] Los errores de usuario son claros y no exponen detalles internos.
- [ ] Autorización, secretos, dependencias y configuración pasaron revisión de seguridad.
- [ ] Se verificó minimización de datos en cada audiencia.
- [ ] El QA visual cubre 377, 500, 768 y 1440 px o una matriz equivalente aprobada.
- [ ] Teclado, foco, contraste, lector de pantalla y movimiento reducido fueron revisados.
- [ ] Skeletons, estados vacíos y errores están implementados en las vistas acordadas.
- [ ] El despliegue es reproducible y tiene rollback documentado.
- [ ] El candidato pasó smoke test en Azure.
- [ ] Existe acta de cierre de MVP 4 con fecha máxima 2026-11-27.

## 6. Ambientes y migraciones

### Local

- [ ] Versiones de frontend, backend y esquema están identificadas.
- [ ] Las pruebas automatizadas aplicables pasan.
- [ ] Los flujos manuales críticos tienen evidencia.

### Copia recuperable de datos

- [ ] El origen, fecha y procedimiento de recuperación están registrados.
- [ ] Las migraciones requeridas son idempotentes.
- [ ] Se verificó integridad antes y después.

### Azure

- [ ] Frontend, API, Entra ID, CORS y Azure SQL corresponden a la misma versión.
- [ ] Variables y secretos están fuera del repositorio y de la evidencia pública.
- [ ] Migraciones aplicadas coinciden con el inventario.
- [ ] Smoke, E2E y rollback tienen resultado fechado.
- [ ] Costos/cuotas y disponibilidad del ambiente están controlados durante la entrega.

## 7. Evidencia técnica y de QA

- [ ] Cada prueba registra fecha, ambiente, versión, procedimiento y resultado.
- [ ] Se conservan resultados de backend, frontend y base de datos.
- [ ] La matriz E2E cubre usuario, administrador, grupo y talleres.
- [ ] La matriz de regresión cubre caminos felices y errores críticos.
- [ ] La evidencia visual no contiene RUT, tokens, códigos u otros datos indebidos.
- [ ] Los riesgos residuales tienen decisión de aceptación.

## 8. Documentación y diagramas

- [ ] Índice, resumen, alcance, requisitos, arquitectura, datos y backlog están sincronizados.
- [ ] [`12-auditoria-alcance-implementacion-2026-08-11.md`](12-auditoria-alcance-implementacion-2026-08-11.md) está integrado y sus hallazgos están resueltos o aceptados.
- [ ] [`13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md`](13-diagramas-arquitectura-flujos-y-secuencias-mvp1-mvp2.md) coincide con los contratos y despliegue finales.
- [ ] La memoria de tesis cita evidencia vigente y distingue resultados locales de online.
- [ ] Los históricos se preservan y las correcciones se realizan mediante nuevas actas o anotaciones.
- [ ] Enlaces Markdown, UTF-8 y formato final están validados.
- [ ] Anexos, glosario y referencias están completos.

## 9. Defensa y entrega

- [ ] El guion de demostración usa datos seguros y un ambiente disponible.
- [ ] Existe contingencia si Azure no está disponible.
- [ ] Se ensayaron tiempos, preguntas críticas y recuperación ante fallos.
- [ ] El candidato final pasó regresión y smoke entre el 2026-12-07 y el 2026-12-09.
- [ ] El paquete final fue revisado contra este checklist.
- [ ] El prototipo y la documentación fueron entregados el 2026-12-10 o antes.
- [ ] La recepción de la entrega quedó registrada.

## 10. Dictamen final

- [ ] **APROBABLE:** todos los criterios obligatorios están completos y los riesgos residuales están aceptados.
- [ ] **APROBABLE CON OBSERVACIONES:** no existen fallas críticas y las observaciones no comprometen objetivos, seguridad ni defensa.
- [ ] **NO APROBABLE:** existe al menos una brecha crítica de alcance, integridad, seguridad, ambiente o evidencia.

**Responsable del dictamen por rol:** ____________________

**Fecha:** ____________________

**Referencia de evidencia/acta:** ____________________
