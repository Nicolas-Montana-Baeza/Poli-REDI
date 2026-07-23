SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

/*
OPEN_USE no consume la frecuencia de solicitudes y tampoco queda limitado por
solicitudes RESERVABLE anteriores. La incompatibilidad por solape del mismo
usuario permanece en trg_reservations_validate_conflicts.

Esta migracion modifica solamente el bloque de frecuencia del trigger vigente.
Es reejecutable y se detiene si la definicion encontrada no coincide con la
version esperada, para no reemplazar silenciosamente cambios posteriores.
*/

IF OBJECT_ID('dbo.trg_reservations_validate_conflicts', 'TR') IS NULL
    THROW 53000, 'Preflight: falta trg_reservations_validate_conflicts.', 1;
GO

DECLARE @definition NVARCHAR(MAX) =
    REPLACE(OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts')), CHAR(13), N'');

IF @definition LIKE N'%inserted_resource.reservation_mode <> ''OPEN_USE''%'
   AND @definition LIKE N'%previous_resource.reservation_mode <> ''OPEN_USE''%'
BEGIN
    PRINT '003 ya estaba aplicada.';
END
ELSE
BEGIN
    DECLARE @old_from NVARCHAR(MAX) =
N'    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    INNER JOIN dbo.reservations previous WITH (UPDLOCK, HOLDLOCK)';

    DECLARE @new_from NVARCHAR(MAX) =
N'    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    INNER JOIN dbo.resources inserted_resource ON inserted_resource.id = i.resource_id
    INNER JOIN dbo.reservations previous WITH (UPDLOCK, HOLDLOCK)';

    DECLARE @old_policy_join NVARCHAR(MAX) =
N'       AND previous.status IN (''PENDING'', ''CONFIRMED'')
    INNER JOIN dbo.reservation_policies previous_policy ON previous_policy.id = previous.policy_id
    WHERE d.id IS NULL
      AND CONVERT(DATE, i.created_at';

    DECLARE @new_policy_join NVARCHAR(MAX) =
N'       AND previous.status IN (''PENDING'', ''CONFIRMED'')
    INNER JOIN dbo.resources previous_resource ON previous_resource.id = previous.resource_id
    INNER JOIN dbo.reservation_policies previous_policy ON previous_policy.id = previous.policy_id
    WHERE d.id IS NULL
      AND inserted_resource.reservation_mode <> ''OPEN_USE''
      AND previous_resource.reservation_mode <> ''OPEN_USE''
      AND CONVERT(DATE, i.created_at';

    IF CHARINDEX(@old_from, @definition) = 0 OR CHARINDEX(@old_policy_join, @definition) = 0
        THROW 53001, 'Preflight: definicion inesperada del trigger; no se modifico.', 1;

    SET @definition = REPLACE(@definition, @old_from, @new_from);
    SET @definition = REPLACE(@definition, @old_policy_join, @new_policy_join);
    EXEC sys.sp_executesql @definition;
END;
GO

SELECT
    CASE WHEN OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts'))
                   LIKE N'%inserted_resource.reservation_mode <> ''OPEN_USE''%'
          AND OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts'))
                   LIKE N'%previous_resource.reservation_mode <> ''OPEN_USE''%'
         THEN 1 ELSE 0 END AS open_use_frequency_scope_ok,
    CASE WHEN OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts'))
                   LIKE N'%existing.user_id = i.user_id%'
          AND OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts'))
                   LIKE N'%i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)%'
         THEN 1 ELSE 0 END AS user_overlap_guard_ok;
GO
