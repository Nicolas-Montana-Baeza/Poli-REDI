-- ============================================================================
-- POLI-REDI
-- PG16_0004_mvp2_group_participants.sql
--
-- MVP2 - Reservas grupales y participantes
--
-- Objetivos:
--   1. Incorporar participantes a reservas grupales.
--   2. Mantener las reglas variables en reservation_policies.
--   3. Mantener en PostgreSQL solo invariantes estructurales.
--   4. Evitar IDs de recursos hardcodeados.
--   5. Mantener compatibilidad con el MVP1 PostgreSQL estable.
--   6. Preparar el modelo para futuras infracciones/bloqueos sin implementarlos.
--
-- Filosofía:
--   PostgreSQL -> integridad y concurrencia.
--   Go         -> comportamiento y reglas de negocio.
--   Policies   -> parámetros configurables.
--
-- IMPORTANTE:
--   Esta migración NO crea todavía el sistema de infracciones.
--   Los retiros tardíos podrán auditarse y convertirse en infracciones
--   en una migración posterior.
-- ============================================================================

BEGIN;

SET TIME ZONE 'America/Santiago';


-- ============================================================================
-- 1. PARÁMETROS CONFIGURABLES PARA FLUJO GRUPAL
-- ============================================================================

-- Cantidad de minutos antes del inicio en que un retiro se considera tardío.
--
-- Ejemplo:
--   late_withdrawal_minutes = 60
--
-- Un participante que abandona dentro de los últimos 60 minutos puede
-- posteriormente generar una infracción.
--
-- Esta migración NO genera la infracción; solo deja disponible el parámetro.

ALTER TABLE reservation_policies
    ADD COLUMN IF NOT EXISTS late_withdrawal_minutes integer
        NOT NULL DEFAULT 60;


-- Minutos antes del inicio hasta los cuales el owner puede recuperar
-- participantes después de quedar bajo el mínimo.
--
-- 0 significa:
--   puede recuperar participantes hasta la hora exacta de inicio.

ALTER TABLE reservation_policies
    ADD COLUMN IF NOT EXISTS group_recovery_deadline_minutes integer
        NOT NULL DEFAULT 0;


DO $migration$
BEGIN

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_policies_late_withdrawal_minutes'
    ) THEN

        ALTER TABLE reservation_policies
            ADD CONSTRAINT ck_policies_late_withdrawal_minutes
            CHECK (late_withdrawal_minutes >= 0);

    END IF;


    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_policies_group_recovery_deadline'
    ) THEN

        ALTER TABLE reservation_policies
            ADD CONSTRAINT ck_policies_group_recovery_deadline
            CHECK (group_recovery_deadline_minutes >= 0);

    END IF;

END
$migration$;


-- ============================================================================
-- 2. EXTENSIÓN DE RESERVATIONS PARA RESERVAS GRUPALES
-- ============================================================================

-- Solo se almacena el hash del código.
-- El código real nunca debe persistirse en texto plano.

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS join_code_hash varchar(64);


-- Snapshot de la capacidad del recurso al momento de crear la reserva.
--
-- Esto evita que un cambio futuro de capacidad modifique retroactivamente
-- una reserva ya creada.

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS group_capacity_snapshot integer;


DO $migration$
BEGIN

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_reservations_join_code_hash'
    ) THEN

        ALTER TABLE reservations
            ADD CONSTRAINT ck_reservations_join_code_hash
            CHECK (
                join_code_hash IS NULL
                OR join_code_hash ~ '^[0-9a-f]{64}$'
            );

    END IF;


    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_reservations_group_capacity_snapshot'
    ) THEN

        ALTER TABLE reservations
            ADD CONSTRAINT ck_reservations_group_capacity_snapshot
            CHECK (
                group_capacity_snapshot IS NULL
                OR group_capacity_snapshot > 0
            );

    END IF;

END
$migration$;


CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_join_code_hash
    ON reservations (join_code_hash)
    WHERE join_code_hash IS NOT NULL;


-- ============================================================================
-- 3. RECURSOS QUE UTILIZAN FLUJO GRUPAL
-- ============================================================================

-- Una política puede permitir múltiples recursos, pero solo algunos requieren
-- flujo grupal.
--
-- Ejemplo actual:
--
--   Cancha 1     -> RESERVABLE + GROUP
--   Cancha 2     -> RESERVABLE + GROUP
--   Muro         -> OPEN_USE
--   Piscina      -> OPEN_USE
--   Gimnasio     -> OPEN_USE
--
-- La FK compuesta garantiza además que el recurso esté incluido dentro de
-- reservation_policy_resources.

