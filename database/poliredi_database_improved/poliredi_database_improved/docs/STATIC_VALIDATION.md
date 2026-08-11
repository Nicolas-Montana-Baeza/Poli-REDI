# Validación estática

**Fecha:** 2026-08-05  
**Alcance:** revisión local de estructura T-SQL; no corresponde a una ejecución contra Azure SQL Database o SQL Server.

## Resultado

- 15 archivos SQL revisados.
- 238 lotes no vacíos separados mediante `GO`.
- 51 definiciones de triggers y 4 definiciones de vistas analizadas.
- Secuencia de migraciones continua: `001` a `009`.
- No se detectaron paréntesis desbalanceados ni definiciones `CREATE OR ALTER TRIGGER/VIEW` fuera del inicio de su lote.
- No existen definiciones duplicadas de triggers en `schema.sql`.
- Los triggers redundantes de conflictos `PENDING` no forman parte del esquema canónico.
- Los RUT literales incluidos en `seed.sql` tienen dígito verificador válido.
- El hash SHA-256 documentado en la migración `007` coincide con su payload canónico.
- El `INSERT` de reservas de desarrollo conserva la misma cantidad de columnas y valores en sus 8 filas.
- Los scripts de seed y reset incluyen guardas explícitas mediante `SESSION_CONTEXT`.

## Validaciones pendientes en una instancia real

1. Ejecutar `schema.sql` en una base vacía y luego `verify.sql`.
2. Ejecutar `001` a `009` sobre una copia representativa de la base existente.
3. Reejecutar cada migración para comprobar idempotencia real.
4. Probar concurrencia, bloqueos y posibles deadlocks.
5. Probar solapes, intervalos contiguos, `OPEN_USE`, confirmaciones y talleres.
6. Verificar planes de ejecución e índices con volumen representativo.
7. Ensayar recuperación desde backup/export antes del despliegue definitivo.
