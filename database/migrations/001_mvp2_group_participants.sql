-- POLI-REDI MVP2: migracion acumulativa desde la base MVP1 mas antigua.
-- Ejecutar el archivo completo con soporte de separadores GO.
-- Los batches estructurales son idempotentes. La publicacion prospectiva es
-- el ultimo paso y ocurre en una unica transaccion.

SET NOCOUNT ON;
SET XACT_ABORT ON;

-- PREFLIGHT: solo objetos que la migracion no puede reconstruir sin inventar datos.
IF OBJECT_ID('dbo.users','U') IS NULL THROW 52000,'Preflight: falta dbo.users.',1;
IF OBJECT_ID('dbo.resources','U') IS NULL THROW 52000,'Preflight: falta dbo.resources.',1;
IF OBJECT_ID('dbo.reservation_policies','U') IS NULL THROW 52000,'Preflight: falta dbo.reservation_policies.',1;
IF OBJECT_ID('dbo.reservations','U') IS NULL THROW 52000,'Preflight: falta dbo.reservations.',1;
IF OBJECT_ID('dbo.participants','U') IS NULL THROW 52000,'Preflight: falta dbo.participants.',1;
IF COL_LENGTH('dbo.reservation_policies','reservable_window_days') IS NULL THROW 52000,'Preflight: falta reservable_window_days.',1;
IF COL_LENGTH('dbo.reservation_policies','request_frequency_days') IS NULL THROW 52000,'Preflight: falta request_frequency_days.',1;
IF COL_LENGTH('dbo.reservation_policies','confirmation_deadline_minutes') IS NULL THROW 52000,'Preflight: falta confirmation_deadline_minutes.',1;
IF COL_LENGTH('dbo.reservation_policies','minimum_participants') IS NULL THROW 52000,'Preflight: falta minimum_participants.',1;
IF COL_LENGTH('dbo.reservation_policies','effective_from') IS NULL THROW 52000,'Preflight: falta effective_from.',1;
IF COL_LENGTH('dbo.reservation_policies','effective_to') IS NULL THROW 52000,'Preflight: falta effective_to.',1;
IF NOT EXISTS(SELECT 1 FROM dbo.reservation_policies) THROW 52000,'Preflight: no existe politica base.',1;
IF (SELECT COUNT(*) FROM dbo.resources WHERE id IN(1,2,7))<>3 THROW 52001,'Preflight: faltan multicanchas 1, 2 o 7.',1;
IF EXISTS(SELECT 1 FROM dbo.resources WHERE id IN(1,2,7) AND (is_active=0 OR reservation_mode='OPEN_USE' OR capacity IS NULL OR capacity<10))
 THROW 52002,'Preflight: una multicancha no esta activa o no tiene capacidad valida.',1;
GO

-- FASE 1: columnas historicamente ausentes. EXEC fuerza compilacion posterior.
BEGIN TRY
 BEGIN TRANSACTION;
 IF COL_LENGTH('dbo.reservation_policies','opening_minute') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD opening_minute INT NOT NULL CONSTRAINT df_reservation_policies_opening DEFAULT(480) WITH VALUES;');
 IF COL_LENGTH('dbo.reservation_policies','closing_minute') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD closing_minute INT NOT NULL CONSTRAINT df_reservation_policies_closing DEFAULT(1320) WITH VALUES;');
 IF COL_LENGTH('dbo.reservation_policies','slot_interval_minutes') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD slot_interval_minutes INT NOT NULL CONSTRAINT df_reservation_policies_slot DEFAULT(15) WITH VALUES;');
 IF COL_LENGTH('dbo.reservation_policies','idempotency_key') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD idempotency_key NVARCHAR(100) NULL;');
 IF COL_LENGTH('dbo.reservation_policies','idempotency_payload_hash') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD idempotency_payload_hash CHAR(64) NULL;');
 IF COL_LENGTH('dbo.reservation_policies','is_published') IS NULL
  EXEC(N'ALTER TABLE dbo.reservation_policies ADD is_published BIT NOT NULL CONSTRAINT df_reservation_policies_published DEFAULT(1) WITH VALUES;');
 IF COL_LENGTH('dbo.reservations','policy_id') IS NULL
  EXEC(N'ALTER TABLE dbo.reservations ADD policy_id INT NULL;');
 IF COL_LENGTH('dbo.reservations','join_code_hash') IS NULL
  EXEC(N'ALTER TABLE dbo.reservations ADD join_code_hash CHAR(64) NULL;');
 IF COL_LENGTH('dbo.reservations','group_capacity_snapshot') IS NULL
  EXEC(N'ALTER TABLE dbo.reservations ADD group_capacity_snapshot INT NULL;');
 IF COL_LENGTH('dbo.participants','is_owner') IS NULL
  EXEC(N'ALTER TABLE dbo.participants ADD is_owner BIT NOT NULL CONSTRAINT df_participants_is_owner DEFAULT(0) WITH VALUES;');
 IF COL_LENGTH('dbo.participants','updated_at') IS NULL
  EXEC(N'ALTER TABLE dbo.participants ADD updated_at DATETIME2(0) NOT NULL CONSTRAINT df_participants_updated_at DEFAULT(SYSUTCDATETIME()) WITH VALUES;');
 COMMIT;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO

-- FASE 2: tablas de colecciones y auditoria.
BEGIN TRY
 BEGIN TRANSACTION;
 IF OBJECT_ID('dbo.reservation_policy_resources','U') IS NULL
  CREATE TABLE dbo.reservation_policy_resources(
   policy_id INT NOT NULL,resource_id INT NOT NULL,
   CONSTRAINT pk_reservation_policy_resources PRIMARY KEY(policy_id,resource_id),
   CONSTRAINT fk_reservation_policy_resources_policy FOREIGN KEY(policy_id) REFERENCES dbo.reservation_policies(id),
   CONSTRAINT fk_reservation_policy_resources_resource FOREIGN KEY(resource_id) REFERENCES dbo.resources(id));
 IF OBJECT_ID('dbo.reservation_policy_durations','U') IS NULL
  CREATE TABLE dbo.reservation_policy_durations(
   policy_id INT NOT NULL,duration_minutes INT NOT NULL,
   CONSTRAINT pk_reservation_policy_durations PRIMARY KEY(policy_id,duration_minutes),
   CONSTRAINT fk_reservation_policy_durations_policy FOREIGN KEY(policy_id) REFERENCES dbo.reservation_policies(id),
   CONSTRAINT ck_reservation_policy_durations_value CHECK(duration_minutes>0));
 IF EXISTS(SELECT policy_id,resource_id FROM dbo.reservation_policy_resources GROUP BY policy_id,resource_id HAVING COUNT(*)>1)
  THROW 52006,'Preflight: reservation_policy_resources contiene duplicados.',1;
 IF NOT EXISTS(SELECT 1 FROM sys.key_constraints WHERE type='PK' AND parent_object_id=OBJECT_ID('dbo.reservation_policy_resources'))
  ALTER TABLE dbo.reservation_policy_resources ADD CONSTRAINT pk_reservation_policy_resources PRIMARY KEY(policy_id,resource_id);
 IF EXISTS(SELECT policy_id,duration_minutes FROM dbo.reservation_policy_durations GROUP BY policy_id,duration_minutes HAVING COUNT(*)>1)
  THROW 52007,'Preflight: reservation_policy_durations contiene duplicados.',1;
 IF NOT EXISTS(SELECT 1 FROM sys.key_constraints WHERE type='PK' AND parent_object_id=OBJECT_ID('dbo.reservation_policy_durations'))
  ALTER TABLE dbo.reservation_policy_durations ADD CONSTRAINT pk_reservation_policy_durations PRIMARY KEY(policy_id,duration_minutes);
 IF OBJECT_ID('dbo.reservation_policy_scope_migrations','U') IS NULL
  CREATE TABLE dbo.reservation_policy_scope_migrations(
   policy_id INT NOT NULL CONSTRAINT pk_reservation_policy_scope_migrations PRIMARY KEY,
   migrated_at DATETIME2(0) NOT NULL CONSTRAINT df_reservation_policy_scope_migrations_at DEFAULT SYSUTCDATETIME(),
   CONSTRAINT fk_reservation_policy_scope_migrations_policy FOREIGN KEY(policy_id) REFERENCES dbo.reservation_policies(id));
 IF OBJECT_ID('dbo.reservation_policy_group_resources','U') IS NULL
  CREATE TABLE dbo.reservation_policy_group_resources(
   policy_id INT NOT NULL,resource_id INT NOT NULL,
   CONSTRAINT pk_reservation_policy_group_resources PRIMARY KEY(policy_id,resource_id),
   CONSTRAINT fk_group_resources_allowed FOREIGN KEY(policy_id,resource_id) REFERENCES dbo.reservation_policy_resources(policy_id,resource_id),
   CONSTRAINT fk_group_resources_resource FOREIGN KEY(resource_id) REFERENCES dbo.resources(id));
 IF OBJECT_ID('dbo.reservation_participant_audit','U') IS NULL
  CREATE TABLE dbo.reservation_participant_audit(
   id BIGINT IDENTITY(1,1) PRIMARY KEY,reservation_id INT NOT NULL,actor_user_id INT NOT NULL,participant_user_id INT NOT NULL,
   action NVARCHAR(30) NOT NULL,previous_status NVARCHAR(30) NULL,new_status NVARCHAR(30) NOT NULL,
   previous_reservation_status NVARCHAR(30) NOT NULL,new_reservation_status NVARCHAR(30) NOT NULL,
   created_at DATETIME2(0) NOT NULL DEFAULT SYSUTCDATETIME(),
   FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id),FOREIGN KEY(actor_user_id) REFERENCES dbo.users(id),FOREIGN KEY(participant_user_id) REFERENCES dbo.users(id));
 COMMIT;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO

