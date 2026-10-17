-- MVP3 / ADMIN-003
-- Impide crear o revalidar reservas activas
-- asociadas a recursos o sedes inactivas.

CREATE OR REPLACE FUNCTION validate_mvp3_reservation_venue()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    -- Las cancelaciones y estados finalizados no deben
    -- depender del estado actual de la sede.
    IF NEW.status NOT IN ('PENDING', 'CONFIRMED') THEN
        RETURN NEW;
    END IF;

    -- Bloquea las filas hasta terminar la transaccion
    -- para coordinarse con cambios concurrentes de estado.
    PERFORM 1
    FROM resources r
    INNER JOIN venues v
        ON v.id = r.venue_id
    WHERE r.id = NEW.resource_id
      AND r.is_active = true
      AND v.is_active = true
      AND r.reservation_mode IN ('RESERVABLE', 'OPEN_USE')
    FOR SHARE OF r, v;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING
            ERRCODE = 'P1002',
            MESSAGE = 'recurso o sede no disponible para reservas';
    END IF;

    RETURN NEW;
END;
$function$;

DROP TRIGGER IF EXISTS trg_check_mvp3_reservation_venue
ON reservations;

CREATE TRIGGER trg_check_mvp3_reservation_venue
BEFORE INSERT OR UPDATE OF
    policy_id,
    user_id,
    resource_id,
    start_time,
    end_time,
    duration_minutes,
    reservation_mode_snapshot,
    status
ON reservations
FOR EACH ROW
EXECUTE FUNCTION validate_mvp3_reservation_venue();
