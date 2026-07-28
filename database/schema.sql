-- ============================================================
-- POLI-REDI - DDL BASE DE DATOS
-- Azure SQL Database / SQL Server T-SQL
-- ============================================================
-- Canonical target schema for local, test, and CI environments.
-- The migration history in database/migrations is reflected here as
-- a readable, idempotent target state:
--   001_mvp2_group_participants.sql
--   002_mvp2_target_participants.sql
--   003_open_use_frequency_scope.sql
--   004_group_flow_completion.sql
--   005_rut_integrity_and_admin_exemption.sql
--   006_workshop_occurrences.sql
-- ============================================================

SET ANSI_NULLS ON;
GO
SET QUOTED_IDENTIFIER ON;
GO

-- ============================================================
-- TABLE: venues
-- Representa una sede o recinto institucional.
-- ============================================================

IF OBJECT_ID('dbo.venues', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.venues (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_venues PRIMARY KEY,
        name NVARCHAR(150) NOT NULL,
        address_line NVARCHAR(200) NULL,
        commune NVARCHAR(100) NULL,
        city NVARCHAR(100) NULL,
        region NVARCHAR(100) NULL,
        country NVARCHAR(100) NOT NULL CONSTRAINT df_venues_country DEFAULT ('Chile'),
        latitude DECIMAL(10, 7) NULL,
        longitude DECIMAL(10, 7) NULL,
        is_active BIT NOT NULL CONSTRAINT df_venues_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_venues_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_venues_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT uq_venues_name UNIQUE (name),
        CONSTRAINT ck_venues_latitude CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
        CONSTRAINT ck_venues_longitude CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
    );
END;
GO

-- ============================================================
-- TABLE: users
-- Usuarios autenticados mediante Microsoft Entra ID.
-- ============================================================

IF OBJECT_ID('dbo.users', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.users (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_users PRIMARY KEY,
        email NVARCHAR(150) NOT NULL,
        full_name NVARCHAR(150) NOT NULL,
        rut NVARCHAR(12) NULL,
        is_admin BIT NOT NULL CONSTRAINT df_users_is_admin DEFAULT (0),
        is_blocked BIT NOT NULL CONSTRAINT df_users_is_blocked DEFAULT (0),
        entra_oid NVARCHAR(100) NULL,
        tenant_id NVARCHAR(100) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_users_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_users_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT uq_users_email UNIQUE (email),
        CONSTRAINT ck_users_email_format CHECK (email LIKE '%_@_%._%'),
        CONSTRAINT ck_users_rut_basic_format CHECK (
            rut IS NULL OR (
                LEN(rut) BETWEEN 9 AND 10
                AND rut LIKE '%-[0-9K]'
                AND rut NOT LIKE '%[^0-9K-]%'
            )
        )
    );
END;
GO

IF COL_LENGTH('dbo.users', 'rut') IS NULL
BEGIN
    ALTER TABLE dbo.users ADD rut NVARCHAR(12) NULL;
END;
GO

IF COL_LENGTH('dbo.users', 'entra_oid') IS NULL
BEGIN
    ALTER TABLE dbo.users ADD entra_oid NVARCHAR(100) NULL;
END;
GO

IF COL_LENGTH('dbo.users', 'tenant_id') IS NULL
BEGIN
    ALTER TABLE dbo.users ADD tenant_id NVARCHAR(100) NULL;
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE name = 'ck_users_rut_basic_format' AND parent_object_id = OBJECT_ID('dbo.users'))
BEGIN
    ALTER TABLE dbo.users ADD CONSTRAINT ck_users_rut_basic_format CHECK (
        rut IS NULL OR (
            LEN(rut) BETWEEN 9 AND 10
            AND rut LIKE '%-[0-9K]'
            AND rut NOT LIKE '%[^0-9K-]%'
        )
    );
END;
GO

-- ============================================================
-- TABLE: resources
-- Recursos deportivos asociados a una sede.
-- ============================================================

IF OBJECT_ID('dbo.resources', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.resources (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_resources PRIMARY KEY,
        venue_id INT NOT NULL,
        name NVARCHAR(120) NOT NULL,
        type NVARCHAR(80) NOT NULL,
        reservation_mode NVARCHAR(50) NOT NULL CONSTRAINT df_resources_reservation_mode DEFAULT ('RESERVABLE'),
        image_url NVARCHAR(500) NULL,
        capacity INT NULL,
        is_active BIT NOT NULL CONSTRAINT df_resources_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_resources_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_resources_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_resources_venue FOREIGN KEY (venue_id) REFERENCES dbo.venues(id) ON DELETE NO ACTION,
        CONSTRAINT ck_resources_reservation_mode CHECK (reservation_mode IN ('RESERVABLE', 'OPEN_USE', 'INFORMATIVE', 'ADMIN_ONLY')),
        CONSTRAINT ck_resources_capacity CHECK (capacity IS NULL OR capacity > 0),
        CONSTRAINT uq_resources_venue_name UNIQUE (venue_id, name)
    );
END;
GO

IF COL_LENGTH('dbo.resources', 'reservation_mode') IS NULL
BEGIN
    ALTER TABLE dbo.resources ADD reservation_mode NVARCHAR(50) NOT NULL CONSTRAINT df_resources_reservation_mode DEFAULT ('RESERVABLE');
END;
GO

IF COL_LENGTH('dbo.resources', 'image_url') IS NULL
BEGIN
    ALTER TABLE dbo.resources ADD image_url NVARCHAR(500) NULL;
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE name = 'ck_resources_reservation_mode' AND parent_object_id = OBJECT_ID('dbo.resources'))
BEGIN
    ALTER TABLE dbo.resources ADD CONSTRAINT ck_resources_reservation_mode CHECK (reservation_mode IN ('RESERVABLE', 'OPEN_USE', 'INFORMATIVE', 'ADMIN_ONLY'));
END;
GO

-- ============================================================
-- TABLE: activities
-- Actividades asociadas a reservas o programacion institucional.
-- ============================================================

IF OBJECT_ID('dbo.activities', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.activities (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_activities PRIMARY KEY,
        name NVARCHAR(120) NOT NULL,
        description NVARCHAR(MAX) NULL,
        is_active BIT NOT NULL CONSTRAINT df_activities_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_activities_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_activities_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT uq_activities_name UNIQUE (name)
    );
END;
GO

-- ============================================================
-- TABLE: reservation_policies
-- Versiones prospectivas e inmutables de las reglas de solicitud.
-- effective_from/effective_to son timestamps tecnicos UTC.
-- ============================================================

IF OBJECT_ID('dbo.reservation_policies', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservation_policies (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_reservation_policies PRIMARY KEY,
        reservable_window_days INT NOT NULL,
        request_frequency_days INT NOT NULL,
        confirmation_deadline_minutes INT NOT NULL,
        minimum_participants INT NOT NULL,
		opening_minute INT NOT NULL CONSTRAINT df_reservation_policies_opening DEFAULT (480),
		closing_minute INT NOT NULL CONSTRAINT df_reservation_policies_closing DEFAULT (1320),
		slot_interval_minutes INT NOT NULL CONSTRAINT df_reservation_policies_slot DEFAULT (15),
        effective_from DATETIME2(0) NOT NULL,
        effective_to DATETIME2(0) NULL,
        created_by_user_id INT NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_reservation_policies_created_at DEFAULT (SYSUTCDATETIME()),
		idempotency_key NVARCHAR(100) NULL,
		idempotency_payload_hash CHAR(64) NULL,
		is_published BIT NOT NULL CONSTRAINT df_reservation_policies_published DEFAULT (1),
        CONSTRAINT fk_reservation_policies_created_by FOREIGN KEY (created_by_user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT ck_reservation_policies_window CHECK (reservable_window_days > 0),
        CONSTRAINT ck_reservation_policies_frequency CHECK (request_frequency_days > 0),
        CONSTRAINT ck_reservation_policies_deadline CHECK (confirmation_deadline_minutes >= 0),
        CONSTRAINT ck_reservation_policies_minimum CHECK (minimum_participants > 0),
		CONSTRAINT ck_reservation_policies_schedule CHECK (opening_minute >= 0 AND opening_minute < closing_minute AND closing_minute <= 1439 AND slot_interval_minutes > 0 AND slot_interval_minutes <= 1440),
        CONSTRAINT ck_reservation_policies_effective_range CHECK (effective_to IS NULL OR effective_to >= effective_from)
    );
END;
GO

IF OBJECT_ID('dbo.ck_reservation_policies_effective_range', 'C') IS NOT NULL ALTER TABLE dbo.reservation_policies DROP CONSTRAINT ck_reservation_policies_effective_range;
ALTER TABLE dbo.reservation_policies ADD CONSTRAINT ck_reservation_policies_effective_range CHECK (effective_to IS NULL OR effective_to >= effective_from);
GO

IF COL_LENGTH('dbo.reservation_policies', 'opening_minute') IS NULL ALTER TABLE dbo.reservation_policies ADD opening_minute INT NOT NULL CONSTRAINT df_reservation_policies_opening DEFAULT (480);
IF COL_LENGTH('dbo.reservation_policies', 'closing_minute') IS NULL ALTER TABLE dbo.reservation_policies ADD closing_minute INT NOT NULL CONSTRAINT df_reservation_policies_closing DEFAULT (1320);
IF COL_LENGTH('dbo.reservation_policies', 'slot_interval_minutes') IS NULL ALTER TABLE dbo.reservation_policies ADD slot_interval_minutes INT NOT NULL CONSTRAINT df_reservation_policies_slot DEFAULT (15);
IF COL_LENGTH('dbo.reservation_policies', 'idempotency_key') IS NULL ALTER TABLE dbo.reservation_policies ADD idempotency_key NVARCHAR(100) NULL;
IF COL_LENGTH('dbo.reservation_policies', 'idempotency_payload_hash') IS NULL ALTER TABLE dbo.reservation_policies ADD idempotency_payload_hash CHAR(64) NULL;
IF COL_LENGTH('dbo.reservation_policies', 'is_published') IS NULL ALTER TABLE dbo.reservation_policies ADD is_published BIT NOT NULL CONSTRAINT df_reservation_policies_published DEFAULT (1);
GO

IF OBJECT_ID('dbo.ck_reservation_policies_schedule', 'C') IS NULL
    ALTER TABLE dbo.reservation_policies ADD CONSTRAINT ck_reservation_policies_schedule CHECK (opening_minute >= 0 AND opening_minute < closing_minute AND closing_minute <= 1439 AND slot_interval_minutes > 0 AND slot_interval_minutes <= 1440);
GO

IF NOT EXISTS (SELECT 1 FROM dbo.reservation_policies)
BEGIN
    INSERT INTO dbo.reservation_policies (
        reservable_window_days,
        request_frequency_days,
        confirmation_deadline_minutes,
        minimum_participants,
        effective_from
    )
    VALUES (7, 7, 60, 10, '19000101');
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'uq_reservation_policies_current' AND object_id = OBJECT_ID('dbo.reservation_policies'))
    CREATE UNIQUE INDEX uq_reservation_policies_current ON dbo.reservation_policies(effective_to) WHERE effective_to IS NULL;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'uq_reservation_policies_idempotency' AND object_id = OBJECT_ID('dbo.reservation_policies'))
    CREATE UNIQUE INDEX uq_reservation_policies_idempotency ON dbo.reservation_policies(idempotency_key) WHERE idempotency_key IS NOT NULL;
GO

-- ============================================================
-- TABLE: reservations
-- Reservas realizadas por usuarios.
-- start_time usa DATETIME2 como hora institucional de muro
-- (America/Santiago); no representa un instante UTC.
-- ============================================================

IF OBJECT_ID('dbo.reservations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservations (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_reservations PRIMARY KEY,
        policy_id INT NOT NULL,
        user_id INT NOT NULL,
        resource_id INT NULL,
        activity_id INT NULL,
        start_time DATETIME2(0) NOT NULL,
        duration_minutes INT NOT NULL,
        status NVARCHAR(30) NOT NULL CONSTRAINT df_reservations_status DEFAULT ('PENDING'),
        group_capacity_snapshot INT NULL,
        target_participants INT NULL,
        cancellation_reason NVARCHAR(MAX) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_reservations_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_reservations_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_reservations_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT fk_reservations_policy FOREIGN KEY (policy_id) REFERENCES dbo.reservation_policies(id) ON DELETE NO ACTION,
        CONSTRAINT fk_reservations_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE NO ACTION,
        CONSTRAINT fk_reservations_activity FOREIGN KEY (activity_id) REFERENCES dbo.activities(id) ON DELETE SET NULL,
        CONSTRAINT ck_reservations_duration CHECK (duration_minutes > 0),
        CONSTRAINT ck_reservations_group_capacity_snapshot CHECK (group_capacity_snapshot IS NULL OR group_capacity_snapshot > 0),
        CONSTRAINT ck_reservations_target_participants CHECK (target_participants IS NULL OR target_participants > 0),
        CONSTRAINT ck_reservations_status CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'REJECTED', 'EXPIRED'))
    );
