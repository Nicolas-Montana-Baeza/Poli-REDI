-- ==========================================================================
-- PG16_0010_mvp2_group_resource_rules.sql
--
-- MVP2 14A:
--   - minimum_participants pasa a configurarse por recurso grupal;
--   - cada reserva grupal conserva el minimo como snapshot;
--   - Sala Multiuso se incorpora al flujo grupal;
--   - el deadline de confirmacion vigente queda en 60 minutos;
--   - group_recovery_deadline_minutes se conserva solo por compatibilidad.
--
-- Esta migracion es prospectiva: crea una nueva version publicada de la
-- politica y no modifica la politica asociada a reservas existentes.
-- ==========================================================================

BEGIN;


-- ==========================================================================
-- 1. MINIMO POR RECURSO GRUPAL
-- ==========================================================================

ALTER TABLE reservation_policy_group_resources
    ADD COLUMN IF NOT EXISTS minimum_participants integer;


UPDATE reservation_policy_group_resources group_rule

SET minimum_participants = policy.minimum_participants

FROM reservation_policies policy

WHERE policy.id = group_rule.policy_id
  AND group_rule.minimum_participants IS NULL;


ALTER TABLE reservation_policy_group_resources
    ALTER COLUMN minimum_participants SET NOT NULL;


DO $migration$
BEGIN

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_group_resources_minimum_participants'
          AND conrelid =
              'reservation_policy_group_resources'::regclass
    ) THEN

        ALTER TABLE reservation_policy_group_resources
            ADD CONSTRAINT ck_group_resources_minimum_participants
            CHECK (minimum_participants > 0);

    END IF;

END
$migration$;


COMMENT ON COLUMN
    reservation_policy_group_resources.minimum_participants
IS
    'Minimo versionado para este recurso grupal dentro de la politica.';


-- ==========================================================================
-- 2. SNAPSHOT DEL MINIMO EN LA RESERVA
-- ==========================================================================

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS group_minimum_participants_snapshot integer;


UPDATE reservations reservation

SET group_minimum_participants_snapshot =
    COALESCE(
        (
            SELECT group_rule.minimum_participants
            FROM reservation_policy_group_resources group_rule
            WHERE group_rule.policy_id = reservation.policy_id
              AND group_rule.resource_id = reservation.resource_id
        ),
        policy.minimum_participants
    )

FROM reservation_policies policy

WHERE reservation.policy_id = policy.id
  AND reservation.group_capacity_snapshot IS NOT NULL
  AND reservation.group_minimum_participants_snapshot IS NULL;


DO $migration$
BEGIN

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_reservations_group_minimum_snapshot'
          AND conrelid = 'reservations'::regclass
    ) THEN

        ALTER TABLE reservations
            ADD CONSTRAINT ck_reservations_group_minimum_snapshot
            CHECK (
                (
                    group_capacity_snapshot IS NULL
                    AND group_minimum_participants_snapshot IS NULL
                )
                OR
                (
                    group_capacity_snapshot IS NOT NULL
                    AND group_minimum_participants_snapshot IS NOT NULL
                    AND group_minimum_participants_snapshot > 0
                    AND group_minimum_participants_snapshot
                        <= group_capacity_snapshot
                )
            );

    END IF;

END
$migration$;


COMMENT ON COLUMN reservations.group_minimum_participants_snapshot
IS
    'Minimo historico exigido al crear la reserva grupal.';


-- ==========================================================================
-- 3. VALIDACION DEL RECURSO GRUPAL
-- ==========================================================================

CREATE OR REPLACE FUNCTION validate_group_policy_resource()
RETURNS trigger
LANGUAGE plpgsql
AS $function$

DECLARE

    v_capacity integer;
    v_mode varchar(50);

BEGIN

    SELECT
        resource.capacity,
        resource.reservation_mode

    INTO
        v_capacity,
        v_mode

    FROM resources resource

    WHERE resource.id = NEW.resource_id;


    -- Las foreign keys resuelven normalmente este escenario.
    IF NOT FOUND THEN
        RETURN NEW;
    END IF;


    IF v_mode <> 'RESERVABLE' THEN

        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE =
                'un recurso grupal debe utilizar reservation_mode RESERVABLE';

    END IF;


    IF NEW.minimum_participants <= 0 THEN

        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE =
                'el minimo del recurso grupal debe ser mayor que cero';

    END IF;


    IF v_capacity IS NULL
       OR v_capacity < NEW.minimum_participants THEN

        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE =
                'la capacidad del recurso grupal es inferior a su minimo';

    END IF;


    RETURN NEW;

