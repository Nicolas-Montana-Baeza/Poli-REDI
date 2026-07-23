# Poli-REDI - Requisitos, Historias de Usuario, Casos de Uso y Roadmap MVP

> **Fecha de consolidación:** 2026-07-23  
> **Propósito:** Agrupar el catálogo funcional, historias de usuario, casos de uso formales y la hoja de ruta de incrementos de software (MVP 1 a MVP 4).

---

## 1. Roadmap por Incrementos (MVPs)

```mermaid
gantt
    title Roadmap Poli-REDI
    dateFormat  YYYY-MM-DD
    section MVP 1 Base Operativa
    Consulta Disponibilidad & Reservas Básicas :done, mvp1, 2026-01-01, 2026-03-31
    section MVP 2 Reglas & Grupal
    Quorum 10 Persona & Frecuencia Semanal   :done, mvp2, 2026-04-01, 2026-06-30
    section MVP 3 Administración
    Bloqueos, Infracciones & Notificaciones   :active, mvp3, 2026-07-01, 2026-08-31
    section MVP 4 Trazabilidad
    Reportes Avanzados & BI                  :future, mvp4, 2026-09-01, 2026-10-31
```

* **MVP 1 (Base Operativa):** Demo funcional. Inicios de sesión, disponibilidad en agenda, reservas simples, captura de RUT y panel admin inicial.
* **MVP 2 (Reglas y Flujo Grupal):** Accepted locally. Frecuencia semanal, mínimo 10 participantes, código grupal cifrado, prioridad institucional y ventana de tiempo.
* **MVP 3 (Administración Extendida):** Parcial. Bloqueo con motivos, notificaciones internas y pantalla informativa pública.
* **MVP 4 (Trazabilidad y Futuro):** En desarrollo/Trabajo futuro. Reportes avanzados e integración con sistemas académicos institucionales.

---

## 2. Historias de Usuario (HU)

* **HU-01 (Estudiante - Consultar Disponibilidad):** *Como estudiante, quiero consultar la disponibilidad de las canchas por fecha y hora para saber cuándo puedo reservar.*
* **HU-02 (Estudiante - Crear Reserva Particular):** *Como estudiante, quiero reservar una cancha ingresando mis datos para asegurar un bloque horario de juego.*
* **HU-03 (Estudiante - Flujo Grupal):** *Como estudiante organizador, quiero invitar a 9 compañeros mediante un código de unión para alcanzar el quorum mínimo requerido.*
* **HU-04 (Administrador - Cancelación Excepcional):** *Como administrador, quiero cancelar una reserva particular ante un evento institucional prioritario para reasignar la cancha.*
* **HU-05 (Estudiante - Inscripción a Talleres):** *Como estudiante, quiero inscribirme a talleres deportivos semanales según los cupos disponibles.*

---

## 3. Casos de Uso Formales (CU)

### CU-01: Solicitar Reserva Particular
* **Actor:** Estudiante Autenticado.
* **Precondición:** El usuario debe tener su RUT registrado y no tener bloqueos activos ni reservas en los últimos 7 días.
* **Flujo Principal:**
  1. El estudiante ingresa a la vista `/disponibilidad`.
  2. Selecciona la cancha, fecha y horario deseado.
  3. El sistema valida disponibilidad y restricción semanal en servidor.
  4. Si la cancha requiere grupo, el sistema crea la reserva en estado `PENDING` y genera un código de unión AES.
  5. El usuario comparte el código hasta completar los 10 participantes.
  6. Al alcanzar el quorum, la reserva pasa automáticamente a `CONFIRMED`.

### CU-02: Gestionar Conflicto Institucional (Administrador)
* **Actor:** Administrador.
* **Precondición:** Existen reservas particulares `CONFIRMED` en un horario donde se programará una clase/EFI.
* **Flujo Principal:**
  1. El administrador ingresa la actividad institucional en la agenda.
  2. El sistema detecta el solapamiento de horarios.
  3. El servidor cancela automáticamente la reserva particular afectada.
  4. El sistema emite una notificación interna explicativa al alumno desasociado.