END;
GO

IF COL_LENGTH('dbo.reservations', 'policy_id') IS NULL
BEGIN
    ALTER TABLE dbo.reservations ADD policy_id INT NULL;
END;
GO

IF COL_LENGTH('dbo.reservations', 'group_capacity_snapshot') IS NULL ALTER TABLE dbo.reservations ADD group_capacity_snapshot INT NULL;
IF COL_LENGTH('dbo.reservations', 'target_participants') IS NULL ALTER TABLE dbo.reservations ADD target_participants INT NULL;
GO
IF OBJECT_ID('dbo.ck_reservations_group_capacity_snapshot','C') IS NULL ALTER TABLE dbo.reservations ADD CONSTRAINT ck_reservations_group_capacity_snapshot CHECK(group_capacity_snapshot IS NULL OR group_capacity_snapshot>0);
GO

UPDATE dbo.reservations
SET policy_id = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from ASC, id ASC)
WHERE policy_id IS NULL;
GO

IF EXISTS (SELECT 1 FROM sys.columns WHERE object_id = OBJECT_ID('dbo.reservations') AND name = 'policy_id' AND is_nullable = 1)
BEGIN
    ALTER TABLE dbo.reservations ALTER COLUMN policy_id INT NOT NULL;
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.foreign_keys WHERE name = 'fk_reservations_policy' AND parent_object_id = OBJECT_ID('dbo.reservations'))
BEGIN
    ALTER TABLE dbo.reservations ADD CONSTRAINT fk_reservations_policy FOREIGN KEY (policy_id) REFERENCES dbo.reservation_policies(id) ON DELETE NO ACTION;
