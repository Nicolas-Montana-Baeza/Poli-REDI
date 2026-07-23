SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

-- PREFLIGHT: solo depende de las migraciones anteriores y no modifica datos.
IF OBJECT_ID('dbo.reservations','U') IS NULL
 OR COL_LENGTH('dbo.reservations','join_code_hash') IS NULL
 OR OBJECT_ID('dbo.participants','U') IS NULL
 THROW 54000,'Preflight: ejecute y verifique primero las migraciones 001 a 003.',1;
GO

-- FASE 1: tablas. CREATE es atomico; una reejecucion conserva la tabla.
IF OBJECT_ID('dbo.reservation_join_code_secrets','U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_join_code_secrets(
  reservation_id INT NOT NULL,
  key_version INT NOT NULL,
  nonce VARBINARY(32) NOT NULL,
  ciphertext VARBINARY(512) NOT NULL,
  rotated_at DATETIME2(0) NOT NULL CONSTRAINT df_join_code_secret_rotated DEFAULT SYSUTCDATETIME());
END;
GO

IF OBJECT_ID('dbo.reservation_group_expirations','U') IS NULL
BEGIN
 CREATE TABLE dbo.reservation_group_expirations(
  reservation_id INT NOT NULL,
  participant_count INT NOT NULL,
  minimum_participants INT NOT NULL,
  expired_at DATETIME2(0) NOT NULL CONSTRAINT df_group_expired_at DEFAULT SYSUTCDATETIME());
END;
GO

-- PREFLIGHT DE ESTADO PARCIAL: no intenta reparar una definicion divergente.
IF EXISTS (
 SELECT 1 FROM (VALUES
  ('reservation_id','int',4,10,0,0),
  ('key_version','int',4,10,0,0),
  ('nonce','varbinary',32,0,0,0),
  ('ciphertext','varbinary',512,0,0,0),
  ('rotated_at','datetime2',6,19,0,0)
 ) expected(name,type_name,max_length,precision,scale,is_nullable)
 LEFT JOIN sys.columns c ON c.object_id=OBJECT_ID('dbo.reservation_join_code_secrets') AND c.name=expected.name
 WHERE c.column_id IS NULL OR TYPE_NAME(c.user_type_id)<>expected.type_name OR c.max_length<>expected.max_length
    OR c.precision<>expected.precision OR c.scale<>expected.scale OR c.is_nullable<>expected.is_nullable
) OR (SELECT COUNT(*) FROM sys.columns WHERE object_id=OBJECT_ID('dbo.reservation_join_code_secrets'))<>5
 THROW 54001,'Estado parcial incompatible: reservation_join_code_secrets tiene columnas divergentes.',1;
GO

IF EXISTS (
 SELECT 1 FROM (VALUES
  ('reservation_id','int',4,10,0,0),
  ('participant_count','int',4,10,0,0),
  ('minimum_participants','int',4,10,0,0),
  ('expired_at','datetime2',6,19,0,0)
 ) expected(name,type_name,max_length,precision,scale,is_nullable)
 LEFT JOIN sys.columns c ON c.object_id=OBJECT_ID('dbo.reservation_group_expirations') AND c.name=expected.name
 WHERE c.column_id IS NULL OR TYPE_NAME(c.user_type_id)<>expected.type_name OR c.max_length<>expected.max_length
    OR c.precision<>expected.precision OR c.scale<>expected.scale OR c.is_nullable<>expected.is_nullable
) OR (SELECT COUNT(*) FROM sys.columns WHERE object_id=OBJECT_ID('dbo.reservation_group_expirations'))<>4
 THROW 54002,'Estado parcial incompatible: reservation_group_expirations tiene columnas divergentes.',1;
GO

-- FASE 2: PK/UNIQUE. Una PK distinta es incompatible; la ausencia es reparable.
IF EXISTS(SELECT 1 FROM sys.key_constraints WHERE parent_object_id=OBJECT_ID('dbo.reservation_join_code_secrets') AND type='PK' AND name<>'pk_reservation_join_code_secrets')
 THROW 54003,'Estado parcial incompatible: PK inesperada en reservation_join_code_secrets.',1;
IF OBJECT_ID('dbo.pk_reservation_join_code_secrets','PK') IS NULL
 ALTER TABLE dbo.reservation_join_code_secrets ADD CONSTRAINT pk_reservation_join_code_secrets PRIMARY KEY(reservation_id);
GO

IF EXISTS(SELECT 1 FROM sys.key_constraints WHERE parent_object_id=OBJECT_ID('dbo.reservation_group_expirations') AND type='PK' AND name<>'pk_reservation_group_expirations')
 THROW 54004,'Estado parcial incompatible: PK inesperada en reservation_group_expirations.',1;
