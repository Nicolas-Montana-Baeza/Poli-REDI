# Poli-REDI - Resumen Ejecutivo y Estado Actual del Producto

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

| Módulo | Estado | Descripción y Evidencia |
| :--- | :---: | :--- |
| **Autenticación y Sesión** | `IMPLEMENTADO` | Integración con Microsoft Entra ID (JWT) + Modo local Dev Auth (`X-Dev-Auth-*`). |
| **Identidad y Perfil (RUT)** | `IMPLEMENTADO` | Modal obligatorio de captura de RUT para usuarios normales antes de reservar. |
| **Recursos y Disponibilidad** | `IMPLEMENTADO` | Consulta de disponibilidad por fecha/recurso entre 08:00 y 22:00 (`GET /api/resources`). |
| **Creación de Reservas** | `IMPLEMENTADO` | Reservas particulares e institucionales controladas por servidor. |
| **Flujo Grupal y Código de Unión**| `ACCEPTED LOCALLY` | Progreso, código cifrado recuperable solo por propietario, rotación, `/join`, confirmación, retiro, reconfirmación, deadline inclusivo y expiración. Migración 004 y concurrencia real en Azure SQL pendientes. |
| **Diseño y UI/UX (Tokens & Components)**| `IMPLEMENTADO LOCAL` | `StatusBadge` se usa en tarjetas, detalle y próximas reservas; `ConfirmModal` reemplaza confirmaciones nativas de cancelación y protege la rotación de códigos; `MetricCard` cubre las tres métricas administrativas; `PrimaryButton` se usa en los modales y acciones del progreso grupal. Calendarios y bloques con semántica especial conservan sus controles propios. La validación integrada/online continúa pendiente. |
| **Control de Frecuencia Semanal** | `IMPLEMENTADO` | Restricción de 7 días corridos entre reservas solicitadas por el mismo usuario. |
| **Prioridad Institucional** | `PENDIENTE` | La regla está aprobada, pero el flujo administrativo de resolución y cancelación automática aún no está implementado. |
| **Cancelación de Reservas** | `IMPLEMENTADO` | Propietario o administrador pueden cancelar reservas activas (`PATCH /api/reservations/cancel`). |
| **Talleres e Inscripciones** | `IMPLEMENTADO` | Talleres recurrentes con control de cupos e inscripción con RUT (`/api/activities`). |
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
