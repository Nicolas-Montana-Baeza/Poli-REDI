BEGIN;

SET TIME ZONE 'America/Santiago';

INSERT INTO venues (name, address_line, commune, city, region, country, latitude, longitude, is_active)
VALUES (
    'Centro Deportivo',
    'Dirección local de demostración',
    'Santiago',
    'Santiago',
    'Región Metropolitana',
    'Chile',
    -33.4551000,
    -70.6415000,
    true
)
ON CONFLICT (name) DO UPDATE SET
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

UPDATE users SET full_name = 'Administrador Local', is_admin = true, is_blocked = false
WHERE lower(email) = 'admin@poliredi.local';
INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
SELECT 'admin@poliredi.local', 'Administrador Local', NULL, true, false
WHERE NOT EXISTS (SELECT 1 FROM users WHERE lower(email) = 'admin@poliredi.local');

UPDATE users SET full_name = 'Usuario Prueba', rut = '12345678-5', is_admin = false, is_blocked = false
WHERE lower(email) = 'usuario@poliredi.local';
INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
SELECT 'usuario@poliredi.local', 'Usuario Prueba', '12345678-5', false, false
WHERE NOT EXISTS (SELECT 1 FROM users WHERE lower(email) = 'usuario@poliredi.local');

UPDATE users SET full_name = 'Usuario Bloqueado', rut = '11111111-1', is_admin = false, is_blocked = true
WHERE lower(email) = 'bloqueado@poliredi.local';
INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
SELECT 'bloqueado@poliredi.local', 'Usuario Bloqueado', '11111111-1', false, true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE lower(email) = 'bloqueado@poliredi.local');

INSERT INTO resources (venue_id, name, type, reservation_mode, image_url, capacity, is_active)
SELECT venue.id, seed.name, seed.type, seed.mode, seed.image_url, seed.capacity, true
FROM venues venue
CROSS JOIN (VALUES
    ('Cancha 1, Centro Deportivo', 'Cancha', 'RESERVABLE', 'https://images.unsplash.com/photo-1574629810360-7efbbe195018?auto=format&fit=crop&w=900&q=80', 22),
    ('Cancha 2, Centro Deportivo', 'Cancha', 'RESERVABLE', 'https://images.unsplash.com/photo-1574629810360-7efbbe195018?auto=format&fit=crop&w=900&q=80', 22),
    ('Muro Escalada, Centro Deportivo', 'Muro Escalada', 'RESERVABLE', 'https://images.unsplash.com/photo-1522163182402-834f871fd851?auto=format&fit=crop&w=900&q=80', 20),
    ('Piscina, Centro Deportivo', 'Piscina', 'OPEN_USE', 'https://images.unsplash.com/photo-1575429198097-0414ec08e8cd?auto=format&fit=crop&w=900&q=80', 20),
    ('Gimnasio, Centro Deportivo', 'Gimnasio', 'OPEN_USE', 'https://images.unsplash.com/photo-1534438327276-14e5300c3a48?auto=format&fit=crop&w=900&q=80', 40)
) AS seed(name, type, mode, image_url, capacity)
WHERE venue.name = 'Centro Deportivo'
ON CONFLICT (venue_id, name) DO UPDATE SET
    type = EXCLUDED.type,
    reservation_mode = EXCLUDED.reservation_mode,
    image_url = EXCLUDED.image_url,
    capacity = EXCLUDED.capacity,
    is_active = true,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO activities (name, description, is_active)
VALUES
    ('Fútbol', 'Partidos o entrenamientos de fútbol.', true),
    ('Natación', 'Uso deportivo de piscina.', true),
    ('Entrenamiento libre', 'Uso general de gimnasio.', true)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = true,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO reservation_policies (
    reservable_window_days,
    request_frequency_days,
    confirmation_deadline_minutes,
    minimum_participants,
    opening_minute,
    closing_minute,
    slot_interval_minutes,
    effective_from,
    created_by_user_id,
    idempotency_key,
    idempotency_payload_hash,
    is_published
)
SELECT
    14,
    1,
    0,
    1,
    480,
    1320,
    15,
    timestamptz '2026-08-13 00:00:00 America/Santiago',
    admin.id,
    'mvp1-local-baseline-20260813',
    repeat('0', 64),
    true
FROM users admin
WHERE lower(admin.email) = 'admin@poliredi.local'
  AND NOT EXISTS (
      SELECT 1 FROM reservation_policies
      WHERE idempotency_key = 'mvp1-local-baseline-20260813'
  );

INSERT INTO reservation_policy_resources (policy_id, resource_id)
SELECT policy.id, resource.id
FROM reservation_policies policy
CROSS JOIN resources resource
WHERE policy.idempotency_key = 'mvp1-local-baseline-20260813'
  AND resource.is_active
  AND resource.reservation_mode IN ('RESERVABLE', 'OPEN_USE')
ON CONFLICT DO NOTHING;

INSERT INTO reservation_policy_durations (policy_id, duration_minutes)
SELECT policy.id, duration.minutes
FROM reservation_policies policy
CROSS JOIN (VALUES (30), (45), (60), (90), (120), (150), (180)) AS duration(minutes)
WHERE policy.idempotency_key = 'mvp1-local-baseline-20260813'
ON CONFLICT DO NOTHING;

COMMIT;
