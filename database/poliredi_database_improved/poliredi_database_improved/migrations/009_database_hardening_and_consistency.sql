-- ============================================================================
-- POLI-REDI | Azure SQL Database / SQL Server
-- Archivo: 009_database_hardening_and_consistency.sql
-- Propósito: alinear bases ya migradas con el esquema canónico mejorado.
-- Reejecución: idempotente y sin reclasificar reservas históricas.
-- ============================================================================

SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET ANSI_PADDING ON;
SET ANSI_WARNINGS ON;
SET ARITHABORT ON;
SET CONCAT_NULL_YIELDS_NULL ON;
SET NUMERIC_ROUNDABORT OFF;
SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

IF OBJECT_ID(N'dbo.users', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservations', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_policies', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_policy_durations', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_target_audit', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_participant_audit', N'U') IS NULL
   OR OBJECT_ID(N'dbo.audit_logs', N'U') IS NULL
   OR OBJECT_ID(N'dbo.availability_blocks', N'U') IS NULL
   OR OBJECT_ID(N'dbo.scheduled_activities', N'U') IS NULL
   OR OBJECT_ID(N'dbo.participants', N'U') IS NULL
   OR OBJECT_ID(N'dbo.resources', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_policy_group_resources', N'U') IS NULL
    THROW 59000, 'Preflight: faltan objetos requeridos; ejecute y verifique 001-008.', 1;

IF COL_LENGTH(N'dbo.reservations', N'updated_at') IS NULL
   OR COL_LENGTH(N'dbo.reservations', N'group_capacity_snapshot') IS NULL
   OR COL_LENGTH(N'dbo.resources', N'capacity') IS NULL
    THROW 59001, 'Preflight: faltan columnas requeridas por el endurecimiento 009.', 1;
GO

-- 1) La auditoría debe representar cambios NULL -> valor y valor -> NULL.
BEGIN TRY
    BEGIN TRANSACTION;

    IF EXISTS (
        SELECT 1 FROM sys.columns
        WHERE object_id = OBJECT_ID(N'dbo.reservation_target_audit')
          AND name = N'old_target_participants' AND is_nullable = 0
    )
        ALTER TABLE dbo.reservation_target_audit ALTER COLUMN old_target_participants INT NULL;

    IF EXISTS (
        SELECT 1 FROM sys.columns
        WHERE object_id = OBJECT_ID(N'dbo.reservation_target_audit')
          AND name = N'new_target_participants' AND is_nullable = 0
    )
        ALTER TABLE dbo.reservation_target_audit ALTER COLUMN new_target_participants INT NULL;

    IF OBJECT_ID(N'dbo.ck_reservation_target_audit_changed', N'C') IS NULL
        ALTER TABLE dbo.reservation_target_audit WITH CHECK
        ADD CONSTRAINT ck_reservation_target_audit_changed CHECK (
            (old_target_participants IS NULL AND new_target_participants IS NOT NULL)
            OR (old_target_participants IS NOT NULL AND new_target_participants IS NULL)
            OR old_target_participants <> new_target_participants
        );

    -- Completa duraciones por política, no solo cuando la tabla completa está vacía.
    INSERT INTO dbo.reservation_policy_durations (policy_id, duration_minutes)
    SELECT p.id, d.duration_minutes
    FROM dbo.reservation_policies AS p
    CROSS JOIN (VALUES (30), (60), (90), (120), (150), (180)) AS d(duration_minutes)
    WHERE NOT EXISTS (
        SELECT 1 FROM dbo.reservation_policy_durations AS existing
        WHERE existing.policy_id = p.id
          AND existing.duration_minutes = d.duration_minutes
    );

    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    THROW;
END CATCH;
GO

-- 2) El trigger canónico ya cubre estados PENDING/CONFIRMED.
IF OBJECT_ID(N'dbo.trg_reservations_pending_conflicts', N'TR') IS NOT NULL
    DROP TRIGGER dbo.trg_reservations_pending_conflicts;
IF OBJECT_ID(N'dbo.trg_blocks_pending_conflicts', N'TR') IS NOT NULL
    DROP TRIGGER dbo.trg_blocks_pending_conflicts;
IF OBJECT_ID(N'dbo.trg_scheduled_activities_pending_conflicts', N'TR') IS NOT NULL
    DROP TRIGGER dbo.trg_scheduled_activities_pending_conflicts;
GO

