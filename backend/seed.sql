-- =========================================
-- USERS
-- =========================================
-- Se crea 1 administrador para gestionar recursos, bloqueos y reservas prioritarias.
-- Se crean 10 usuarios normales para simular uso real del sistema.
-- Se crea 1 usuario bloqueado para probar restricciones de acceso.

INSERT INTO users (email, full_name, is_admin, is_blocked)
VALUES
('admin@universidad.cl', 'Administrador General', TRUE, FALSE), -- Admin principal
('nicolas@universidad.cl', 'Nicolás Montaña', FALSE, FALSE), -- Usuario normal
('maria@universidad.cl', 'María González', FALSE, FALSE),
('juan@universidad.cl', 'Juan Pérez', FALSE, FALSE),
('camila@universidad.cl', 'Camila Soto', FALSE, FALSE),
('pedro@universidad.cl', 'Pedro Ramírez', FALSE, FALSE),
('valentina@universidad.cl', 'Valentina Fuentes', FALSE, FALSE),
('sebastian@universidad.cl', 'Sebastián Morales', FALSE, FALSE),
('fernanda@universidad.cl', 'Fernanda Rojas', FALSE, FALSE),
('diego@universidad.cl', 'Diego Herrera', FALSE, FALSE),
('sofia@universidad.cl', 'Sofía Castillo', FALSE, FALSE),
('bloqueado@universidad.cl', 'Usuario Bloqueado', FALSE, TRUE); -- Usuario bloqueado

-- =========================================
-- RESOURCES
-- =========================================
-- Recursos deportivos disponibles para reservas.

INSERT INTO resources (name, type, reservation_mode)
VALUES
('Cancha de Fútbol', 'Cancha', 'exclusive'), -- Solo una reserva a la vez
('Cancha de Básquetbol', 'Cancha', 'exclusive'),
('Piscina', 'Piscina', 'exclusive'),
('Gimnasio', 'Gimnasio', 'shared'), -- Permite múltiples usuarios
('Sala Multiuso', 'Sala', 'shared');

-- =========================================
-- ACTIVITIES
-- =========================================
-- Actividades asociables a las reservas.

INSERT INTO activities (name)
VALUES
('Fútbol'),
('Básquetbol'),
('Natación'),
('Entrenamiento Libre'),
('Yoga');

-- =========================================
-- RESERVATIONS
-- =========================================
-- Reservas de ejemplo en distintos horarios para probar:
-- - reservas confirmadas
-- - reservas pendientes
-- - reportes de uso
-- - horas punta

INSERT INTO reservations (user_id, resource_id, activity_id, start_time, duration_minutes, status)
VALUES
(2, 1, 1, '2026-04-30 10:00:00', 60, 'CONFIRMED'), -- Nicolás reserva cancha fútbol
(3, 2, 2, '2026-04-30 11:30:00', 90, 'CONFIRMED'), -- María reserva básquet
(4, 3, 3, '2026-04-30 13:30:00', 60, 'CONFIRMED'), -- Juan piscina
(5, 4, 4, '2026-04-30 15:00:00', 120, 'CONFIRMED'), -- Camila gimnasio
(6, 5, 5, '2026-04-30 18:00:00', 60, 'PENDING'), -- Pedro pendiente aprobación
(7, 1, 1, '2026-05-01 09:00:00', 60, 'CONFIRMED'), -- Valentina otro día
(8, 2, 2, '2026-05-01 11:00:00', 60, 'CONFIRMED'),
(9, 3, 3, '2026-05-01 14:00:00', 60, 'CONFIRMED');

-- =========================================
-- PARTICIPANTS
-- =========================================
-- Participantes invitados o asociados a reservas grupales.

INSERT INTO participants (reservation_id, user_id, confirmed)
VALUES
(1, 3, TRUE), -- María participa con Nicolás
(1, 4, TRUE), -- Juan participa con Nicolás
(2, 2, TRUE), -- Nicolás participa con María
(4, 5, FALSE), -- Camila invitó a Pedro pero no confirmó
(6, 8, TRUE),
(7, 9, TRUE);

-- =========================================
-- VIOLATIONS
-- =========================================
-- Infracciones para probar sanciones y reportes.

INSERT INTO violations (user_id, reservation_id, reason)
VALUES
(4, 3, 'No asistió a la reserva'), -- Juan no-show
(12, NULL, 'Intentó reservar estando bloqueado'); -- Usuario bloqueado

-- =========================================
-- PRIORITY RESERVATIONS
-- =========================================
-- Reservas especiales creadas por administrador.

INSERT INTO priority_reservations (reservation_id, reason, created_by)
VALUES
(1, 'Torneo universitario', 1),
(3, 'Evento institucional', 1);

-- =========================================
-- AVAILABILITY BLOCKS
-- =========================================
-- Bloqueos por mantención o limpieza.

INSERT INTO availability_blocks (resource_id, start_time, end_time, reason)
VALUES
(1, '2026-05-02 08:00:00', '2026-05-02 12:00:00', 'Mantención'),
(3, '2026-05-03 14:00:00', '2026-05-03 18:00:00', 'Limpieza');

-- =========================================
-- NOTIFICATIONS
-- =========================================
-- Notificaciones para probar lectura y tipos.

INSERT INTO notifications (user_id, message, notification_type)
VALUES
(2, 'Tu reserva ha sido confirmada.', 'GENERAL'),
(4, 'Has recibido una infracción.', 'VIOLATION'),
(6, 'Tu reserva está pendiente.', 'GENERAL'),
(12, 'Tu cuenta está bloqueada.', 'VIOLATION');

-- =========================================
-- LOGS
-- =========================================
-- Registro de auditoría para trazabilidad.

INSERT INTO logs (user_id, action, details)
VALUES
(1, 'CREATE_RESOURCE', 'Se creó Cancha de Fútbol'),
(2, 'CREATE_RESERVATION', 'Reserva creada para Cancha de Fútbol'),
(4, 'NO_SHOW', 'Usuario no asistió a reserva'),
(12, 'BLOCKED_ATTEMPT', 'Intento de reserva bloqueado');