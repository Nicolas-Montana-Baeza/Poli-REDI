# Plan de corte desde Google Calendar hacia Poli-REDI

**Estado:** PLAN OPERATIVO; fecha oficial de congelamiento pendiente

## 1. Objetivo

Convertir Poli-REDI en la fuente oficial de reservas sin perder eventos futuros, duplicar horarios ni eliminar trazabilidad del sistema legado.

## 2. Principio de corte

No mantener ambos sistemas como fuentes activas por un período indefinido. Google Calendar debe quedar en consulta o respaldo temporal después de una fecha y hora formal de congelamiento.

## 3. Inventario documentado

Se recibieron cinco calendarios únicos. La segunda copia de Sala Multiuso era duplicada.

| Calendario legado | Destino | Eventos base | Cobertura observada |
|---|---|---:|---|
| Cancha 1 | Cancha 1 | 91 | 2026-03-18 a 2026-09-05 |
| Cancha 2 | Cancha 2 | 102 | 2026-03-18 a 2026-07-15 |
| Cancha 3 | Cancha 3 | 54 | 2026-03-18 a 2026-08-08 |
| Sala Multiuso | Sala Multiuso | 13 | 2026-04-06 a 2026-07-31 |
| Sala de Musculación | Gimnasio | 2 | 2026-08-01 a 2026-09-05 |

No se recibieron exportaciones para Muro Escalada, Sala Spinning ni Piscina; debe confirmarse si no existían calendarios o faltan archivos.

## 4. Resultado preliminar

La expansión entre 2026-07-14 y 2026-09-30 produjo 217 ocurrencias futuras y detectó 14 solapamientos internos del legado: 8 en Cancha 1 y 6 en Cancha 2.

Los eventos deben modelarse como `scheduled_activities`, no como reservas personales. Para usuarios normales, el título debe sanitizarse como actividad institucional cuando contenga información privada.

## 5. Fases

### Fase 1 — Inventario y limpieza

- confirmar calendarios y responsables;
- resolver duplicados;
- limitar recurrencias sin término;
- revisar eventos reprogramados;
- clasificar solapamientos.

### Fase 2 — Preparación

- crear backup/export de fuentes;
- mapear cada calendario a un recurso;
- definir rango de migración;
- preparar carga en ambiente de prueba;
- validar privacidad de títulos.

### Fase 3 — Validación

- comparar conteos por recurso;
- revisar ocurrencias y excepciones;
- resolver conflictos con responsable operativo;
- comprobar disponibilidad desde usuario y administrador.

### Fase 4 — Congelamiento

- fijar fecha y hora oficial;
- comunicar que no se crean nuevos eventos en Google;
- registrar cambios de última hora;
- ejecutar carga final.

### Fase 5 — Operación post-corte

- Poli-REDI como única fuente de escritura;
- Google Calendar en modo consulta durante plazo acotado;
- monitoreo diario inicial;
- cierre formal y conservación del export.

## 6. Criterios de aceptación

- [ ] inventario completo y aprobado;
- [ ] mapeo unívoco de recursos;
- [ ] recurrencias expandidas y revisadas;
- [ ] solapamientos resueltos o aceptados explícitamente;
- [ ] privacidad de títulos validada;
- [ ] fecha de congelamiento comunicada;
- [ ] carga final reconciliada;
- [ ] responsables y plan de reversa definidos.

## 7. Fuera de alcance

La sincronización automática bidireccional con Google Calendar no forma parte del MVP 1 documentado.
