# Migraciones de base de datos

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
