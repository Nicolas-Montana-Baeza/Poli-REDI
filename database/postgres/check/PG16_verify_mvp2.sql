\set ON_ERROR_STOP on

BEGIN;

SET LOCAL TIME ZONE 'America/Santiago';

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
            'no existe la politica final de reglas grupales MVP2';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM reservation_policies
        WHERE id = v_policy_id
          AND is_published
          AND confirmation_deadline_minutes = 60
          AND group_recovery_deadline_minutes = 0
    ) THEN
        RAISE EXCEPTION
            'la politica MVP2 no conserva el deadline unico de 60 minutos';
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
      AND resource.is_active
      AND resource.reservation_mode = 'RESERVABLE'
      AND group_rule.minimum_participants = 10
      AND resource.capacity >= group_rule.minimum_participants;

    IF v_group_count <> 3 THEN
        RAISE EXCEPTION
            'la politica MVP2 no contiene los tres recursos grupales esperados';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM reservation_policy_group_resources
        WHERE minimum_participants IS NULL
           OR minimum_participants <= 0
    ) THEN
        RAISE EXCEPTION
            'existe una regla grupal sin minimo valido';
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

    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'uq_reservations_join_code_hash'
    ) THEN
        RAISE EXCEPTION
            'falta la unicidad del hash de codigo de invitacion';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'participants'::regclass
          AND conname = 'uq_participants_reservation_user'
    ) THEN
        RAISE EXCEPTION
            'falta la unicidad de participante por reserva';
    END IF;
END
$verification$;

SELECT
    policy.confirmation_deadline_minutes AS deadline,
    resource.name,
    group_rule.minimum_participants AS minimum,
    resource.capacity,
    resource.reservation_mode
FROM reservation_policies policy
INNER JOIN reservation_policy_group_resources group_rule
    ON group_rule.policy_id = policy.id
INNER JOIN resources resource
    ON resource.id = group_rule.resource_id
WHERE policy.idempotency_key = 'mvp2-group-resource-rules-v1'
ORDER BY resource.name;

ROLLBACK;