CREATE TABLE IF NOT EXISTS reservation_policy_group_resources (

    policy_id integer NOT NULL,

    resource_id integer NOT NULL,

    CONSTRAINT pk_reservation_policy_group_resources
        PRIMARY KEY (policy_id, resource_id),

    CONSTRAINT fk_group_resources_policy_resource
        FOREIGN KEY (policy_id, resource_id)
        REFERENCES reservation_policy_resources(policy_id, resource_id)
        ON DELETE CASCADE

);


-- ============================================================================
-- 4. VALIDACIÓN ESTRUCTURAL DE RECURSOS GRUPALES
-- ============================================================================

-- No hardcodeamos IDs ni capacidades.
--
-- Solo protegemos dos invariantes:
--
--   - un recurso grupal debe ser RESERVABLE;
--   - su capacidad debe soportar el mínimo configurado en la política.

CREATE OR REPLACE FUNCTION validate_group_policy_resource()
RETURNS trigger
LANGUAGE plpgsql
AS $function$

DECLARE

    v_minimum integer;
    v_capacity integer;
    v_mode varchar(50);

BEGIN

    SELECT
        policy.minimum_participants,
        resource.capacity,
        resource.reservation_mode

    INTO
        v_minimum,
        v_capacity,
        v_mode

    FROM reservation_policies policy

    INNER JOIN resources resource
        ON resource.id = NEW.resource_id

    WHERE policy.id = NEW.policy_id;


    -- Las foreign keys resolverán normalmente este escenario.
    IF NOT FOUND THEN
        RETURN NEW;
    END IF;


    IF v_mode <> 'RESERVABLE' THEN

        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'un recurso grupal debe utilizar reservation_mode RESERVABLE';

    END IF;


    IF v_capacity IS NULL OR v_capacity < v_minimum THEN

        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'la capacidad del recurso grupal es inferior al minimo configurado';

    END IF;


    RETURN NEW;

END
$function$;


DROP TRIGGER IF EXISTS trg_validate_group_policy_resource
    ON reservation_policy_group_resources;


CREATE TRIGGER trg_validate_group_policy_resource

BEFORE INSERT OR UPDATE
ON reservation_policy_group_resources

FOR EACH ROW
EXECUTE FUNCTION validate_group_policy_resource();


-- ============================================================================
-- 5. PARTICIPANTES
-- ============================================================================

CREATE TABLE IF NOT EXISTS participants (

    id integer
        GENERATED BY DEFAULT AS IDENTITY
        PRIMARY KEY,

    reservation_id integer NOT NULL
        REFERENCES reservations(id)
        ON DELETE CASCADE,

    user_id integer NOT NULL
        REFERENCES users(id)
        ON DELETE RESTRICT,

    status varchar(30)
        NOT NULL DEFAULT 'PENDING',

    is_owner boolean
        NOT NULL DEFAULT false,

    confirmed_at timestamptz,

    created_at timestamptz
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at timestamptz
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_participants_reservation_user
        UNIQUE (reservation_id, user_id),

    CONSTRAINT ck_participants_status
        CHECK (
            status IN (
                'PENDING',
                'CONFIRMED',
                'REJECTED',
                'CANCELLED'
            )
        )

);


-- Solo puede existir un owner por reserva.

CREATE UNIQUE INDEX IF NOT EXISTS uq_participants_owner
    ON participants (reservation_id)
    WHERE is_owner;


CREATE INDEX IF NOT EXISTS ix_participants_reservation
    ON participants (reservation_id);


CREATE INDEX IF NOT EXISTS ix_participants_user
    ON participants (user_id);


CREATE INDEX IF NOT EXISTS ix_participants_reservation_status
    ON participants (reservation_id, status);


-- Reutilizamos la función set_updated_at() creada en PG16_0003.

DROP TRIGGER IF EXISTS trg_participants_updated_at
    ON participants;


CREATE TRIGGER trg_participants_updated_at

BEFORE UPDATE
ON participants

FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- ============================================================================
-- 6. AUDITORÍA DEL FLUJO DE PARTICIPANTES
-- ============================================================================

-- Esta tabla será especialmente importante cuando incorporemos:
--
--   LATE_PARTICIPANT_WITHDRAWAL
--   GROUP_BELOW_MINIMUM_AT_START
--   NO_SHOW
--   infracciones
--   bloqueos
--
-- Por ahora registra los cambios del flujo de participantes.

