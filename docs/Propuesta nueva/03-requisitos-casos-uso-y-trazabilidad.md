# Requisitos, casos de uso y trazabilidad de Poli-REDI

**Estado:** CANÓNICO ADOPTADO

**Corte de revisión:** 2026-08-11

**Fecha límite del prototipo:** 2026-12-10
**Criterio:** distinguir alcance aprobado, implementación observable y verificación en el ambiente objetivo.

## 1. Actores

| Actor | Responsabilidad |
|---|---|
| Usuario normal | Consultar disponibilidad, completar RUT solo cuando corresponde, reservar, gestionar participaciones grupales y talleres, cancelar lo permitido y consultar su historial. |
| Administrador | Operar inventario, usuarios, bloqueos, programación, prioridades e historial institucional con permisos y auditoría. Puede agendar sin RUT cuando la regla aprobada así lo permite. |
| Microsoft Entra ID | Autenticar y emitir tokens para la API desplegada. |
| Backend | Resolver identidad, permisos, reglas, conflictos y errores de dominio. |
| Azure SQL | Persistir y defender integridad, concurrencia, historia y auditoría. |
| Pantalla informativa | Presentar disponibilidad institucional sanitizada como parte de MVP 3, sin datos personales ni acciones privadas. |

## 2. Requisitos funcionales y asignación por MVP

| Grupo | Requisitos principales | MVP | Dictamen al 2026-08-11 |
|---|---|---:|---|
| Base técnica e identidad | Autenticación, sincronización, roles, bloqueo, RUT condicionado y persistencia | 1 | Demostrable en local; integración online no cerrada. |
| Reserva básica | Recursos, fecha, hora, duración, creación, consulta, cancelación y conflicto básico | 1 | Demostrable en local; requiere evidencia integrada online. |
| Disponibilidad | Recursos, reservas, talleres y bloques por rango de fechas | 1–2 | El backend por rango existe; la integración completa en frontend está pendiente. |
| Flujo grupal | Mínimo, objetivo, participantes, código, plazo, expiración y privacidad | 2 | Parcial; depende de integración frontend, migraciones `007`/`008` y validación Azure. |
| Talleres | Listado, inscripción, desinscripción, cupo, solape taller↔taller, historial propio y reinscripción mientras el taller esté activo | 2 | Parcialmente verificado en local; no compara taller↔reserva personal entre recursos y falta cierre online. |
| Administración | Usuarios, inventario, bloqueos, programación, prioridades, notificación específica de prioridad, pantalla pública sanitizada, conflictos e historial institucional | 3 | Parcial. |
| Calidad y soporte | Sistema core de notificaciones, reportes básicos, auditoría consultable, seguridad, accesibilidad y despliegue | 4 | Pendiente y acotado al cierre del prototipo. |

El catálogo detallado RF-001 a RF-025 y RNF-001 a RNF-012 se conserva como referencia en [`referencia/08-requisitos-completos.md`](referencia/08-requisitos-completos.md). Si difiere de esta asignación o del alcance definitivo, prevalece el canon vigente.

## 3. Requisitos no funcionales críticos

1. **Seguridad y privacidad:** autorización en backend y minimización de datos por audiencia.
2. **Secretos:** códigos, claves, tokens y contraseñas no deben exponerse en respuestas, trazas ni repositorio.
3. **Trazabilidad:** decisiones, cambios de estado y operaciones críticas deben poder auditarse.
4. **Consistencia temporal:** reglas de negocio en `America/Santiago` y timestamps técnicos normalizados.
5. **Integridad concurrente:** validaciones coherentes en servicio y base de datos.
6. **Accesibilidad:** teclado, foco visible, texto además de color, contraste y movimiento reducido.
7. **Responsive:** funcionamiento verificable en anchos representativos.
8. **Errores seguros:** mensajes útiles de dominio sin filtrar detalles internos.
9. **Despliegue reproducible:** migraciones, configuración, reversión y evidencia del ambiente objetivo.
10. **Documentación viva:** aprobado, implementado y verificado deben registrarse como estados distintos.

## 4. Historias de usuario clave

### HU-01 — Consultar disponibilidad

Como usuario autenticado, quiero ver recursos y bloques ocupados por rango para elegir un horario sin conocer información privada de terceros.

### HU-02 — Crear una reserva

Como usuario habilitado, quiero solicitar un recurso para una fecha, hora, duración y cantidad de participantes válidas, recibiendo el estado definido por la política vigente.

### HU-03 — Reunir participantes

Como propietario de una solicitud grupal, quiero compartir un código, consultar el progreso y editar el objetivo dentro del plazo permitido para alcanzar el mínimo de confirmación.

### HU-04 — Confirmar o retirar participación

Como usuario autenticado, quiero confirmar o retirar mi participación cuando esté permitido, respetando los solapes definidos para reservas y participaciones grupales.

### HU-05 — Gestionar talleres

Como usuario, quiero inscribirme, desinscribirme y reinscribirme mientras el taller esté activo, según cupo y ausencia de otro taller solapado, conservando el historial de episodios.

### HU-06 — Administrar la operación

Como administrador, quiero gestionar usuarios, recursos, bloqueos y programación institucional con prioridad definida, notificación específica, pantalla pública sanitizada, permisos y auditoría.

