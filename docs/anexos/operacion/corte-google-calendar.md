# Corte operativo desde Google Calendar

**Audiencia:** operación, Analista, DevOps, QA y responsables institucionales

**Propósito:** migrar eventos futuros y dejar Poli-REDI como fuente de escritura

**Estado:** plan operativo; fecha oficial de congelamiento pendiente

**Fuente:** inventario y análisis documental del sistema legado

## Resumen

El corte debe evitar doble escritura, pérdida de eventos, duplicados y exposición de títulos privados. Google Calendar quedará temporalmente en consulta; Poli-REDI será la única fuente de escritura tras una fecha y hora comunicadas.

La sincronización bidireccional no forma parte del prototipo. El corte depende de la programación institucional de MVP 3 y de una carga recuperable y reconciliada.

## Inventario recibido

Se identificaron cinco calendarios únicos; una segunda copia de Sala Multiuso era duplicada.

| Calendario legado | Destino | Eventos base | Cobertura observada |
|---|---|---:|---|
| Cancha 1 | Cancha 1 | 91 | 2026-03-18 a 2026-09-05 |
| Cancha 2 | Cancha 2 | 102 | 2026-03-18 a 2026-07-15 |
| Cancha 3 | Cancha 3 | 54 | 2026-03-18 a 2026-08-08 |
| Sala Multiuso | Sala Multiuso | 13 | 2026-04-06 a 2026-07-31 |
| Sala de Musculación | Gimnasio | 2 | 2026-08-01 a 2026-09-05 |

No se recibieron exportaciones para Muro Escalada, Sala Spinning ni Piscina. Debe confirmarse si nunca existieron o si faltan fuentes.

## Hallazgos preliminares

La expansión del 2026-07-14 al 2026-09-30 produjo 217 ocurrencias futuras y 14 solapamientos internos: 8 en Cancha 1 y 6 en Cancha 2.

Los eventos se modelan como actividades programadas, no como reservas personales. Títulos con información privada se muestran al usuario normal como actividad institucional sanitizada.

## Fases

### Inventario y limpieza

- confirmar calendarios, propietarios y recursos destino;
- resolver duplicados y archivos faltantes;
- limitar recurrencias sin término;
- revisar excepciones y reprogramaciones;
- clasificar cada solape para decisión institucional.

### Preparación

- conservar export y respaldo de las fuentes;
- fijar rango de migración y zona horaria;
- preparar carga idempotente en ambiente de prueba;
- definir reglas de privacidad de títulos;
- asignar responsables por rol y reversión.

### Validación

- reconciliar conteos por recurso y fecha;
- revisar muestras de recurrencias y excepciones;
- resolver o aceptar explícitamente solapamientos;
- comprobar disponibilidad desde usuario, administrador y pantalla pública;
- verificar que la importación no genere acciones o notificaciones indebidas.

### Congelamiento

- fijar fecha y hora oficial;
- comunicar que Google deja de recibir escrituras;
- registrar cambios producidos desde el último export;
- ejecutar export y carga final;
- impedir doble escritura.

### Operación posterior

- declarar Poli-REDI como fuente oficial;
- mantener Google en consulta por un plazo acordado;
- monitorear diariamente el periodo inicial;
- reconciliar incidentes con el export conservado;
- cerrar formalmente el legado.

## Criterios de aceptación

- [ ] Inventario completo y aprobado por operación.
- [ ] Mapeo unívoco de calendario a recurso.
- [ ] Recurrencias y excepciones expandidas y revisadas.
- [ ] Solapamientos resueltos o aceptados con motivo.
- [ ] Privacidad de títulos validada por audiencia.
- [ ] Fecha de congelamiento comunicada.
- [ ] Carga final idempotente y reconciliada.
- [ ] Plan de reversión probado y responsables asignados.
- [ ] Google queda sin nuevas escrituras después del corte.

## Reversión

Si la reconciliación falla, detener la activación, conservar Google como fuente temporal, revertir la carga con el procedimiento probado y comunicar el nuevo estado. No corregir eventos manualmente en ambos sistemas.

La preparación técnica sigue [05-instalacion-despliegue-recuperacion.md](../../05-instalacion-despliegue-recuperacion.md) y la decisión de alcance, [ADR-003](../../decisiones/ADR-003-alcance-mvp-y-exclusiones.md).
