-- ============================================================
-- POLI-REDI - SEED TEMPORAL PARA PRUEBAS DE HOY
-- Ejecutar despues de database/schema.sql y database/seed.sql.
-- Fecha de prueba: 2026-07-14
-- ============================================================

DELETE FROM dbo.notifications WHERE id IN (101, 102, 103);
DELETE FROM dbo.violations WHERE id IN (1, 2);
DELETE FROM dbo.participants WHERE id IN (1, 2, 3, 4, 5, 6);
DELETE FROM dbo.availability_blocks WHERE id IN (1, 2);
DELETE FROM dbo.scheduled_activities WHERE id IN (1, 2, 3);
DELETE FROM dbo.reservations WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8);
GO

SET IDENTITY_INSERT dbo.reservations ON;
DECLARE @initial_policy_id INT = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from ASC, id ASC);
INSERT INTO dbo.reservations (id, policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status, cancellation_reason, created_at)
VALUES
(1, @initial_policy_id, 2, 1, 1, '2026-07-14T08:30:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(2, @initial_policy_id, 3, 2, 2, '2026-07-14T09:45:00', 90, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(3, @initial_policy_id, 4, 3, 6, '2026-07-14T11:30:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(4, @initial_policy_id, 5, 4, 5, '2026-07-14T13:00:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(5, @initial_policy_id, 6, 5, NULL, '2026-07-14T15:00:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(6, @initial_policy_id, 7, 8, NULL, '2026-07-14T16:30:00', 90, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(7, @initial_policy_id, 8, 1, 1, '2026-07-14T18:30:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00'),
(8, @initial_policy_id, 9, 7, 2, '2026-07-14T20:00:00', 60, 'CONFIRMED', NULL, '2026-07-13T12:00:00');
SET IDENTITY_INSERT dbo.reservations OFF;
GO

SET IDENTITY_INSERT dbo.participants ON;
INSERT INTO dbo.participants (id, reservation_id, user_id, status, confirmed_at)
VALUES
(1, 1, 3, 'CONFIRMED', '2026-07-14T08:15:00'),
(2, 1, 4, 'CONFIRMED', '2026-07-14T08:20:00'),
(3, 2, 2, 'CONFIRMED', '2026-07-14T09:30:00'),
(4, 4, 6, 'PENDING', NULL),
(5, 6, 8, 'CONFIRMED', '2026-07-14T16:00:00'),
(6, 7, 9, 'CONFIRMED', '2026-07-14T18:00:00');
SET IDENTITY_INSERT dbo.participants OFF;
GO

SET IDENTITY_INSERT dbo.availability_blocks ON;
INSERT INTO dbo.availability_blocks (id, resource_id, created_by_user_id, block_type, reason, start_time, end_time, is_active)
VALUES
(1, 1, 1, 'MAINTENANCE', 'Mantencion programada de cancha.', '2026-07-14T12:00:00', '2026-07-14T13:00:00', 1),
(2, 3, 1, 'CLOSED', 'Limpieza profunda de muro.', '2026-07-14T16:00:00', '2026-07-14T17:00:00', 1);
SET IDENTITY_INSERT dbo.availability_blocks OFF;
GO

SET IDENTITY_INSERT dbo.scheduled_activities ON;
INSERT INTO dbo.scheduled_activities (id, resource_id, activity_id, created_by_user_id, title, description, activity_type, start_time, end_time, recurrence_rule, is_active)
VALUES
(1, 4, 4, 1, 'Entrenamiento funcional', 'Clase guiada para estudiantes.', 'TRAINING', '2026-07-14T17:00:00', '2026-07-14T18:00:00', NULL, 1),
(2, 6, 5, 1, 'Taller de yoga', 'Actividad institucional en sala multiuso.', 'WORKSHOP', '2026-07-14T10:00:00', '2026-07-14T11:00:00', NULL, 1),
(3, 2, 6, 1, 'Campeonato interno', 'Evento deportivo universitario.', 'CHAMPIONSHIP', '2026-07-14T19:30:00', '2026-07-14T21:00:00', NULL, 1);
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
