# Poli-REDI - Plan de corte desde Google Calendar

## Objetivo

Definir el plan minimo para mover la operacion desde calendarios Google legados hacia Poli-REDI sin perder reservas, sin duplicar horarios y sin dejar al equipo operativo sin trazabilidad.

## Principio de corte

Poli-REDI debe convertirse en la fuente oficial de reservas. Google Calendar debe quedar solo como referencia historica o respaldo temporal durante la transicion.

Mientras ambos sistemas esten activos, existe riesgo de doble reserva. Por eso el corte debe tener fecha, responsables y reglas claras.

## Alcance MVP 1

Para MVP 1 no se implementa sincronizacion automatica con Google Calendar.

El alcance recomendado es operativo:

- Inventariar calendarios Google vigentes.
- Exportar o registrar reservas futuras relevantes.
- Cargar en Poli-REDI las reservas que deban mantenerse.
- Definir fecha y hora de congelamiento del calendario legado.
- Comunicar que nuevas reservas se crean solo en Poli-REDI.
- Mantener Google Calendar en modo consulta durante un periodo acotado.

## Datos a levantar

Antes del corte, registrar:

- Nombre de cada calendario Google usado por la operacion.
- Recurso Poli-REDI equivalente para cada calendario.
- Rango de fechas a migrar.
- Eventos futuros confirmados.
- Eventos recurrentes o institucionales.
- Responsables operativos de validar la migracion.

## Inventario recibido el 14-07-2026

Se analizaron cinco calendarios unicos exportados desde Google Calendar en formato iCal. La segunda copia de Sala Multiuso se excluye por ser un duplicado del mismo calendario.

| Calendario legado | Recurso Poli-REDI propuesto | Eventos base | Cobertura detectada | Observaciones |
| --- | --- | ---: | --- | --- |
| CANCHA 1 - 2026 | Cancha 1, Centro Deportivo | 91 | 18-03-2026 a 05-09-2026 | Incluye series recurrentes y eventos futuros. |
| CANCHA 2 - 2026 | Cancha 2, Centro Deportivo | 102 | 18-03-2026 a 15-07-2026 | Incluye series recurrentes hasta fines de julio. |
| CANCHA 3 - 2026 | Cancha 3, Centro Deportivo | 54 | 18-03-2026 a 08-08-2026 | Contiene una serie sin termino explicito que debe limitarse al torneo del 4 al 8 de agosto. |
| SALA MULTIUSO - 2026 | Sala Multiuso, Centro Deportivo | 13 | 06-04-2026 a 31-07-2026 | El ZIP fue recibido dos veces; solo se considera una copia. |
| SALA DE MUSCULACION | Gimnasio, Centro Deportivo | 2 | 01-08-2026 a 05-09-2026 | La ocurrencia del 1 de septiembre fue reprogramada al 5 de septiembre. |

No se recibieron exportaciones para Muro Escalada, Sala Spinning ni Piscina. Se debe confirmar si esos recursos no usaban calendario legado o si faltan archivos.

Los eventos importados se modelaran como `scheduled_activities`, no como reservas personales. En Disponibilidad, los administradores podran ver el titulo original y los usuarios normales veran solamente `Actividad institucional`, evitando exponer nombres personales incluidos en algunos titulos del calendario legado.

El corte propuesto para preparar la carga es el 14-07-2026. La fecha y hora oficial de congelamiento siguen pendientes de confirmacion operativa.

### Resultado de expansion preliminar

Al expandir las recurrencias entre el 14-07-2026 y el 30-09-2026 se obtienen 217 ocurrencias futuras:

| Recurso | Ocurrencias |
| --- | ---: |
| Cancha 1, Centro Deportivo | 81 |
| Cancha 2, Centro Deportivo | 52 |
| Cancha 3, Centro Deportivo | 56 |
| Sala Multiuso, Centro Deportivo | 26 |
| Gimnasio, Centro Deportivo | 2 |

La expansion detecto 14 solapamientos dentro del propio legado: 8 en Cancha 1 y 6 en Cancha 2. Tambien existen eventos que representan talleres ya cargados desde la planilla oficial, por ejemplo Judo, Esgrima, Pilates, Entrenamiento funcional y Aikido.

Por estas razones no se debe importar el iCal completo sin depuracion. Antes de generar la carga final se debe:

1. Excluir los talleres cuya fuente oficial ya es la planilla de talleres.
2. Resolver o consolidar los 14 solapamientos legados.
3. Confirmar el limite de la serie de voleibol del 4 al 8 de agosto.
4. Confirmar la fecha y hora oficial de congelamiento.
5. Generar una carga idempotente solo con las ocurrencias aprobadas.

## Estrategia recomendada

### 1. Inventario

Crear una tabla simple:

```txt
Calendario Google | Recurso Poli-REDI | Responsable | Observaciones
```

### 2. Congelamiento

Definir una fecha de corte:

```txt
Desde YYYY-MM-DD HH:mm, Google Calendar queda solo lectura para reservas nuevas.
```

Si Google Calendar no permite solo lectura para todos los usuarios, se debe comunicar la regla y retirar permisos de edicion cuando sea posible.

### 3. Migracion manual controlada

Para MVP 1, migrar manualmente las reservas futuras criticas:

- Crear la reserva en Poli-REDI.
- Confirmar que aparece en disponibilidad.
- Comparar contra el evento original de Google Calendar.
- Marcar el evento legado como migrado o mantener una nota operativa externa.

### 4. Validacion

Validar una muestra o el total de eventos migrados:

- Fecha.
- Hora inicio.
- Hora termino.
- Recurso.
- Actividad.
- Responsable o usuario asociado cuando exista.

### 5. Operacion post-corte

Despues del corte:

- Toda nueva reserva se crea en Poli-REDI.
- Google Calendar se usa solo para consulta historica temporal.
- Cualquier diferencia detectada se corrige en Poli-REDI.
- El equipo registra incidencias durante la primera semana.

## Riesgos

- Doble reserva si ambos sistemas aceptan cambios.
- Eventos recurrentes mal interpretados.
- Reservas antiguas sin usuario equivalente en Poli-REDI.
- Diferencias de zona horaria o formato de hora.
- Usuarios que sigan usando el flujo anterior por costumbre.

## Criterios de aceptacion para mover operacion

- [ ] Calendarios Google inventariados.
- [ ] Recursos equivalentes definidos en Poli-REDI.
- [ ] Reservas futuras criticas migradas o registradas.
- [ ] Fecha de congelamiento comunicada.
- [ ] Nuevas reservas se crean solo en Poli-REDI.
- [ ] Admin puede ver todas las reservas e historial en Poli-REDI.
- [ ] Usuario normal solo ve sus reservas e historial.
- [ ] Existe responsable operativo para resolver diferencias durante la transicion.

## Fuera de MVP 1

Queda para iteraciones futuras:

- Importador automatico desde `.ics` o Google Calendar API.
- Sincronizacion bidireccional.
- Marcado automatico de eventos migrados.
- Auditoria avanzada de diferencias entre Google Calendar y Poli-REDI.
