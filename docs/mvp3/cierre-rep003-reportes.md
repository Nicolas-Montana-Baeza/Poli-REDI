# MVP3 — Cierre REP-003: Reportes e indicadores

Fecha: 2026-10-24
Estado: VALIDADO — Prototipo MVP3

## Objetivo

Corregir los indicadores administrativos de reservas,
duración y utilización de recursos.

## Cambios implementados

1. Habilitación de Reportes en el alcance MVP3.
2. Corrección de la navegación a /reports.
3. Cálculo de reservas vigentes considerando:
   - Estados PENDING y CONFIRMED.
   - Reservas cuya hora de finalización aún no transcurre.
   - Exclusión de reservas canceladas, rechazadas y expiradas.
4. Representación exacta de duraciones en horas y minutos.
5. Recuento de recursos distintos con reservas vigentes.
6. Distribución de reservas según su estado.
7. Agrupación de reservas por hora de inicio en America/Santiago.
8. Exclusión de fechas y duraciones inválidas.

## Pruebas

- 5 pruebas específicas REP-003: aprobadas.
- 43 pruebas de utilidades frontend: aprobadas.
- 73 pruebas de componentes Vue: aprobadas.
- Compilación de producción Vite: aprobada.
- Validación git diff --check: sin errores.

## Alcance y limitaciones

Los indicadores representan las reservas obtenidas de la API.

Las horas reservadas corresponden a la duración total
de reservas vigentes; no representan necesariamente
horas de utilización efectiva.

Los horarios más frecuentes reflejan horas de inicio,
no ocupación por franjas completas.

El módulo no calcula todavía porcentajes de ocupación
respecto de la capacidad disponible.

Los indicadores temporales se actualizan al recargar
los datos de la vista; no existe actualización
automática en tiempo real.

No se realizaron modificaciones en PostgreSQL
ni se añadieron endpoints de reportes.

## Resultado

REP-003 validado técnicamente para el prototipo MVP3.
