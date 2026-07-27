SET XACT_ABORT ON;
GO

IF OBJECT_ID('dbo.users','U') IS NULL
    THROW 55000, 'Falta dbo.users.', 1;
IF COL_LENGTH('dbo.users','id') IS NULL OR COL_LENGTH('dbo.users','rut') IS NULL
    THROW 55005, 'dbo.users no contiene id/rut.', 1;
IF NOT EXISTS (
    SELECT 1 FROM sys.columns c JOIN sys.types t ON t.user_type_id=c.user_type_id
    WHERE c.object_id=OBJECT_ID('dbo.users') AND c.name='id'
      AND t.name='int' AND c.is_nullable=0
) OR NOT EXISTS (
    SELECT 1 FROM sys.columns c JOIN sys.types t ON t.user_type_id=c.user_type_id
    WHERE c.object_id=OBJECT_ID('dbo.users') AND c.name='rut'
      AND t.name='nvarchar' AND c.max_length>=20 AND c.is_nullable=1
)
    THROW 55006, 'dbo.users.id/rut tienen tipos o nulabilidad incompatibles.', 1;
IF OBJECT_ID('dbo.reservations','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policies','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_durations','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_resources','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_group_resources','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_scope_migrations','U') IS NULL
   OR OBJECT_ID('dbo.resources','U') IS NULL
   OR OBJECT_ID('dbo.availability_blocks','U') IS NULL
   OR OBJECT_ID('dbo.scheduled_activities','U') IS NULL
    THROW 55007, 'Faltan prerrequisitos del trigger canonico de reservas.', 1;
GO

DECLARE @normalized TABLE(user_id INT PRIMARY KEY, rut NVARCHAR(12) NULL);
INSERT INTO @normalized(user_id,rut)
SELECT id,
 CASE WHEN cleaned='' THEN NULL
      WHEN cleaned LIKE '%-%' THEN cleaned
      ELSE LEFT(cleaned,LEN(cleaned)-1)+'-'+RIGHT(cleaned,1)
 END
FROM dbo.users
CROSS APPLY(SELECT UPPER(REPLACE(REPLACE(LTRIM(RTRIM(COALESCE(rut,''))),'.',''),' ','')) cleaned) n;

IF EXISTS (
 SELECT 1 FROM @normalized
 WHERE rut IS NOT NULL AND (
   LEN(rut) NOT BETWEEN 9 AND 10 OR rut NOT LIKE '%-[0-9K]'
   OR rut LIKE '%[^0-9K-]%' OR LEN(rut)-LEN(REPLACE(rut,'-',''))<>1
 )
)
    THROW 55002, 'Hay RUT con formato no canonico. Corrija los datos antes de continuar.', 1;

IF EXISTS (
    SELECT 1
    FROM @normalized u
    CROSS APPLY (SELECT LEFT(u.rut, LEN(u.rut)-2) AS body, RIGHT(u.rut, 1) AS verifier) parts
    CROSS APPLY (
        SELECT SUM(
            CONVERT(INT, SUBSTRING(parts.body, LEN(parts.body)-n.n+1, 1))
            * (2 + ((n.n-1) % 6))
        ) AS weighted_sum
        FROM (VALUES (1),(2),(3),(4),(5),(6),(7),(8)) n(n)
        WHERE n.n <= LEN(parts.body)
    ) calculation
    CROSS APPLY (
        SELECT CASE 11-(calculation.weighted_sum % 11)
            WHEN 11 THEN '0'
            WHEN 10 THEN 'K'
            ELSE CONVERT(VARCHAR(1), 11-(calculation.weighted_sum % 11))
        END AS expected_verifier
    ) validation
    WHERE u.rut IS NOT NULL
      AND validation.expected_verifier <> parts.verifier
)
    THROW 55004, 'Hay RUT con digito verificador invalido. Corrija los datos antes de continuar.', 1;

IF EXISTS (SELECT rut FROM @normalized WHERE rut IS NOT NULL GROUP BY rut HAVING COUNT(*)>1)
    THROW 55001, 'Hay RUT duplicados. Corrija los datos antes de continuar.', 1;

BEGIN TRANSACTION;
UPDATE u SET rut=n.rut FROM dbo.users u JOIN @normalized n ON n.user_id=u.id
WHERE ISNULL(u.rut,'')<>ISNULL(n.rut,'');

IF EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE object_id = OBJECT_ID('dbo.users') AND name = 'ux_users_rut'
)
    DROP INDEX ux_users_rut ON dbo.users;

CREATE UNIQUE INDEX ux_users_rut ON dbo.users(rut) WHERE rut IS NOT NULL;

COMMIT TRANSACTION;
GO

