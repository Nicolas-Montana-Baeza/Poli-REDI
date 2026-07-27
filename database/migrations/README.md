# Migraciones de base de datos

El archivo principal [../schema.sql](../schema.sql) representa el estado canónico de la base de datos y sirve como referencia para lectura y uso diario. Las migraciones en esta carpeta quedan como deltas incrementales, reejecutables y ordenados para evolución controlada.

## `001_mvp2_group_participants.sql`

Ejecutar el archivo completo con una herramienta que interprete separadores
`GO`, como SSMS, Azure Data Studio o `sqlcmd`. No agrupar todos los batches
dentro de una sola llamada `EXEC`.

La migracion es acumulativa y reejecutable. SQL Server compila cada batch antes
de ejecutar un `ALTER TABLE`, por lo que las columnas nuevas se crean con SQL
dinamico y no se referencian hasta despues de un separador `GO`.

Los `GO` impiden una transaccion unica para todo el archivo. Cada fase
estructural es idempotente y puede quedar aplicada si una fase posterior falla.
La politica prospectiva se publica exclusivamente en la ultima fase, que si es
atomica: ante un error no deja cerrada la politica vigente.

La migracion no asigna modalidad grupal, codigo ni capacidad a reservas
historicas. El unico backfill permitido es el `policy_id` tecnico ya documentado
para reservas MVP1 y las colecciones base de la politica bootstrap.

## Recuperacion de la unica base despues de un intento fallido

1. Crear primero un backup o export. No continuar sin una copia recuperable.
2. Abrir una sesion nueva y comprobar:

   ```sql
   SELECT @@TRANCOUNT AS transaction_count, XACT_STATE() AS transaction_state;
   ```

   Ambos valores deben ser `0`. Si no lo son, cerrar la sesion y abrir otra.
3. Inspeccionar el estado parcial sin modificarlo:

   ```sql
   SELECT COL_LENGTH('dbo.reservation_policies','opening_minute') AS opening_minute;
   SELECT COL_LENGTH('dbo.reservation_policies','is_published') AS is_published;
   SELECT COL_LENGTH('dbo.reservation_policies','idempotency_key') AS idempotency_key;
   SELECT COL_LENGTH('dbo.reservations','group_capacity_snapshot') AS capacity_snapshot;
   SELECT OBJECT_ID('dbo.reservation_policy_group_resources','U') AS group_resources;
   SELECT name
   FROM sys.columns
   WHERE object_id = OBJECT_ID('dbo.reservation_policies')
     AND name IN ('opening_minute','is_published','idempotency_key');
   SELECT id,effective_from,effective_to
   FROM dbo.reservation_policies ORDER BY id;
   ```
4. Conservar los separadores `GO` al ejecutar el archivo.
5. Ejecutar solamente `001_mvp2_group_participants.sql`. No ejecutar
   `drop.sql`, `schema.sql` ni `seed.sql` durante esta recuperacion.
6. Revisar el `POSTCHECK`: `policy_columns_ok`, `snapshot_ok`,
   `group_tables_ok` y `policy_published_ok` deben valer `1`.

Si aparece el error de politica parcial `52004`, no editar vigencias ni publicar
manualmente. Conservar el backup y revisar la fila cuya `idempotency_key` es
`migration-mvp2-group-v1`. La migracion se detiene antes de alterar la politica
vigente. Esa fila requiere revision tecnica; despues se ejecuta nuevamente solo
la migracion.

Despues de completar y verificar `001`, ejecutar
`002_mvp2_target_participants.sql`. Esta segunda migracion es acumulativa, no
rellena reservas historicas y conserva `NULL` con semantica efectiva igual a la
capacidad congelada de cada solicitud.

## `002_mvp2_target_participants.sql`

1. Crear un backup o export recuperable antes de comenzar.
2. Abrir una sesion nueva y confirmar `@@TRANCOUNT = 0` y `XACT_STATE() = 0`.
3. Usar SSMS, Azure Data Studio o `sqlcmd`, conservando todos los separadores
   `GO`.
4. Ejecutar solamente `002_mvp2_target_participants.sql`, y solo despues de que
   el `POSTCHECK` de `001` haya informado todos sus indicadores en `1`.
5. Si una fase falla, no ejecutar `drop.sql`, `schema.sql` ni rellenar datos
   historicos. Abrir otra sesion, inspeccionar columna, constraint, tabla y
   triggers, y volver a ejecutar solamente `002`; todas sus fases son
   reejecutables y toleran estado parcial.

   ```sql
   SELECT COL_LENGTH('dbo.reservations','target_participants') AS target_column;
   SELECT OBJECT_ID('dbo.ck_reservations_target_participants','C') AS target_constraint;
   SELECT OBJECT_ID('dbo.reservation_target_audit','U') AS target_audit;
   SELECT OBJECT_ID('dbo.trg_reservations_target_validate','TR') AS validation_trigger;
   SELECT OBJECT_ID('dbo.trg_reservation_target_audit_append_only','TR') AS append_only_trigger;
   ```