CREATE TABLE IF NOT EXISTS reservation_participant_audit (

    id bigint
        GENERATED BY DEFAULT AS IDENTITY
        PRIMARY KEY,

    reservation_id integer NOT NULL
        REFERENCES reservations(id)
        ON DELETE RESTRICT,

    actor_user_id integer NOT NULL
        REFERENCES users(id)
        ON DELETE RESTRICT,

    participant_user_id integer NOT NULL
        REFERENCES users(id)
        ON DELETE RESTRICT,

    action varchar(40) NOT NULL,

    previous_status varchar(30),

    new_status varchar(30) NOT NULL,

    previous_reservation_status varchar(30),

    new_reservation_status varchar(30) NOT NULL,

    created_at timestamptz
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT ck_participant_audit_action
        CHECK (length(trim(action)) > 0)

);


CREATE INDEX IF NOT EXISTS ix_participant_audit_reservation
    ON reservation_participant_audit (reservation_id, created_at DESC);


CREATE INDEX IF NOT EXISTS ix_participant_audit_participant
    ON reservation_participant_audit (participant_user_id, created_at DESC);


CREATE INDEX IF NOT EXISTS ix_participant_audit_actor
    ON reservation_participant_audit (actor_user_id, created_at DESC);


-- ============================================================================
-- 7. MURO DE ESCALADA PASA A OPEN_USE
-- ============================================================================

-- Se identifica mediante nombre, NO mediante resource_id.

UPDATE resources

SET
    reservation_mode = 'OPEN_USE',
    updated_at = CURRENT_TIMESTAMP

WHERE name = 'Muro Escalada, Centro Deportivo'
  AND reservation_mode <> 'OPEN_USE';


-- ============================================================================
-- 8. NUEVA VERSIÓN DE POLÍTICA PARA MVP2
-- ============================================================================

-- No modificamos retroactivamente la política del MVP1.
--
-- Creamos una nueva versión prospectiva.
--
-- Valores iniciales:
--
--   minimum_participants                = 10
--   late_withdrawal_minutes             = 60
--   group_recovery_deadline_minutes     = 0
--
-- Estos valores son configurables y podrán cambiar mediante nuevas versiones
-- de políticas.

DO $migration$

DECLARE

    v_now timestamptz := CURRENT_TIMESTAMP;

    v_old_policy_id integer;

    v_new_policy_id integer;

    v_group_resource_count integer;

    v_minimum_participants integer := 10;

BEGIN

    -- ------------------------------------------------------------------------
    -- La política MVP2 ya fue creada.
    -- ------------------------------------------------------------------------

    IF EXISTS (
        SELECT 1
        FROM reservation_policies
        WHERE idempotency_key = 'mvp2-group-participants-v1'
    ) THEN

        RETURN;

    END IF;


    -- ------------------------------------------------------------------------
    -- Buscamos la política actualmente vigente.
    -- ------------------------------------------------------------------------

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
            'no existe una politica vigente desde la cual crear MVP2';

    END IF;


    -- ------------------------------------------------------------------------
    -- Validamos las dos canchas grupales actuales.
    --
    -- NO dependemos de sus IDs.
    -- ------------------------------------------------------------------------

    SELECT COUNT(*)

    INTO v_group_resource_count

    FROM resources

    WHERE name IN (
        'Cancha 1, Centro Deportivo',
        'Cancha 2, Centro Deportivo'
    )

      AND is_active
      AND reservation_mode = 'RESERVABLE'
      AND capacity IS NOT NULL
      AND capacity >= v_minimum_participants;


    IF v_group_resource_count <> 2 THEN

        RAISE EXCEPTION
            'no se encontraron las dos canchas grupales activas con capacidad suficiente';

    END IF;


    -- ------------------------------------------------------------------------
    -- Cerramos la política anterior.
    --
    -- Esto libera la restricción uq_policies_current.
    -- ------------------------------------------------------------------------

    UPDATE reservation_policies

    SET effective_to = v_now

    WHERE id = v_old_policy_id;


    -- ------------------------------------------------------------------------
    -- Creamos la nueva política MVP2.
    --
    -- Se conservan:
    --   ventana reservable
    --   frecuencia
    --   deadline
    --   horario
    --   intervalos
    --   creador
    --
    -- Solo se introduce la semántica grupal.
    -- ------------------------------------------------------------------------

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

        old.confirmation_deadline_minutes,

        v_minimum_participants,

        old.opening_minute,

        old.closing_minute,

        old.slot_interval_minutes,

        v_now,

        NULL,

        old.created_by_user_id,

        'mvp2-group-participants-v1',

        repeat('2', 64),

        true,

        60,

        0

    FROM reservation_policies old

    WHERE old.id = v_old_policy_id

    RETURNING id INTO v_new_policy_id;


    -- ------------------------------------------------------------------------
    -- Copiamos duraciones permitidas.
    -- ------------------------------------------------------------------------

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


    -- ------------------------------------------------------------------------
    -- Copiamos recursos permitidos por la política anterior.
    -- ------------------------------------------------------------------------

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


    -- ------------------------------------------------------------------------
    -- Garantizamos explícitamente que las dos canchas estén permitidas.
    -- ------------------------------------------------------------------------

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
        'Cancha 2, Centro Deportivo'
    )

    ON CONFLICT DO NOTHING;


    -- ------------------------------------------------------------------------
    -- Marcamos exclusivamente Cancha 1 y Cancha 2 como flujo grupal.
    -- ------------------------------------------------------------------------

    INSERT INTO reservation_policy_group_resources (
        policy_id,
        resource_id
    )

    SELECT
        v_new_policy_id,
        resource.id

    FROM resources resource

    WHERE resource.name IN (
        'Cancha 1, Centro Deportivo',
        'Cancha 2, Centro Deportivo'
    )

      AND resource.is_active
      AND resource.reservation_mode = 'RESERVABLE'

    ON CONFLICT DO NOTHING;