END
$function$;


-- El trigger creado por PG16_0004 conserva su nombre y pasa a utilizar esta
-- nueva definicion de la funcion.


-- ==========================================================================
-- 4. SALA MULTIUSO EN POSTGRESQL
-- ==========================================================================

INSERT INTO resources (
    venue_id,
    name,
    type,
    reservation_mode,
    image_url,
    capacity,
    is_active
)

SELECT
    venue.id,
    'Sala Multiuso, Centro Deportivo',
    'Sala',
    'RESERVABLE',
    'https://images.unsplash.com/photo-1517457373958-b7bdd4587205?auto=format&fit=crop&w=900&q=80',
    25,
    true

FROM venues venue

WHERE venue.name = 'Centro Deportivo'

ON CONFLICT (venue_id, name)
DO UPDATE SET
    type = EXCLUDED.type,
    reservation_mode = 'RESERVABLE',
    image_url = COALESCE(resources.image_url, EXCLUDED.image_url),
    capacity = COALESCE(resources.capacity, EXCLUDED.capacity),
    is_active = true,
    updated_at = CURRENT_TIMESTAMP;


-- ==========================================================================
-- 5. NUEVA POLITICA PROSPECTIVA MVP2
-- ==========================================================================

DO $migration$

DECLARE

    v_now timestamptz := CURRENT_TIMESTAMP;
    v_old_policy_id integer;
    v_new_policy_id integer;
    v_target_count integer;
    v_default_minimum integer := 10;

BEGIN

    SELECT id

    INTO v_new_policy_id

    FROM reservation_policies

    WHERE idempotency_key = 'mvp2-group-resource-rules-v1';


    IF v_new_policy_id IS NOT NULL THEN
        RETURN;
    END IF;


    SELECT id

    INTO v_old_policy_id

    FROM reservation_policies

    WHERE is_published
      AND effective_to IS NULL

    ORDER BY effective_from DESC, id DESC

    LIMIT 1

    FOR UPDATE;


    IF v_old_policy_id IS NULL THEN
        RAISE EXCEPTION
            'no existe una politica vigente desde la cual crear PG16_0010';
    END IF;


    SELECT COUNT(*)

    INTO v_target_count

    FROM resources

    WHERE name IN (
        'Cancha 1, Centro Deportivo',
        'Cancha 2, Centro Deportivo',
        'Sala Multiuso, Centro Deportivo'
    )
      AND is_active
      AND reservation_mode = 'RESERVABLE'
      AND capacity >= v_default_minimum;


    IF v_target_count <> 3 THEN
        RAISE EXCEPTION
            'no existen los tres recursos grupales MVP2 con capacidad valida';
    END IF;


    UPDATE reservation_policies

    SET effective_to = v_now

    WHERE id = v_old_policy_id;


    INSERT INTO reservation_policies (
        reservable_window_days,
        request_frequency_days,
        confirmation_deadline_minutes,
        minimum_participants,
        opening_minute,
        closing_minute,
        slot_interval_minutes,
        effective_from,
        effective_to,
        created_by_user_id,
        idempotency_key,
        idempotency_payload_hash,
        is_published,
        late_withdrawal_minutes,
        group_recovery_deadline_minutes
    )

    SELECT
        old.reservable_window_days,
        old.request_frequency_days,
        60,
        v_default_minimum,
        old.opening_minute,
        old.closing_minute,
        old.slot_interval_minutes,
        v_now,
        NULL,
        old.created_by_user_id,
        'mvp2-group-resource-rules-v1',
        repeat('a', 64),
        true,
        old.late_withdrawal_minutes,
        0

    FROM reservation_policies old

    WHERE old.id = v_old_policy_id

    RETURNING id INTO v_new_policy_id;


    INSERT INTO reservation_policy_durations (
        policy_id,
        duration_minutes
    )

    SELECT
        v_new_policy_id,
        duration_minutes

    FROM reservation_policy_durations

    WHERE policy_id = v_old_policy_id

    ON CONFLICT DO NOTHING;


    INSERT INTO reservation_policy_resources (
        policy_id,
        resource_id
    )

    SELECT
        v_new_policy_id,
        resource_id

    FROM reservation_policy_resources

    WHERE policy_id = v_old_policy_id

    ON CONFLICT DO NOTHING;


    INSERT INTO reservation_policy_resources (
        policy_id,
        resource_id
    )

    SELECT
        v_new_policy_id,
        resource.id

    FROM resources resource

    WHERE resource.name IN (
        'Cancha 1, Centro Deportivo',
        'Cancha 2, Centro Deportivo',
        'Sala Multiuso, Centro Deportivo'
    )

    ON CONFLICT DO NOTHING;


    -- Conserva cualquier regla grupal versionada de la politica anterior.
    INSERT INTO reservation_policy_group_resources (
        policy_id,
        resource_id,
        minimum_participants
    )

    SELECT
        v_new_policy_id,
        old_rule.resource_id,
        old_rule.minimum_participants

    FROM reservation_policy_group_resources old_rule

    INNER JOIN reservation_policy_resources new_scope
        ON new_scope.policy_id = v_new_policy_id
       AND new_scope.resource_id = old_rule.resource_id

    WHERE old_rule.policy_id = v_old_policy_id

    ON CONFLICT (policy_id, resource_id)
    DO UPDATE SET
        minimum_participants = EXCLUDED.minimum_participants;


    -- Valor inicial MVP2 para Cancha 1, Cancha 2 y Sala Multiuso.
    INSERT INTO reservation_policy_group_resources (
        policy_id,
        resource_id,
        minimum_participants
    )

    SELECT
        v_new_policy_id,
        resource.id,
        v_default_minimum

    FROM resources resource

    WHERE resource.name IN (
        'Cancha 1, Centro Deportivo',
        'Cancha 2, Centro Deportivo',
        'Sala Multiuso, Centro Deportivo'
    )

    ON CONFLICT (policy_id, resource_id)
    DO UPDATE SET
        minimum_participants = EXCLUDED.minimum_participants;

