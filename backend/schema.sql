-- =========================================
-- EXTENSIONS
-- =========================================
-- Se habilita btree_gist para poder usar EXCLUDE con igualdad (=) y rangos (&&)
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- =========================================
-- FUNCTIONS
-- =========================================

-- Función para actualizar automáticamente updated_at en cada UPDATE
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función genérica de auditoría
-- Registra INSERT, UPDATE y DELETE en tabla logs
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO logs(user_id, action, table_name, record_id, new_data)
        VALUES (NEW.user_id, 'INSERT', TG_TABLE_NAME, NEW.id, to_jsonb(NEW));
        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO logs(user_id, action, table_name, record_id, old_data, new_data)
        VALUES (NEW.user_id, 'UPDATE', TG_TABLE_NAME, NEW.id, to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO logs(user_id, action, table_name, record_id, old_data)
        VALUES (OLD.user_id, 'DELETE', TG_TABLE_NAME, OLD.id, to_jsonb(OLD));
        RETURN OLD;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Función específica para detectar bloqueos/desbloqueos de usuarios
CREATE OR REPLACE FUNCTION log_user_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.is_blocked IS DISTINCT FROM NEW.is_blocked THEN
        INSERT INTO logs(user_id, action, table_name, record_id, details, old_data, new_data)
        VALUES (
            NEW.id,
            CASE WHEN NEW.is_blocked THEN 'USER_BLOCKED' ELSE 'USER_UNBLOCKED' END,
            'users',
            NEW.id,
            'Cambio en estado de bloqueo',
            to_jsonb(OLD),
            to_jsonb(NEW)
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función para registrar creación de infracciones
CREATE OR REPLACE FUNCTION log_violations()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO logs(user_id, action, table_name, record_id, details, new_data)
    VALUES (
        NEW.user_id,
        'CREATE_VIOLATION',
        'violations',
        NEW.id,
        NEW.reason,
        to_jsonb(NEW)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función para generar notificación automática al crear infracción
CREATE OR REPLACE FUNCTION notify_violation()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO notifications(user_id, message, notification_type)
    VALUES (
        NEW.user_id,
        'Has recibido una infracción: ' || NEW.reason,
        'VIOLATION'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================================
-- USERS
-- =========================================
-- Almacena usuarios del sistema (admins y usuarios normales)
CREATE TABLE if NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL, -- Identificador único
    full_name VARCHAR(255),
    is_blocked BOOLEAN DEFAULT FALSE, -- Control de acceso
    is_admin BOOLEAN DEFAULT FALSE, -- Rol administrativo
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$') -- Validación de email
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- =========================================
-- RESOURCES
-- =========================================
-- Recursos reservables (canchas, gimnasio, etc.)
CREATE TABLE if NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(100),
    reservation_mode VARCHAR(50) DEFAULT 'exclusive', -- exclusive o shared
    is_active BOOLEAN DEFAULT TRUE
);

-- =========================================
-- ACTIVITIES
-- =========================================
-- Actividades asociadas a reservas
CREATE TABLE if NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- =========================================
-- RESERVATIONS
-- =========================================
-- Tabla principal del sistema
CREATE TABLE if NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id), -- Usuario que reserva
    resource_id INT NOT NULL REFERENCES resources(id), -- Recurso reservado
    activity_id INT REFERENCES activities(id),
    start_time TIMESTAMP NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    status VARCHAR(50) DEFAULT 'CONFIRMED'
        CHECK (status IN ('CONFIRMED', 'CANCELLED', 'PENDING')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índice para optimizar consultas por recurso y tiempo
CREATE INDEX IF NOT EXISTS idx_reservations_resource_time 
ON reservations(resource_id, start_time);

-- Restricción crítica: evita solapamiento de reservas por recurso
ALTER TABLE reservations
DROP CONSTRAINT IF EXISTS no_overlap_resource;
ALTER TABLE reservations
ADD CONSTRAINT no_overlap_resource
EXCLUDE USING gist (
    resource_id WITH =,
        tsrange(
        start_time,
        start_time + (duration_minutes * INTERVAL '1 minute')
    ) WITH &&
)
WHERE (status = 'CONFIRMED');

-- Restricción: evita que un usuario tenga dos reservas simultáneas
ALTER TABLE reservations
DROP CONSTRAINT IF EXISTS no_overlap_user;
ALTER TABLE reservations
ADD CONSTRAINT no_overlap_user
EXCLUDE USING gist (
    user_id WITH =,
        tsrange(
        start_time,
        start_time + (duration_minutes * INTERVAL '1 minute')
    ) WITH &&
)
WHERE (status = 'CONFIRMED');

-- =========================================
-- PARTICIPANTS
-- =========================================
-- Usuarios adicionales en una reserva
CREATE TABLE if NOT EXISTS participants (
    id SERIAL PRIMARY KEY,
    reservation_id INT REFERENCES reservations(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id),
    confirmed BOOLEAN DEFAULT FALSE
);

-- =========================================
-- VIOLATIONS
-- =========================================
-- Registro de infracciones (ej: no asistir)
CREATE TABLE  if NOT EXISTS violations (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    reservation_id INT REFERENCES reservations(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- PRIORITY RESERVATIONS
-- =========================================
-- Reservas con prioridad (eventos, torneos)
CREATE TABLE if NOT EXISTS priority_reservations (
    id SERIAL PRIMARY KEY,
    reservation_id INT REFERENCES reservations(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- AVAILABILITY BLOCKS
-- =========================================
-- Bloqueos de disponibilidad (mantención, limpieza)
CREATE TABLE if NOT EXISTS availability_blocks (
    id SERIAL PRIMARY KEY,
    resource_id INT REFERENCES resources(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_time > start_time)
);

-- =========================================
-- NOTIFICATIONS
-- =========================================
-- Notificaciones para usuarios
CREATE TABLE if NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    message TEXT NOT NULL,
    notification_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- LOGS (AUDITORÍA)
-- =========================================
-- Registro completo de cambios del sistema
CREATE TABLE if NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    user_id INT, -- usuario asociado a la acción
    action VARCHAR(100), -- tipo de acción
    table_name VARCHAR(50), -- tabla afectada
    record_id INT, -- id del registro afectado
    old_data JSONB, -- estado anterior
    new_data JSONB, -- estado nuevo
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- TRIGGERS
-- =========================================

-- Actualización automática de timestamps
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS trg_reservations_updated_at ON reservations;
CREATE TRIGGER trg_reservations_updated_at
BEFORE UPDATE ON reservations
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Auditoría automática de reservas
DROP TRIGGER IF EXISTS trg_audit_reservations ON reservations;
CREATE TRIGGER trg_audit_reservations
AFTER INSERT OR UPDATE OR DELETE ON reservations
FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Logs de bloqueo/desbloqueo de usuarios
DROP TRIGGER IF EXISTS trg_users_block_log ON users;
CREATE TRIGGER trg_users_block_log
AFTER UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION log_user_changes();

-- Logs de infracciones
DROP TRIGGER IF EXISTS trg_log_violations ON violations;
CREATE TRIGGER trg_log_violations
AFTER INSERT ON violations
FOR EACH ROW EXECUTE FUNCTION log_violations();

-- Notificación automática por infracción
DROP TRIGGER IF EXISTS trg_notify_violation ON violations;
CREATE TRIGGER  trg_notify_violation
AFTER INSERT ON violations
FOR EACH ROW EXECUTE FUNCTION notify_violation();

-- =========================================
-- VIEWS (REPORTES)
-- =========================================

-- Uso de recursos
DROP VIEW IF EXISTS vw_resource_usage;
CREATE VIEW vw_resource_usage AS
SELECT r.name, COUNT(*) total
FROM reservations res
JOIN resources r ON r.id = res.resource_id
WHERE res.status = 'CONFIRMED'
GROUP BY r.name;

-- Horas punta
DROP VIEW IF EXISTS vw_peak_hours;
CREATE VIEW vw_peak_hours AS
SELECT EXTRACT(HOUR FROM start_time) _hour, COUNT(*) total
FROM reservations
WHERE status = 'CONFIRMED'
GROUP BY _hour
ORDER BY total DESC;

-- Infracciones por usuario
DROP VIEW IF EXISTS vw_user_violations;
CREATE VIEW vw_user_violations AS
SELECT u.email, COUNT(v.id) total
FROM users u
LEFT JOIN violations v ON v.user_id = u.id
GROUP BY u.email;