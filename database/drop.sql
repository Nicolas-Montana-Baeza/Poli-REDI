-- ============================================================
-- POLI-REDI - LIMPIEZA BASE DE DATOS
-- Azure SQL Database / SQL Server T-SQL
-- ============================================================

-- VIEWS
IF OBJECT_ID('dbo.vw_resource_calendar', 'V') IS NOT NULL DROP VIEW dbo.vw_resource_calendar;
IF OBJECT_ID('dbo.vw_user_violations', 'V') IS NOT NULL DROP VIEW dbo.vw_user_violations;
IF OBJECT_ID('dbo.vw_peak_hours', 'V') IS NOT NULL DROP VIEW dbo.vw_peak_hours;
IF OBJECT_ID('dbo.vw_resource_usage', 'V') IS NOT NULL DROP VIEW dbo.vw_resource_usage;
IF OBJECT_ID('dbo.vw_priority_reservations', 'V') IS NOT NULL DROP VIEW dbo.vw_priority_reservations;
IF OBJECT_ID('dbo.vw_activity_usage', 'V') IS NOT NULL DROP VIEW dbo.vw_activity_usage;
GO

-- TRIGGERS
IF OBJECT_ID('dbo.trg_reservations_audit', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservations_audit;
IF OBJECT_ID('dbo.trg_reservation_policies_immutable', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservation_policies_immutable;
IF OBJECT_ID('dbo.trg_reservation_policy_resources_immutable', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservation_policy_resources_immutable;
IF OBJECT_ID('dbo.trg_reservation_policy_durations_immutable', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservation_policy_durations_immutable;
IF OBJECT_ID('dbo.trg_violations_notify', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_violations_notify;
IF OBJECT_ID('dbo.trg_scheduled_activities_validate_conflicts', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_scheduled_activities_validate_conflicts;
IF OBJECT_ID('dbo.trg_blocks_validate_conflicts', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_blocks_validate_conflicts;
IF OBJECT_ID('dbo.trg_reservations_validate_conflicts', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservations_validate_conflicts;
IF OBJECT_ID('dbo.trg_scheduled_activities_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_scheduled_activities_updated_at;
IF OBJECT_ID('dbo.trg_workshops_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_workshops_updated_at;
IF OBJECT_ID('dbo.trg_reservations_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_reservations_updated_at;
IF OBJECT_ID('dbo.trg_activities_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_activities_updated_at;
IF OBJECT_ID('dbo.trg_resources_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_resources_updated_at;
IF OBJECT_ID('dbo.trg_users_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_users_updated_at;
IF OBJECT_ID('dbo.trg_venues_updated_at', 'TR') IS NOT NULL DROP TRIGGER dbo.trg_venues_updated_at;
GO

-- TABLES
IF OBJECT_ID('dbo.audit_logs', 'U') IS NOT NULL DROP TABLE dbo.audit_logs;
IF OBJECT_ID('dbo.logs', 'U') IS NOT NULL DROP TABLE dbo.logs;
IF OBJECT_ID('dbo.notifications', 'U') IS NOT NULL DROP TABLE dbo.notifications;
IF OBJECT_ID('dbo.violations', 'U') IS NOT NULL DROP TABLE dbo.violations;
IF OBJECT_ID('dbo.priority_reservations', 'U') IS NOT NULL DROP TABLE dbo.priority_reservations;
IF OBJECT_ID('dbo.workshop_enrollments', 'U') IS NOT NULL DROP TABLE dbo.workshop_enrollments;
IF OBJECT_ID('dbo.workshops', 'U') IS NOT NULL DROP TABLE dbo.workshops;
IF OBJECT_ID('dbo.scheduled_activities', 'U') IS NOT NULL DROP TABLE dbo.scheduled_activities;
IF OBJECT_ID('dbo.availability_blocks', 'U') IS NOT NULL DROP TABLE dbo.availability_blocks;
IF OBJECT_ID('dbo.participants', 'U') IS NOT NULL DROP TABLE dbo.participants;
IF OBJECT_ID('dbo.reservations', 'U') IS NOT NULL DROP TABLE dbo.reservations;
IF OBJECT_ID('dbo.reservation_policy_resources', 'U') IS NOT NULL DROP TABLE dbo.reservation_policy_resources;
IF OBJECT_ID('dbo.reservation_policy_durations', 'U') IS NOT NULL DROP TABLE dbo.reservation_policy_durations;
IF OBJECT_ID('dbo.reservation_policy_scope_migrations', 'U') IS NOT NULL DROP TABLE dbo.reservation_policy_scope_migrations;
IF OBJECT_ID('dbo.reservation_policies', 'U') IS NOT NULL DROP TABLE dbo.reservation_policies;
IF OBJECT_ID('dbo.activities', 'U') IS NOT NULL DROP TABLE dbo.activities;
IF OBJECT_ID('dbo.resources', 'U') IS NOT NULL DROP TABLE dbo.resources;
IF OBJECT_ID('dbo.users', 'U') IS NOT NULL DROP TABLE dbo.users;
IF OBJECT_ID('dbo.venues', 'U') IS NOT NULL DROP TABLE dbo.venues;
GO