END;
GO

-- Recursos que requieren confirmacion grupal. El consumo de esta relacion se
-- incorpora en el incremento tecnico de participantes y estados.
IF OBJECT_ID('dbo.reservation_policy_resources', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservation_policy_resources (
        policy_id INT NOT NULL,
        resource_id INT NOT NULL,
        CONSTRAINT pk_reservation_policy_resources PRIMARY KEY (policy_id, resource_id),
        CONSTRAINT fk_reservation_policy_resources_policy FOREIGN KEY (policy_id) REFERENCES dbo.reservation_policies(id) ON DELETE CASCADE,
        CONSTRAINT fk_reservation_policy_resources_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE NO ACTION
    );
END;
GO

IF OBJECT_ID('dbo.reservation_policy_durations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservation_policy_durations (
        policy_id INT NOT NULL,
        duration_minutes INT NOT NULL,
        CONSTRAINT pk_reservation_policy_durations PRIMARY KEY (policy_id, duration_minutes),
        CONSTRAINT fk_reservation_policy_durations_policy FOREIGN KEY (policy_id) REFERENCES dbo.reservation_policies(id) ON DELETE CASCADE,
        CONSTRAINT ck_reservation_policy_durations_value CHECK (duration_minutes > 0)
    );
END;
GO

IF OBJECT_ID('dbo.reservation_policy_group_resources', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservation_policy_group_resources (
        policy_id INT NOT NULL,
        resource_id INT NOT NULL,
        CONSTRAINT pk_reservation_policy_group_resources PRIMARY KEY (policy_id, resource_id),
        CONSTRAINT fk_group_resources_allowed FOREIGN KEY (policy_id, resource_id) REFERENCES dbo.reservation_policy_resources(policy_id, resource_id),
        CONSTRAINT fk_group_resources_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id)
    );
END;
GO

-- Migracion unica desde la semantica historica (recursos de confirmacion) a
-- una lista completa de recursos permitidos. La marca evita ampliar versiones
-- antiguas cuando se agreguen recursos en ejecuciones futuras del esquema.
IF OBJECT_ID('dbo.reservation_policy_scope_migrations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservation_policy_scope_migrations (
        policy_id INT NOT NULL CONSTRAINT pk_reservation_policy_scope_migrations PRIMARY KEY,
        migrated_at DATETIME2(0) NOT NULL CONSTRAINT df_reservation_policy_scope_migrations_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_reservation_policy_scope_migrations_policy FOREIGN KEY (policy_id) REFERENCES dbo.reservation_policies(id) ON DELETE NO ACTION
    );
END;
GO

INSERT INTO dbo.reservation_policy_durations (policy_id, duration_minutes)
SELECT p.id, d.duration_minutes
FROM dbo.reservation_policies p
CROSS JOIN (VALUES (30), (60), (90), (120), (150), (180)) d(duration_minutes)
WHERE NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_durations);
GO

-- ============================================================
-- TABLE: participants
-- Participantes asociados a una reserva.
-- ============================================================

IF OBJECT_ID('dbo.participants', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.participants (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_participants PRIMARY KEY,
        reservation_id INT NOT NULL,
        user_id INT NOT NULL,
        status NVARCHAR(30) NOT NULL CONSTRAINT df_participants_status DEFAULT ('PENDING'),
        confirmed_at DATETIME2(0) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_participants_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_participants_reservation FOREIGN KEY (reservation_id) REFERENCES dbo.reservations(id) ON DELETE CASCADE,
        CONSTRAINT fk_participants_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT ck_participants_status CHECK (status IN ('PENDING', 'CONFIRMED', 'REJECTED', 'CANCELLED')),
        CONSTRAINT uq_participants_reservation_user UNIQUE (reservation_id, user_id)
    );
END;
GO

IF COL_LENGTH('dbo.reservations', 'join_code_hash') IS NULL ALTER TABLE dbo.reservations ADD join_code_hash CHAR(64) NULL;
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name='uq_reservations_join_code_hash' AND object_id=OBJECT_ID('dbo.reservations')) CREATE UNIQUE INDEX uq_reservations_join_code_hash ON dbo.reservations(join_code_hash) WHERE join_code_hash IS NOT NULL;
GO
IF OBJECT_ID('dbo.reservation_join_code_secrets','U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_join_code_secrets(
  reservation_id INT NOT NULL CONSTRAINT pk_reservation_join_code_secrets PRIMARY KEY,
  key_version INT NOT NULL, nonce VARBINARY(32) NOT NULL, ciphertext VARBINARY(512) NOT NULL,
  rotated_at DATETIME2(0) NOT NULL CONSTRAINT df_join_code_secret_rotated DEFAULT SYSUTCDATETIME(),
  CONSTRAINT fk_join_code_secret_reservation FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id) ON DELETE CASCADE,
  CONSTRAINT ck_join_code_secret_key_version CHECK(key_version>0));
END;
GO
IF OBJECT_ID('dbo.reservation_group_expirations','U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_group_expirations(
  reservation_id INT NOT NULL CONSTRAINT pk_reservation_group_expirations PRIMARY KEY,
  participant_count INT NOT NULL, minimum_participants INT NOT NULL,
  expired_at DATETIME2(0) NOT NULL CONSTRAINT df_group_expired_at DEFAULT SYSUTCDATETIME(),
  CONSTRAINT fk_group_expiration_reservation FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id),
  CONSTRAINT ck_group_expiration_counts CHECK(participant_count>=0 AND minimum_participants>0 AND participant_count<minimum_participants));
END;
GO
IF COL_LENGTH('dbo.participants', 'is_owner') IS NULL ALTER TABLE dbo.participants ADD is_owner BIT NOT NULL CONSTRAINT df_participants_is_owner DEFAULT(0);
IF COL_LENGTH('dbo.participants', 'updated_at') IS NULL ALTER TABLE dbo.participants ADD updated_at DATETIME2(0) NOT NULL CONSTRAINT df_participants_updated_at DEFAULT(SYSUTCDATETIME());
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name='uq_participants_owner' AND object_id=OBJECT_ID('dbo.participants')) CREATE UNIQUE INDEX uq_participants_owner ON dbo.participants(reservation_id) WHERE is_owner=1;
GO

IF OBJECT_ID('dbo.reservation_participant_audit', 'U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_participant_audit (
  id BIGINT IDENTITY(1,1) PRIMARY KEY, reservation_id INT NOT NULL, actor_user_id INT NOT NULL,
  participant_user_id INT NOT NULL, action NVARCHAR(30) NOT NULL, previous_status NVARCHAR(30) NULL,
  new_status NVARCHAR(30) NOT NULL, previous_reservation_status NVARCHAR(30) NOT NULL,
  new_reservation_status NVARCHAR(30) NOT NULL, created_at DATETIME2(0) NOT NULL DEFAULT SYSUTCDATETIME(),
  FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id), FOREIGN KEY(actor_user_id) REFERENCES dbo.users(id),
  FOREIGN KEY(participant_user_id) REFERENCES dbo.users(id)
 );
