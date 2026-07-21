-- ============================================================
-- POLI-REDI - SEED TEMPORAL PARA PRUEBAS DE HOY
-- Ejecutar despues de database/schema.sql y database/seed.sql.
-- La fecha se calcula dinamicamente en America/Santiago.
-- ============================================================

IF OBJECT_ID('tempdb..#seed_today_context', 'U') IS NOT NULL DROP TABLE #seed_today_context;
CREATE TABLE #seed_today_context (business_date DATE NOT NULL);
INSERT INTO #seed_today_context (business_date)
VALUES (CONVERT(DATE, SYSUTCDATETIME() AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'));
GO

DELETE FROM dbo.notifications WHERE id IN (101, 102, 103);
DELETE FROM dbo.violations WHERE id IN (1, 2);
DELETE FROM dbo.participants WHERE id IN (1, 2, 3, 4, 5, 6);
DELETE FROM dbo.availability_blocks WHERE id IN (1, 2);
DELETE FROM dbo.scheduled_activities WHERE id IN (1, 2, 3);
DELETE FROM dbo.reservations WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8);
GO

SET IDENTITY_INSERT dbo.reservations ON;
DECLARE @initial_policy_id INT = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from ASC, id ASC);
DECLARE @business_date DATE = (SELECT business_date FROM #seed_today_context);
INSERT INTO dbo.reservations (id, policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status, cancellation_reason)
SELECT
    seed.id,
    @initial_policy_id,
    seed.user_id,
    seed.resource_id,
    seed.activity_id,
    DATEADD(MINUTE, seed.start_minute, CAST(@business_date AS DATETIME2(0))),
    seed.duration_minutes,
    'CONFIRMED',
    NULL
FROM (VALUES
    (1, 2, 1, 1,  510, 60),
    (2, 3, 2, 2,  585, 90),
    (3, 4, 3, 6,  690, 60),
    (4, 5, 4, 5,  780, 60),
    (5, 6, 5, NULL, 900, 60),
    (6, 7, 8, NULL, 990, 90),
    (7, 8, 1, 1, 1110, 60),
    (8, 9, 7, 2, 1200, 60)
) seed(id, user_id, resource_id, activity_id, start_minute, duration_minutes);
SET IDENTITY_INSERT dbo.reservations OFF;
GO

SET IDENTITY_INSERT dbo.participants ON;
INSERT INTO dbo.participants (id, reservation_id, user_id, status, confirmed_at)
SELECT
    seed.id,
    seed.reservation_id,
    seed.user_id,
    seed.status,
    CASE
        WHEN seed.confirmed_minute IS NULL THEN NULL
        ELSE CONVERT(
            DATETIME2(0),
            DATEADD(MINUTE, seed.confirmed_minute, CAST(context.business_date AS DATETIME2(0)))
                AT TIME ZONE 'Pacific SA Standard Time'
                AT TIME ZONE 'UTC'
        )
    END
FROM (VALUES
    (1, 1, 3, 'CONFIRMED',  495),
    (2, 1, 4, 'CONFIRMED',  500),
    (3, 2, 2, 'CONFIRMED',  570),
    (4, 4, 6, 'PENDING',   NULL),
    (5, 6, 8, 'CONFIRMED',  960),
    (6, 7, 9, 'CONFIRMED', 1080)
) seed(id, reservation_id, user_id, status, confirmed_minute)
CROSS JOIN #seed_today_context context;
SET IDENTITY_INSERT dbo.participants OFF;
GO

SET IDENTITY_INSERT dbo.availability_blocks ON;
INSERT INTO dbo.availability_blocks (id, resource_id, created_by_user_id, block_type, reason, start_time, end_time, is_active)
SELECT
    seed.id,
    seed.resource_id,
    1,
    seed.block_type,
    seed.reason,
    DATEADD(MINUTE, seed.start_minute, CAST(context.business_date AS DATETIME2(0))),
    DATEADD(MINUTE, seed.end_minute, CAST(context.business_date AS DATETIME2(0))),
    1
FROM (VALUES
    (1, 1, 'MAINTENANCE', 'Mantencion programada de cancha.', 720, 780),
    (2, 3, 'CLOSED', 'Limpieza profunda de muro.', 960, 1020)
) seed(id, resource_id, block_type, reason, start_minute, end_minute)
CROSS JOIN #seed_today_context context;
SET IDENTITY_INSERT dbo.availability_blocks OFF;
GO

SET IDENTITY_INSERT dbo.scheduled_activities ON;
INSERT INTO dbo.scheduled_activities (id, resource_id, activity_id, created_by_user_id, title, description, activity_type, start_time, end_time, recurrence_rule, is_active)
SELECT
    seed.id,
    seed.resource_id,
    seed.activity_id,
    1,
    seed.title,
    seed.description,
    seed.activity_type,
    DATEADD(MINUTE, seed.start_minute, CAST(context.business_date AS DATETIME2(0))),
    DATEADD(MINUTE, seed.end_minute, CAST(context.business_date AS DATETIME2(0))),
    NULL,
    1
FROM (VALUES
    (1, 4, 4, 'Entrenamiento funcional', 'Clase guiada para estudiantes.', 'TRAINING', 1020, 1080),
    (2, 6, 5, 'Taller de yoga', 'Actividad institucional en sala multiuso.', 'WORKSHOP', 600, 660),
    (3, 2, 6, 'Campeonato interno', 'Evento deportivo universitario.', 'CHAMPIONSHIP', 1170, 1260)
) seed(id, resource_id, activity_id, title, description, activity_type, start_minute, end_minute)
CROSS JOIN #seed_today_context context;
SET IDENTITY_INSERT dbo.scheduled_activities OFF;
GO

SET IDENTITY_INSERT dbo.notifications ON;
INSERT INTO dbo.notifications (id, user_id, reservation_id, title, message, type, is_read)
VALUES
(101, 2, 1, 'Reserva confirmada', 'Tu reserva de hoy ha sido confirmada.', 'RESERVATION_CONFIRMED', 0),
(102, 5, 4, 'Reserva confirmada', 'Tu reserva de hoy ha sido confirmada.', 'RESERVATION_CONFIRMED', 0),
(103, 7, 6, 'Reserva confirmada', 'Tu reserva de hoy ha sido confirmada.', 'RESERVATION_CONFIRMED', 0);
SET IDENTITY_INSERT dbo.notifications OFF;
GO

DROP TABLE #seed_today_context;
GO
