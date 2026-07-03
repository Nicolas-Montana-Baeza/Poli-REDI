-- ============================================================
-- POLI-REDI - DATOS INICIALES
-- Azure SQL Database / SQL Server T-SQL
-- ============================================================

-- VENUES
SET IDENTITY_INSERT dbo.venues ON;
INSERT INTO dbo.venues (id, name, address_line, commune, city, region, country, latitude, longitude, is_active)
VALUES
(1, 'Campus Principal', 'Av. Santa Isabel 1186', 'Santiago', 'Santiago', 'Region Metropolitana', 'Chile', -33.4551000, -70.6415000, 1),
(2, 'Campo Deportivo', 'Av. Ejemplo 1234', 'Santiago', 'Santiago', 'Region Metropolitana', 'Chile', -33.4372000, -70.6506000, 1);
SET IDENTITY_INSERT dbo.venues OFF;
GO

-- USERS
SET IDENTITY_INSERT dbo.users ON;
INSERT INTO dbo.users (id, email, full_name, is_admin, is_blocked, entra_oid, tenant_id)
VALUES
(1, 'admin@universidad.cl', 'Administrador General', 1, 0, NULL, NULL),
(2, 'nicolas@universidad.cl', 'Nicolas Montana', 0, 0, NULL, NULL),
(3, 'maria@universidad.cl', 'Maria Gonzalez', 0, 0, NULL, NULL),
(4, 'juan@universidad.cl', 'Juan Perez', 0, 0, NULL, NULL),
(5, 'camila@universidad.cl', 'Camila Soto', 0, 0, NULL, NULL),
(6, 'pedro@universidad.cl', 'Pedro Ramirez', 0, 0, NULL, NULL),
(7, 'valentina@universidad.cl', 'Valentina Fuentes', 0, 0, NULL, NULL),
(8, 'sebastian@universidad.cl', 'Sebastian Morales', 0, 0, NULL, NULL),
(9, 'fernanda@universidad.cl', 'Fernanda Rojas', 0, 0, NULL, NULL),
(10, 'diego@universidad.cl', 'Diego Herrera', 0, 0, NULL, NULL),
(11, 'sofia@universidad.cl', 'Sofia Castillo', 0, 0, NULL, NULL),
(12, 'bloqueado@universidad.cl', 'Usuario Bloqueado', 0, 1, NULL, NULL);
SET IDENTITY_INSERT dbo.users OFF;
GO

-- RESOURCES
SET IDENTITY_INSERT dbo.resources ON;
INSERT INTO dbo.resources (id, venue_id, name, type, reservation_mode, capacity, is_active)
VALUES
(1, 2, 'Cancha de Futbol 1', 'Cancha', 'RESERVABLE', 22, 1),
(2, 2, 'Cancha de Basquetbol', 'Cancha', 'RESERVABLE', 10, 1),
(3, 1, 'Piscina', 'Piscina', 'RESERVABLE', 20, 1),
(4, 1, 'Gimnasio', 'Gimnasio', 'RESERVABLE', 40, 1),
(5, 1, 'Sala Multiuso', 'Sala', 'ADMIN_ONLY', 25, 1),
(6, 1, 'Muro Informativo Deportivo', 'Informativo', 'INFORMATIVE', NULL, 1);
SET IDENTITY_INSERT dbo.resources OFF;
GO

-- ACTIVITIES
SET IDENTITY_INSERT dbo.activities ON;
INSERT INTO dbo.activities (id, name, description, is_active)
VALUES
(1, 'Futbol', 'Partidos o entrenamientos de futbol.', 1),
(2, 'Basquetbol', 'Partidos o entrenamientos de basquetbol.', 1),
(3, 'Natacion', 'Uso deportivo de piscina.', 1),
(4, 'Entrenamiento Libre', 'Uso general de gimnasio.', 1),
(5, 'Yoga', 'Actividad grupal guiada.', 1),
(6, 'Campeonato', 'Competencia institucional.', 1);
SET IDENTITY_INSERT dbo.activities OFF;
GO

-- RESERVATIONS
SET IDENTITY_INSERT dbo.reservations ON;
INSERT INTO dbo.reservations (id, user_id, resource_id, activity_id, start_time, duration_minutes, status, cancellation_reason)
VALUES
(1, 2, 1, 1, '2026-04-30T10:00:00', 60, 'CONFIRMED', NULL),
(2, 3, 2, 2, '2026-04-30T11:30:00', 90, 'CONFIRMED', NULL),
(3, 4, 3, 3, '2026-04-30T13:30:00', 60, 'CONFIRMED', NULL),
(4, 5, 4, 4, '2026-04-30T15:00:00', 120, 'CONFIRMED', NULL),
(5, 1, 5, 5, '2026-04-30T18:00:00', 60, 'PENDING', NULL),
(6, 7, 1, 1, '2026-05-01T09:00:00', 60, 'CONFIRMED', NULL),
(7, 8, 2, 2, '2026-05-01T11:00:00', 60, 'CONFIRMED', NULL),
(8, 9, 3, 3, '2026-05-01T14:00:00', 60, 'CONFIRMED', NULL);
SET IDENTITY_INSERT dbo.reservations OFF;
GO