6. El `POSTCHECK` final debe mostrar en `1`: `target_column_ok`,
   `target_constraint_ok`, `target_validation_trigger_ok`, `target_audit_ok` y
   `target_audit_append_only_ok`.

## `003_open_use_frequency_scope.sql`

Ejecutar despues de `001` y `002`. Esta migracion hace que `OPEN_USE` no
consuma la frecuencia configurada ni sea limitado por reservas normales
anteriores. Conserva la prohibicion de reservas solapadas para el mismo usuario;
los limites contiguos, por ejemplo `12:00-13:00` y `13:00-14:00`, no se
consideran solape.

La migracion es reejecutable e instala la definicion canonica completa con
`CREATE OR ALTER TRIGGER`; no intenta interpretar el encabezado almacenado por
SQL Server. El `POSTCHECK` debe mostrar `open_use_frequency_scope_ok = 1` y
`user_overlap_guard_ok = 1`.

## `004_group_flow_completion.sql`

Ejecutar solamente después de verificar 001, 002 y 003. Es incremental,
reejecutable y no genera secretos ni cambia estados históricos. Añade el
almacenamiento cifrado del código y la marca idempotente de expiración.
La migracion separa tablas, validacion de estado parcial y constraints mediante
`GO`. Si se interrumpe, abrir una sesion nueva, confirmar `@@TRANCOUNT = 0` y
`XACT_STATE() = 0`, conservar el backup y ejecutar nuevamente solo 004. Las
tablas con columnas compatibles y constraints faltantes se completan; columnas,
tipos, nulabilidad, precision o PK divergentes detienen el proceso con
`54001`-`54004` sin reinterpretar datos.

El `POSTCHECK` devuelve doce indicadores para tabla, columnas/tipos/nulabilidad,
PK/unicidad, FK, CHECK y
DEFAULT. Todos deben valer `1`. No continuar con el despliegue si alguno vale
`0`.

## `005_rut_integrity_and_admin_exemption.sql`

Ejecutar después de `004`. Normaliza blancos y mayúsculas, se detiene ante
formatos inválidos o duplicados, reinstala el índice único filtrado y protege el
RUT como dato de escritura única. También reinstala completa la definición
canónica del trigger de reservas (incluidas las reglas de 003 y 004), con
exención administrativa solamente para la exigencia de RUT del titular grupal.
Su `POSTCHECK` confirma la protección de RUT y esa cláusula administrativa.

Crear un backup recuperable antes de ejecutarla. El preflight valida estructura, normalización, formato, dígito verificador y duplicados antes de abrir la transacción. La verificación en Azure SQL real, incluida la reejecución idempotente, sigue pendiente.

Si 005 falla antes de `BEGIN TRANSACTION`, no modificó ningún RUT. Si falla
dentro de la transacción, abrir una sesión nueva, confirmar
`@@TRANCOUNT = 0` y `XACT_STATE() = 0`, corregir solo el dato o estructura
reportada y volver a ejecutar 005 completo. El preflight acepta RUT legacy sin
guion y calcula su representación canónica antes de escribir.

## `006_workshop_occurrences.sql`

Recuperación parcial: abrir una sesión nueva y revisar tabla, columnas, PK, FK,
CHECK, UNIQUE, índice y trigger con los nombres del `POSTCHECK`. La migración
completa objetos faltantes cuando la estructura es compatible y se detiene con
`56000`-`56011` ante divergencias; no borrar ni recrear la tabla para recuperar.
El repositorio es el camino soportado para inscripciones y toma locks en orden
usuario→taller. El trigger defiende escrituras externas de forma set-based,
pero no sustituye ese orden para evitar deadlocks.

Ejecutar después de `005`. Crea los horarios normalizados y carga el mapeo
explícito del catálogo institucional. La migración falla de forma segura si
encuentra un taller activo fuera de ese catálogo o si alguno queda sin horario.

Crear un backup antes de ejecutarla y revisar todos los indicadores del
`POSTCHECK`. La semántica instalada compara únicamente talleres activos e
inscripciones `CONFIRMED`, usa intervalos semiabiertos y admite múltiples días;
inscripciones `CANCELLED` y talleres inactivos no bloquean. DDL, idempotencia y
carreras reales en Azure SQL continúan pendientes.