-- FASE 3: backfill exclusivamente tecnico. No clasifica reservas historicas.
BEGIN TRY
 BEGIN TRANSACTION;
 DECLARE @bootstrap_policy INT=(SELECT TOP(1) id FROM dbo.reservation_policies ORDER BY effective_from,id);
 UPDATE dbo.reservations SET policy_id=@bootstrap_policy WHERE policy_id IS NULL;
 IF EXISTS(SELECT 1 FROM sys.columns WHERE object_id=OBJECT_ID('dbo.reservations') AND name='policy_id' AND is_nullable=1)
  ALTER TABLE dbo.reservations ALTER COLUMN policy_id INT NOT NULL;
 IF NOT EXISTS(SELECT 1 FROM sys.foreign_keys WHERE name='fk_reservations_policy' AND parent_object_id=OBJECT_ID('dbo.reservations'))
  ALTER TABLE dbo.reservations ADD CONSTRAINT fk_reservations_policy FOREIGN KEY(policy_id) REFERENCES dbo.reservation_policies(id);
 IF NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_scope_migrations WHERE policy_id=@bootstrap_policy)
 BEGIN
  INSERT INTO dbo.reservation_policy_resources(policy_id,resource_id)
  SELECT @bootstrap_policy,id FROM dbo.resources r WHERE is_active=1 AND reservation_mode<>'INFORMATIVE'
   AND NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_resources x WHERE x.policy_id=@bootstrap_policy AND x.resource_id=r.id);
  INSERT INTO dbo.reservation_policy_scope_migrations(policy_id) VALUES(@bootstrap_policy);
 END;
 INSERT INTO dbo.reservation_policy_durations(policy_id,duration_minutes)
 SELECT p.id,d.v FROM dbo.reservation_policies p CROSS JOIN(VALUES(30),(60),(90),(120),(150),(180))d(v)
 WHERE NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_durations x WHERE x.policy_id=p.id AND x.duration_minutes=d.v);
 COMMIT;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO

-- FASE 4: constraints e indices; ya se compilan contra columnas existentes.
IF EXISTS(SELECT reservation_id,user_id FROM dbo.participants GROUP BY reservation_id,user_id HAVING COUNT(*)>1)
 THROW 52005,'Preflight: hay participantes duplicados; se requiere correccion manual.',1;
IF NOT EXISTS(SELECT 1 FROM sys.key_constraints WHERE name='uq_participants_reservation_user' AND parent_object_id=OBJECT_ID('dbo.participants'))
 ALTER TABLE dbo.participants ADD CONSTRAINT uq_participants_reservation_user UNIQUE(reservation_id,user_id);
IF NOT EXISTS(SELECT 1 FROM sys.key_constraints WHERE type='PK' AND parent_object_id=OBJECT_ID('dbo.reservation_policy_group_resources'))
 ALTER TABLE dbo.reservation_policy_group_resources ADD CONSTRAINT pk_reservation_policy_group_resources PRIMARY KEY(policy_id,resource_id);
IF NOT EXISTS(SELECT 1 FROM sys.foreign_keys WHERE name='fk_group_resources_allowed' AND parent_object_id=OBJECT_ID('dbo.reservation_policy_group_resources'))
 ALTER TABLE dbo.reservation_policy_group_resources ADD CONSTRAINT fk_group_resources_allowed FOREIGN KEY(policy_id,resource_id) REFERENCES dbo.reservation_policy_resources(policy_id,resource_id);
