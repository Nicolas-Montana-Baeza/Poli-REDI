-- ============================================================================
-- POLI-REDI
-- PG16_0008_mvp2_schedule_exception_availability.sql
--
-- MVP2 - Excepciones institucionales dentro de la disponibilidad
--
-- Una excepción CANCEL elimina una ocurrencia original.
--
-- Una excepción RESCHEDULE:
--
--   1. elimina la ocurrencia original;
--   2. materializa una nueva ocurrencia concreta.
--
-- Esta migración actualiza la función de disponibilidad creada en 0006.
-- ============================================================================

BEGIN;

SET TIME ZONE 'America/Santiago';


-- ============================================================================
-- SOLAPAMIENTO INSTITUCIONAL CON EXCEPCIONES
-- ============================================================================

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

        -- ====================================================================
        -- OCURRENCIAS ORIGINALES ACTIVAS
        -- ====================================================================

        SELECT 1

        FROM institutional_activities activity

        INNER JOIN institutional_activity_schedules schedule
            ON schedule.activity_id = activity.id

        WHERE activity.resource_id = p_resource_id
          AND activity.status = 'SCHEDULED'
          AND schedule.is_active = true

          AND (

                -- ------------------------------------------------------------
                -- SINGLE
                -- ------------------------------------------------------------

                (
                    schedule.schedule_type = 'SINGLE'

                    -- Una excepción elimina la ocurrencia original.
                    AND NOT EXISTS (

                        SELECT 1

                        FROM institutional_activity_schedule_exceptions exception

                        WHERE exception.schedule_id = schedule.id
                          AND exception.original_date =
                                schedule.specific_date
                    )

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


                -- ------------------------------------------------------------
                -- WEEKLY
                -- ------------------------------------------------------------

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


                          -- Una excepción CANCEL o RESCHEDULE reemplaza la
                          -- ocurrencia original de esa fecha.
                          AND NOT EXISTS (

                                SELECT 1

                                FROM institutional_activity_schedule_exceptions exception

                                WHERE exception.schedule_id = schedule.id
                                  AND exception.original_date =
                                        occurrence_date::date
                          )


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


        UNION ALL


        -- ====================================================================
        -- OCURRENCIAS REPROGRAMADAS
        -- ====================================================================

        SELECT 1

        FROM institutional_activity_schedule_exceptions exception

        INNER JOIN institutional_activities activity
            ON activity.id = exception.activity_id

        INNER JOIN institutional_activity_schedules schedule
            ON schedule.id = exception.schedule_id
           AND schedule.activity_id = exception.activity_id

        WHERE activity.resource_id = p_resource_id
          AND activity.status = 'SCHEDULED'
          AND schedule.is_active = true
          AND exception.exception_type = 'RESCHEDULE'

          AND tstzrange(
                (
                    (
                        exception.new_date
                        + exception.new_start_time
                    )
                    AT TIME ZONE 'America/Santiago'
                ),
                (
                    (
                        exception.new_date
                        + exception.new_end_time
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
    );

$function$;


GRANT EXECUTE
ON FUNCTION institutional_schedule_overlaps(
    integer,
    timestamptz,
    timestamptz
)
TO poliredi_app;


COMMIT;
