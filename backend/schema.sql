-- =========================================
-- USERS TABLE
-- =========================================
-- Creating users table...

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255),
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    is_admin BOOLEAN DEFAULT FALSE


);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- =========================================
-- RESOURCES TABLE
-- =========================================
-- Creating resources table...

CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(100),
    reservation_mode VARCHAR(50) DEFAULT 'exclusive',
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_resources_name ON resources(name);

-- =========================================
-- ACTIVITIES TABLE
-- =========================================
-- Creating activities table...

CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- =========================================
-- RESERVATIONS TABLE
-- =========================================
-- Creating reservations table...

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    resource_id INT NOT NULL REFERENCES resources(id),
    activity_id INT REFERENCES activities(id),
    start_time TIMESTAMP NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    status VARCHAR(50) DEFAULT 'CONFIRMED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reservations_user_id ON reservations(user_id);
CREATE INDEX IF NOT EXISTS idx_reservations_resource_id ON reservations(resource_id);
CREATE INDEX IF NOT EXISTS idx_reservations_start_time ON reservations(start_time);

-- =========================================
-- PARTICIPANTS TABLE
-- =========================================
-- Creating participants table...

CREATE TABLE IF NOT EXISTS participants (
    id SERIAL PRIMARY KEY,
    reservation_id INT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    confirmed BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_participants_reservation_id ON participants(reservation_id);

-- =========================================
-- VIOLATIONS TABLE
-- =========================================
-- Creating violations table...

CREATE TABLE IF NOT EXISTS violations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    reservation_id INT REFERENCES reservations(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- PRIORITY RESERVATIONS TABLE
-- =========================================
-- Creating priority reservations table...

CREATE TABLE IF NOT EXISTS priority_reservations (
    id SERIAL PRIMARY KEY,
    reservation_id INT NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- AVAILABILITY BLOCKS TABLE
-- =========================================
-- Creating availability blocks table...

CREATE TABLE IF NOT EXISTS availability_blocks (
    id SERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resources(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);

CREATE TABLE If not EXISTS violation_notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    reservation_id INT NOT NULL REFERENCES reservations(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_violation_notifications_user_id ON violation_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_violation_notifications_reservation_id ON violation_notifications(reservation_id);

CREATE INDEX IF NOT EXISTS idx_availability_blocks_start_time ON availability_blocks(start_time);
CREATE INDEX IF NOT EXISTS idx_availability_blocks_end_time ON availability_blocks(end_time);   
CREATE INDEX IF NOT EXISTS idx_availability_blocks_resource_id ON availability_blocks(resource_id);

CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- REPORT VIEWS
-- =========================================
-- Creating views...

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

CREATE OR REPLACE VIEW vw_activity_usage AS
SELECT
    a.name AS activity_name,
    COUNT(*) AS total
FROM reservations r
JOIN activities a ON a.id = r.activity_id
GROUP BY a.name;

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

-- Schema completed successfully.