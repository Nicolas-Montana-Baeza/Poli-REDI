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
* `PENDIENTE`: Funcionalidad planeada o en trabajo futuro.

---

## 3. Estado Funcional por Módulos

| Módulo | Estado | Descripción y Evidencia |
| :--- | :---: | :--- |
| **Autenticación y Sesión** | `IMPLEMENTADO` | Integración con Microsoft Entra ID (JWT) + Modo local Dev Auth (`X-Dev-Auth-*`). |
| **Identidad y Perfil (RUT)** | `IMPLEMENTADO` | Modal obligatorio de captura de RUT para usuarios normales antes de reservar. |
| **Recursos y Disponibilidad** | `IMPLEMENTADO` | Consulta de disponibilidad por fecha/recurso entre 08:00 y 22:00 (`GET /api/resources`). |
| **Creación de Reservas** | `IMPLEMENTADO` | Reservas particulares e institucionales controladas por servidor. |
| **Flujo Grupal y Código de Unión**| `IMPLEMENTADO` | Cifrado AES de códigos de unión, rotación de llaves y quorum obligatorio de 10 personas. |
| **Diseño y UI/UX (Tokens & Components)**| `VERIFICADO` | Tokens de estado HSL/Hex en `variables.css` y componentes canónicos (`StatusBadge.vue`, `PrimaryButton.vue`, `ConfirmModal.vue`, `MetricCard.vue`) integrados sin errores de compilación (`npm run build`). |
| **Control de Frecuencia Semanal** | `IMPLEMENTADO` | Restricción de 7 días corridos entre reservas solicitadas por el mismo usuario. |
| **Prioridad Institucional** | `IMPLEMENTADO` | Bloqueo o cancelación de reservas particulares ante clases o actividades académicas. |
| **Cancelación de Reservas** | `IMPLEMENTADO` | Propietario o administrador pueden cancelar reservas activas (`PATCH /api/reservations/cancel`). |
| **Talleres e Inscripciones** | `IMPLEMENTADO` | Talleres recurrentes con control de cupos e inscripción con RUT (`/api/activities`). |
| **Notificaciones Internas** | `PARCIAL` | Consulta y contador interno (`GET /api/notifications`); notificaciones push multicanal pendientes. |
| **Panel Administrador** | `IMPLEMENTADO` | Gestión básica de usuarios, reservas, recursos e indicadores iniciales. |
| **Auditoría y Trazabilidad** | `IMPLEMENTADO` | Esquema Azure SQL registra marcas temporales y estados de cambios en reservas. |

---

## 4. Brechas y Decisiones Aprobadas

1. **Duración de Reservas:** Se permiten selecciones de 30, 60, 90, 120, 150 y 180 minutos.
2. **Ventana de Reserva:** El usuario puede reservar desde el día actual hasta el día anterior al mismo día de la semana siguiente.
3. **Mínimo de Participantes:** Mínimo 10 participantes obligatorios para Canchas 1, 2 y 3.
4. **Cancelación Automática por Prioridad:** Un conflicto entre actividad institucional y reserva particular cancela automáticamente la reserva particular y notifica al afectado.