END
$migration$;


-- ============================================================================
-- 9. REPARACIÓN IDEMPOTENTE
-- ============================================================================

-- Si la migración fue ejecutada parcialmente anteriormente, garantizamos que
-- las canchas existan en el scope de la política MVP2.

INSERT INTO reservation_policy_resources (
    policy_id,
    resource_id
)

SELECT
    policy.id,
    resource.id

FROM reservation_policies policy

CROSS JOIN resources resource

WHERE policy.idempotency_key = 'mvp2-group-participants-v1'

  AND resource.name IN (
      'Cancha 1, Centro Deportivo',
      'Cancha 2, Centro Deportivo'
  )

ON CONFLICT DO NOTHING;


INSERT INTO reservation_policy_group_resources (
    policy_id,
    resource_id
)

SELECT
    policy.id,
    resource.id

FROM reservation_policies policy

CROSS JOIN resources resource

WHERE policy.idempotency_key = 'mvp2-group-participants-v1'

  AND resource.name IN (
      'Cancha 1, Centro Deportivo',
      'Cancha 2, Centro Deportivo'
  )

  AND resource.is_active

  AND resource.reservation_mode = 'RESERVABLE'

ON CONFLICT DO NOTHING;


-- ============================================================================
-- 10. VERIFICACIÓN DE CONFIGURACIÓN
-- ============================================================================

DO $verification$

DECLARE

    v_policy_id integer;

    v_group_count integer;

BEGIN

    SELECT id

    INTO v_policy_id

    FROM reservation_policies

    WHERE idempotency_key = 'mvp2-group-participants-v1';


    IF v_policy_id IS NULL THEN

        RAISE EXCEPTION
            'la politica MVP2 de participantes no fue creada';

    END IF;


    SELECT COUNT(*)

    INTO v_group_count

    FROM reservation_policy_group_resources group_resource

    WHERE group_resource.policy_id = v_policy_id;


    IF v_group_count <> 2 THEN

        RAISE EXCEPTION
            'la politica MVP2 debe contener exactamente dos recursos grupales';

    END IF;


    IF EXISTS (

        SELECT 1

        FROM reservation_policy_group_resources group_resource

        INNER JOIN reservation_policies policy
            ON policy.id = group_resource.policy_id

        INNER JOIN resources resource
            ON resource.id = group_resource.resource_id

        WHERE group_resource.policy_id = v_policy_id

          AND (
              resource.reservation_mode <> 'RESERVABLE'
              OR resource.capacity IS NULL
              OR resource.capacity < policy.minimum_participants
          )

    ) THEN

        RAISE EXCEPTION
            'existe un recurso grupal incompatible con la politica';

    END IF;


    IF EXISTS (

        SELECT 1

        FROM resources

        WHERE name = 'Muro Escalada, Centro Deportivo'

          AND reservation_mode <> 'OPEN_USE'

    ) THEN

        RAISE EXCEPTION
            'Muro Escalada debe utilizar OPEN_USE';

    END IF;

END
$verification$;


-- ============================================================================
-- 11. PERMISOS
-- ============================================================================

GRANT SELECT, INSERT, UPDATE, DELETE
ON participants
TO poliredi_app;


GRANT SELECT, INSERT, UPDATE, DELETE
ON reservation_participant_audit
TO poliredi_app;


GRANT SELECT, INSERT, UPDATE, DELETE
ON reservation_policy_group_resources
TO poliredi_app;


GRANT USAGE, SELECT
ON ALL SEQUENCES IN SCHEMA public
TO poliredi_app;


GRANT EXECUTE
ON FUNCTION validate_group_policy_resource()
TO poliredi_app;


-- ============================================================================
-- 12. FIN
-- ============================================================================

COMMIT;
