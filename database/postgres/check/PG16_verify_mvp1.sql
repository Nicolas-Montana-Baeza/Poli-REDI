\set ON_ERROR_STOP on

BEGIN;

DO $verify$
DECLARE
    active_policy_id integer;
    owner_id integer;
    first_user_id integer;
    second_user_id integer;
    reservable_id integer;
    pool_id integer;
    gym_id integer;
    activity_id integer;
    slot_date date;
    slot_10 timestamptz;
    slot_12 timestamptz;
    rejected boolean;
BEGIN
    IF current_setting('server_version_num')::integer < 160000 THEN
        RAISE EXCEPTION 'PostgreSQL 16 o superior es obligatorio';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'poliredi_app' AND NOT rolsuper) THEN
        RAISE EXCEPTION 'poliredi_app no existe o tiene privilegios de superusuario';
    END IF;

    IF to_regclass('public.reservations') IS NULL
       OR to_regclass('public.availability_blocks') IS NULL THEN
        RAISE EXCEPTION 'baseline MVP1 incompleto';
    END IF;

	SELECT id INTO active_policy_id
    FROM reservation_policies
    WHERE is_published AND effective_from <= CURRENT_TIMESTAMP
      AND (effective_to IS NULL OR effective_to > CURRENT_TIMESTAMP)
    ORDER BY effective_from DESC, id DESC
    LIMIT 1;

	IF active_policy_id IS NULL THEN
        RAISE EXCEPTION 'no existe politica vigente';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM reservation_policy_durations
		WHERE policy_id = active_policy_id AND duration_minutes = 180
    ) THEN
        RAISE EXCEPTION 'seed de duraciones incompleto';
    END IF;

    SELECT id INTO owner_id FROM users WHERE lower(email) = 'admin@poliredi.local';
    SELECT id INTO first_user_id FROM users WHERE lower(email) = 'usuario@poliredi.local';

    INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
    VALUES ('verify-mvp1@poliredi.local', 'Verificacion MVP1', '22222222-2', false, false)
    RETURNING id INTO second_user_id;

    SELECT id INTO reservable_id FROM resources WHERE reservation_mode = 'RESERVABLE' ORDER BY id LIMIT 1;
    SELECT id INTO pool_id FROM resources WHERE reservation_mode = 'OPEN_USE' ORDER BY id LIMIT 1;
    SELECT id INTO gym_id FROM resources WHERE reservation_mode = 'OPEN_USE' AND id <> pool_id ORDER BY id LIMIT 1;
    SELECT id INTO activity_id FROM activities WHERE is_active ORDER BY id LIMIT 1;

    slot_date := (CURRENT_TIMESTAMP AT TIME ZONE 'America/Santiago')::date + 1;
    slot_10 := (slot_date + time '10:00') AT TIME ZONE 'America/Santiago';
    slot_12 := (slot_date + time '12:00') AT TIME ZONE 'America/Santiago';

    INSERT INTO reservations (policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status)
	VALUES (active_policy_id, first_user_id, reservable_id, activity_id, slot_10, 60, 'CONFIRMED');

    rejected := false;
    BEGIN
        INSERT INTO reservations (policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status)
		VALUES (active_policy_id, second_user_id, reservable_id, activity_id, slot_10, 60, 'CONFIRMED');
    EXCEPTION WHEN exclusion_violation THEN
        rejected := true;
    END;
    IF NOT rejected THEN
        RAISE EXCEPTION 'RESERVABLE acepto un solape de recurso';
    END IF;

    INSERT INTO reservations (policy_id, user_id, resource_id, start_time, duration_minutes, status)
    VALUES
		(active_policy_id, first_user_id, pool_id, slot_12, 60, 'CONFIRMED'),
		(active_policy_id, second_user_id, pool_id, slot_12, 60, 'CONFIRMED');

    rejected := false;
    BEGIN
        INSERT INTO reservations (policy_id, user_id, resource_id, start_time, duration_minutes, status)
		VALUES (active_policy_id, first_user_id, gym_id, slot_12 + interval '30 minutes', 60, 'CONFIRMED');
    EXCEPTION WHEN exclusion_violation THEN
        rejected := true;
    END;
    IF NOT rejected THEN
        RAISE EXCEPTION 'OPEN_USE acepto un solape personal';
    END IF;

    rejected := false;
    BEGIN
        INSERT INTO availability_blocks (
            resource_id, created_by_user_id, block_type, reason, start_time, end_time, is_active
        ) VALUES (
            reservable_id, owner_id, 'ADMINISTRATIVE', 'verificacion', slot_10, slot_10 + interval '60 minutes', true
        );
    EXCEPTION WHEN SQLSTATE 'P1010' THEN
        rejected := true;
    END;
    IF NOT rejected THEN
        RAISE EXCEPTION 'bloqueo administrativo acepto solape con reserva';
    END IF;

    UPDATE reservations
    SET status = 'CANCELLED'
    WHERE user_id = first_user_id AND resource_id = reservable_id AND start_time = slot_10;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'cancelacion de verificacion no actualizo la reserva';
    END IF;
END
$verify$;

ROLLBACK;

SELECT 'PG16 MVP1 verificado correctamente' AS result;
