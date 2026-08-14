BEGIN;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    NEW.updated_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS trg_venues_updated_at ON venues;
CREATE TRIGGER trg_venues_updated_at BEFORE UPDATE ON venues
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_resources_updated_at ON resources;
CREATE TRIGGER trg_resources_updated_at BEFORE UPDATE ON resources
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_activities_updated_at ON activities;
CREATE TRIGGER trg_activities_updated_at BEFORE UPDATE ON activities
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_reservations_updated_at ON reservations;
CREATE TRIGGER trg_reservations_updated_at BEFORE UPDATE ON reservations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION validate_mvp1_reservation()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    policy reservation_policies%ROWTYPE;
    local_start timestamp;
    local_end timestamp;
    start_minute integer;
    resource_mode varchar(50);
	local_today date;
BEGIN
    NEW.end_time := NEW.start_time + make_interval(mins => NEW.duration_minutes);

    IF NEW.status NOT IN ('PENDING', 'CONFIRMED') THEN
        RETURN NEW;
    END IF;

    -- Serializa las escrituras que compiten por el mismo recurso, incluida la
    -- tabla availability_blocks. Fase 2 debe mantener la transaccion hasta commit.
    PERFORM pg_advisory_xact_lock(73001, NEW.resource_id);

    IF NOT EXISTS (
        SELECT 1 FROM users u
        WHERE u.id = NEW.user_id
          AND NOT u.is_blocked
          AND (u.is_admin OR u.rut IS NOT NULL)
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = 'P1001',
            MESSAGE = 'usuario bloqueado o sin RUT habilitante';
    END IF;

	IF NEW.activity_id IS NOT NULL AND NOT EXISTS (
		SELECT 1 FROM activities activity
		WHERE activity.id = NEW.activity_id AND activity.is_active
	) THEN
		RAISE EXCEPTION USING ERRCODE = 'P1009', MESSAGE = 'actividad inexistente o inactiva';
	END IF;

    SELECT r.reservation_mode INTO resource_mode
    FROM resources r
        WHERE r.id = NEW.resource_id
          AND r.is_active
          AND r.reservation_mode IN ('RESERVABLE', 'OPEN_USE');

    IF resource_mode IS NULL THEN
        RAISE EXCEPTION USING
            ERRCODE = 'P1002',
            MESSAGE = 'recurso no disponible para reserva individual';
    END IF;

    NEW.reservation_mode_snapshot := resource_mode;

    SELECT p.* INTO STRICT policy
    FROM reservation_policies p
    WHERE p.id = NEW.policy_id
      AND p.is_published
      AND p.effective_from <= CURRENT_TIMESTAMP
      AND (p.effective_to IS NULL OR p.effective_to > CURRENT_TIMESTAMP);

    IF NOT EXISTS (
        SELECT 1 FROM reservation_policy_resources scope
        WHERE scope.policy_id = NEW.policy_id
          AND scope.resource_id = NEW.resource_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P1003', MESSAGE = 'recurso fuera de la politica vigente';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM reservation_policy_durations duration
        WHERE duration.policy_id = NEW.policy_id
          AND duration.duration_minutes = NEW.duration_minutes
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P1004', MESSAGE = 'duracion fuera de la politica vigente';
    END IF;

    local_start := NEW.start_time AT TIME ZONE 'America/Santiago';
    local_end := (NEW.start_time + make_interval(mins => NEW.duration_minutes)) AT TIME ZONE 'America/Santiago';
    start_minute := extract(hour FROM local_start)::integer * 60 + extract(minute FROM local_start)::integer;

    IF local_start::date <> local_end::date
       OR start_minute < policy.opening_minute
       OR start_minute + NEW.duration_minutes > policy.closing_minute
       OR start_minute % policy.slot_interval_minutes <> 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P1005', MESSAGE = 'horario fuera de la politica vigente';
    END IF;

	local_today := (CURRENT_TIMESTAMP AT TIME ZONE 'America/Santiago')::date;

	IF local_start::date < local_today
	   OR local_start::date >= local_today + policy.reservable_window_days THEN
        RAISE EXCEPTION USING ERRCODE = 'P1006', MESSAGE = 'fecha fuera de la ventana reservable';
    END IF;

    IF EXISTS (
        SELECT 1 FROM availability_blocks block
        WHERE block.resource_id = NEW.resource_id
          AND block.is_active
          AND tstzrange(block.start_time, block.end_time, '[)')
              && tstzrange(NEW.start_time, NEW.start_time + make_interval(mins => NEW.duration_minutes), '[)')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P1007', MESSAGE = 'horario bloqueado para el recurso';
    END IF;

    RETURN NEW;
EXCEPTION
    WHEN no_data_found THEN
        RAISE EXCEPTION USING ERRCODE = 'P1008', MESSAGE = 'politica inexistente o no vigente';
END
$function$;

DROP TRIGGER IF EXISTS trg_validate_mvp1_reservation ON reservations;
CREATE TRIGGER trg_validate_mvp1_reservation
BEFORE INSERT OR UPDATE OF policy_id, user_id, resource_id, start_time, end_time, duration_minutes, reservation_mode_snapshot, status
ON reservations
FOR EACH ROW EXECUTE FUNCTION validate_mvp1_reservation();

CREATE OR REPLACE FUNCTION validate_mvp1_availability_block()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    IF NOT NEW.is_active THEN
        RETURN NEW;
    END IF;

    PERFORM pg_advisory_xact_lock(73001, NEW.resource_id);

    IF EXISTS (
        SELECT 1 FROM reservations reservation
        WHERE reservation.resource_id = NEW.resource_id
          AND reservation.status IN ('PENDING', 'CONFIRMED')
          AND tstzrange(reservation.start_time, reservation.end_time, '[)')
              && tstzrange(NEW.start_time, NEW.end_time, '[)')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P1010', MESSAGE = 'el bloqueo se solapa con una reserva activa';
    END IF;

    RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS trg_validate_mvp1_availability_block ON availability_blocks;
CREATE TRIGGER trg_validate_mvp1_availability_block
BEFORE INSERT OR UPDATE OF resource_id, start_time, end_time, is_active
ON availability_blocks
FOR EACH ROW EXECUTE FUNCTION validate_mvp1_availability_block();

GRANT EXECUTE ON FUNCTION set_updated_at() TO poliredi_app;
GRANT EXECUTE ON FUNCTION validate_mvp1_reservation() TO poliredi_app;
GRANT EXECUTE ON FUNCTION validate_mvp1_availability_block() TO poliredi_app;

COMMIT;
