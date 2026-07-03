-- ============================================================
-- POLI-REDI - DDL BASE DE DATOS
-- Azure SQL Database / SQL Server T-SQL
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
        is_admin BIT NOT NULL CONSTRAINT df_users_is_admin DEFAULT (0),
        is_blocked BIT NOT NULL CONSTRAINT df_users_is_blocked DEFAULT (0),
        entra_oid NVARCHAR(100) NULL,
        tenant_id NVARCHAR(100) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_users_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_users_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT uq_users_email UNIQUE (email),
        CONSTRAINT ck_users_email_format CHECK (email LIKE '%_@_%._%')
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
        capacity INT NULL,
        is_active BIT NOT NULL CONSTRAINT df_resources_is_active DEFAULT (1),
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_resources_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_resources_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_resources_venue FOREIGN KEY (venue_id) REFERENCES dbo.venues(id) ON DELETE NO ACTION,
        CONSTRAINT ck_resources_reservation_mode CHECK (reservation_mode IN ('RESERVABLE', 'INFORMATIVE', 'ADMIN_ONLY')),
        CONSTRAINT ck_resources_capacity CHECK (capacity IS NULL OR capacity > 0),
        CONSTRAINT uq_resources_venue_name UNIQUE (venue_id, name)
    );
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
-- TABLE: reservations
-- Reservas realizadas por usuarios.
-- ============================================================

IF OBJECT_ID('dbo.reservations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.reservations (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_reservations PRIMARY KEY,
        user_id INT NOT NULL,
        resource_id INT NOT NULL,
        activity_id INT NULL,
        start_time DATETIME2(0) NOT NULL,
        duration_minutes INT NOT NULL,
        status NVARCHAR(30) NOT NULL CONSTRAINT df_reservations_status DEFAULT ('PENDING'),
        cancellation_reason NVARCHAR(MAX) NULL,
        created_at DATETIME2(0) NOT NULL CONSTRAINT df_reservations_created_at DEFAULT (SYSUTCDATETIME()),
        updated_at DATETIME2(0) NOT NULL CONSTRAINT df_reservations_updated_at DEFAULT (SYSUTCDATETIME()),
        CONSTRAINT fk_reservations_user FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE NO ACTION,
        CONSTRAINT fk_reservations_resource FOREIGN KEY (resource_id) REFERENCES dbo.resources(id) ON DELETE NO ACTION,
        CONSTRAINT fk_reservations_activity FOREIGN KEY (activity_id) REFERENCES dbo.activities(id) ON DELETE SET NULL,
        CONSTRAINT ck_reservations_duration CHECK (duration_minutes > 0),
        CONSTRAINT ck_reservations_status CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'REJECTED', 'EXPIRED'))
    );
END;
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

CREATE OR ALTER TRIGGER dbo.trg_scheduled_activities_updated_at ON dbo.scheduled_activities AFTER UPDATE AS
BEGIN
    SET NOCOUNT ON; IF TRIGGER_NESTLEVEL() > 1 RETURN;
    UPDATE target SET updated_at = SYSUTCDATETIME() FROM dbo.scheduled_activities target INNER JOIN inserted i ON i.id = target.id;
END;
GO

-- ============================================================
-- BUSINESS RULE TRIGGERS
-- Reemplazan las restricciones EXCLUDE de PostgreSQL.
-- ============================================================

CREATE OR ALTER TRIGGER dbo.trg_reservations_validate_conflicts
ON dbo.reservations
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (SELECT 1 FROM inserted i INNER JOIN dbo.users u ON u.id = i.user_id WHERE i.status IN ('PENDING', 'CONFIRMED') AND u.is_blocked = 1)
        THROW 51000, 'El usuario se encuentra bloqueado y no puede crear reservas.', 1;

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
        WHERE i.status = 'CONFIRMED'
          AND existing.status = 'CONFIRMED'
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51004, 'Ya existe una reserva confirmada para ese recurso en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations existing ON existing.user_id = i.user_id
        WHERE i.status = 'CONFIRMED'
          AND existing.status = 'CONFIRMED'
          AND existing.id <> i.id
          AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
    )
        THROW 51005, 'El usuario ya tiene una reserva confirmada en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.availability_blocks b ON b.resource_id = i.resource_id
        WHERE i.status = 'CONFIRMED'
          AND b.is_active = 1
          AND i.start_time < b.end_time
          AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > b.start_time
    )
        THROW 51006, 'El recurso esta bloqueado en ese horario.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.scheduled_activities s ON s.resource_id = i.resource_id
        WHERE i.status = 'CONFIRMED'
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
          AND r.status = 'CONFIRMED'
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
          AND r.status = 'CONFIRMED'
          AND i.start_time < DATEADD(MINUTE, r.duration_minutes, r.start_time)
          AND i.end_time > r.start_time
    )
        THROW 51202, 'La actividad programada se cruza con una reserva confirmada.', 1;
END;
GO

-- ============================================================
-- AUDIT AND NOTIFICATION TRIGGERS
-- ============================================================

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