## 5. Casos de uso principales

### CU-01 — Crear reserva

**Precondiciones:** sesión válida; usuario no bloqueado; RUT válido cuando aplica; recurso, capacidad y política disponibles.

**Flujo:** seleccionar horario → indicar actividad y participantes → validar identidad y reglas → comprobar solapes y política → persistir → devolver detalle seguro.

**Alternativas:** horario o cantidad inválidos, frecuencia, conflicto, política ausente, recurso no elegible o datos incompletos.

### CU-02 — Confirmar participación grupal

**Precondiciones:** código válido y disponible; usuario autenticado; plazo vigente.

**Flujo:** consultar invitación sanitizada → confirmar → validar solapes y cupo → actualizar conteo → confirmar la reserva al alcanzar el mínimo.

**Alternativas:** código inválido, cancelado o vencido; propietario intenta retirarse; objetivo alcanzado; solape personal.

### CU-03 — Gestionar inscripción a taller

**Alta:** exige RUT únicamente al usuario normal que aún no lo tenga, taller activo, cupo y ausencia de otro taller activo solapado con inscripción confirmada. No compara el taller con una reserva personal ubicada en otro recurso.

**Baja:** se permite mientras el taller esté activo, sin cutoff adicional; afecta solo la inscripción activa propia, es idempotente, libera cupo y conserva historia.

**Reinscripción:** crea un episodio nuevo y conserva el anterior como cancelado.

### CU-04 — Cancelar reserva

**Actores:** propietario o administrador autorizado.

**Reglas:** la reserva debe estar activa y no haber finalizado; la acción conserva trazabilidad. Un cutoff configurable es una mejora futura y no constituye criterio de cierre de MVP 2.

### CU-05 — Configurar política

**Actor:** administrador.

**Regla:** los cambios son prospectivos; las solicitudes existentes conservan la versión aplicable. La interfaz administrativa completa corresponde a MVP 3.

## 6. Roadmap y criterio de cierre

| MVP | Alcance final | Estado real al corte | Fecha objetivo de cierre |
|---|---|---|---:|
| MVP 1 | Base técnica, identidad, persistencia, disponibilidad y reserva básica | Demostrable local; no online | 2026-08-28 |
| MVP 2 | Flujo de usuario, reserva grupal y talleres | Parcial; no cerrado | 2026-09-25 |
| MVP 3 | Administración, inventario, bloqueos, programación, prioridad, notificación específica de prioridad, pantalla pública e historial institucional | Parcial | 2026-10-30 |
| MVP 4 | Calidad, soporte, sistema core de notificaciones, reportes básicos, auditoría y despliegue | Pendiente y acotado | 2026-11-27 |
| Entrega | Integración documental, evidencia y defensa | Pendiente | 2026-12-10 |

Los criterios de salida completos se encuentran en [`11-cronograma-cierre-2026.md`](11-cronograma-cierre-2026.md) y [`14-checklist-cierre-total-2026-12-10.md`](14-checklist-cierre-total-2026-12-10.md).

## 7. Matriz de trazabilidad resumida

| Necesidad | Requisito | Implementación observable | Evidencia pendiente para cierre |
|---|---|---|---|
| Identidad segura | Identidad / seguridad | Middleware y rutas protegidas | Flujo online con Entra ID, CORS y API desplegada. |
| Disponibilidad por rango | Consulta unificada | Backend por rango existente | Integración frontend, regresión y evidencia online. |
| Reserva válida | Creación / conflictos | Servicio, repositorio y defensas DB | Migraciones y E2E en Azure. |
| Quórum grupal | Participantes / política | Endpoints, estados y progreso local | `007`/`008`, frontend y Azure. |
| Privacidad del código | Consulta sanitizada | Contrato por audiencia | Casos válido, inexistente, cancelado y vencido. |
| Talleres | Alta, baja, solape taller↔taller e historial propio | Flujo local observable | E2E integrado y online sin imponer taller↔reserva personal entre recursos. |
| Administración | Inventario, usuarios, programación, prioridad y pantalla pública | Componentes parciales | Contratos completos, permisos, sanitización, auditoría y E2E. |
| Calidad operativa | Notificaciones, reportes, seguridad y despliegue | Parcial o pendiente | QA visual/accesible, seguridad, despliegue y reversión. |

## 8. Decisiones y reservas

- La migración `009` es una **propuesta no aprobada**. No forma parte del cierre ni debe ejecutarse hasta registrar aprobación, preflight, reversión y efecto sobre datos existentes.
- `007` y `008` siguen siendo dependencias de cierre y deben verificarse primero en copia recuperable y luego en Azure.
- Una capacidad demostrable en local no equivale a verificación online.
- No se amplía el prototipo a BI avanzado, IA, integración académica, multisede, sincronización bidireccional con Google, gestión avanzada de campeonatos ni detección automatizada de abuso.

## 9. Regla de mantenimiento

Cuando cambie una regla, actualizar requisito, caso de uso, contrato técnico, prueba, backlog y evidencia en el mismo incremento. No cambiar un estado a “verificado” sin registrar ambiente, fecha, procedimiento y resultado.
