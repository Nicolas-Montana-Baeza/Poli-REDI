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