END;
GO

IF OBJECT_ID('dbo.reservation_target_audit','U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_target_audit(
  id BIGINT IDENTITY(1,1) PRIMARY KEY,reservation_id INT NOT NULL,actor_user_id INT NOT NULL,
  old_target_participants INT NOT NULL,new_target_participants INT NOT NULL,
  created_at DATETIME2(0) NOT NULL DEFAULT SYSUTCDATETIME(),
  FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id),FOREIGN KEY(actor_user_id) REFERENCES dbo.users(id));
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_target_audit_append_only
ON dbo.reservation_target_audit
INSTEAD OF UPDATE, DELETE
AS
BEGIN
 SET NOCOUNT ON;
 THROW 51022,'La auditoria de objetivo es inmutable.',1;
END;
GO

-- ============================================================
-- TABLE: availability_blocks
-- Bloqueos administrativos o mantenciones sobre un recurso.
-- ============================================================

IF OBJECT_ID('dbo.availability_blocks', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.availability_blocks (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_availability_blocks PRIMARY KEY,
        resource_id INT NOT NULL,
        created_by_user_id INT NOT NULL,
        block_type NVARCHAR(50) NOT NULL,
        reason NVARCHAR(MAX) NULL,
        start_time DATETIME2(0) NOT NULL,
        end_time DATETIME2(0) NOT NULL,
        is_active BIT NOT NULL CONSTRAINT df_availability_blocks_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_availability_blocks_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_blocks_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE CASCADE,
        CONSTRAINT fk_blocks_created_by FOREIGN KEY (created_by_user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT ck_blocks_time CHECK (end_time > start_time),
        CONSTRAINT ck_blocks_type CHECK (block_type IN ('MAINTENANCE', 'ADMINISTRATIVE', 'EVENT', 'CLOSED', 'OTHER'))
    );
END;
GO

-- ============================================================
-- TABLE: scheduled_activities
-- Programacion institucional: clases, talleres, eventos y entrenamientos.
-- ============================================================

IF OBJECT_ID('dbo.scheduled_activities', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.scheduled_activities (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_scheduled_activities PRIMARY KEY,
        resource_id INT NOT NULL,
        activity_id INT NULL,
        created_by_user_id INT NOT NULL,
        title NVARCHAR(150) NOT NULL,
        description NVARCHAR(MAX) NULL,
        activity_type NVARCHAR(50) NOT NULL,
        start_time DATETIME2(0) NOT NULL,
        end_time DATETIME2(0) NOT NULL,
        recurrence_rule NVARCHAR(MAX) NULL,
        is_active BIT NOT NULL CONSTRAINT df_scheduled_activities_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_scheduled_activities_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_scheduled_activities_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_scheduled_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE CASCADE,
        CONSTRAINT fk_scheduled_activity FOREIGN KEY (activity_id) REFERENCES dbo.activities(id) ON DELETE SET NULL,
        CONSTRAINT fk_scheduled_created_by FOREIGN KEY (created_by_user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT ck_scheduled_time CHECK (end_time > start_time),
        CONSTRAINT ck_scheduled_type CHECK (activity_type IN ('CLASS', 'WORKSHOP', 'EVENT', 'CHAMPIONSHIP', 'TRAINING', 'OTHER'))
    );
END;
GO

-- ============================================================
-- TABLE: workshops
-- Talleres deportivos recurrentes con inscripcion de estudiantes.
-- ============================================================

IF OBJECT_ID('dbo.workshops', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.workshops (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_workshops PRIMARY KEY,
        resource_id INT NOT NULL,
        title NVARCHAR(150) NOT NULL,
        description NVARCHAR(MAX) NULL,
        day_text NVARCHAR(120) NOT NULL,
        schedule_text NVARCHAR(180) NOT NULL,
        location NVARCHAR(180) NULL,
        instructor_name NVARCHAR(150) NULL,
        capacity INT NOT NULL CONSTRAINT df_workshops_capacity DEFAULT (25),
        is_active BIT NOT NULL CONSTRAINT df_workshops_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_workshops_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_workshops_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_workshops_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE NO ACTION,
        CONSTRAINT ck_workshops_capacity CHECK (capacity > 0)
    );
END;
GO

-- ============================================================
-- TABLE: workshop_occurrences
-- Horario normalizado; weekday_iso usa lunes=1 ... domingo=7.
-- Los intervalos son semiabiertos [inicio, fin), por lo que pueden ser contiguos.
-- ============================================================

IF OBJECT_ID('dbo.workshop_occurrences', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.workshop_occurrences (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_workshop_occurrences PRIMARY KEY,
        workshop_id INT NOT NULL,
        weekday_iso TINYINT NOT NULL,
        start_minute SMALLINT NOT NULL,
        end_minute SMALLINT NOT NULL,
        CONSTRAINT fk_workshop_occurrences_workshop FOREIGN KEY (workshop_id) REFERENCES dbo.workshops(id) ON DELETE CASCADE,
        CONSTRAINT ck_workshop_occurrences_weekday CHECK (weekday_iso BETWEEN 1 AND 7),
        CONSTRAINT ck_workshop_occurrences_minutes CHECK (
            start_minute >= 0 AND start_minute < 1440
            AND end_minute > 0 AND end_minute <= 1440
            AND start_minute < end_minute
        ),
        CONSTRAINT uq_workshop_occurrences_slot UNIQUE (workshop_id, weekday_iso, start_minute, end_minute)
    );
END;
GO

-- ============================================================
-- TABLE: workshop_enrollments
-- Inscripciones de estudiantes a talleres deportivos.
-- ============================================================

IF OBJECT_ID('dbo.workshop_enrollments', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.workshop_enrollments (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_workshop_enrollments PRIMARY KEY,
        workshop_id INT NOT NULL,
        user_id INT NOT NULL,
        status NVARCHAR(30) NOT NULL CONSTRAINT df_workshop_enrollments_status DEFAULT ('CONFIRMED'),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_workshop_enrollments_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_workshop_enrollments_workshop FOREIGN KEY (workshop_id) REFERENCES dbo.workshops(id) ON DELETE CASCADE,
        CONSTRAINT fk_workshop_enrollments_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT ck_workshop_enrollments_status CHECK (status IN ('CONFIRMED', 'CANCELLED'))
    );
END;
GO

-- ============================================================
-- TABLE: violations
-- Infracciones o incumplimientos de usuarios.
-- ============================================================

IF OBJECT_ID('dbo.violations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.violations (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_violations PRIMARY KEY,
        user_id INT NOT NULL,
        reservation_id INT NULL,
        created_by_user_id INT NULL,
        violation_type NVARCHAR(60) NOT NULL,
        description NVARCHAR(MAX) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_violations_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_violations_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT fk_violations_reservation FOREIGN KEY (reservation_id) REFERENCES dbo.reservations(id) ON DELETE SET NULL,
        CONSTRAINT fk_violations_created_by FOREIGN KEY (created_by_user_id) REFERENCES dbo.users(id) ON DELETE SET NULL,
        CONSTRAINT ck_violation_type CHECK (violation_type IN ('NO_SHOW', 'LATE_CANCEL', 'MISUSE', 'PARTICIPANTS_NOT_MET', 'OTHER'))
    );
END;
GO

-- ============================================================
-- TABLE: notifications
-- Notificaciones internas del sistema.
-- ============================================================

IF OBJECT_ID('dbo.notifications', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.notifications (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_notifications PRIMARY KEY,
        user_id INT NOT NULL,
        reservation_id INT NULL,
        title NVARCHAR(150) NOT NULL,
        message NVARCHAR(MAX) NOT NULL,
        type NVARCHAR(50) NOT NULL,
        is_read BIT NOT NULL CONSTRAINT df_notifications_is_read DEFAULT (0),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_notifications_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE CASCADE,
        CONSTRAINT fk_notifications_reservation FOREIGN KEY (reservation_id) REFERENCES dbo.reservations(id) ON DELETE SET NULL,
        CONSTRAINT ck_notification_type CHECK (type IN ('RESERVATION_CREATED', 'RESERVATION_CONFIRMED', 'RESERVATION_CANCELLED', 'RESERVATION_MODIFIED', 'REMINDER', 'SYSTEM'))
    );
END;
GO

-- ============================================================
-- TABLE: audit_logs
-- Auditoria y trazabilidad de acciones relevantes.
-- ============================================================

IF OBJECT_ID('dbo.audit_logs', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.audit_logs (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_audit_logs PRIMARY KEY,
        user_id INT NULL,
        action NVARCHAR(100) NOT NULL,
        entity_type NVARCHAR(100) NOT NULL,
        entity_id INT NULL,
        description NVARCHAR(MAX) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_audit_logs_created_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE SET NULL
    );
END;
GO

-- ============================================================
-- INDEXES
-- ============================================================

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_venues_name' AND object_id = OBJECT_ID('dbo.venues')) CREATE INDEX idx_venues_name ON dbo.venues(name);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_users_email' AND object_id = OBJECT_ID('dbo.users')) CREATE INDEX idx_users_email ON dbo.users(email);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ux_users_rut' AND object_id = OBJECT_ID('dbo.users')) CREATE UNIQUE INDEX ux_users_rut ON dbo.users(rut) WHERE rut IS NOT NULL;
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_users_entra_oid' AND object_id = OBJECT_ID('dbo.users')) CREATE INDEX idx_users_entra_oid ON dbo.users(entra_oid);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ux_users_entra_identity' AND object_id = OBJECT_ID('dbo.users')) CREATE UNIQUE INDEX ux_users_entra_identity ON dbo.users(tenant_id, entra_oid) WHERE tenant_id IS NOT NULL AND entra_oid IS NOT NULL;
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_resources_venue_id' AND object_id = OBJECT_ID('dbo.resources')) CREATE INDEX idx_resources_venue_id ON dbo.resources(venue_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_resources_type' AND object_id = OBJECT_ID('dbo.resources')) CREATE INDEX idx_resources_type ON dbo.resources(type);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_reservations_user_id' AND object_id = OBJECT_ID('dbo.reservations')) CREATE INDEX idx_reservations_user_id ON dbo.reservations(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_reservations_resource_id' AND object_id = OBJECT_ID('dbo.reservations')) CREATE INDEX idx_reservations_resource_id ON dbo.reservations(resource_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_reservations_activity_id' AND object_id = OBJECT_ID('dbo.reservations')) CREATE INDEX idx_reservations_activity_id ON dbo.reservations(activity_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_reservations_start_time' AND object_id = OBJECT_ID('dbo.reservations')) CREATE INDEX idx_reservations_start_time ON dbo.reservations(start_time);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_reservations_status' AND object_id = OBJECT_ID('dbo.reservations')) CREATE INDEX idx_reservations_status ON dbo.reservations(status);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_participants_reservation_id' AND object_id = OBJECT_ID('dbo.participants')) CREATE INDEX idx_participants_reservation_id ON dbo.participants(reservation_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_participants_user_id' AND object_id = OBJECT_ID('dbo.participants')) CREATE INDEX idx_participants_user_id ON dbo.participants(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_blocks_resource_id' AND object_id = OBJECT_ID('dbo.availability_blocks')) CREATE INDEX idx_blocks_resource_id ON dbo.availability_blocks(resource_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_blocks_start_end' AND object_id = OBJECT_ID('dbo.availability_blocks')) CREATE INDEX idx_blocks_start_end ON dbo.availability_blocks(start_time, end_time);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_scheduled_resource_id' AND object_id = OBJECT_ID('dbo.scheduled_activities')) CREATE INDEX idx_scheduled_resource_id ON dbo.scheduled_activities(resource_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_scheduled_activity_id' AND object_id = OBJECT_ID('dbo.scheduled_activities')) CREATE INDEX idx_scheduled_activity_id ON dbo.scheduled_activities(activity_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_scheduled_start_end' AND object_id = OBJECT_ID('dbo.scheduled_activities')) CREATE INDEX idx_scheduled_start_end ON dbo.scheduled_activities(start_time, end_time);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_workshops_active' AND object_id = OBJECT_ID('dbo.workshops')) CREATE INDEX idx_workshops_active ON dbo.workshops(is_active, title);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_workshops_resource_id' AND object_id = OBJECT_ID('dbo.workshops')) CREATE INDEX idx_workshops_resource_id ON dbo.workshops(resource_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_workshop_occurrences_overlap' AND object_id = OBJECT_ID('dbo.workshop_occurrences')) CREATE INDEX idx_workshop_occurrences_overlap ON dbo.workshop_occurrences(weekday_iso, start_minute, end_minute, workshop_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_workshop_enrollments_workshop_id' AND object_id = OBJECT_ID('dbo.workshop_enrollments')) CREATE INDEX idx_workshop_enrollments_workshop_id ON dbo.workshop_enrollments(workshop_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_workshop_enrollments_user_id' AND object_id = OBJECT_ID('dbo.workshop_enrollments')) CREATE INDEX idx_workshop_enrollments_user_id ON dbo.workshop_enrollments(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'ux_workshop_enrollments_active_user' AND object_id = OBJECT_ID('dbo.workshop_enrollments')) CREATE UNIQUE INDEX ux_workshop_enrollments_active_user ON dbo.workshop_enrollments(workshop_id, user_id) WHERE status = 'CONFIRMED';
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_violations_user_id' AND object_id = OBJECT_ID('dbo.violations')) CREATE INDEX idx_violations_user_id ON dbo.violations(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_notifications_user_id' AND object_id = OBJECT_ID('dbo.notifications')) CREATE INDEX idx_notifications_user_id ON dbo.notifications(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_audit_user_id' AND object_id = OBJECT_ID('dbo.audit_logs')) CREATE INDEX idx_audit_user_id ON dbo.audit_logs(user_id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_audit_entity' AND object_id = OBJECT_ID('dbo.audit_logs')) CREATE INDEX idx_audit_entity ON dbo.audit_logs(entity_type, entity_id);
GO

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE OR ALTER TRIGGER dbo.trg_venues_updated_at ON dbo.venues AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.venues target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_users_updated_at ON dbo.users AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.users target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_users_rut_write_once ON dbo.users AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN deleted d ON d.id = i.id
        WHERE NULLIF(LTRIM(RTRIM(d.rut)), '') IS NOT NULL
          AND (
              NULLIF(LTRIM(RTRIM(i.rut)), '') IS NULL
              OR UPPER(REPLACE(REPLACE(LTRIM(RTRIM(d.rut)), '.', ''), ' ', ''))
                 <> UPPER(REPLACE(REPLACE(LTRIM(RTRIM(i.rut)), '.', ''), ' ', ''))
          )
    )
        THROW 51010, 'El RUT no puede modificarse una vez registrado.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_resources_updated_at ON dbo.resources AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.resources target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_activities_updated_at ON dbo.activities AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.activities target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_updated_at ON dbo.reservations AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.reservations target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_group_snapshot_immutable ON dbo.reservations AFTER UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN deleted d ON d.id=i.id WHERE ISNULL(i.group_capacity_snapshot,-1)<>ISNULL(d.group_capacity_snapshot,-1) OR ISNULL(i.join_code_hash,'')<>ISNULL(d.join_code_hash,''))
  THROW 51020,'El snapshot grupal de una solicitud es inmutable.',1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_target_validate ON dbo.reservations AFTER INSERT,UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i LEFT JOIN dbo.reservation_policies p ON p.id=i.policy_id
  WHERE (i.target_participants IS NOT NULL AND i.group_capacity_snapshot IS NULL)
     OR (i.target_participants IS NOT NULL AND (i.target_participants<p.minimum_participants OR i.target_participants>i.group_capacity_snapshot)))
  THROW 51021,'El objetivo de participantes no cumple minimo o capacidad.',1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_scheduled_activities_updated_at ON dbo.scheduled_activities AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.scheduled_activities target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_workshops_updated_at ON dbo.workshops AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.workshops target INNER JOIN inserted i ON i.id = target.id;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_policies_immutable
ON dbo.reservation_policies
AFTER UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (SELECT 1 FROM deleted d LEFT JOIN inserted i ON i.id = d.id WHERE i.id IS NULL)
        THROW 51011, 'Las versiones de politica utilizadas no se pueden eliminar.', 1;

    IF EXISTS (
        SELECT 1
        FROM deleted d
        INNER JOIN inserted i ON i.id = d.id
        WHERE i.reservable_window_days <> d.reservable_window_days
           OR i.request_frequency_days <> d.request_frequency_days
           OR i.confirmation_deadline_minutes <> d.confirmation_deadline_minutes
           OR i.minimum_participants <> d.minimum_participants
		   OR i.opening_minute <> d.opening_minute
		   OR i.closing_minute <> d.closing_minute
		   OR i.slot_interval_minutes <> d.slot_interval_minutes
           OR i.effective_from <> d.effective_from
           OR ISNULL(i.created_by_user_id, -1) <> ISNULL(d.created_by_user_id, -1)
           OR i.created_at <> d.created_at
		   OR ISNULL(i.idempotency_key, N'') <> ISNULL(d.idempotency_key, N'')
		   OR ISNULL(i.idempotency_payload_hash, '') <> ISNULL(d.idempotency_payload_hash, '')
		   OR (i.is_published <> d.is_published AND NOT (d.is_published = 0 AND i.is_published = 1))
    )
        THROW 51012, 'Una version de politica publicada es inmutable.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_resources_immutable
ON dbo.reservation_policy_resources
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (SELECT 1 FROM deleted)
	   OR EXISTS (
	       SELECT 1 FROM inserted i
	       INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
	       WHERE p.is_published = 1
	         AND NOT (
	             TRY_CONVERT(INT, SESSION_CONTEXT(N'legacy_policy_scope_bootstrap')) = 1
	             AND p.idempotency_key IS NULL
	             AND p.id = (SELECT TOP (1) id FROM dbo.reservation_policies ORDER BY effective_from, id)
	             AND NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_scope_migrations m WHERE m.policy_id = p.id)
	         )
	   )
        THROW 51013, 'Los recursos de una version publicada son inmutables.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_group_resources_immutable
