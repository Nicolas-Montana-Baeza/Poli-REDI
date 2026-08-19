-- ============================================================================
-- POLI-REDI
-- PG16_0006_mvp2_institutional_availability.sql
--
-- MVP2 - Integración entre Programación Institucional y Disponibilidad
--
-- Objetivos:
--
--   1. Considerar actividades institucionales SCHEDULED como ocupaciones.
--   2. Impedir nuevas reservas sobre programación institucional existente.
--   3. Impedir availability_blocks sobre programación institucional existente.
--   4. Mantener semántica de intervalos [inicio, fin).
--   5. Mantener America/Santiago como calendario institucional.
--   6. Compartir el advisory lock utilizado por reservas y bloqueos.
--
-- Regla fundamental:
--
--   actividad institucional -> puede generar conflicto con ocupaciones previas
--   reserva posterior        -> NO puede ocupar ese intervalo
--   availability_block       -> NO puede ocupar ese intervalo
--
-- La detección de scheduling_conflicts sigue perteneciendo a Go.
-- PostgreSQL protege aquí la integridad/concurrencia de la disponibilidad.
-- ============================================================================

BEGIN;

SET TIME ZONE 'America/Santiago';


-- ============================================================================
-- 1. FUNCIÓN COMÚN DE SOLAPAMIENTO INSTITUCIONAL
-- ============================================================================

-- institutional_schedule_overlaps responde si un intervalo concreto intersecta
-- al menos una ocurrencia activa de una actividad institucional SCHEDULED.
--
-- Soporta:
--
--   SINGLE
--   WEEKLY
--
-- La recurrencia WEEKLY se expande únicamente dentro del rango temporal
-- consultado. No materializamos permanentemente todas sus ocurrencias.
--
-- Los intervalos utilizan semántica [inicio, fin):
--
--   actividad 10:00-11:00
--   reserva   11:00-12:00
--
-- NO se consideran solapadas.

CREATE OR REPLACE FUNCTION institutional_schedule_overlaps(
    p_resource_id integer,
    p_start timestamptz,
    p_end timestamptz
)
RETURNS boolean
LANGUAGE sql
STABLE
AS $function$

    SELECT EXISTS (

        SELECT 1

        FROM institutional_activities activity

        INNER JOIN institutional_activity_schedules schedule
            ON schedule.activity_id = activity.id

        WHERE activity.resource_id = p_resource_id

          AND activity.status = 'SCHEDULED'

          AND schedule.is_active = true

          AND (

                -- ============================================================
                -- SINGLE
                -- ============================================================

                (
                    schedule.schedule_type = 'SINGLE'

                    AND tstzrange(
                        (
                            (
                                schedule.specific_date
                                + schedule.start_time
                            )
                            AT TIME ZONE 'America/Santiago'
                        ),
                        (
                            (
                                schedule.specific_date
                                + schedule.end_time
                            )
                            AT TIME ZONE 'America/Santiago'
                        ),
                        '[)'
                    )
                    &&
                    tstzrange(
                        p_start,
                        p_end,
                        '[)'
                    )
                )


                OR


                -- ============================================================
                -- WEEKLY
                -- ============================================================

                (
                    schedule.schedule_type = 'WEEKLY'

                    AND EXISTS (

                        SELECT 1

                        FROM generate_series(

                            GREATEST(
                                schedule.valid_from,
                                (
                                    p_start
                                    AT TIME ZONE 'America/Santiago'
                                )::date
                            )::timestamp,

                            LEAST(
                                schedule.valid_to,
                                (
                                    (
                                        p_end
                                        - interval '1 microsecond'
                                    )
                                    AT TIME ZONE 'America/Santiago'
                                )::date
                            )::timestamp,

                            interval '1 day'

                        ) occurrence_date

                        WHERE extract(
                            isodow
                            FROM occurrence_date
                        )::integer = schedule.day_of_week

                          AND tstzrange(
                                (
                                    (
                                        occurrence_date::date
                                        + schedule.start_time
                                    )
                                    AT TIME ZONE 'America/Santiago'
                                ),
                                (
                                    (
                                        occurrence_date::date
                                        + schedule.end_time
                                    )
                                    AT TIME ZONE 'America/Santiago'
                                ),
                                '[)'
                          )
                          &&
                          tstzrange(
                                p_start,
                                p_end,
                                '[)'
                          )
                    )
                )
          )
    );

$function$;


-- ============================================================================
-- 2. RESERVAS VS PROGRAMACIÓN INSTITUCIONAL
-- ============================================================================

