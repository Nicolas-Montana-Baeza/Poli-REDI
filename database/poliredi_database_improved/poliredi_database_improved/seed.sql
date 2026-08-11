-- ============================================================
-- POLI-REDI - DATOS INICIALES
-- Script idempotente para ambientes locales y de desarrollo.
-- ============================================================

SET NOCOUNT ON;
SET XACT_ABORT ON;

IF TRY_CONVERT(bit, SESSION_CONTEXT(N'allow_development_seed')) <> 1
    THROW 59991, 'Seed bloqueado. Habilite allow_development_seed=1 en esta sesion.', 1;
GO


-- Asegurar el usuario administrador base.
IF NOT EXISTS (SELECT 1 FROM dbo.users WHERE email = N'admin@polirediucen.onmicrosoft.com')
BEGIN
    SET IDENTITY_INSERT dbo.users ON;
    INSERT INTO dbo.users (id, email, full_name, rut, is_admin, is_blocked, entra_oid, tenant_id)
    VALUES
    (1, N'admin@polirediucen.onmicrosoft.com', N'Administrador', NULL, 1, 0, N'2e02ef4b-97d8-4359-acce-3431018f8745', NULL);
    SET IDENTITY_INSERT dbo.users OFF;
END;
GO

UPDATE dbo.users
SET is_admin = 1,
    updated_at = SYSUTCDATETIME()
WHERE email = N'admin@polirediucen.onmicrosoft.com'
  AND is_admin = 0;
GO