-- PARTICIPANTS
SET IDENTITY_INSERT dbo.participants ON;
INSERT INTO dbo.participants (id, reservation_id, user_id, status, confirmed_at)
VALUES
(1, 1, 3, 'CONFIRMED', '2026-04-29T09:00:00'),
(2, 1, 4, 'CONFIRMED', '2026-04-29T09:05:00'),
(3, 2, 2, 'CONFIRMED', '2026-04-29T10:00:00'),
(4, 4, 6, 'PENDING', NULL),
(5, 6, 8, 'CONFIRMED', '2026-04-30T12:00:00'),
(6, 7, 9, 'CONFIRMED', '2026-04-30T12:30:00');
SET IDENTITY_INSERT dbo.participants OFF;
GO

-- AVAILABILITY BLOCKS
SET IDENTITY_INSERT dbo.availability_blocks ON;
INSERT INTO dbo.availability_blocks (id, resource_id, created_by_user_id, block_type, reason, start_time, end_time, is_active)
VALUES
(1, 1, 1, 'MAINTENANCE', 'Mantencion programada de cancha.', '2026-05-02T08:00:00', '2026-05-02T12:00:00', 1),
(2, 3, 1, 'CLOSED', 'Limpieza profunda de piscina.', '2026-05-03T14:00:00', '2026-05-03T18:00:00', 1);
SET IDENTITY_INSERT dbo.availability_blocks OFF;
GO

-- SCHEDULED ACTIVITIES
SET IDENTITY_INSERT dbo.scheduled_activities ON;
INSERT INTO dbo.scheduled_activities (id, resource_id, activity_id, created_by_user_id, title, description, activity_type, start_time, end_time, recurrence_rule, is_active)
VALUES
(1, 4, 4, 1, 'Entrenamiento funcional', 'Clase guiada para estudiantes.', 'TRAINING', '2026-05-02T16:00:00', '2026-05-02T17:00:00', NULL, 1),
(2, 5, 5, 1, 'Taller de yoga', 'Actividad institucional en sala multiuso.', 'WORKSHOP', '2026-05-03T10:00:00', '2026-05-03T11:00:00', NULL, 1),
(3, 2, 6, 1, 'Campeonato interno', 'Evento deportivo universitario.', 'CHAMPIONSHIP', '2026-05-04T09:00:00', '2026-05-04T13:00:00', NULL, 1);
SET IDENTITY_INSERT dbo.scheduled_activities OFF;
GO

-- VIOLATIONS
SET IDENTITY_INSERT dbo.violations ON;
INSERT INTO dbo.violations (id, user_id, reservation_id, created_by_user_id, violation_type, description)
VALUES
(1, 4, 3, 1, 'NO_SHOW', 'No asistio a la reserva confirmada.'),
(2, 12, NULL, 1, 'MISUSE', 'Intento reservar estando bloqueado.');
SET IDENTITY_INSERT dbo.violations OFF;
GO

-- NOTIFICATIONS
SET IDENTITY_INSERT dbo.notifications ON;
INSERT INTO dbo.notifications (id, user_id, reservation_id, title, message, type, is_read)
VALUES
(101, 2, 1, 'Reserva confirmada', 'Tu reserva ha sido confirmada.', 'RESERVATION_CONFIRMED', 0),
(102, 4, 3, 'Infraccion registrada', 'Se registro una infraccion asociada a tu reserva.', 'SYSTEM', 0),
(103, 6, 5, 'Reserva pendiente', 'Tu reserva esta pendiente de confirmacion.', 'RESERVATION_CREATED', 0);
SET IDENTITY_INSERT dbo.notifications OFF;
GO

-- AUDIT LOGS
SET IDENTITY_INSERT dbo.audit_logs ON;
INSERT INTO dbo.audit_logs (id, user_id, action, entity_type, entity_id, description)
VALUES
(101, 1, 'SEED_CREATED', 'database', NULL, 'Datos iniciales cargados para ambiente de desarrollo.');
SET IDENTITY_INSERT dbo.audit_logs OFF;
GO