CREATE OR ALTER TRIGGER dbo.trg_users_rut_write_once ON dbo.users AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (
        SELECT 1 FROM inserted i JOIN deleted d ON d.id = i.id
        WHERE NULLIF(LTRIM(RTRIM(d.rut)), '') IS NOT NULL
          AND (
              NULLIF(LTRIM(RTRIM(i.rut)), '') IS NULL
              OR UPPER(REPLACE(REPLACE(LTRIM(RTRIM(d.rut)),'.',''),' ',''))
                 <> UPPER(REPLACE(REPLACE(LTRIM(RTRIM(i.rut)),'.',''),' ',''))
          )
    )
        THROW 55003, 'El RUT no puede modificarse una vez registrado.', 1;
END;
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
            CONVERT(
                DATE,
                previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
            )
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
          CONVERT(
              DATE,
              previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
          )
      )
    ORDER BY DATEADD(
        DAY,
        previous_policy.request_frequency_days,
        CONVERT(
            DATE,
            previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
        )
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

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id=i.user_id INNER JOIN dbo.reservation_policy_group_resources g ON g.policy_id=i.policy_id AND g.resource_id=i.resource_id WHERE i.status IN ('PENDING','CONFIRMED') AND u.is_admin=0 AND NULLIF(LTRIM(RTRIM(u.rut)),'') IS NULL)
        THROW 51017, 'El usuario debe registrar su RUT antes de crear reservas.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.is_active = 0)
        THROW 51001, 'El recurso no esta activo.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.reservation_mode = 'INFORMATIVE')
        THROW 51002, 'El recurso es solo informativo y no permite reservas.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        INNER JOIN dbo.users u ON u.id = i.user_id
        WHERE i.status IN ('PENDING', 'CONFIRMED')
          AND r.reservation_mode = 'ADMIN_ONLY'
          AND u.is_admin = 0
    )
        THROW 51003, 'El recurso solo puede ser reservado por administradores.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
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
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.user_id = i.user_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND existing.status IN ('PENDING','CONFIRMED')
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51005, 'El usuario ya tiene una reserva confirmada en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > b.start_time
    )
        THROW 51006, 'El recurso esta bloqueado en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
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

CREATE OR ALTER TRIGGER dbo.trg_blocks_validate_conflicts
ON dbo.availability_blocks
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks existing ON existing.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND existing.is_active = 1
          AND existing.id <> i.id
          AND i.start_time < existing.end_time
          AND i.end_time > existing.start_time
    )
        THROW 51100, 'Ya existe un bloqueo activo para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations r ON r.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND r.status IN ('PENDING','CONFIRMED')
          AND i.start_time < DATEADD(MINUTE, r.duration_minutes, r.start_time)
          AND i.end_time > r.start_time
    )
        THROW 51101, 'El bloqueo se cruza con una reserva confirmada.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_scheduled_activities_validate_conflicts
ON dbo.scheduled_activities
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.scheduled_activities existing ON existing.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND existing.is_active = 1
          AND existing.id <> i.id
          AND i.start_time < existing.end_time
          AND i.end_time > existing.start_time
    )
        THROW 51200, 'Ya existe una actividad programada para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND i.end_time > b.start_time
    )
        THROW 51201, 'La actividad programada se cruza con un bloqueo activo.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations r ON r.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND r.status IN ('PENDING','CONFIRMED')
          AND i.start_time < DATEADD(MINUTE, r.duration_minutes, r.start_time)
          AND i.end_time > r.start_time
    )
        THROW 51202, 'La actividad programada se cruza con una reserva confirmada.', 1;
END;
GO

-- ============================================================
-- AUDIT AND NOTIFICATION TRIGGERS
-- ============================================================

-- Garantias complementarias para instalaciones migradas desde MVP1. Se
GO

SELECT
    CASE WHEN NOT EXISTS (
        SELECT rut FROM dbo.users WHERE rut IS NOT NULL
        GROUP BY rut HAVING COUNT(*) > 1
    ) THEN 1 ELSE 0 END AS rut_unique_ok,
    CASE WHEN OBJECT_ID('dbo.trg_users_rut_write_once','TR') IS NOT NULL
         THEN 1 ELSE 0 END AS rut_write_once_ok,
    CASE WHEN EXISTS (
        SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID('dbo.users')
          AND name='ux_users_rut' AND is_unique=1 AND has_filter=1
    ) THEN 1 ELSE 0 END AS rut_filtered_unique_index_ok,
    CASE WHEN NOT EXISTS (
        SELECT 1 FROM dbo.users WHERE rut IS NOT NULL
          AND (rut LIKE '%.%' OR rut LIKE '% %' OR rut NOT LIKE '%-[0-9K]')
    ) THEN 1 ELSE 0 END AS rut_canonical_data_ok,
    CASE WHEN OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_conflicts'))
              LIKE '%u.is_admin=0%NULLIF(LTRIM(RTRIM(u.rut))%'
         THEN 1 ELSE 0 END AS reservation_admin_rut_exemption_ok;
GO