-- 3) Refuerza el CHECK de formato y valida también el dígito verificador.
BEGIN TRY
    BEGIN TRANSACTION;

    IF EXISTS (
        SELECT 1 FROM dbo.users
        WHERE rut IS NOT NULL AND (
            LEN(rut) NOT BETWEEN 9 AND 10
            OR LEN(rut)-LEN(REPLACE(rut,N'-',N''))<>1
            OR CHARINDEX(N'-',rut)<>LEN(rut)-1
            OR LEFT(rut,LEN(rut)-2) LIKE N'%[^0-9]%'
            OR RIGHT(rut,1) NOT LIKE N'[0-9K]'
            OR rut<>UPPER(rut)
            OR rut LIKE N'%[ .]%'
        )
    )
        THROW 59001, 'Hay RUT no canonicos; corrija los datos antes de continuar.', 1;

    IF EXISTS (
        SELECT 1
        FROM dbo.users AS u
        CROSS APPLY (SELECT LEFT(u.rut,LEN(u.rut)-2),RIGHT(u.rut,1)) AS parts(body,verifier)
        CROSS APPLY (
            SELECT SUM(CONVERT(INT,SUBSTRING(parts.body,LEN(parts.body)-n.n+1,1))*(2+((n.n-1)%6))) AS weighted_sum
            FROM (VALUES(1),(2),(3),(4),(5),(6),(7),(8)) AS n(n)
            WHERE n.n<=LEN(parts.body)
        ) AS calculation
        CROSS APPLY (
            SELECT CASE 11-(calculation.weighted_sum%11)
                WHEN 11 THEN N'0' WHEN 10 THEN N'K'
                ELSE CONVERT(NVARCHAR(1),11-(calculation.weighted_sum%11)) END
        ) AS validation(expected_verifier)
        WHERE u.rut IS NOT NULL AND validation.expected_verifier<>parts.verifier
    )
        THROW 59004, 'Hay RUT con digito verificador invalido; corrija los datos antes de continuar.', 1;

    IF OBJECT_ID(N'dbo.ck_users_rut_basic_format', N'C') IS NOT NULL
        ALTER TABLE dbo.users DROP CONSTRAINT ck_users_rut_basic_format;

    ALTER TABLE dbo.users WITH CHECK
    ADD CONSTRAINT ck_users_rut_basic_format CHECK (
        rut IS NULL OR (
            LEN(rut) BETWEEN 9 AND 10
            AND LEN(rut)-LEN(REPLACE(rut,N'-',N''))=1
            AND CHARINDEX(N'-',rut)=LEN(rut)-1
            AND LEFT(rut,LEN(rut)-2) NOT LIKE N'%[^0-9]%'
            AND RIGHT(rut,1) LIKE N'[0-9K]'
            AND rut=UPPER(rut)
            AND rut NOT LIKE N'%[ .]%'
        )
    );

    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    THROW;
END CATCH;
GO

CREATE OR ALTER TRIGGER dbo.trg_users_rut_validate
ON dbo.users
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1 FROM inserted AS i
        WHERE i.rut IS NOT NULL
          AND (
              LEN(i.rut) NOT BETWEEN 9 AND 10
              OR LEN(i.rut) - LEN(REPLACE(i.rut, N'-', N'')) <> 1
              OR CHARINDEX(N'-', i.rut) <> LEN(i.rut) - 1
              OR LEFT(i.rut, LEN(i.rut) - 2) LIKE N'%[^0-9]%'
              OR RIGHT(i.rut, 1) NOT LIKE N'[0-9K]'
              OR i.rut <> UPPER(i.rut)
              OR i.rut LIKE N'%[ .]%'
          )
    )
        THROW 55008, 'El RUT debe almacenarse en formato canonico sin puntos: 12345678-9.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted AS i
        CROSS APPLY (SELECT LEFT(i.rut, LEN(i.rut) - 2), RIGHT(i.rut, 1)) AS parts(body, verifier)
        CROSS APPLY (
            SELECT SUM(CONVERT(INT, SUBSTRING(parts.body, LEN(parts.body)-n.n+1, 1)) * (2 + ((n.n-1) % 6))) AS weighted_sum
            FROM (VALUES (1),(2),(3),(4),(5),(6),(7),(8)) AS n(n)
            WHERE n.n <= LEN(parts.body)
        ) AS calculation
        CROSS APPLY (
            SELECT CASE 11-(calculation.weighted_sum % 11)
                WHEN 11 THEN N'0' WHEN 10 THEN N'K'
                ELSE CONVERT(NVARCHAR(1), 11-(calculation.weighted_sum % 11)) END
        ) AS validation(expected_verifier)
        WHERE i.rut IS NOT NULL AND validation.expected_verifier <> parts.verifier
    )
        THROW 55009, 'El digito verificador del RUT es invalido.', 1;
