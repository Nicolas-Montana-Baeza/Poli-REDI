BEGIN;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_ci ON users (lower(email));
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_rut_present ON users (rut) WHERE rut IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_entra_identity
    ON users (tenant_id, entra_oid)
    WHERE tenant_id IS NOT NULL AND entra_oid IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_policies_current
    ON reservation_policies ((true))
    WHERE is_published AND effective_to IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_policies_idempotency
    ON reservation_policies (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_resources_active ON resources (is_active, reservation_mode);
CREATE INDEX IF NOT EXISTS ix_reservations_user_start ON reservations (user_id, start_time DESC);
CREATE INDEX IF NOT EXISTS ix_reservations_resource_start ON reservations (resource_id, start_time);
CREATE INDEX IF NOT EXISTS ix_blocks_resource_time
    ON availability_blocks USING gist (resource_id, tstzrange(start_time, end_time, '[)'))
    WHERE is_active;

DO $migration$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ex_reservations_resource_overlap') THEN
        ALTER TABLE reservations
            ADD CONSTRAINT ex_reservations_resource_overlap
            EXCLUDE USING gist (
                resource_id WITH =,
                tstzrange(start_time, end_time, '[)') WITH &&
            )
            WHERE (
                status IN ('PENDING', 'CONFIRMED')
                AND reservation_mode_snapshot = 'RESERVABLE'
            );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ex_reservations_user_overlap') THEN
        ALTER TABLE reservations
            ADD CONSTRAINT ex_reservations_user_overlap
            EXCLUDE USING gist (
                user_id WITH =,
                tstzrange(start_time, end_time, '[)') WITH &&
            )
            WHERE (status IN ('PENDING', 'CONFIRMED'));
    END IF;
END
$migration$;

COMMIT;
