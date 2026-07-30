# Poli-REDI - Resumen Ejecutivo y Estado Actual del Producto

> **Revision vigente:** 2026-07-30. El resultado local es `APROBABLE` para
> continuar el cierre condicionado de MVP 2. La evidencia, riesgos y limites
> constan en el
> [acta integral](./historico_y_checklists/15-acta-revision-integral-2026-07-30.md).

> **Fecha de consolidación:** 2026-07-23  
> **Propósito:** Documento único para comprender la visión ejecutiva, el estado funcional de cada módulo, las reglas vigentes y la matriz de brechas del prototipo.

---

## 1. Problema y Objetivo

El proceso de reserva del Polideportivo institucional (3 canchas y 1 sala multiusos) dependía históricamente de una gestión manual en Google Calendar, concentrando decisiones en el encargado sin visibilidad pública de disponibilidad ni trazabilidad.

**Objetivo:** Desarrollar un prototipo funcional (aplicación web desacoplada) que centralice la consulta de disponibilidad, reservas particulares e institucionales, aplicando validaciones de horario, frecuencia semanal y quorum mínimo en servidor.

---

## 2. Convenciones de Estado

* `APROBADO`: Definido explícitamente en el alcance o acordado con contraparte.
* `IMPLEMENTADO`: Comportamiento observable en código Go/Vue o base de datos.
* `VERIFICADO`: Comprobado mediante pruebas unitarias/builds con resultado satisfactorio.
* `ACCEPTED LOCALLY`: Incremento aceptado mediante pruebas locales, todavía pendiente de validación integrada en Azure SQL.
* `IMPLEMENTADO LOCAL`: Funcionalidad integrada en el código y comprobada en el ambiente local mediante pruebas o build; todavía pendiente de validación en el ambiente desplegado.
* `PENDIENTE`: Funcionalidad planeada o en trabajo futuro.

---

## 3. Estado Funcional por Módulos

### Delta verificado el 2026-07-30

* La revision de roles y contratos no mantiene hallazgos P0/P1 abiertos dentro
  del alcance revisado. Esto no equivale a una prueba de penetracion.
* La politica de reservas falla cerrada si no existe una politica publicada
  valida; la interfaz filtra recursos segun estado, modalidad y rol.
* Disponibilidad, Mis Reservas, Historial y confirmacion por codigo reutilizan el
  mismo detalle con capacidades explicitas. El codigo se obtiene bajo demanda y
  solo para el propietario autorizado.
* Las tarjetas personales son seleccionables completas. Los dialogos administran
  foco, Escape y restauracion; la linea de tiempo y el progreso tienen semantica
  de teclado/ARIA.
* El dashboard evita duplicar la proxima reserva.
* Talleres e inscripciones propias son MVP 2; clases, campeonatos y otros eventos
  institucionales son MVP 3 y requieren relacion explicita antes de atribuir
  participacion personal.
* `007` y `008` son prospectivas y no reescriben reservas historicas.

**Validacion local exacta:** 18 pruebas Node, 77 pruebas Vitest y build frontend
de produccion aprobados. Siguen pendientes Azure SQL `007`/`008` y el flujo
online completo con Entra ID/CORS/API.