IF OBJECT_ID('dbo.pk_reservation_group_expirations','PK') IS NULL
 ALTER TABLE dbo.reservation_group_expirations ADD CONSTRAINT pk_reservation_group_expirations PRIMARY KEY(reservation_id);
GO

-- FASE 3: defaults, checks y FKs canonicos.
IF OBJECT_ID('dbo.pk_reservation_join_code_secrets','PK') IS NOT NULL AND (
 (SELECT COUNT(*) FROM sys.index_columns WHERE object_id=OBJECT_ID('dbo.reservation_join_code_secrets') AND index_id=(SELECT unique_index_id FROM sys.key_constraints WHERE name='pk_reservation_join_code_secrets'))<>1
 OR NOT EXISTS(SELECT 1 FROM sys.index_columns ic INNER JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id WHERE ic.object_id=OBJECT_ID('dbo.reservation_join_code_secrets') AND ic.index_id=(SELECT unique_index_id FROM sys.key_constraints WHERE name='pk_reservation_join_code_secrets') AND c.name='reservation_id' AND ic.key_ordinal=1))
 THROW 54005,'Estado parcial incompatible: PK de secretos no garantiza reservation_id.',1;
IF OBJECT_ID('dbo.pk_reservation_group_expirations','PK') IS NOT NULL AND (
 (SELECT COUNT(*) FROM sys.index_columns WHERE object_id=OBJECT_ID('dbo.reservation_group_expirations') AND index_id=(SELECT unique_index_id FROM sys.key_constraints WHERE name='pk_reservation_group_expirations'))<>1
 OR NOT EXISTS(SELECT 1 FROM sys.index_columns ic INNER JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id WHERE ic.object_id=OBJECT_ID('dbo.reservation_group_expirations') AND ic.index_id=(SELECT unique_index_id FROM sys.key_constraints WHERE name='pk_reservation_group_expirations') AND c.name='reservation_id' AND ic.key_ordinal=1))
 THROW 54006,'Estado parcial incompatible: unicidad de expiracion no usa reservation_id.',1;
GO

IF OBJECT_ID('dbo.ck_join_code_secret_key_version','C') IS NOT NULL AND OBJECT_DEFINITION(OBJECT_ID('dbo.ck_join_code_secret_key_version')) NOT LIKE '%[[]key_version]>(0)%'
 THROW 54007,'Estado parcial incompatible: CHECK de key_version divergente.',1;
IF OBJECT_ID('dbo.ck_group_expiration_counts','C') IS NOT NULL AND (
 OBJECT_DEFINITION(OBJECT_ID('dbo.ck_group_expiration_counts')) NOT LIKE '%[[]participant_count]>=(0)%'
 OR OBJECT_DEFINITION(OBJECT_ID('dbo.ck_group_expiration_counts')) NOT LIKE '%[[]minimum_participants]>(0)%'
 OR OBJECT_DEFINITION(OBJECT_ID('dbo.ck_group_expiration_counts')) NOT LIKE '%[[]participant_count]<[[]minimum_participants]%')
 THROW 54008,'Estado parcial incompatible: CHECK de conteos divergente.',1;
GO

IF OBJECT_ID('dbo.df_join_code_secret_rotated','D') IS NULL
 ALTER TABLE dbo.reservation_join_code_secrets ADD CONSTRAINT df_join_code_secret_rotated DEFAULT SYSUTCDATETIME() FOR rotated_at;
IF OBJECT_ID('dbo.ck_join_code_secret_key_version','C') IS NULL
 ALTER TABLE dbo.reservation_join_code_secrets ADD CONSTRAINT ck_join_code_secret_key_version CHECK(key_version>0);
IF OBJECT_ID('dbo.fk_join_code_secret_reservation','F') IS NULL
 ALTER TABLE dbo.reservation_join_code_secrets ADD CONSTRAINT fk_join_code_secret_reservation FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id) ON DELETE CASCADE;
GO

IF OBJECT_ID('dbo.df_group_expired_at','D') IS NULL
 ALTER TABLE dbo.reservation_group_expirations ADD CONSTRAINT df_group_expired_at DEFAULT SYSUTCDATETIME() FOR expired_at;
IF OBJECT_ID('dbo.ck_group_expiration_counts','C') IS NULL
 ALTER TABLE dbo.reservation_group_expirations ADD CONSTRAINT ck_group_expiration_counts CHECK(participant_count>=0 AND minimum_participants>0 AND participant_count<minimum_participants);
IF OBJECT_ID('dbo.fk_group_expiration_reservation','F') IS NULL
 ALTER TABLE dbo.reservation_group_expirations ADD CONSTRAINT fk_group_expiration_reservation FOREIGN KEY(reservation_id) REFERENCES dbo.reservations(id);
GO

