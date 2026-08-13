# ADR-002 — Evolución de una base de datos única

**Audiencia:** Arquitecto, Backend, QA y DevOps

**Propósito:** limitar el riesgo de cambios de esquema sin crear otra base

**Estado:** aceptada el 2026-08-11

**Fuente:** restricción operativa, scripts existentes y auditoría de migraciones

## Resumen

Poli-REDI evoluciona la base existente mediante migraciones incrementales, recuperables e idempotentes. No se crea una segunda base productiva ni se ejecutan scripts destructivos. Las migraciones aprobadas son `001`–`008`; `009` permanece excluida.

## Contexto

MVP 2 requiere columnas, índices, triggers e historial adicionales. Errores de preflight demostraron que asumir una definición de trigger o un esquema previo puede interrumpir la actualización y poner en riesgo la única base disponible.

## Decisión

1. Ejecutar migraciones en orden numérico, una por vez.
2. Antes de cada cambio, identificar versión, crear respaldo recuperable y ejecutar precheck.
3. Si el objeto existente no coincide con la definición esperada, detenerse sin modificarlo.
4. Después, ejecutar postcheck de esquema y datos, prueba funcional e idempotencia.
5. Probar primero en una copia recuperable y luego en Azure SQL.
6. Conservar snapshots para que cambios administrativos posteriores sean prospectivos.
7. No usar `drop.sql`, reinstalación limpia ni DDL improvisado para corregir producción.

## Secuencia aprobada

`001` → `002` → `003` → `004` → `005` → `006` → `007` → `008`.

`009_database_hardening_and_consistency.sql` es una propuesta no aprobada. Requeriría un ADR posterior, comparación con `007` y `008`, análisis de datos, preflight, reversión y criterio explícito de adopción.

## Recuperación

Ante fallo, detener la publicación, abrir una sesión nueva, verificar transacciones y objetos, y ejecutar la reversión aprobada o restaurar el respaldo. La aplicación solo vuelve a servicio cuando versión de código y esquema son compatibles.

## Consecuencias

- Aumenta el tiempo de preparación, pero reduce pérdida y divergencia de datos.
- `007` y `008` son dependencias reales de cierre de MVP 2.
- Una migración que “termina sin error” no está validada sin postcheck y prueba funcional.

Ver procedimiento en [04-base-de-datos-y-migraciones.md](../04-base-de-datos-y-migraciones.md) y [05-instalacion-despliegue-recuperacion.md](../05-instalacion-despliegue-recuperacion.md).
