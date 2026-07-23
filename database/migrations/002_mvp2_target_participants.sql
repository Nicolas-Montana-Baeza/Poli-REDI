-- Target editable para solicitudes grupales. Requiere migration 001.
SET XACT_ABORT ON;
IF OBJECT_ID('dbo.reservations','U') IS NULL OR COL_LENGTH('dbo.reservations','group_capacity_snapshot') IS NULL
 THROW 52100,'Preflight: ejecute primero migration 001.',1;
GO
BEGIN TRY
 BEGIN TRANSACTION;
 IF COL_LENGTH('dbo.reservations','target_participants') IS NULL
  EXEC(N'ALTER TABLE dbo.reservations ADD target_participants INT NULL;');
 COMMIT TRANSACTION;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO
IF OBJECT_ID('dbo.ck_reservations_target_participants','C') IS NULL
 ALTER TABLE dbo.reservations ADD CONSTRAINT ck_reservations_target_participants CHECK(target_participants IS NULL OR target_participants>0);
IF OBJECT_ID('dbo.reservation_target_audit','U') IS NULL
 CREATE TABLE dbo.reservation_target_audit(
  id BIGINT IDENTITY(1,1) PRIMARY KEY,reservation_id INT NOT NULL,actor_user_id INT NOT NULL,
  old_target_participants INT NOT NULL,new_target_participants INT NOT NULL,
  created_at DATETIME2(0) NOT NULL DEFAULT SYSUTCDATETIME(),
  FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id),FOREIGN KEY(actor_user_id) REFERENCES dbo.users(id));
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
CREATE OR ALTER TRIGGER dbo.trg_reservation_target_audit_append_only ON dbo.reservation_target_audit INSTEAD OF UPDATE,DELETE AS
BEGIN
 SET NOCOUNT ON;
 THROW 51022,'La auditoria de objetivo es inmutable.',1;
END;
GO
SELECT CONVERT(bit,CASE WHEN COL_LENGTH('dbo.reservations','target_participants') IS NOT NULL THEN 1 ELSE 0 END) target_column_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.ck_reservations_target_participants','C') IS NOT NULL THEN 1 ELSE 0 END) target_constraint_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.trg_reservations_target_validate','TR') IS NOT NULL THEN 1 ELSE 0 END) target_validation_trigger_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.reservation_target_audit','U') IS NOT NULL THEN 1 ELSE 0 END) target_audit_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.trg_reservation_target_audit_append_only','TR') IS NOT NULL THEN 1 ELSE 0 END) target_audit_append_only_ok;
GO