END;
GO

-- 4) Endurece inmutabilidad de políticas y auditorías.
CREATE OR ALTER TRIGGER dbo.trg_reservation_policies_immutable
ON dbo.reservation_policies
AFTER UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS(SELECT 1 FROM deleted d LEFT JOIN inserted i ON i.id=d.id WHERE i.id IS NULL)
        THROW 51011, 'Las versiones de politica utilizadas no se pueden eliminar.', 1;
    IF EXISTS(
        SELECT 1 FROM deleted d INNER JOIN inserted i ON i.id=d.id
        WHERE i.reservable_window_days<>d.reservable_window_days
           OR i.request_frequency_days<>d.request_frequency_days
           OR i.confirmation_deadline_minutes<>d.confirmation_deadline_minutes
           OR i.minimum_participants<>d.minimum_participants
           OR i.opening_minute<>d.opening_minute
           OR i.closing_minute<>d.closing_minute
           OR i.slot_interval_minutes<>d.slot_interval_minutes
           OR i.effective_from<>d.effective_from
           OR (d.effective_to IS NOT NULL AND (i.effective_to IS NULL OR i.effective_to<>d.effective_to))
           OR ISNULL(i.created_by_user_id,-1)<>ISNULL(d.created_by_user_id,-1)
           OR i.created_at<>d.created_at
           OR ISNULL(i.idempotency_key,N'')<>ISNULL(d.idempotency_key,N'')
           OR ISNULL(i.idempotency_payload_hash,'')<>ISNULL(d.idempotency_payload_hash,'')
           OR (i.is_published<>d.is_published AND NOT(d.is_published=0 AND i.is_published=1))
    ) THROW 51012, 'Una politica solo puede cerrarse una vez; el resto de sus datos es inmutable.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_group_snapshot_validate
ON dbo.reservations
AFTER INSERT
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted AS i
        INNER JOIN dbo.reservation_policy_group_resources AS g
          ON g.policy_id=i.policy_id AND g.resource_id=i.resource_id
        INNER JOIN dbo.resources AS r ON r.id=i.resource_id
        WHERE i.group_capacity_snapshot IS NULL
           OR i.group_capacity_snapshot<>r.capacity
    )
        THROW 59005, 'Una reserva grupal debe congelar la capacidad vigente del recurso.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted AS i
        WHERE i.group_capacity_snapshot IS NOT NULL
          AND NOT EXISTS (
              SELECT 1 FROM dbo.reservation_policy_group_resources AS g
              WHERE g.policy_id=i.policy_id AND g.resource_id=i.resource_id
          )
    )
        THROW 59006, 'Solo las reservas grupales pueden registrar un snapshot de capacidad.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_participant_audit_append_only
ON dbo.reservation_participant_audit
INSTEAD OF UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    THROW 59003, 'La auditoria de participantes es inmutable.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_updated_at
ON dbo.reservations
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF TRIGGER_NESTLEVEL()>1 RETURN;
    BEGIN TRY
        EXEC sys.sp_set_session_context @key=N'poliredi.skip_reservation_audit',@value=1;
        UPDATE target SET updated_at=SYSUTCDATETIME()
        FROM dbo.reservations target INNER JOIN inserted i ON i.id=target.id;
        EXEC sys.sp_set_session_context @key=N'poliredi.skip_reservation_audit',@value=NULL;
    END TRY
    BEGIN CATCH
        EXEC sys.sp_set_session_context @key=N'poliredi.skip_reservation_audit',@value=NULL;
        THROW;
    END CATCH;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_audit
ON dbo.reservations
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    IF TRY_CONVERT(bit,SESSION_CONTEXT(N'poliredi.skip_reservation_audit'))=1 RETURN;
    INSERT INTO dbo.audit_logs(user_id,action,entity_type,entity_id,description)
    SELECT COALESCE(i.user_id,d.user_id),
           CASE WHEN i.id IS NOT NULL AND d.id IS NULL THEN N'RESERVATION_CREATED'
                WHEN i.id IS NOT NULL AND d.id IS NOT NULL THEN N'RESERVATION_UPDATED'
                ELSE N'RESERVATION_DELETED' END,
           N'reservations',COALESCE(i.id,d.id),N'Cambio registrado sobre una reserva'
    FROM inserted i FULL OUTER JOIN deleted d ON d.id=i.id;
END;
GO