IF NOT EXISTS(SELECT 1 FROM sys.foreign_keys WHERE name='fk_group_resources_resource' AND parent_object_id=OBJECT_ID('dbo.reservation_policy_group_resources'))
 ALTER TABLE dbo.reservation_policy_group_resources ADD CONSTRAINT fk_group_resources_resource FOREIGN KEY(resource_id) REFERENCES dbo.resources(id);
IF OBJECT_ID('dbo.ck_reservations_group_capacity_snapshot','C') IS NULL
 ALTER TABLE dbo.reservations ADD CONSTRAINT ck_reservations_group_capacity_snapshot CHECK(group_capacity_snapshot IS NULL OR group_capacity_snapshot>0);
IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_reservations_join_code_hash' AND object_id=OBJECT_ID('dbo.reservations'))
 CREATE UNIQUE INDEX uq_reservations_join_code_hash ON dbo.reservations(join_code_hash) WHERE join_code_hash IS NOT NULL;
IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_participants_owner' AND object_id=OBJECT_ID('dbo.participants'))
 CREATE UNIQUE INDEX uq_participants_owner ON dbo.participants(reservation_id) WHERE is_owner=1;
IF EXISTS(SELECT idempotency_key FROM dbo.reservation_policies WHERE idempotency_key IS NOT NULL GROUP BY idempotency_key HAVING COUNT(*)>1)
 THROW 52009,'Preflight: hay claves de idempotencia duplicadas.',1;
IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_reservation_policies_idempotency' AND object_id=OBJECT_ID('dbo.reservation_policies'))
 CREATE UNIQUE INDEX uq_reservation_policies_idempotency ON dbo.reservation_policies(idempotency_key) WHERE idempotency_key IS NOT NULL;
IF (SELECT COUNT(*) FROM dbo.reservation_policies WHERE effective_to IS NULL AND is_published=1)<>1
 THROW 52008,'Preflight: debe existir exactamente una politica vigente publicada.',1;
IF (SELECT COUNT(*) FROM dbo.reservation_policies WHERE effective_to IS NULL)<>1
 THROW 52009,'Preflight: hay vigencias abiertas adicionales; se requiere revision manual.',1;
IF NOT EXISTS(SELECT 1 FROM sys.indexes WHERE name='uq_reservation_policies_current' AND object_id=OBJECT_ID('dbo.reservation_policies'))
 CREATE UNIQUE INDEX uq_reservation_policies_current ON dbo.reservation_policies(effective_to) WHERE effective_to IS NULL;
GO

-- FASE 5: triggers canónicos. CREATE OR ALTER repara intentos parciales.
CREATE OR ALTER TRIGGER dbo.trg_reservation_policies_immutable
ON dbo.reservation_policies
AFTER UPDATE, DELETE
AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM deleted d LEFT JOIN inserted i ON i.id=d.id WHERE i.id IS NULL)
  THROW 51011,'Las versiones de politica utilizadas no se pueden eliminar.',1;
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
     OR ISNULL(i.created_by_user_id,-1)<>ISNULL(d.created_by_user_id,-1)
     OR i.created_at<>d.created_at
     OR ISNULL(i.idempotency_key,N'')<>ISNULL(d.idempotency_key,N'')
     OR ISNULL(i.idempotency_payload_hash,'')<>ISNULL(d.idempotency_payload_hash,'')
     OR (i.is_published<>d.is_published AND NOT(d.is_published=0 AND i.is_published=1))
 )
  THROW 51012,'Una version de politica publicada es inmutable.',1;
END;
GO
CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_resources_immutable
ON dbo.reservation_policy_resources
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM deleted)
    OR EXISTS(
     SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id
     WHERE p.is_published=1
       AND NOT(
        TRY_CONVERT(INT,SESSION_CONTEXT(N'legacy_policy_scope_bootstrap'))=1
        AND p.idempotency_key IS NULL
        AND p.id=(SELECT TOP(1) id FROM dbo.reservation_policies ORDER BY effective_from,id)
        AND NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_scope_migrations m WHERE m.policy_id=p.id)
       )
    )
  THROW 51013,'Los recursos de una version publicada son inmutables.',1;
END;
GO
CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_durations_immutable
ON dbo.reservation_policy_durations
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM deleted)
    OR EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id WHERE p.is_published=1)
  THROW 51014,'Las duraciones de una version publicada son inmutables.',1;
