# Base de datos Poli-REDI — versión mejorada

Este paquete separa claramente **instalación limpia**, **evolución de una base existente**, **datos de desarrollo** y **diagnóstico**. Está diseñado para Azure SQL Database y SQL Server mediante T-SQL con separadores `GO`.

## Estructura

```text
poliredi_database_improved/
├── schema.sql                 # Estado canónico para una base nueva/local/CI
├── seed.sql                   # Catálogo y datos de desarrollo protegidos
├── seed_today_temp.sql        # Datos dinámicos para pruebas del día
├── verify.sql                 # Verificación posterior al despliegue, solo lectura
├── queries.sql                # Consultas operativas; cambios deshabilitados por defecto
├── drop.sql                   # Reset destructivo protegido
├── migrations/
│   ├── 001_...sql
│   ├── ...
│   └── 009_database_hardening_and_consistency.sql
└── docs/CHANGES.md
```

## Qué ejecutar

### Base nueva, local o CI

1. `schema.sql`
2. Habilitar el seed en **la misma sesión**:

```sql
EXEC sys.sp_set_session_context @key=N'allow_development_seed', @value=1;
```

3. `seed.sql`
4. `verify.sql`

`schema.sql` representa el estado final. No ejecute las migraciones `001` a `009` después de una instalación limpia salvo que esté probando explícitamente el camino de upgrade.

### Base existente

1. Crear backup/export recuperable.
2. Abrir una sesión nueva y verificar:

```sql
SELECT @@TRANCOUNT AS transaction_count, XACT_STATE() AS transaction_state;
```

Ambos valores deben ser `0`.

3. Ejecutar solo las migraciones pendientes, en orden numérico.
4. Después de cada archivo, revisar su `POSTCHECK`; todos los indicadores deben valer `1`.
5. Reejecutar la migración para comprobar idempotencia.
6. Ejecutar `verify.sql`.

La secuencia vigente es `001` → `009`. La migración `009` corrige bases que ya ejecutaron las versiones anteriores de `001` a `008`.

## Scripts protegidos

`seed.sql`, `seed_today_temp.sql` y `drop.sql` no se ejecutan accidentalmente. Requieren una bandera de sesión:

```sql
-- Datos de desarrollo
EXEC sys.sp_set_session_context @key=N'allow_development_seed', @value=1;

-- Datos dinámicos del día
EXEC sys.sp_set_session_context @key=N'allow_today_seed', @value=1;

-- Reset destructivo local/CI
EXEC sys.sp_set_session_context @key=N'allow_destructive_reset', @value=1;
```

Estas banderas duran únicamente durante la conexión actual.

## Reglas de seguridad

- No usar `drop.sql`, `schema.sql` ni `seed.sql` para recuperar una base productiva o la base única del proyecto.
- Conservar los separadores `GO` y usar SSMS, Azure Data Studio o `sqlcmd`.
- No editar manualmente vigencias de políticas para “arreglar” una migración fallida.
- Los secretos de códigos grupales se guardan cifrados; el seed no genera secretos reales.
- La revisión estática local no sustituye una prueba real en Azure SQL, incluida concurrencia e idempotencia.

## Mejoras principales

- Se añadió `009_database_hardening_and_consistency.sql`.
- La auditoría de `target_participants` ahora admite cambios desde y hacia `NULL`.
- Se eliminan triggers redundantes que duplicaban las reglas de `PENDING` y podían bloquear `OPEN_USE`.
- El RUT queda validado también en inserciones futuras, incluido el dígito verificador.
- Las duraciones permitidas se completan por política, no solo cuando la tabla completa está vacía.
- Se agregaron índices para búsquedas de solape, participantes y talleres.
- Una política cerrada ya no puede reabrirse ni cambiar su fecha de cierre.
- La auditoría de participantes es append-only y la auditoría de reservas evita duplicados internos por `updated_at`.
- Solo puede existir una inscripción `CONFIRMED` por usuario y taller.
- Las reservas grupales nuevas deben congelar una capacidad coherente con el recurso y las reservas no grupales no pueden llevar ese snapshot.
- Los scripts destructivos y de seed requieren habilitación explícita.
- Se agregó una verificación posterior al despliegue y consultas de mantenimiento seguras por defecto.