ON dbo.reservation_policy_group_resources
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
 IF EXISTS(SELECT 1 FROM deleted d INNER JOIN dbo.reservation_policies p ON p.id=d.policy_id WHERE p.is_published=1)
    OR EXISTS(
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id
        WHERE p.is_published=1
          AND NOT (
              TRY_CONVERT(INT, SESSION_CONTEXT(N'legacy_policy_scope_bootstrap'))=1
              AND p.idempotency_key IS NULL
              AND p.id=(SELECT TOP(1) id FROM dbo.reservation_policies ORDER BY effective_from,id)
              AND NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_scope_migrations m WHERE m.policy_id=p.id)
          )
    )
  THROW 51018, 'Los recursos grupales de una politica publicada son inmutables.', 1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id INNER JOIN dbo.resources r ON r.id=i.resource_id WHERE r.capacity IS NULL OR r.capacity<p.minimum_participants OR r.reservation_mode='OPEN_USE')
  THROW 51019, 'El recurso grupal requiere capacidad suficiente y no puede ser OPEN_USE.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_durations_immutable
ON dbo.reservation_policy_durations
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (SELECT 1 FROM deleted)
	   OR EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id WHERE p.is_published = 1)
        THROW 51014, 'Las duraciones de una version publicada son inmutables.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_workshop_enrollments_validate
