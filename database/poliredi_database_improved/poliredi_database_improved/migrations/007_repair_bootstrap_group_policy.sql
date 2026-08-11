-- ============================================================================
-- POLI-REDI | Azure SQL Database / SQL Server
-- Archivo: 007_repair_bootstrap_group_policy.sql
-- Propósito: Reparar de forma cerrada la política bootstrap inequívoca.
-- Reejecución: Idempotente; revisar PRECHECK y POSTCHECK.
-- Requiere cliente con soporte para separadores GO.
-- ============================================================================

SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET ANSI_PADDING ON;
SET ANSI_WARNINGS ON;
SET ARITHABORT ON;
SET CONCAT_NULL_YIELDS_NULL ON;
SET NUMERIC_ROUNDABORT OFF;
SET XACT_ABORT ON;
SET NOCOUNT ON;
GO

/*
Repara exclusivamente la politica bootstrap inequívoca creada por schema.sql.
No intenta inferir una politica administrada ni completar configuraciones
parciales. Toda divergencia de identidad, modo, alcance o huella falla cerrada.

Payload canonico trazable:
repair-bootstrap-group-policy-v2|source=bootstrap-19000101|groups=1,2,7|allowed=1,2,3,4,5,6,7,8
SHA-256:
8a33d6dbedf56c20dbb857c3579ed986ccfce686afb60278dff06190aafe6ed3
*/

IF OBJECT_ID('dbo.reservation_policies', 'U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_resources', 'U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_group_resources', 'U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_durations', 'U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_scope_migrations', 'U') IS NULL
   OR OBJECT_ID('dbo.resources', 'U') IS NULL
    THROW 57000, 'Preflight: faltan objetos de politica, alcance o recursos.', 1;

IF COL_LENGTH('dbo.reservation_policies', 'idempotency_key') IS NULL
   OR COL_LENGTH('dbo.reservation_policies', 'idempotency_payload_hash') IS NULL
   OR COL_LENGTH('dbo.reservation_policies', 'is_published') IS NULL
    THROW 57000, 'Preflight: faltan columnas de trazabilidad de politica.', 1;
GO

IF EXISTS (
    SELECT 1
    FROM (
        VALUES
            (1, 2, N'Cancha 1, Centro Deportivo', N'Cancha', N'RESERVABLE', 22),
            (2, 2, N'Cancha 2, Centro Deportivo', N'Cancha', N'RESERVABLE', 22),
            (3, 2, N'Muro Escalada, Centro Deportivo', N'Muro Escalada', N'RESERVABLE', 20),
            (4, 2, N'Sala Spinning, Centro Deportivo', N'Sala', N'RESERVABLE', 25),
            (5, 2, N'Piscina, Centro Deportivo', N'Piscina', N'OPEN_USE', 20),
            (6, 2, N'Sala Multiuso, Centro Deportivo', N'Sala', N'ADMIN_ONLY', 25),
            (7, 2, N'Cancha 3, Centro Deportivo', N'Cancha', N'RESERVABLE', 22),
            (8, 2, N'Gimnasio, Centro Deportivo', N'Gimnasio', N'OPEN_USE', 40)
    ) expected(resource_id, venue_id, resource_name, resource_type, reservation_mode, capacity)
    LEFT JOIN dbo.resources actual ON actual.id = expected.resource_id
    WHERE actual.id IS NULL
       OR actual.venue_id <> expected.venue_id
       OR actual.name <> expected.resource_name
       OR actual.type <> expected.resource_type
       OR actual.reservation_mode <> expected.reservation_mode
       OR actual.capacity <> expected.capacity
       OR actual.is_active <> 1
)
    THROW 57001, 'Preflight: la identidad o modo de los recursos bootstrap diverge.', 1;
GO

