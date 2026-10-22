# ADMIN-005 — Cierre de gestión de sesiones institucionales

## Alcance implementado

Se amplió la gestión de actividades institucionales para
administrar ocurrencias individuales de programaciones semanales.

Funcionalidades:
- Suspender una sesión sin cancelar la actividad completa.
- Reprogramar una sesión a otra fecha y horario.
- Modificar nuevamente una sesión previamente reprogramada.
- Reactivar o reprogramar sesiones suspendidas.
- Conservar el historial de revisiones y sus motivos.
- Mantener intactas las demás ocurrencias semanales.

## Validaciones

- Autorización de administradores y gestores institucionales.
- Validación de fechas, horarios y jornada vigente.
- Control de bloqueos administrativos.
- Prevención de modificaciones inconsistentes.
- Integración con Disponibilidad y conflictos de programación.
- Restricciones aplicables a actividades con inscripciones.

## Persistencia

Migración:
PG16_0013_mvp3_occurrence_revisions.sql

Se conserva el registro de excepciones institucionales
y se incorpora trazabilidad para modificaciones sucesivas.

## Pruebas

- Pruebas automatizadas del backend: aprobadas.
- Pruebas automatizadas del frontend: aprobadas.
- Compilación del frontend: aprobada.
- Despliegue y pruebas funcionales: confirmados por el responsable.

## Resultado

Ampliación de ADMIN-005 validada para el alcance del prototipo.

Las mejoras futuras sobre notificaciones e impacto individual
en inscripciones deberán tratarse como alcance independiente.
