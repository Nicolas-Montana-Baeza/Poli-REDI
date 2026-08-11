# ADR-004 — Reglas temporales y solapes

**Audiencia:** Analista, Arquitecto, Backend, Frontend y QA

**Propósito:** fijar semántica temporal de reservas, uso libre, grupos y talleres

**Estado:** aceptada el 2026-08-11

**Fuente:** decisiones funcionales confirmadas y auditoría de implementación

## Resumen

Los intervalos se solapan cuando `inicioA < finB` y `finA > inicioB`; extremos contiguos son válidos. La agenda personal evita ocupaciones simultáneas, pero MVP 2 trata talleres y reservas personales como dominios separados entre recursos. Cancelación y baja de taller se permiten mientras la actividad siga vigente, sin cutoff adicional.

## Reservas y `open_use`

1. Una reserva personal no puede solaparse con otra reserva o participación grupal confirmada del mismo usuario.
2. Dos bloques contiguos, por ejemplo 12:00–13:00 y 13:00–14:00, no se solapan.
3. Un recurso `open_use` puede reservarse libremente sin restricción semanal.
4. `open_use` no elimina la regla de agenda personal: se rechaza un intervalo solapado con otra reserva o participación del usuario.
5. La frecuencia semanal sigue aplicando a recursos reservables según su política.

## Reservas grupales

1. El solicitante cuenta como participante.
2. La solicitud permanece pendiente bajo el mínimo y se confirma al alcanzarlo.
3. El objetivo puede cambiar hasta el deadline derivado de la política vigente.
4. Confirmar o retirar participación respeta plazo, estado, cupo y agenda personal.
5. El propietario no puede retirar su propia participación sin cancelar la reserva.
6. La consulta pública del código minimiza información y no distingue indebidamente inexistente, cancelado o vencido.

## Talleres

1. El usuario puede inscribirse si el taller está activo, existe cupo y no tiene otro taller activo solapado.
2. MVP 2 solo compara taller↔taller. No bloquea taller↔reserva personal entre recursos.
3. El usuario puede desinscribirse mientras el taller esté activo, sin cutoff adicional.
4. La baja repetida es idempotente, libera cupo y conserva el episodio cancelado.
5. La reinscripción crea un episodio nuevo; no reactiva ni borra el histórico.
6. El historial de inscripción no acredita asistencia.

## Cancelación y cambios administrativos

- El propietario o administrador autorizado puede cancelar una reserva activa antes de que finalice.
- Un cutoff configurable para cancelaciones es una mejora futura, no criterio de cierre de MVP 2.
- Cambios de políticas, horarios, mínimo, objetivo o ventanas administrativas son prospectivos.
- Una reserva existente conserva snapshots de la política aplicable, salvo corrección excepcional explícita y auditada.
- La interfaz administrativa de estas políticas corresponde a MVP 3.

## Tiempo y concurrencia

Las reglas de negocio se interpretan en `America/Santiago`; los timestamps técnicos se normalizan. Backend y base de datos deben defender las mismas invariantes bajo concurrencia. Los casos de prueba están en [07-calidad-y-evidencia.md](../07-calidad-y-evidencia.md).

## Consecuencias

- No se debe introducir un bloqueo global de frecuencia para `open_use`.
- No se promete prevenir taller↔reserva personal en MVP 2.
- Una futura unificación de agendas o cutoff requiere nuevo ADR, migración compatible y pruebas de no retroactividad.