-- Una actividad institucional ya SCHEDULED protege su intervalo frente a
-- reservas creadas posteriormente.
--
-- Esto aplica tanto a:
--
--   RESERVABLE
--   OPEN_USE
--
-- La programación institucional representa ocupación institucional exclusiva
-- del recurso completo durante MVP2.
--
-- IMPORTANTE:
--
-- No utilizamos NEW.end_time como fuente principal porque otro BEFORE trigger
-- puede ejecutarse antes o después de este dependiendo de su nombre.
--
-- El final se deriva siempre desde:
--
--   start_time + duration_minutes

CREATE OR REPLACE FUNCTION
validate_mvp2_reservation_institutional_occupancy()
RETURNS trigger
LANGUAGE plpgsql
AS $function$

DECLARE

    v_end timestamptz;

BEGIN

    IF NEW.status NOT IN ('PENDING', 'CONFIRMED') THEN
        RETURN NEW;
    END IF;


    -- ------------------------------------------------------------------------
    -- Serialización por recurso.
    -- ------------------------------------------------------------------------
    --
    -- Compartimos exactamente la misma familia de advisory lock usada por
    -- reservations y availability_blocks.
    --
    -- Así una actividad institucional y una reserva concurrentes no pueden
    -- validar simultáneamente sobre estados diferentes.

    PERFORM pg_advisory_xact_lock(
        73001,
        NEW.resource_id
    );


    v_end :=
        NEW.start_time
        + make_interval(
            mins => NEW.duration_minutes
        );


    IF institutional_schedule_overlaps(
        NEW.resource_id,
        NEW.start_time,
        v_end
    ) THEN

        RAISE EXCEPTION USING

            ERRCODE = 'P1011',

            MESSAGE =
                'el recurso posee programación institucional en ese horario';

    END IF;


    RETURN NEW;

END
$function$;


DROP TRIGGER IF EXISTS
trg_validate_mvp2_reservation_institutional_occupancy
ON reservations;


CREATE TRIGGER
trg_validate_mvp2_reservation_institutional_occupancy

BEFORE INSERT OR UPDATE OF
    resource_id,
    start_time,
    duration_minutes,
    status

ON reservations

FOR EACH ROW

EXECUTE FUNCTION
validate_mvp2_reservation_institutional_occupancy();


-- ============================================================================
-- 3. AVAILABILITY BLOCKS VS PROGRAMACIÓN INSTITUCIONAL
-- ============================================================================

-- availability_blocks representan indisponibilidad dura.
--
-- Si una actividad institucional ya está programada, crear posteriormente un
-- bloqueo administrativo encima de ella produciría una contradicción:
--
--   actividad SCHEDULED
--   +
--   recurso declarado no disponible
--
-- Por eso el bloqueo es rechazado.
--
-- La dirección contraria ya se valida durante la creación de actividades:
--
--   bloque existente + actividad nueva -> actividad rechazada
--
-- Con este trigger protegemos ahora también:
--
--   actividad existente + bloque nuevo -> bloque rechazado

CREATE OR REPLACE FUNCTION
validate_mvp2_availability_block_institutional_occupancy()
RETURNS trigger
LANGUAGE plpgsql
AS $function$

BEGIN

    IF NOT NEW.is_active THEN
        RETURN NEW;
    END IF;


    PERFORM pg_advisory_xact_lock(
        73001,
        NEW.resource_id
    );


    IF institutional_schedule_overlaps(
        NEW.resource_id,
        NEW.start_time,
        NEW.end_time
    ) THEN

        RAISE EXCEPTION USING

            ERRCODE = 'P1012',

            MESSAGE =
                'el bloqueo se solapa con programación institucional activa';

    END IF;


    RETURN NEW;

END
$function$;


DROP TRIGGER IF EXISTS
trg_validate_mvp2_block_institutional_occupancy
ON availability_blocks;


CREATE TRIGGER
trg_validate_mvp2_block_institutional_occupancy

BEFORE INSERT OR UPDATE OF
    resource_id,
    start_time,
    end_time,
    is_active

ON availability_blocks

FOR EACH ROW

EXECUTE FUNCTION
validate_mvp2_availability_block_institutional_occupancy();


-- ============================================================================
-- 4. PERMISOS
-- ============================================================================

GRANT EXECUTE
ON FUNCTION institutional_schedule_overlaps(
    integer,
    timestamptz,
    timestamptz
)
TO poliredi_app;


GRANT EXECUTE
ON FUNCTION validate_mvp2_reservation_institutional_occupancy()
TO poliredi_app;


GRANT EXECUTE
ON FUNCTION validate_mvp2_availability_block_institutional_occupancy()
TO poliredi_app;


COMMIT;