| Módulo | Estado | Descripción y Evidencia |
| :--- | :---: | :--- |
| **Autenticación y Sesión** | `IMPLEMENTADO` | Integración con Microsoft Entra ID (JWT) + Modo local Dev Auth (`X-Dev-Auth-*`). |
| **Identidad y Perfil (RUT)** | `ACCEPTED LOCALLY` | Modal condicionado a `/api/me` listo, usuario no administrador y RUT ausente/inválido. Registro write-once, lectura posterior e idempotencia del mismo valor. |
| **Recursos y Disponibilidad** | `IMPLEMENTADO` | Consulta de disponibilidad por fecha/recurso entre 08:00 y 22:00 (`GET /api/resources`). |
| **Creación de Reservas** | `IMPLEMENTADO` | Reservas particulares e institucionales controladas por servidor. |
| **Flujo Grupal y Código de Unión**| `ACCEPTED LOCALLY` | Progreso, código cifrado recuperable solo por propietario, rotación, `/join`, confirmación, retiro, reconfirmación, deadline inclusivo y expiración. Migración 004 y concurrencia real en Azure SQL pendientes. |
| **Diseño y UI/UX (Tokens & Components)**| `IMPLEMENTADO LOCAL` | `StatusBadge` se usa en tarjetas, detalle y próximas reservas; `ConfirmModal` reemplaza confirmaciones nativas de cancelación y protege la rotación de códigos; `MetricCard` cubre las tres métricas administrativas; `PrimaryButton` se usa en los modales y acciones del progreso grupal. Calendarios y bloques con semántica especial conservan sus controles propios. La validación integrada/online continúa pendiente. |
| **Control de Frecuencia Semanal** | `IMPLEMENTADO` | Restricción de 7 días corridos entre reservas solicitadas por el mismo usuario. |
| **Prioridad Institucional** | `PENDIENTE` | La regla está aprobada, pero el flujo administrativo de resolución y cancelación automática aún no está implementado. |
| **Cancelación de Reservas** | `IMPLEMENTADO` | Propietario o administrador pueden cancelar reservas activas (`PATCH /api/reservations/cancel`). |
| **Talleres e Inscripciones** | `ACCEPTED LOCALLY` | Cupos, ocurrencias normalizadas y prevención de solapes taller↔taller para inscripciones activas. Contrato `POST /api/workshops/:id/enroll`. |
| **Historial Personal** | `PARCIAL` | El historial básico de reservas propias o participadas pertenece a MVP 1 y está implementado. La consulta de talleres e inscripciones se incorpora como ampliación controlada de MVP 2. |
| **Historial Institucional** | `PENDIENTE` | Las clases, actividades institucionales y otros eventos se incorporarán en MVP 3. No se atribuirán como participación personal mientras no exista una relación explícita usuario–actividad. |
| **Notificaciones Internas** | `PARCIAL` | Consulta y contador (`GET /api/notifications`) y notificación única de expiración verificada localmente; lectura, destinos, otros eventos y sistema completo pendientes. |
| **Panel Administrador** | `PARCIAL` | Lectura operacional, indicadores, imágenes de recursos y políticas; gestión completa de usuarios, recursos, bloqueos y programación pendiente. |
| **Auditoría y Trazabilidad** | `PARCIAL` | El esquema registra cambios de reservas, participantes, objetivos y expiraciones; falta consulta administrativa integral. |

---

## 4. Brechas y Decisiones Aprobadas

1. **Duración de Reservas:** Se permiten selecciones de 30, 60, 90, 120, 150 y 180 minutos.
2. **Ventana de Reserva:** El usuario puede reservar desde el día actual hasta el día anterior al mismo día de la semana siguiente.
3. **Mínimo de Participantes:** Mínimo 10 participantes obligatorios para Canchas 1, 2 y 3.
4. **Cancelación Automática por Prioridad:** Un conflicto entre actividad institucional y reserva particular cancela automáticamente la reserva particular y notifica al afectado.

> **Límite de verificación:** El MVP 2 está `ACCEPTED LOCALLY`. Continúan pendientes la ejecución de la migración 004, su idempotencia y las pruebas de concurrencia real en Azure SQL.

Las extensiones de integridad de RUT y horarios de talleres también están `ACCEPTED LOCALLY`. Migraciones 005/006, DDL, idempotencia y carreras reales en Azure SQL siguen pendientes.

### Alcance aprobado del historial

1. **MVP 1:** historial básico de reservas propias o participadas, incluyendo
   reservas pasadas y canceladas.
2. **MVP 2 (ampliación controlada):** consulta de talleres e inscripciones del
   usuario, sin presentar una inscripción como asistencia comprobada.
3. **MVP 3:** historial institucional de clases, actividades programadas y otros
   eventos, con acceso acorde al rol.
4. Una clase o evento solo podrá aparecer en el historial personal cuando el
   modelo registre una relación explícita de participación del usuario. La mera
   existencia del evento en la agenda no permite inferir asistencia.

### Matriz RUT por rol y acción

| Actor sin RUT | Reserva normal | Reserva grupal propia | Inscripción a taller | Confirmar como participante |
| :--- | :---: | :---: | :---: | :---: |
| Usuario normal | No | No | No | No |
| Administrador | Sí | Sí | Sí | No |

Una vez registrado, el RUT no puede cambiarse ni borrarse. Repetir el mismo valor es idempotente; un valor diferente o duplicado devuelve `409`. El inicio de sesión local conserva el RUT existente.
