-- =========================================
-- STOP ON FIRST ERROR
-- =========================================
\set ON_ERROR_STOP on

\echo 'Checking database...'

SELECT 'CREATE DATABASE poliredi'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'poliredi'
)\gexec

\echo 'Connecting to poliredi...'

\c poliredi

-- =========================================
-- USERS
-- =========================================
\echo 'Creating users table...'

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255),
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- =========================================
-- RESOURCES
-- =========================================
\echo 'Creating resources table...'

CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(100),
    reservation_mode VARCHAR(50) DEFAULT 'exclusive',
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_resources_name ON resources(name);

-- =========================================
-- SPORTS
-- =========================================
\echo 'Creating sports table...'

CREATE TABLE IF NOT EXISTS sports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- =========================================
-- RESERVATIONS
-- =========================================
\echo 'Creating reservations table...'

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    resource_id INT NOT NULL REFERENCES resources(id),
    sport_id INT REFERENCES sports(id),
    start_time TIMESTAMP NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    status VARCHAR(50) DEFAULT 'CONFIRMED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reservations_user_id ON reservations(user_id);
CREATE INDEX IF NOT EXISTS idx_reservations_resource_id ON reservations(resource_id);
CREATE INDEX IF NOT EXISTS idx_reservations_start_time ON reservations(start_time);

-- =========================================
-- PARTICIPANTS
-- =========================================
\echo 'Creating participants table...'

CREATE TABLE IF NOT EXISTS participants (
    id SERIAL PRIMARY KEY,
    reservation_id INT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    confirmed BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_participants_reservation_id ON participants(reservation_id);

-- =========================================
-- VIOLATIONS
-- =========================================
\echo 'Creating violations table...'

CREATE TABLE IF NOT EXISTS violations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    reservation_id INT REFERENCES reservations(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- PRIORITY RESERVATIONS
-- =========================================
\echo 'Creating priority reservations table...'

CREATE TABLE IF NOT EXISTS priority_reservations (
    id SERIAL PRIMARY KEY,
    reservation_id INT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- AVAILABILITY BLOCKS
-- =========================================
\echo 'Creating availability blocks table...'

CREATE TABLE IF NOT EXISTS availability_blocks (
    id SERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resources(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_availability_blocks_resource_id ON availability_blocks(resource_id);

-- =========================================
-- SEED USERS
-- =========================================
\echo 'Seeding users...'

INSERT INTO users (email, full_name, is_blocked) VALUES
('nicolas@ucentral.cl', 'Nicolás Montaña', FALSE),
('maria@ucentral.cl', 'María González', FALSE),
('juan@ucentral.cl', 'Juan Pérez', FALSE),
('camila@ucentral.cl', 'Camila Soto', TRUE),
('dave@ucentral.cl', 'Dave Coordinador', FALSE)
ON CONFLICT (email) DO NOTHING;

-- =========================================
-- SEED RESOURCES
-- =========================================
\echo 'Seeding resources...'

INSERT INTO resources (name, type, reservation_mode) VALUES
('Cancha 1', 'Cancha', 'exclusive'),
('Cancha 2', 'Cancha', 'exclusive'),
('Cancha 3', 'Cancha', 'exclusive'),
('Gimnasio', 'Free Pass', 'concurrent'),
('Piscina', 'Free Pass', 'concurrent'),
('Sala Multiuso', 'Sala', 'exclusive')
ON CONFLICT (name) DO NOTHING;

-- =========================================
-- SEED SPORTS
-- =========================================
\echo 'Seeding sports...'

INSERT INTO sports (name) VALUES
('Fútbol'),
('Vóley'),
('Básquet'),
('Tenis'),
('Natación'),
('Cardio'),
('Musculación'),
('Libre'),
('Yoga'),
('Baile')
ON CONFLICT (name) DO NOTHING;

-- =========================================
-- REPORT VIEWS
-- =========================================
\echo 'Creating views...'

CREATE OR REPLACE VIEW vw_resource_usage AS
SELECT
    r.name AS resource_name,
    COUNT(res.id) AS total_reservations
FROM reservations res
JOIN resources r ON r.id = res.resource_id
GROUP BY r.name;

CREATE OR REPLACE VIEW vw_peak_hours AS
SELECT
    EXTRACT(HOUR FROM start_time) AS hour,
    COUNT(*) AS reservations_count
FROM reservations
GROUP BY hour
ORDER BY reservations_count DESC;

CREATE OR REPLACE VIEW vw_sport_usage AS
SELECT
    s.name AS sport_name,
    COUNT(*) AS total
FROM reservations r
JOIN sports s ON s.id = r.sport_id
GROUP BY s.name;

CREATE OR REPLACE VIEW vw_user_violations AS
SELECT
    u.full_name,
    u.email,
    COUNT(v.id) AS total_violations
FROM users u
LEFT JOIN violations v ON v.user_id = u.id
GROUP BY u.id;

CREATE OR REPLACE VIEW vw_priority_reservations AS
SELECT
    r.id,
    u.full_name,
    res.name AS resource_name,
    r.start_time,
    p.reason
FROM priority_reservations p
JOIN reservations r ON r.id = p.reservation_id
JOIN users u ON u.id = r.user_id
JOIN resources res ON res.id = r.resource_id;

\echo 'Schema completed successfully.'