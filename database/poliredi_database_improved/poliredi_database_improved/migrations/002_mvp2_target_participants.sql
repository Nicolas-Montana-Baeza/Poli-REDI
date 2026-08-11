-- ============================================================================
-- POLI-REDI | Azure SQL Database / SQL Server
-- Archivo: 002_mvp2_target_participants.sql
-- Propósito: agregar un objetivo editable de participantes para reservas grupales.
-- Reejecución: idempotente; no rellena reservas históricas.
-- ============================================================================

SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

IF OBJECT_ID(N'dbo.reservations', N'U') IS NULL
   OR OBJECT_ID(N'dbo.users', N'U') IS NULL
   OR OBJECT_ID(N'dbo.reservation_policies', N'U') IS NULL
   OR COL_LENGTH(N'dbo.reservations', N'group_capacity_snapshot') IS NULL
    THROW 52100, 'Preflight: ejecute y verifique primero la migracion 001.', 1;
GO

BEGIN TRY
    BEGIN TRANSACTION;

    IF COL_LENGTH(N'dbo.reservations', N'target_participants') IS NULL
        EXEC(N'ALTER TABLE dbo.reservations ADD target_participants INT NULL;');

    IF OBJECT_ID(N'dbo.ck_reservations_target_participants', N'C') IS NULL
        ALTER TABLE dbo.reservations WITH CHECK
        ADD CONSTRAINT ck_reservations_target_participants
        CHECK (target_participants IS NULL OR target_participants > 0);

    IF OBJECT_ID(N'dbo.reservation_target_audit', N'U') IS NULL
    BEGIN
        CREATE TABLE dbo.reservation_target_audit (
            id BIGINT IDENTITY(1,1) NOT NULL CONSTRAINT pk_reservation_target_audit PRIMARY KEY,
            reservation_id INT NOT NULL,
            actor_user_id INT NOT NULL,
            old_target_participants INT NULL,
            new_target_participants INT NULL,
            created_at DATETIME2(0) NOT NULL CONSTRAINT df_reservation_target_audit_created_at DEFAULT (SYSUTCDATETIME()),
            CONSTRAINT fk_reservation_target_audit_reservation FOREIGN KEY (reservation_id) REFERENCES dbo.reservations(id),
            CONSTRAINT fk_reservation_target_audit_actor FOREIGN KEY (actor_user_id) REFERENCES dbo.users(id),
            CONSTRAINT ck_reservation_target_audit_changed CHECK (
                (old_target_participants IS NULL AND new_target_participants IS NOT NULL)
                OR (old_target_participants IS NOT NULL AND new_target_participants IS NULL)
                OR old_target_participants <> new_target_participants
            )
        );
    END;

    -- Compatibilidad con una ejecución anterior de 002: NULL es un estado válido.
    IF EXISTS (
        SELECT 1
        FROM sys.columns
        WHERE object_id = OBJECT_ID(N'dbo.reservation_target_audit')
          AND name IN (N'old_target_participants', N'new_target_participants')
          AND is_nullable = 0
    )
    BEGIN
        ALTER TABLE dbo.reservation_target_audit ALTER COLUMN old_target_participants INT NULL;
        ALTER TABLE dbo.reservation_target_audit ALTER COLUMN new_target_participants INT NULL;
    END;

    IF OBJECT_ID(N'dbo.ck_reservation_target_audit_changed', N'C') IS NULL
        ALTER TABLE dbo.reservation_target_audit WITH CHECK
        ADD CONSTRAINT ck_reservation_target_audit_changed CHECK (
            (old_target_participants IS NULL AND new_target_participants IS NOT NULL)
            OR (old_target_participants IS NOT NULL AND new_target_participants IS NULL)
            OR old_target_participants <> new_target_participants
        );

    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    THROW;
END CATCH;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_target_validate
ON dbo.reservations
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted AS i
        INNER JOIN dbo.reservation_policies AS p ON p.id = i.policy_id
        WHERE i.target_participants IS NOT NULL
          AND (
              i.group_capacity_snapshot IS NULL
              OR i.target_participants < p.minimum_participants
              OR i.target_participants > i.group_capacity_snapshot
          )
    )
        THROW 51021, 'El objetivo de participantes no cumple el minimo o la capacidad congelada.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_target_audit_append_only
ON dbo.reservation_target_audit
INSTEAD OF UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    THROW 51022, 'La auditoria de objetivo es inmutable.', 1;
END;
GO

SELECT
    CONVERT(bit, CASE WHEN COL_LENGTH(N'dbo.reservations', N'target_participants') IS NOT NULL THEN 1 ELSE 0 END) AS target_column_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.ck_reservations_target_participants', N'C') IS NOT NULL THEN 1 ELSE 0 END) AS target_constraint_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservations_target_validate', N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS target_validation_trigger_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.reservation_target_audit', N'U') IS NOT NULL THEN 1 ELSE 0 END) AS target_audit_ok,
    CONVERT(bit, CASE WHEN NOT EXISTS (
        SELECT 1 FROM sys.columns
        WHERE object_id = OBJECT_ID(N'dbo.reservation_target_audit')
          AND name IN (N'old_target_participants', N'new_target_participants')
          AND is_nullable = 0
    ) THEN 1 ELSE 0 END) AS target_audit_nullable_endpoints_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.ck_reservation_target_audit_changed', N'C') IS NOT NULL THEN 1 ELSE 0 END) AS target_audit_changed_check_ok,
    CONVERT(bit, CASE WHEN OBJECT_ID(N'dbo.trg_reservation_target_audit_append_only', N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS target_audit_append_only_ok;
GO