ON dbo.workshop_enrollments
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    -- El repositorio toma primero un lock de usuario y luego del taller. Ese es
    -- el camino soportado para serializar altas concurrentes sin deadlocks.
    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops w WITH (UPDLOCK, HOLDLOCK) ON w.id=i.workshop_id
        WHERE i.status='CONFIRMED'
          AND (
              w.is_active=0
              OR NOT EXISTS (
                  SELECT 1 FROM dbo.workshop_occurrences o WITH (HOLDLOCK)
                  WHERE o.workshop_id=i.workshop_id
                    AND o.weekday_iso BETWEEN 1 AND 7
                    AND o.start_minute>=0 AND o.end_minute<=1440
                    AND o.start_minute<o.end_minute
              )
          )
    )
        THROW 51300, 'El taller no esta activo o no tiene horario valido.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops w WITH (UPDLOCK, HOLDLOCK) ON w.id=i.workshop_id
        CROSS APPLY (
            SELECT COUNT_BIG(*) AS confirmed_count
            FROM dbo.workshop_enrollments e WITH (UPDLOCK, HOLDLOCK)
            WHERE e.workshop_id=i.workshop_id AND e.status='CONFIRMED'
        ) counts
        WHERE i.status='CONFIRMED' AND counts.confirmed_count>w.capacity
    )
        THROW 51301, 'El taller no tiene cupos disponibles.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops target_w WITH (HOLDLOCK)
          ON target_w.id=i.workshop_id AND target_w.is_active=1
        JOIN dbo.workshop_occurrences target_o WITH (HOLDLOCK)
          ON target_o.workshop_id=i.workshop_id
        JOIN dbo.workshop_enrollments existing WITH (UPDLOCK, HOLDLOCK)
          ON existing.user_id=i.user_id AND existing.status='CONFIRMED'
         AND existing.id<>i.id AND existing.workshop_id<>i.workshop_id
        JOIN dbo.workshops existing_w WITH (HOLDLOCK)
          ON existing_w.id=existing.workshop_id AND existing_w.is_active=1
        JOIN dbo.workshop_occurrences existing_o WITH (HOLDLOCK)
          ON existing_o.workshop_id=existing.workshop_id
         AND existing_o.weekday_iso=target_o.weekday_iso
         AND existing_o.start_minute<target_o.end_minute
         AND target_o.start_minute<existing_o.end_minute
        WHERE i.status='CONFIRMED'
    )
        THROW 51300, 'El horario se superpone con otro taller confirmado.', 1;
