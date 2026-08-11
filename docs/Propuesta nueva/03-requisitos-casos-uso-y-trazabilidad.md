# Requisitos, casos de uso y trazabilidad de Poli-REDI

**Estado:** CANÓNICO  
**Criterio:** separar necesidad aprobada, implementación observable y verificación.

## 1. Actores

| Actor | Responsabilidad |
|---|---|
| Usuario normal | Consultar disponibilidad, completar RUT, reservar, confirmar participación, gestionar talleres, cancelar lo permitido y revisar su información. |
| Administrador | Consultar operación y ejecutar acciones administrativas autorizadas. |
| Microsoft Entra ID | Autenticar y emitir tokens para la API. |
| Backend | Resolver identidad, permisos, reglas y errores de dominio. |
| Azure SQL | Persistir y defender integridad, conflictos y auditoría. |
| Pantalla informativa | Presentar información pública sanitizada; alcance completo pendiente de MVP 3. |

## 2. Requisitos funcionales agrupados

| Grupo | Requisitos principales | Estado resumido |
|---|---|---|
| Identidad | Autenticación, sincronización, rol, bloqueo y RUT | Implementado; integración online pendiente. |
| Disponibilidad | Recursos, reservas, talleres y actividades por fecha | Parcial; bloqueos y filtros backend pendientes. |
| Reservas | Crear, cancelar, consultar propias, historial y conflictos | Implementado localmente con brechas operativas. |
| Flujo grupal | Frecuencia, mínimo, participantes, objetivo, código, deadline y expiración | Aprobable condicionado. |
| Talleres | Listado, inscripción, desinscripción, cupo, historial y reinscripción | Verificado localmente. |
| Administración | Usuarios, recursos, políticas, bloqueos, programación, conflictos e infracciones | Parcial. |
| Notificaciones | Consultar, marcar y generar eventos relevantes | Parcial. |
| Reportes y auditoría | Indicadores, vistas, consultas y trazabilidad | Parcial. |

El catálogo detallado original RF-001 a RF-025 y RNF-001 a RNF-012 se conserva en [`referencia/08-requisitos-completos.md`](referencia/08-requisitos-completos.md).

## 3. Requisitos no funcionales críticos

1. **Seguridad:** autorización en backend y minimización de datos por audiencia.
2. **No exposición de secretos:** código, claves, tokens y contraseñas fuera de respuestas y repositorio.
3. **Trazabilidad:** decisiones, cambios de estado y operaciones críticas auditables.
4. **Consistencia temporal:** reglas aplicadas en `America/Santiago` con timestamps técnicos UTC.
5. **Integridad concurrente:** validaciones en servicio y base de datos.
6. **Accesibilidad:** teclado, foco, texto además de color y movimiento reducido.
7. **Responsive:** funcionamiento verificable en anchos representativos.
8. **Errores seguros:** mensajes de dominio sin detalles internos.
9. **Documentación viva:** diferencias entre aprobado, implementado y verificado deben ser explícitas.

## 4. Historias de usuario clave

### HU-01 — Consultar disponibilidad

Como usuario autenticado, quiero ver los recursos y sus bloques ocupados para seleccionar un horario disponible sin conocer información privada de terceros.

### HU-02 — Crear una reserva

Como usuario con RUT cuando corresponda, quiero solicitar un recurso para una fecha, hora y duración válidas, recibiendo un estado definido por la política vigente.

### HU-03 — Reunir participantes

Como propietario de una solicitud grupal, quiero compartir un código y ver el progreso para alcanzar el mínimo antes del plazo.

### HU-04 — Confirmar participación

Como usuario autenticado, quiero confirmar o retirar mi participación sin solapar mi agenda personal.

### HU-05 — Gestionar talleres

Como usuario, quiero inscribirme, desinscribirme y reinscribirme según cupo y estado del taller, conservando historial.

### HU-06 — Administrar operación

Como administrador, quiero consultar y modificar únicamente los elementos autorizados, con trazabilidad y sin reescribir historia.

## 5. Casos de uso principales

### CU-01 — Crear reserva

**Precondiciones:** sesión válida; usuario no bloqueado; RUT válido cuando aplica; recurso y política disponibles.  
**Flujo:** seleccionar horario → validar en frontend → validar identidad y reglas en backend → verificar conflictos y política en DB → persistir → devolver detalle seguro.  
**Alternativas:** horario inválido, frecuencia, conflicto, política ausente, recurso no elegible o datos incompletos.

### CU-02 — Confirmar participación grupal

**Precondiciones:** código válido; usuario autenticado; plazo vigente.  
**Flujo:** consultar progreso → confirmar → validar solapes y cupo → actualizar conteo → cambiar estado si alcanza mínimo.  
**Alternativas:** código inválido, owner intenta retirarse, deadline vencido, objetivo alcanzado o solape personal.

### CU-03 — Gestionar inscripción a taller

**Alta:** exige RUT para usuario normal, taller activo, cupo y ausencia de conflicto.  
**Baja:** no exige RUT, afecta solo el episodio `CONFIRMED` propio, es idempotente y libera cupo.  
**Reinscripción:** crea un episodio nuevo y conserva el anterior `CANCELLED`.

### CU-04 — Cancelar reserva

**Actores:** propietario o administrador autorizado.  
**Reglas:** la reserva debe estar activa, no finalizada y en un estado cancelable; la acción conserva trazabilidad.

### CU-05 — Configurar política

**Actor:** administrador.  
**Regla:** los cambios son prospectivos; solicitudes existentes conservan su versión. La interfaz administrativa completa sigue pendiente.

## 6. Roadmap y criterio de cierre

| MVP | Alcance | Estado |
|---|---|---|
| MVP 1 | Base técnica, autenticación, Azure SQL, reserva básica y demo | Demostrable. |
| MVP 2 | Flujo usuario, grupo y talleres | Aprobable condicionado. |
| MVP 3 | Administración institucional y eventos | Parcial. |
| MVP 4 | Calidad, reportes, notificaciones y soporte | En desarrollo. |

## 7. Matriz de trazabilidad resumida

| Necesidad | Requisito | Implementación | Evidencia |
|---|---|---|---|
| Identidad segura | RF identidad / RNF seguridad | Middleware y rutas protegidas | Pruebas backend; online pendiente. |
| Reserva válida | RF creación/conflictos | Servicio + repositorio + triggers | Pruebas locales; Azure pendiente. |
| Quórum grupal | RF participantes/política | Endpoints, estados y migraciones | Flujo local; `007`/`008` pendientes en Azure. |
| Privacidad | RNF minimización | Endpoint sanitizado y UI por audiencia | Pruebas y revisión local. |
| Talleres | RF talleres | Alta, baja, historial y auditoría | Evidencia 2026-08-04. |
| Administración | RF administración | Panel y lecturas base | Implementación parcial. |

## 8. Regla de mantenimiento

Cuando cambie una regla, actualizar requisito, caso de uso, contrato técnico, prueba y backlog en el mismo incremento. No cambiar un estado a “verificado” sin registrar ambiente, fecha, comando y resultado.