BEGIN TRY
    SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
    BEGIN TRANSACTION;

    DECLARE @key NVARCHAR(100) = N'repair-bootstrap-group-policy-v1';
    DECLARE @payload_hash CHAR(64) = '8a33d6dbedf56c20dbb857c3579ed986ccfce686afb60278dff06190aafe6ed3';
    DECLARE @now DATETIME2(0) = SYSUTCDATETIME();
    DECLARE @current INT;
    DECLARE @new INT;
    DECLARE @replayed BIT = 0;

    SELECT @new = id
    FROM dbo.reservation_policies WITH (UPDLOCK, HOLDLOCK)
    WHERE idempotency_key = @key;

    IF @new IS NOT NULL
    BEGIN
        SET @replayed = 1;

        IF NOT EXISTS (
            SELECT 1
            FROM dbo.reservation_policies
            WHERE id = @new
              AND reservable_window_days = 7
              AND request_frequency_days = 7
              AND confirmation_deadline_minutes = 60
              AND minimum_participants = 10
              AND opening_minute = 480
              AND closing_minute = 1320
              AND slot_interval_minutes = 15
              AND idempotency_payload_hash = @payload_hash
              AND effective_to IS NULL
              AND is_published = 1
        )
            THROW 57005, 'Replay: la politica reparada no coincide con su huella.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_durations WHERE policy_id = @new) <> 6
           OR EXISTS (
               SELECT 1 FROM dbo.reservation_policy_durations
               WHERE policy_id = @new AND duration_minutes NOT IN (30, 60, 90, 120, 150, 180)
           )
            THROW 57005, 'Replay: las duraciones reparadas divergen.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_resources WHERE policy_id = @new) <> 8
           OR EXISTS (
               SELECT 1 FROM dbo.reservation_policy_resources
               WHERE policy_id = @new AND resource_id NOT IN (1, 2, 3, 4, 5, 6, 7, 8)
           )
            THROW 57005, 'Replay: los recursos permitidos reparados divergen.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_group_resources WHERE policy_id = @new) <> 3
           OR EXISTS (
               SELECT 1 FROM dbo.reservation_policy_group_resources
               WHERE policy_id = @new AND resource_id NOT IN (1, 2, 7)
           )
            THROW 57005, 'Replay: los recursos grupales reparados divergen.', 1;
    END;

    IF @new IS NULL
    BEGIN
        SELECT TOP (1) @current = id
        FROM dbo.reservation_policies WITH (UPDLOCK, HOLDLOCK)
        WHERE is_published = 1
          AND effective_from <= @now
          AND (effective_to IS NULL OR effective_to > @now)
        ORDER BY effective_from DESC, id DESC;

        IF @current IS NULL
            THROW 57002, 'No existe una politica vigente publicada.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policies WITH (HOLDLOCK)) <> 1
           OR NOT EXISTS (
               SELECT 1
               FROM dbo.reservation_policies
               WHERE id = @current
                 AND reservable_window_days = 7
                 AND request_frequency_days = 7
                 AND confirmation_deadline_minutes = 60
                 AND minimum_participants = 10
                 AND opening_minute = 480
                 AND closing_minute = 1320
                 AND slot_interval_minutes = 15
                 AND effective_from = CONVERT(DATETIME2(0), '19000101', 112)
                 AND effective_to IS NULL
                 AND created_by_user_id IS NULL
                 AND idempotency_key IS NULL
                 AND idempotency_payload_hash IS NULL
                 AND is_published = 1
           )
            THROW 57003, 'La politica vigente no tiene la huella unica del bootstrap; no se modifico.', 1;

        IF NOT EXISTS (
            SELECT 1 FROM dbo.reservation_policy_scope_migrations WHERE policy_id = @current
        )
            THROW 57003, 'La politica bootstrap no tiene la marca de alcance esperada; no se modifico.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_durations WHERE policy_id = @current) <> 6
           OR EXISTS (
               SELECT 1 FROM dbo.reservation_policy_durations
               WHERE policy_id = @current AND duration_minutes NOT IN (30, 60, 90, 120, 150, 180)
           )
            THROW 57003, 'Las duraciones de la politica bootstrap divergen; no se modifico.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_resources WHERE policy_id = @current) <> 8
           OR EXISTS (
               SELECT 1 FROM dbo.reservation_policy_resources
               WHERE policy_id = @current AND resource_id NOT IN (1, 2, 3, 4, 5, 6, 7, 8)
           )
            THROW 57003, 'Los recursos permitidos del bootstrap divergen; no se modifico.', 1;

        IF (SELECT COUNT(*) FROM dbo.reservation_policy_group_resources WHERE policy_id = @current) = 3
           AND NOT EXISTS (
               SELECT 1 FROM dbo.reservation_policy_group_resources
               WHERE policy_id = @current AND resource_id NOT IN (1, 2, 7)
           )
        BEGIN
            COMMIT TRANSACTION;
            SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
            SELECT
                @current AS repaired_policy_id,
                CAST(NULL AS NVARCHAR(100)) AS idempotency_key,
                CAST(NULL AS CHAR(64)) AS idempotency_payload_hash,
                CONVERT(bit, 1) AS replayed,
                CONVERT(bit, 1) AS already_correct;
            RETURN;
        END;

        IF EXISTS (
            SELECT 1 FROM dbo.reservation_policy_group_resources WHERE policy_id = @current
        )
            THROW 57003, 'La politica vigente tiene alcance grupal parcial o administrado; no se modifico.', 1;

        UPDATE dbo.reservation_policies
        SET effective_to = @now
        WHERE id = @current AND effective_to IS NULL;

        IF @@ROWCOUNT <> 1
            THROW 57003, 'La politica bootstrap cambio concurrentemente; no se modifico.', 1;

        INSERT INTO dbo.reservation_policies (
            reservable_window_days, request_frequency_days,
            confirmation_deadline_minutes, minimum_participants,
            opening_minute, closing_minute, slot_interval_minutes,
            effective_from, created_by_user_id, idempotency_key,
            idempotency_payload_hash, is_published
        )
        SELECT
            reservable_window_days, request_frequency_days,
            confirmation_deadline_minutes, minimum_participants,
            opening_minute, closing_minute, slot_interval_minutes,
            @now, NULL, @key, @payload_hash, 0
        FROM dbo.reservation_policies
        WHERE id = @current;

        SET @new = SCOPE_IDENTITY();

        INSERT INTO dbo.reservation_policy_durations (policy_id, duration_minutes)
        SELECT @new, duration_minutes
        FROM dbo.reservation_policy_durations
        WHERE policy_id = @current;

        INSERT INTO dbo.reservation_policy_resources (policy_id, resource_id)
        SELECT @new, resource_id
        FROM dbo.reservation_policy_resources
        WHERE policy_id = @current;

        INSERT INTO dbo.reservation_policy_group_resources (policy_id, resource_id)
        SELECT @new, resource_id
        FROM (VALUES (1), (2), (7)) expected(resource_id);

        UPDATE dbo.reservation_policies
        SET is_published = 1
        WHERE id = @new AND is_published = 0;

        IF @@ROWCOUNT <> 1
            THROW 57004, 'Postcheck: la politica reparada no se pudo publicar.', 1;

        IF NOT EXISTS (
            SELECT 1
            FROM dbo.reservation_policies
            WHERE id = @new
              AND effective_to IS NULL
              AND is_published = 1
              AND idempotency_key = @key
              AND idempotency_payload_hash = @payload_hash
        )
            THROW 57004, 'Postcheck: la politica reparada no quedo vigente y trazable.', 1;
    END;

    COMMIT TRANSACTION;
    SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

    SELECT
        @new AS repaired_policy_id,
        @key AS idempotency_key,
        @payload_hash AS idempotency_payload_hash,
        @replayed AS replayed,
        CONVERT(bit, 0) AS already_correct;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
    THROW;
END CATCH;
GO

SELECT
    CONVERT(bit, CASE WHEN EXISTS (
        SELECT 1
        FROM dbo.reservation_policies p
        WHERE p.is_published = 1
          AND p.effective_to IS NULL
          AND (
              (p.idempotency_key = N'repair-bootstrap-group-policy-v1'
               AND p.idempotency_payload_hash = '8a33d6dbedf56c20dbb857c3579ed986ccfce686afb60278dff06190aafe6ed3')
              OR (
                  p.idempotency_key IS NULL
                  AND p.effective_from = CONVERT(DATETIME2(0), '19000101', 112)
                  AND p.created_by_user_id IS NULL
              )
          )
          AND (SELECT COUNT(*) FROM dbo.reservation_policy_group_resources g
               WHERE g.policy_id = p.id AND g.resource_id IN (1, 2, 7)) = 3
          AND NOT EXISTS (
              SELECT 1 FROM dbo.reservation_policy_group_resources g
              WHERE g.policy_id = p.id AND g.resource_id NOT IN (1, 2, 7)
          )
    ) THEN 1 ELSE 0 END) AS bootstrap_group_policy_repaired;
GO