END;
GO

-- ============================================================
-- BUSINESS RULE TRIGGERS
-- Reemplazan las restricciones EXCLUDE de PostgreSQL.
-- Estos triggers son parte del contrato de concurrencia: evitan que dos
-- clientes creen conflictos entre el chequeo de disponibilidad y el INSERT.
-- ============================================================

CREATE OR ALTER TRIGGER dbo.trg_reservations_validate_conflicts
ON dbo.reservations
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

	IF EXISTS (
		SELECT 1 FROM inserted i
		INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
		WHERE NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_durations d WHERE d.policy_id = i.policy_id AND d.duration_minutes = i.duration_minutes)
		   OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) < p.opening_minute
		   OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) >= p.closing_minute
		   OR (DATEPART(HOUR, i.start_time) * 60 + DATEPART(MINUTE, i.start_time)) % p.slot_interval_minutes <> 0
		   OR (DATEPART(HOUR, DATEADD(MINUTE, i.duration_minutes, i.start_time)) * 60 + DATEPART(MINUTE, DATEADD(MINUTE, i.duration_minutes, i.start_time))) > p.closing_minute
		   OR CONVERT(DATE, i.start_time) <> CONVERT(DATE, DATEADD(MINUTE, i.duration_minutes, i.start_time))
	)
		THROW 51015, 'El horario o la duracion no estan permitidos por la politica asociada.', 1;

	IF EXISTS (
		SELECT 1 FROM inserted i
		INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
		WHERE (p.idempotency_key IS NOT NULL OR EXISTS (SELECT 1 FROM dbo.reservation_policy_scope_migrations m WHERE m.policy_id = p.id))
		  AND NOT EXISTS (SELECT 1 FROM dbo.reservation_policy_resources pr WHERE pr.policy_id = i.policy_id AND pr.resource_id = i.resource_id)
	)
		THROW 51016, 'El recurso no esta permitido por la politica asociada.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        LEFT JOIN deleted d ON d.id = i.id
        INNER JOIN dbo.reservation_policies p ON p.id = i.policy_id
        WHERE d.id IS NULL
          AND (
              CONVERT(DATE, i.start_time) < CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
              OR CONVERT(DATE, i.start_time) >= DATEADD(DAY, p.reservable_window_days, CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'))
          )
    )
        THROW 51009, 'La fecha solicitada esta fuera de la ventana reservable vigente.', 1;

    DECLARE @next_request_date DATE;

    SELECT TOP (1)
        @next_request_date = DATEADD(
            DAY,
            previous_policy.request_frequency_days,
            CONVERT(
                DATE,
                previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
            )
        )
    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    INNER JOIN dbo.resources inserted_resource ON inserted_resource.id = i.resource_id
    INNER JOIN dbo.reservations previous WITH (UPDLOCK, HOLDLOCK)
        ON previous.user_id = i.user_id
       AND previous.id <> i.id
       AND previous.status IN ('PENDING', 'CONFIRMED')
    INNER JOIN dbo.resources previous_resource ON previous_resource.id = previous.resource_id
    INNER JOIN dbo.reservation_policies previous_policy ON previous_policy.id = previous.policy_id
    WHERE d.id IS NULL
      AND inserted_resource.reservation_mode <> 'OPEN_USE'
      AND previous_resource.reservation_mode <> 'OPEN_USE'
      AND CONVERT(DATE, i.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time') < DATEADD(
          DAY,
          previous_policy.request_frequency_days,
          CONVERT(
              DATE,
              previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
          )
      )
    ORDER BY DATEADD(
        DAY,
        previous_policy.request_frequency_days,
        CONVERT(
            DATE,
            previous.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time'
        )
    ) DESC, previous.id DESC;

    IF @next_request_date IS NOT NULL
    BEGIN
        DECLARE @frequency_message NVARCHAR(2048) = CONCAT(
            N'Ya existe una solicitud vigente. Proxima fecha permitida: ',
            CONVERT(NVARCHAR(10), @next_request_date, 23),
            N'.'
        );
        THROW 51010, @frequency_message, 1;
    END;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id = i.user_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND u.is_blocked = 1)
        THROW 51000, 'El usuario se encuentra bloqueado y no puede crear reservas.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id=i.user_id INNER JOIN dbo.reservation_policy_group_resources g ON g.policy_id=i.policy_id AND g.resource_id=i.resource_id WHERE i.status IN ('PENDING','CONFIRMED') AND u.is_admin=0 AND NULLIF(LTRIM(RTRIM(u.rut)),'') IS NULL)
        THROW 51017, 'El usuario debe registrar su RUT antes de crear reservas.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.is_active = 0)
        THROW 51001, 'El recurso no esta activo.', 1;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.resources r ON r.id = i.resource_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND r.reservation_mode = 'INFORMATIVE')
        THROW 51002, 'El recurso es solo informativo y no permite reservas.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        INNER JOIN dbo.users u ON u.id = i.user_id
        WHERE i.status IN ('PENDING', 'CONFIRMED')
          AND r.reservation_mode = 'ADMIN_ONLY'
          AND u.is_admin = 0
    )
        THROW 51003, 'El recurso solo puede ser reservado por administradores.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.resource_id = i.resource_id
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND r.reservation_mode <> 'OPEN_USE'
          AND existing.status IN ('PENDING','CONFIRMED')
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51004, 'Ya existe una reserva confirmada para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.user_id = i.user_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND existing.status IN ('PENDING','CONFIRMED')
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51005, 'El usuario ya tiene una reserva confirmada en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > b.start_time
    )
        THROW 51006, 'El recurso esta bloqueado en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.scheduled_activities s ON s.resource_id = i.resource_id
        INNER JOIN dbo.resources r ON r.id = i.resource_id
        WHERE i.status IN ('PENDING','CONFIRMED')
          AND r.reservation_mode <> 'OPEN_USE'
          AND s.is_active = 1
          AND i.start_time < s.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > s.start_time
    )
        THROW 51007, 'El recurso tiene una actividad programada en ese horario.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_blocks_validate_conflicts
ON dbo.availability_blocks
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks existing ON existing.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND existing.is_active = 1
          AND existing.id <> i.id
          AND i.start_time < existing.end_time
          AND i.end_time > existing.start_time
    )
        THROW 51100, 'Ya existe un bloqueo activo para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations r ON r.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND r.status IN ('PENDING','CONFIRMED')
          AND i.start_time < DATEADD(MINUTE, r.duration_minutes, r.start_time)
          AND i.end_time > r.start_time
    )
        THROW 51101, 'El bloqueo se cruza con una reserva confirmada.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_scheduled_activities_validate_conflicts