-- VENUES
IF NOT EXISTS (SELECT 1 FROM dbo.venues WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.venues ON;
    INSERT INTO dbo.venues (id, name, address_line, commune, city, region, country, latitude, longitude, is_active)
    VALUES
    (1, N'Campus Principal', N'Av. Santa Isabel 1186', N'Santiago', N'Santiago', N'Region Metropolitana', N'Chile', -33.4551000, -70.6415000, 1),
    (2, N'Campo Deportivo', N'Av. Ejemplo 1234', N'Santiago', N'Santiago', N'Region Metropolitana', N'Chile', -33.4372000, -70.6506000, 1);
    SET IDENTITY_INSERT dbo.venues OFF;
END;
GO

-- USERS
IF NOT EXISTS (SELECT 1 FROM dbo.users WHERE id = 2)
BEGIN
    SET IDENTITY_INSERT dbo.users ON;
    INSERT INTO dbo.users (id, email, full_name, rut, is_admin, is_blocked, entra_oid, tenant_id)
    VALUES
    (2, N'nicolas@polirediucen.onmicrosoft.com', N'Nicolás Montaña', N'10000001-6', 0, 0, N'1cc13d5f-8baa-427c-9f04-a09c4e52121a', NULL),
    (3, N'maria@polirediucen.onmicrosoft.com', N'María González', N'10000002-4', 0, 0, N'43831395-6628-4c1d-b3e8-289c54076dfd', NULL),
    (4, N'juan@polirediucen.onmicrosoft.com', N'Juan Pérez', N'10000003-2', 0, 0, N'0fe3819a-c6f7-420f-85c4-9ccc578e72c0', NULL),
    (5, N'camila@polirediucen.onmicrosoft.com', N'Camila Soto', N'10000004-0', 0, 0, N'5bb1fded-f7e2-4dd8-b591-58756b5bb0b4', NULL),
    (6, N'pedro@polirediucen.onmicrosoft.com', N'Pedro Ramírez', N'10000005-9', 0, 0, N'22110319-7b8b-4c4c-ac18-0d7897a3200f', NULL),
    (7, N'valentina@polirediucen.onmicrosoft.com', N'Valentina Fuentes', N'10000006-7', 0, 0, N'1078591b-b341-4762-9f0d-41a5ae9a3cfc', NULL),
    (8, N'sebastian@polirediucen.onmicrosoft.com', N'Sebastián Morales', N'10000007-5', 0, 0, N'4baf08a6-75a7-4ecc-8e91-fc537c019106', NULL),
    (9, N'fernanda@polirediucen.onmicrosoft.com', N'Fernanda Rojas', N'10000008-3', 0, 0, N'a633168d-5d10-4dd3-875d-fd58f224c67e', NULL),
    (10, N'diego@polirediucen.onmicrosoft.com', N'Diego Herrera', N'10000009-1', 0, 0, N'175b7686-e1ea-4b9a-8999-77bc1be7f800', NULL),
    (11, N'sofia@polirediucen.onmicrosoft.com', N'Sofía Castillo', N'10000010-5', 0, 0, N'a0c7a873-6e4b-4934-8780-1c40f26f7b50', NULL),
    (12, N'bloqueado@polirediucen.onmicrosoft.com', N'Usuario Bloqueado', N'10000011-3', 0, 1, N'c0e0c85d-a7f9-4e54-a213-c452d4546ba0', NULL),
    (13, N'ni.cco896_gmail.com#EXT#@polirediucen.onmicrosoft.com', N'Nicolás Montaña', N'10000012-1', 0, 0, N'3da380aa-1b93-4418-adeb-a1ae8d862776', NULL);
    SET IDENTITY_INSERT dbo.users OFF;
END;
GO

-- RESOURCES
IF NOT EXISTS (SELECT 1 FROM dbo.resources WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.resources ON;
    INSERT INTO dbo.resources (id, venue_id, name, type, reservation_mode, image_url, capacity, is_active)
    VALUES
    (1, 2, N'Cancha 1, Centro Deportivo', N'Cancha', N'RESERVABLE', N'https://images.unsplash.com/photo-1574629810360-7efbbe195018?auto=format&fit=crop&w=900&q=80', 22, 1),
    (2, 2, N'Cancha 2, Centro Deportivo', N'Cancha', N'RESERVABLE', N'https://images.unsplash.com/photo-1574629810360-7efbbe195018?auto=format&fit=crop&w=900&q=80', 22, 1),
    (3, 2, N'Muro Escalada, Centro Deportivo', N'Muro Escalada', N'RESERVABLE', N'https://images.unsplash.com/photo-1522163182402-834f871fd851?auto=format&fit=crop&w=900&q=80', 20, 1),
    (4, 2, N'Sala Spinning, Centro Deportivo', N'Sala', N'RESERVABLE', N'https://images.unsplash.com/photo-1518611012118-696072aa579a?auto=format&fit=crop&w=900&q=80', 25, 1),
    (5, 2, N'Piscina, Centro Deportivo', N'Piscina', N'OPEN_USE', N'https://images.unsplash.com/photo-1575429198097-0414ec08e8cd?auto=format&fit=crop&w=900&q=80', 20, 1),
    (6, 2, N'Sala Multiuso, Centro Deportivo', N'Sala', N'ADMIN_ONLY', N'https://images.unsplash.com/photo-1517457373958-b7bdd4587205?auto=format&fit=crop&w=900&q=80', 25, 1),
    (7, 2, N'Cancha 3, Centro Deportivo', N'Cancha', N'RESERVABLE', N'https://images.unsplash.com/photo-1577223625816-7546f13df25d?auto=format&fit=crop&w=900&q=80', 22, 1),
    (8, 2, N'Gimnasio, Centro Deportivo', N'Gimnasio', N'OPEN_USE', N'https://images.unsplash.com/photo-1534438327276-14e5300c3a48?auto=format&fit=crop&w=900&q=80', 40, 1);
    SET IDENTITY_INSERT dbo.resources OFF;
END;
GO

-- Bootstrap técnico único de la política creada por schema.sql.
BEGIN TRY
    BEGIN TRANSACTION;
    EXEC sys.sp_set_session_context @key=N'legacy_policy_scope_bootstrap', @value=1;

    DECLARE @bootstrap_policy_id INT = (
        SELECT TOP (1) p.id
        FROM dbo.reservation_policies AS p WITH (UPDLOCK, HOLDLOCK)
        WHERE p.idempotency_key IS NULL
          AND p.id = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from, id)
          AND NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_scope_migrations AS m WHERE m.policy_id = p.id)
    );

    IF @bootstrap_policy_id IS NOT NULL
    BEGIN
        INSERT INTO dbo.reservation_policy_resources (policy_id, resource_id)
        SELECT @bootstrap_policy_id, r.id
        FROM dbo.resources AS r
        WHERE r.is_active = 1
          AND r.reservation_mode <> N'INFORMATIVE'
          AND NOT EXISTS (
              SELECT 1
              FROM dbo.reservation_policy_resources AS existing
              WHERE existing.policy_id = @bootstrap_policy_id
                AND existing.resource_id = r.id
          );

        INSERT INTO dbo.reservation_policy_group_resources (policy_id, resource_id)
        SELECT @bootstrap_policy_id, r.id
        FROM dbo.resources AS r
        WHERE r.id IN (1, 2, 7)
          AND r.is_active = 1
          AND r.reservation_mode <> N'OPEN_USE'
          AND r.capacity >= (
              SELECT p.minimum_participants
              FROM dbo.reservation_policies AS p
              WHERE p.id = @bootstrap_policy_id
          )
          AND NOT EXISTS (
              SELECT 1
              FROM dbo.reservation_policy_group_resources AS existing
              WHERE existing.policy_id = @bootstrap_policy_id
                AND existing.resource_id = r.id
          );

        IF NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_scope_migrations WHERE policy_id = @bootstrap_policy_id)
        BEGIN
            INSERT INTO dbo.reservation_policy_scope_migrations (policy_id)
            VALUES (@bootstrap_policy_id);
        END;
    END;

    EXEC sys.sp_set_session_context @key=N'legacy_policy_scope_bootstrap', @value=NULL;
    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    EXEC sys.sp_set_session_context @key=N'legacy_policy_scope_bootstrap', @value=NULL;
    THROW;
