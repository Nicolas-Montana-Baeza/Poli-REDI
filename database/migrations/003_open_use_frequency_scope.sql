SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

/*
OPEN_USE no consume la frecuencia de solicitudes y tampoco queda limitado por
solicitudes RESERVABLE anteriores. La incompatibilidad por solape del mismo
usuario permanece activa.

La migracion instala la definicion canonica completa. No interpreta ni
reconstruye el encabezado textual que OBJECT_DEFINITION puede devolver como
CREATE, ALTER, CREATE OR ALTER o con identificadores delimitados.
*/

IF OBJECT_ID('dbo.trg_reservations_validate_conflicts', 'TR') IS NULL
    THROW 53000, 'Preflight: falta trg_reservations_validate_conflicts.', 1;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_validate_conflicts
ON dbo.reservations
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
        WHERE NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_durations d WHERE d.policy_id = i.policy_id AND d.duration_minutes = i.duration_minutes)
           OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) < p.opening_minute
           OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) >= p.closing_minute
           OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) % p.slot_interval_minutes <> 0
           OR (DATEPART(HOUR, DATEADD(MINUTE, i.duration_minutes, i.start_time)) * 60 + DATEPART(MINUTE, DATEADD(MINUTE, i.duration_minutes, i.start_time))) > p.closing_minute
           OR CONVERT(DATE, i.start_time) <> CONVERT(DATE, DATEADD(MINUTE, i.duration_minutes, i.start_time))
    )
        THROW 51015, 'El horario o la duracion no estan permitidos por la politica asociada.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
        WHERE (p.idempotency_key IS NOT NULL OR EXISTS (SELECT 1 FROM dbo.reservation_policy_scope_migrations m WHERE m.policy_id = p.id))
          AND NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_resources pr WHERE pr.policy_id = i.policy_id AND pr.resource_id = i.resource_id)
    )
        THROW 51016, 'El recurso no esta permitido por la politica asociada.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        LEFT JOIN deleted d ON d.id = i.id
        INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
        WHERE d.id IS NULL
          AND (
              CONVERT(DATE, i.start_time) < CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
              OR CONVERT(DATE, i.start_time) >= DATEADD(DAY, p.reservable_window_days, CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'))
          )
    )
        THROW 51009, 'La fecha solicitada esta fuera de la ventana reservable vigente.', 1;

    DECLARE @next_request_date DATE;

    SELECT TOP (1)
        @next_request_date = DATEADD(
            DAY,
            previous_policy.request_frequency_days,
            CONVERT(DATE, previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
        )
    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    INNER JOIN dbo.resources inserted_resource ON inserted_resource.id = i.resource_id
    INNER JOIN dbo.reservations previous WITH (UPDLOCK, HOLDLOCK)
        ON previous.user_id = i.user_id
       AND previous.id <> i.id
       AND previous.status IN ('PENDING', 'CONFIRMED')
    INNER JOIN dbo.resources previous_resource ON previous_resource.id = previous.resource_id
    INNER JOIN dbo.reservation_policies previous_policy ON previous_policy.id = previous.policy_id
    WHERE d.id IS NULL
      AND inserted_resource.reservation_mode <> 'OPEN_USE'
      AND previous_resource.reservation_mode <> 'OPEN_USE'
      AND CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time') < DATEADD(
          DAY,
          previous_policy.request_frequency_days,
          CONVERT(DATE, previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
      )
    ORDER BY DATEADD(
        DAY,
        previous_policy.request_frequency_days,
        CONVERT(DATE, previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
    ) DESC, previous.id DESC;

    IF @next_request_date IS NOT NULL
    BEGIN
        DECLARE @frequency_message NVARCHAR(2048) = CONCAT(
            N'Ya existe una solicitud vigente. Proxima fecha permitida: ',
            CONVERT(NVARCHAR(10), @next_request_date, 23),
            N'.'
        );
        THROW 51010, @frequency_message, 1;
    END;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id = i.user_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND u.is_blocked = 1)
        THROW 51000, 'El usuario se encuentra bloqueado y no puede crear reservas.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id=i.user_id INNER JOIN dbo.reservation_policy_group_resources g ON g.policy_id=i.policy_id AND g.resource_id=i.resource_id WHERE i.status IN ('PENDING','CONFIRMED') AND NULLIF(LTRIM(RTRIM(u.rut)),'') IS NULL)
        THROW 51017, 'El usuario debe registrar su RUT antes de crear reservas.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.is_active = 0)
        THROW 51001, 'El recurso no esta activo.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.reservation_mode = 'INFORMATIVE')
        THROW 51002, 'El recurso es solo informativo y no permite reservas.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        INNER JOIN dbo.users u ON u.id = i.user_id
        WHERE i.status IN ('PENDING', 'CONFIRMED')
          AND r.reservation_mode = 'ADMIN_ONLY'
          AND u.is_admin = 0
    )
        THROW 51003, 'El recurso solo puede ser reservado por administradores.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.resource_id = i.resource_id
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND r.reservation_mode <> 'OPEN_USE'
          AND existing.status IN ('PENDING','CONFIRMED')
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51004, 'Ya existe una reserva confirmada para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.user_id = i.user_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND existing.status IN ('PENDING','CONFIRMED')
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51005, 'El usuario ya tiene una reserva confirmada en ese horario.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > b.start_time
    )
        THROW 51006, 'El recurso esta bloqueado en ese horario.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted i
        INNER JOIN dbo.scheduled_activities s ON s.resource_id = i.resource_id
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND r.reservation_mode <> 'OPEN_USE'
          AND s.is_active = 1
          AND i.start_time < s.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > s.start_time
    )
        THROW 51007, 'El recurso tiene una actividad programada en ese horario.', 1;
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