-- 5) Índices alineados con filtros y comprobaciones de solape.
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.reservations') AND name=N'idx_reservations_resource_status_start')
    CREATE INDEX idx_reservations_resource_status_start
    ON dbo.reservations(resource_id, status, start_time)
    INCLUDE(duration_minutes, user_id, policy_id);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.reservations') AND name=N'idx_reservations_user_status_start')
    CREATE INDEX idx_reservations_user_status_start
    ON dbo.reservations(user_id, status, start_time)
    INCLUDE(duration_minutes, resource_id, policy_id, created_at);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.participants') AND name=N'idx_participants_user_status_reservation')
    CREATE INDEX idx_participants_user_status_reservation
    ON dbo.participants(user_id, status, reservation_id);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.availability_blocks') AND name=N'idx_availability_blocks_resource_active_start')
    CREATE INDEX idx_availability_blocks_resource_active_start
    ON dbo.availability_blocks(resource_id, is_active, start_time)
    INCLUDE(end_time);

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.scheduled_activities') AND name=N'idx_scheduled_activities_resource_active_start')
    CREATE INDEX idx_scheduled_activities_resource_active_start
    ON dbo.scheduled_activities(resource_id, is_active, start_time)
    INCLUDE(end_time);

IF OBJECT_ID(N'dbo.workshop_enrollments', N'U') IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.workshop_enrollments') AND name=N'idx_workshop_enrollments_user_status')
    CREATE INDEX idx_workshop_enrollments_user_status
    ON dbo.workshop_enrollments(user_id, status, workshop_id);

IF OBJECT_ID(N'dbo.workshop_enrollments', N'U') IS NOT NULL
   AND EXISTS (
       SELECT workshop_id,user_id FROM dbo.workshop_enrollments
       WHERE status=N'CONFIRMED' GROUP BY workshop_id,user_id HAVING COUNT(*)>1
   )
    THROW 59002, 'Hay inscripciones CONFIRMED duplicadas para un mismo usuario y taller.', 1;

IF OBJECT_ID(N'dbo.workshop_enrollments', N'U') IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.workshop_enrollments') AND name=N'uq_workshop_enrollments_confirmed')
    CREATE UNIQUE INDEX uq_workshop_enrollments_confirmed
    ON dbo.workshop_enrollments(workshop_id,user_id) WHERE status=N'CONFIRMED';
GO

SELECT
    CONVERT(bit, CASE WHEN NOT EXISTS (
        SELECT 1 FROM sys.columns WHERE object_id=OBJECT_ID(N'dbo.reservation_target_audit')
        AND name IN (N'old_target_participants', N'new_target_participants') AND is_nullable=0
    ) THEN 1 ELSE 0 END) AS target_audit_nullable_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.ck_reservation_target_audit_changed', N'C') IS NOT NULL THEN 1 ELSE 0 END) AS target_audit_changed_check_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.ck_users_rut_basic_format', N'C') IS NOT NULL THEN 1 ELSE 0 END) AS rut_format_check_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservation_participant_audit_append_only', N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS participant_audit_append_only_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservations_group_snapshot_validate', N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS group_snapshot_validation_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_users_rut_validate', N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS rut_validation_trigger_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservations_pending_conflicts', N'TR') IS NULL
                       AND OBJECT_ID(N'dbo.trg_blocks_pending_conflicts', N'TR') IS NULL
                       AND OBJECT_ID(N'dbo.trg_scheduled_activities_pending_conflicts', N'TR') IS NULL THEN 1 ELSE 0 END) AS redundant_triggers_removed_ok,
    CONVERT(bit, CASE WHEN NOT EXISTS (
        SELECT 1 FROM dbo.reservation_policies AS p
        CROSS JOIN (VALUES (30),(60),(90),(120),(150),(180)) AS d(duration_minutes)
        WHERE NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_durations AS x WHERE x.policy_id=p.id AND x.duration_minutes=d.duration_minutes)
    ) THEN 1 ELSE 0 END) AS policy_durations_complete_ok,
    CONVERT(bit, CASE WHEN (SELECT COUNT(*) FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.reservations') AND name IN (N'idx_reservations_resource_status_start',N'idx_reservations_user_status_start'))=2 THEN 1 ELSE 0 END) AS reservation_indexes_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservations_audit',N'TR') IS NOT NULL AND OBJECT_DEFINITION(OBJECT_ID(N'dbo.trg_reservations_audit')) LIKE N'%skip_reservation_audit%' THEN 1 ELSE 0 END) AS reservation_audit_deduplicated_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.workshop_enrollments',N'U') IS NULL OR EXISTS(SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID(N'dbo.workshop_enrollments') AND name=N'uq_workshop_enrollments_confirmed' AND is_unique=1 AND has_filter=1) THEN 1 ELSE 0 END) AS workshop_confirmed_unique_ok;
GO
