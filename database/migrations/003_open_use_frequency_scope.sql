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
    DECLARE @frequency_start INT = CHARINDEX(N'DECLARE @next_request_date DATE;', @definition);
    DECLARE @frequency_end INT = CHARINDEX(
        N'IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id = i.user_id',
        @definition,
        @frequency_start
    );

    IF @frequency_start = 0 OR @frequency_end = 0 OR @frequency_end <= @frequency_start
        THROW 53001, 'Preflight: definicion inesperada del trigger; no se modifico.', 1;

    DECLARE @new_frequency NVARCHAR(MAX) =
N'DECLARE @next_request_date DATE;

    SELECT TOP (1)
        @next_request_date = DATEADD(
            DAY,
            previous_policy.request_frequency_days,
            CONVERT(
                DATE,
                previous.created_at AT TIME ZONE ''UTC'' AT TIME ZONE ''Pacific SA Standard Time''
            )
        )
    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    INNER JOIN dbo.resources inserted_resource ON inserted_resource.id = i.resource_id
    INNER JOIN dbo.reservations previous WITH (UPDLOCK, HOLDLOCK)
        ON previous.user_id = i.user_id
       AND previous.id <> i.id
       AND previous.status IN (''PENDING'', ''CONFIRMED'')
    INNER JOIN dbo.resources previous_resource ON previous_resource.id = previous.resource_id
    INNER JOIN dbo.reservation_policies previous_policy ON previous_policy.id = previous.policy_id
    WHERE d.id IS NULL
      AND inserted_resource.reservation_mode <> ''OPEN_USE''
      AND previous_resource.reservation_mode <> ''OPEN_USE''
      AND CONVERT(DATE, i.created_at AT TIME ZONE ''UTC'' AT TIME ZONE ''Pacific SA Standard Time'') < DATEADD(
          DAY,
          previous_policy.request_frequency_days,
          CONVERT(
              DATE,
              previous.created_at AT TIME ZONE ''UTC'' AT TIME ZONE ''Pacific SA Standard Time''
          )
      )
    ORDER BY DATEADD(
        DAY,
        previous_policy.request_frequency_days,
        CONVERT(
            DATE,
            previous.created_at AT TIME ZONE ''UTC'' AT TIME ZONE ''Pacific SA Standard Time''
        )
    ) DESC, previous.id DESC;

    IF @next_request_date IS NOT NULL
    BEGIN
        DECLARE @frequency_message NVARCHAR(2048) = CONCAT(
            N''Ya existe una solicitud vigente. Proxima fecha permitida: '',
            CONVERT(NVARCHAR(10), @next_request_date, 23),
            N''.''
        );
        THROW 51010, @frequency_message, 1;
    END;

    ';

    SET @definition =
        LEFT(@definition, @frequency_start - 1)
        + @new_frequency
        + SUBSTRING(@definition, @frequency_end, LEN(@definition) - @frequency_end + 1);

    DECLARE @create_or_alter INT = CHARINDEX(N'CREATE OR ALTER TRIGGER', UPPER(@definition));
    DECLARE @create_only INT = CHARINDEX(N'CREATE TRIGGER', UPPER(@definition));
    IF @create_or_alter > 0
        SET @definition = STUFF(@definition, @create_or_alter, LEN(N'CREATE OR ALTER TRIGGER'), N'ALTER TRIGGER');
    ELSE IF @create_only > 0
        SET @definition = STUFF(@definition, @create_only, LEN(N'CREATE TRIGGER'), N'ALTER TRIGGER');
    ELSE
        THROW 53002, 'Preflight: no se pudo identificar la sentencia CREATE del trigger.', 1;

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