IF EXISTS(SELECT 1 FROM sys.foreign_keys WHERE name='fk_join_code_secret_reservation' AND (referenced_object_id<>OBJECT_ID('dbo.reservations') OR delete_referential_action_desc<>'CASCADE'))
 OR NOT EXISTS(SELECT 1 FROM sys.foreign_key_columns fkc INNER JOIN sys.columns pc ON pc.object_id=fkc.parent_object_id AND pc.column_id=fkc.parent_column_id INNER JOIN sys.columns rc ON rc.object_id=fkc.referenced_object_id AND rc.column_id=fkc.referenced_column_id WHERE fkc.constraint_object_id=OBJECT_ID('dbo.fk_join_code_secret_reservation') AND pc.name='reservation_id' AND rc.name='id')
 THROW 54009,'Estado parcial incompatible: FK de secretos divergente.',1;
IF EXISTS(SELECT 1 FROM sys.foreign_keys WHERE name='fk_group_expiration_reservation' AND (referenced_object_id<>OBJECT_ID('dbo.reservations') OR delete_referential_action_desc<>'NO_ACTION'))
 OR NOT EXISTS(SELECT 1 FROM sys.foreign_key_columns fkc INNER JOIN sys.columns pc ON pc.object_id=fkc.parent_object_id AND pc.column_id=fkc.parent_column_id INNER JOIN sys.columns rc ON rc.object_id=fkc.referenced_object_id AND rc.column_id=fkc.referenced_column_id WHERE fkc.constraint_object_id=OBJECT_ID('dbo.fk_group_expiration_reservation') AND pc.name='reservation_id' AND rc.name='id')
 THROW 54010,'Estado parcial incompatible: FK de expiracion divergente.',1;
IF EXISTS(SELECT 1 FROM sys.default_constraints WHERE name IN('df_join_code_secret_rotated','df_group_expired_at') AND definition NOT LIKE '%sysutcdatetime%')
 THROW 54011,'Estado parcial incompatible: DEFAULT temporal divergente.',1;
GO

-- POSTCHECK exhaustivo: todos los indicadores deben valer 1.
SELECT
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.reservation_join_code_secrets','U') IS NOT NULL THEN 1 ELSE 0 END) join_code_secrets_table_ok,
 CONVERT(bit,CASE WHEN (SELECT COUNT(*) FROM sys.columns c WHERE c.object_id=OBJECT_ID('dbo.reservation_join_code_secrets') AND c.is_nullable=0 AND ((c.name IN('reservation_id','key_version') AND TYPE_NAME(c.user_type_id)='int' AND c.max_length=4 AND c.precision=10 AND c.scale=0) OR (c.name='nonce' AND TYPE_NAME(c.user_type_id)='varbinary' AND c.max_length=32) OR (c.name='ciphertext' AND TYPE_NAME(c.user_type_id)='varbinary' AND c.max_length=512) OR (c.name='rotated_at' AND TYPE_NAME(c.user_type_id)='datetime2' AND c.max_length=6 AND c.precision=19 AND c.scale=0)))=5 THEN 1 ELSE 0 END) join_code_secrets_columns_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.pk_reservation_join_code_secrets','PK') IS NOT NULL THEN 1 ELSE 0 END) join_code_secrets_pk_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.fk_join_code_secret_reservation','F') IS NOT NULL THEN 1 ELSE 0 END) join_code_secrets_fk_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.ck_join_code_secret_key_version','C') IS NOT NULL THEN 1 ELSE 0 END) join_code_secrets_check_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.df_join_code_secret_rotated','D') IS NOT NULL THEN 1 ELSE 0 END) join_code_secrets_default_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.reservation_group_expirations','U') IS NOT NULL THEN 1 ELSE 0 END) group_expirations_table_ok,
 CONVERT(bit,CASE WHEN (SELECT COUNT(*) FROM sys.columns c WHERE c.object_id=OBJECT_ID('dbo.reservation_group_expirations') AND c.is_nullable=0 AND ((c.name IN('reservation_id','participant_count','minimum_participants') AND TYPE_NAME(c.user_type_id)='int' AND c.max_length=4 AND c.precision=10 AND c.scale=0) OR (c.name='expired_at' AND TYPE_NAME(c.user_type_id)='datetime2' AND c.max_length=6 AND c.precision=19 AND c.scale=0)))=4 THEN 1 ELSE 0 END) group_expirations_columns_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.pk_reservation_group_expirations','PK') IS NOT NULL THEN 1 ELSE 0 END) group_expirations_unique_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.fk_group_expiration_reservation','F') IS NOT NULL THEN 1 ELSE 0 END) group_expirations_fk_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.ck_group_expiration_counts','C') IS NOT NULL THEN 1 ELSE 0 END) group_expirations_check_ok,
 CONVERT(bit,CASE WHEN OBJECT_ID('dbo.df_group_expired_at','D') IS NOT NULL THEN 1 ELSE 0 END) group_expirations_default_ok;
GO
