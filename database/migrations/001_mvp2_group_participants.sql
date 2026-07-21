-- Migracion prospectiva e idempotente desde una base MVP1 poblada.
-- No clasifica ni rellena reservas historicas.
SET XACT_ABORT ON;
BEGIN TRY
 BEGIN TRANSACTION;

 IF OBJECT_ID('dbo.reservation_policy_group_resources','U') IS NULL
  CREATE TABLE dbo.reservation_policy_group_resources(
   policy_id INT NOT NULL, resource_id INT NOT NULL,
   CONSTRAINT pk_reservation_policy_group_resources PRIMARY KEY(policy_id,resource_id),
   CONSTRAINT fk_group_resources_allowed FOREIGN KEY(policy_id,resource_id) REFERENCES dbo.reservation_policy_resources(policy_id,resource_id),
   CONSTRAINT fk_group_resources_resource FOREIGN KEY(resource_id) REFERENCES dbo.resources(id));

 IF COL_LENGTH('dbo.reservations','join_code_hash') IS NULL ALTER TABLE dbo.reservations ADD join_code_hash CHAR(64) NULL;
 IF COL_LENGTH('dbo.reservations','group_capacity_snapshot') IS NULL ALTER TABLE dbo.reservations ADD group_capacity_snapshot INT NULL;
	IF OBJECT_ID('dbo.ck_reservations_group_capacity_snapshot','C') IS NULL ALTER TABLE dbo.reservations ADD CONSTRAINT ck_reservations_group_capacity_snapshot CHECK(group_capacity_snapshot IS NULL OR group_capacity_snapshot>0);
 IF COL_LENGTH('dbo.participants','is_owner') IS NULL ALTER TABLE dbo.participants ADD is_owner BIT NOT NULL CONSTRAINT df_participants_is_owner DEFAULT(0);
 IF COL_LENGTH('dbo.participants','updated_at') IS NULL ALTER TABLE dbo.participants ADD updated_at DATETIME2(0) NOT NULL CONSTRAINT df_participants_updated_at DEFAULT(SYSUTCDATETIME());

 IF OBJECT_ID('dbo.reservation_participant_audit','U') IS NULL
  CREATE TABLE dbo.reservation_participant_audit(
   id BIGINT IDENTITY(1,1) PRIMARY KEY,reservation_id INT NOT NULL,actor_user_id INT NOT NULL,participant_user_id INT NOT NULL,
   action NVARCHAR(30) NOT NULL,previous_status NVARCHAR(30) NULL,new_status NVARCHAR(30) NOT NULL,
   previous_reservation_status NVARCHAR(30) NOT NULL,new_reservation_status NVARCHAR(30) NOT NULL,
   created_at DATETIME2(0) NOT NULL DEFAULT SYSUTCDATETIME(),
   FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id),FOREIGN KEY(actor_user_id) REFERENCES dbo.users(id),FOREIGN KEY(participant_user_id) REFERENCES dbo.users(id));

 IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_reservations_join_code_hash' AND object_id=OBJECT_ID('dbo.reservations')) CREATE UNIQUE INDEX uq_reservations_join_code_hash ON dbo.reservations(join_code_hash) WHERE join_code_hash IS NOT NULL;
 IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_participants_owner' AND object_id=OBJECT_ID('dbo.participants')) CREATE UNIQUE INDEX uq_participants_owner ON dbo.participants(reservation_id) WHERE is_owner=1;

 IF NOT EXISTS(SELECT 1 FROM dbo.reservation_policies WHERE idempotency_key='migration-mvp2-group-v1')
 BEGIN
	IF (SELECT COUNT(*) FROM dbo.resources WHERE id IN(1,2,7))<>3 THROW 52001,'Faltan multicanchas canonicas.',1;
  IF EXISTS(SELECT 1 FROM dbo.resources WHERE id IN(1,2,7) AND (is_active=0 OR reservation_mode='OPEN_USE' OR capacity IS NULL OR capacity<10)) THROW 52002,'Multicancha sin capacidad valida.',1;
  DECLARE @now DATETIME2(0)=SYSUTCDATETIME(),@old INT,@new INT;
  SELECT TOP(1) @old=id FROM dbo.reservation_policies WITH(UPDLOCK,HOLDLOCK) WHERE effective_to IS NULL ORDER BY effective_from DESC,id DESC;
  INSERT INTO dbo.reservation_policies(reservable_window_days,request_frequency_days,confirmation_deadline_minutes,minimum_participants,opening_minute,closing_minute,slot_interval_minutes,effective_from,idempotency_key,idempotency_payload_hash,is_published)
  SELECT reservable_window_days,request_frequency_days,confirmation_deadline_minutes,10,opening_minute,closing_minute,slot_interval_minutes,@now,'migration-mvp2-group-v1','1111111111111111111111111111111111111111111111111111111111111111',0 FROM dbo.reservation_policies WHERE id=@old;
  SET @new=SCOPE_IDENTITY();
  INSERT INTO dbo.reservation_policy_durations(policy_id,duration_minutes) SELECT @new,duration_minutes FROM dbo.reservation_policy_durations WHERE policy_id=@old;
	INSERT INTO dbo.reservation_policy_resources(policy_id,resource_id)
	SELECT @new,resource_id FROM (SELECT resource_id FROM dbo.reservation_policy_resources WHERE policy_id=@old UNION SELECT id FROM dbo.resources WHERE id IN(1,2,7)) allowed;
  INSERT INTO dbo.reservation_policy_group_resources(policy_id,resource_id) VALUES(@new,1),(@new,2),(@new,7);
  UPDATE dbo.reservation_policies SET effective_to=@now WHERE id=@old;
  UPDATE dbo.reservation_policies SET is_published=1 WHERE id=@new;
 END;
 COMMIT TRANSACTION;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_group_snapshot_immutable ON dbo.reservations AFTER UPDATE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN deleted d ON d.id=i.id WHERE ISNULL(i.group_capacity_snapshot,-1)<>ISNULL(d.group_capacity_snapshot,-1) OR ISNULL(i.join_code_hash,'')<>ISNULL(d.join_code_hash,'')) THROW 51020,'El snapshot grupal de una solicitud es inmutable.',1;
END;
GO

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

CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_group_resources_immutable
ON dbo.reservation_policy_group_resources AFTER INSERT,UPDATE,DELETE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM deleted d INNER JOIN dbo.reservation_policies p ON p.id=d.policy_id WHERE p.is_published=1)
    OR EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id WHERE p.is_published=1)
  THROW 51018,'Los recursos grupales de una politica publicada son inmutables.',1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id INNER JOIN dbo.resources r ON r.id=i.resource_id WHERE r.capacity IS NULL OR r.capacity<p.minimum_participants OR r.reservation_mode='OPEN_USE')
  THROW 51019,'El recurso grupal requiere capacidad suficiente y no puede ser OPEN_USE.',1;
END;
GO