END;
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
CREATE OR ALTER TRIGGER dbo.trg_reservation_policy_group_resources_immutable ON dbo.reservation_policy_group_resources AFTER INSERT,UPDATE,DELETE AS
BEGIN
 SET NOCOUNT ON;
 IF EXISTS(SELECT 1 FROM deleted d INNER JOIN dbo.reservation_policies p ON p.id=d.policy_id WHERE p.is_published=1)
    OR EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id WHERE p.is_published=1)
  THROW 51018,'Los recursos grupales de una politica publicada son inmutables.',1;
 IF EXISTS(SELECT 1 FROM inserted i INNER JOIN dbo.reservation_policies p ON p.id=i.policy_id INNER JOIN dbo.resources r ON r.id=i.resource_id WHERE r.capacity IS NULL OR r.capacity<p.minimum_participants OR r.reservation_mode='OPEN_USE')
  THROW 51019,'El recurso grupal requiere capacidad suficiente y no puede ser OPEN_USE.',1;
END;
GO

-- FASE 6: única fase que publica/cierra vigencias. Si falla, revierte completa.
BEGIN TRY
 BEGIN TRANSACTION;
 DECLARE @key NVARCHAR(100)=N'migration-mvp2-group-v1',@now DATETIME2(0)=SYSUTCDATETIME(),@old INT,@new INT;
 SELECT @new=id FROM dbo.reservation_policies WITH(UPDLOCK,HOLDLOCK) WHERE idempotency_key=@key;
 IF @new IS NULL
 BEGIN
  SELECT TOP(1) @old=id FROM dbo.reservation_policies WITH(UPDLOCK,HOLDLOCK) WHERE effective_to IS NULL AND is_published=1 ORDER BY effective_from DESC,id DESC;
  IF @old IS NULL THROW 52003,'Publicacion: no existe politica vigente publicada.',1;
  UPDATE dbo.reservation_policies SET effective_to=@now WHERE id=@old;
  INSERT INTO dbo.reservation_policies(reservable_window_days,request_frequency_days,confirmation_deadline_minutes,minimum_participants,opening_minute,closing_minute,slot_interval_minutes,effective_from,idempotency_key,idempotency_payload_hash,is_published)
  SELECT reservable_window_days,request_frequency_days,confirmation_deadline_minutes,10,opening_minute,closing_minute,slot_interval_minutes,@now,@key,'1111111111111111111111111111111111111111111111111111111111111111',0
  FROM dbo.reservation_policies WHERE id=@old;
  SET @new=SCOPE_IDENTITY();
  INSERT INTO dbo.reservation_policy_durations(policy_id,duration_minutes) SELECT @new,duration_minutes FROM dbo.reservation_policy_durations WHERE policy_id=@old;
  INSERT INTO dbo.reservation_policy_resources(policy_id,resource_id)
   SELECT @new,resource_id FROM(SELECT resource_id FROM dbo.reservation_policy_resources WHERE policy_id=@old UNION SELECT id FROM dbo.resources WHERE id IN(1,2,7))s;
  INSERT INTO dbo.reservation_policy_group_resources(policy_id,resource_id) VALUES(@new,1),(@new,2),(@new,7);
  UPDATE dbo.reservation_policies SET is_published=1 WHERE id=@new;
 END
 ELSE IF NOT EXISTS(SELECT 1 FROM dbo.reservation_policies WHERE id=@new AND is_published=1 AND effective_to IS NULL)
  THROW 52004,'Recuperacion: existe una politica MVP2 parcial; no se modifico la vigencia. Revise la fila antes de reintentar.',1;
 COMMIT;
END TRY
BEGIN CATCH
 IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
 THROW;
END CATCH;
GO

-- POSTCHECK (devuelve una fila; cualquier 0 exige revision).
SELECT
 CONVERT(bit,CASE WHEN COL_LENGTH('dbo.reservation_policies','opening_minute') IS NOT NULL AND COL_LENGTH('dbo.reservation_policies','is_published') IS NOT NULL THEN 1 ELSE 0 END) AS policy_columns_ok,
 CONVERT(bit,CASE WHEN COL_LENGTH('dbo.reservations','group_capacity_snapshot') IS NOT NULL THEN 1 ELSE 0 END) AS snapshot_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.reservation_policy_group_resources','U') IS NOT NULL AND OBJECT_ID('dbo.reservation_participant_audit','U') IS NOT NULL THEN 1 ELSE 0 END) AS group_tables_ok,
 CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.reservation_policies WHERE idempotency_key='migration-mvp2-group-v1' AND is_published=1 AND effective_to IS NULL) THEN 1 ELSE 0 END) AS policy_published_ok;
GO