END CATCH;
GO

-- ACTIVITIES
IF NOT EXISTS (SELECT 1 FROM dbo.activities WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.activities ON;
    INSERT INTO dbo.activities (id, name, description, is_active)
    VALUES
    (1, N'Futbol', N'Partidos o entrenamientos de futbol.', 1),
    (2, N'Basquetbol', N'Partidos o entrenamientos de basquetbol.', 1),
    (3, N'Natacion', N'Uso deportivo de piscina.', 1),
    (4, N'Entrenamiento Libre', N'Uso general de gimnasio.', 1),
    (5, N'Yoga', N'Actividad grupal guiada.', 1),
    (6, N'Campeonato', N'Competencia institucional.', 1);
    SET IDENTITY_INSERT dbo.activities OFF;
END;
GO

-- RESERVATIONS
DECLARE @initial_policy_id INT = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from ASC, id ASC);
IF NOT EXISTS (SELECT 1 FROM dbo.reservations WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.reservations ON;
    INSERT INTO dbo.reservations (id, policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status, group_capacity_snapshot, target_participants, cancellation_reason, created_at)
    VALUES
    (1, @initial_policy_id, 2, 1, 1, '2026-04-30T10:00:00', 60, 'CONFIRMED', 22, 10, NULL, '2026-04-29T12:00:00'),
    (2, @initial_policy_id, 3, 2, 2, '2026-04-30T11:30:00', 90, 'CONFIRMED', 22, 10, NULL, '2026-04-29T12:00:00'),
    (3, @initial_policy_id, 4, 3, 3, '2026-04-30T13:30:00', 60, 'CONFIRMED', NULL, NULL, NULL, '2026-04-29T12:00:00'),
    (4, @initial_policy_id, 5, 4, 4, '2026-04-30T15:00:00', 120, 'CONFIRMED', NULL, NULL, NULL, '2026-04-29T12:00:00'),
    (5, @initial_policy_id, 1, 5, 5, '2026-04-30T18:00:00', 60, 'PENDING', NULL, NULL, NULL, '2026-04-29T12:00:00'),
    (6, @initial_policy_id, 7, 1, 1, '2026-05-01T09:00:00', 60, 'CONFIRMED', 22, 10, NULL, '2026-04-30T12:00:00'),
    (7, @initial_policy_id, 8, 2, 2, '2026-05-01T11:00:00', 60, 'CONFIRMED', 22, 10, NULL, '2026-04-30T12:00:00'),
    (8, @initial_policy_id, 9, 3, 3, '2026-05-01T14:00:00', 60, 'CONFIRMED', NULL, NULL, NULL, '2026-04-30T12:00:00');
    SET IDENTITY_INSERT dbo.reservations OFF;
END;
GO

-- PARTICIPANTS
IF NOT EXISTS (SELECT 1 FROM dbo.participants WHERE id = 1)
BEGIN
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
END;
GO

-- AVAILABILITY BLOCKS
IF NOT EXISTS (SELECT 1 FROM dbo.availability_blocks WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.availability_blocks ON;
    INSERT INTO dbo.availability_blocks (id, resource_id, created_by_user_id, block_type, reason, start_time, end_time, is_active)
    VALUES
    (1, 1, 1, 'MAINTENANCE', N'Mantencion programada de cancha.', '2026-05-02T08:00:00', '2026-05-02T12:00:00', 1),
    (2, 3, 1, 'CLOSED', N'Limpieza profunda de piscina.', '2026-05-03T14:00:00', '2026-05-03T18:00:00', 1);
    SET IDENTITY_INSERT dbo.availability_blocks OFF;
END;
GO

-- SCHEDULED ACTIVITIES
IF NOT EXISTS (SELECT 1 FROM dbo.scheduled_activities WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.scheduled_activities ON;
    INSERT INTO dbo.scheduled_activities (id, resource_id, activity_id, created_by_user_id, title, description, activity_type, start_time, end_time, recurrence_rule, is_active)
    VALUES
    (1, 4, 4, 1, N'Entrenamiento funcional', N'Clase guiada para estudiantes.', 'TRAINING', '2026-05-02T16:00:00', '2026-05-02T17:00:00', NULL, 1),
    (2, 5, 5, 1, N'Taller de yoga', N'Actividad institucional en sala multiuso.', 'WORKSHOP', '2026-05-03T10:00:00', '2026-05-03T11:00:00', NULL, 1),
    (3, 2, 6, 1, N'Campeonato interno', N'Evento deportivo universitario.', 'CHAMPIONSHIP', '2026-05-04T09:00:00', '2026-05-04T13:00:00', NULL, 1);
    SET IDENTITY_INSERT dbo.scheduled_activities OFF;
END;
GO

-- WORKSHOPS
IF NOT EXISTS (SELECT 1 FROM dbo.workshops WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.workshops ON;
    INSERT INTO dbo.workshops (id, resource_id, title, description, day_text, schedule_text, location, instructor_name, capacity, is_active)
    VALUES
    (1, 1, N'Judo', N'Taller deportivo DAVE.', N'Martes y jueves', N'15:45 a 17:45 h', N'Centro Deportivo', NULL, 25, 1),
    (2, 2, N'Entrenamiento funcional', N'Taller deportivo DAVE.', N'Martes y jueves', N'15:30 a 16:15 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (3, 2, N'Entrenamiento funcional', N'Taller deportivo DAVE.', N'Martes y jueves', N'17:00 a 18:00 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (4, 2, N'Entrenamiento funcional', N'Taller deportivo DAVE.', N'Lunes y miercoles', N'17:00 a 17:45 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (5, 2, N'Entrenamiento funcional', N'Taller deportivo DAVE.', N'Lunes y miercoles', N'18:30 a 19:30 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (6, 3, N'Escalada', N'Taller deportivo DAVE.', N'Martes y jueves', N'12:00 a 15:00 h', N'Muro Escalada, Centro Deportivo', NULL, 25, 1),
    (7, 4, N'Spinning', N'Taller deportivo DAVE.', N'Lunes a jueves', N'14:15 a 15:00 h', N'Sala Spinning', NULL, 25, 1),
    (8, 2, N'Pilates', N'Taller deportivo DAVE.', N'Lunes y miercoles', N'15:30 a 16:15 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (9, 1, N'Basquetbol', N'Taller deportivo DAVE.', N'Lunes y viernes', N'Lunes 16:30 a 18:30 h / Viernes 17:30 a 19:00 h', N'Centro Deportivo', NULL, 25, 1),
    (10, 2, N'Futsal', N'Taller deportivo DAVE.', N'Miercoles', N'14:00 a 15:30 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (11, 5, N'Natacion', N'Taller deportivo DAVE.', N'Martes a jueves', N'12:45 a 13:45 h', N'Piscina, Centro Deportivo', NULL, 25, 1),
    (12, 2, N'Tenis mesa', N'Taller deportivo DAVE.', N'Lunes y miercoles', N'Lunes 19:30 a 21:00 h / Miercoles 18:00 a 19:30 h', N'Cancha 2, Centro Deportivo', NULL, 25, 1),
    (13, 1, N'Voleibol femenino', N'Taller deportivo DAVE.', N'Miercoles', N'20:00 a 21:00 h', N'Centro Deportivo', NULL, 25, 1),
    (14, 1, N'Voleibol masculino', N'Taller deportivo DAVE.', N'Miercoles', N'17:00 a 19:00 h', N'Centro Deportivo', NULL, 25, 1),
    (15, 6, N'Aikido', N'Taller deportivo DAVE.', N'Miercoles', N'17:00 a 18:00 h', N'Sala Multiuso, Centro Deportivo', NULL, 25, 1),
    (16, 1, N'Esgrima', N'Taller deportivo DAVE.', N'Martes y sabado', N'Martes 19:00 a 21:00 h / Sabado 11:45 a 13:00 h', N'Centro Deportivo', NULL, 25, 1),
    (17, 7, N'Gimnasia artistica', N'Taller deportivo DAVE.', N'Viernes', N'15:30 a 16:30 h', N'Cancha 3, Centro Deportivo', NULL, 25, 1);
    SET IDENTITY_INSERT dbo.workshops OFF;
END;
GO

IF NOT EXISTS (SELECT 1 FROM dbo.workshop_occurrences WHERE workshop_id = 1 AND weekday_iso = 2 AND start_minute = 945)
BEGIN
    INSERT INTO dbo.workshop_occurrences (workshop_id, weekday_iso, start_minute, end_minute)
    VALUES
    (1,2,945,1065),(1,4,945,1065),
    (2,2,930,975),(2,4,930,975),
    (3,2,1020,1080),(3,4,1020,1080),
    (4,1,1020,1065),(4,3,1020,1065),
    (5,1,1110,1170),(5,3,1110,1170),
    (6,2,720,900),(6,4,720,900),
    (7,1,855,900),(7,2,855,900),(7,3,855,900),(7,4,855,900),
    (8,1,930,975),(8,3,930,975),
    (9,1,990,1110),(9,5,1050,1140),
    (10,3,840,930),
    (11,2,765,825),(11,3,765,825),(11,4,765,825),
    (12,1,1170,1260),(12,3,1080,1170),
    (13,3,1200,1260),
    (14,3,1020,1140),
    (15,3,1020,1080),
    (16,2,1140,1260),(16,6,705,780),
    (17,5,930,990);
END;
GO

-- VIOLATIONS
IF NOT EXISTS (SELECT 1 FROM dbo.violations WHERE id = 1)
BEGIN
    SET IDENTITY_INSERT dbo.violations ON;
    INSERT INTO dbo.violations (id, user_id, reservation_id, created_by_user_id, violation_type, description)
    VALUES
    (1, 4, 3, 1, 'NO_SHOW', N'No asistio a la reserva confirmada.'),
    (2, 12, NULL, 1, 'MISUSE', N'Intento reservar estando bloqueado.');
    SET IDENTITY_INSERT dbo.violations OFF;
END;
GO

-- NOTIFICATIONS
IF NOT EXISTS (SELECT 1 FROM dbo.notifications WHERE id = 101)
BEGIN
    SET IDENTITY_INSERT dbo.notifications ON;
    INSERT INTO dbo.notifications (id, user_id, reservation_id, title, message, type, is_read)
    VALUES
    (101, 2, 1, N'Reserva confirmada', N'Tu reserva ha sido confirmada.', 'RESERVATION_CONFIRMED', 0),
    (102, 4, 3, N'Infraccion registrada', N'Se registro una infraccion asociada a tu reserva.', 'SYSTEM', 0),
    (103, 6, 5, N'Reserva pendiente', N'Tu reserva esta pendiente de confirmacion.', 'RESERVATION_CREATED', 0);
    SET IDENTITY_INSERT dbo.notifications OFF;
END;
GO

-- AUDIT LOGS
IF NOT EXISTS (SELECT 1 FROM dbo.audit_logs WHERE id = 101)
BEGIN
    SET IDENTITY_INSERT dbo.audit_logs ON;
    INSERT INTO dbo.audit_logs (id, user_id, action, entity_type, entity_id, description)
    VALUES
    (101, 1, 'SEED_CREATED', 'database', NULL, N'Datos iniciales cargados para ambiente de desarrollo.');
    SET IDENTITY_INSERT dbo.audit_logs OFF;
END;
GO