ON dbo.scheduled_activities
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.scheduled_activities existing ON existing.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND existing.is_active = 1
          AND existing.id <> i.id
          AND i.start_time < existing.end_time
          AND i.end_time > existing.start_time
    )
        THROW 51200, 'Ya existe una actividad programada para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND i.end_time > b.start_time
    )
        THROW 51201, 'La actividad programada se cruza con un bloqueo activo.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations r ON r.resource_id = i.resource_id
        WHERE i.is_active = 1
          AND r.status IN ('PENDING','CONFIRMED')
          AND i.start_time < DATEADD(MINUTE, r.duration_minutes, r.start_time)
          AND i.end_time > r.start_time
    )
        THROW 51202, 'La actividad programada se cruza con una reserva confirmada.', 1;
END;
GO

-- ============================================================
-- AUDIT AND NOTIFICATION TRIGGERS
-- ============================================================

-- Garantias complementarias para instalaciones migradas desde MVP1. Se
-- mantienen tambien en instalaciones limpias para paridad schema/migracion.
CREATE OR ALTER TRIGGER dbo.trg_reservations_pending_conflicts ON dbo.reservations AFTER INSERT,UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservations r ON r.resource_id=i.resource_id WHERE i.status='PENDING' AND r.status IN('PENDING','CONFIRMED') AND r.id<>i.id AND i.start_time<DATEADD(MINUTE,r.duration_minutes,r.start_time) AND DATEADD(MINUTE,i.duration_minutes,i.start_time)>r.start_time) THROW 52010,'La solicitud pendiente se cruza con otra solicitud activa.',1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservations r ON r.user_id=i.user_id WHERE i.status='PENDING' AND r.status IN('PENDING','CONFIRMED') AND r.id<>i.id AND i.start_time<DATEADD(MINUTE,r.duration_minutes,r.start_time) AND DATEADD(MINUTE,i.duration_minutes,i.start_time)>r.start_time) THROW 52015,'El usuario ya tiene una solicitud activa en ese horario.',1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.availability_blocks b ON b.resource_id=i.resource_id WHERE i.status='PENDING' AND b.is_active=1 AND i.start_time<b.end_time AND DATEADD(MINUTE,i.duration_minutes,i.start_time)>b.start_time) THROW 52011,'La solicitud pendiente se cruza con un bloqueo.',1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.scheduled_activities s ON s.resource_id=i.resource_id WHERE i.status='PENDING' AND s.is_active=1 AND i.start_time<s.end_time AND DATEADD(MINUTE,i.duration_minutes,i.start_time)>s.start_time) THROW 52012,'La solicitud pendiente se cruza con una actividad.',1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_blocks_pending_conflicts ON dbo.availability_blocks AFTER INSERT,UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservations r ON r.resource_id=i.resource_id WHERE i.is_active=1 AND r.status='PENDING' AND i.start_time<DATEADD(MINUTE,r.duration_minutes,r.start_time) AND i.end_time>r.start_time) THROW 52013,'El bloqueo se cruza con una solicitud pendiente.',1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_scheduled_activities_pending_conflicts ON dbo.scheduled_activities AFTER INSERT,UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservations r ON r.resource_id=i.resource_id WHERE i.is_active=1 AND r.status='PENDING' AND i.start_time<DATEADD(MINUTE,r.duration_minutes,r.start_time) AND i.end_time>r.start_time) THROW 52014,'La actividad se cruza con una solicitud pendiente.',1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_violations_notify
ON dbo.violations
AFTER INSERT
AS
BEGIN
    SET NOCOUNT ON;

    INSERT INTO dbo.notifications (user_id, reservation_id, title, message, type)
    SELECT
        i.user_id,
        i.reservation_id,
        'Infraccion registrada',
        COALESCE(i.description, 'Se ha registrado una infraccion asociada a tu usuario.'),
        'SYSTEM'
    FROM inserted i;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_audit
ON dbo.reservations
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;

    INSERT INTO dbo.audit_logs (user_id, action, entity_type, entity_id, description)
    SELECT
        COALESCE(i.user_id, d.user_id),
        CASE
            WHEN i.id IS NOT NULL AND d.id IS NULL THEN 'RESERVATION_CREATED'
            WHEN i.id IS NOT NULL AND d.id IS NOT NULL THEN 'RESERVATION_UPDATED'
            ELSE 'RESERVATION_DELETED'
        END,
        'reservations',
        COALESCE(i.id, d.id),
        'Cambio registrado sobre una reserva'
    FROM inserted i
    FULL OUTER JOIN deleted d ON d.id = i.id;
END;
GO

-- ============================================================
-- VIEWS
-- ============================================================

CREATE OR ALTER VIEW dbo.vw_resource_usage AS
SELECT
    v.id AS venue_id,
    v.name AS venue_name,
    r.id AS resource_id,
    r.name AS resource_name,
    COUNT(res.id) AS confirmed_reservations
FROM dbo.resources r
INNER JOIN dbo.venues v ON v.id = r.venue_id
LEFT JOIN dbo.reservations res
    ON res.resource_id = r.id
   AND res.status = 'CONFIRMED'
GROUP BY v.id, v.name, r.id, r.name;
GO

CREATE OR ALTER VIEW dbo.vw_peak_hours AS
SELECT
    DATEPART(HOUR, start_time) AS reservation_hour,
    COUNT(*) AS total_reservations
FROM dbo.reservations
WHERE status = 'CONFIRMED'
GROUP BY DATEPART(HOUR, start_time);
GO

CREATE OR ALTER VIEW dbo.vw_user_violations AS
SELECT
    u.id AS user_id,
    u.email,
    u.full_name,
    COUNT(v.id) AS total_violations
FROM dbo.users u
LEFT JOIN dbo.violations v ON v.user_id = u.id
GROUP BY u.id, u.email, u.full_name;
GO

CREATE OR ALTER VIEW dbo.vw_resource_calendar AS
SELECT
    'RESERVATION' AS item_type,
    r.id AS item_id,
    r.resource_id,
    res.name AS resource_name,
    r.start_time,
    DATEADD(MINUTE, r.duration_minutes, r.start_time) AS end_time,
    r.status,
    COALESCE(a.name, 'Reserva') AS title
FROM dbo.reservations r
INNER JOIN dbo.resources res ON res.id = r.resource_id
LEFT JOIN dbo.activities a ON a.id = r.activity_id
WHERE r.status IN ('PENDING', 'CONFIRMED')
UNION ALL
SELECT
    'BLOCK' AS item_type,
    b.id AS item_id,
    b.resource_id,
    res.name AS resource_name,
    b.start_time,
    b.end_time,
    CASE WHEN b.is_active = 1 THEN 'ACTIVE' ELSE 'INACTIVE' END AS status,
    b.block_type AS title
FROM dbo.availability_blocks b
INNER JOIN dbo.resources res ON res.id = b.resource_id
UNION ALL
SELECT
    'SCHEDULED_ACTIVITY' AS item_type,
    s.id AS item_id,
    s.resource_id,
    res.name AS resource_name,
    s.start_time,
    s.end_time,
    CASE WHEN s.is_active = 1 THEN 'ACTIVE' ELSE 'INACTIVE' END AS status,
    s.title
FROM dbo.scheduled_activities s
INNER JOIN dbo.resources res ON res.id = s.resource_id;
GO