END
$migration$;


COMMENT ON COLUMN reservation_policies.group_recovery_deadline_minutes
IS
    'Compatibilidad temporal. MVP2 14B unificara la participacion en confirmation_deadline_minutes.';


-- ==========================================================================
-- 6. VERIFICACION
-- ==========================================================================

DO $verification$

DECLARE

    v_policy_id integer;
    v_group_count integer;

BEGIN

    SELECT id

    INTO v_policy_id

    FROM reservation_policies

    WHERE idempotency_key = 'mvp2-group-resource-rules-v1';


    IF v_policy_id IS NULL THEN
        RAISE EXCEPTION
            'PG16_0010 no creo su politica prospectiva';
    END IF;


    IF NOT EXISTS (
        SELECT 1
        FROM reservation_policies
        WHERE id = v_policy_id
          AND confirmation_deadline_minutes = 60
    ) THEN
        RAISE EXCEPTION
            'el deadline vigente de PG16_0010 debe ser 60 minutos';
    END IF;


    SELECT COUNT(*)

    INTO v_group_count

    FROM reservation_policy_group_resources group_rule

    INNER JOIN resources resource
        ON resource.id = group_rule.resource_id

    WHERE group_rule.policy_id = v_policy_id
      AND resource.name IN (
          'Cancha 1, Centro Deportivo',
          'Cancha 2, Centro Deportivo',
          'Sala Multiuso, Centro Deportivo'
      )
      AND resource.reservation_mode = 'RESERVABLE'
      AND group_rule.minimum_participants = 10
      AND resource.capacity >= group_rule.minimum_participants;


    IF v_group_count <> 3 THEN
        RAISE EXCEPTION
            'PG16_0010 debe configurar los tres recursos grupales MVP2';
    END IF;


    IF EXISTS (
        SELECT 1
        FROM reservations
        WHERE group_capacity_snapshot IS NOT NULL
          AND group_minimum_participants_snapshot IS NULL
    ) THEN
        RAISE EXCEPTION
            'existe una reserva grupal sin snapshot de minimo';
    END IF;

END
$verification$;


COMMIT;
